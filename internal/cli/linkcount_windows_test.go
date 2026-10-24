//go:build windows

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestWindowsSingleNameTraceIsUsed proves an existing trace.jsonl is not
// refused. The link count is not read on Windows, so the load treats the
// file as one name.
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
