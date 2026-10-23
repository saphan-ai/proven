// Package skeptic makes one refutation attempt per finding.
//
// The Skeptic gets each finding and makes one attempt to
// refute it (ADR-0006). It runs on another model family than the finder,
// through PROVEN_SKEPTIC_MODEL. It only proposes: a refuting test goes to the
// Prover's runner, a refuting query to the Investigator's engine, and the
// Judge decides on what they return. It never drops a finding itself.
package skeptic

import (
	"context"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
)

// refuteSpec is the JSON the skeptic asks a model to return.
// A test or a query is a proposal. The prover or the investigator runs it.
// The model's refuted flag does not drop the finding.
var refuteSpec = answer.Spec{
	Task:    "Make one attempt to refute this finding.",
	Shape:   "one JSON object. It has refuted (boolean), reason (string), path (string), name (string), command (string) and sql (string). Empty path, name, command and sql mean there is no refuting test and no refuting query.",
	Example: `{"refuted":false,"reason":"no refutation","path":"","name":"","command":"","sql":""}`,
}

// Prompt asks for one refutation of a finding.
func Prompt(id, title string) string {
	return answer.Instruct(refuteSpec, "id: "+id+"\ntitle: "+title+"\n")
}

// row is the refutation as the model returns it.
type row struct {
	Refuted bool   `json:"refuted"`
	Reason  string `json:"reason"`
	Path    string `json:"path"`
	Name    string `json:"name"`
	Command string `json:"command"`
	SQL     string `json:"sql"`
}

// Attempt asks for one refutation. It does not run the proposal and it does
// not drop the finding. The prover or the investigator runs a proposal, and
// the judge decides.
func Attempt(ctx context.Context, c candidate.Candidate, gen generator.Generator) (candidate.Candidate, error) {
	var got row
	if err := answer.Read(ctx, gen, "skeptic", Prompt(c.ID, c.Title), refuteSpec, &got); err != nil {
		return c, err
	}
	reason := got.Reason
	// The trace always says why, even when the model gave no reason.
	if reason == "" {
		reason = "no refutation"
	}
	c.Refutation = candidate.Refutation{
		Attempted: true,
		Refuted:   got.Refuted,
		Reason:    reason,
		TestPath:  got.Path,
		TestName:  got.Name,
		Command:   got.Command,
		SQL:       strings.TrimSpace(got.SQL),
	}
	return c, nil
}
