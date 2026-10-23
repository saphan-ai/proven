package review

// This file wires the reviewer: Open builds the adapters and the pipeline, and the runner turns
// one Request into one Result.

import (
	"context"
	"errors"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// open is Open. It refuses a missing LogDir, loads proven.yaml when the shell named one, and in live
// mode loads every adapter's settings from the environment (ADR-0001, ADR-0013). Any refused setting
// returns a *ConfigError, so the shell exits 7 before any review starts.
func open(opts Options) (Reviewer, error) {
	// The trace is written under LogDir; a review with no trace could not show how a finding was judged.
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
	// An explicit --config is read once, here. Without one, each Review reads proven.yaml from the
	// change's baseline (see runner.Review).
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
	// In sample mode the pipeline uses its recorded adapters, so none of the live settings are read.
	if !opts.Sample {
		env := opts.Env
		if env == nil {
			env = func(string) string { return "" }
		}
		settings, lerr := generator.Load(env)
		if lerr != nil {
			return nil, generatorConfig(lerr)
		}
		// ADR-0006: the Skeptic runs on a model of another family than the finder's; that model is a setting.
		popts.SkepticModel = settings.SkepticModel
		gs, gerr := guardrail.Load(env)
		if gerr != nil {
			return nil, guardrailConfig(gerr)
		}
		// ADR-0005: the guardrail is on unless a setting turns it off. When off, the pipeline is told
		// so and the trace records one "no guardrail" line.
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
		if opts.gen != nil {
			// A test supplies the generator. The live client above is not called,
			// so the review does not dial. A review leaves gen unset.
			popts.Generator = opts.gen
		}
		zset, zerr := query.Load(env)
		if zerr != nil {
			return nil, queryConfig(zerr)
		}
		popts.LiveQuery = query.NewLive(zset, nil)
		if opts.liveQuery != nil {
			// A test supplies the query adapter. The live client above is not called.
			popts.LiveQuery = opts.liveQuery
		}
		table, terr := osv.LoadTable(env)
		if terr != nil {
			return nil, osvConfig(terr)
		}
		popts.OSVTable = table
		view, verr := epss.LoadView(env)
		if verr != nil {
			return nil, epssConfig(verr)
		}
		popts.EPSSView = view
		if opts.testRun != nil {
			// A test supplies the runner, so Open does not look for a container.
			// Production leaves testRun unset and uses the container below.
			popts.TestRun = opts.testRun
		} else {
			boxCfg, berr := testrun.Load(env)
			if berr != nil {
				return nil, testrunConfig(berr)
			}
			popts.MaxWrittenBytes = boxCfg.MaxWrittenBytes
			// ADR-0025: the Prover runs a test only in a container with no network and no secret.
			// A missing container runtime is not a refusal: the Prover is off, the reason is named in the
			// review, and the other agents still run. Any other error stops Open.
			box, err := testrun.NewContainer(boxCfg, nil)
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
	}
	p, err := pipeline.Build(popts)
	if err != nil {
		return nil, err
	}
	return &runner{p: p, cfg: cfg, explicit: explicit}, nil
}

// runner is the seam's Reviewer. It copies the writer's draft into Result.
// cfg is the explicit configuration when explicit is true; otherwise it is unused and each
// Review loads the settings of the change it reviews.
type runner struct {
	p        *pipeline.Pipeline
	cfg      fileSettings
	explicit bool
}

// Review runs one review: settings, deadline, class, the six agents, then the decision lines.
func (r *runner) Review(ctx context.Context, req Request) (Result, error) {
	cfg := r.cfg
	// Settings come from the baseline commit, not from the change, so a change cannot loosen its
	// own review. With no baseline there is only the working tree to read.
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
	// The timeout setting applies only when the shell set no deadline: --timeout wins.
	// Zero means no deadline from the settings.
	timeout := cfg.Timeout
	if timeout > 0 {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}
	// ADR-0009: a doubled class runs the agents a second time with the special-review model.
	// With no special model configured there is nothing to double, so the class is not needed here;
	// applyDecision still states it and says the special model is not configured.
	double := false
	if cfg.SpecialReviewer != "" {
		files, ferr := changedPaths(req)
		if ferr != nil {
			return Result{}, ferr
		}
		class, _ := chooseClass(files, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
		double = cfg.Doubled[class]
	}
	d, err := r.p.Review(ctx, pipeline.Request{
		Target:         req.Target,
		BaselineCommit: req.BaselineCommit,
		HeadCommit:     req.HeadCommit,
		Include:        req.Include,
		Exclude:        req.Exclude,
		DoubleRun:      double,
		DefaultModel:   cfg.DefaultReviewer,
		SpecialModel:   cfg.SpecialReviewer,
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

// fill copies the writer's draft into the seam's Result. Slices are copied, so the Result shares no
// memory with the pipeline. Findings and FailedServices are never nil, so an empty review prints
// an empty list rather than a missing one.
func fill(d candidate.Draft) Result {
	out := Result{Partial: d.Partial, TracePath: d.TracePath}
	for _, c := range d.Findings {
		f := Finding{
			ID: c.ID, RuleID: c.RuleID, Title: c.Title, Message: c.Message,
			Severity:   Severity(c.Severity),
			Location:   Location{Path: c.Path, StartLine: c.StartLine, EndLine: c.EndLine},
			Unresolved: c.Unresolved,
			Models:     append([]string(nil), c.Models...),
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
	if len(d.SkippedPaths) > 0 {
		out.SkippedPaths = append([]string(nil), d.SkippedPaths...)
	}
	if out.Findings == nil {
		out.Findings = []Finding{}
	}
	if out.FailedServices == nil {
		out.FailedServices = []ServiceFailure{}
	}
	return out
}

// toEvidence copies one piece of evidence across the seam, field by field.
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

// queryConfig turns a refused Zetaris query setting into the seam's *ConfigError.
// The shell exits 7 on that type. Any other error passes through unchanged.
// The five functions below do the same for the other adapters' settings.
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

// guardrailConfig turns a refused guardrail setting into the seam's *ConfigError.
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

// epssConfig turns a refused EPSS view setting into the seam's *ConfigError.
func epssConfig(err error) error {
	var cfg *epss.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
}

// osvConfig turns a refused OSV table setting into the seam's *ConfigError.
func osvConfig(err error) error {
	var cfg *osv.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
}

// testrunConfig turns a refused container setting into the seam's *ConfigError.
func testrunConfig(err error) error {
	var cfg *testrun.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
}

// generatorConfig turns a refused Nightshift generator setting into the seam's *ConfigError.
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

// copyRows copies query rows deeply. A nil input stays nil, so "no rows" and "zero rows" stay distinct.
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
