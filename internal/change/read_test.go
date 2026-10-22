package change

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadRefusesASymlink(t *testing.T) {
	const marker = "outside-the-tree"
	for _, side := range []string{"base", "change"} {
		t.Run(side, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "base"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(root, "change"), 0o755); err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(t.TempDir(), "outside.txt")
			if err := os.WriteFile(outside, []byte(marker), 0o644); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(root, side, "x")
			if err := os.Symlink(outside, link); err != nil {
				t.Fatal(err)
			}
			files, err := Read(root)
			if err == nil {
				for _, f := range files {
					if strings.Contains(string(f.Base), marker) || strings.Contains(string(f.Change), marker) {
						t.Fatalf("read %s; the symlink target was returned", f.Path)
					}
				}
				t.Fatal("symlink was accepted")
			}
			if !strings.Contains(err.Error(), link) {
				t.Fatalf("error %q does not name %s", err, link)
			}
		})
	}
}

func TestReadReturnsARegularFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "change"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "change", "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "a.txt" || string(files[0].Change) != "hello" {
		t.Fatalf("%#v", files)
	}
}
