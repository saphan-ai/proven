package review

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
)

// failingGen is a scripted generator whose planner call fails with err.
type failingGen struct {
	*scriptGen
	err error
}

func (g failingGen) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "planner" {
		return "", g.err
	}
	return g.scriptGen.Complete(ctx, agent, prompt)
}

// failedReview runs one review whose default planner fails with err, and returns what the review returned.
func failedReview(t *testing.T, mem *memTelemetry, rd *redact.Redactor, err error) (Result, error) {
	t.Helper()
	def := failingGen{scriptGen: newScript("NVIDIA", risks(), `{"refuted":false,"reason":"x"}`), err: err}
	p, berr := pipeline.Build(pipeline.Options{
		Sample: true, LogDir: t.TempDir(), Generator: def, Telemetry: mem.t, Redactor: rd,
	})
	if berr != nil {
		t.Fatal(berr)
	}
	r := &runner{p: p, cfg: defaultSettings(), explicit: true, tel: mem.t, sample: true}
	return r.Review(context.Background(), Request{Target: discountFixture(t)})
}

// noSecretInAnySpan fails when the secret is in a span's name, status, attributes, events or event attributes.
// It returns each non-empty status as "span: description", so a caller can show the secret had a way in.
func noSecretInAnySpan(t *testing.T, mem *memTelemetry, secret string) (statuses []string) {
	t.Helper()
	for _, s := range mem.spans.GetSpans() {
		if strings.Contains(s.Name, secret) {
			t.Fatalf("span name %q holds the secret", s.Name)
		}
		if s.Status.Description != "" {
			statuses = append(statuses, s.Name+": "+s.Status.Description)
		}
		if strings.Contains(s.Status.Description, secret) {
			t.Fatalf("span %q status description holds the secret: %q", s.Name, s.Status.Description)
		}
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
	return statuses
}

// hasStatus reports whether a status "span: description" starts with span and holds every part.
func hasStatus(statuses []string, span string, parts ...string) bool {
	for _, s := range statuses {
		if !strings.HasPrefix(s, span+": ") {
			continue
		}
		all := true
		for _, p := range parts {
			all = all && strings.Contains(s, p)
		}
		if all {
			return true
		}
	}
	return false
}

// TestTelemetryDoesNotChangeWhatAReviewDecides: the same sample review runs with the exporter off and on, and
// the findings, the proof counts, the verdicts and the failed services are the same. Only where the trace file
// sits differs, since each run has its own log directory. It fails if a span, a metric or a flush ever feeds
// back into a decision.
func TestTelemetryDoesNotChangeWhatAReviewDecides(t *testing.T) {
	c := newCollector(t)
	run := func(env map[string]string) Result {
		t.Helper()
		rev, err := Open(Options{Sample: true, LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
		if err != nil {
			t.Fatal(err)
		}
		res, err := rev.Review(context.Background(), Request{Target: discountFixture(t)})
		if err != nil {
			t.Fatal(err)
		}
		res.TracePath = ""
		return res
	}
	off := run(nil)
	on := run(map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": c.srv.URL})
	if c.traces.Load() == 0 {
		t.Fatal("control: the exporter was not on, so the two runs do not differ in telemetry")
	}
	if len(off.Findings) == 0 {
		t.Fatal("control: the sample review kept no finding, so there is nothing to compare")
	}
	if !reflect.DeepEqual(off, on) {
		t.Fatalf("telemetry changed the result:\noff: %+v\non:  %+v", off, on)
	}
}

// TestAFailedReviewKeepsTheSecretOutOfTheExportedStatus: the planner fails with an error that echoes a
// configured secret, and the review fails with it. The root span and the agent span are Error, and neither
// status description holds the secret; the redactor's marker stands in its place, which shows the secret did
// reach the status path. It fails if either status is set from the raw error text.
func TestAFailedReviewKeepsTheSecretOutOfTheExportedStatus(t *testing.T) {
	const secret = "sk-not-for-the-status-4d1e"
	mem := newMemTelemetry(t)
	rd := redact.New(map[string]string{"TEST_SECRET_KEY": secret})
	_, err := failedReview(t, mem, rd, errors.New("adapter refused the key "+secret))
	if err == nil {
		t.Fatal("control: the review did not fail, so the failed path was not exercised")
	}
	statuses := noSecretInAnySpan(t, mem, secret)
	marker := rd.Text(secret)
	for _, name := range []string{"proven.review", "agent planner"} {
		if !hasStatus(statuses, name, "adapter refused the key ", marker) {
			t.Fatalf("span %q has no status with the redaction marker %q; statuses: %v", name, marker, statuses)
		}
	}
}

// TestAFailedServiceKeepsTheSecretOutOfTheAgentStatus: the planner's service is down and its reason echoes the
// secret. The review finishes degraded, and the planner span's status holds the marker, not the secret.
func TestAFailedServiceKeepsTheSecretOutOfTheAgentStatus(t *testing.T) {
	const secret = "sk-not-for-the-service-77ab"
	mem := newMemTelemetry(t)
	rd := redact.New(map[string]string{"TEST_SECRET_KEY": secret})
	res, err := failedReview(t, mem, rd, &service.Down{Service: "nightshift", Reason: "401 for " + secret})
	if err != nil {
		t.Fatalf("a down service must degrade the review, not fail it: %v", err)
	}
	if len(res.FailedServices) == 0 {
		t.Fatal("control: no failed service, so the failed path was not exercised")
	}
	statuses := noSecretInAnySpan(t, mem, secret)
	if marker := rd.Text(secret); !hasStatus(statuses, "agent planner", marker) {
		t.Fatalf("the planner's status has no redaction marker %q; statuses: %v", marker, statuses)
	}
}
