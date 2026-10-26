//go:build windows

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestWindowsSingleNameTraceIsUsed proves an existing trace.jsonl with one
// name is not refused: the count read on Windows is one.
// Unplug: linkCount returns 2. This load exits 7.
// Mutated field: the count linkCount reports for trace.jsonl.
func TestWindowsSingleNameTraceIsUsed(t *testing.T) {
	target := t.TempDir()
	logDir := t.TempDir()
	path := filepath.Join(logDir, traceFile)
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := linkCount(path)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("link count %d, want 1", n)
	}
	var stdout, stderr bytes.Buffer
	opened := false
	code := Run([]string{"scan", target}, envOf(map[string]string{EnvLogDir: logDir}), &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		opened = true
		return &captureReviewer{}, nil
	})
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !opened {
		t.Fatal("reviewer did not open for a trace file with one name")
	}
}

// TestWindowsHardLinkedTraceIsRefused proves a trace.jsonl that is a second
// name of a file inside the review target is refused at load with exit 7, and
// the target file is not written.
// Unplug: linkCount returns 1. The load opens the reviewer and appends.
// Mutated field: the count linkCount reports for trace.jsonl.
func TestWindowsHardLinkedTraceIsRefused(t *testing.T) {
	target := t.TempDir()
	logDir := t.TempDir()
	inside := filepath.Join(target, "kept.txt")
	if err := os.WriteFile(inside, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(logDir, traceFile)
	if err := os.Link(inside, path); err != nil {
		t.Skipf("hard link not available here: %v", err)
	}
	n, err := linkCount(path)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("link count %d, want 2", n)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", target}, envOf(map[string]string{EnvLogDir: logDir}), &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		t.Fatal("reviewer opened for a trace file with two names")
		return nil, nil
	})
	if code != exitConfig {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	if b, _ := os.ReadFile(inside); string(b) != "kept\n" {
		t.Fatalf("the target file was written: %q", b)
	}
}
