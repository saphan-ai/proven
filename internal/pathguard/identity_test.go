package pathguard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/pathguard"
)

// TestInsideDecidesByFileIdentityNotSpelling proves a directory spelled in
// another letter case than the root is still inside it on a file system that
// folds case, and a not-yet-created path below such a spelling is placed by its
// nearest existing ancestor. It skips, by name, on a case-sensitive file system.
func TestInsideDecidesByFileIdentityNotSpelling(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	folded := filepath.Join(base, "REPO")
	if _, err := os.Stat(folded); err != nil {
		t.Skip("the file system is case-sensitive: a differently cased name is another directory")
	}
	for name, dir := range map[string]string{
		"the root, respelled":        folded,
		"below the root, respelled":  filepath.Join(folded, "logs"),
		"not yet created, respelled": filepath.Join(folded, "new", "deep"),
	} {
		got, err := pathguard.Inside(dir, root)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !got {
			t.Errorf("%s: judged outside the root it is the same directory as", name)
		}
	}
	sibling := filepath.Join(base, "REPO2")
	if got, err := pathguard.Inside(sibling, root); err != nil || got {
		t.Errorf("a sibling sharing a prefix: inside = %v, err = %v", got, err)
	}
}
