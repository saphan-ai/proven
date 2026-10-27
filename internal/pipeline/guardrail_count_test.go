package pipeline

import (
	"context"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/guardrail"
)

// scriptedRail answers safe for its first safeChecks checks and unsafe after,
// and keeps its own tally of the checks it was asked, so a test can compare the
// pipeline's count with a count that did not come from the pipeline.
// It names itself, as the NVIDIA adapter does.
type scriptedRail struct {
	mu         sync.Mutex
	safeChecks int
	asked      int
	prompts    int
	answers    int
}

func (r *scriptedRail) Identity() (string, string) { return "nvidia", "nvidia/test-guard-model" }

func (r *scriptedRail) Check(_ context.Context, _, answer string) (guardrail.Verdict, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.asked++
	if answer == "" {
		r.prompts++
	} else {
		r.answers++
	}
	if r.asked <= r.safeChecks {
		return guardrail.Verdict{User: "safe", Response: "safe"}, nil
	}
	return guardrail.Verdict{User: "unsafe", Categories: []string{"Other"}}, nil
}

// Property: the draft carries how many prompts and answers the guardrail
// checked and how many it blocked, counted at the callback every check passes
// through, with the backend and the model that answered.
// Mutated line for the negative control: the p.countGuard call in (*Pipeline).guard;
// without it every count below reads 0 and this test is red.
func TestDraftCountsTheGuardrailChecksItMade(t *testing.T) {
	fx := loadFixture(t)
	rail := &scriptedRail{safeChecks: 2}
	d, _ := reviewWithRail(t, fx, rail, false)

	got := d.Guardrail
	if got.Prompts+got.Answers != rail.asked {
		t.Fatalf("the draft counts %d prompts and %d answers; the guardrail was asked %d times", got.Prompts, got.Answers, rail.asked)
	}
	if got.Prompts != rail.prompts || got.Answers != rail.answers {
		t.Fatalf("the draft counts %d prompts and %d answers; the guardrail was asked %d and %d", got.Prompts, got.Answers, rail.prompts, rail.answers)
	}
	// The first prompt and its answer are safe; every later prompt is unsafe and
	// is blocked, so one answer is checked and each other check is a blocked
	// prompt. The 13 and 12 are what this fixture's review asks, read off a run
	// and held here so a change in what is counted turns the test red.
	if got.Prompts != 13 || got.Answers != 1 || got.Blocked != 12 {
		t.Fatalf("counts = %+v, want 13 prompts, 1 answer and 12 blocked", got)
	}
	if got.Unanswered != 0 {
		t.Fatalf("unanswered = %d, want 0: every check got a verdict", got.Unanswered)
	}
	if got.Backend != "nvidia" || got.Model != "nvidia/test-guard-model" {
		t.Fatalf("the guardrail is named %q %q", got.Backend, got.Model)
	}
}

// Property: a review run with the guardrail off counts no check, and says so.
// Mutated field: the GuardOff option.
func TestDraftWithTheGuardrailOffCountsNothing(t *testing.T) {
	fx := loadFixture(t)
	d, _ := reviewWithRail(t, fx, nil, true)
	if d.Guardrail.Prompts+d.Guardrail.Answers+d.Guardrail.Blocked+d.Guardrail.Unanswered != 0 {
		t.Fatalf("a review with the guardrail off counted checks: %+v", d.Guardrail)
	}
	if d.GuardrailStatement == "" {
		t.Fatal("a review with the guardrail off does not say so")
	}
}
