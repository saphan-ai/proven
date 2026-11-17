package codecomments

import (
	"context"
	data "github.com/saphan-ai/proven/internal/codecomments"
	"testing"
)

type model struct {
	answer string
	calls  int
}

func (m *model) Complete(context.Context, string, string) (string, error) {
	m.calls++
	return m.answer, nil
}
func TestCitationCannotReferToOtherCodeAndFailurePreservesStructure(t *testing.T) {
	for _, tc := range []struct {
		quote    string
		complete bool
	}{{"return amount", true}, {"if amount < 0", false}} {
		gen := &model{answer: `{"unit_id":"u","status":"needs_attention","reason":"The comment and supplied body differ.","quote":"` + tc.quote + `"}`}
		r := data.Result{Complete: true, Units: []data.Unit{{ID: "u", Text: "Reject negative amounts.", Code: "return amount"}}, Issues: []data.Issue{{UnitID: "u", Rule: "stale_param_tag"}}}
		r = Assess(context.Background(), gen, r)
		if r.Complete != tc.complete || len(r.Issues) != 1 || len(r.Assessments) != 1 {
			t.Fatalf("%+v", r)
		}
	}
}
func TestCancelledAndExhaustedBudgetsNeverDisappear(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	gen := &model{}
	r := Assess(ctx, gen, data.Result{Complete: true, Units: []data.Unit{{ID: "u", Text: "Documented behavior", Code: "return"}}})
	if gen.calls != 0 || r.Complete || len(r.Assessments) != 1 || r.Assessments[0].Status != "not_assessed" {
		t.Fatal(r)
	}
}
