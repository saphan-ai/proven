package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/redact"
)

// rewriting answers like the sample generator, except that the planner's answer
// passes through rewrite, the way a model can echo a configured value in its
// own text. The marker is removed from every prompt, so the recorded answers
// for the later agents still match.
type rewriting struct {
	inner   generator.Generator
	marker  string
	rewrite func(answer string) string
}

func (g *rewriting) Complete(ctx context.Context, agent, prompt string) (string, error) {
	answer, err := g.inner.Complete(ctx, agent, strings.ReplaceAll(prompt, g.marker, ""))
	if err != nil || agent != "planner" {
		return answer, err
	}
	return g.rewrite(answer), nil
}

// reviewRewriting runs one sample review whose planner answer is rewritten,
// with a redactor holding marker as the configured token.
func reviewRewriting(t *testing.T, marker string, rewrite func(string) string) (candidate.Draft, error) {
	t.Helper()
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	env := secretEnv(map[string]string{generator.EnvToken: marker})
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), Redactor: redact.FromEnv(env),
		Generator: &rewriting{inner: inner, marker: marker, rewrite: rewrite},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fx.Root})
	return d, err
}

// TestTokenInAFindingDoesNotReachTheDraft proves that a configured token a
// model echoes in a finding's title and message does not reach the draft the
// review returns, and that the finding is still reported.
// Mutated field: the title and message of the planner's risk.
func TestTokenInAFindingDoesNotReachTheDraft(t *testing.T) {
	const marker = "marker-w30-find-91ab"
	res, err := reviewRewriting(t, marker, func(a string) string {
		a = strings.ReplaceAll(a, "Discount is applied twice", "Discount is applied twice"+marker)
		return a
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) == 0 {
		t.Fatal("the review reported no finding, so the property was not exercised")
	}
	for _, f := range res.Findings {
		if strings.Contains(f.Title+f.Message, marker) {
			t.Errorf("the finding carries the token: %q", f.Title)
		}
	}
}

// TestTokenInAFatalAgentErrorDoesNotReachTheError proves that a configured
// token a model echoes as a repeated risk id is not part of the error the
// review returns.
// Mutated field: the ids of the risks the planner names.
func TestTokenInAFatalAgentErrorDoesNotReachTheError(t *testing.T) {
	const marker = "marker-w30-dup-44cd"
	_, err := reviewRewriting(t, marker, func(string) string {
		return `[{"id":"` + marker + `","title":"a","message":"a","severity":"INFO","path":"billing/discount.go","start_line":1,"end_line":1},` +
			`{"id":"` + marker + `","title":"b","message":"b","severity":"INFO","path":"billing/discount.go","start_line":1,"end_line":1}]`
	})
	if err == nil {
		t.Fatal("a repeated risk id was accepted")
	}
	if strings.Contains(err.Error(), marker) {
		t.Errorf("the error carries the token: %v", err)
	}
}

// TestRedactFindingsReachesEvidence proves that the redaction of a finding
// covers its evidence and its refutation, and leaves the caller's candidates
// unchanged.
// Mutated field: the test output, the query rows, the tool output, the call path and the refutation reason.
func TestRedactFindingsReachesEvidence(t *testing.T) {
	const marker = "marker-w30-evid-7be1"
	p := &Pipeline{redactor: redact.New(map[string]string{generator.EnvToken: marker})}
	in := []candidate.Candidate{{
		Risk:     candidate.Risk{ID: "F1", Title: "t " + marker},
		CallPath: "path " + marker,
		Evidence: []candidate.Evidence{
			{Kind: candidate.KindTest, Test: &candidate.TestEvidence{Output: "out " + marker}},
			{Kind: candidate.KindQuery, Query: &candidate.QueryEvidence{Rows: [][]string{{"r " + marker}}}},
			{Kind: candidate.KindTool, Tool: &candidate.ToolEvidence{Command: "cmd " + marker, Output: "tool " + marker, BaseOutput: "base " + marker}},
		},
		Refutation: candidate.Refutation{Reason: "why " + marker, QueryRows: [][]string{{marker}}},
	}}
	out := p.redactFindings(in)
	got := out[0]
	tool := got.Evidence[2].Tool
	if strings.Contains(got.Title+got.CallPath+got.Evidence[0].Test.Output+got.Evidence[1].Query.Rows[0][0]+tool.Command+tool.Output+tool.BaseOutput+got.Refutation.Reason+got.Refutation.QueryRows[0][0], marker) {
		t.Errorf("a redacted finding still carries the token: %#v", got)
	}
	if !strings.Contains(in[0].Evidence[0].Test.Output, marker) {
		t.Error("the caller's evidence was changed")
	}
}

// TestLivePipelineWithoutARedactorIsRefused proves that Build refuses a live
// pipeline that has a guardrail and no redactor, and accepts the same options
// once a redactor is set.
// Mutated field: Options.Redactor.
func TestLivePipelineWithoutARedactorIsRefused(t *testing.T) {
	opts := Options{LogDir: t.TempDir(), Guardrail: guardrail.Safe{}}
	if _, err := Build(opts); err == nil {
		t.Fatal("a live pipeline with no redactor was accepted")
	}
	opts.Redactor = redact.New(nil)
	if _, err := Build(opts); err != nil {
		t.Fatalf("a live pipeline with a redactor was refused: %v", err)
	}
}
