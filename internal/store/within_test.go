package store

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWithinResolvesLinks proves a path counts as inside a directory when it
// is written inside, when it does not exist yet, and when a link from outside
// leads into it; a sibling with the same name prefix is outside. The mutated
// field is the file path. Negative control: dropping the link resolution
// turns the link case red.
func TestWithinResolvesLinks(t *testing.T) {
	root := t.TempDir()
	tree := filepath.Join(root, "repo")
	outside := filepath.Join(root, "repo-data")
	for _, d := range []string{tree, outside} {
		if err := os.Mkdir(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(outside, "into")
	if err := os.Symlink(tree, link); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		file string
		want bool
	}{
		{"inside, not yet created", filepath.Join(tree, "sub", "e.sql"), true},
		{"the tree itself", tree, true},
		{"through a link", filepath.Join(link, "e.sql"), true},
		{"sibling with the same prefix", filepath.Join(outside, "e.sql"), false},
	}
	for _, c := range cases {
		got, err := Within(c.file, tree)
		if err != nil || got != c.want {
			t.Errorf("%s: got %v, %v; want %v", c.name, got, err, c.want)
		}
	}
}
