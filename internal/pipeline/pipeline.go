// Package pipeline runs the six agents in order and implements the review.
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
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/query"
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
	// ProverIdle, when set, is named on the review as the Prover being off.
	// The review runs no test code in that case.
	ProverIdle string
	// LiveQuery, when set, is the Zetaris adapter. Sample mode does not use it.
	LiveQuery query.Adapter
	// StateDir is the durable directory for H-MEM and the world model.
	// Empty uses a state directory beside the logs, never a path inside the module.
	StateDir string
	// Guardrail checks every prompt and every answer. Sample mode with none
	// uses the safe sample. A live pipeline with none is refused.
	Guardrail guardrail.Adapter
	// GuardOff writes one "no guardrail" line and checks nothing.
	GuardOff bool
}

// Pipeline is one reviewer.
type Pipeline struct {
	sample     bool
	testRun    testrun.Adapter
	proverIdle string
	gen        generator.Generator
	liveQuery  query.Adapter
	trace      *trace.Writer
	rail       guardrail.Adapter
	guardOff   bool
	atBoundary func(stage string)
	mem        *hmem.Store
	world      *world.Store
}

// Build opens the trace. Sample mode uses the recorded adapters. Those
// adapters do not dial, so sample mode holds no HTTP client.
func Build(opts Options) (*Pipeline, error) {
	// opts.Env is not called. review.Open loads the live generator before Build.
	// Sample mode reads no key.
	if opts.LogDir == "" {
		return nil, fmt.Errorf("pipeline: logs directory is empty")
	}
	if opts.GuardOff && opts.Guardrail != nil {
		return nil, fmt.Errorf("pipeline: the guardrail is off and an adapter was set")
	}
	tw, err := trace.Open(opts.LogDir)
	if err != nil {
		return nil, err
	}
	stateDir := opts.StateDir
	if stateDir == "" {
		var derr error
		stateDir, derr = defaultStateDir(opts.LogDir)
		if derr != nil {
			return nil, derr
		}
	}
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
	p := &Pipeline{
		sample: opts.Sample, testRun: opts.TestRun, proverIdle: opts.ProverIdle,
		gen: opts.Generator, liveQuery: opts.LiveQuery, trace: tw, rail: opts.Guardrail, guardOff: opts.GuardOff,
		mem: mem, world: wm,
	}
	if !opts.GuardOff && p.rail == nil {
		if opts.Sample {
			p.rail = guardrail.Safe{}
		} else {
			return nil, fmt.Errorf("pipeline: the guardrail is on and no adapter was set")
		}
	}
	return p, nil
}

// boundary lets a test cancel the context at one named stage. Production leaves it unset.
func (p *Pipeline) boundary(stage string) {
	if p.atBoundary != nil {
		p.atBoundary(stage)
	}
}

// Request is the part of a review request this wagon reads.
type Request struct {
	Target         string
	BaselineCommit string
	HeadCommit     string
	Include        []string
	Exclude        []string
}

// Review runs planner, investigator, prover, skeptic, judge and writer.
// Findings whose path is outside the selected files are dropped after the agents run.
func (p *Pipeline) Review(ctx context.Context, req Request) (d candidate.Draft, err error) {
	defer func() {
		if err != nil {
			return
		}
		if ferr := p.fold(); ferr != nil {
			d = candidate.Draft{}
			err = ferr
		}
	}()
	return p.review(ctx, req)
}

func (p *Pipeline) review(ctx context.Context, req Request) (candidate.Draft, error) {
	files, err := changedFiles(req)
	if err != nil {
		return candidate.Draft{}, err
	}
	selected, dropped, err := change.Select(files, req.Include, req.Exclude)
	if err != nil {
		return candidate.Draft{}, err
	}
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
	fx, err := fixture.Load(req.Target)
	if err != nil {
		return candidate.Draft{}, err
	}
	gen, q, run, err := p.adapters(fx)
	if err != nil {
		return candidate.Draft{}, err
	}
	p.mem.SetGenerator(gen)
	p.world.SetGenerator(gen)
	st := &state{trace: p.trace, changeKey: changeKey(req)}
	if p.proverIdle != "" {
		st.note(&service.Down{Service: "prover", Reason: p.proverIdle})
	}
	gen = recording{inner: gen, models: &st.models}
	gen, err = p.applyGuard(gen)
	if err != nil {
		return candidate.Draft{}, err
	}

	risks, err := planner.Plan(ctx, req.Target, gen)
	if err != nil {
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
		c := candidate.Candidate{Risk: r}
		if err := st.event(r.ID, "claim", "planner", r.Title); err != nil {
			return candidate.Draft{}, err
		}
		cands = append(cands, c)
	}
	p.boundary("planned")
	if ctx.Err() != nil {
		return p.cut(st, cands, allow)
	}

	for i := range cands {
		next, err := investigator.Investigate(ctx, cands[i], fx.Snapshot, gen, q)
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
		next, err := prover.Prove(ctx, cands[i], gen, run, prover.Options{
			BaseDir:         filepath.Join(fx.Root, "base"),
			ChangeDir:       filepath.Join(fx.Root, "change"),
			MaxWrittenBytes: testrun.DefaultMaxWrittenBytes,
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

	for i := range cands {
		next, err := skeptic.Attempt(ctx, cands[i], gen)
		if err != nil {
			if _, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return p.cut(st, cands, allow)
				}
				return candidate.Draft{}, fatal
			}
			cands[i].Refutation = candidate.Refutation{Attempted: false, Reason: "not attempted"}
			continue
		}
		cands[i] = next
	}
	for _, c := range cands {
		if err := st.refutation(c); err != nil {
			return candidate.Draft{}, err
		}
	}
	p.boundary("refuted")
	if ctx.Err() != nil {
		return p.cut(st, cands, allow)
	}

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
	return keep(writer.Write(writer.Input{
		Candidates: cands, Failed: st.failed, Models: st.models, TracePath: p.trace.Path(),
	})), nil
}

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
			c, nerr := testrun.NewContainer(testrun.DefaultContainerConfig(), nil)
			if nerr != nil {
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

func (a checkoutRun) Run(ctx context.Context, req testrun.Request) (testrun.Outcome, error) {
	if req.BaseDir == "" {
		req.BaseDir = a.base
	}
	if req.ChangeDir == "" {
		req.ChangeDir = a.change
	}
	return a.inner.Run(ctx, req)
}

type state struct {
	trace     *trace.Writer
	failed    []candidate.Failure
	models    []candidate.ModelUse
	changeKey string
}

func (s *state) event(id, kind, agent, text string) error {
	return s.trace.Append(trace.Event{FindingID: id, Kind: kind, Agent: agent, Text: text})
}

func (s *state) evidence(c candidate.Candidate) error {
	return s.trace.Append(trace.Event{
		FindingID: c.ID,
		Kind:      "evidence",
		Agent:     evidenceAgent(c),
		Text:      evidenceText(c),
		Evidence:  traceEvidence(c.Evidence),
	})
}

func (s *state) refutation(c candidate.Candidate) error {
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
		},
		ChangeKey:   s.changeKey,
		RuleID:      c.RuleID,
		Path:        c.Path,
		Title:       c.Title,
		HasEvidence: judge.Evidenced(c.Evidence),
	})
}

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

var errPartial = errors.New("review partial")

func (s *state) note(d *service.Down) {
	for _, f := range s.failed {
		if f.Service == d.Service {
			return
		}
	}
	s.failed = append(s.failed, candidate.Failure{Service: d.Service, Reason: d.Reason})
}

func (p *Pipeline) cut(st *state, cands []candidate.Candidate, allow map[string]struct{}) (candidate.Draft, error) {
	d, err := st.partial(cands)
	if err != nil {
		return candidate.Draft{}, err
	}
	d.Findings = keepFindings(d.Findings, allow)
	return d, nil
}

func (s *state) partial(cands []candidate.Candidate) (candidate.Draft, error) {
	// partial-path filter: a candidate the judge has not judged is not a finding.
	judged, err := s.omitUnjudged(cands)
	if err != nil {
		return candidate.Draft{}, err
	}
	return writer.Write(writer.Input{
		Candidates: judged, Failed: s.failed, Models: s.models, Partial: true, TracePath: s.trace.Path(),
	}), nil
}

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

// applyGuard writes "no guardrail" once when the switch is off.
// Otherwise it checks the prompt before the model runs and the answer before it is used.
func (p *Pipeline) applyGuard(gen generator.Generator) (generator.Generator, error) {
	if p.guardOff {
		err := p.trace.Append(trace.Event{
			FindingID: "guardrail",
			Kind:      "guardrail",
			Agent:     "pipeline",
			Text:      "no guardrail",
		})
		return gen, err
	}
	return guardrail.Wrap(gen, p.rail, func(agent, side string, v guardrail.Verdict, refused string) error {
		return p.trace.Append(guardEvent(agent, side, v, refused))
	})
}

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

type recording struct {
	inner  generator.Generator
	models *[]candidate.ModelUse
}

func (g recording) Complete(ctx context.Context, agent, prompt string) (string, error) {
	text, err := g.inner.Complete(ctx, agent, prompt)
	if rep, ok := g.inner.(generator.Reported); ok {
		if u, ok := rep.LastUse(); ok {
			*g.models = append(*g.models, candidate.ModelUse{
				Agent: agent, Route: u.Route, Model: u.Model,
				ReceiptID: u.ReceiptID, FromReceipt: u.FromReceipt,
			})
		}
		return text, err
	}
	if err != nil {
		return "", err
	}
	*g.models = append(*g.models, candidate.ModelUse{Agent: agent, Model: "unknown", FromReceipt: false})
	return text, nil
}

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

func selectionText(dropped []string) string {
	if len(dropped) == 0 {
		return "selected out: none"
	}
	cp := append([]string(nil), dropped...)
	sort.Strings(cp)
	return "selected out: " + strings.Join(cp, ", ")
}

func keepFindings(in []candidate.Candidate, allow map[string]struct{}) []candidate.Candidate {
	var out []candidate.Candidate
	for _, c := range in {
		if _, ok := allow[c.Path]; ok {
			out = append(out, c)
		}
	}
	return out
}

func evidenceText(c candidate.Candidate) string {
	if len(c.Evidence) == 0 {
		if c.ProverNote != "" {
			return c.ProverNote
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
	}
	if text == "" {
		if c.ProverNote != "" {
			return c.ProverNote
		}
		return "none"
	}
	if c.ProverNote != "" {
		return text + "; " + c.ProverNote
	}
	return text
}

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
		out = append(out, ev)
	}
	return out
}

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

func evidenceAgent(c candidate.Candidate) string {
	for _, e := range c.Evidence {
		if e.Kind == candidate.KindTest {
			return "prover"
		}
	}
	return "investigator"
}
