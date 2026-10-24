package review_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestReviewRefusesLogsAndStateInsideTheReviewedTarget proves the property that
// the trace, the state and the work area stay outside the reviewed target: a
// review whose logs directory sits inside the target is refused and leaves no
// logs and no state there. The mutated field is LogDir (a path under Target);
// the refusal is compared with the target, not with Proven's own module.
func TestReviewRefusesLogsAndStateInsideTheReviewedTarget(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(file), "..", "..", "testdata", "fixture", "discount")
	target := t.TempDir()
	copyTargetTree(t, src, target)

	rev, err := review.Open(review.Options{Sample: true, LogDir: filepath.Join(target, "logs")})
	if err == nil {
		_, err = rev.Review(context.Background(), review.Request{Target: target})
	}
	var inside []string
	for _, name := range []string{"logs", "state"} {
		if _, serr := os.Lstat(filepath.Join(target, name)); serr == nil {
			inside = append(inside, name)
		}
	}
	if err == nil || len(inside) > 0 {
		t.Fatalf("review err %v; written inside the reviewed target: %v; want a refusal and nothing written there", err, inside)
	}
}

// TestReviewRefusesATraceInsideAnEmptyTarget proves the same property for a
// target that holds nothing: no trace file and no state directory appear in it.
func TestReviewRefusesATraceInsideAnEmptyTarget(t *testing.T) {
	target := t.TempDir()
	logs := filepath.Join(target, "logs")
	rev, err := review.Open(review.Options{Sample: true, LogDir: logs})
	if err == nil {
		_, err = rev.Review(context.Background(), review.Request{Target: target})
	}
	if err == nil {
		t.Fatal("a review with its trace and state inside the reviewed target was accepted")
	}
	for _, path := range []string{filepath.Join(logs, "trace.jsonl"), filepath.Join(target, "state")} {
		if _, serr := os.Lstat(path); serr == nil {
			t.Fatalf("%s was created inside the reviewed target", path)
		}
	}
}

// TestReviewRefusesALogsDirectoryReachedThroughALinkIntoTheTarget proves the
// target comparison resolves links: a logs path that is a link into the target
// is refused and the target stays untouched.
func TestReviewRefusesALogsDirectoryReachedThroughALinkIntoTheTarget(t *testing.T) {
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, "inner"), 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "logs")
	if err := os.Symlink(filepath.Join(target, "inner"), link); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	rev, err := review.Open(review.Options{Sample: true, LogDir: link})
	if err == nil {
		_, err = rev.Review(context.Background(), review.Request{Target: target})
	}
	if err == nil {
		t.Fatal("a logs directory that resolves inside the reviewed target was accepted")
	}
	entries, rerr := os.ReadDir(filepath.Join(target, "inner"))
	if rerr != nil || len(entries) > 0 {
		t.Fatalf("something was written inside the target: %v %v", entries, rerr)
	}
}

func copyTargetTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
