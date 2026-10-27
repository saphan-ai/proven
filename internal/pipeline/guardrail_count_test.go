package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
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

// downRail never gives a verdict; err is what each check returns.
type downRail struct {
	err   error
	asked int
}

func (r *downRail) Identity() (string, string) { return "nvidia", "nvidia/probe-guard" }

func (r *downRail) Check(context.Context, string, string) (guardrail.Verdict, error) {
	r.asked++
	return guardrail.Verdict{}, r.err
}

// Property: a check the guardrail could not make is counted as not checked, never as
// a checked (safe) prompt, whatever its refusal says, an empty reason included
// (review claude r3 C1). Mutated line for the negative control: the Down test in
// countGuard; keyed on the refusal text again, the empty-reason cases are red.
func TestDraftCountsADownGuardrailAsNotChecked(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"named-reason", guardrail.Refuse("unreachable")},
		{"empty-reason", guardrail.Refuse("")},
		{"foreign-empty-error", errors.New("")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rail := &downRail{err: tc.err}
			d, _ := reviewWithRail(t, loadFixture(t), rail, false)
			g := d.Guardrail
			t.Logf("asked=%d count=%+v", rail.asked, g)
			if rail.asked == 0 {
				t.Fatal("the rail was never asked")
			}
			if g.Prompts != 0 || g.Answers != 0 || g.Unanswered != rail.asked {
				t.Errorf("a down check was counted as made: asked %d, count %+v", rail.asked, g)
			}
		})
	}
}

// Property: a review's guardrail count is that review's alone, because a Pipeline
// runs one review: a second Review on it is refused (its trace is closed), so no
// count can carry over (review codex r2/r4, owed). The reset at the start of Review
// is kept as a second line of defence. Mutated condition for the negative control:
// a Pipeline that ran a second review would make this test red, and the reset would
// then be the line that keeps the counts apart.
func TestAPipelineRunsOneReviewSoItsCountIsThatReviews(t *testing.T) {
	fx := loadFixture(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), Generator: inner, Guardrail: &scriptedRail{safeChecks: 1000},
		Env: func(k string) string { t.Fatalf("sample mode read %s", k); return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := p.Review(context.Background(), Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	if first.Guardrail.Prompts == 0 {
		t.Fatal("the first review counted no prompt")
	}
	if _, err := p.Review(context.Background(), Request{Target: fx.Root}); err == nil {
		t.Fatal("a second review ran on the same Pipeline; its guardrail count must then be shown to start from zero")
	}
}
