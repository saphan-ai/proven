package review

// This file wires the reviewer: Open builds the adapters and the pipeline, and the runner turns
// one Request into one Result.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// open is Open. It refuses a missing LogDir, loads proven.yaml when the shell named one, and in live
// mode loads every adapter's settings from the environment (ADR-0001, ADR-0013). A wrong non-empty
// value returns a *ConfigError, so the shell exits 7 before any review starts. An empty secret does
// not: that service is down and the review names it (ADR-0013 point 4a).
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
	// ADR-0013: the size bounds on every read of the repository under review are
	// settings. They are loaded first, so the explicit configuration below is read
	// under them too. A wrong value stops Open before any review and is never
	// replaced by the default. Sample mode reads no setting, so it puts the defaults back: the bounds
	// are process state, and a sample review must not run under an earlier review's bounds.
	reads := change.DefaultLimits()
	if !opts.Sample {
		loaded, rerr := change.LoadLimits(opts.Env)
		if rerr != nil {
			return nil, readLimitConfig(rerr)
		}
		reads = loaded
	}
	change.UseLimits(reads)
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
	var kevView string
	var adapterLimit time.Duration
	// secondModel is the connected second reviewer's model; empty when none is connected (ADR-0037).
	var secondModel string
	// In sample mode the pipeline uses its recorded adapters, so none of the live settings are read.
	if !opts.Sample {
		env := opts.Env
		if env == nil {
			env = func(string) string { return "" }
		}
		settings, lerr := generator.Load(env)
		if d, ok := service.As(lerr); ok {
			// ADR-0013 point 4a: an empty token does not build a live client.
			popts.Down = append(popts.Down, *d)
			popts.Generator = generator.Absent{Service: d.Service, Reason: d.Reason}
		} else if lerr != nil {
			return nil, generatorConfig(lerr)
		} else {
			// ADR-0006: the Skeptic runs on a model of another family than the finder's; that model is a setting.
			popts.SkepticModel = settings.SkepticModel
			popts.Generator = generator.NewLive(settings, nil)
		}
		if opts.gen != nil {
			// A test supplies the generator. The live client above is not called,
			// so the review does not dial. A review leaves gen unset.
			popts.Generator = opts.gen
		}
		// ADR-0037: a key for a second reviewer connects it. Its settings are checked here, so a wrong
		// address stops Open before any review. With no key nothing is connected and the review has
		// one reviewer, as before.
		sec, connected, serr := generator.LoadSecond(env)
		if serr != nil {
			return nil, generatorConfig(serr)
		}
		if connected {
			popts.Special = generator.NewSecond(sec, nil)
			secondModel = sec.Model
		}
		// ADR-0005, ADR-0013 point 4: the redactor holds the value of every secret this review loads, so no key
		// a service echoes back reaches the trace or the result.
		popts.Redactor = redact.FromEnv(env)
		gs, gerr := guardrail.Load(env)
		if d, ok := service.As(gerr); ok {
			// ADR-0005 point 5: an empty key is not off. Every model call is
			// refused by this adapter, and the trace gets one line per refusal.
			popts.Down = append(popts.Down, *d)
			popts.Guardrail = guardrail.Absent{Reason: d.Reason}
		} else if gerr != nil {
			return nil, guardrailConfig(gerr)
		} else if gs.Mode == guardrail.ModeOff {
			// ADR-0005: the guardrail is on unless a setting turns it off. When off, the pipeline is told
			// so and the trace records one "no guardrail" line.
			popts.GuardOff = true
		} else {
			rail, nerr := guardrail.New(gs, nil)
			if nerr != nil {
				return nil, nerr
			}
			popts.Guardrail = rail
		}
		zset, zerr := query.Load(env)
		if d, ok := service.As(zerr); ok {
			popts.Down = append(popts.Down, *d)
			popts.LiveQuery = query.Absent{Reason: d.Reason}
		} else if zerr != nil {
			return nil, queryConfig(zerr)
		} else {
			popts.LiveQuery = query.NewLive(zset, nil)
		}
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
		// ADR-0013: the KEV view and the adapter bound are settings. A wrong
		// value stops Open before any review. Sample mode skips this block
		// and does not read either variable.
		kev, kerr := osv.LoadView(env)
		if kerr != nil {
			return nil, osvConfig(kerr)
		}
		kevView = kev
		bound, aerr := limit.Load(env)
		if aerr != nil {
			return nil, limitConfig(aerr)
		}
		adapterLimit = bound
	}
	p, err := pipeline.Build(popts)
	if err != nil {
		return nil, err
	}
	return &runner{p: p, cfg: cfg, explicit: explicit, kevView: kevView, adapterLimit: adapterLimit, secondModel: secondModel}, nil
}

// withSecond makes the connected second reviewer the special reviewer (ADR-0037 point 2) and widens
// the doubled classes (point 3): with a second reviewer connected every class is doubled, unless
// proven.yaml names the classes, and security is doubled either way. With none connected cfg is
// returned as it is, so one reviewer reviews exactly as before.
func withSecond(cfg fileSettings, model string) fileSettings {
	if model == "" {
		return cfg
	}
	cfg.SpecialReviewer = model
	doubled := map[string]bool{}
	if cfg.DoubledSet {
		for class, on := range cfg.Doubled {
			doubled[class] = on
		}
	} else {
		for _, class := range classNames() {
			doubled[class] = true
		}
	}
	// Security is always doubled, even under "doubled_classes: off".
	doubled["security"] = true
	cfg.Doubled = doubled
	return cfg
}

// runner is the seam's Reviewer. It copies the writer's draft into Result.
// cfg is the explicit configuration when explicit is true; otherwise it is unused and each
// Review loads the settings of the change it reviews.
type runner struct {
	p            *pipeline.Pipeline
	cfg          fileSettings
	explicit     bool
	kevView      string
	adapterLimit time.Duration
	// secondModel is the second reviewer's model when one is connected; empty otherwise.
	secondModel string
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
	cfg = withSecond(cfg, r.secondModel)
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
	// The KEV view and the adapter bound ride the context. The pipeline's
	// calls do not take them, and this change does not edit that package.
	// Sample mode leaves both unset, so the defaults apply and the
	// environment is not read.
	if r.kevView != "" {
		ctx = osv.WithKEVView(ctx, r.kevView)
	}
	if r.adapterLimit > 0 {
		ctx = limit.WithBound(ctx, r.adapterLimit)
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
		// ADR-0042: the settings that turn proposed fixes off and bound their attempts.
		FixesOff:    !cfg.Fixes,
		FixAttempts: cfg.FixAttempts,
	})
	if err != nil {
		return Result{}, err
	}
	out, err := fill(d)
	if err != nil {
		return Result{}, err
	}
	if err := applyDecision(&out, cfg, req); err != nil {
		return Result{}, err
	}
	return out, nil
}

// fill copies the writer's draft into the seam's Result. Slices are copied, so the Result shares no
// memory with the pipeline. Findings and FailedServices are never nil, so an empty review prints
// an empty list rather than a missing one.
//
// ADR-0044: fill sets each finding's ProofLevel from its strongest evidence and counts the findings by
// level. A finding whose evidence gives no level is an error, so it cannot reach the Result.
func fill(d candidate.Draft) (Result, error) {
	out := Result{
		Partial: d.Partial, TracePath: d.TracePath, ProofCounts: map[ProofLevel]int{},
		// A guardrail that was off is named in the result, not only in the trace.
		GuardrailStatement: d.GuardrailStatement,
	}
	for _, l := range ProofLevels {
		out.ProofCounts[l] = 0
	}
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
		f.Fix = toFix(c.Fix)
		for _, v := range c.Verdicts {
			f.Verdicts = append(f.Verdicts, ReviewerVerdict{Role: v.Run, Model: v.Model, Verdict: v.Verdict, Reason: v.Reason})
		}
		f.Disagree = disagree(f.Verdicts)
		if f.Disagree {
			out.Disagreements++
		}
		level, err := ProofLevelOf(f.Evidence)
		if err != nil {
			return Result{}, fmt.Errorf("finding %s: %w", c.ID, err)
		}
		f.ProofLevel = level
		out.ProofCounts[level]++
		out.Findings = append(out.Findings, f)
	}
	// ADR-0041: an opinion is copied beside the findings. It has no proof level,
	// so it is not added to ProofCounts and it cannot fail the review.
	for _, o := range d.Opinions {
		op := Opinion{
			ID: o.ID, Title: o.Title, Level: OpinionLevel,
			Location: Location{Path: o.Path, StartLine: o.StartLine, EndLine: o.EndLine},
		}
		for _, v := range o.Reasons {
			op.Reasons = append(op.Reasons, ReviewerVerdict{Role: v.Run, Model: v.Model, Verdict: v.Verdict, Reason: v.Reason})
		}
		out.Opinions = append(out.Opinions, op)
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
	if out.Opinions == nil {
		out.Opinions = []Opinion{}
	}
	if out.FailedServices == nil {
		out.FailedServices = []ServiceFailure{}
	}
	return out, nil
}

// toFix copies a proposed fix across the seam, field by field. A nil fix stays nil.
func toFix(f *candidate.Fix) *Fix {
	if f == nil {
		return nil
	}
	out := &Fix{
		Kind: f.Kind, Proved: f.Proved, Attempts: f.Attempts,
		What: f.What, Why: f.Why, Note: f.Note,
	}
	for _, e := range f.Edits {
		out.Edits = append(out.Edits, FixEdit{Path: e.Path, Old: e.Old, New: e.New})
	}
	for _, r := range f.Runs {
		out.Runs = append(out.Runs, FixRun{Label: r.Label, Command: r.Command, Result: r.Result, ExitCode: r.ExitCode, Output: r.Output})
	}
	if d := f.Dependency; d != nil {
		out.Dependency = &DependencyFix{
			Ecosystem: d.Ecosystem, Package: d.Package, From: d.From, To: d.To,
			AdvisoryID:      d.AdvisoryID,
			AdvisoryColumns: append([]string(nil), d.AdvisoryColumns...),
			AdvisoryRow:     append([]string(nil), d.AdvisoryRow...),
		}
	}
	return out
}

// disagree reports whether the reviewers who answered gave different verdicts on one finding
// (ADR-0037 point 4). The verdicts are compared, never merged: two different verdicts stay two.
// A reviewer with no answer is left out, because a failed call says nothing about the finding.
func disagree(vs []ReviewerVerdict) bool {
	first := ""
	for _, v := range vs {
		if v.Verdict == VerdictNoAnswer {
			continue
		}
		if first == "" {
			first = v.Verdict
			continue
		}
		if v.Verdict != first {
			return true
		}
	}
	return false
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
	if e.Tool != nil {
		// ADR-0040: the tool, its version, the command and both sides cross the seam together.
		t := *e.Tool
		out.Tool = &ToolEvidence{
			Tool: t.Tool, Version: t.Version, Command: t.Command,
			Output: t.Output, BaseOutput: t.BaseOutput,
			PresentOnChange: t.PresentOnChange, PresentOnBase: t.PresentOnBase,
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

// osvConfig turns a refused OSV setting into the seam's *ConfigError.
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

// limitConfig turns a refused adapter-bound setting into the seam's *ConfigError.
func limitConfig(err error) error {
	var cfg *limit.ConfigError
	if errors.As(err, &cfg) {
		return &ConfigError{
			Field: cfg.Field, Value: cfg.Value, Reason: cfg.Reason,
			Allowed: append([]string(nil), cfg.Allowed...),
		}
	}
	return err
}

// readLimitConfig turns a refused read-bound setting into the seam's *ConfigError.
func readLimitConfig(err error) error {
	var cfg *change.LimitError
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
