// Package pipeline runs the six agents in order and implements the review.
package pipeline

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/saphan-ai/proven/internal/netguard"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
	"github.com/saphan-ai/proven/internal/trace"
)

// Options select the adapters. TestRun, when set, replaces the sample or live test-run adapter.
type Options struct {
	Sample  bool
	LogDir  string
	Env     func(string) string
	TestRun testrun.Adapter
}

// Pipeline is one reviewer.
type Pipeline struct {
	sample  bool
	testRun testrun.Adapter
	trace   *trace.Writer
	guard   *netguard.Guard
}

// Build opens the trace and, in sample mode, a client that refuses every dial.
func Build(opts Options) (*Pipeline, error) {
	// opts.Env is not called. This wagon reads no key.
	_ = opts.Env
	if opts.LogDir == "" {
		return nil, fmt.Errorf("pipeline: logs directory is empty")
	}
	tw, err := trace.Open(opts.LogDir)
	if err != nil {
		return nil, err
	}
	p := &Pipeline{sample: opts.Sample, testRun: opts.TestRun, trace: tw}
	if opts.Sample {
		p.guard = &netguard.Guard{}
		p.guard.Arm()
	}
	return p, nil
}

// Guard is the sample-mode dial counter. It is nil when sample mode is off.
func (p *Pipeline) Guard() *netguard.Guard { return p.guard }

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
func (p *Pipeline) Review(ctx context.Context, req Request) (candidate.Draft, error) {
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
	st := &state{trace: p.trace}
	gen = recording{inner: gen, models: &st.models}

	risks, err := planner.Plan(ctx, req.Target, gen)
	if err != nil {
		if d, fatal := st.soften(ctx, err); fatal != nil {
			if errors.Is(fatal, errPartial) {
				return keep(st.partial(nil)), nil
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
	if ctx.Err() != nil {
		return keep(st.partial(cands)), nil
	}

	for i := range cands {
		next, err := investigator.Investigate(ctx, cands[i], fx.Snapshot, gen, q)
		if err != nil {
			if _, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return keep(st.partial(cands)), nil
				}
				return candidate.Draft{}, fatal
			}
			continue
		}
		cands[i] = next
	}
	if ctx.Err() != nil {
		return keep(st.partial(cands)), nil
	}
	for i := range cands {
		next, err := prover.Prove(ctx, cands[i], gen, run)
		if err != nil {
			if _, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return keep(st.partial(cands)), nil
				}
				return candidate.Draft{}, fatal
			}
			continue
		}
		cands[i] = next
	}
	for _, c := range cands {
		if err := st.event(c.ID, "evidence", evidenceAgent(c), evidenceText(c)); err != nil {
			return candidate.Draft{}, err
		}
	}
	if ctx.Err() != nil {
		return keep(st.partial(cands)), nil
	}

	for i := range cands {
		next, err := skeptic.Attempt(ctx, cands[i], gen)
		if err != nil {
			if _, fatal := st.soften(ctx, err); fatal != nil {
				if errors.Is(fatal, errPartial) {
					return keep(st.partial(cands)), nil
				}
				return candidate.Draft{}, fatal
			}
			cands[i].Refutation = candidate.Refutation{Attempted: false, Reason: "not attempted"}
			continue
		}
		cands[i] = next
	}
	for _, c := range cands {
		text := c.Refutation.Reason
		if text == "" {
			text = "no refutation"
		}
		if err := st.event(c.ID, "refutation", "skeptic", text); err != nil {
			return candidate.Draft{}, err
		}
	}
	if ctx.Err() != nil {
		return keep(st.partial(cands)), nil
	}

	for i := range cands {
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
		gen, err = generator.NewSample(fx.Answers)
		if err != nil {
			return nil, nil, nil, err
		}
		q, err = query.OpenSample(fx.Tables)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		gen = generator.Live{}
		q = query.Live{}
	}
	if p.testRun != nil {
		return gen, q, p.testRun, nil
	}
	if p.sample {
		return gen, q, testrun.NewSample(fx.Tests), nil
	}
	return gen, q, testrun.Live{}, nil
}

type state struct {
	trace  *trace.Writer
	failed []candidate.Failure
	models []candidate.ModelUse
}

func (s *state) event(id, kind, agent, text string) error {
	return s.trace.Append(trace.Event{FindingID: id, Kind: kind, Agent: agent, Text: text})
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

func (s *state) partial(cands []candidate.Candidate) candidate.Draft {
	d := writer.Write(writer.Input{
		Candidates: cands, Failed: s.failed, Models: s.models, Partial: true, TracePath: s.trace.Path(),
	})
	return d
}

type recording struct {
	inner  generator.Generator
	models *[]candidate.ModelUse
}

func (g recording) Complete(ctx context.Context, agent, prompt string) (string, error) {
	text, err := g.inner.Complete(ctx, agent, prompt)
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
		return "none"
	}
	return text
}

func evidenceAgent(c candidate.Candidate) string {
	for _, e := range c.Evidence {
		if e.Kind == candidate.KindTest {
			return "prover"
		}
	}
	return "investigator"
}
