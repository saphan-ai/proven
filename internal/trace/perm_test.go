package trace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/trace"
)

// TestTraceIsReadableOnlyByItsUser proves the logs directory is created 0700
// and trace.jsonl 0600 (ADR-0013: Proven's own state is private to its user).
func TestTraceIsReadableOnlyByItsUser(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new", "logs")
	w, err := trace.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{dir: 0o700, w.Path(): 0o600} {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Mode().Perm(); got != want {
			t.Errorf("%s: mode %o, want %o", path, got, want)
		}
	}
}
