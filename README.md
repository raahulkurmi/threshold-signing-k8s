# threshold-signing

**Threshold-signed Kubernetes service account tokens.**

Kubernetes gives every program running in a cluster a *service account token*: a signed
JSON Web Token (JWT, a short signed text that says "this is service account X"). Normally
one private key, kept on the cluster's control plane, signs all of these tokens. Whoever
steals that one key can make tokens for anyone. This project removes that single key. It
splits the signing key into 5 *shares*, gives each share to a different *signer*, and
makes a token only when at least 3 of the 5 signers each produce a partial signature
("3-of-5 threshold RSA", using the Shoup scheme). After the key is created, no program
ever holds the whole private key. Kubernetes (kube-apiserver v1.36 or later) talks to
this signer through its standard ExternalJWTSigner interface, and it checks the result
as an ordinary RS256 token, so nothing in Kubernetes is modified.

What it does **not** protect against, in plain words:

- **The key is made by one trusted dealer.** One program (`cmd/dealer`) creates the key
  and splits it. During that moment it sees the whole key. A dishonest dealer could keep
  a copy and forge tokens forever.
- **It is an online oracle.** Whoever controls a coordinator (the component that asks
  the signers for partial signatures) can get any token the signers' rules allow, for as
  long as they are in control. That includes tokens for powerful `kube-system` accounts
  under the shipped rules. They lose that power as soon as they lose control.
- **The threshold library is unaudited.** It uses `niclabs/tcrsa` v0.0.5 (2020), which no
  one has audited and which is no longer maintained.
- **The signers are not independent in any tested setup.** The threshold only helps if
  the 5 signers fail independently. The widest test put one signer in each of 5 AWS
  regions, but still with **one provider, one account, one operator, one build and one
  dealer**. Whoever controls that account controls every signer.

> **Status: open-source research prototype (Apache-2.0).** Not for production. Read
> [Trust assumptions](#trust-assumptions) and [docs/THREAT_MODEL.md](docs/THREAT_MODEL.md)
> before relying on anything here.

## Papers

- **Systematization:** "Service-Account Token Signing in Kubernetes: A Systematization of
  Mechanisms, Threats, and Threshold-Based Defenses". Zenodo.
  <https://doi.org/10.5281/zenodo.20733862>
- **System and measurements:** "Threshold-Signed Service-Account Tokens for Kubernetes:
  Design, Measured Cost, and Quorum Waste under Overload". Zenodo.
  <https://doi.org/10.5281/zenodo.20734723>

## Contents

1. [What you need](#what-you-need)
2. [Quick start: build and run the tests](#quick-start-build-and-run-the-tests)
3. [Run the full end-to-end test on a local cluster](#run-the-full-end-to-end-test-on-a-local-cluster)
4. [Reproduce everything on a fresh machine](#reproduce-everything-on-a-fresh-machine)
5. [AWS experiments (optional, costs money)](#aws-experiments-optional-costs-money)
6. [Troubleshooting](#troubleshooting)
7. [Repository layout](#repository-layout)
8. [How it works](#how-it-works), [Trust assumptions](#trust-assumptions), [Tests and evidence](#tests-and-evidence)
9. [Reference: defaults and operations](#reference-defaults-and-operations)
10. [Licensing and patents](#licensing-and-patents)
11. [About the git history](#about-the-git-history)

## What you need

### Which computer

| What you want to run | Linux (Ubuntu 24.04) | macOS |
|---|---|---|
| Build and unit tests (`make build`, `make test`) | yes | yes |
| Image check (`make check-images`) | yes | yes, with Docker Desktop |
| End-to-end test on a local cluster (`make e2e`) | yes, with **Docker Engine** | **no** (see below) |
| Full reproduction (`make repro`) | **Ubuntu 24.04 only** | **no** |

- **Linux:** tested on Ubuntu 24.04 (amd64 on GitHub's runners, arm64 in a local VM).
  Other recent Linux distributions should work for the build, tests and e2e, but only
  Ubuntu 24.04 was tested. `make repro` refuses anything except Ubuntu 24.04.
- **macOS:** tested on macOS 26 (Apple silicon). The build, the tests and the image check
  work. The end-to-end test does **not** run on macOS. It stops at once with
  `FATAL: e2e runs on a Linux host only`, because it needs Docker Engine on Linux, not
  Docker Desktop. To run it from a Mac, use an Ubuntu 24.04 virtual machine (for example
  [Multipass](https://canonical.com/multipass): `multipass launch 24.04 --cpus 4 --memory 8G --disk 40G`)
  and follow the Linux steps inside it.
- **Windows:** not tested. Use an Ubuntu 24.04 virtual machine.

**Hardware.** At least **4 CPU cores, 8 GB of RAM and 25 GB of free disk**. All steps
passed on a VM with exactly 4 cores, 8 GB of RAM and a 40 GB disk. The tests use a lot
of CPU (they generate RSA keys), so fewer cores means much longer runs.

**Network.** The first run downloads Go modules, Docker images and, for the end-to-end
test, about 320 MB of Kubernetes release files.

### Tools and exact versions

You need these tools. The version numbers are the ones the project pins and tests with.

| Tool | Version | Needed for | Official install page |
|---|---|---|---|
| git | any recent | everything | <https://git-scm.com/downloads> |
| make | any (GNU Make) | everything | Ubuntu: `sudo apt-get install -y make`; macOS: `xcode-select --install` |
| Go | **1.27.2** | build, tests | <https://go.dev/dl/> |
| Docker | Docker **Engine** on Linux; Docker Desktop on macOS (image check only) | image check, e2e | <https://docs.docker.com/engine/install/ubuntu/>, <https://docs.docker.com/desktop/setup/install/mac-install/> |
| kind | **v0.31.0** | e2e | <https://kind.sigs.k8s.io/docs/user/quick-start/#installation> |
| kubectl | **v1.36.5** | e2e | <https://kubernetes.io/docs/tasks/tools/> |
| gitleaks | **8.30.1** | the "no secrets in the tree" test (skipped if missing) | <https://github.com/gitleaks/gitleaks/releases/tag/v8.30.1> |
| jq, openssl, curl | any recent | e2e | Ubuntu: `sudo apt-get install -y jq openssl curl` |

Some words used below:

- **Go** is the programming language the project is written in.
- **Docker** runs programs in isolated *containers*. **Docker Engine** is the Linux
  version. **Docker Desktop** is the Mac/Windows app, which runs Docker inside a hidden
  virtual machine.
- **Kubernetes** runs containers across machines. **kind** ("Kubernetes in Docker")
  creates a small Kubernetes cluster inside Docker containers on your own computer.
  **kubectl** is the command-line tool that talks to a Kubernetes cluster.
- **gitleaks** scans files for passwords and private keys.

#### Install on Ubuntu 24.04

Run these commands one block at a time. They need `sudo` (administrator rights). The
`ARCH` line works out whether your computer is `amd64` (Intel/AMD) or `arm64` (ARM).

```bash
# 1. Basic tools
sudo apt-get update
sudo apt-get install -y git make jq openssl curl ca-certificates

# 2. Docker Engine (Docker's official convenience script), and permission to use it
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker "$USER"
# Now log out and log back in (or reboot) so the new group takes effect.

# 3. Go 1.27.2
ARCH=$(dpkg --print-architecture)
curl -fsSLO "https://go.dev/dl/go1.27.2.linux-$ARCH.tar.gz"
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "go1.27.2.linux-$ARCH.tar.gz"
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.profile
export PATH=/usr/local/go/bin:$PATH

# 4. kind v0.31.0 and kubectl v1.36.5
curl -fsSLo kind "https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-$ARCH"
curl -fsSLo kubectl "https://dl.k8s.io/release/v1.36.5/bin/linux/$ARCH/kubectl"
sudo install -m 0755 kind kubectl /usr/local/bin/

# 5. gitleaks 8.30.1 (its file names say x64 instead of amd64)
GL_ARCH=$([ "$ARCH" = amd64 ] && echo x64 || echo arm64)
curl -fsSLO "https://github.com/gitleaks/gitleaks/releases/download/v8.30.1/gitleaks_8.30.1_linux_${GL_ARCH}.tar.gz"
tar -xzf "gitleaks_8.30.1_linux_${GL_ARCH}.tar.gz" gitleaks
sudo install -m 0755 gitleaks /usr/local/bin/
```

These commands do not check download checksums. `scripts/repro.sh` installs the same
versions and does check them (see [Reproduce everything](#reproduce-everything-on-a-fresh-machine)).

#### Install on macOS

1. **make and git:** run `xcode-select --install` and accept the dialog.
2. **Go 1.27.2:** download and open the installer from <https://go.dev/dl/>
   (`go1.27.2.darwin-arm64.pkg` for Apple silicon, `go1.27.2.darwin-amd64.pkg` for Intel).
3. **Docker Desktop** (only for `make check-images`):
   <https://docs.docker.com/desktop/setup/install/mac-install/>. Start it once.
4. **gitleaks 8.30.1** (Apple silicon shown; on Intel replace `arm64` with `x64`):

   ```bash
   curl -fsSLO https://github.com/gitleaks/gitleaks/releases/download/v8.30.1/gitleaks_8.30.1_darwin_arm64.tar.gz
   tar -xzf gitleaks_8.30.1_darwin_arm64.tar.gz gitleaks
   sudo install -m 0755 gitleaks /usr/local/bin/
   ```

kind and kubectl are not needed on macOS, because the end-to-end test does not run there.

The macOS install steps were checked by downloading each file and running it from a
temporary folder; they were not repeated on a clean Mac.

### Check that everything is installed

Run each command. Your output should look like the sample (on Intel/AMD Linux you will
see `amd64` instead of `arm64`; Docker's version number may be newer).

```console
$ go version
go version go1.27.2 linux/arm64

$ docker version --format '{{.Server.Version}}'
29.9.1

$ kind version
kind v0.31.0 go1.25.5 linux/arm64

$ kubectl version --client
Client Version: v1.36.5
Kustomize Version: v5.8.1

$ gitleaks version
8.30.1

$ make --version
GNU Make 4.3

$ docker ps
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
```

If `docker ps` says `permission denied`, see [Troubleshooting](#docker-permission-denied).

## Quick start: build and run the tests

This works on Linux and macOS. It builds the code and runs the unit and integration
tests. It does not need Docker or Kubernetes.

**1. Download the code** (about 1 to 3 minutes; the repository is about 200 MB because
it contains measurement results).

```console
$ git clone https://github.com/raahulkurmi/threshold-signing-k8s.git
Cloning into 'threshold-signing-k8s'...
```

**2. Go into the folder.**

```console
$ cd threshold-signing-k8s
```

**3. Build** (about 15 seconds the first time, because Go also downloads the
libraries; a few seconds after that).

```console
$ make build
go build ./...
go: downloading github.com/niclabs/tcrsa v0.0.5
go: downloading google.golang.org/grpc v1.84.0
...
go: downloading golang.org/x/text v0.42.0
```

The `go: downloading` lines appear only the first time. Success means no line contains
`error` and the command returns you to the prompt.

**4. Run the tests** (about 4 minutes on a 10-core Mac, 8 to 11 minutes on a 4-core
machine). Keep the computer awake: on a laptop, run `caffeinate -i make test` on macOS,
because a sleeping computer makes the clock-check tests fail.

```console
$ make test
go test -count=1 -timeout 30m ./...
ok  	frost-k8s-threshold-signing/cmd/dealer	241.209s
ok  	frost-k8s-threshold-signing/cmd/grpc-proxy	140.079s
ok  	frost-k8s-threshold-signing/cmd/signer	118.318s
ok  	frost-k8s-threshold-signing/internal/coordinator	50.085s
...
ok  	frost-k8s-threshold-signing/test	76.685s
go test -count=1 -timeout 30m -tags testmalicious -run 'TestMaliciousSignerExcluded' -v ./test/
...
--- PASS: TestMaliciousSignerExcluded (64.13s)
PASS
ok  	frost-k8s-threshold-signing/test	75.753s
cd spike && go test -count=1 -timeout 30m -v .
...
PASS
ok  	github.com/raahulkurmi/frost-k8s-threshold-signing/spike	124.938s
```

**It passed if** every package line starts with `ok` (lines with `[no test files]` are
fine), and the command ends without `FAIL` or `make: *** ... Error`. You can check right
after it finishes:

```console
$ echo $?
0
```

`0` means success. Any other number means something failed; see
[Troubleshooting](#troubleshooting).

**5. Optional: check the container images** (needs Docker; about 1 minute). This builds
the coordinator and signer images and proves that no key or share is inside them.

```console
$ make check-images
scripts/check-images.sh
== coordinator image frost-k8s/coordinator:dev (sha256:902488e64ede)
  layers inspected: 3
  files in image: 90, binaries: grpc-proxy
  history: 6 entries, no secret references
== signer image frost-k8s/signer:dev (sha256:343dfcd7136c)
  layers inspected: 3
  files in image: 90, binaries: signer
  history: 6 entries, no secret references
check-images: PASS (no shares, keys, certs, encrypted blobs or private keys in any layer of either image)
```

The `sha256:` values differ on your machine. That is expected.

## Run the full end-to-end test on a local cluster

**Linux with Docker Engine only.** This test creates a real Kubernetes v1.36.5 cluster
with kind, connects it to 5 signers, 3 coordinators and an nginx proxy running in Docker,
and checks that every token in the cluster is threshold-signed. It also stops signers to
show that 2 missing signers are fine and 3 missing signers stop token issuance. It needs
`sudo`, because the signing socket's folder must belong to root. It may ask for your
password.

**1. Do the [Quick start](#quick-start-build-and-run-the-tests) steps 1 and 2 first**
(clone and `cd`). You do not have to run the tests first.

**2. Build the Kubernetes node image** (once per machine). kind has no ready-made image
for Kubernetes v1.36.5, so you build one from the official release. It downloads about
320 MB. This took under 1 minute on GitHub's runners and 10 to 25 minutes on a slow
home connection.

```console
$ kind build node-image --type release v1.36.5 --image kindest/node:v1.36.5-tk8s
...
Image "kindest/node:v1.36.5-tk8s" build completed.
```

Check that it is there:

```console
$ docker image ls kindest/node
IMAGE                       ID             DISK USAGE   CONTENT SIZE   EXTRA
kindest/node:v1.36.5-tk8s   655d67b3019c       1.35GB          374MB
```

If the download keeps failing part-way, see
[Troubleshooting](#the-node-image-download-keeps-failing).

**3. Run the end-to-end test** (about 3 to 6 minutes).

```console
$ make e2e
test/e2e/run.sh

================ Environment ================
git commit: fc207aef1e27e793769ea7e6a514970d39e04029
...
================ Setup: fresh certs, key ceremony, signer stack ================
Generated in secrets/:
...
Generating 3-of-5 threshold RSA key, 2048-bit modulus (safe primes; this can take minutes)...
...
================ Summary ================
git commit: fc207aef1e27e793769ea7e6a514970d39e04029
kubernetes: v1.36.5, node image sha256:655d67b3019c5eab2058e71c4f779039a39a941f20729bd891b8322487c39ef4
PASS  SETUP  apiserver uses only --service-account-signing-endpoint; no in-tree SA key flags on apiserver or controller-manager
PASS  E1  alg=RS256 kid=Jo9puoTBb-W1-Y52PQGbLA (equals public-meta.json)
PASS  E2  authenticated=true username=system:serviceaccount:default:default
PASS  E3  pod reached the API (HTTP 200) and authenticated as system:serviceaccount:default:default with its projected token
...
PASS  E6  2 down: issued; 3 down: refused with threshold error (59ms); old tokens still authenticate
...
PASS  N76-PRIO  5 signers ready with admission=priority, stable-identity priority (N77), rotation 32 (shared priority.key mounted); token issued, kid Jo9puoTBb-W1-Y52PQGbLA, TokenReview authenticated
E2E: PASS
cluster, stack and generated secrets removed
```

**It passed if** the last lines say `E2E: PASS` and every line in the Summary starts
with `PASS`. The `kid` and hash values are different on every run, because every run
makes a fresh key. The full log is saved in `test/e2e/results/<date>-<commit>/e2e.log`.

**4. Clean up.** `make e2e` cleans up after itself, whether it passes or fails: it
deletes the cluster, the containers and every generated key. You can check:

```console
$ kind get clusters
No kind clusters found.
```

The only thing left is the new log folder, which git shows as untracked:

```console
$ git status --short
?? test/e2e/results/
```

Delete it with `rm -rf test/e2e/results/` if you do not want it.

**To keep the cluster running** after the test (for example, to look around with
`kubectl --context kind-tk8s get pods -A`), run `make e2e-keep` instead of `make e2e`.
When you are done, remove the cluster, containers and generated keys with:

```console
$ make e2e-down
kind delete cluster --name tk8s
Deleting cluster "tk8s" ...
FROST_UID=$(id -u) docker compose -p tk8s -f deploy/docker-compose.yml down -v --remove-orphans
rm -rf secrets audit bin/e2e
sudo rm -rf run
```

`make e2e-down` is safe to run even when nothing is running.

## Reproduce everything on a fresh machine

`make repro` does everything in one command on a **fresh Ubuntu 24.04** machine: it
installs the pinned tools (checking their checksums and signatures), builds the
Kubernetes node image, then runs `make test`, `make check-images` and `make e2e`, and
writes a report to `reports/REPRO.md`. It needs passwordless `sudo` (as on cloud VMs and
Multipass VMs) and refuses Docker Desktop and any other operating system.

**1. Install git and make, and clone** (a fresh Ubuntu cloud image has git but no make):

```console
$ sudo apt-get install -y make
$ git clone https://github.com/raahulkurmi/threshold-signing-k8s.git
$ cd threshold-signing-k8s
```

**2. Run it** (about 16 minutes on a fast connection; about 45 to
55 minutes on a slow one, mostly downloads):

```console
$ make repro
scripts/repro.sh
== repro step: install-base-packages
   PASS (5s)
== repro step: install-docker-engine
   PASS (0s)
== repro step: install-kind-kubectl-gitleaks-go
   PASS (5s)
== repro step: build-kind-node-image
   PASS (29s)
== repro step: make-test
   PASS (652s)
== repro step: make-check-images
   PASS (37s)
== repro step: make-e2e
   PASS (217s)
...
Total: 946 s. Overall: **PASS**
```

(This output is from a fresh GitHub runner, Ubuntu 24.04, 4 CPUs, on commit `fc207ae`.)

**It passed if** the report ends with `Overall: **PASS**`. The report is in
`reports/REPRO.md` and each step's log is in `reports/repro-logs/`. If only the
`make-test` step failed, read its log and see [Troubleshooting](#troubleshooting).

Note that `make repro` **overwrites** the committed `reports/REPRO.md` in your clone,
which records the maintainers' own reproduction run. Run `git checkout reports/REPRO.md`
to restore it. The same reproduction also runs on a fresh GitHub runner through
`.github/workflows/repro.yml`.

## AWS experiments (optional, costs money)

The multi-region results in the papers come from Amazon Web Services (AWS) virtual
machines. **You do not need AWS to build, test or run this project.** These scripts
were not run for this README.

**Cost warning.** These scripts start many machines in up to 5 regions: the Phase 7C
setup uses one m7i-flex.large, three c7i-flex.large, one t3.small and up to ten
t3.micro instances, plus Elastic IPs and disks. The maintainers' estimate for one
5.7-hour measurement session was about **US $3.50** (NOTES N79). Prices change, and
anything you forget to delete keeps costing money. Always run the teardown script and
check its output.

What you need:

- an AWS account and the [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html),
  with a profile named `tk8s`;
- enough vCPU quota in the regions in `deploy/aws/common.sh`;
- to **edit `deploy/aws/common.sh`** first: `EXPECTED_ARN` is the maintainer's own AWS
  user, and every script stops if your identity does not match it. That check is
  deliberate, so you do not create machines in the wrong account.

Where to start:

| Script | What it does |
|---|---|
| `deploy/aws/provision.sh check` / `deploy/aws/provision-7c.sh check` | read-only: checks identity, quotas and a dry run of every planned machine |
| `deploy/aws/provision.sh`, `deploy/aws/provision-7c.sh` | create the machines (Phase 7A/7B, Phase 7C), step by step |
| `deploy/aws/teardown.sh` | deletes every resource tagged `Project=tk8s-research` and proves nothing is left |
| `deploy/multihost/README.md`, `test/e2e/multihost.sh` | multi-host deployment and its end-to-end test |
| `benchmark/k8s7c/run.sh`, `benchmark/multihost/run-l2.sh` | the Phase 7C and Level 2 benchmarks |
| `reports/aws/`, `reports/aws/7c/README.md` | records of the real runs and their teardowns |

## Troubleshooting

### Docker is not running

```
Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
```

- **Linux:** start it with `sudo systemctl start docker`. To start it at every boot:
  `sudo systemctl enable docker`.
- **macOS:** open the Docker Desktop app and wait until it says it is running.

### Docker: permission denied

```
permission denied while trying to connect to the docker API at unix:///var/run/docker.sock
```

Your user is not in the `docker` group yet. Run `sudo usermod -aG docker "$USER"`, then
**log out and log back in** (or reboot). Check with `groups`: the list must include
`docker`.

### Wrong Go version

`go version` must show `go1.27.2`. The Makefile asks Go for exactly 1.27.2
(`GOTOOLCHAIN=go1.27.2`), so any Go from 1.21 on downloads 1.27.2 by itself the first
time. That needs internet access. With Ubuntu's own Go 1.22 it looks like this:

```console
$ make build
go build ./...
go: downloading go1.27.2 (linux/arm64)
```
 If you see `go: command not found`, Go is not
on your `PATH`: run `export PATH=/usr/local/go/bin:$PATH` (Linux) or open a new terminal
after the macOS installer. If you see a Go older than 1.21, remove it and install 1.27.2
as shown above.

### `make test` fails on a laptop with "is -404s from signer clock"

```
rej=policy: iat: iat 1792165025 is -404s from signer clock (max skew 60s)
```

The computer went to sleep during the test. The test's clock check then sees a token
that is minutes old and rejects it, as designed. On macOS run `caffeinate -i make test`,
which keeps the Mac awake until the command ends. On Linux, turn off automatic suspend
while the tests run.

If you run the Linux steps in a virtual machine on a Mac, the VM is paused whenever the
Mac sleeps, so a step can look many times slower than the times given here. Keep the
Mac awake with `caffeinate -i` running in another terminal window.

### `make e2e` says "e2e runs on a Linux host only" or "refusing to run on Docker Desktop"

The end-to-end test needs Linux with Docker Engine. On a Mac or Windows computer, create
an Ubuntu 24.04 virtual machine (for example with Multipass, see
[What you need](#which-computer)) and run it there.

### A kind cluster already exists

`make e2e` always uses the cluster name `tk8s`, and it deletes any old `tk8s` cluster
before it starts, so a cluster left by an earlier run is not a problem. Clusters with
other names are not touched. To remove leftovers by hand: `make e2e-down`. To list all
kind clusters: `kind get clusters`; to delete one: `kind delete cluster --name <name>`.

### A port is already in use

The test stack publishes **no** ports on your computer: the components talk over private
Docker networks and a Unix socket, and the e2e test checks this (test N2). kind picks a
free local port for the Kubernetes API by itself. So other programs on your computer
using ports such as 80, 443, 6443 or 8080 do not block it. If kind reports a port
conflict anyway, another kind cluster is probably still running: see the section above.

### Permission errors on the `run/` folder

```
rm: cannot remove 'run': Permission denied
```

`run/` holds the signing socket. It belongs to root on purpose (mode 0700), so only root
can request signatures. `make e2e` and `make e2e-down` remove it with `sudo`. If you
delete the clone or the folder by hand, use `sudo rm -rf run` first. If `make e2e` stops
at a `sudo` password prompt, type your password; on machines without a password for
sudo, ask your administrator.

### The node image download keeps failing

`kind build node-image --type release` downloads one 320 MB file without resuming. On
some networks it is cut off every time (NOTES N55). `scripts/repro.sh` works around this
by downloading the official file with resume and checking its checksum, then using
`kind build node-image --type file`. You can do the same (replace `arm64` with `amd64`
on Intel/AMD):

```bash
curl -L --http1.1 -C - -o kubernetes-server-linux-arm64.tar.gz https://dl.k8s.io/v1.36.5/kubernetes-server-linux-arm64.tar.gz
echo "$(curl -fsSL https://dl.k8s.io/v1.36.5/kubernetes-server-linux-arm64.tar.gz.sha256)  kubernetes-server-linux-arm64.tar.gz" | sha256sum -c -
kind build node-image --type file kubernetes-server-linux-arm64.tar.gz --image kindest/node:v1.36.5-tk8s
```

Run the `curl` line again until it finishes, then the `sha256sum` line must print `OK`.

## Repository layout

| Folder | What it contains |
|---|---|
| `cmd/` | The programs: `dealer` (makes the key and the 5 shares), `signer` (holds one share), `grpc-proxy` (the coordinator that Kubernetes talks to) |
| `internal/` | The libraries those programs use: claims policy, signer, coordinator, key and share formats, TLS setup, audit log |
| `deploy/` | How to run it: Docker Compose files, nginx config, the signer policy, multi-host and AWS scripts |
| `test/` | Integration tests (T1–T12); `test/e2e/` holds the Kubernetes end-to-end test and the multi-host test |
| `scripts/` | Helper scripts: certificate generation, image check, govulncheck, `repro.sh` |
| `benchmark/` | Benchmark tools and drivers, and all measurement results in `benchmark/results/` |
| `spike/` | The first experiment showing that the threshold library makes tokens Kubernetes accepts |
| `docs/` | Threat model, key ceremony, invariants I1–I10, design notes |
| `reports/` | Audit, claims check, reproduction reports, test evidence, commit hash map |
| `legacy/frost/` | The abandoned first prototype (not built; see [About the git history](#about-the-git-history)) |
| `NOTES.md` | Every discrepancy, decision and finding, numbered N1, N2, … |

## How it works

```
kube-apiserver ──unix socket (dir 0700)──▶ nginx ──mTLS "lb"──▶ coordinator ×3 ──mTLS "coordinator-<k>"──▶ signer ×5
  (FetchKeys: group public key; Sign: header + claims)            (public metadata only)          (one share each,
                                                                                                    own claims policy)
```

(mTLS means both sides of a network connection prove who they are with certificates.)

1. **Key ceremony (trusted dealer).** `cmd/dealer` generates a 2048-bit Shoup 3-of-5 RSA key
   and writes `public-meta.json` (group public key, share verification keys, `kid`) and one
   `share-<i>.json` per signer. The full private key exists only in the dealer's memory
   during generation ([docs/KEY_CEREMONY.md](docs/KEY_CEREMONY.md)).
2. **Signers** (`cmd/signer`) each hold **exactly one** share. For a request they re-check the
   JWT header byte-for-byte, hash and PKCS#1 v1.5-encode the signing input **themselves**,
   apply an **independent claims policy** (issuer, audience allowlist, maximum lifetime, clock
   skew, subject form, deny lists, rate limit), write an audit record, and only then return a
   signature share. They admit work only if it can finish before the coordinator's deadline.
3. **Coordinators** (`cmd/grpc-proxy`) implement ExternalJWTSigner v1. They hold **only public
   metadata**. They fan the signing input out to the signers, combine 3 shares into a
   standard RS256 signature and verify it against the group key before returning it. By
   default (`optimistic`) shares are verified individually (Shoup proof) only after a failed
   combine, for signers already marked suspect, or in `strict` mode; the final-signature check
   always runs (NOTES N65). Fewer than 3 valid shares → a generic error, never a token.
4. **kube-apiserver** verifies tokens with the key from `FetchKeys` (also served at
   `/openid/v1/jwks`). Verification needs no signer.

The repository and Go module names still say "frost" because of the first prototype
(see [About the git history](#about-the-git-history)).

## Trust assumptions

State these whenever you describe the system:

- **Trusted dealer.** Shoup threshold RSA needs one. The dealer sees the whole key during
  the ceremony; a dishonest or compromised dealer can forge forever.
- **Signer independence is required for the security claim.** Threshold signing protects
  against compromise of fewer than 3 signers only if signers fail independently (different
  hosts, operators, credentials). The single-host deployment has **none**; the Level 1
  multi-host deployment is **multi-VM on one physical host**; Level 2 puts one signer in each
  of 5 AWS regions but keeps **one provider, one account, one operator, one build and one
  dealer**: whoever controls that account or operator controls every signer
  ([reports/INDEPENDENCE.md](reports/INDEPENDENCE.md)).
- **The signer claims policy limits *what* can be signed, not *who* asks.** Whoever controls
  a coordinator (or kube-apiserver) can obtain tokens for any **policy-compliant** claims
  while in control (an online oracle). What they cannot do is obtain policy-violating tokens,
  or mint tokens after losing control (no share or key is on the coordinator).
  **Under the shipped policy (`deploy/policy.json`, empty deny lists) "policy-compliant"
  includes tokens for kube-system service accounts**, some of which can escalate to cluster-admin (e.g.
  `kube-system:clusterrole-aggregation-controller`, which may escalate ClusterRoles). A live coordinator compromise is
  therefore, while it lasts, effectively a cluster compromise. What bounds it: the issuer,
  audience allowlist and maximum lifetime (7200 s), the ±60 s `iat` window, the per-signer rate
  limit, every signer's audit log, and the end of control (THREAT_MODEL §4). kube-system
  cannot simply be denied, because kube-controller-manager's own controllers use those
  tokens.
- **3 colluding or co-compromised signers can forge.** That is the threshold.
- **Unaudited cryptography.** `niclabs/tcrsa` v0.0.5 (2020) is unaudited and unmaintained;
  the coordinator guards its known hazards (docs/THREAT_MODEL.md §5).
- **Signers need correct clocks.** A signer whose clock is off by more than the policy's skew
  refuses every token (fail closed).

**Test coverage.** It has been tested end to end against a real kube-apiserver v1.36.5:
single host (`make e2e`), multi-VM on one physical host (Level 1), and one signer per AWS
region (Level 2: **5 regions, one provider, one account, one operator, one build, one
dealer**; [reports/INDEPENDENCE.md](reports/INDEPENDENCE.md)). Final benchmark numbers
come from Phase 7C (kubeadm v1.36.5 on AWS); Phase 7A is single-host and CPU-contended,
Phase 7B is T only with one run per configuration, and the Level 1 numbers are
preliminary (arm64).

## Tests and evidence

| What | Where |
|---|---|
| Invariants I1–I10 (definition, enforcing code, tests) | [docs/INVARIANTS.md](docs/INVARIANTS.md) |
| Unit / integration (T1–T12) | `make test`; `test/`, `internal/*/..._test.go` |
| Image isolation (T13) | `make check-images`; negative control in `reports/gates/gate6-t13-negative-control.txt` |
| Kubernetes e2e (single host) | `make e2e`; `reports/gates/gate6*.log`, `gate6.5*.log` |
| Multi-host e2e + isolation (L1–L5) | `test/e2e/multihost.sh`; `reports/multihost/e2e-*/` (Level 1 and Level 2) |
| Fresh-host reproduction | `make repro` / `.github/workflows/repro.yml`; [reports/REPRO.md](reports/REPRO.md), step logs in `reports/repro/` |
| Benchmarks | `benchmark/results/*/summary.md` (generated from raw CSVs): final Phase 7C `20260927T131349Z-73be90f-7C`; 7A, 7B and Level 1 with their labels |
| Phase 12 adversarial audit | [reports/audit/AUDIT.md](reports/audit/AUDIT.md) (mutation, fuzz, claims recheck) |
| Claims audit | [reports/CLAIMS_AUDIT.md](reports/CLAIMS_AUDIT.md) |
| Every discrepancy and finding | [NOTES.md](NOTES.md) |

Every benchmark number must carry its label and have a row in a script-generated
`summary.md`. **Final numbers: Phase 7C** (kubeadm v1.36.5 on AWS; B0 in-tree / B1 single-key
external signer / T threshold, one session, 3 runs; NOTES N69–N74). Phase 7A (single host,
2 vCPU, all components co-located; CPU-contended) and Phase 7B (Level 2, T only, one run per
configuration) keep their labels; the Level 1 numbers remain **preliminary: arm64, multi-VM on
one overloaded 16 GB host; not for publication**.

Other make targets: `make vet`, `make lint` (needs staticcheck), `make vuln` (needs
govulncheck), `make legacy` (checks that the old prototype still compiles).

## Reference: defaults and operations

### Defaults (final, NOTES N82)

| Setting | Default | Notes |
|---|---|---|
| Verification | `VERIFY_STRATEGY=optimistic` plus the breaker | combine t shares and verify the final signature; suspects and a strict fallback after repeated failed combines (N65) |
| Signer admission | `SIGNER_ADMISSION=n48` | deadline-aware bounded admission: `SIGNER_MAX_CONCURRENT` = CPU count, `SIGNER_MAX_QUEUE` = 64, deadline header capped at `SIGNER_MAX_DEADLINE` = 4 s (N48, N76) |
| Quorum-impossible abort | `QUORUM_ABORT=off` | opt-in (`on`). Evaluated: it saves signer work under queueing, but by its pre-registered rule it raised retry-storm wait and amplification (N78, N81) |
| Priority admission | off (`SIGNER_ADMISSION=n48`) | `priority` with controller `v1` or `v2` is selectable but **evaluated, not recommended**: neither met its pre-registered rules ([N78](NOTES.md), [N81](NOTES.md); `docs/PRIORITY_ADMISSION.md`, `docs/PRIORITY_ADMISSION_V2.md`) |
| Coordinator identity | one client certificate per replica (`coordinator-<k>`, `COORDINATOR_ID`) | signers accept only that canonical form (N76) |

### Operations notes

- **Signers need reliable time sync.** Each signer rejects claims whose `iat` is more
  than `clock_skew_seconds` (default 60 s) from its own clock, so a signer with a bad
  clock refuses everything. Run chrony/NTP on every signer host and monitor its offset.
  After a VM resume or snapshot restore, a signer fails closed until its clock resyncs
  (docs/THREAT_MODEL.md §3, NOTES N44). The multihost scripts refuse to run when any host is
  more than 1 s off (NOTES N50).
- **Verification strategy.** Coordinators default to `VERIFY_STRATEGY=optimistic`:
  combine t shares, verify only the final signature, and verify shares individually
  only after a failed combine. A signer whose share fails verification becomes a
  *suspect* for `BREAKER_SUSPECT_COOLDOWN` (default 10m): its shares are verified
  before they are combined. `BREAKER_FALLBACK_AFTER` (default 3) failed combines within
  `BREAKER_FALLBACK_WINDOW` (1m) switch to strict for `BREAKER_FALLBACK_COOLDOWN`
  (10m). `strict` verifies shares until t are valid. Both strategies verify the final
  signature against the group key before returning a token (NOTES N65).
- **Timing logs.** nginx writes one JSON timing line per request, and the coordinators'
  `signed`/`sign failed` lines carry `nginx_request_id`, `nginx_to_coordinator_ms`,
  `queued_before_sign_ms` and `incoming_deadline_ms`; `benchmark/single/timing-breakdown.sh`
  joins them (NOTES N64).
- **Signer capacity.** Each signer computes at most `SIGNER_MAX_CONCURRENT` shares at once
  (default: CPU count). Further requests wait only while they can still meet the
  coordinator's deadline (sent in `X-Frost-Deadline-Ms`) and fewer than `SIGNER_MAX_QUEUE`
  (default 64) are waiting; otherwise they get `503` immediately (NOTES N46, N48).
- **Who can call Sign.** Only root on the coordinator host / control-plane node (the socket
  directory is `root:root 0700`) and nginx holding the `lb` key. Nothing is published to the
  host network (docs/THREAT_MODEL.md §1).
- **Errors are generic by design.** A token requester sees only `token signing failed: …`;
  which signer refused and why is in the coordinator log and the signers' audit logs (N33).

## Licensing and patents

- The threshold RSA dependency is **`github.com/niclabs/tcrsa` v0.0.5, used under the MIT
  License** (see [NOTICE](NOTICE)). It is pinned to exactly that version.
- **Later upstream versions of tcrsa reference US patent 10735188** in their license.
- This repository is an **open-source research prototype**, licensed under the
  **Apache License 2.0** ([LICENSE](LICENSE)). Apache-2.0 is compatible with tcrsa v0.0.5's MIT
  license and with every other dependency (BSD-3-Clause, Apache-2.0, MIT).
- **Any commercial use requires independent legal review.**

## About the git history

- **Why the names say "frost".** The project started as a prototype that used FROST, a
  threshold scheme that makes Schnorr signatures. That design could not work: Kubernetes
  verifies only RS256/ES256/ES384/ES512, and the prototype in fact signed every token with
  a single coordinator-held ECDSA key. It is kept, excluded from every build, in
  [legacy/frost/](legacy/frost/README.md), with its original README.
- **Burned keys.** The early history contains development keys and secrets of that
  abandoned prototype. They were committed to this public repository and are **burned**.
  Per the owner they were never used outside this repository, and nothing in the rebuilt
  system uses them: every key the system uses is generated fresh. The author's employer
  name appears in the history of one audit-instructions commit (0785797) and nowhere in
  the current tree. See [reports/HISTORY_PURGE.md](reports/HISTORY_PURGE.md), which also
  records the plan for an anonymized, history-free snapshot for double-blind review.
- **Rewritten history.** On 2026-10-11 the history was rewritten with `git filter-repo`
  for one purpose only: to remove co-author trailer lines from commit messages. File
  contents, authors, committers, dates and commit order are unchanged, but 137 commits got
  new hashes. The burned keys are still in the history; the rewrite did not remove them.
  [reports/COMMIT_MAP.md](reports/COMMIT_MAP.md) maps every old hash to its new one.
  Results folders under `benchmark/results/` and raw logs keep the old short hashes in
  their names; use the map to translate them.
