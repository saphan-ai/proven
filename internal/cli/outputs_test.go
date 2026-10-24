package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// outputFlag is one way to name an output file on the command line.
type outputFlag struct {
	name string // the flag as the refusal must print it
	arg  string // the argument prefix that sets it; the path is appended
}

var outputFlags = []outputFlag{
	{"-o", "-o="},
	{"--text-output", "--text-output="},
	{"--json-output", "--json-output="},
	{"--sarif-output", "--sarif-output="},
}

// TestCollidingOutputPathsAreRefused proves that two outputs naming one file
// are refused at load with exit 7, naming both flags and the path (ADR-0013),
// and that the review never runs. Every pair of the four flags is tried, and
// the second flag spells the path with a ".." through a directory that
// exists, so the two name one file. The mutated field is the path of the
// second flag. With the collision check removed every row returns 0.
func TestCollidingOutputPathsAreRefused(t *testing.T) {
	for i := range outputFlags {
		for j := i + 1; j < len(outputFlags); j++ {
			first, second := outputFlags[i], outputFlags[j]
			t.Run(first.name+" with "+second.name, func(t *testing.T) {
				dir := t.TempDir()
				if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, "report")
				alias := dir + string(filepath.Separator) + "sub" + string(filepath.Separator) + ".." + string(filepath.Separator) + "report"
				rev := &captureReviewer{}
				var stdout, stderr bytes.Buffer
				code := Run([]string{"scan", first.arg + path, second.arg + alias}, emptyEnv, &stdout, &stderr, openCapture(rev))
				if code != exitConfig {
					t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
				}
				msg := stderr.String()
				for _, want := range []string{first.name, second.name, path} {
					if !strings.Contains(msg, want) {
						t.Errorf("refusal does not name %q: %s", want, msg)
					}
				}
				if _, err := os.Stat(path); err == nil {
					t.Error("a file was written despite the refusal")
				}
				if rev.req.Target != "" {
					t.Error("the review ran despite the refusal")
				}
			})
		}
	}
}

// TestOutputsThroughASymlinkedDirectoryAreRefused proves that two outputs
// reaching one file through a directory link are refused with exit 7 and
// write nothing: -o real/r.json and --sarif-output link/r.json, link -> real.
// Before the fix this returned 0 and the SARIF body replaced the JSON body.
// The mutated field is the directory of the second path: a link, not the real
// directory. With the symlink resolution removed the row is red.
func TestOutputsThroughASymlinkedDirectoryAreRefused(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(dir, "link")); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	rev := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--json", "-o", filepath.Join(real, "r.json"), "--sarif-output", filepath.Join(dir, "link", "r.json")}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitConfig {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	for _, want := range []string{"-o", "--sarif-output", "r.json"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("refusal does not name %q: %s", want, stderr.String())
		}
	}
	if _, err := os.Stat(filepath.Join(real, "r.json")); err == nil {
		t.Error("a file was written despite the refusal")
	}
	if rev.req.Target != "" {
		t.Error("the review ran despite the refusal")
	}
}

// TestSymlinkedFileAndHardLinkAreRefused proves the os.SameFile half: two
// existing outputs that are one file under different names are refused. The
// mutated field is the second path: a link to the file, or a hard link to it.
func TestSymlinkedFileAndHardLinkAreRefused(t *testing.T) {
	for _, kind := range []string{"symlink", "hard link"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			first := filepath.Join(dir, "first.json")
			second := filepath.Join(dir, "second.json")
			if err := os.WriteFile(first, []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
			var err error
			if kind == "symlink" {
				err = os.Symlink(first, second)
			} else {
				err = os.Link(first, second)
			}
			if err != nil {
				t.Skipf("%s is not available here: %v", kind, err)
			}
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--json-output", first, "--sarif-output", second}, emptyEnv, &stdout, &stderr, openCapture(&captureReviewer{}))
			if code != exitConfig {
				t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
			}
		})
	}
}

// TestRelativeAndAbsoluteSpellingsAreOneFile proves that "r.json" and its
// absolute spelling are refused as one file. The mutated field is the second
// path: absolute where the first is relative. With filepath.Abs replaced by
// filepath.Clean the two keys differ and the row is red.
func TestRelativeAndAbsoluteSpellingsAreOneFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	rev := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "-o", "r.json", "--sarif-output", filepath.Join(dir, "r.json")}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitConfig {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	if rev.req.Target != "" {
		t.Error("the review ran despite the refusal")
	}
}

// TestCollisionReproduction is the reproduction of the reported defect:
// --json -o X --sarif-output X returned 0 and left the SARIF body in X.
// It proves the refusal, and that X was not written.
func TestCollisionReproduction(t *testing.T) {
	x := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--json", "-o", x, "--sarif-output", x}, emptyEnv, &stdout, &stderr, openCapture(&captureReviewer{}))
	if code != exitConfig {
		t.Fatalf("exit %d, want %d (before the fix this returned 0); stderr %s", code, exitConfig, stderr.String())
	}
	if _, err := os.Stat(x); err == nil {
		t.Fatal("X was written despite the refusal")
	}
}

// TestDistinctOutputPathsAreNotCollisions proves the check refuses only the
// same file: four different paths are all written, with exit 0.
func TestDistinctOutputPathsAreNotCollisions(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	args := []string{"scan", "--json"}
	for _, f := range outputFlags {
		p := filepath.Join(dir, strings.TrimLeft(f.name, "-"))
		paths = append(paths, p)
		args = append(args, f.arg+p)
	}
	var stdout, stderr bytes.Buffer
	code := Run(args, emptyEnv, &stdout, &stderr, openCapture(&captureReviewer{}))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("not written: %v", err)
		}
	}
}

// TestOutputWriteFailureExits2 proves that a file that cannot be written is
// exit 2 (Proven could not report), for -o and for each extra, and that the
// error names the path. The mutated field is the directory of the path: it
// does not exist.
func TestOutputWriteFailureExits2(t *testing.T) {
	for _, f := range outputFlags {
		t.Run(f.name, func(t *testing.T) {
			bad := filepath.Join(t.TempDir(), "missing-dir", "out")
			var stdout, stderr bytes.Buffer
			rev := &captureReviewer{res: review.Result{}}
			code := Run([]string{"scan", f.arg + bad}, emptyEnv, &stdout, &stderr, openCapture(rev))
			if code != exitFatal {
				t.Fatalf("exit %d, want %d; stderr %s", code, exitFatal, stderr.String())
			}
			if !strings.Contains(stderr.String(), "missing-dir") {
				t.Fatalf("error does not name the path: %s", stderr.String())
			}
		})
	}
}

// funcReviewer runs a function as the review, so a test can change the disk
// between the load checks and the writes.
type funcReviewer func() review.Result

func (f funcReviewer) Review(context.Context, review.Request) (review.Result, error) {
	return f(), nil
}

// TestDanglingAliasOutputsAreRefused proves that an output named through a
// symlink whose target does not exist yet is the same file as the output named
// by the target: exit 7, both flags named, no review, nothing written. Every
// ordered pair of the four flags is tried, with the alias on either side. The
// mutated field is the path of the alias flag: a link to the other output's
// path, still absent. Without the Readlink resolution every row returns 0 and
// the later format replaces the earlier one.
func TestDanglingAliasOutputsAreRefused(t *testing.T) {
	for _, first := range outputFlags {
		for _, second := range outputFlags {
			if first == second {
				continue
			}
			t.Run(first.name+" actual, "+second.name+" alias", func(t *testing.T) {
				dir := t.TempDir()
				actual := filepath.Join(dir, "actual.json")
				alias := filepath.Join(dir, "alias.json")
				if err := os.Symlink("actual.json", alias); err != nil {
					t.Skipf("symlinks are not available here: %v", err)
				}
				rev := &captureReviewer{}
				var stdout, stderr bytes.Buffer
				code := Run([]string{"scan", first.arg + actual, second.arg + alias}, emptyEnv, &stdout, &stderr, openCapture(rev))
				if code != exitConfig {
					t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
				}
				for _, want := range []string{first.name, second.name} {
					if !strings.Contains(stderr.String(), want) {
						t.Errorf("refusal does not name %q: %s", want, stderr.String())
					}
				}
				if _, err := os.Stat(actual); err == nil {
					t.Error("a file was written despite the refusal")
				}
				if rev.req.Target != "" {
					t.Error("the review ran despite the refusal")
				}
			})
		}
	}
}

// TestDanglingAliasChainsAreFollowed proves that the resolution follows a
// chain of links, with a relative target taken against the directory of the
// link that holds it: sub/b.json -> ../a.json -> actual.json, none present.
// The mutated field is the alias: two hops, one of them in another directory.
func TestDanglingAliasChainsAreFollowed(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("actual.json", filepath.Join(dir, "a.json")); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	if err := os.Symlink(filepath.Join("..", "a.json"), filepath.Join(dir, "sub", "b.json")); err != nil {
		t.Fatal(err)
	}
	rev := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--json-output", filepath.Join(dir, "actual.json"), "--sarif-output", filepath.Join(dir, "sub", "b.json")}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitConfig {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	if rev.req.Target != "" {
		t.Error("the review ran despite the refusal")
	}
}

// TestLinkLoopAndLongChainAreRefusedAtLoad proves that an output that is a
// link loop, or a chain longer than the bound, is refused at load with exit 7
// naming its flag, and that the review does not run. The mutated field is the
// output path: a link to itself, or a chain of more than the bound.
func TestLinkLoopAndLongChainAreRefusedAtLoad(t *testing.T) {
	dir := t.TempDir()
	loop := filepath.Join(dir, "loop.json")
	if err := os.Symlink("loop.json", loop); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	prev := "end.json"
	for i := 0; i <= maxOutputLinkHops; i++ {
		name := "hop" + strings.Repeat("x", i) + ".json"
		if err := os.Symlink(prev, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
		prev = name
	}
	for name, path := range map[string]string{"loop": loop, "long chain": filepath.Join(dir, prev)} {
		t.Run(name, func(t *testing.T) {
			rev := &captureReviewer{}
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--sarif-output", path}, emptyEnv, &stdout, &stderr, openCapture(rev))
			if code != exitConfig {
				t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
			}
			if !strings.Contains(stderr.String(), "--sarif-output") {
				t.Errorf("refusal does not name the flag: %s", stderr.String())
			}
			if rev.req.Target != "" {
				t.Error("the review ran despite the refusal")
			}
		})
	}
}

// TestAliasCreatedAfterLoadIsRefusedAtWrite proves the second line of
// defence: a link that appears after the load checks passed, made here by the
// review itself, is caught when the second output is about to be written. Exit
// 2, both flags named, and the first output is left as it was written. The
// mutated field is the disk between load and write: alias.json becomes a link
// to the first output. Without the write-time check the SARIF body replaces
// the JSON body and the exit is 0.
func TestAliasCreatedAfterLoadIsRefusedAtWrite(t *testing.T) {
	dir := t.TempDir()
	actual := filepath.Join(dir, "actual.json")
	alias := filepath.Join(dir, "alias.json")
	rev := funcReviewer(func() review.Result {
		if err := os.Symlink(actual, alias); err != nil {
			t.Errorf("symlinks are not available here: %v", err)
		}
		return review.Result{}
	})
	var stdout, stderr bytes.Buffer
	open := func(review.Options) (review.Reviewer, error) { return rev, nil }
	code := Run([]string{"scan", "--json", "-o", actual, "--sarif-output", alias}, emptyEnv, &stdout, &stderr, open)
	if code != exitFatal {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitFatal, stderr.String())
	}
	for _, want := range []string{"-o", "--sarif-output"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("refusal does not name %q: %s", want, stderr.String())
		}
	}
	body, err := os.ReadFile(actual)
	if err != nil {
		t.Fatalf("the first output is gone: %v", err)
	}
	if strings.Contains(string(body), "$schema") {
		t.Errorf("the first output was replaced by the SARIF body: %s", body)
	}
}
