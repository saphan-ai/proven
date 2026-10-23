package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMaterializedTreesAreReadableByOthers proves the trees a live review
// writes for the container (base and change) can be read and searched by a
// user that does not own them, and are not writable by others. The mutated
// field is the mode of the trees' root and of the files: the tree root is made
// private by MkdirTemp, so without the grant in materialize this test fails.
func TestMaterializedTreesAreReadableByOthers(t *testing.T) {
	repo := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@example.org"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "go.mod")
	git("commit", "-q", "-m", "[readable-trees] test: first")
	head, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	commit := string(head[:len(head)-1])

	p := &Pipeline{workDir: t.TempDir()}
	fx, _, remove, err := p.materialize(Request{Target: repo, BaselineCommit: commit, HeadCommit: commit})
	if err != nil {
		t.Fatal(err)
	}
	defer remove()
	for _, path := range []string{
		fx.Root,
		filepath.Join(fx.Root, "base"),
		filepath.Join(fx.Root, "base", "go.mod"),
		filepath.Join(fx.Root, "change"),
		filepath.Join(fx.Root, "change", "go.mod"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		perm := info.Mode().Perm()
		if perm&0o004 == 0 || (info.IsDir() && perm&0o001 == 0) {
			t.Errorf("%s mode %o is not readable by others", path, perm)
		}
		if perm&0o022 != 0 {
			t.Errorf("%s mode %o is writable by group or others", path, perm)
		}
	}
}
