package telemetry

import (
	"context"
	"errors"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	ptrace "github.com/saphan-ai/proven/internal/trace"
)

func mem(t *testing.T) (*T, *tracetest.InMemoryExporter, *sdkmetric.ManualReader) {
	t.Helper()
	spans := tracetest.NewInMemoryExporter()
	reader := sdkmetric.NewManualReader()
	tel, err := New(sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans)), sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	if err != nil {
		t.Fatal(err)
	}
	return tel, spans, reader
}

// states reads the agent-state gauge: "agent/reviewer" to its last value.
func states(t *testing.T, reader *sdkmetric.ManualReader) map[string]int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	out := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if g, ok := m.Data.(metricdata.Gauge[int64]); ok && m.Name == "proven.agent.state" {
				for _, dp := range g.DataPoints {
					a, _ := dp.Attributes.Value("proven.agent")
					r, _ := dp.Attributes.Value("proven.reviewer")
					out[a.AsString()+"/"+r.AsString()] = dp.Value
				}
			}
		}
	}
	return out
}

// TestANilTelemetryIsOff: the off state is a nil *T, and every call on it is safe and does nothing, so the
// pipeline needs no "if enabled" around a call.
func TestANilTelemetryIsOff(t *testing.T) {
	var tel *T
	ctx, root := tel.StartReview(context.Background())
	ctx, stage := tel.StartStage(ctx, "planner", ReviewerDefault, "m")
	ctx, call := tel.StartModelCall(ctx, "planner")
	call.End(ctx, "m", errors.New("x"))
	stage.End(nil)
	tel.Verdict(ctx, "default", "upheld")
	tel.Disagreement(ctx, "verdict")
	tel.Findings(ctx, map[string]int{"proved-by-test": 1})
	tel.ExpectReviewer(ctx, ReviewerSpecial)
	tel.Observe(ptrace.Event{Kind: "claim", Text: "x"})
	root.End(nil)
	if err := tel.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if tel.Enabled() {
		t.Fatal("a nil telemetry is enabled")
	}
}

// TestTheTimelineShowsEveryAgentFromTheStart: a new review reports each agent as waiting, a reviewer announced
// later joins as waiting, a running agent is working and a finished one is done. It would fail if a row only
// appeared when its agent started.
func TestTheTimelineShowsEveryAgentFromTheStart(t *testing.T) {
	tel, _, reader := mem(t)
	ctx, root := tel.StartReview(context.Background())
	got := states(t, reader)
	for _, k := range []string{"planner/default", "skeptic/default", "investigator/shared", "prover/shared", "judge/shared", "writer/shared"} {
		if got[k] != StateWaiting {
			t.Fatalf("%s is %d at the start, want waiting; have %v", k, got[k], got)
		}
	}
	if _, ok := got["planner/special"]; ok {
		t.Fatal("a second reviewer is on the timeline before it was announced")
	}
	tel.ExpectReviewer(ctx, ReviewerSpecial)
	_, stage := tel.StartStage(ctx, "planner", ReviewerSpecial, "m")
	if got := states(t, reader); got["planner/special"] != StateWorking || got["planner/default"] != StateWaiting {
		t.Fatalf("while the special planner works: %v", got)
	}
	stage.End(nil)
	if got := states(t, reader); got["planner/special"] != StateDone {
		t.Fatalf("after the special planner: %v", got)
	}
	root.End(nil)
}

// TestAStageLeftOpenIsFailedWhenTheReviewEnds: a review that stops early (an error, a deadline) must not leave an
// agent working on the dashboard forever. A stage ended twice counts once.
func TestAStageLeftOpenIsFailedWhenTheReviewEnds(t *testing.T) {
	tel, spans, reader := mem(t)
	ctx, root := tel.StartReview(context.Background())
	_, left := tel.StartStage(ctx, "prover", ReviewerShared, "")
	_, done := tel.StartStage(ctx, "investigator", ReviewerShared, "")
	done.End(nil)
	done.End(errors.New("a second end must change nothing"))
	root.End(errors.New("stopped"))
	left.End(nil) // already ended by the review; no effect
	got := states(t, reader)
	if got["prover/shared"] != StateFailed {
		t.Fatalf("prover left open is %d, want failed; have %v", got["prover/shared"], got)
	}
	if got["investigator/shared"] != StateDone {
		t.Fatalf("investigator is %d, want done", got["investigator/shared"])
	}
	for _, s := range spans.GetSpans() {
		if s.Name == "agent investigator" && s.Status.Code != 0 {
			t.Fatalf("a second End changed the span's status: %v", s.Status)
		}
	}
}

// TestADropReasonIsAClosedSet: the label a drop carries is one of a few words, never the Judge's free text.
func TestADropReasonIsAClosedSet(t *testing.T) {
	for text, want := range map[string]string{
		"kept":                         "",
		"":                             "",
		"unresolved: not settled":      "",
		"dropped: refuted: it is fine": "refuted",
		"dropped: no evidence":         "no_evidence",
		"suppressed: hmem abc":         "suppressed",
		"something new":                "other",
	} {
		if got := dropReason(text); got != want {
			t.Errorf("dropReason(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestSetupIsOffWithoutAnEndpoint(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"empty":               {},
		"only a service name": {"OTEL_SERVICE_NAME": "x"},
		"disabled, endpoint":  {"OTEL_SDK_DISABLED": "true", "OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4318"},
		"a blank endpoint":    {"OTEL_EXPORTER_OTLP_ENDPOINT": "  "},
	} {
		tel, err := Setup(context.Background(), func(k string) string { return env[k] })
		if err != nil || tel != nil {
			t.Errorf("%s: Setup = %v, %v; want off", name, tel, err)
		}
	}
	if tel, err := Setup(context.Background(), nil); err != nil || tel != nil {
		t.Errorf("no environment: Setup = %v, %v; want off", tel, err)
	}
}
