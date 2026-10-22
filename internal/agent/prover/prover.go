// Package prover asks the test-run adapter for a test that fails on the change and passes on the base.
package prover

import (
	"context"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// testSpec is the JSON the prover asks a model to return.
var testSpec = answer.Spec{
	Task: "Write one test that fails on the change and passes on the base. " +
		"The test asserts the correct behaviour and fails because the change breaks it. " +
		"It does not attack anything. " +
		"Leave content empty to name one test that already exists.",
	Shape:   "one JSON object. It has path (string), content (string), name (string) and command (string).",
	Example: `{"path":"pkg/file_test.go","content":"package pkg\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tif Got() != 1 {\n\t\tt.Fatal(Got())\n\t}\n}\n","name":"TestExample","command":"go test ./pkg -run TestExample"}`,
}

const (
	notePassedOnChange = "written test passed on the change"
	noteFailedOnBase   = "written test failed on the base"
)

// Prompt asks which test would prove one risk.
func Prompt(id, title string) string {
	return answer.Instruct(testSpec, "id: "+id+"\ntitle: "+title+"\n")
}

type row struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Execute runs a test the skeptic named. It does not ask a model.
func Execute(ctx context.Context, run testrun.Adapter, req testrun.Request) (testrun.Outcome, error) {
	if run == nil {
		return testrun.Outcome{}, &service.Down{Service: "prover", Reason: "no test runner"}
	}
	return run.Run(ctx, req)
}

// Options are the checkouts and the size limit for a test the model writes.
// MaxWrittenBytes <= 0 uses the testrun default.
type Options struct {
	BaseDir         string
	ChangeDir       string
	MaxWrittenBytes int
}

// Prove asks for a test, then asks the adapter for its outcome.
// Content empty keeps the recorded path: name a test that already exists.
// Content set is a file the runner writes into a copy of each checkout.
// An empty name and empty content means the generator proposed no test.
func Prove(ctx context.Context, c candidate.Candidate, gen generator.Generator, run testrun.Adapter, opts Options) (candidate.Candidate, error) {
	var got row
	if err := answer.Read(ctx, gen, "prover", Prompt(c.ID, c.Title), testSpec, &got); err != nil {
		return c, err
	}
	if strings.TrimSpace(got.Content) == "" {
		return proveNamed(ctx, c, run, got)
	}
	return proveWritten(ctx, c, run, got, opts)
}

func proveNamed(ctx context.Context, c candidate.Candidate, run testrun.Adapter, got row) (candidate.Candidate, error) {
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
	c.Evidence = append(c.Evidence, testEvidence(out))
	return c, nil
}

func proveWritten(ctx context.Context, c candidate.Candidate, run testrun.Adapter, got row, opts Options) (candidate.Candidate, error) {
	if reason := testrun.RefuseWritten(got.Path, got.Content, got.Command, opts.BaseDir, opts.ChangeDir, opts.MaxWrittenBytes); reason != "" {
		c.ProverNote = "refused: " + reason
		return c, nil
	}
	out, err := run.Run(ctx, testrun.Request{
		FindingID: c.ID, Path: got.Path, Name: got.Name, Command: got.Command, Content: got.Content,
		BaseDir: opts.BaseDir, ChangeDir: opts.ChangeDir,
	})
	if err != nil {
		return c, err
	}
	if line := refusedLine(out.Output); line != "" {
		c.ProverNote = line
		return c, nil
	}
	if out.Found && out.FailsOnChange && out.PassesOnBase {
		c.Evidence = append(c.Evidence, testEvidence(out))
		return c, nil
	}
	// The passes-on-base check. A written test that passed on the base did not
	// fail on the change, so it is not evidence.
	// Negative control (a): delete this branch. TestPassesOnBothWritesNoEvidence turns red.
	if out.PassesOnBase {
		c.ProverNote = notePassedOnChange
		return c, nil
	}
	if !out.FailsOnChange && !out.PassesOnBase && out.Found && !couldNotRun(out) {
		c.ProverNote = notePassedOnChange
		return c, nil
	}
	if couldNotRun(out) {
		c.ProverNote = "the test could not run"
		return c, nil
	}
	c.ProverNote = noteFailedOnBase
	return c, nil
}

func testEvidence(out testrun.Outcome) candidate.Evidence {
	return candidate.Evidence{
		Kind: candidate.KindTest,
		Test: &candidate.TestEvidence{
			Path: out.Path, Name: out.Name, Command: out.Command,
			Output:        strings.TrimSpace(out.Output),
			FailsOnChange: out.FailsOnChange, PassesOnBase: out.PassesOnBase,
		},
	}
}

func refusedLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "refused:") {
			return line
		}
	}
	return ""
}

func couldNotRun(out testrun.Outcome) bool {
	if !out.Found {
		return true
	}
	return strings.Contains(out.Output, "could not run") || strings.Contains(out.Output, "not proof")
}
