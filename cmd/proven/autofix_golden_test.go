package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWithAutofixOffTheSampleOutputIsTheOutputBeforeTheAttemptExisted proves ADR-0050 point 6: with the
// whole-change attempt off, which is the default, the review text of two sample scenarios (one with proposed
// fixes, one with a dependency finding) is byte for byte the text the base commit printed. The golden files
// were captured by running `proven scan --sample` on those fixtures at the base commit, before any line of
// the attempt was written, with the trace line cut because it names the machine's directory.
func TestWithAutofixOffTheSampleOutputIsTheOutputBeforeTheAttemptExisted(t *testing.T) {
	// repoRoot reads the working directory, which scanWith changes, so it is read once, first, and each
	// scenario is a subtest, whose working directory is restored when it ends.
	root := repoRoot(t)
	for _, name := range []string{"fix", "osv-go-module"} {
		t.Run(name, func(t *testing.T) {
			body, _ := scanWith(t, name)
			got := body[:strings.LastIndex(body, "trace: ")]
			want, err := os.ReadFile(filepath.Join(root, "testdata", "golden", "autofix-off", name+".txt"))
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Fatalf("the output with the attempt off differs from the base commit's\n--- got\n%s\n--- want\n%s", got, want)
			}
		})
	}
}

// TestWithAutofixOnAScenarioWithNoRecordedAttemptPrintsTheSameReview proves a review that turns the attempt
// on, on a change whose sample records no answer for it, makes no claim: the review text for `osv-go-module`
// is still the base commit's. (`fix` is not used: with the attempt on, the per-finding suggestions of ADR-0042
// are replaced by it, which is the one intended difference.)
func TestWithAutofixOnAScenarioWithNoRecordedAttemptPrintsTheSameReview(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(cfg, []byte("autofix: on\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := repoRoot(t)
	body, _ := scanWith(t, "osv-go-module", "--config", cfg)
	got := body[:strings.LastIndex(body, "trace: ")]
	want, err := os.ReadFile(filepath.Join(root, "testdata", "golden", "autofix-off", "osv-go-module.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("the output with the attempt on and no recorded answer differs\n--- got\n%s\n--- want\n%s", got, want)
	}
}
