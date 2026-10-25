package review

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/telemetry"
)

// memTelemetry is telemetry over in-memory exporters: the spans and the numbers of a review can be read back.
type memTelemetry struct {
	t       *telemetry.T
	spans   *tracetest.InMemoryExporter
	metrics *sdkmetric.ManualReader
}

func newMemTelemetry(t *testing.T) *memTelemetry {
	t.Helper()
	spans := tracetest.NewInMemoryExporter()
	reader := sdkmetric.NewManualReader()
	tel, err := telemetry.New(
		sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans)),
		sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return &memTelemetry{t: tel, spans: spans, metrics: reader}
}

// reviewWithTelemetry runs one scripted, doubled review with tel watching it.
func reviewWithTelemetry(t *testing.T, tel *telemetry.T, rd *redact.Redactor, def, spec *scriptGen) Result {
	t.Helper()
	p, err := pipeline.Build(pipeline.Options{
		Sample: true, LogDir: t.TempDir(), Generator: def, Special: spec, Telemetry: tel, Redactor: rd,
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultSettings()
	cfg.SpecialReviewer = "qwen-special"
	r := &runner{p: p, cfg: cfg, explicit: true, tel: tel, sample: true}
	res, err := r.Review(context.Background(), Request{Target: discountFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// twoReviewers is a doubled review: B1 is named by both reviewers, D1 only by the default one.
// The special Skeptic upholds B1 where the default one refutes it, so the two disagree on B1.
func twoReviewers() (def, spec *scriptGen) {
	def = newScript("NVIDIA", risks(risk("B1", "seen by both"), risk("D1", "default only")),
		`{"refuted":true,"reason":"default-refutes"}`)
	def.sql = map[string]string{"B1": rowAnswer, "D1": rowAnswer}
	spec = newScript("qwen-special", risks(risk("B1", "seen by both")),
		`{"refuted":false,"reason":"special-upholds"}`)
	spec.sql = map[string]string{"B1": rowAnswer}
	return def, spec
}

func attrOf(kvs []attribute.KeyValue, key string) string {
	for _, kv := range kvs {
		if string(kv.Key) == key {
			return kv.Value.Emit()
		}
	}
	return ""
}

// sumOf adds the value of every point of the named counter whose attributes include want.
func sumOf(t *testing.T, rm metricdata.ResourceMetrics, name string, want ...attribute.KeyValue) int64 {
	t.Helper()
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("%s is %T, not an int64 sum", name, m.Data)
			}
			for _, dp := range sum.DataPoints {
				match := true
				for _, w := range want {
					if v, ok := dp.Attributes.Value(w.Key); !ok || v != w.Value {
						match = false
					}
				}
				if match {
					total += dp.Value
				}
			}
		}
	}
	return total
}

func collect(t *testing.T, m *memTelemetry) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := m.metrics.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	return rm
}

// TestReviewIsOneTraceWithAnAgentSpanAndTheTwoReviewers is ADR-0045 point 1: one trace per review, one span per
// agent and per reviewer model, the model calls under the agent that made them, the trace file's events on the
// spans. It would fail if a reviewer's agent had no span of its own, if a model call sat outside its agent, or if
// the review split into two traces.
func TestReviewIsOneTraceWithAnAgentSpanAndTheTwoReviewers(t *testing.T) {
	mem := newMemTelemetry(t)
	def, spec := twoReviewers()
	reviewWithTelemetry(t, mem.t, nil, def, spec)

	spans := mem.spans.GetSpans()
	if len(spans) == 0 {
		t.Fatal("no span was exported")
	}
	trace := spans[0].SpanContext.TraceID()
	byID := map[string]tracetest.SpanStub{}
	roots := 0
	for _, s := range spans {
		if s.SpanContext.TraceID() != trace {
			t.Fatalf("span %q is in another trace: a review is one trace", s.Name)
		}
		byID[s.SpanContext.SpanID().String()] = s
		if s.Name == "proven.review" {
			roots++
		}
	}
	if roots != 1 {
		t.Fatalf("%d root spans, want 1", roots)
	}
	has := func(name, reviewer string) tracetest.SpanStub {
		t.Helper()
		for _, s := range spans {
			if s.Name == name && attrOf(s.Attributes, "proven.reviewer") == reviewer {
				return s
			}
		}
		t.Fatalf("no span %q for reviewer %q; have %d spans", name, reviewer, len(spans))
		return tracetest.SpanStub{}
	}
	has("agent planner", telemetry.ReviewerDefault)
	has("agent planner", telemetry.ReviewerSpecial)
	has("agent skeptic", telemetry.ReviewerDefault)
	has("agent skeptic", telemetry.ReviewerSpecial)
	for _, agent := range []string{"investigator", "prover", "judge", "writer"} {
		has("agent "+agent, telemetry.ReviewerShared)
	}
	// A model call is a child of the agent that made it, and it names its reviewer's model.
	special := has("chat NVIDIA", telemetry.ReviewerDefault)
	if attrOf(special.Attributes, "gen_ai.operation.name") != "chat" {
		t.Fatalf("model call without gen_ai.operation.name: %v", special.Attributes)
	}
	calls := 0
	for _, s := range spans {
		if !strings.HasPrefix(s.Name, "chat") {
			continue
		}
		calls++
		parent, ok := byID[s.Parent.SpanID().String()]
		if !ok || !strings.HasPrefix(parent.Name, "agent ") {
			t.Fatalf("model call %q has parent %q, want an agent span", s.Name, parent.Name)
		}
		if attrOf(s.Attributes, "proven.agent") != strings.TrimPrefix(parent.Name, "agent ") {
			t.Fatalf("model call of %s sits under %s", attrOf(s.Attributes, "proven.agent"), parent.Name)
		}
	}
	if calls == 0 {
		t.Fatal("no model call span")
	}
	has("chat qwen-special", telemetry.ReviewerSpecial)
	// The trace file's events ride along: the claim, the evidence and the verdict are span events.
	seen := map[string]bool{}
	for _, s := range spans {
		for _, e := range s.Events {
			seen[e.Name] = true
		}
	}
	for _, kind := range []string{"claim", "evidence", "refutation", "verdict", "model", "guardrail"} {
		if !seen[kind] {
			t.Fatalf("no span event of kind %q; have %v", kind, seen)
		}
	}
}

// TestReviewCountsWhatADashboardShows is ADR-0045 point 1's numbers: findings by level of proof, claims dropped,
// reviewer verdicts and their disagreements, model calls and their duration, guardrail checks.
func TestReviewCountsWhatADashboardShows(t *testing.T) {
	mem := newMemTelemetry(t)
	def, spec := twoReviewers()
	res := reviewWithTelemetry(t, mem.t, nil, def, spec)
	rm := collect(t, mem)

	var kept int64
	for level, n := range res.ProofCounts {
		got := sumOf(t, rm, "proven.review.findings", attribute.String("proven.proof_level", string(level)))
		if got != int64(n) {
			t.Fatalf("findings at %s: metric %d, result %d", level, got, n)
		}
		kept += int64(n)
	}
	if kept == 0 {
		t.Fatal("the review kept no finding, so the count above proves nothing")
	}
	// The default Skeptic refutes B1 and the special one upholds it: one disagreement on a verdict.
	// D1 is named by one reviewer only: one disagreement of the other kind.
	if got := sumOf(t, rm, "proven.reviewer.disagreements", attribute.String("kind", "verdict")); got != 1 {
		t.Fatalf("verdict disagreements %d, want 1", got)
	}
	if got := sumOf(t, rm, "proven.reviewer.disagreements", attribute.String("kind", "one_reviewer")); got != 1 {
		t.Fatalf("one-reviewer disagreements %d, want 1", got)
	}
	if got := sumOf(t, rm, "proven.reviewer.verdicts", attribute.String("proven.reviewer", "default"), attribute.String("verdict", "refuted")); got < 1 {
		t.Fatalf("default reviewer's refutations %d, want at least 1", got)
	}
	if got := sumOf(t, rm, "proven.reviewer.verdicts", attribute.String("proven.reviewer", "special"), attribute.String("verdict", "upheld")); got < 1 {
		t.Fatalf("special reviewer's upheld verdicts %d, want at least 1", got)
	}
	if got := sumOf(t, rm, "proven.model.calls", attribute.String("proven.reviewer", "special")); got < 2 {
		t.Fatalf("special reviewer's model calls %d, want the planner's and the Skeptic's", got)
	}
	if got := sumOf(t, rm, "proven.guardrail.checks"); got < 1 {
		t.Fatalf("guardrail checks %d, want at least 1", got)
	}
	var durations uint64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "gen_ai.client.operation.duration" {
				for _, dp := range m.Data.(metricdata.Histogram[float64]).DataPoints {
					durations += dp.Count
				}
			}
		}
	}
	if durations == 0 {
		t.Fatal("no model call duration was recorded")
	}
}

// TestDroppedClaimsAreCounted: a claim with no evidence is dropped by the Judge, and the count says why.
func TestDroppedClaimsAreCounted(t *testing.T) {
	mem := newMemTelemetry(t)
	def := newScript("NVIDIA", risks(risk("N1", "no evidence from anyone")), `{"refuted":false,"reason":"x"}`)
	spec := newScript("qwen-special", risks(), `{"refuted":false,"reason":"x"}`)
	res := reviewWithTelemetry(t, mem.t, nil, def, spec)
	for _, f := range res.Findings {
		if f.ID == "N1" {
			t.Fatal("N1 has no evidence and must not be kept")
		}
	}
	rm := collect(t, mem)
	if got := sumOf(t, rm, "proven.review.claims.dropped", attribute.String("reason", "no_evidence")); got != 1 {
		t.Fatalf("claims dropped for no evidence: %d, want 1", got)
	}
}

// TestNoSecretReachesASpanOrANumber is ADR-0045 point 2: the redactor applies. The claim's title, the model's
// name and the finding's text carry a configured secret; no attribute or event value of any span holds it.
func TestNoSecretReachesASpanOrANumber(t *testing.T) {
	const secret = "sk-not-for-the-trace-9f2c"
	mem := newMemTelemetry(t)
	rd := redact.New(map[string]string{"TEST_SECRET_KEY": secret})
	def := newScript("NVIDIA-"+secret, risks(risk("B1", "claim echoing "+secret)), `{"refuted":false,"reason":"echo `+secret+`"}`)
	def.sql = map[string]string{"B1": rowAnswer}
	spec := newScript("qwen-special", risks(), `{"refuted":false,"reason":"x"}`)
	reviewWithTelemetry(t, mem.t, rd, def, spec)

	for _, s := range mem.spans.GetSpans() {
		for _, kv := range s.Attributes {
			if strings.Contains(kv.Value.Emit(), secret) {
				t.Fatalf("span %q attribute %s holds the secret", s.Name, kv.Key)
			}
		}
		for _, e := range s.Events {
			for _, kv := range e.Attributes {
				if strings.Contains(kv.Value.Emit(), secret) {
					t.Fatalf("span %q event %q attribute %s holds the secret", s.Name, e.Name, kv.Key)
				}
			}
		}
	}
	rm := collect(t, mem)
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			for _, dp := range dataPointAttrs(m) {
				for _, kv := range dp.ToSlice() {
					if strings.Contains(kv.Value.Emit(), secret) {
						t.Fatalf("metric %s attribute %s holds the secret", m.Name, kv.Key)
					}
				}
			}
		}
	}
	// The control: the secret was in the run, so the test above could fail. The model's name reached the span
	// as the redactor's marker.
	found := false
	for _, s := range mem.spans.GetSpans() {
		if strings.Contains(attrOf(s.Attributes, "gen_ai.response.model"), "NVIDIA-") {
			found = true
		}
	}
	if !found {
		t.Fatal("control: no span named the response model, so the secret never had a way in")
	}
}

func dataPointAttrs(m metricdata.Metrics) []attribute.Set {
	var out []attribute.Set
	switch d := m.Data.(type) {
	case metricdata.Sum[int64]:
		for _, p := range d.DataPoints {
			out = append(out, p.Attributes)
		}
	case metricdata.Gauge[int64]:
		for _, p := range d.DataPoints {
			out = append(out, p.Attributes)
		}
	case metricdata.Histogram[float64]:
		for _, p := range d.DataPoints {
			out = append(out, p.Attributes)
		}
	}
	return out
}

// collector is a stand-in OTLP/HTTP endpoint that counts what it is sent.
type collector struct {
	srv      *httptest.Server
	traces   atomic.Int64
	metrics  atomic.Int64
	requests atomic.Int64
}

func newCollector(t *testing.T) *collector {
	t.Helper()
	c := &collector{}
	c.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		c.requests.Add(1)
		switch r.URL.Path {
		case "/v1/traces":
			c.traces.Add(1)
		case "/v1/metrics":
			c.metrics.Add(1)
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(c.srv.Close)
	return c
}

// sampleReviewWith opens a sample review with the given environment and runs it once.
func sampleReviewWith(t *testing.T, env map[string]string) {
	t.Helper()
	rev, err := Open(Options{
		Sample: true, LogDir: t.TempDir(),
		Env: func(k string) string { return env[k] },
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rev.Review(context.Background(), Request{Target: discountFixture(t)}); err != nil {
		t.Fatal(err)
	}
}

// TestAReviewIsExportedOnlyWhenAnEndpointIsSet is ADR-0045 point 2 with its control. The same sample review runs
// twice against a stand-in collector. With OTEL_EXPORTER_OTLP_ENDPOINT set, the collector is sent the trace and the
// metrics (the control: it fails when the exporter is not wired). With it unset, the collector is sent nothing,
// and neither is any other address: the review opens no connection.
func TestAReviewIsExportedOnlyWhenAnEndpointIsSet(t *testing.T) {
	c := newCollector(t)

	sampleReviewWith(t, map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT": c.srv.URL,
		"OTEL_SERVICE_NAME":           "proven-test",
	})
	if c.traces.Load() == 0 || c.metrics.Load() == 0 {
		t.Fatalf("control: with an endpoint set the collector got %d trace and %d metric requests, want both", c.traces.Load(), c.metrics.Load())
	}

	before := c.requests.Load()
	sampleReviewWith(t, nil)
	if got := c.requests.Load() - before; got != 0 {
		t.Fatalf("with no endpoint set the collector was sent %d requests, want 0", got)
	}
}

// TestADisabledSDKExportsNothing: OTEL_SDK_DISABLED is the standard off switch and it wins over an endpoint.
func TestADisabledSDKExportsNothing(t *testing.T) {
	c := newCollector(t)
	sampleReviewWith(t, map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT": c.srv.URL,
		"OTEL_SDK_DISABLED":           "true",
	})
	if got := c.requests.Load(); got != 0 {
		t.Fatalf("a disabled SDK sent %d requests, want 0", got)
	}
}

// TestABadTelemetrySettingIsRefusedByName: Open refuses a wrong OTEL_* value before any review, as for any setting.
func TestABadTelemetrySettingIsRefusedByName(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"OTEL_EXPORTER_OTLP_PROTOCOL": {"OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4318", "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc"},
		"OTEL_EXPORTER_OTLP_ENDPOINT": {"OTEL_EXPORTER_OTLP_ENDPOINT": "collector:4318"},
		"OTEL_METRIC_EXPORT_INTERVAL": {"OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4318", "OTEL_METRIC_EXPORT_INTERVAL": "soon"},
	} {
		_, err := Open(Options{Sample: true, LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
		cfg, ok := err.(*ConfigError)
		if !ok {
			t.Fatalf("%s: error %v, want *ConfigError", name, err)
		}
		if cfg.Field != name {
			t.Fatalf("%s: refusal names field %q", name, cfg.Field)
		}
	}
}
