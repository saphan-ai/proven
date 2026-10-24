// Package pipeline runs the six agents in order and implements the review.
//
// It is the reviewer behind review.Open. One Review call reads the change, lets
// the Planner name candidate risks, has the Investigator pick a query template and the
// Prover write a test for each, lets the Skeptic try to refute every finding,
// has the Judge keep only what carries evidence (ADR-0002), adds the dependency
// check against the OSV table and the EPSS scores, and hands the result to the
// Writer. Every step writes to the trace, which is then folded into H-MEM and
// the World Model.
//
// The pipeline also owns what wraps the agents: the guardrail on every prompt
// and answer (ADR-0005), the record of every model call, the doubled review
// with a second model (ADR-0009, ADR-0012), the partial review a deadline
// leaves behind, and the redaction of every configured secret from the trace,
// the recorded models, a failure's reason and every finding, because a service
// can echo a value it was sent (ADR-0005, ADR-0013 point 4).
//
// Live mode reads a directory that already has base/ and change/, or a git
// repository. For a repository it writes those two trees for the review and
// removes them when the review returns. A tracked symlink is not written into
// either tree; the trace names it and the draft lists the path. Sample mode
// still loads a fixture and does not write trees. An empty secret is named in
// the review's failed services before any agent runs (ADR-0013 point 4a).
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/investigator"
	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/agent/prover"
	"github.com/saphan-ai/proven/internal/agent/skeptic"
	"github.com/saphan-ai/proven/internal/agent/writer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
	"github.com/saphan-ai/proven/internal/trace"
	"github.com/saphan-ai/proven/internal/world"
)

// Options select the adapters. TestRun, when set, replaces the sample or live test-run adapter.
// Generator, when set, is the live generator. Sample mode does not use it and does not read Env.
type Options struct {
	Sample    bool
	LogDir    string
	Env       func(string) string
	TestRun   testrun.Adapter
	Generator generator.Generator
	// Special, when set, is the generator pointed at the special-review model.
	// A doubled review uses it for the second planner and the second skeptic.
	Special generator.Generator
	// ProverIdle, when set, is named on the review as the Prover being off.
	// The review runs no test code in that case.
	ProverIdle string
	// LiveQuery, when set, is the Zetaris adapter. Sample mode does not use it.
	LiveQuery query.Adapter
	// StateDir is the durable directory for H-MEM and the world model.
	// Empty uses a state directory beside the logs, never a path inside the module.
	StateDir string
	// Guardrail checks every prompt and every answer. Sample mode with none
	// uses the safe sample, unless the fixture records a verdict. A live
	// pipeline with none is refused.
	Guardrail guardrail.Adapter
	// GuardOff writes one "no guardrail" line and checks nothing.
	GuardOff bool
	// SkepticModel is the skeptic's configured model. Empty means the Nemotron default.
	// Sample mode does not read the environment to fill it.
	SkepticModel string
	// OSVTable is the package table the dependency check queries.
	// Empty means osv.DefaultTable. Sample mode leaves this empty and does not read the environment.
	OSVTable string
	// EPSSView is the view a finding's CVE is scored from.
	// Empty means epss.DefaultView. Sample mode leaves this empty and does not read the environment.
	EPSSView string
	// MaxWrittenBytes is the loaded container limit for a test the prover writes.
	// Zero means DefaultMaxWrittenBytes. Sample mode leaves it zero.
	MaxWrittenBytes int
	// Down names services that are already down when the review starts.
	// An empty secret is recorded here so the review names the service even
	// when a later check never calls it (ADR-0013 point 4a, ADR-0005 point 5).
	Down []service.Down
	// Redactor replaces the configured secrets in the trace, in a failure's
	// reason, in the models a review records and in every finding. A live
	// pipeline with none is refused; nil in sample mode redacts nothing, which is
	// right because sample mode loads no secret.
	Redactor *redact.Redactor
}

// Pipeline is one reviewer. Build makes it; Review runs it once.
type Pipeline struct {
	sample       bool
	testRun      testrun.Adapter
	proverIdle   string
	gen          generator.Generator
	special      generator.Generator
	liveQuery    query.Adapter
	trace        *trace.Writer
	rail         guardrail.Adapter
	guardOff     bool
	skepticModel string
	osvTable     string
	epssView     string
	env          func(string) string
	maxWritten   int
	down         []service.Down
	atBoundary   func(stage string)
	mem          *hmem.Store
	world        *world.Store
	// workDir is the run's working area. Live git trees are written here, not
	// inside the repository under review, and removed when the review returns.
	workDir string
	// redactor replaces the configured secrets in text a service returned.
	redactor *redact.Redactor
}

// Build opens the trace. Sample mode uses the recorded adapters. Those
// adapters do not dial, so sample mode holds no HTTP client.
func Build(opts Options) (*Pipeline, error) {
	// opts.Env is not read here. review.Open loads live settings before Build.
	// adapters reads Env when a live review left TestRun unset.
	// Sample mode reads no key.
	if opts.LogDir == "" {
		return nil, fmt.Errorf("pipeline: logs directory is empty")
	}
	// ADR-0005: "off" and an adapter together are a contradiction, refused here
	// rather than settled silently one way.
	if opts.GuardOff && opts.Guardrail != nil {
		return nil, fmt.Errorf("pipeline: the guardrail is off and an adapter was set")
	}
	// ADR-0005, ADR-0013 point 4: a live pipeline holds keys, so it never runs
	// with nothing to redact them; the refusal is made here, before any file is
	// opened, like the refusal of a live pipeline with no guardrail.
	if !opts.Sample && opts.Redactor == nil {
		return nil, fmt.Errorf("pipeline: a live pipeline needs a redactor and none was set")
	}
	tw, err := trace.Open(opts.LogDir)
	if err != nil {
		return nil, err
	}
	// ADR-0005, ADR-0013 point 4: one redactor, built by review.Open from the values the live settings
	// loaded, stands at every place a service's text becomes trace or result text.
	// Build does not read the environment to make it, so sample mode reads no
	// setting; with none given, nothing is redacted.
	rd := opts.Redactor
	tw.SetRedactor(rd)
	stateDir := opts.StateDir
	if stateDir == "" {
		var derr error
		stateDir, derr = defaultStateDir(opts.LogDir)
		if derr != nil {
			return nil, derr
		}
	}
	// The stores refuse a state directory inside the proven module, so they are
	// told where the module is.
	root := moduleRoot()
	mem, err := hmem.Open(hmem.Options{
		Dir: filepath.Join(stateDir, "hmem"), LogDir: opts.LogDir, RepoRoot: root,
	})
	if err != nil {
		return nil, err
	}
	wm, err := world.Open(world.Options{
		Dir: filepath.Join(stateDir, "world"), LogDir: opts.LogDir, RepoRoot: root, Memory: mem,
	})
	if err != nil {
		return nil, err
	}
	area, aerr := workArea(opts.LogDir)
	if aerr != nil {
		return nil, aerr
	}
	p := &Pipeline{
		sample: opts.Sample, testRun: opts.TestRun, proverIdle: opts.ProverIdle,
		gen: opts.Generator, special: opts.Special, liveQuery: opts.LiveQuery, trace: tw, rail: opts.Guardrail, guardOff: opts.GuardOff,
		skepticModel: opts.SkepticModel, osvTable: opts.OSVTable, epssView: opts.EPSSView, env: opts.Env, maxWritten: opts.MaxWrittenBytes,
		down: append([]service.Down(nil), opts.Down...), mem: mem, world: wm,
		workDir: area, redactor: rd,
	}
	// ADR-0005: the guardrail is on by default. Sample mode falls back to the
	// sample guardrail, which answers safe; a live pipeline without one is
	// refused rather than run unchecked.
	if !opts.GuardOff && p.rail == nil {
		if opts.Sample {
			p.rail = guardrail.Safe{}
		} else {
			return nil, fmt.Errorf("pipeline: the guardrail is on and no adapter was set")
		}
	}
	return p, nil
}

// writtenLimit is the size cap for a test the prover writes.
// Zero means the container default, which sample mode keeps.
func (p *Pipeline) writtenLimit() int {
	if p == nil || p.maxWritten <= 0 {
		return testrun.DefaultMaxWrittenBytes
	}
	return p.maxWritten
}

// boundary lets a test cancel the context at one named stage. Production leaves it unset.
func (p *Pipeline) boundary(stage string) {
	if p.atBoundary != nil {
		p.atBoundary(stage)
	}
}

// Request is the part of a review request the pipeline reads. Target is the
// directory under review. BaselineCommit and HeadCommit name a git change.
// With no baseline, a directory that has base/ and change/ is read as those
// trees, and a git repository without them is refused. Include and Exclude
// narrow the changed files the review reports on.
type Request struct {
	Target         string
	BaselineCommit string
	HeadCommit     string
	Include        []string
	Exclude        []string
	// DoubleRun is set when the class is doubled and a special-review model is configured
	// (ADR-0012).
	DoubleRun bool
	// DefaultModel and SpecialModel are the configured reviewer names.
	DefaultModel string
	SpecialModel string
}

// Review runs planner, investigator, prover, skeptic, judge and writer.
// Findings whose path is outside the selected files are dropped after the agents run.
// When ctx ends before the Judge has decided, the draft is partial and holds only
// judged findings. After a review with no error the trace is folded into the World
// Model and H-MEM; if that fails, the review fails.
func (p *Pipeline) Review(ctx context.Context, req Request) (d candidate.Draft, err error) {
	if p.trace != nil {
		defer p.trace.Close()
	}
	// Fold only a review that ended without an error. This defer runs before
	// the trace is closed, and reads what the review wrote.
	defer func() {
		if err != nil {
			return
		}
		if ferr := p.fold(); ferr != nil {
			d = candidate.Draft{}
			err = ferr
		}
	}()
	d, err = p.review(ctx, req)
	if err != nil {
		return d, err
	}
	// ADR-0005, ADR-0013 point 4: the draft is the last place model and service text
	// can reach the result, so it is redacted once here, after every agent has run.
	d.Findings = p.redactFindings(d.Findings)
	// A review run without the guardrail says so in its result, not only in its
	// trace: a reader of the result alone must not take it for a guarded review.
	if p.guardOff {
		d.GuardrailStatement = candidate.GuardrailOffStatement
	}
	return d, nil
}

// redactFindings returns a copy of in with the configured secrets replaced in
// every string of every finding and its evidence. The copy shares no memory
// with in, because the same candidates were already written to the trace and
// the stores.
func (p *Pipeline) redactFindings(in []candidate.Candidate) []candidate.Candidate {
	if in == nil {
		return nil
	}
	rd := p.redactor
	out := make([]candidate.Candidate, len(in))
	for i, c := range in {
		c.ID, c.RuleID = rd.Text(c.ID), rd.Text(c.RuleID)
		c.Title, c.Message = rd.Text(c.Title), rd.Text(c.Message)
		c.Severity, c.Path = rd.Text(c.Severity), rd.Text(c.Path)
		c.DropReason, c.ProverNote = rd.Text(c.DropReason), rd.Text(c.ProverNote)
		c.Runs, c.Models = rd.Texts(c.Runs), rd.Texts(c.Models)
		r := &c.Refutation
		r.Reason, r.TestPath, r.TestName = rd.Text(r.Reason), rd.Text(r.TestPath), rd.Text(r.TestName)
		// ADR-0032: a refutation names a query template and its values, never SQL;
		// the values are redacted like any other text a review keeps.
		r.Query.Template = rd.Text(r.Query.Template)
		if r.Query.Params != nil {
			params := make(map[string]string, len(r.Query.Params))
			for k, v := range r.Query.Params {
				params[k] = rd.Text(v)
			}
			r.Query.Params = params
		}
		r.QueryRows = redactRows(rd, r.QueryRows)
		if c.Evidence != nil {
			ev := make([]candidate.Evidence, len(c.Evidence))
			for j, e := range c.Evidence {
				ev[j] = candidate.Evidence{Kind: rd.Text(e.Kind)}
				if t := e.Test; t != nil {
					cp := *t
					cp.Path, cp.Name = rd.Text(t.Path), rd.Text(t.Name)
					cp.Command, cp.Output = rd.Text(t.Command), rd.Text(t.Output)
					ev[j].Test = &cp
				}
				if q := e.Query; q != nil {
					cp := *q
					cp.SQL, cp.Engine, cp.Snapshot = rd.Text(q.SQL), rd.Text(q.Engine), rd.Text(q.Snapshot)
					cp.Columns, cp.Rows = rd.Texts(q.Columns), redactRows(rd, q.Rows)
					ev[j].Query = &cp
				}
				if t := e.Tool; t != nil {
					// ADR-0040: a tool's output is checkout text and can carry a secret.
					cp := *t
					cp.Tool, cp.Version = rd.Text(t.Tool), rd.Text(t.Version)
					cp.Command, cp.Output, cp.BaseOutput = rd.Text(t.Command), rd.Text(t.Output), rd.Text(t.BaseOutput)
					ev[j].Tool = &cp
				}
			}
			c.Evidence = ev
		}
		out[i] = c
	}
	return out
}

// redactRows returns a copy of rows with rd applied to every cell. A nil slice stays nil.
func redactRows(rd *redact.Redactor, rows [][]string) [][]string {
	if rows == nil {
		return nil
	}
	out := make([][]string, len(rows))
	for i, row := range rows {
		out[i] = rd.Texts(row)
	}
	return out
}

// review is the body of Review. Every agent step follows one pattern: an error
// from a service that is down is noted and the review goes on without it; a
// context that has ended returns the partial draft through cut; any other
// error stops the review.
func (p *Pipeline) review(ctx context.Context, req Request) (candidate.Draft, error) {
	// A git repository with neither a baseline commit nor prepared trees has
	// nothing for the agents to read. Say what to pass, rather than failing
	// later as a missing fixture file.
	if !p.sample {
		if err := p.refuseBareRepository(req); err != nil {
			return candidate.Draft{}, err
		}
	}
	files, err := changedFiles(req)
	if err != nil {
		return candidate.Draft{}, err
	}
	selected, dropped, err := change.Select(files, req.Include, req.Exclude)
	if err != nil {
		return candidate.Draft{}, err
	}
	// The files left out are named on the trace, so a reader sees what the
	// review did not cover.
	if err := p.trace.Append(trace.Event{
		FindingID: "selection",
		Kind:      "selection",
		Agent:     "pipeline",
		Text:      selectionText(dropped),
	}); err != nil {
		return candidate.Draft{}, err
	}
	allow := make(map[string]struct{}, len(selected))
	for _, f := range selected {
		allow[f] = struct{}{}
	}
	keep := func(d candidate.Draft) candidate.Draft {
		d.Findings = keepFindings(d.Findings, allow)
		return d
	}

	if ctx.Err() != nil {
		return candidate.Draft{Partial: true, TracePath: p.trace.Path()}, nil
	}
	// Sample mode loads fixture.json and checks it. Live mode does not: an
	// ordinary repository has no fixture file. The trees the Planner and the
	// Prover read are fx.Root either way.
	fx, skipped, cleanup, err := p.loadTarget(req)
	if err != nil {
		return candidate.Draft{}, err
	}
	if cleanup != nil {
		defer cleanup()
	}
	if !p.sample {
		// The snapshot name has no fixture.json to come from. It is written
		// here, not only on query evidence, so a review that keeps no row
		// still records which trees it read. An empty finding id keeps the
		// world fold from treating the line as a finding.
		if err := p.trace.Append(trace.Event{
			Kind: "snapshot", Agent: "pipeline", Text: fx.Snapshot,
		}); err != nil {
			return candidate.Draft{}, err
		}
		// F31: one line per skipped symlink. The path is the tracked path.
		// Text is the reason. An empty finding id is ignored by the world
		// fold, the same way a snapshot line is.
		for _, path := range skipped {
			if err := p.trace.Append(trace.Event{
				Kind: "skipped", Agent: "pipeline", Path: path, Text: "symlink",
			}); err != nil {
				return candidate.Draft{}, err
			}
		}
	}
	gen, q, run, err := p.adapters(fx)
	if err != nil {
		return candidate.Draft{}, err
	}
	st :=&state{trace: p.trace, redactor: p.redactor, changeKey: changeKey(req), skipped: skipped}
	if p.proverIdle != "" {
		st.note(&service.Down{Service: "prover", Reason: p.proverIdle})
	}
	// ADR-0013 point 4a: name a down service even when no later step calls it.
	// The guardrail refuses the planner's call before the generator runs
	// (ADR-0005 point 5), and a review with no risks never queries Zetaris.
	for _, d := range p.down {
		noted := d
		st.note(&noted)
	}
	// Sample mode: a fixture may carry recorded guardrail verdicts, so a
	// scenario shows an unsafe prompt refused with no model server (ADR-0005).
	rail := p.rail
	if p.sample && len(fx.Guardrail) > 0 {
		sampled, serr := guardrail.NewSample(fx.Guardrail)
		if serr != nil {
			return candidate.Draft{}, serr
		}
		rail = sampled
	}
	// raw is kept unwrapped: the special run builds its own wrapped generator
	// from it, so each run's calls are guarded and recorded once.
	raw := gen
	if err := p.noteGuardOff(); err != nil {
		return candidate.Draft{}, err
	}
	// The guardrail wraps the recorder: a call the guardrail refuses never
	// reaches the model, so it is not recorded as a model call.
	gen, err = p.guard(recording{inner: raw, models: &st.models, trace: p.trace, redactor: p.redactor}, rail)
	if err != nil {
		return candidate.Draft{}, err
	}
	// The stores get the guarded generator, never the raw one: a call a store
	// ever makes is checked before and after, and recorded, like an agent call.
	p.mem.SetGenerator(gen)
	p.world.SetGenerator(gen)
	// ADR-0009, ADR-0012: a doubled review runs a second Planner and a second
	// Skeptic on the special-review model.
	var special generator.Generator
	if req.DoubleRun {
		second := p.secondGenerator(req.SpecialModel, raw)
		if second != nil {
			special, err = p.guard(recording{inner: second, models: &st.models, trace: p.trace, redactor: p.redactor}, rail)
			if err != nil {
				return candidate.Draft{}, err
			}
		}
	}

	risks, err := planner.Plan(ctx, fx.Root, gen)
	if err != nil {
		// A Planner whose service is down leaves an empty review that names
		// the service; any other Planner error stops the review.
		if d, fatal := st.soften(ctx, err); fatal != nil {
			if errors.Is(fatal, errPartial) {
				return p.cut(st, nil, allow)
			}
			return candidate.Draft{}, fatal
		} else if !d {
			return candidate.Draft{}, err
		}
		risks = nil
	}
	cands := make([]candidate.Candidate, 0, len(risks))
	for _, r := range risks {
		c := candidate.Candidate{
			Risk: r, Runs: []string{runDefault}, Models: []string{reviewerModel(req, runDefault)},
		}
		if err := st.claim(r.ID, r.Title, runDefault); err != nil {
			return candidate.Draft{}, err
		}
		cands = append(cands, c)
	}
	if special != nil {
		if ctx.Err() != nil {
			return p.cut(st, cands, allow)
		}
		more, err := planner.Plan(ctx, fx.Root, special)
		if err != nil {
			if d, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return p.cut(st, cands, allow)
				}
				return candidate.Draft{}, fatal
			} else if !d {
				return candidate.Draft{}, err
			}
			more = nil
		}
		extra := make([]candidate.Candidate, 0, len(more))
		for _, r := range more {
			extra = append(extra, candidate.Candidate{
				Risk: r, Runs: []string{runSpecial}, Models: []string{reviewerModel(req, runSpecial)},
			})
		}
		// The two runs number their risks on their own, so they are merged by
		// what the risk is. The claim names a renamed risk under both ids.
		var ids []string
		cands, ids = mergeCandidates(cands, extra)
		for i, r := range more {
			title := r.Title
			if ids[i] != r.ID {
				title = r.Title + " (special run named it " + r.ID + "; renamed " + ids[i] + ")"
			}
			if err := st.claim(ids[i], title, runSpecial); err != nil {
				return candidate.Draft{}, err
			}
		}
	}
	// After each stage, a context that has ended returns what is judged so far.
	p.boundary("planned")
	if ctx.Err() != nil {
		return p.cut(st, cands, allow)
	}

	// An agent step that fails for one candidate leaves that candidate as it
	// was; the others go on.
	// ADR-0032: the Investigator and the Skeptic name a template; the table and
	// view names come from the review's settings, read once.
	rel := p.relations(ctx)
	for i := range cands {
		next, err := investigator.Investigate(ctx, cands[i], fx.Snapshot, gen, q, rel)
		if err != nil {
			if _, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return p.cut(st, cands, allow)
				}
				return candidate.Draft{}, fatal
			}
			continue
		}
		cands[i] = next
	}
	p.boundary("investigated")
	if ctx.Err() != nil {
		return p.cut(st, cands, allow)
	}
	for i := range cands {
		// ADR-0025: the Prover's test runs in the container on the fixture's
		// base and change trees, with no network and no secret.
		next, err := prover.Prove(ctx, cands[i], gen, run, prover.Options{
			BaseDir:         filepath.Join(fx.Root, "base"),
			ChangeDir:       filepath.Join(fx.Root, "change"),
			MaxWrittenBytes: p.writtenLimit(),
		})
		if err != nil {
			if _, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return p.cut(st, cands, allow)
				}
				return candidate.Draft{}, fatal
			}
			continue
		}
		cands[i] = next
	}
	for _, c := range cands {
		if err := st.evidence(c); err != nil {
			return candidate.Draft{}, err
		}
	}
	p.boundary("proved")
	if ctx.Err() != nil {
		return p.cut(st, cands, allow)
	}

	// ADR-0006: every finding faces the Skeptic, and what it proposes is run
	// as a test or a query; its word alone is not evidence.
	for i := range cands {
		if hasRun(cands[i], runDefault) {
			next, err := skeptic.Attempt(ctx, cleared(cands[i]), gen)
			if err != nil {
				if _, fatal := st.soften(ctx, err); fatal != nil {
					if errors.Is(fatal, errPartial) {
						return p.cut(st, cands, allow)
					}
					return candidate.Draft{}, fatal
				}
				cands[i].Refutation = candidate.Refutation{Attempted: false, Reason: "not attempted"}
			} else {
				cands[i] = next
				if err := p.runRefutation(ctx, &cands[i], q, run); err != nil {
					if _, fatal := st.soften(ctx, err); fatal != nil {
						if errors.Is(fatal, errPartial) {
							return p.cut(st, cands, allow)
						}
						return candidate.Draft{}, fatal
					}
				}
			}
			if err := st.refutation(cands[i], runDefault); err != nil {
				return candidate.Draft{}, err
			}
		}
		// The special Skeptic sees only risks its own run named. Its result
		// is written to the trace under the special run; it becomes the
		// finding's refutation only when the default run did not name the risk.
		if special != nil && hasRun(cands[i], runSpecial) {
			next, err := skeptic.Attempt(ctx, cleared(cands[i]), special)
			ref := candidate.Refutation{Attempted: false, Reason: "not attempted"}
			if err != nil {
				if _, fatal := st.soften(ctx, err); fatal != nil {
					if errors.Is(fatal, errPartial) {
						return p.cut(st, cands, allow)
					}
					return candidate.Draft{}, fatal
				}
			} else {
				if err := p.runRefutation(ctx, &next, q, run); err != nil {
					if _, fatal := st.soften(ctx, err); fatal != nil {
						if errors.Is(fatal, errPartial) {
							return p.cut(st, cands, allow)
						}
						return candidate.Draft{}, fatal
					}
				}
				ref = next.Refutation
				if !hasRun(cands[i], runDefault) {
					cands[i] = next
				}
			}
			if !hasRun(cands[i], runDefault) {
				cands[i].Refutation = ref
			}
			traced := cands[i]
			traced.Refutation = ref
			if err := st.refutation(traced, runSpecial); err != nil {
				return candidate.Draft{}, err
			}
		}
	}
	if err := p.noteFamily(st); err != nil {
		return candidate.Draft{}, err
	}
	p.boundary("refuted")
	if ctx.Err() != nil {
		return p.cut(st, cands, allow)
	}

	// H-MEM is read before the Judge: a finding an earlier review of this
	// change refuted is dropped and does not come back.
	for i := range cands {
		if rec, ok := p.mem.MatchRefutation(st.changeKey, cands[i].RuleID, cands[i].Path, cands[i].Title); ok {
			cands[i].Dropped = true
			cands[i].DropReason = "suppressed: hmem " + rec.ID
			if err := st.event(cands[i].ID, "memory", "hmem", "matched "+rec.ID); err != nil {
				return candidate.Draft{}, err
			}
			continue
		}
		cands[i] = judge.Decide(cands[i])
		if err := st.event(cands[i].ID, "verdict", "judge", cands[i].DropReason); err != nil {
			return candidate.Draft{}, err
		}
	}
	// The dependency check and the EPSS scores come after the Judge: they add
	// findings judged on their own query rows and do not pass the Skeptic.
	cands, err = p.addOSV(ctx, st, q, fx.Root, fx.Snapshot, cands)
	if err != nil {
		if _, fatal := st.soften(ctx, err); fatal != nil {
			if errors.Is(fatal, errPartial) {
				return p.cut(st, cands, allow)
			}
			return candidate.Draft{}, fatal
		}
	}
	cands, err = p.attachEPSS(ctx, st, q, fx.Snapshot, cands)
	if err != nil {
		if _, fatal := st.soften(ctx, err); fatal != nil {
			if errors.Is(fatal, errPartial) {
				return p.cut(st, cands, allow)
			}
			return candidate.Draft{}, fatal
		}
	}
	return keep(writer.Write(writer.Input{
		Candidates: cands, Failed: st.failed, Models: st.models, TracePath: p.trace.Path(),
		SkippedPaths: st.skipped,
	})), nil
}

// attachEPSS scores every kept finding whose evidence names a CVE.
// A missing row is a trace line and adds no evidence.
func (p *Pipeline) attachEPSS(ctx context.Context, st *state, q query.Adapter, snapshot string, cands []candidate.Candidate) ([]candidate.Candidate, error) {
	// ADR-0013: the view is a setting; empty means the named default.
	view := p.epssView
	if view == "" {
		view = epss.DefaultView
	}
	next, notes, err := epss.Attach(ctx, q, view, snapshot, cands)
	for _, n := range notes {
		ev := trace.Event{FindingID: n.FindingID, Kind: "evidence", Agent: "epss", Text: n.Text}
		if n.HasEvidence {
			ev.Evidence = traceEvidence([]candidate.Evidence{n.Evidence})
		}
		if e := st.trace.Append(ev); e != nil {
			return nil, e
		}
	}
	return next, err
}

// addOSV names added and raised modules without a model, queries the OSV
// table, and compares versions in Go. A module with no affected row is a
// trace line, not a finding.
func (p *Pipeline) addOSV(ctx context.Context, st *state, q query.Adapter, root, snapshot string, cands []candidate.Candidate) ([]candidate.Candidate, error) {
	if ctx.Err() != nil {
		return cands, ctx.Err()
	}
	checked, err := osv.Check(ctx, q, root, p.osvTable, snapshot)
	for _, note := range checked.Notes {
		if e := st.event("osv", "osv", "osv", note); e != nil {
			return nil, e
		}
	}
	if checked.Down != nil {
		st.note(checked.Down)
	}
	if err != nil {
		return cands, err
	}
	for _, c := range checked.Candidates {
		// The Judge still decides an OSV finding: a matching range is
		// evidence, nothing else is.
		c = judge.Decide(c)
		if e := st.trace.Append(trace.Event{
			FindingID: c.ID, Kind: "claim", Agent: "osv", Text: c.Title,
			RuleID: c.RuleID, Path: c.Path, Title: c.Title,
		}); e != nil {
			return nil, e
		}
		text := evidenceText(c)
		if c.Message != "" && text != "none" {
			text += "; " + c.Message
		}
		if e := st.trace.Append(trace.Event{
			FindingID: c.ID,
			Kind:      "evidence",
			Agent:     "osv",
			Text:      text,
			Evidence:  traceEvidence(c.Evidence),
		}); e != nil {
			return nil, e
		}
		if e := st.event(c.ID, "verdict", "judge", c.DropReason); e != nil {
			return nil, e
		}
		cands = append(cands, c)
	}
	return cands, nil
}

// relations is the table and view names the review's settings give, for the
// templates the Investigator and the Skeptic name (ADR-0032, ADR-0013). An
// empty name means the default, and querytpl checks each name again when it
// builds, so a setting that is not a plain name is refused there.
func (p *Pipeline) relations(ctx context.Context) querytpl.Relations {
	return querytpl.Relations{
		OSVTable: p.osvTable,
		KEVView:  osv.KEVViewOf(ctx),
		EPSSView: p.epssView,
	}
}

// runRefutation asks the prover or the investigator to run what the skeptic proposed.
// A proposal with no test and no query is not evidence. The caller writes the
// refutation to the trace, and H-MEM learns it from the trace when the review is
// folded in.
func (p *Pipeline) runRefutation(ctx context.Context, c *candidate.Candidate, q query.Adapter, run testrun.Adapter) error {
	r := &c.Refutation
	if r.TestName != "" {
		out, err := prover.Execute(ctx, run, testrun.Request{
			FindingID: c.ID, Path: r.TestPath, Name: r.TestName,
		})
		if err != nil {
			// A deadline is not a test that cannot run; it is passed up as is.
			if ctx.Err() != nil {
				return err
			}
			r.TestCannotRun = true
			r.Reason = "the refuting test could not run"
			return err
		}
		if !out.Found {
			r.TestCannotRun = true
			r.Reason = "the refuting test could not run"
		} else {
			r.TestFound = true
			r.TestFailsOnChange = out.FailsOnChange
			if !out.FailsOnChange {
				name := out.Name
				if name == "" {
					name = r.TestName
				}
				r.Reason = "test " + name + " passes on the change"
			}
		}
	}
	if !r.Query.Empty() {
		res, err := investigator.Execute(ctx, q, r.Query, p.relations(ctx))
		if err != nil {
			// ADR-0032: a request Proven refuses was checked by the Skeptic
			// already, so this is a settings refusal. It is not a query that
			// could not run; the instrument never answered.
			var refusal *querytpl.Refusal
			if errors.As(err, &refusal) {
				r.Query = querytpl.Request{}
				r.Reason = "refused: " + refusal.Error()
				return nil
			}
			if ctx.Err() != nil {
				return err
			}
			r.QueryCannotRun = true
			if !r.TestFound || r.TestFailsOnChange {
				r.Reason = "the refuting query could not run"
			}
			return err
		}
		r.QueryRows = copyRows(res.Rows)
		// A test result already gives the reason when it refutes; the query
		// sets it only when the test did not run or did not refute.
		if judge.QueryContradicts(*c, r.QueryRows) && (r.Reason == "" || r.TestFailsOnChange || !r.TestFound) {
			r.Reason = "query rows contradict the finding"
		}
	}
	return nil
}

// noteFamily writes the same-family sentence when the skeptic and the finder
// share a family. Two unknown models are not that case. The review is not refused.
func (p *Pipeline) noteFamily(st *state) error {
	sk := p.skepticModel
	if sk == "" {
		sk = generator.DefaultSkepticModel
	}
	// A provider receipt is the stronger word on which model answered, so it
	// replaces the configured name.
	for _, m := range st.models {
		if m.Agent == "skeptic" && m.FromReceipt && m.Model != "" && m.Model != "unknown" {
			sk = m.Model
			break
		}
	}
	finder := "unknown"
	for _, m := range st.models {
		if m.Agent == "planner" && m.Model != "" {
			finder = m.Model
			break
		}
	}
	sentence, ok := skeptic.SameFamilySentence(sk, finder)
	if !ok {
		return nil
	}
	return st.event("review", "family", "skeptic", sentence)
}

// adapters picks the generator, the query adapter and the test runner for this
// review. Sample mode uses the fixture's recorded answers, tables and test
// outcomes; live mode uses the adapters Build was given, and loads the container
// settings only when no runner was given.
func (p *Pipeline) adapters(fx *fixture.Fixture) (generator.Generator, query.Adapter, testrun.Adapter, error) {
	var gen generator.Generator
	var q query.Adapter
	var err error
	if p.sample {
		if p.gen != nil {
			// A test may replace the recorded answers. The sample query and
			// the recorded tests stay in place. A review leaves Generator unset.
			gen = p.gen
		} else {
			gen, err = generator.NewSample(fx.Answers)
			if err != nil {
				return nil, nil, nil, err
			}
		}
		q, err = query.OpenSample(fx.Tables)
		if err != nil {
			return nil, nil, nil, err
		}
	} else if p.gen != nil {
		gen = p.gen
		q = p.liveQuery
	} else {
		gen = &generator.Live{}
		q = p.liveQuery
	}
	if !p.sample && q == nil {
		q = &query.Live{}
	}
	run := p.testRun
	if run == nil {
		if p.sample {
			run = testrun.NewSample(fx.Tests)
		} else {
			cfg, lerr := testrun.Load(p.env)
			if lerr != nil {
				return nil, nil, nil, lerr
			}
			// An explicit Options.MaxWrittenBytes wins over the setting.
			if p.maxWritten <= 0 {
				p.maxWritten = cfg.MaxWrittenBytes
			}
			c, nerr := testrun.NewContainer(cfg, nil)
			if nerr != nil {
				// ADR-0025: a missing container runtime turns the Prover off
				// and names why; it does not stop the review. A bad setting does.
				d, ok := service.As(nerr)
				if !ok {
					return nil, nil, nil, nerr
				}
				if p.proverIdle == "" {
					p.proverIdle = d.Reason
				}
				run = testrun.Off{Reason: d.Reason}
			} else {
				run = c
			}
		}
	}
	// Only the container needs the two trees on disk; a recorded or injected
	// runner does not.
	if _, ok := run.(*testrun.Container); ok {
		run = checkoutRun{
			inner:  run,
			base:   filepath.Join(fx.Root, "base"),
			change: filepath.Join(fx.Root, "change"),
		}
	}
	return gen, q, run, nil
}

// checkoutRun fills the two checkout directories the container mounts when
// the caller left them empty. A directory the caller set is left as it is.
type checkoutRun struct {
	inner        testrun.Adapter
	base, change string
}

// Run fills the empty checkout directories and runs req on the inner adapter.
func (a checkoutRun) Run(ctx context.Context, req testrun.Request) (testrun.Outcome, error) {
	if req.BaseDir == "" {
		req.BaseDir = a.base
	}
	if req.ChangeDir == "" {
		req.ChangeDir = a.change
	}
	return a.inner.Run(ctx, req)
}

// state is what one review gathers as it runs: the trace it writes to, the
// services that failed, the model calls made, the key of the change, and the
// tracked symlinks the trees left out.
type state struct {
	trace *trace.Writer
	// redactor is applied to the text of a service that becomes a failure
	// reason in the draft: the trace writer redacts what it writes, and the
	// draft is a second place that text lands.
	redactor  *redact.Redactor
	failed    []candidate.Failure
	models    []candidate.ModelUse
	changeKey string
	skipped   []string
}

// event writes one plain trace line about a finding.
func (s *state) event(id, kind, agent, text string) error {
	return s.trace.Append(trace.Event{FindingID: id, Kind: kind, Agent: agent, Text: text})
}

// claim writes the Planner's claim for a risk, under the run that named it.
func (s *state) claim(id, text, run string) error {
	return s.trace.Append(trace.Event{FindingID: id, Kind: "claim", Agent: "planner", Text: text, Run: run})
}

// evidence writes what the Investigator and the Prover found for a candidate.
func (s *state) evidence(c candidate.Candidate) error {
	return s.trace.Append(trace.Event{
		FindingID: c.ID,
		Kind:      "evidence",
		Agent:     evidenceAgent(c),
		Text:      evidenceText(c),
		Evidence:  traceEvidence(c.Evidence),
	})
}

// refutation writes the Skeptic's result for a candidate. It carries the change
// key, rule, path and title, which is what H-MEM stores a refutation by.
func (s *state) refutation(c candidate.Candidate, run string) error {
	text := c.Refutation.Reason
	if text == "" {
		text = "no refutation"
	}
	r := c.Refutation
	return s.trace.Append(trace.Event{
		FindingID: c.ID,
		Kind:      "refutation",
		Agent:     "skeptic",
		Text:      text,
		Refutation: &trace.Refutation{
			Attempted: r.Attempted,
			Refuted:   r.Refuted,
			Reason:    r.Reason,
			Evidenced: judge.Refutes(c),
		},
		ChangeKey:   s.changeKey,
		RuleID:      c.RuleID,
		Path:        c.Path,
		Title:       c.Title,
		HasEvidence: judge.Evidenced(c),
		Run:         run,
	})
}

// soften sorts an agent error. A service that is down is noted and reported as
// degraded, so the review goes on without it. An ended context becomes
// errPartial, so the caller returns a partial draft. Anything else is fatal.
func (s *state) soften(ctx context.Context, err error) (degraded bool, fatal error) {
	if err == nil {
		return false, nil
	}
	if d, ok := service.As(err); ok {
		s.note(d)
		return true, nil
	}
	if ctx.Err() != nil {
		return false, errPartial
	}
	return false, err
}

// errPartial tells the review loop to stop and return what is judged so far.
var errPartial = errors.New("review partial")

// note adds a failed service once; the first reason given is kept.
func (s *state) note(d *service.Down) {
	for _, f := range s.failed {
		if f.Service == d.Service {
			return
		}
	}
	// A refusal reason can carry a service's own words, such as a routed model or a
	// safety category, which may echo a configured value.
	s.failed = append(s.failed, candidate.Failure{Service: d.Service, Reason: s.redactor.Text(d.Reason)})
}

// cut ends the review early with a partial draft, limited to the selected files.
func (p *Pipeline) cut(st *state, cands []candidate.Candidate, allow map[string]struct{}) (candidate.Draft, error) {
	d, err := st.partial(cands)
	if err != nil {
		return candidate.Draft{}, err
	}
	d.Findings = keepFindings(d.Findings, allow)
	return d, nil
}

// partial writes the draft of a review the deadline cut short.
func (s *state) partial(cands []candidate.Candidate) (candidate.Draft, error) {
	// partial-path filter: a candidate the judge has not judged is not a finding.
	judged, err := s.omitUnjudged(cands)
	if err != nil {
		return candidate.Draft{}, err
	}
	return writer.Write(writer.Input{
		Candidates: judged, Failed: s.failed, Models: s.models, Partial: true, TracePath: s.trace.Path(),
		SkippedPaths: s.skipped,
	}), nil
}

// omitUnjudged drops every candidate the Judge has not decided, and writes a
// "not judged: deadline" verdict for each so the trace accounts for it.
func (s *state) omitUnjudged(cands []candidate.Candidate) ([]candidate.Candidate, error) {
	var out []candidate.Candidate
	for _, c := range cands {
		if c.DropReason == "" {
			if err := s.event(c.ID, "verdict", "pipeline", "not judged: deadline"); err != nil {
				return nil, err
			}
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

// noteGuardOff writes "no guardrail" once when the switch is off, so a review
// run without the guardrail says so on its trace (ADR-0005).
func (p *Pipeline) noteGuardOff() error {
	if !p.guardOff {
		return nil
	}
	return p.trace.Append(trace.Event{
		FindingID: "guardrail",
		Kind:      "guardrail",
		Agent:     "pipeline",
		Text:      "no guardrail",
	})
}

// guard checks the prompt before the model runs and the answer before it is used
// (ADR-0005). Each check is written to the trace.
// When the switch is off it returns the generator unchanged.
// rail nil uses the pipeline's adapter.
func (p *Pipeline) guard(gen generator.Generator, rail guardrail.Adapter) (generator.Generator, error) {
	if p.guardOff {
		return gen, nil
	}
	if rail == nil {
		rail = p.rail
	}
	return guardrail.Wrap(gen, rail, func(agent, side string, v guardrail.Verdict, refused string) error {
		return p.trace.Append(guardEvent(agent, side, v, refused))
	})
}

// guardEvent is the trace line for one guardrail check: the refusal when the
// check could not be made, otherwise the side, verdict, latency and categories.
func guardEvent(agent, side string, v guardrail.Verdict, refused string) trace.Event {
	ev := trace.Event{FindingID: "guardrail", Kind: "guardrail", Agent: agent}
	if refused != "" {
		ev.Text = "guardrail: " + refused
		return ev
	}
	verdict := "safe"
	if v.Unsafe() {
		verdict = "unsafe"
	}
	cats := strings.Join(v.Categories, ", ")
	ev.Text = fmt.Sprintf("%s %s latency=%dms", side, verdict, v.Latency.Milliseconds())
	if cats != "" {
		ev.Text += " categories=" + cats
	}
	ev.Guardrail = &trace.Check{
		Side: side, Verdict: verdict, LatencyMS: v.Latency.Milliseconds(),
		Categories: append([]string(nil), v.Categories...),
	}
	return ev
}

// recording wraps a generator and records every model call, on the review and
// on the trace.
type recording struct {
	inner  generator.Generator
	models *[]candidate.ModelUse
	trace  *trace.Writer
	// redactor is applied to the route, model and receipt id a service reported,
	// before they are kept on the review: a service can echo a configured value there.
	redactor *redact.Redactor
}

// Complete calls the inner generator and records which model answered. A
// generator that reports its use is recorded even when the call failed, since
// a call was made. One that does not is recorded as "unknown", never guessed.
func (g recording) Complete(ctx context.Context, agent, prompt string) (string, error) {
	text, err := g.inner.Complete(ctx, agent, prompt)
	if rep, ok := g.inner.(generator.Reported); ok {
		if u, ok := rep.LastUse(); ok {
			m := candidate.ModelUse{
				Agent: agent, Route: u.Route, Model: u.Model,
				ReceiptID: u.ReceiptID, FromReceipt: u.FromReceipt,
			}
			// Each place the use lands redacts for itself: the trace writer redacts what it
			// writes, and the copy kept on the review is redacted here.
			kept := m
			kept.Route, kept.Model, kept.ReceiptID = g.redactor.Text(m.Route), g.redactor.Text(m.Model), g.redactor.Text(m.ReceiptID)
			*g.models = append(*g.models, kept)
			if terr := g.record(m); terr != nil && err == nil {
				return "", terr
			}
		}
		return text, err
	}
	if err != nil {
		return "", err
	}
	m := candidate.ModelUse{Agent: agent, Model: "unknown", FromReceipt: false}
	*g.models = append(*g.models, m)
	if terr := g.record(m); terr != nil {
		return "", terr
	}
	return text, nil
}

// record writes one "model" event per generator call: the route, the routed
// model and the Nightshift receipt id, so a reader can match the trace to the
// provider's signed receipt.
func (g recording) record(m candidate.ModelUse) error {
	if g.trace == nil {
		return nil
	}
	return g.trace.Append(trace.Event{
		Kind:  "model",
		Agent: m.Agent,
		Text:  "model call",
		Model: &trace.Model{
			Route: m.Route, Model: m.Model, ReceiptID: m.ReceiptID, FromReceipt: m.FromReceipt,
		},
	})
}

// changedFiles lists the files of the change: from git when a baseline commit
// is named, otherwise from the fixture's two trees.
func changedFiles(req Request) ([]string, error) {
	if req.BaselineCommit != "" {
		return change.Between(req.Target, req.BaselineCommit, req.HeadCommit)
	}
	files, err := change.Read(req.Target)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out, nil
}

// selectionText is the trace text naming the files left out, sorted.
func selectionText(dropped []string) string {
	if len(dropped) == 0 {
		return "selected out: none"
	}
	cp := append([]string(nil), dropped...)
	sort.Strings(cp)
	return "selected out: " + strings.Join(cp, ", ")
}

// keepFindings keeps the findings whose path is one of the selected files.
func keepFindings(in []candidate.Candidate, allow map[string]struct{}) []candidate.Candidate {
	var out []candidate.Candidate
	for _, c := range in {
		if _, ok := allow[c.Path]; ok {
			out = append(out, c)
		}
	}
	return out
}

// evidenceText is the one-line trace text for a candidate's evidence, followed
// by the Prover's note when there is one, or "none".
func evidenceText(c candidate.Candidate) string {
	// The Investigator's refusal (ADR-0032) leads the notes, so the trace says
	// why the finding has no query evidence.
	notes := c.ProverNote
	if c.QueryNote != "" {
		if notes != "" {
			notes = c.QueryNote + "; " + notes
		} else {
			notes = c.QueryNote
		}
	}
	if len(c.Evidence) == 0 {
		if notes != "" {
			return notes
		}
		return "none"
	}
	text := ""
	for i, e := range c.Evidence {
		if i > 0 {
			text += "; "
		}
		if e.Test != nil {
			text += fmt.Sprintf("test %s %s", e.Test.Path, e.Test.Name)
		}
		if e.Query != nil {
			text += fmt.Sprintf("query %s rows=%d", e.Query.SQL, len(e.Query.Rows))
		}
		if e.Tool != nil {
			text += fmt.Sprintf("tool %s %s", e.Tool.Tool, e.Tool.Command)
		}
	}
	if text == "" {
		if notes != "" {
			return notes
		}
		return "none"
	}
	if notes != "" {
		return text + "; " + notes
	}
	return text
}

// traceEvidence copies evidence into its trace form. Rows and columns are
// copied, so the trace never shares a slice with a candidate.
func traceEvidence(in []candidate.Evidence) []trace.Evidence {
	if len(in) == 0 {
		return nil
	}
	out := make([]trace.Evidence, 0, len(in))
	for _, e := range in {
		ev := trace.Evidence{Kind: e.Kind}
		if e.Test != nil {
			t := e.Test
			ev.Test = &trace.TestEvidence{
				Path: t.Path, Name: t.Name, Command: t.Command, Output: t.Output,
				FailsOnChange: t.FailsOnChange, PassesOnBase: t.PassesOnBase,
			}
		}
		if e.Query != nil {
			q := e.Query
			ev.Query = &trace.QueryEvidence{
				SQL: q.SQL, Engine: q.Engine, Snapshot: q.Snapshot,
				Columns: append([]string(nil), q.Columns...),
				Rows:    copyRows(q.Rows),
			}
		}
		if e.Tool != nil {
			t := e.Tool
			ev.Tool = &trace.ToolEvidence{
				Tool: t.Tool, Version: t.Version, Command: t.Command,
				Output: t.Output, BaseOutput: t.BaseOutput,
				PresentOnChange: t.PresentOnChange, PresentOnBase: t.PresentOnBase,
			}
		}
		out = append(out, ev)
	}
	return out
}

// copyRows deep-copies query rows.
func copyRows(rows [][]string) [][]string {
	if rows == nil {
		return nil
	}
	out := make([][]string, len(rows))
	for i, row := range rows {
		out[i] = append([]string(nil), row...)
	}
	return out
}

// The two runs of a doubled review. A review that is not doubled has only
// the default run.
const (
	runDefault = "default"
	runSpecial = "special"
)

// secondGenerator is the generator the special-review run calls.
// An injected Special wins. A live generator is pointed at the special model.
// Sample mode, with none injected, reuses the sample generator for a fresh call.
func (p *Pipeline) secondGenerator(model string, sample generator.Generator) generator.Generator {
	if p.special != nil {
		return p.special
	}
	if model == "" {
		return nil
	}
	if live, ok := p.gen.(*generator.Live); ok {
		return live.WithModel(model)
	}
	return sample
}

// reviewerModel is the model name a run is credited to on the review: the
// configured name, or the run's own name when none was configured.
func reviewerModel(req Request, run string) string {
	if run == runSpecial {
		if req.SpecialModel != "" {
			return req.SpecialModel
		}
		return runSpecial
	}
	if req.DefaultModel != "" {
		return req.DefaultModel
	}
	return runDefault
}

// hasRun reports whether run named the candidate.
func hasRun(c candidate.Candidate, run string) bool {
	for _, r := range c.Runs {
		if r == run {
			return true
		}
	}
	return false
}

// cleared is a copy whose refutation is empty, so the second skeptic does not
// see the first run's attempt. The prompt is still only this finding's id and title.
func cleared(c candidate.Candidate) candidate.Candidate {
	c.Refutation = candidate.Refutation{}
	return c
}

// mergeCandidates adds the special run's risks to the default run's. The same risk is the same rule at the same
// place (rule id, path and lines; the title when there is no rule id), never the same model-chosen id: each run numbers
// its own risks. A different risk whose id is already taken gets a new id. ids holds each extra risk's final id.
func mergeCandidates(base, extra []candidate.Candidate) (merged []candidate.Candidate, ids []string) {
	at := make(map[string]int, len(base))
	taken := make(map[string]bool, len(base)+len(extra))
	for i, c := range base {
		at[riskKey(c.Risk)] = i
		taken[c.ID] = true
	}
	ids = make([]string, len(extra))
	for j, c := range extra {
		if i, ok := at[riskKey(c.Risk)]; ok {
			base[i].Runs = appendNew(base[i].Runs, c.Runs)
			base[i].Models = appendNew(base[i].Models, c.Models)
			ids[j] = base[i].ID
			continue
		}
		id := c.ID
		for n := 1; taken[id]; n++ {
			id = fmt.Sprintf("%s-s%d", c.ID, n)
		}
		c.ID = id
		taken[id] = true
		base = append(base, c)
		at[riskKey(c.Risk)] = len(base) - 1
		ids[j] = id
	}
	return base, ids
}

// riskKey is what makes two risks the same risk: the rule id with the path and
// lines, or, with no rule id, the title with spaces and case evened out and the
// path.
func riskKey(r candidate.Risk) string {
	if strings.TrimSpace(r.RuleID) == "" {
		return "title\x00" + strings.ToLower(strings.Join(strings.Fields(r.Title), " ")) + "\x00" + r.Path
	}
	return fmt.Sprintf("rule\x00%s\x00%s\x00%d\x00%d", r.RuleID, r.Path, r.StartLine, r.EndLine)
}

// appendNew appends the values of src that dst does not hold yet.
func appendNew(dst, src []string) []string {
	for _, s := range src {
		seen := false
		for _, d := range dst {
			if d == s {
				seen = true
				break
			}
		}
		if !seen {
			dst = append(dst, s)
		}
	}
	return dst
}

// evidenceAgent names the agent the evidence is credited to: the Prover when a
// test is among it, otherwise the Investigator.
func evidenceAgent(c candidate.Candidate) string {
	for _, e := range c.Evidence {
		if e.Kind == candidate.KindTest {
			return "prover"
		}
	}
	return "investigator"
}
