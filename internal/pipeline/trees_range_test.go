package pipeline

import (
	"github.com/saphan-ai/proven/internal/change"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Property: with a baseline commit, the git range decides the trees. A
// repository that has its own base/ and change/ directories does not choose
// what the agents read.
func TestBaselineCommitDecidesTheTreesNotTheRepositoryDirectories(t *testing.T) {
	repo := t.TempDir()
	rangeGit(t, repo, "init")
	rangeWrite(t, filepath.Join(repo, "app.go"), "package app\n")
	rangeGit(t, repo, "add", "app.go")
	rangeGit(t, repo, "commit", "-m", "[range] repo: base")
	baseline := strings.TrimSpace(rangeGit(t, repo, "rev-parse", "HEAD"))
	rangeWrite(t, filepath.Join(repo, "app.go"), "package app\n\nvar Changed = true\n")
	rangeWrite(t, filepath.Join(repo, "base", "decoy.txt"), "decoy\n")
	rangeWrite(t, filepath.Join(repo, "change", "decoy.txt"), "decoy\n")
	rangeGit(t, repo, "add", ".")
	rangeGit(t, repo, "commit", "-m", "[range] repo: change with base and change directories")
	head := strings.TrimSpace(rangeGit(t, repo, "rev-parse", "HEAD"))

	p := &Pipeline{workDir: t.TempDir()}
	fx, _, cleanup, err := p.loadLiveTarget(Request{Target: repo, BaselineCommit: baseline, HeadCommit: head}, change.ProcessPolicy())
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		t.Fatal(err)
	}
	if fx.Root == repo {
		t.Fatal("a review with a baseline commit read the repository's own base/ and change/ directories")
	}
	if _, err := os.Stat(filepath.Join(fx.Root, "change", "app.go")); err != nil {
		t.Fatalf("the change tree does not hold the changed file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(fx.Root, "base", "decoy.txt")); err == nil {
		t.Fatal("the base tree holds a file that is not in the baseline commit")
	}
}

// Property: with no baseline commit, prepared base/ and change/ directories are
// still read as they are, so the demo layout keeps working.
func TestPreparedTreesAreReadWhenNoBaselineIsGiven(t *testing.T) {
	root := t.TempDir()
	rangeWrite(t, filepath.Join(root, "base", "a.txt"), "a\n")
	rangeWrite(t, filepath.Join(root, "change", "a.txt"), "b\n")
	p := &Pipeline{workDir: t.TempDir()}
	fx, _, _, err := p.loadLiveTarget(Request{Target: root}, change.ProcessPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if fx.Root != root {
		t.Fatalf("prepared trees were not used: root %s", fx.Root)
	}
}

// rangeGit runs git in dir with a fixed identity and returns its output.
func rangeGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{
		"-c", "user.name=Proven Test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// rangeWrite writes body at path and creates the parent directories.
func rangeWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
