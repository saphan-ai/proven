// Package telemetry makes a review observable through OpenTelemetry (ADR-0045).
//
// One review is one trace. Each agent, and each model call inside it, is a
// span; each line of the trace file rides along as a span event. Counters and
// histograms carry the numbers a dashboard needs: findings by level of proof,
// claims dropped, reviewer verdicts and disagreements, model calls and their
// duration, guardrail checks. Telemetry only watches. It never changes what a
// review decides, and a failure to export is never a failure of the review.
//
// Telemetry is off unless the standard OTEL_* variables name an endpoint
// (setup.go). A nil *T is the off state: every method is safe on it and does
// nothing, so the pipeline needs no "if enabled" around a call.
package telemetry

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	ptrace "github.com/saphan-ai/proven/internal/trace"
)

// scope is the instrumentation scope every span and instrument is made under.
const scope = "github.com/saphan-ai/proven"

// Agent states, as the proven.agent.state gauge reports them. The dashboard's
// state timeline maps these numbers to words.
const (
	StateWaiting = 0
	StateWorking = 1
	StateDone    = 2
	StateFailed  = 3
)

// Reviewer names the run a reviewer model belongs to. Shared names an agent
// that does not belong to one reviewer.
const (
	ReviewerDefault = "default"
	ReviewerSpecial = "special"
	ReviewerShared  = "shared"
)

// Flusher is what a provider offers to push what it holds to the exporter.
type Flusher interface {
	ForceFlush(ctx context.Context) error
}

// maxEventText bounds the text copied into one span event. The redactor has
// already run on it; the bound keeps one long evidence line out of the trace
// backend.
const maxEventText = 256

// T is the telemetry of one reviewer. The zero value is not usable; New and
// Setup build one, and a nil *T is off.
type T struct {
	tracer trace.Tracer
	flush  []Flusher

	findings      metric.Int64Counter
	dropped       metric.Int64Counter
	verdicts      metric.Int64Counter
	disagreements metric.Int64Counter
	modelCalls    metric.Int64Counter
	modelDuration metric.Float64Histogram
	guardChecks   metric.Int64Counter
	guardUnsafe   metric.Int64Counter
	agentState    metric.Int64Gauge

	mu      sync.Mutex
	current trace.Span // the span a trace-file event is attached to
	open    []*Stage   // stages not yet ended, ended with the review
}

// New builds a T over the given providers. A provider that also implements
// Flusher is flushed by Flush. Tests pass in-memory providers here.
func New(tp trace.TracerProvider, mp metric.MeterProvider) (*T, error) {
	t := &T{tracer: tp.Tracer(scope)}
	if f, ok := tp.(Flusher); ok {
		t.flush = append(t.flush, f)
	}
	if f, ok := mp.(Flusher); ok {
		t.flush = append(t.flush, f)
	}
	m := mp.Meter(scope)
	var err error
	if t.findings, err = m.Int64Counter("proven.review.findings",
		metric.WithDescription("Findings a review kept, by level of proof (ADR-0044)")); err != nil {
		return nil, err
	}
	if t.dropped, err = m.Int64Counter("proven.review.claims.dropped",
		metric.WithDescription("Claims the Judge did not keep, by reason")); err != nil {
		return nil, err
	}
	if t.verdicts, err = m.Int64Counter("proven.reviewer.verdicts",
		metric.WithDescription("The Skeptic's verdict on a finding, by reviewer")); err != nil {
		return nil, err
	}
	if t.disagreements, err = m.Int64Counter("proven.reviewer.disagreements",
		metric.WithDescription("Findings on which the two reviewers do not agree")); err != nil {
		return nil, err
	}
	if t.modelCalls, err = m.Int64Counter("proven.model.calls",
		metric.WithDescription("Model calls, by agent, reviewer and outcome")); err != nil {
		return nil, err
	}
	if t.modelDuration, err = m.Float64Histogram("gen_ai.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of a model call (gen_ai semantic conventions)"),
		metric.WithExplicitBucketBoundaries(0.25, 0.5, 1, 2, 4, 8, 16, 32, 64, 128)); err != nil {
		return nil, err
	}
	if t.guardChecks, err = m.Int64Counter("proven.guardrail.checks",
		metric.WithDescription("Guardrail checks, by side and verdict")); err != nil {
		return nil, err
	}
	if t.guardUnsafe, err = m.Int64Counter("proven.guardrail.unsafe",
		metric.WithDescription("Guardrail checks that returned unsafe")); err != nil {
		return nil, err
	}
	if t.agentState, err = m.Int64Gauge("proven.agent.state",
		metric.WithDescription("State of an agent: 0 waiting, 1 working, 2 done, 3 failed")); err != nil {
		return nil, err
	}
	return t, nil
}

// Enabled reports whether this T exports anything.
func (t *T) Enabled() bool { return t != nil }

// Flush pushes what is buffered to the exporter. It is bounded by ctx. A
// failure is returned for the caller to log; it never fails a review.
func (t *T) Flush(ctx context.Context) error {
	if t == nil {
		return nil
	}
	var first error
	for _, f := range t.flush {
		if err := f.ForceFlush(ctx); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// agents are the agents that belong to no one reviewer. A new review reports
// each as waiting, so the timeline shows the whole row from the start. The
// planner and the Skeptic wait once per reviewer (ExpectReviewer).
var agents = []string{"investigator", "prover", "judge", "writer"}

// Review is one review's root span.
type Review struct {
	t    *T
	span trace.Span
}

// StartReview opens the trace of one review. The returned context carries the
// root span; every agent and model span below it joins the same trace.
// doubled says whether a second reviewer model runs, so its row is announced.
func (t *T) StartReview(ctx context.Context, attrs ...attribute.KeyValue) (context.Context, *Review) {
	if t == nil {
		return ctx, nil
	}
	ctx, span := t.tracer.Start(ctx, "proven.review", trace.WithAttributes(attrs...))
	t.mu.Lock()
	t.current = span
	t.mu.Unlock()
	for _, a := range agents {
		t.state(ctx, a, ReviewerShared, StateWaiting)
	}
	t.ExpectReviewer(ctx, ReviewerDefault)
	return ctx, &Review{t: t, span: span}
}

// ExpectReviewer announces a reviewer's agents as waiting, so a second
// reviewer shows on the timeline before it starts. The review calls it for the
// special reviewer when the class doubles the review.
func (t *T) ExpectReviewer(ctx context.Context, reviewer string) {
	if t == nil {
		return
	}
	t.state(ctx, "planner", reviewer, StateWaiting)
	t.state(ctx, "skeptic", reviewer, StateWaiting)
}

// End closes the review. A stage still open (the review stopped early) is
// ended as failed first, so the timeline does not show it working forever.
func (r *Review) End(err error, attrs ...attribute.KeyValue) {
	if r == nil {
		return
	}
	r.t.mu.Lock()
	left := append([]*Stage(nil), r.t.open...)
	r.t.mu.Unlock()
	for _, s := range left {
		s.End(context.Canceled)
	}
	r.span.SetAttributes(attrs...)
	if err != nil {
		r.span.SetStatus(codes.Error, err.Error())
	}
	r.span.End()
	r.t.mu.Lock()
	r.t.current = nil
	r.t.mu.Unlock()
}

// Findings counts a finished review's findings by level of proof.
func (t *T) Findings(ctx context.Context, byLevel map[string]int) {
	if t == nil {
		return
	}
	for level, n := range byLevel {
		t.findings.Add(ctx, int64(n), metric.WithAttributes(attribute.String("proven.proof_level", level)))
	}
}

// Stage is one agent at work: a span, and the agent's state on the gauge.
type Stage struct {
	t        *T
	span     trace.Span
	prev     trace.Span
	agent    string
	reviewer string
	once     sync.Once
}

// StartStage opens an agent's span under ctx. reviewer is ReviewerDefault,
// ReviewerSpecial or ReviewerShared; model is the reviewer model the agent
// asks, "" when it asks none.
func (t *T) StartStage(ctx context.Context, agent, reviewer, model string) (context.Context, *Stage) {
	if t == nil {
		return ctx, nil
	}
	attrs := []attribute.KeyValue{
		attribute.String("proven.agent", agent),
		attribute.String("proven.reviewer", reviewer),
	}
	if model != "" {
		attrs = append(attrs, attribute.String("gen_ai.request.model", model))
	}
	ctx, span := t.tracer.Start(ctx, "agent "+agent, trace.WithAttributes(attrs...))
	s := &Stage{t: t, span: span, agent: agent, reviewer: reviewer}
	t.mu.Lock()
	s.prev = t.current
	t.current = span
	t.open = append(t.open, s)
	t.mu.Unlock()
	t.state(ctx, agent, reviewer, StateWorking)
	return withReviewer(ctx, agent, reviewer, model), s
}

// End closes the stage. err nil is done; any error is failed. A second call
// does nothing.
func (s *Stage) End(err error) {
	if s == nil {
		return
	}
	s.once.Do(func() {
		st := int64(StateDone)
		if err != nil {
			st = StateFailed
			s.span.SetStatus(codes.Error, err.Error())
		}
		s.t.state(context.Background(), s.agent, s.reviewer, st)
		s.span.End()
		s.t.mu.Lock()
		s.t.current = s.prev
		for i, o := range s.t.open {
			if o == s {
				s.t.open = append(s.t.open[:i], s.t.open[i+1:]...)
				break
			}
		}
		s.t.mu.Unlock()
	})
}

func (t *T) state(ctx context.Context, agent, reviewer string, state int64) {
	t.agentState.Record(ctx, state, metric.WithAttributes(
		attribute.String("proven.agent", agent), attribute.String("proven.reviewer", reviewer)))
}

// who is what a model call needs to know of the agent that makes it.
type who struct{ agent, reviewer, model string }

type whoKey struct{}

func withReviewer(ctx context.Context, agent, reviewer, model string) context.Context {
	return context.WithValue(ctx, whoKey{}, who{agent, reviewer, model})
}

// WithReviewer names the reviewer and the model a call under ctx belongs to.
// The Skeptic's two runs share one stage kind and are told apart by it.
func WithReviewer(ctx context.Context, reviewer, model string) context.Context {
	w, _ := ctx.Value(whoKey{}).(who)
	w.reviewer, w.model = reviewer, model
	return context.WithValue(ctx, whoKey{}, w)
}

// ModelCall times one generator call (gen_ai semantic conventions: a client
// span named "chat <model>", and the operation duration histogram).
type ModelCall struct {
	t     *T
	span  trace.Span
	start time.Time
	w     who
	agent string
}

// StartModelCall opens the span of one call to a model.
func (t *T) StartModelCall(ctx context.Context, agent string) (context.Context, *ModelCall) {
	if t == nil {
		return ctx, nil
	}
	w, _ := ctx.Value(whoKey{}).(who)
	if w.reviewer == "" {
		w.reviewer = ReviewerShared
	}
	name := "chat"
	if w.model != "" {
		name += " " + w.model
	}
	ctx, span := t.tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.request.model", w.model),
		attribute.String("proven.agent", agent),
		attribute.String("proven.reviewer", w.reviewer),
	))
	return ctx, &ModelCall{t: t, span: span, start: time.Now(), w: w, agent: agent}
}

// End closes the call. responseModel is the model the provider named, "" when
// it named none. err nil is a call that answered.
func (c *ModelCall) End(ctx context.Context, responseModel string, err error) {
	if c == nil {
		return
	}
	secs := time.Since(c.start).Seconds()
	outcome := "ok"
	attrs := []attribute.KeyValue{
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.request.model", c.w.model),
		attribute.String("proven.agent", c.agent),
		attribute.String("proven.reviewer", c.w.reviewer),
	}
	if responseModel != "" {
		attrs = append(attrs, attribute.String("gen_ai.response.model", responseModel))
		c.span.SetAttributes(attribute.String("gen_ai.response.model", responseModel))
	}
	if err != nil {
		outcome = "error"
		attrs = append(attrs, attribute.String("error.type", "generator"))
		c.span.SetAttributes(attribute.String("error.type", "generator"))
		c.span.SetStatus(codes.Error, "model call failed")
	}
	c.t.modelDuration.Record(ctx, secs, metric.WithAttributes(attrs...))
	c.t.modelCalls.Add(ctx, 1, metric.WithAttributes(
		attribute.String("proven.agent", c.agent),
		attribute.String("proven.reviewer", c.w.reviewer),
		attribute.String("outcome", outcome)))
	c.span.End()
}

// Verdict counts the Skeptic's verdict on one finding for one reviewer:
// "refuted", "upheld" or "not_attempted".
func (t *T) Verdict(ctx context.Context, reviewer, verdict string) {
	if t == nil {
		return
	}
	t.verdicts.Add(ctx, 1, metric.WithAttributes(
		attribute.String("proven.reviewer", reviewer), attribute.String("verdict", verdict)))
}

// Disagreement counts one finding on which the two reviewers differ. kind is
// "verdict" (both ran it, and they differ) or "one_reviewer" (only one named it).
func (t *T) Disagreement(ctx context.Context, kind string) {
	if t == nil {
		return
	}
	t.disagreements.Add(ctx, 1, metric.WithAttributes(attribute.String("kind", kind)))
}

// Observe is the trace writer's observer: each event of the trace file becomes
// an event on the span that is open, and the events that carry a number feed
// the instruments. The event arrives already redacted. Nothing here fails.
func (t *T) Observe(e ptrace.Event) {
	if t == nil {
		return
	}
	ctx := context.Background()
	t.mu.Lock()
	span := t.current
	t.mu.Unlock()
	if span != nil {
		text := e.Text
		if len(text) > maxEventText {
			text = text[:maxEventText] + "…"
		}
		attrs := []attribute.KeyValue{
			attribute.String("proven.event.kind", e.Kind),
			attribute.String("proven.event.agent", e.Agent),
			attribute.String("proven.event.text", text),
		}
		if e.FindingID != "" {
			attrs = append(attrs, attribute.String("proven.finding", e.FindingID))
		}
		if e.Run != "" {
			attrs = append(attrs, attribute.String("proven.reviewer", e.Run))
		}
		span.AddEvent(e.Kind, trace.WithAttributes(attrs...))
	}
	switch e.Kind {
	case "guardrail":
		verdict, side := "refused", ""
		if e.Guardrail != nil {
			verdict, side = e.Guardrail.Verdict, e.Guardrail.Side
		}
		t.guardChecks.Add(ctx, 1, metric.WithAttributes(
			attribute.String("side", side), attribute.String("verdict", verdict)))
		if verdict == "unsafe" {
			t.guardUnsafe.Add(ctx, 1, metric.WithAttributes(attribute.String("side", side)))
		}
	case "verdict":
		if reason := dropReason(e.Text); reason != "" {
			t.dropped.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
		}
	case "memory":
		// H-MEM matched an earlier refutation: the claim never reaches the Judge.
		t.dropped.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", "suppressed")))
	}
}

// dropReason names the kind of drop a Judge verdict line states, or "" when
// the finding was kept. The kinds are a closed set, so the label stays small.
func dropReason(text string) string {
	switch {
	case text == "kept", text == "", strings.HasPrefix(text, "unresolved"):
		// An unresolved finding stays in the review, marked, so it is not a drop.
		return ""
	case strings.HasPrefix(text, "dropped: refuted"):
		return "refuted"
	case strings.HasPrefix(text, "dropped: no evidence"):
		return "no_evidence"
	case strings.HasPrefix(text, "suppressed"):
		return "suppressed"
	}
	return "other"
}
