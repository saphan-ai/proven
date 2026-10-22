package review

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/pipeline"
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
	var timeout time.Duration
	explicit := false
	if opts.ConfigPath != "" {
		d, err := loadConfigFile(opts.ConfigPath)
		if err != nil {
			return nil, err
		}
		timeout = d
		explicit = true
	}
	p, err := pipeline.Build(pipeline.Options{
		Sample: opts.Sample,
		LogDir: opts.LogDir,
		Env:    opts.Env,
	})
	if err != nil {
		return nil, err
	}
	return &runner{p: p, timeout: timeout, explicit: explicit}, nil
}

// runner is the seam's Reviewer. It copies the writer's draft into Result.
type runner struct {
	p        *pipeline.Pipeline
	timeout  time.Duration
	explicit bool
}

func (r *runner) Review(ctx context.Context, req Request) (Result, error) {
	timeout := r.timeout
	if !r.explicit {
		d, err := timeoutInTarget(req.Target)
		if err != nil {
			return Result{}, err
		}
		timeout = d
	}
	if timeout > 0 {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}
	d, err := r.p.Review(ctx, pipeline.Request{
		Target:         req.Target,
		BaselineCommit: req.BaselineCommit,
		HeadCommit:     req.HeadCommit,
		Include:        req.Include,
		Exclude:        req.Exclude,
	})
	if err != nil {
		return Result{}, err
	}
	return fill(d), nil
}

// timeoutInTarget reads proven.yaml in the review target when Open was given
// no path. A missing file is the defaults.
func timeoutInTarget(target string) (time.Duration, error) {
	path := filepath.Join(target, "proven.yaml")
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if st.IsDir() {
		return 0, &ConfigError{
			Field: "ConfigPath", Value: path, Reason: "path is a directory",
			Allowed: []string{"a path to an existing file"},
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return parseConfig(b)
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
