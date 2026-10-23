package hmem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// TestStateIsReadableOnlyByItsUser proves the state and log directories this
// package creates are 0700 and the files in them 0600 (ADR-0013: Proven's own
// state is private to its user).
func TestStateIsReadableOnlyByItsUser(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "new", "state")
	logs := filepath.Join(root, "new", "logs")
	s := openAt(t, dir, logs)
	if _, err := s.Add(hmem.AddInput{Content: "a private fact", Source: "test", At: fixed}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{
		dir:                               0o700,
		logs:                              0o700,
		filepath.Join(dir, "state.json"):  0o600,
		filepath.Join(logs, "hmem.jsonl"): 0o600,
	} {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Mode().Perm(); got != want {
			t.Errorf("%s: mode %o, want %o", path, got, want)
		}
	}
}
