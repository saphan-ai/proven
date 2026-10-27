package change

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A symlink in either tree is refused, its target is never returned, and the
// error names the link.
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

// A regular file that exists only in change/ is returned once, with its path
// and its content.
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

// Changed drops a file whose two sides are equal and keeps an edit, an addition and a deletion, an empty added file
// included; a change in which nothing differs is returned whole.
func TestChangedKeepsOnlyFilesThatDiffer(t *testing.T) {
	files := []File{
		{Path: "same", Base: []byte("x"), Change: []byte("x")},
		{Path: "edit", Base: []byte("x"), Change: []byte("y")},
		{Path: "added-empty", Base: nil, Change: []byte{}},
		{Path: "gone", Base: []byte("x"), Change: nil},
	}
	var got []string
	for _, f := range Changed(files) {
		got = append(got, f.Path)
	}
	if strings.Join(got, ",") != "edit,added-empty,gone" {
		t.Fatalf("%v", got)
	}
	if same := files[:1]; len(Changed(same)) != 1 {
		t.Fatal("a change with no differing file was emptied")
	}
}
