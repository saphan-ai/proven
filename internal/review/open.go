package review

import (
	"context"
	"errors"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/query"
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
	cfg := defaultSettings()
	explicit := false
	if opts.ConfigPath != "" {
		loaded, err := loadConfigFile(opts.ConfigPath)
		if err != nil {
			return nil, err
		}
		cfg = loaded
		explicit = true
	}
	popts := pipeline.Options{
		Sample: opts.Sample,
		LogDir: opts.LogDir,
		Env:    opts.Env,
	}
	if !opts.Sample {
		env := opts.Env
		if env == nil {
			env = func(string) string { return "" }
		}
		settings, lerr := generator.Load(env)
		if lerr != nil {
			return nil, generatorConfig(lerr)
		}
		gs, gerr := guardrail.Load(env)
		if gerr != nil {
			return nil, guardrailConfig(gerr)
		}
		if gs.Mode == guardrail.ModeOff {
			popts.GuardOff = true
		} else {
			rail, nerr := guardrail.New(gs, nil)
			if nerr != nil {
				return nil, nerr
			}
			popts.Guardrail = rail
		}
		popts.Generator = generator.NewLive(settings, nil)
		zset, zerr := query.Load(env)
		if zerr != nil {
			return nil, queryConfig(zerr)
		}
		popts.LiveQuery = query.NewLive(zset, nil)
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
	return &runner{p: p, cfg: cfg, explicit: explicit}, nil
}

// runner is the seam's Reviewer. It copies the writer's draft into Result.
type runner struct {
	p        *pipeline.Pipeline
	cfg      fileSettings
	explicit bool
}

func (r *runner) Review(ctx context.Context, req Request) (Result, error) {
	cfg := r.cfg
	if !r.explicit {
		var loaded fileSettings
		var err error
		if req.BaselineCommit != "" {
			loaded, err = loadSettingsAtBaseline(req.Target, req.BaselineCommit)
		} else {
			loaded, err = loadSettingsInTarget(req.Target)
		}
		if err != nil {
			return Result{}, err
		}
		cfg = loaded
	}
	timeout := cfg.Timeout
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
	out := fill(d)
	if err := applyDecision(&out, cfg, req); err != nil {
		return Result{}, err
	}
	return out, nil
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
			Agent: m.Agent, Route: m.Route, Model: m.Model, ReceiptID: m.ReceiptID, FromReceipt: m.FromReceipt,
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

// generatorConfig is the seam's *ConfigError for a refused generator setting.
// The shell exits 7 on that type.
func queryConfig(err error) error {
	var cfg *query.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
}

func guardrailConfig(err error) error {
	var cfg *guardrail.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
}

func generatorConfig(err error) error {
	var cfg *generator.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
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
