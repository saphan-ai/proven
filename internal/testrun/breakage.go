package testrun

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/service"
)

// RunBreakage is the label prefix of the runs that show a patch broke nothing it touched
// (ADR-0050 point 3): one run for each package a patched file belongs to.
const RunBreakage = "breakage"

// BreakageRequest asks for the runs that show a patch broke nothing in the places it touched.
// ChangeDir is the checkout the patch applies to; it is never written.
type BreakageRequest struct {
	ChangeDir string
	Edits     []Edit
}

// BreakageOutcome is the record of those runs. Runs is empty when the patch was refused, and
// Refused then says why; no code ran in that case.
type BreakageOutcome struct {
	Runs    []FixRun
	Refused string
}

// BreakageRunner runs every package a patch touched on the patched change. Container runs them
// in the Prover's container; Sample replays recorded package runs.
type BreakageRunner interface {
	RunBreakage(ctx context.Context, req BreakageRequest) (BreakageOutcome, error)
}

// manifestFiles are the files that define a module: an edit to one can change what every package
// of the module builds against, so the whole module is run.
var manifestFiles = map[string]bool{"go.mod": true, "go.sum": true, "go.work": true, "go.work.sum": true}

// BreakageArgvs builds the commands that show a patch broke nothing it touched, from the edit
// paths alone: go test for the package of each edited file, once each, in name order. A file
// under testdata belongs to the package above it, and an edit to a module manifest runs every
// package below it. The words are Proven's; no edit text reaches a command. The paths are the
// ones RefuseEdits has already checked.
func BreakageArgvs(edits []Edit) [][]string {
	targets := map[string]bool{}
	for _, e := range edits {
		p := path.Clean(filepath.ToSlash(e.Path))
		dir := path.Dir(p)
		segs := strings.Split(dir, "/")
		for i, s := range segs {
			if s == "testdata" {
				dir = strings.Join(segs[:i], "/")
				if dir == "" {
					dir = "."
				}
				break
			}
		}
		switch {
		case manifestFiles[path.Base(p)] && dir == ".":
			targets["./..."] = true
		case manifestFiles[path.Base(p)]:
			targets["./"+dir+"/..."] = true
		case dir == ".":
			targets["."] = true
		default:
			targets["./"+dir] = true
		}
	}
	if targets["./..."] {
		// The whole module covers every package below it.
		targets = map[string]bool{"./...": true}
	}
	names := make([]string, 0, len(targets))
	for t := range targets {
		names = append(names, t)
	}
	sort.Strings(names)
	argvs := make([][]string, 0, len(names))
	for _, n := range names {
		argvs = append(argvs, []string{"go", "test", jsonFlag, n})
	}
	return argvs
}

// unresolvable says why go test, run at the module root, cannot resolve target, or "" when it
// can. A target in a nested module belongs to another module, and a directory with no Go file
// is in no package: an edit to a script, a document or a config there touched no package.
func unresolvable(root, target string) string {
	if target == "./..." {
		return ""
	}
	dir := strings.TrimSuffix(strings.TrimPrefix(target, "./"), "/...")
	for d := dir; d != "." && d != ""; d = path.Dir(d) {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(d), "go.mod")); err == nil {
			return "the edited file belongs to the nested module at " + d + ", which this module's go test does not run"
		}
	}
	if strings.HasSuffix(target, "/...") {
		return ""
	}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
	if err != nil {
		return "the edited file's directory " + dir + " is not in the patched copy"
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			return ""
		}
	}
	return "the edited file is in " + dir + ", which holds no Go package"
}

// breakageLabel names one breakage run by the package it covers.
func breakageLabel(argv []string) string { return RunBreakage + " " + argv[len(argv)-1] }

// RunBreakage runs, in the container, every package the patch touched on a scratch copy of the
// change with the patch applied (ADR-0050 point 3). The review's own trees are never written.
// A run that does not end passed is a run that did not show nothing broke: the caller reads
// every Result.
func (c *Container) RunBreakage(ctx context.Context, req BreakageRequest) (BreakageOutcome, error) {
	if c == nil || c.runtime == "" {
		return BreakageOutcome{}, &service.Down{Service: "prover", Reason: "Prover off: no container runtime"}
	}
	if err := ctx.Err(); err != nil {
		return BreakageOutcome{}, err
	}
	if req.ChangeDir == "" {
		return BreakageOutcome{Refused: "change directory is required"}, nil
	}
	if reason := RefuseEdits(req.ChangeDir, req.Edits, c.cfg.MaxWrittenBytes); reason != "" {
		return BreakageOutcome{Refused: reason}, nil
	}
	if linkedCheckout(req.ChangeDir) {
		return BreakageOutcome{Refused: "a checkout is a symbolic link and is not mounted"}, nil
	}
	patched, err := stagePatched(req.ChangeDir, req.Edits, c.cfg.MaxWrittenBytes)
	if err != nil {
		return BreakageOutcome{}, err
	}
	defer os.RemoveAll(patched)

	var out BreakageOutcome
	for _, argv := range BreakageArgvs(req.Edits) {
		if why := unresolvable(patched, argv[len(argv)-1]); why != "" {
			// go test would fail on a target it cannot resolve, and the record would then say the
			// package failed. It is a run that cannot happen, and the attempt stays rejected.
			out.Runs = append(out.Runs, FixRun{
				Label: breakageLabel(argv), Command: displayCommand(argv), Result: ResultNotRun,
				ExitCode: -1, Output: why,
			})
			continue
		}
		text, kind, code := c.sideExit(ctx, patched, argv)
		if kind == kindCancelled {
			return BreakageOutcome{}, withRemoval(ctx.Err(), text)
		}
		if down := c.absentImage(text, kind); down != nil {
			return BreakageOutcome{}, down
		}
		out.Runs = append(out.Runs, FixRun{
			Label: breakageLabel(argv), Command: displayCommand(argv), Result: resultOf(kind),
			ExitCode: code, Output: trimOutput(text),
		})
	}
	return out, nil
}

// RunBreakage replays the recorded package runs of the sample's fixes. It executes no code, and
// it checks the patch against the real change tree as RunFix does. A package no recording ran
// is a run that did not happen, and is reported so.
func (s *Sample) RunBreakage(_ context.Context, req BreakageRequest) (BreakageOutcome, error) {
	if reason := RefuseEdits(req.ChangeDir, req.Edits, 0); reason != "" {
		return BreakageOutcome{Refused: reason}, nil
	}
	var out BreakageOutcome
	for _, argv := range BreakageArgvs(req.Edits) {
		cmd := displayCommand(argv)
		run := FixRun{Label: breakageLabel(argv), Command: cmd, Result: ResultNotRun, ExitCode: -1,
			Output: "the sample records no run of this package"}
		for _, rec := range s.fixes {
			for _, r := range rec.Runs {
				if r.Label != RunPackage || r.Command != cmd || r.Result != ResultPassed || !RecordedPass(r.Command, r.Output) {
					continue
				}
				run.Result, run.ExitCode, run.Output = ResultPassed, r.ExitCode, r.Output
			}
		}
		out.Runs = append(out.Runs, run)
	}
	return out, nil
}
