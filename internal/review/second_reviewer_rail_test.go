package review

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/service"
)

// downPlanner is a first reviewer whose Planner call fails as a service failure and which answers
// every other agent as the wrapped script does.
type downPlanner struct {
	*scriptGen
}

// Complete fails the Planner's call as a service that is down.
func (g downPlanner) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "planner" {
		return "", &service.Down{Service: "nvidia", Reason: "unavailable"}
	}
	return g.scriptGen.Complete(ctx, agent, prompt)
}

// TestAFailedFirstReviewerIsNotADisagreement proves that when the first reviewer's Planner fails as a
// service failure and the second answers, the first reviewer's verdict on the second's risk is
// "no answer", and nothing is flagged or counted as a disagreement: a missing answer is not a
// different answer, whichever reviewer it is missing from.
// Mutated field: whether the first reviewer's Planner call fails (the second reviewer answers either way).
func TestAFailedFirstReviewerIsNotADisagreement(t *testing.T) {
	def := downPlanner{newScript("NVIDIA", risks(), `{}`)}
	def.sql = map[string]string{"S1": rowAnswer}
	api := &secondAPI{planner: risks(risk("S1", "second only")), skeptic: `{"refuted":false,"reason":"second-keeps-it"}`}
	res := reviewWithSecond(t, defaultSettings(), def, api)
	s1, ok := finding(res, "S1")
	if !ok {
		t.Fatal("S1 missing")
	}
	if v, _ := verdictOf(s1, RoleDefault); v.Verdict != VerdictNoAnswer || s1.Disagree || res.Disagreements != 0 {
		t.Fatalf("S1 %+v disagree %v disagreements %d", s1.Verdicts, s1.Disagree, res.Disagreements)
	}
	if len(s1.Verdicts) != 2 || s1.Verdicts[0].Role != RoleDefault {
		t.Fatalf("the first reviewer is listed first: %+v", s1.Verdicts)
	}
}

// TestTheGuardrailWrapsTheSecondReviewer proves ADR-0005 for the second reviewer's endpoint: no prompt
// reaches it unless the guardrail passed that prompt. With the guardrail absent (live mode, key empty) or
// refusing every prompt, zero prompts reach the second API and the trace carries a guardrail refusal for
// the second reviewer's call as well as the first's. With a guardrail that answers safe the same review
// does reach the second API, so the zero above is the guardrail's doing.
// Mutated field: the guardrail adapter on the pipeline (the second reviewer's generator is the same).
func TestTheGuardrailWrapsTheSecondReviewer(t *testing.T) {
	sample, err := guardrail.NewSample(nil)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		rail    guardrail.Adapter
		reaches bool
		reason  string
	}{
		{"absent", guardrail.Absent{Reason: "PROVEN_GUARDRAIL_KEY is empty"}, false, "guardrail: PROVEN_GUARDRAIL_KEY is empty"},
		{"refusing sample", sample, false, "guardrail: no recorded verdict"},
		{"safe", guardrail.Safe{}, true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			def := newScript("NVIDIA", risks(), `{}`)
			api := &secondAPI{planner: risks(), skeptic: `{}`}
			res := reviewWithSecondRail(t, defaultSettings(), def, api, c.rail)
			if got := len(api.prompts) > 0; got != c.reaches {
				t.Fatalf("prompts reaching the second reviewer: %d, want reaching %v", len(api.prompts), c.reaches)
			}
			if c.reaches {
				return
			}
			// One refusal for each reviewer's Planner call: the first reviewer's and the second's.
			body := traceBody(t, res.TracePath)
			if n := strings.Count(body, c.reason); n < 2 {
				t.Fatalf("trace holds %d refusals %q, want one per reviewer\n%s", n, c.reason, body)
			}
		})
	}
}
