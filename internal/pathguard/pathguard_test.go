package pathguard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/pathguard"
)

// TestInsideComparesResolvedPaths proves the comparison follows links: a path
// reached through a link into the root is inside, a sibling that merely shares a
// name prefix is not, and a path that does not exist yet is placed by its nearest
// existing parent.
func TestInsideComparesResolvedPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "in"), 0o700); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir()
	link := filepath.Join(other, "link")
	if err := os.Symlink(filepath.Join(root, "in"), link); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	for name, tc := range map[string]struct {
		dir  string
		want bool
	}{
		"the root itself":           {root, true},
		"below the root":            {filepath.Join(root, "in", "x"), true},
		"not yet created":           {filepath.Join(root, "new", "deep"), true},
		"through a link":            {filepath.Join(link, "state"), true},
		"outside":                   {filepath.Join(other, "plain"), false},
		"a name starting with dots": {filepath.Join(root, "..state"), true},
	} {
		got, err := pathguard.Inside(tc.dir, root)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != tc.want {
			t.Errorf("%s: inside = %v, want %v", name, got, tc.want)
		}
	}
}

// TestResolveRefusesALinkThatDoesNotResolve proves a link to nowhere is an error
// and is never treated as a path outside the root.
func TestResolveRefusesALinkThatDoesNotResolve(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "dangling")
	if err := os.Symlink(filepath.Join(dir, "missing"), link); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	if _, err := pathguard.Resolve(filepath.Join(link, "state")); err == nil {
		t.Fatal("a path behind a dangling link was resolved")
	}
}
