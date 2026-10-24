package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSessionDirectoriesAreReadableOnlyByTheUser proves the session folders
// the conformance command creates under the temp dir are 0700 (ADR-0013:
// Proven's own state is private to its user).
func TestSessionDirectoriesAreReadableOnlyByTheUser(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	id, err := createSession("h")
	if err != nil {
		t.Fatal(err)
	}
	base := sessionBase(id)
	for _, p := range []string{filepath.Dir(base), base, filepath.Join(base, "state"), filepath.Join(base, "logs"), filepath.Join(base, "hmem")} {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Mode().Perm(); got != 0o700 {
			t.Errorf("%s: mode %o, want 700", p, got)
		}
	}
}
