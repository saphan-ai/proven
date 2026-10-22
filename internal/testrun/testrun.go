// Package testrun asks whether a test fails on the change and passes on the base.
package testrun

import "context"

// Request names the test the prover wants run.
// Content is the full text of a test file the prover wrote. Empty means the
// test already exists and the runner executes nothing it was not already going to run.
type Request struct {
	FindingID string
	Path      string
	Name      string
	Command   string
	Content   string
	BaseDir   string // checkout of the base; the container mounts it alone for the base run
	ChangeDir string // checkout of the change; the container mounts it alone for the change run
}

// Outcome is the result. Found is false when the fixture recorded no such test.
type Outcome struct {
	Found         bool
	Path          string
	Name          string
	Command       string
	Output        string
	FailsOnChange bool
	PassesOnBase  bool
}

// Adapter returns a test outcome. The sample adapter executes no code.
type Adapter interface {
	Run(ctx context.Context, req Request) (Outcome, error)
}
