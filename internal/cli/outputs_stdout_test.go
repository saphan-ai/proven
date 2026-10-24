package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestStdoutRedirectedOntoAnOutputIsRefused proves that standard output, when
// it is a regular file, is compared with every output file: the shell has
// already created the file (proven scan --sarif --json-output r.json > r.json),
// so one body would silently replace the other. The mutated field is the
// output path: it names the file standard output is open on, directly or
// through a symlink.
func TestStdoutRedirectedOntoAnOutputIsRefused(t *testing.T) {
	for _, f := range outputFlags {
		for _, via := range []string{"direct", "symlink"} {
			t.Run(f.name+"/"+via, func(t *testing.T) {
				dir := t.TempDir()
				x := filepath.Join(dir, "x.json")
				out, err := os.Create(x)
				if err != nil {
					t.Fatal(err)
				}
				defer out.Close()
				path := x
				if via == "symlink" {
					path = filepath.Join(dir, "alias.json")
					symlink(t, x, path)
				}
				rev := &captureReviewer{}
				var stderr bytes.Buffer
				code := Run([]string{"scan", "--sarif", f.arg + path}, emptyEnv, out, &stderr, openCapture(rev))
				if code != exitConfig {
					t.Fatalf("exit %d, want %d (before the fix this returned 0); stderr %s", code, exitConfig, stderr.String())
				}
				for _, want := range []string{"standard output", f.name} {
					if !strings.Contains(stderr.String(), want) {
						t.Errorf("refusal does not name %q: %s", want, stderr.String())
					}
				}
				if rev.req.Target != "" {
					t.Error("the review ran despite the refusal")
				}
				if info, err := os.Stat(x); err != nil || info.Size() != 0 {
					t.Errorf("the file was written despite the refusal: %v", err)
				}
			})
		}
	}
}

// TestStdoutOnAnotherFileOrAWriterIsNotACollision proves the check refuses
// only the same file: standard output on a different regular file, and a
// plain writer (a pipe or a terminal is not a regular file), both write with
// exit 0.
func TestStdoutOnAnotherFileOrAWriterIsNotACollision(t *testing.T) {
	dir := t.TempDir()
	other, err := os.Create(filepath.Join(dir, "stdout.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	var buf bytes.Buffer
	for name, w := range map[string]io.Writer{"other file": other, "writer": &buf} {
		var stderr bytes.Buffer
		code := Run([]string{"scan", "--sarif", "--json-output=" + filepath.Join(dir, name+".json")}, emptyEnv, w, &stderr, openCapture(&captureReviewer{}))
		if code != exitOK {
			t.Errorf("%s: exit %d stderr %s", name, code, stderr.String())
		}
	}
}

// TestStdoutAliasCreatedAfterLoadIsRefusedAtWrite proves the write-time
// guard counts standard output: the output path is a link made while the
// review ran, so the load check saw nothing. The mutated field is the stdout
// entry of the written list. The run stops with exit 2 and standard output
// keeps what it was written.
func TestStdoutAliasCreatedAfterLoadIsRefusedAtWrite(t *testing.T) {
	dir := t.TempDir()
	x := filepath.Join(dir, "x.json")
	out, err := os.Create(x)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	link := filepath.Join(dir, "later.json")
	rev := funcReviewer(func() review.Result {
		symlink(t, x, link)
		return review.Result{}
	})
	var stderr bytes.Buffer
	open := func(review.Options) (review.Reviewer, error) { return rev, nil }
	code := Run([]string{"scan", "--sarif", "--json-output=" + link}, emptyEnv, out, &stderr, open)
	if code != exitFatal {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitFatal, stderr.String())
	}
	if !strings.Contains(stderr.String(), "standard output") {
		t.Errorf("error does not name standard output: %s", stderr.String())
	}
	body, _ := os.ReadFile(x)
	if !strings.Contains(string(body), "$schema") {
		t.Errorf("standard output was replaced: %s", body)
	}
}
