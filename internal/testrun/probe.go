package testrun

import (
	"context"
	"errors"
	"strings"

	"github.com/saphan-ai/proven/internal/service"
)

// Probe asks the configured runtime whether it answers, then whether the
// image is already on the host. The argument list is version, then
// image inspect. There is no pull: a missing image is not fetched.
// The call is bounded by the container time limit. A run past that limit
// is a timeout, not proof that the image is absent.
func (c *Container) Probe(ctx context.Context) error {
	if c == nil || c.runtime == "" {
		return &service.Down{Service: "prover", Reason: "container runtime is not configured"}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, c.cfg.Timeout)
	defer cancel()
	if out, err := c.invoke(ctx, []string{c.runtime, "version"}); err != nil {
		return probeDown(parent, ctx, "runtime did not answer", out, err)
	}
	if out, err := c.invoke(ctx, []string{c.runtime, "image", "inspect", c.cfg.Image}); err != nil {
		return probeDown(parent, ctx, "image "+c.cfg.Image+" is not present", out, err)
	}
	return nil
}

func probeDown(parent, ctx context.Context, reason, output string, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return &service.Down{Service: "prover", Reason: "timeout"}
	}
	msg := reason
	detail := oneLine(output)
	if detail == "" && err != nil {
		detail = oneLine(err.Error())
	}
	if detail != "" {
		msg += ": " + detail
	}
	return &service.Down{Service: "prover", Reason: msg}
}

func oneLine(s string) string {
	var last string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			last = line
		}
	}
	last = strings.Join(strings.Fields(last), " ")
	if len(last) > 160 {
		return last[:160]
	}
	return last
}
