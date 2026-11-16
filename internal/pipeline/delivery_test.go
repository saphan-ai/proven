package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
)

type deliveryModel struct {
	inner    generator.Generator
	response string
	calls    int
}

func (m *deliveryModel) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "delivery" {
		m.calls++
		return m.response, nil
	}
	return m.inner.Complete(ctx, agent, prompt)
}

func TestDeliveryRunsInsideExistingPipelineAndPreservesFindingsOnFailure(t *testing.T) {
	fx, err := fixture.LoadUnder(change.ProcessPolicy(), fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, valid := range []bool{true, false} {
		inner, err := generator.NewSample(fx.Answers)
		if err != nil {
			t.Fatal(err)
		}
		response := `{"claims":[{"requirement":"Preserve behavior","source_kind":"description","source_index":0,"source_quote":"Preserve behavior","status":"unclear","reason":"Needs acceptance coverage beyond this bounded comparison","citations":[]}]}`
		if !valid {
			response = `{"claims":[{"requirement":"Preserve behavior","source_kind":"description","source_quote":"invented","status":"unclear","reason":"not supported","citations":[]}]}`
		}
		gen := &deliveryModel{inner: inner, response: response}
		p, err := Build(Options{Sample: true, LogDir: t.TempDir(), Generator: gen})
		if err != nil {
			t.Fatal(err)
		}
		d, err := p.Review(context.Background(), Request{Target: fixturePath(t), FixesOff: true, Delivery: &delivery.Metadata{Description: "Preserve behavior"}})
		if err != nil || d.Delivery == nil || d.Delivery.Complete != valid || gen.calls != 1 {
			t.Fatalf("delivery=%+v err=%v calls=%d", d.Delivery, err, gen.calls)
		}
		if _, ok := kept(d, "F1"); !ok {
			t.Fatal("delivery assessment discarded an existing test-proved finding")
		}
		failed := false
		for _, f := range d.Failed {
			if f.Service == "delivery" {
				failed = true
			}
		}
		if failed == valid {
			t.Fatalf("failure classification: %+v", d.Failed)
		}
		if !valid && !strings.Contains(d.Delivery.Issue, "invalid") {
			t.Fatal(d.Delivery.Issue)
		}
	}
}
