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
		add   func(path string) error
	}{
		{"symlink named in the error", func(p string) error { return os.Symlink("go.mod", p) }},
		// Sample mode puts the default bounds back whatever the process held,
		// so the file is one byte over the default bound. It is sparse: the
		// size is checked before a byte is read.
		{"over-bound file named in the error", func(p string) error {
			f, err := os.Create(p)
			if err != nil {
				return err
			}
			if err := f.Truncate(change.DefaultLimits().File + 1); err != nil {
				f.Close()
				return err
			}
			return f.Close()
		}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			dir := stderrFixture(t, name, c.add)
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
