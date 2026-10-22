// Package skeptic makes one refutation attempt per finding.
package skeptic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
)

// Prompt asks for one refutation of a finding.
func Prompt(id, title string) string {
	return "refute " + id + "\n" + title
}

type row struct {
	Refuted bool   `json:"refuted"`
	Reason  string `json:"reason"`
}

// Attempt asks the generator once. It does not drop the finding; the judge does that.
func Attempt(ctx context.Context, c candidate.Candidate, gen generator.Generator) (candidate.Candidate, error) {
	raw, err := gen.Complete(ctx, "skeptic", Prompt(c.ID, c.Title))
	if err != nil {
		return c, err
	}
	var got row
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		return c, fmt.Errorf("skeptic: response is not a refutation: %w", err)
	}
	reason := got.Reason
	if reason == "" {
		reason = "no refutation"
	}
	c.Refutation = candidate.Refutation{Attempted: true, Refuted: got.Refuted, Reason: reason}
	return c, nil
}
