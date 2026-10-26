package review_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// sampleBoundedLine opens a sample review with the given environment, reviews the sample change, and returns
// the text of the one "bounded-steps" trace line. That line says whether the engine ran or why it did not.
func sampleBoundedLine(t *testing.T, env map[string]string) string {
	t.Helper()
	rev, err := review.Open(review.Options{
		Sample: true, LogDir: t.TempDir(),
		Env: func(k string) string { return env[k] },
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(res.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.Contains(l, `"kind":"bounded-steps"`) {
			lines = append(lines, l)
		}
	}
	if len(lines) != 1 {
		t.Fatalf("%d bounded-steps lines in the trace, want 1", len(lines))
	}
	return lines[0]
}

// TestSampleReviewRunsMarkovianByDefault proves ADR-0053 for the lab: a sample review (the lab's pull-request
// review is one) has Markovian on when nothing is set. The sample change is below the threshold, so the line
// says the change is not above it; an engine that was off would say the engine is off. Mutated field:
// PROVEN_MARKOVIAN, unset against off.
func TestSampleReviewRunsMarkovianByDefault(t *testing.T) {
	line := sampleBoundedLine(t, nil)
	if !strings.Contains(line, "not above the threshold") || strings.Contains(line, "the engine is off") {
		t.Fatalf("unset: %s", line)
	}
	line = sampleBoundedLine(t, map[string]string{"PROVEN_MARKOVIAN": "off"})
	if !strings.Contains(line, "the engine is off (PROVEN_MARKOVIAN is off)") {
		t.Fatalf("off: %s", line)
	}
}

// TestEngineSwitchesReachTheStoresThroughOpen proves the seam wires PROVEN_HMEM and PROVEN_WORLD_MODEL to
// the pipeline: with the defaults a sample review writes both stores under its logs directory, and with
// each variable set to off its store is not made. Mutated field: the one variable under test.
func TestEngineSwitchesReachTheStoresThroughOpen(t *testing.T) {
	has := func(logDir, store string) bool {
		_, err := os.Stat(logDir + "/state/" + store)
		return err == nil
	}
	for _, tc := range []struct {
		name        string
		env         map[string]string
		hmem, world bool
	}{
		{"defaults", nil, true, true},
		{"hmem off", map[string]string{"PROVEN_HMEM": "off"}, false, true},
		{"world off", map[string]string{"PROVEN_WORLD_MODEL": "off"}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logDir := t.TempDir()
			rev, err := review.Open(review.Options{Sample: true, LogDir: logDir, Env: func(k string) string { return tc.env[k] }})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)}); err != nil {
				t.Fatal(err)
			}
			if has(logDir, "hmem") != tc.hmem || has(logDir, "world") != tc.world {
				t.Fatalf("hmem %t world %t, want %t %t", has(logDir, "hmem"), has(logDir, "world"), tc.hmem, tc.world)
			}
		})
	}
}

// TestSampleReviewRefusesAWrongEngineSwitchByName proves a wrong engine switch stops Open in sample mode,
// naming the variable, as it does in a live review (ADR-0013). Each row mutates one variable.
func TestSampleReviewRefusesAWrongEngineSwitchByName(t *testing.T) {
	for _, name := range []string{"PROVEN_MARKOVIAN", "PROVEN_HMEM", "PROVEN_WORLD_MODEL"} {
		_, err := review.Open(review.Options{
			Sample: true, LogDir: t.TempDir(),
			Env: func(k string) string {
				if k == name {
					return "maybe"
				}
				return ""
			},
		})
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Fatalf("%s=maybe: %v", name, err)
		}
	}
}
