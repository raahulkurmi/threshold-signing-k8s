package coordinator_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"frost-k8s-threshold-signing/internal/coordinator"
	"frost-k8s-threshold-signing/internal/testutil"
	"frost-k8s-threshold-signing/internal/wire"
)

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestResponseSANCheckWithoutTLSPinning (audit C-3, mutation M5a): the
// coordinator checks the responder's certificate SAN against the endpoint id
// itself, not only through the TLS client configuration. Here endpoint 3 uses a
// client that trusts the CA but does not pin signer-3 (it expects signer-4) and
// points at signer 4's server. TLS succeeds; the coordinator must still refuse
// the response as not coming from signer-3, before any share verification.
//
// The coordinator stops collecting once 3 valid shares are in, so a reply from
// endpoint 3 that arrives after that is never recorded (as in signWithRogue,
// audit C-4/C-5). Honest endpoints 4 and 5 hold their sign-share request until
// endpoint 3 has replied (plus a margin for the coordinator to evaluate it): the
// quorum then needs one of them, and endpoint 3's reply is always evaluated first.
func TestResponseSANCheckWithoutTLSPinning(t *testing.T) {
	c := testutil.StartCluster(t, testutil.ClusterOpts{})
	eps := c.Endpoints(t, 1, 2, 4, 5)
	cert, err := tls.LoadX509KeyPair(c.CoordinatorCert().Cert, c.CoordinatorCert().Key)
	if err != nil {
		t.Fatal(err)
	}
	caPEM, err := os.ReadFile(c.PKI.CA)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("bad CA")
	}
	replied := make(chan struct{})
	var once sync.Once
	for i := range eps {
		if eps[i].ID != 4 && eps[i].ID != 5 {
			continue
		}
		honest := eps[i].Client.Transport
		eps[i].Client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == wire.SignSharePath {
				select {
				case <-replied:
					time.Sleep(200 * time.Millisecond)
				case <-r.Context().Done():
					return nil, r.Context().Err()
				}
			}
			return honest.RoundTrip(r)
		})
	}
	unpinned := &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}, RootCAs: pool, ServerName: "signer-4"}
	rogue := &http.Transport{TLSClientConfig: unpinned, ForceAttemptHTTP2: true}
	eps = append(eps, coordinator.Endpoint{ID: 3, URL: c.Servers[4].URL,
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == wire.SignSharePath {
				defer once.Do(func() { close(replied) })
			}
			return rogue.RoundTrip(r)
		})}})
	co, err := coordinator.New(coordinator.Config{Meta: c.Fx.Meta, Endpoints: eps, Deadline: 5 * time.Second, Strategy: coordinator.Strict})
	if err != nil {
		t.Fatal(err)
	}
	res, err := co.Sign(context.Background(), claims(t))
	if err != nil {
		t.Fatal(err) // signers 1, 2, 4, 5 are honest: the token still forms
	}
	for _, id := range res.Signers {
		if id == 3 {
			t.Fatalf("combined a share attributed to signer 3: %v", res.Signers)
		}
	}
	found := false
	for _, f := range res.Excluded {
		if f.SignerID == 3 {
			found = true
			if !strings.Contains(f.Reason, "is not signer-3") {
				t.Fatalf("endpoint 3 excluded for %q; want the response-side identity check (\"is not signer-3\")", f.Reason)
			}
		}
	}
	if !found {
		t.Fatalf("endpoint 3 (answered by signer-4) not excluded: %+v", res.Excluded)
	}
}
