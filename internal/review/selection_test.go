package review_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestExcludeAllDropsFindings proves that excluding every path leaves no finding, and that the trace
// names the paths the selection left out, so a reader sees why the review is empty.
func TestExcludeAllDropsFindings(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target:  fixturePath(t),
		Exclude: []string{"*"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("findings %#v", res.Findings)
	}
	ev := traceEvents(t, res.TracePath, "selection")
	if len(ev) != 1 || !strings.Contains(ev[0].Text, "billing/discount.go") || !strings.Contains(ev[0].Text, "go.mod") {
		t.Fatalf("selection trace %#v", ev)
	}
}

// TestExcludeWinsOverInclude proves that a path both Include and Exclude match is excluded, and that
// the other paths stay in the review.
func TestExcludeWinsOverInclude(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target:  fixturePath(t),
		Include: []string{"*"},
		Exclude: []string{"go.mod"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := find(res, "F1"); !ok {
		t.Fatal("F1 is on billing/discount.go and the exclude does not name it")
	}
	if _, ok := find(res, "F2"); ok {
		t.Fatal("F2 is on go.mod and exclude wins")
	}
	for _, f := range res.Findings {
		if f.Location.Path == "go.mod" {
			t.Fatalf("go.mod finding %s survived the exclude", f.ID)
		}
	}
}

// TestIncludeKeepsOnlyMatchingFiles proves that Include keeps only the findings on matching paths.
func TestIncludeKeepsOnlyMatchingFiles(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target:  fixturePath(t),
		Include: []string{"go.mod"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := find(res, "F2"); !ok {
		t.Fatal("F2 missing")
	}
	if _, ok := find(res, "F1"); ok {
		t.Fatal("F1 is outside the include")
	}
}

// TestBadGlobIsFatal proves that a malformed glob stops the review with an error that quotes it, rather
// than matching nothing.
func TestBadGlobIsFatal(t *testing.T) {
	rev := openSample(t, t.TempDir())
	_, err := rev.Review(context.Background(), review.Request{
		Target:  fixturePath(t),
		Include: []string{"["},
	})
	if err == nil || !strings.Contains(err.Error(), "[") {
		t.Fatal(err)
	}
}

// TestMissingBaselineNamesTheCommit proves that a baseline commit that does not exist is an error naming it.
func TestMissingBaselineNamesTheCommit(t *testing.T) {
	rev := openSample(t, t.TempDir())
	_, err := rev.Review(context.Background(), review.Request{
		Target:         fixturePath(t),
		BaselineCommit: "does-not-exist",
	})
	if err == nil || !strings.Contains(err.Error(), "does-not-exist") {
		t.Fatal(err)
	}
}

// TestBaselineKeepsOnlyTheChangedPath proves that with a baseline, only the paths the commits changed are
// reviewed: a finding on an unchanged file is not reported.
func TestBaselineKeepsOnlyTheChangedPath(t *testing.T) {
	dir := copyFixtureTree(t)
	gitRun(t, dir, "init")
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-m", "[change] git: base")
	base := strings.TrimSpace(string(gitRun(t, dir, "rev-parse", "HEAD")))
	if err := os.MkdirAll(filepath.Join(dir, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "billing", "discount.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "billing/discount.go")
	gitRun(t, dir, "commit", "-m", "[change] git: head")
	head := strings.TrimSpace(string(gitRun(t, dir, "rev-parse", "HEAD")))

	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{
		Target:         dir,
		BaselineCommit: base,
		HeadCommit:     head,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := find(res, "F1"); !ok {
		t.Fatal("F1 is on the file the second commit added")
	}
	if _, ok := find(res, "F2"); ok {
		t.Fatal("F2 is on go.mod, which the second commit did not change")
	}
}

// copyFixtureTree copies the discount fixture into a temporary directory a test may change.
func copyFixtureTree(t *testing.T) string {
	t.Helper()
	src := fixturePath(t)
	dst := filepath.Join(t.TempDir(), "fixture")
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

// gitRun runs git in dir with a fixed identity, no signing and a fixed default branch, so the test does
// not depend on the machine's git settings.
func gitRun(t *testing.T, dir string, args ...string) []byte {
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
	return out
}
