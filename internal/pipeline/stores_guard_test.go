package pipeline

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
)

// Property: with the guardrail off, the draft the pipeline returns says so,
// not only the trace.
// Mutated field: Options.GuardOff.
func TestGuardOffIsNamedInTheDraft(t *testing.T) {
	d, _ := reviewWithRail(t, loadFixture(t), nil, true)
	if d.GuardrailStatement != candidate.GuardrailOffStatement {
		raw, _ := json.Marshal(d.GuardrailStatement)
		t.Fatalf("a review with the guardrail off carries no word of it outside the trace: %s", raw)
	}
	if on, _ := reviewWithRail(t, loadFixture(t), guardrail.Safe{}, false); on.GuardrailStatement != "" {
		t.Fatal("negative control: a review with the guardrail on must not say it was off")
	}
}

// Property: the generator the stores hold is the guarded one, so a call a
// store ever makes is checked before and after, like every agent call.
// Mutated field: the order of the stores' generator and the guard.
func TestStoresHoldTheGuardedGenerator(t *testing.T) {
	fx := loadFixture(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	rail := &countingRail{}
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), Generator: &callCount{inner: inner}, Guardrail: rail,
		Env: func(k string) string { t.Fatalf("sample mode read %s", k); return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Review(context.Background(), Request{Target: fx.Root}); err != nil {
		t.Fatal(err)
	}
	for name, gen := range map[string]generator.Generator{"hmem": p.mem.Generator(), "world": p.world.Generator()} {
		before := rail.checks
		// The sample has no recorded answer for this prompt; the call fails,
		// but the guard checks the prompt first.
		_, _ = gen.Complete(context.Background(), "planner", "store probe")
		if rail.checks == before {
			t.Fatalf("the %s store's generator is not guarded: the guardrail saw no check", name)
		}
	}
}

// countingRail counts the checks it is asked for and calls everything safe.
type countingRail struct{ checks int }

func (r *countingRail) Check(ctx context.Context, prompt, answer string) (guardrail.Verdict, error) {
	r.checks++
	return guardrail.Safe{}.Check(ctx, prompt, answer)
}
