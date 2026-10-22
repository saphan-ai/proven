package change

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBetweenTwoCommits(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")
	first := rev(t, dir)
	write(t, filepath.Join(dir, "b.txt"), "b")
	git(t, dir, "add", "b.txt")
	git(t, dir, "commit", "-m", "[change] git: second")
	second := rev(t, dir)

	got, err := Between(dir, first, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "b.txt" {
		t.Fatalf("files %#v", got)
	}
}

func TestBetweenWorkingTreeIncludesEditsAndUntracked(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")
	base := rev(t, dir)
	write(t, filepath.Join(dir, "a.txt"), "changed")
	write(t, filepath.Join(dir, "c.txt"), "new")

	got, err := Between(dir, base, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "a.txt" || got[1] != "c.txt" {
		t.Fatalf("files %#v", got)
	}
}

func TestBetweenMissingBaselineNamesTheCommit(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")

	_, err := Between(dir, "does-not-exist", "")
	if err == nil || !strings.Contains(err.Error(), "does-not-exist") {
		t.Fatal(err)
	}
}

func TestBetweenMissingHeadNamesTheCommit(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")
	base := rev(t, dir)

	_, err := Between(dir, base, "missing-head")
	if err == nil || !strings.Contains(err.Error(), "missing-head") {
		t.Fatal(err)
	}
}

func TestSelectExcludeWinsAndMatchesBaseName(t *testing.T) {
	files := []string{"billing/discount.go", "go.mod"}
	selected, out, err := Select(files, []string{"*"}, []string{"go.mod"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0] != "billing/discount.go" {
		t.Fatalf("selected %#v", selected)
	}
	if len(out) != 1 || out[0] != "go.mod" {
		t.Fatalf("out %#v", out)
	}

	selected, out, err = Select(files, nil, []string{"*"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 0 || len(out) != 2 {
		t.Fatalf("selected %#v out %#v", selected, out)
	}
}

func TestSelectBadGlob(t *testing.T) {
	_, _, err := Select([]string{"a.go"}, []string{"["}, nil)
	if err == nil || !strings.Contains(err.Error(), "[") {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{
		"-c", "user.name=Proven Test",
		"-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
		"-c", "init.defaultBranch=main",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func rev(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(string(runGit(t, dir, "rev-parse", "HEAD")))
}

func runGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMessagesExcludesTheBaselineCommit(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, "a.txt"), "a")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "[fixture] change: first\n\nalpha-marker")
	first := rev(t, dir)
	write(t, filepath.Join(dir, "b.txt"), "b")
	git(t, dir, "add", "b.txt")
	git(t, dir, "commit", "-m", "[fixture] change: second\n\nModel-of-record: Claude:claude-3-opus")
	second := rev(t, dir)

	got, err := Messages(dir, first, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.Contains(got[0], "Model-of-record: Claude:claude-3-opus") {
		t.Fatalf("%#v", got)
	}
	if strings.Contains(got[0], "alpha-marker") {
		t.Fatal("baseline message was included")
	}
}

func TestBetweenWorkingTreeInSubdirectoryKeepsUntracked(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "sub", "a.txt"), "a")
	git(t, dir, "add", "sub/a.txt")
	git(t, dir, "commit", "-m", "[change] git: first")
	base := rev(t, dir)
	write(t, filepath.Join(dir, "sub", "c.txt"), "new")
	write(t, filepath.Join(dir, "top.txt"), "outside the target")

	got, err := Between(filepath.Join(dir, "sub"), base, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "c.txt" {
		t.Fatalf("files %#v", got)
	}
}
