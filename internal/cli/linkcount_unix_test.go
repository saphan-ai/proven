//go:build unix

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestExistingTraceFileWithOneNameIsUsed proves a trace.jsonl that already
// exists, and has one name, is not refused. The refusal is a link count
// above one, not the fact that the file is there.
// Unplug: linkCount returns a count above one for every regular file.
// This load exits 7.
// Mutated field: the link count of trace.jsonl in PROVEN_LOG_DIR.
func TestExistingTraceFileWithOneNameIsUsed(t *testing.T) {
	target := t.TempDir()
	logDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(logDir, traceFile), nil, 0o644); err != nil {
		t.Fatal(err)
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
