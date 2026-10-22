// Package skeptic makes one refutation attempt per finding.
package skeptic

import (
	"context"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
)

// refuteSpec is the JSON the skeptic asks a model to return.
var refuteSpec = answer.Spec{
	Task:    "Make one attempt to refute this finding.",
	Shape:   "one JSON object. It has refuted (boolean) and reason (string).",
	Example: `{"refuted":false,"reason":"no refutation"}`,
}

// Prompt asks for one refutation of a finding.
func Prompt(id, title string) string {
	return answer.Instruct(refuteSpec, "id: "+id+"\ntitle: "+title+"\n")
}

type row struct {
	Refuted bool   `json:"refuted"`
	Reason  string `json:"reason"`
}

// Attempt asks for one refutation. It does not drop the finding; the judge does that.
func Attempt(ctx context.Context, c candidate.Candidate, gen generator.Generator) (candidate.Candidate, error) {
	var got row
	if err := answer.Read(ctx, gen, "skeptic", Prompt(c.ID, c.Title), refuteSpec, &got); err != nil {
		return c, err
	}
	reason := got.Reason
	if reason == "" {
		reason = "no refutation"
	}
	c.Refutation = candidate.Refutation{Attempted: true, Refuted: got.Refuted, Reason: reason}
	return c, nil
}
