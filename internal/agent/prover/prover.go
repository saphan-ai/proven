// Package prover asks the test-run adapter for a test that fails on the change and passes on the base.
//
// The Prover gets each risk from the Planner and asks the model for a test that
// proves it: one that fails on the change and passes on the base (ADR-0002).
// The test asserts correct behaviour and never attacks anything (ADR-0007). The
// test-run adapter runs it once on each checkout, with no network and no secret
// (ADR-0025). Only that outcome, never the model's word, becomes evidence.
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
	Shape:   "one JSON object. It has path (string), content (string) and name (string). Do not add a command: Proven builds the command that runs the test.",
	Example: `{"path":"pkg/file_test.go","content":"package pkg\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {\n\tif Got() != 1 {\n\t\tt.Fatal(Got())\n\t}\n}\n","name":"TestExample"}`,
}

// Prover notes say why a written test gave no evidence. The trace and the review show them.
const (
	notePassedOnChange = "written test passed on the change"
	noteFailedOnBase   = "written test failed on the base"
	noteNotFound       = "written test was not found"
)

// Prompt asks which test would prove one risk.
func Prompt(id, title string) string {
	return answer.Instruct(testSpec, "id: "+id+"\ntitle: "+title+"\n")
}

// row is the test as the model returns it. The prompt does not ask for Command.
// It is read only so that an answer that carries one can be refused.
type row struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Execute runs a test the skeptic named. It does not ask a model. The name and
// path are checked first (ADR-0025): one that fails is an outcome with no test
// found and a "refused:" line, and the runner is not called.
func Execute(ctx context.Context, run testrun.Adapter, req testrun.Request) (testrun.Outcome, error) {
	if run == nil {
		return testrun.Outcome{}, &service.Down{Service: "prover", Reason: "no test runner"}
	}
	if reason := testrun.RefuseTest(req.Path, req.Name); reason != "" {
		return testrun.Outcome{Path: req.Path, Name: req.Name, Output: "refused: " + reason}, nil
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
	// ADR-0025: only the test's file and name are the model's. An answer that
	// also carries a command is refused whole, so no text of a model is ever
	// close to the line the container runs.
	if strings.TrimSpace(got.Command) != "" {
		c.ProverNote = "refused: " + testrun.ReasonModelCommand
		return c, nil
	}
	// No content names a test that already exists; content is a new test file.
	if strings.TrimSpace(got.Content) == "" {
		return proveNamed(ctx, c, run, got)
	}
	return proveWritten(ctx, c, run, got, opts)
}

// proveNamed runs a test that already exists in the repository. Only a test
// that is found, fails on the change and passes on the base is evidence;
// anything else leaves the candidate as it was.
func proveNamed(ctx context.Context, c candidate.Candidate, run testrun.Adapter, got row) (candidate.Candidate, error) {
	if got.Name == "" {
		return c, nil
	}
	// The same refusal as Execute. A nil runner panics on Run.
	if run == nil {
		return c, &service.Down{Service: "prover", Reason: "no test runner"}
	}
	// A name or path that fails the checks gives the finding no test evidence.
	if reason := testrun.RefuseTest(got.Path, got.Name); reason != "" {
		c.ProverNote = "refused: " + reason
		return c, nil
	}
	out, err := run.Run(ctx, testrun.Request{
		FindingID: c.ID, Path: got.Path, Name: got.Name,
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

// proveWritten runs a test the model wrote. The file is checked before anything
// runs, and every outcome that is not evidence leaves a note saying why.
func proveWritten(ctx context.Context, c candidate.Candidate, run testrun.Adapter, got row, opts Options) (candidate.Candidate, error) {
	// A path outside the repository, a file that is not a test, an overwrite of
	// code, a name that is not a test name or a file over the size limit is
	// refused before the runner is called.
	if reason := testrun.RefuseWritten(got.Path, got.Content, got.Name, opts.BaseDir, opts.ChangeDir, opts.MaxWrittenBytes); reason != "" {
		c.ProverNote = "refused: " + reason
		return c, nil
	}
	// The same refusal as Execute. A nil runner panics on Run.
	if run == nil {
		return c, &service.Down{Service: "prover", Reason: "no test runner"}
	}
	out, err := run.Run(ctx, testrun.Request{
		FindingID: c.ID, Path: got.Path, Name: got.Name, Content: got.Content,
		BaseDir: opts.BaseDir, ChangeDir: opts.ChangeDir,
	})
	if err != nil {
		return c, err
	}
	// The runner itself may refuse; its refusal line becomes the note.
	if line := refusedLine(out.Output); line != "" {
		c.ProverNote = line
		return c, nil
	}
	// A runner can report PassesOnBase on a test it did not find. Found is
	// what happened; the pass flag is not.
	if !out.Found {
		c.ProverNote = noteNotFound
		return c, nil
	}
	if out.FailsOnChange && out.PassesOnBase {
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

// testEvidence is the test evidence the Judge reads: the test, its command,
// its output and its two outcomes.
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

// refusedLine returns the first "refused:" line of the runner's output.
func refusedLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "refused:") {
			return line
		}
	}
	return ""
}

// couldNotRun reports a test that was not found or that the runner says did not
// run, so its outcome cannot count either way.
func couldNotRun(out testrun.Outcome) bool {
	if !out.Found {
		return true
	}
	return strings.Contains(out.Output, "could not run") || strings.Contains(out.Output, "not proof")
}
