package review_test

import (
	"os/exec"
	"slices"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/review"
)

// TestSampleOpenInLocalModeKeepsGitOffline: a sample review with a baseline commit still runs git
// (proven.yaml at the baseline, the commit log). In PROVEN_NETWORK=local that git must not be able to
// fetch, so after Open in sample mode the review's git must carry the offline switches.
// Mutated field: Options.Sample (true; the live path already sets the switch).
func TestSampleOpenInLocalModeKeepsGitOffline(t *testing.T) {
	prev := change.UseOffline(false)
	t.Cleanup(func() { change.UseOffline(prev) })
	env := map[string]string{"PROVEN_NETWORK": "local"}
	r, err := review.Open(review.Options{Sample: true, LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
	if err != nil {
		t.Fatal(err)
	}
	if args := r.(interface{ Git(...string) *exec.Cmd }).Git("status").Args; !slices.Contains(args, "protocol.allow=never") {
		t.Fatalf("sample mode in PROVEN_NETWORK=local builds a git that can fetch: %v", args)
	}
}
