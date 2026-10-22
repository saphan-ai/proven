// Package prover asks the test-run adapter for a test that fails on the change and passes on the base.
package prover

import (
	"context"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/testrun"
)

// testSpec is the JSON the prover asks a model to return.
var testSpec = answer.Spec{
	Task:    "Name one test that fails on the change and passes on the base.",
	Shape:   "one JSON object. It has path (string), name (string) and command (string).",
	Example: `{"path":"pkg/file_test.go","name":"TestExample","command":"go test ./pkg -run TestExample"}`,
}

// Prompt asks which test would prove one risk.
func Prompt(id, title string) string {
	return answer.Instruct(testSpec, "id: "+id+"\ntitle: "+title+"\n")
}

type row struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Prove asks for a test name, then asks the adapter for its outcome.
// An empty name means the generator proposed no test. The adapter executes no code in sample mode.
func Prove(ctx context.Context, c candidate.Candidate, gen generator.Generator, run testrun.Adapter) (candidate.Candidate, error) {
	var got row
	if err := answer.Read(ctx, gen, "prover", Prompt(c.ID, c.Title), testSpec, &got); err != nil {
		return c, err
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
