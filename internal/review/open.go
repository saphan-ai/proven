package review

import (
	"context"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

func open(opts Options) (Reviewer, error) {
	if opts.LogDir == "" {
		return nil, &ConfigError{
			Field:   "LogDir",
			Value:   "",
			Reason:  "the run needs a logs directory",
			Allowed: []string{"a directory path"},
		}
	}
	popts := pipeline.Options{
		Sample: opts.Sample,
		LogDir: opts.LogDir,
		Env:    opts.Env,
	}
	if !opts.Sample {
		box, err := testrun.NewContainer(testrun.DefaultContainerConfig(), nil)
		if err != nil {
			d, ok := service.As(err)
			if !ok {
				return nil, err
			}
			popts.TestRun = testrun.Off{Reason: d.Reason}
			popts.ProverIdle = d.Reason
		} else {
			popts.TestRun = box
		}
	}
	p, err := pipeline.Build(popts)
	if err != nil {
		return nil, err
	}
	return &runner{p: p}, nil
}

// runner is the seam's Reviewer. It copies the writer's draft into Result.
type runner struct {
	p *pipeline.Pipeline
}

func (r *runner) Review(ctx context.Context, req Request) (Result, error) {
	d, err := r.p.Review(ctx, pipeline.Request{Target: req.Target})
	if err != nil {
		return Result{}, err
	}
	return fill(d), nil
}

func fill(d candidate.Draft) Result {
	out := Result{Partial: d.Partial, TracePath: d.TracePath}
	for _, c := range d.Findings {
		f := Finding{
			ID: c.ID, RuleID: c.RuleID, Title: c.Title, Message: c.Message,
			Severity:   Severity(c.Severity),
			Location:   Location{Path: c.Path, StartLine: c.StartLine, EndLine: c.EndLine},
			Unresolved: c.Unresolved,
		}
		for _, e := range c.Evidence {
			f.Evidence = append(f.Evidence, toEvidence(e))
		}
		out.Findings = append(out.Findings, f)
	}
	for _, s := range d.Failed {
		out.FailedServices = append(out.FailedServices, ServiceFailure{Service: s.Service, Reason: s.Reason})
	}
	for _, m := range d.Models {
		out.Models = append(out.Models, ModelUse{
			Agent: m.Agent, Route: m.Route, Model: m.Model, FromReceipt: m.FromReceipt,
		})
	}
	if out.Findings == nil {
		out.Findings = []Finding{}
	}
	if out.FailedServices == nil {
		out.FailedServices = []ServiceFailure{}
	}
	return out
}

func toEvidence(e candidate.Evidence) Evidence {
	out := Evidence{Kind: e.Kind}
	if e.Test != nil {
		te := *e.Test
		out.Test = &TestEvidence{
			Path: te.Path, Name: te.Name, Command: te.Command, Output: te.Output,
			FailsOnChange: te.FailsOnChange, PassesOnBase: te.PassesOnBase,
		}
	}
	if e.Query != nil {
		q := *e.Query
		out.Query = &QueryEvidence{
			SQL: q.SQL, Engine: q.Engine, Snapshot: q.Snapshot,
			Columns: append([]string(nil), q.Columns...),
			Rows:    copyRows(q.Rows),
		}
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
