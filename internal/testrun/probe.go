package testrun

import (
	"context"
	"errors"
	"path/filepath"
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
	// parent is kept apart from the bounded ctx so probeDown can tell the
	// caller's cancellation from the probe's own time limit.
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, c.cfg.Timeout)
	defer cancel()
	if out, err := c.invoke(ctx, []string{c.runtime, "version"}); err != nil {
		return probeDown(parent, ctx, "runtime did not answer", out, err)
	}
	if out, err := c.invoke(ctx, []string{c.runtime, "image", "inspect", c.cfg.Image}); err != nil {
		// The Prover never pulls during a review: a missing image is the
		// service down, and the reason names the act that fixes it.
		down := probeDown(parent, ctx, "image "+c.cfg.Image+" is not present", out, err)
		if d, ok := service.As(down); ok && d.Reason != "timeout" {
			d.Reason += "; next: " + filepath.Base(c.runtime) + " pull " + c.cfg.Image
			return d
		}
		return down
	}
	return nil
}

// probeDown turns a failed probe into the error a review reports. The
// caller's cancellation wins, then the time limit, then the runtime's own
// last line of output, so the failed service names a cause a reader can act on.
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

// oneLine keeps the last non-empty line of s, with spaces collapsed and at
// most 160 bytes: the runtime prints its error last, and a reason must fit one line.
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
