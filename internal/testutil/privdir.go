package testutil

import (
	"os"
	"testing"
)

// PrivateDir returns a new t.TempDir() set to 0700. t.TempDir() creates the
// directory with mode 0777 &^ umask, so under a umask of 002 (the default for
// users with their own group, as on Ubuntu) it is 0775. The dealer refuses to
// write keys or shares into a group-writable directory (audit E-4), so every
// test that writes key, share or CA material uses this instead.
func PrivateDir(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}
