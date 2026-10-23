package trace

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOpenNarrowsAnExistingDirectoryAndFile proves Proven's own state stays
// readable only by its user when an earlier run left the logs directory at
// 0755 and the trace file at 0644: Open narrows both.
func TestOpenNarrowsAnExistingDirectoryAndFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "trace.jsonl")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	for p, want := range map[string]os.FileMode{dir: 0o700, path: 0o600} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Fatalf("%s: mode %o, want %o", p, got, want)
		}
	}
}
