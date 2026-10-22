package testrun

import (
	"context"
	"errors"
	"time"
)

// ADR-0025: the change's code and tests run only inside a container Proven starts
// for that one run, with no network and no secret. This file is the contract of
// wagon W1b (proven-reviewer-core); the bodies are stubs until W1b lands.

// ErrContainerNotImplemented is returned by the stubs below until W1b lands.
var ErrContainerNotImplemented = errors.New("testrun: container runner not implemented")

// ContainerConfig holds the container's settings (ADR-0025 point 5). There is no
// field that gives the container a network, an environment variable or a mount
// beyond the one checkout.
type ContainerConfig struct {
	Runtime string        // container CLI name or path; default "docker"
	Image   string        // image the test runs in
	Timeout time.Duration // limit for one test run, base or change
	CPUs    string        // CPU limit passed to the runtime
	Memory  string        // memory limit passed to the runtime
}

// DefaultContainerConfig returns the documented defaults.
func DefaultContainerConfig() ContainerConfig { return ContainerConfig{} }

// Validate refuses a setting by name with the allowed values (ADR-0013).
func (c ContainerConfig) Validate() error { return ErrContainerNotImplemented }

// Container is the live test-run adapter.
type Container struct{}

// NewContainer resolves the runtime with lookPath. When no runtime is found it
// returns a *service.Down naming the service "prover" (ADR-0025 point 6), and the
// caller runs no code.
func NewContainer(cfg ContainerConfig, lookPath func(string) (string, error)) (*Container, error) {
	return nil, ErrContainerNotImplemented
}

// Argv is the runtime's argument list for running command in dir: the whole
// sentence handed to the runtime, so a test can read every property of the run.
func (c *Container) Argv(dir, command string) []string { return nil }

// Run runs req.Command once in req.BaseDir and once in req.ChangeDir, each in its
// own container, and reports whether it fails on the change and passes on the base.
func (c *Container) Run(ctx context.Context, req Request) (Outcome, error) {
	return Outcome{}, ErrContainerNotImplemented
}
