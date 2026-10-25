//go:build !windows

package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
)

// TestMaterializedTreesAreNotGroupWritableUnderAPermissiveUmask proves the
// modes of the materialized trees are set explicitly and do not follow the
// process umask: under umask 002 no tree, directory or file is writable by
// group or others. The mutated field is the umask (002 instead of 022).
func TestMaterializedTreesAreNotGroupWritableUnderAPermissiveUmask(t *testing.T) {
	repo := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@example.org"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	if err := os.MkdirAll(filepath.Join(repo, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "sub/a.go"} {
		if err := os.WriteFile(filepath.Join(repo, name), []byte("package m\n"), 0o664); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Join(repo, name), 0o664); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-q", "-m", "[readable-trees] test: first")
	head, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	commit := string(head[:len(head)-1])

	old := syscall.Umask(0o002)
	defer syscall.Umask(old)
	p := &Pipeline{workDir: t.TempDir()}
	fx, _, remove, err := p.materialize(Request{Target: repo, BaselineCommit: commit, HeadCommit: commit})
	if err != nil {
		t.Fatal(err)
	}
	defer remove()
	err = filepath.WalkDir(fx.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		perm := info.Mode().Perm()
		if perm&0o022 != 0 {
			t.Errorf("%s mode %o is writable by group or others", path, perm)
		}
		if perm&0o004 == 0 || (info.IsDir() && perm&0o001 == 0) {
			t.Errorf("%s mode %o is not readable by others", path, perm)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
