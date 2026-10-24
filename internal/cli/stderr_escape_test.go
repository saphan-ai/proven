package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// stderrFixture copies the discount fixture and adds one file named name to
// its change tree, through add.
func stderrFixture(t *testing.T, name string, add func(path string) error) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "fixture")
	if err := os.CopyFS(dir, os.DirFS(filepath.Join(moduleRoot(t), "testdata", "fixture", "discount"))); err != nil {
		t.Fatal(err)
	}
	if err := add(filepath.Join(dir, "change", name)); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Property: a file name from the change, carried into an error sentence,
// reaches stderr with every control byte shown and none raw.
func TestStderrCarriesNoRawControlByteFromTheChange(t *testing.T) {
	const name = "x\x1b[2Jy"
	cases := []struct {
		label string
		bound int64
		add   func(path string) error
	}{
		{"symlink named in the error", 0, func(p string) error { return os.Symlink("go.mod", p) }},
		// Sample mode keeps the default bounds, so the file bound is set
		// below the one oversize file; every fixture file is under 2 KiB.
		{"over-bound file named in the error", 2048, func(p string) error { return os.WriteFile(p, make([]byte, 4096), 0o644) }},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			dir := stderrFixture(t, name, c.add)
			if c.bound > 0 {
				previous := change.UseLimits(change.Limits{File: c.bound})
				t.Cleanup(func() { change.UseLimits(previous) })
			}
			t.Chdir(t.TempDir())
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--sample", dir}, emptyEnv, &stdout, &stderr, nil)
			if code != exitFatal {
				t.Fatalf("exit %d, want %d\nstderr: %q", code, exitFatal, stderr.String())
			}
			if !bytes.Contains(stderr.Bytes(), []byte(`x\x1b[2Jy`)) {
				t.Errorf("stderr does not show the name with its control byte shown: %q", stderr.String())
			}
			if i := bytes.IndexByte(stderr.Bytes(), 0x1b); i >= 0 {
				t.Errorf("raw byte 0x1b at offset %d of stderr: %q", i, stderr.String())
			}
		})
	}
}
