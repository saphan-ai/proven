package tool

import (
	"context"

	"github.com/saphan-ai/proven/internal/limit"
)

// Recorded is one tool outcome stored for sample mode.
// The outputs are what the tool printed. Run classifies them with the
// adapter's Read, so a recording cannot claim a race the output does not show.
// Unrun is "change" or "base" when that side did not answer; UnrunDetail
// says why, for example "timeout".
type Recorded struct {
	Version      string
	Command      string
	BaseOutput   string
	ChangeOutput string
	Unrun        string
	UnrunDetail  string
}

// The recorded race scenario (ADR-0040 point 5). The change's output carries
// the detector's banner and the base's does not. The version is the module's
// go line, not a measurement from a container. Sample mode runs no container.
const (
	raceChangeOutput = "WARNING: DATA RACE\nWrite at 0x00c000014090 by goroutine 8:\nFAIL\n"
	raceBaseOutput   = "ok  \texample/counter\t0.012s\n"
	raceCommand      = "go test -race ./counter"
	raceVersion      = "go1.26"
)

// RaceIntroduced is the sample in which the change races and the base does not.
func RaceIntroduced() Recorded {
	return Recorded{
		Version:      raceVersion,
		Command:      raceCommand,
		BaseOutput:   raceBaseOutput,
		ChangeOutput: raceChangeOutput,
	}
}

// RaceOnBoth is the sample in which the base races too.
// That is not evidence of the change (ADR-0040).
func RaceOnBoth() Recorded {
	return Recorded{
		Version:      raceVersion,
		Command:      raceCommand,
		BaseOutput:   raceChangeOutput,
		ChangeOutput: raceChangeOutput,
	}
}

// RaceTimeout is the sample in which the change did not answer.
// The reason names a timeout, and the outcome is not a clean run.
func RaceTimeout() Recorded {
	return Recorded{
		Version:     raceVersion,
		Command:     raceCommand,
		Unrun:       "change",
		UnrunDetail: "timeout",
	}
}

// Sample returns one recorded outcome. It executes no code and starts no container.
type Sample struct {
	tool Adapter
	rec  Recorded
}

// NewSample indexes one recording for one tool.
func NewSample(tool Adapter, rec Recorded) *Sample {
	return &Sample{tool: tool, rec: rec}
}

// Run classifies the recording. It does not read PROVEN_TOOL_TIMEOUT:
// sample mode uses no container, and the bound is the sample adapter bound.
func (s *Sample) Run(ctx context.Context) (Outcome, error) {
	// The sample keeps the live deadline behaviour of the other sample adapters.
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	if s == nil || s.tool == nil {
		return Outcome{Reason: "tool cannot run: no adapter"}, nil
	}
	out := Outcome{Tool: s.tool.ID(), Version: s.rec.Version, Command: s.rec.Command}
	if out.Version == "" {
		out.Version = "unknown"
	}
	if s.rec.Unrun != "" {
		// A recorded miss is not classified from the output: the side did not answer,
		// even when a partial output happens to contain the banner.
		detail := s.rec.UnrunDetail
		if detail == "" {
			detail = "did not answer"
		}
		out.Reason = s.tool.ID() + " cannot run on the " + s.rec.Unrun + ": " + detail
		return out, nil
	}
	change := s.tool.Read(s.rec.ChangeOutput)
	base := s.tool.Read(s.rec.BaseOutput)
	out.Answered = true
	out.PresentOnChange = change.Present
	out.PresentOnBase = base.Present
	out.ChangeOutput = change.Output
	out.BaseOutput = base.Output
	return out, nil
}
