package cli

import (
	"bytes"
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
// the second flag spells the path unclean, so filepath.Clean and filepath.Abs
// must make the two equal. The mutated field is the path of the second flag.
// With the collision check removed every row returns 0, so every row is red.
func TestCollidingOutputPathsAreRefused(t *testing.T) {
	for i := range outputFlags {
		for j := i + 1; j < len(outputFlags); j++ {
			first, second := outputFlags[i], outputFlags[j]
			t.Run(first.name+" with "+second.name, func(t *testing.T) {
				dir := t.TempDir()
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
