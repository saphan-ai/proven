// Package prover asks the test-run adapter for a test that fails on the change and passes on the base.
package prover

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Prompt asks which test would prove one risk.
func Prompt(id, title string) string {
	return "prove " + id + "\n" + title
}

type row struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Prove asks for a test name, then asks the adapter for its outcome.
// An empty name means the generator proposed no test. The adapter executes no code in sample mode.
func Prove(ctx context.Context, c candidate.Candidate, gen generator.Generator, run testrun.Adapter) (candidate.Candidate, error) {
	raw, err := gen.Complete(ctx, "prover", Prompt(c.ID, c.Title))
	if err != nil {
		return c, err
	}
	var got row
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		return c, fmt.Errorf("prover: response is not a test: %w", err)
	}
	if got.Name == "" {
		return c, nil
	}
	out, err := run.Run(ctx, testrun.Request{
		FindingID: c.ID, Path: got.Path, Name: got.Name, Command: got.Command,
	})
	if err != nil {
		return c, err
	}
	if !out.Found || !out.FailsOnChange || !out.PassesOnBase {
		return c, nil
	}
	c.Evidence = append(c.Evidence, candidate.Evidence{
		Kind: candidate.KindTest,
		Test: &candidate.TestEvidence{
			Path: out.Path, Name: out.Name, Command: out.Command,
			Output:        strings.TrimSpace(out.Output),
			FailsOnChange: out.FailsOnChange, PassesOnBase: out.PassesOnBase,
		},
	})
	return c, nil
}
