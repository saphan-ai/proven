package guardrail

import (
	"context"
	"errors"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

// Observe is called once per check.
// refused is empty when the guardrail returned a verdict, and the reason when it did not.
type Observe func(agent, side string, v Verdict, refused string) error

// Wrap checks the prompt before the generator runs, and the answer before it is returned.
// An unsafe verdict, or a guardrail that does not answer, is returned and the text is not.
func Wrap(inner generator.Generator, rail Adapter, observe Observe) (generator.Generator, error) {
	if inner == nil || rail == nil {
		return nil, Refuse("not configured")
	}
	return &guarded{inner: inner, rail: rail, observe: observe}, nil
}

// guarded is the Generator Wrap returns.
type guarded struct {
	inner   generator.Generator
	rail    Adapter
	observe Observe
}

// Complete checks the prompt, calls the model, then checks the prompt and
// answer together. The order matters: an unsafe prompt never reaches the model.
func (g *guarded) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if err := g.side(ctx, agent, "prompt", prompt, ""); err != nil {
		return "", err
	}
	text, err := g.inner.Complete(ctx, agent, prompt)
	if err != nil {
		return "", err
	}
	if err := g.side(ctx, agent, "answer", prompt, text); err != nil {
		return "", err
	}
	return text, nil
}

// side runs one check and reports it to observe, so the trace has a line for
// every check, refused ones included. An error from observe stops the step:
// a check that cannot be traced is not passed.
func (g *guarded) side(ctx context.Context, agent, side, prompt, answer string) error {
	v, err := g.rail.Check(ctx, prompt, answer)
	if err != nil {
		// A cancelled review is a cancellation, not a guardrail failure.
		if ctx.Err() != nil {
			return ctx.Err()
		}
		down := asDown(err)
		if g.observe != nil {
			if oerr := g.observe(agent, side, v, down.Reason); oerr != nil {
				return oerr
			}
		}
		return down
	}
	if g.observe != nil {
		if oerr := g.observe(agent, side, v, ""); oerr != nil {
			return oerr
		}
	}
	if v.Unsafe() {
		return Refuse(unsafeReason(v))
	}
	return nil
}

// asDown turns any check error into a named guardrail refusal, so a failed
// check is never mistaken for a safe one.
func asDown(err error) *service.Down {
	var d *service.Down
	if errors.As(err, &d) && d != nil && d.Service == Service {
		return d
	}
	if err == nil {
		return Refuse("unreachable")
	}
	return Refuse(err.Error())
}
