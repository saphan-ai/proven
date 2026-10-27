package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestCILocalBadAllowStartsOnlyOfflineGit: a malformed allow list in local mode
// is refused by Open, after ci has read the merge base; that git call must
// already carry the offline switches. Mutated field: PROVEN_NETWORK_ALLOW.
func TestCILocalBadAllowStartsOnlyOfflineGit(t *testing.T) {
	defer change.UseOffline(change.UseOffline(false))
	rec := fakeGit(t)
	env := map[string]string{
		"PROVEN_NETWORK":       "local",
		"PROVEN_NETWORK_ALLOW": "bad/host",
		"GITHUB_BASE_REF":      "main",
		"GITHUB_SHA":           "fedcba9876543210fedcba9876543210fedcba98",
	}
	var out, stderr bytes.Buffer
	_ = Run([]string{"ci"}, func(k string) string { return env[k] }, &out, &stderr, openCapture(&captureReviewer{}))
	argv, err := os.ReadFile(filepath.Join(rec, "argv-0"))
	if err != nil {
		t.Fatalf("expected git merge-base to be observed: %v; stderr: %s", err, stderr.String())
	}
	if !strings.Contains(string(argv), "protocol.allow=never") {
		t.Fatalf("ci started git without the offline switches before the allow list was refused; argv: %s; stderr: %s", argv, stderr.String())
	}
}
