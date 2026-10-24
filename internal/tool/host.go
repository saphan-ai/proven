package tool

import (
	"context"
	"errors"
	"time"

	"github.com/saphan-ai/proven/internal/testrun"
)

// ContainerHost runs a tool command in the Prover's container (ADR-0025, ADR-0040).
// Timeout is the tool's own limit, from PROVEN_TOOL_TIMEOUT. It is applied to
// a copy of Box, so the test runner's limit is left as it was.
type ContainerHost struct {
	Box     *testrun.Container
	Timeout time.Duration
}

// Run starts argv in dir. A non-positive Timeout is a named failure, not a
// fall back onto the test runner's limit: that would hide a missing setting.
func (h ContainerHost) Run(ctx context.Context, dir string, argv []string) (Side, error) {
	if h.Box == nil {
		return Side{}, errors.New("no container")
	}
	if h.Timeout <= 0 {
		return Side{}, errors.New("the time limit is not a positive duration")
	}
	box, err := h.Box.WithTimeout(h.Timeout)
	if err != nil {
		return Side{}, err
	}
	res, err := box.RunCommand(ctx, dir, argv)
	if err != nil {
		return Side{}, err
	}
	return Side{Output: res.Output, Status: res.Status}, nil
}
