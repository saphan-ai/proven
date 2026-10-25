package tool

import (
	"context"
	"strings"
)

// ByID returns the adapter that produces a tool's evidence, by the name the evidence carries.
// Only a tool whose evidence is a result the change introduced has one here (ADR-0040):
// govulncheck labels findings and is not evidence of its own. The second result is false for
// any other name, and a caller must then say the tool cannot be run again, never guess a command.
func ByID(id string) (Adapter, bool) {
	if id == (Race{}).ID() {
		return Race{}, true
	}
	return nil, false
}

// Rerunning is one run of a tool's command in one checkout (ADR-0050 point 2). Answered is false
// when the tool did not answer, and Reason then names why; Present is true when the output
// still carries the result. Command is the command Proven built, never one a model wrote.
type Rerunning struct {
	Command  string
	Answered bool
	Present  bool
	Output   string
	Reason   string
}

// Rerun runs the adapter's command for paths in dir and reads the output with the adapter, as
// Runner does for one side. A tool that does not apply, has no command, has no host, or ran past
// its limit has not answered, and is never read as a clean result. A cancelled context is returned
// as that error.
func Rerun(ctx context.Context, a Adapter, host Host, dir string, paths []string) (Rerunning, error) {
	if err := ctx.Err(); err != nil {
		return Rerunning{}, err
	}
	var out Rerunning
	if !a.Applies(paths) {
		out.Reason = a.ID() + " does not apply"
		return out, nil
	}
	argv, reason := a.Argv(paths)
	if reason != "" {
		out.Reason = a.ID() + " cannot run: " + reason
		return out, nil
	}
	out.Command = strings.Join(argv, " ")
	if host == nil {
		out.Reason = a.ID() + " cannot run: no container"
		return out, nil
	}
	side, err := host.Run(ctx, dir, argv)
	if err != nil {
		if ctx.Err() != nil {
			return Rerunning{}, ctx.Err()
		}
		out.Reason = a.ID() + " cannot run: " + err.Error()
		return out, nil
	}
	if why := unrunReason(side); why != "" {
		out.Reason = a.ID() + " cannot run on the patched change: " + why
		return out, nil
	}
	read := a.Read(side.Output)
	out.Answered, out.Present, out.Output = true, read.Present, read.Output
	return out, nil
}
