// Package tool runs a deterministic tool on the base and the change and
// turns a difference into evidence (ADR-0040).
//
// A tool plugs in through Adapter: which files it applies to, the command
// Proven builds, and how one side's output is read. Sample replays a recorded
// outcome and starts no container. Runner runs the command in the Prover's
// container, with no network and no secret (ADR-0025), under the time limit
// from PROVEN_TOOL_TIMEOUT (ADR-0013). A tool that cannot run is named.
// The race detector is the first tool. govulncheck is the second:
// it labels Go dependency findings with reachability (ADR-0043). Rerun runs one
// tool again on a patched copy of the change, for the whole-change attempt (ADR-0050).
package tool

import (
	"context"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Adapter is one deterministic tool (ADR-0040).
// The runner compares the two sides. The adapter does not.
type Adapter interface {
	// ID is the tool's stable name, for example "go-race".
	ID() string
	// Version is the tool's version when the adapter knows it.
	// Empty means the runner records "unknown" rather than guessing.
	Version() string
	// Applies reports whether the tool looks at these changed paths.
	Applies(paths []string) bool
	// Argv is the command Proven runs, with no shell (ADR-0025).
	// A non-empty reason means there is no command and nothing is started.
	Argv(paths []string) (argv []string, reason string)
	// Read classifies one side's output. It is called only after that side ran.
	Read(output string) Reading
}

// Reading is what one side's output means.
type Reading struct {
	Present bool
	Output  string
}

// Request is one comparison of a tool on the base and the change.
type Request struct {
	Paths     []string
	BaseDir   string
	ChangeDir string
}

// Outcome is the comparison. Proves is true only when the tool answered and
// the change shows a result the base does not. A tool that did not answer
// has Answered false and a Reason that names it (ADR-0040).
type Outcome struct {
	Tool            string
	Version         string
	Command         string
	Answered        bool
	Skipped         bool
	Reason          string
	PresentOnChange bool
	PresentOnBase   bool
	ChangeOutput    string
	BaseOutput      string
}

// Proves reports whether this outcome is evidence of the change.
// A result the base already had is the base, not the change (ADR-0040).
func (o Outcome) Proves() bool {
	return Introduced(o.Answered, o.PresentOnChange, o.PresentOnBase)
}

// Introduced reports whether a tool result is evidence of the change.
// Answered is required: a side that did not run is not a clean result.
// Mutated field of the negative control: PresentOnBase must keep a result
// that is also on the base from counting.
func Introduced(answered, presentOnChange, presentOnBase bool) bool {
	return answered && presentOnChange && !presentOnBase
}

// Evidence is the proof a proved outcome carries, or nil when it is not proof.
func (o Outcome) Evidence() *candidate.ToolEvidence {
	if !o.Proves() {
		return nil
	}
	version := o.Version
	if version == "" {
		version = "unknown"
	}
	return &candidate.ToolEvidence{
		Tool:            o.Tool,
		Version:         version,
		Command:         o.Command,
		Output:          o.ChangeOutput,
		BaseOutput:      o.BaseOutput,
		PresentOnChange: true,
		PresentOnBase:   false,
	}
}

// Candidate is the finding a proved outcome reports, located at path.
// The second result is false when the outcome is not evidence.
func (o Outcome) Candidate(path string) (candidate.Candidate, bool) {
	ev := o.Evidence()
	if ev == nil {
		return candidate.Candidate{}, false
	}
	return candidate.Candidate{
		Risk: candidate.Risk{
			ID:        o.Tool,
			RuleID:    o.Tool,
			Title:     "tool result introduced",
			Message:   o.Tool + " reports a result on the change and none on the base.",
			Severity:  "ERROR",
			Path:      path,
			StartLine: 1,
			EndLine:   1,
		},
		Evidence: []candidate.Evidence{{Kind: candidate.KindTool, Tool: ev}},
	}, true
}

// Side is one checkout's run of the tool command.
// Status is one of the testrun.Command* values.
type Side struct {
	Output string
	Status string
}

// Host runs the command in one checkout.
// The live host is the Prover's container (ADR-0025). A test supplies a fake.
type Host interface {
	Run(ctx context.Context, dir string, argv []string) (Side, error)
}

// Runner runs one Adapter on the base and the change.
type Runner struct {
	tool Adapter
	host Host
}

// NewRunner returns a runner for one tool. host may be nil: Run then names
// the tool as unable to run and starts nothing.
func NewRunner(tool Adapter, host Host) *Runner {
	return &Runner{tool: tool, host: host}
}

// Run compares the tool on the change and on the base.
// The change runs first, as a test run does, so a change that does not
// answer does not start the base. A cancelled context is returned as that
// error. Every other failure is an Outcome whose Reason names the tool.
func (r *Runner) Run(ctx context.Context, req Request) (Outcome, error) {
	if err := ctx.Err(); err != nil {
		return Outcome{}, err
	}
	if r == nil || r.tool == nil {
		return Outcome{Reason: "tool cannot run: no adapter"}, nil
	}
	a := r.tool
	out := Outcome{Tool: a.ID(), Version: a.Version()}
	if out.Version == "" {
		// A live race run does not start a second container to ask for the
		// version. Unknown is the honest value; a guess would be a lie.
		out.Version = "unknown"
	}
	if !a.Applies(req.Paths) {
		out.Skipped = true
		out.Reason = a.ID() + " does not apply"
		return out, nil
	}
	argv, reason := a.Argv(req.Paths)
	if reason != "" {
		out.Reason = a.ID() + " cannot run: " + reason
		return out, nil
	}
	out.Command = strings.Join(argv, " ")
	if r.host == nil {
		out.Reason = a.ID() + " cannot run: no container"
		return out, nil
	}
	if req.BaseDir == "" || req.ChangeDir == "" {
		out.Reason = a.ID() + " cannot run: base directory and change directory are required"
		return out, nil
	}
	change, err := r.host.Run(ctx, req.ChangeDir, argv)
	if err != nil {
		if ctx.Err() != nil {
			return Outcome{}, ctx.Err()
		}
		out.Reason = a.ID() + " cannot run: " + err.Error()
		return out, nil
	}
	if why := unrunReason(change); why != "" {
		out.Reason = a.ID() + " cannot run on the change: " + why
		return out, nil
	}
	base, err := r.host.Run(ctx, req.BaseDir, argv)
	if err != nil {
		if ctx.Err() != nil {
			return Outcome{}, ctx.Err()
		}
		out.Reason = a.ID() + " cannot run: " + err.Error()
		return out, nil
	}
	if why := unrunReason(base); why != "" {
		out.Reason = a.ID() + " cannot run on the base: " + why
		return out, nil
	}
	changeRead := a.Read(change.Output)
	baseRead := a.Read(base.Output)
	out.Answered = true
	out.PresentOnChange = changeRead.Present
	out.PresentOnBase = baseRead.Present
	out.ChangeOutput = changeRead.Output
	out.BaseOutput = baseRead.Output
	return out, nil
}

// unrunReason says why a side is not an answer, or "" when the command ran.
// A timeout is not an answer even when the output already contains a result:
// a run past its limit is not proof (ADR-0040).
func unrunReason(side Side) string {
	switch side.Status {
	case testrun.CommandPass, testrun.CommandFail, "":
		return ""
	case testrun.CommandTimeout:
		return "timeout"
	default:
		text := strings.TrimSpace(side.Output)
		if text == "" {
			return side.Status
		}
		line := text
		if i := strings.IndexByte(text, '\n'); i >= 0 {
			line = text[:i]
		}
		return side.Status + ": " + line
	}
}
