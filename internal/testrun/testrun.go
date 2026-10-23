// Package testrun asks whether a test fails on the change and passes on the base.
//
// It is the Prover's evidence source for tests (ADR-0002): a finding is proven
// when a test that names it fails on the change and passes on the base. The
// Prover asks through the Adapter interface. Container
// runs the test in a container with no network and no secret (ADR-0025);
// Sample replays outcomes recorded in a fixture and runs no code; Off and Live
// report the Prover off, so a review without a container runtime never runs
// the change's code on the host (ADR-0025 point 6).
package testrun

import "context"

// Request names the test the prover wants run.
// Content is the full text of a test file the prover wrote. Empty means the
// test already exists and the runner executes nothing it was not already going to run.
// FindingID ties the run to the finding it is meant to prove. Path and Name
// name the test file and the test; they are all a model decides about the run.
// The command that runs the test is built from them by Proven (TestArgv).
type Request struct {
	FindingID string
	Path      string
	Name      string
	Content   string
	BaseDir   string // checkout of the base; the container mounts it alone for the base run
	ChangeDir string // checkout of the change; the container mounts it alone for the change run
}

// Outcome is the result. Found is false when the fixture recorded no such test.
// A finding is proven only when FailsOnChange and PassesOnBase are both true:
// a test that fails on both sides shows a broken base, not the change.
type Outcome struct {
	Found         bool
	Path          string
	Name          string
	Command       string // the line Proven built, shown to the reader of the trace
	Output        string
	FailsOnChange bool
	PassesOnBase  bool
}

// Adapter returns a test outcome. The sample adapter executes no code.
// A *service.Down error means the Prover is off; the review names it among
// the failed services and the finding keeps a query as its evidence.
type Adapter interface {
	Run(ctx context.Context, req Request) (Outcome, error)
}
