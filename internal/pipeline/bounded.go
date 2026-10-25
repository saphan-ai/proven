package pipeline

import (
	"context"
	"fmt"

	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/trace"
)

// traceKindBounded is the trace line that says, once per Planner run, whether bounded steps ran; traceKindStep is
// the line of one step (ADR-0048).
const (
	traceKindBounded = "bounded-steps"
	traceKindStep    = "bounded-step"
)

// plan runs the Planner for one reviewer run, whole or in bounded steps (ADR-0048 point 1). With the engine off,
// or a change at or below the threshold, it is planner.Plan exactly as before. Either way the trace says which,
// and why, by name.
func (p *Pipeline) plan(ctx context.Context, root string, gen generator.Generator, run string) ([]candidate.Risk, error) {
	s := p.markov
	if !s.On {
		if err := p.boundedLine(run, fmt.Sprintf("bounded steps did not run: the engine is off (%s is off); the Planner read the change whole", markov.EnvSwitch)); err != nil {
			return nil, err
		}
		return planner.Plan(ctx, root, gen)
	}
	files, err := change.Read(root)
	if err != nil {
		return nil, err
	}
	size := planner.Size(files)
	// ADR-0048 point 1: only a change above the size set in the settings is cut; at the threshold it is still whole.
	if size <= s.Threshold {
		if err := p.boundedLine(run, fmt.Sprintf("bounded steps did not run: the change is %d estimated tokens, not above the threshold %d (%s); the Planner read the change whole", size, s.Threshold, markov.EnvThreshold)); err != nil {
			return nil, err
		}
		return planner.Plan(ctx, root, gen)
	}
	risks, steps, err := planner.PlanBounded(ctx, files, gen, s)
	// The steps that ran are traced even when a later one failed, so the trace shows how far the review got.
	if terr := p.boundedLine(run, fmt.Sprintf("bounded steps ran: %d of %d steps, unit %s, the change is %d estimated tokens, above the threshold %d; carried-state bound %d estimated tokens",
		len(steps), stepsPlanned(steps), s.Unit, size, s.Threshold, s.Carry)); terr != nil {
		return nil, terr
	}
	for _, st := range steps {
		if terr := p.trace.Append(trace.Event{
			Kind: traceKindStep, Agent: "planner", Run: run, Path: st.Part,
			Text: fmt.Sprintf("step %d of %d: prompt %d estimated tokens; carried state %d estimated tokens, %d items; %d dropped by the bound from the state handed on; %d risks named",
				st.Index, st.Steps, st.PromptTokens, st.CarriedTokens, st.Carried, st.Dropped, st.Risks),
		}); terr != nil {
			return nil, terr
		}
	}
	return risks, err
}

// stepsPlanned is the number of parts the change was cut into, read from the steps that ran.
func stepsPlanned(steps []planner.Step) int {
	if len(steps) == 0 {
		return 0
	}
	return steps[0].Steps
}

// boundedLine writes the one line that says whether bounded steps ran. An empty finding id keeps the world fold
// from reading it as a finding, as with the snapshot line.
func (p *Pipeline) boundedLine(run, text string) error {
	return p.trace.Append(trace.Event{Kind: traceKindBounded, Agent: "pipeline", Run: run, Text: text})
}

// fixCarry is the carried-state bound the Fixer gets: the setting when the engine is on, zero (every earlier attempt
// in the prompt, as before) when it is off (ADR-0048 point 2).
func (p *Pipeline) fixCarry() int {
	if !p.markov.On {
		return 0
	}
	return p.markov.Carry
}
