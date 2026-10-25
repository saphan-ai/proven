package snapshot

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyModes(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, mode := range map[string]fs.FileMode{
		"a.txt":     0o600,
		"sub/b.txt": 0o640,
	} {
		if err := os.WriteFile(filepath.Join(src, name), []byte(name), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Join(src, name), mode); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{src, filepath.Join(src, "sub")} {
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	dst := filepath.Join(t.TempDir(), "dst")
	if err := Copy(src, dst); err != nil {
		t.Fatal(err)
	}

	err := filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		want := fs.FileMode(0o644)
		if d.IsDir() {
			want = 0o755
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s: mode %04o, want %04o", path, got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
