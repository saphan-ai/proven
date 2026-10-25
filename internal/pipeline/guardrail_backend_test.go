package pipeline

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/trace"
)

// guardChecks returns the structured guardrail checks of a trace, in order.
func guardChecks(t *testing.T, path string) []trace.Check {
	t.Helper()
	var out []trace.Check
	for _, e := range readTrace(t, path) {
		if e.Kind == "guardrail" && e.Guardrail != nil {
			out = append(out, *e.Guardrail)
		}
	}
	return out
}

// railMode builds a live adapter for mode at base with key and, when model is
// not empty, that model id.
func railMode(t *testing.T, mode, base, key, model string) guardrail.Adapter {
	t.Helper()
	s, err := guardrail.Load(func(k string) string {
		switch k {
		case guardrail.EnvMode:
			return mode
		case guardrail.EnvBaseURL:
			return base
		case guardrail.EnvKey:
			return key
		case guardrail.EnvModel:
			return model
		case guardrail.EnvTimeout:
			return "1s"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	rail, err := guardrail.New(s, nil)
	if err != nil {
		t.Fatal(err)
	}
	return rail
}

// safeServer answers safe to every check.
func safeServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		text := "User Safety: safe\n"
		if messageCount(t, readBody(t, r)) >= 2 {
			text = "User Safety: safe\nResponse Safety: safe\n"
		}
		_, _ = io.WriteString(w, verdictJSON(text))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// assertBackend fails unless every check names backend and model, and the
// trace has at least one check.
func assertBackend(t *testing.T, checks []trace.Check, backend, model string) {
	t.Helper()
	if len(checks) == 0 {
		t.Fatal("the trace has no guardrail check")
	}
	for i, c := range checks {
		if c.Backend != backend || c.Model != model {
			t.Fatalf("check %d names backend %q model %q, want %q %q", i, c.Backend, c.Model, backend, model)
		}
		if c.Down != "" {
			t.Fatalf("check %d with a verdict names down %q", i, c.Down)
		}
	}
}

// Property: a check answered by our own model server names the backend
// "server" and the model id the settings configured (here the default), and
// the configured key never appears on the trace.
func TestServerCheckNamesServerAndItsModel(t *testing.T) {
	const key = "server-key-must-not-reach-the-trace"
	srv := safeServer(t)
	d, _ := reviewWithRail(t, loadFixture(t), railMode(t, "server", srv.URL, key, ""), false)
	assertBackend(t, guardChecks(t, d.TracePath), "server", "nemotron-3.5-content-safety")
	if bytes.Contains(traceRaw(t, d.TracePath), []byte(key)) {
		t.Fatal("the guardrail key reached the trace")
	}
}

// Property: a check answered by NVIDIA's hosted API names the backend
// "nvidia" and the configured model id, not the default one; the key never
// appears on the trace.
func TestNVIDIACheckNamesNVIDIAAndItsModel(t *testing.T) {
	const key = "nvidia-key-must-not-reach-the-trace"
	srv := safeServer(t)
	d, _ := reviewWithRail(t, loadFixture(t), railMode(t, "nvidia", srv.URL, key, "test/content-safety"), false)
	assertBackend(t, guardChecks(t, d.TracePath), "nvidia", "test/content-safety")
	if bytes.Contains(traceRaw(t, d.TracePath), []byte(key)) {
		t.Fatal("the guardrail key reached the trace")
	}
}

// Property: a sample-mode review with no recorded verdicts uses the sample
// guardrail, and each check names the backend "sample". The model is empty
// here, and only here and on the other non-live adapters: no model is
// configured for an adapter that calls none.
func TestSampleCheckNamesTheSampleBackend(t *testing.T) {
	d, _ := reviewWithRail(t, loadFixture(t), nil, false)
	assertBackend(t, guardChecks(t, d.TracePath), "sample", "")
}

// Property: the recorded sample guardrail names itself "sample-recorded",
// so a count can tell recorded verdicts from the always-safe sample.
func TestRecordedSampleCheckNamesItsBackend(t *testing.T) {
	fx := loadFixture(t)
	rec := guardrail.RecordedUnsafePrompt()
	rec.Prompt = fx.Answers[0].Prompt
	sample, err := guardrail.NewSample([]guardrail.Record{rec})
	if err != nil {
		t.Fatal(err)
	}
	d, _ := reviewWithRail(t, fx, sample, false)
	checks := guardChecks(t, d.TracePath)
	if len(checks) == 0 || checks[0].Backend != "sample-recorded" || checks[0].Verdict != "unsafe" {
		t.Fatalf("checks %#v", checks)
	}
}

// failingRail passes the prompt side and refuses the answer side with err
// when onAnswer is set; otherwise it refuses the prompt side.
type failingRail struct {
	err      error
	onAnswer bool
}

func (r failingRail) Check(_ context.Context, _, answer string) (guardrail.Verdict, error) {
	if r.onAnswer && answer == "" {
		return guardrail.Verdict{User: "safe"}, nil
	}
	return guardrail.Verdict{}, r.err
}

// echo is a generator that answers with a fixed text.
type echo struct{}

func (echo) Complete(context.Context, string, string) (string, error) { return "an answer", nil }

// downEvents runs one guarded call through rail and returns the trace events
// guardEvent made for it.
func downEvents(t *testing.T, rail guardrail.Adapter) []trace.Event {
	t.Helper()
	var events []trace.Event
	var gen generator.Generator = echo{}
	g, err := guardrail.Wrap(gen, rail, func(agent, side string, v guardrail.Verdict, refused string) error {
		events = append(events, guardEvent(agent, side, v, refused))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Complete(context.Background(), "planner", "a prompt"); err == nil {
		t.Fatal("a refused check let the call through")
	}
	return events
}

// Property: each reason of the closed set of down reasons lands on the trace
// as its structured value, with the side whose check could not be made, and
// the text stays "guardrail: <reason>" as before.
func TestEachDownReasonLandsAsItsStructuredValue(t *testing.T) {
	cases := []struct {
		reason string
		down   string
	}{
		{"not configured", "not-configured"},
		{guardrail.EnvKey + " is empty", "no-key"},
		{"off", "off"},
		{"bad verdict", "bad-verdict"},
		{"timeout", "timeout"},
		{"unauthorized", "unauthorized"},
		{"empty answer", "empty-answer"},
		{"unreachable", "unreachable"},
		{"no recorded verdict", "no-recorded-verdict"},
		{"http 503", "http-503"},
		{"http 429", "http-429"},
	}
	for _, tc := range cases {
		for _, side := range []string{"prompt", "answer"} {
			t.Run(tc.down+"/"+side, func(t *testing.T) {
				events := downEvents(t, failingRail{err: guardrail.Refuse(tc.reason), onAnswer: side == "answer"})
				last := events[len(events)-1]
				if last.Text != "guardrail: "+tc.reason {
					t.Fatalf("text %q", last.Text)
				}
				g := last.Guardrail
				if g == nil || g.Down != tc.down || g.Side != side || g.Verdict != "" {
					t.Fatalf("check %#v, want down %q side %q", g, tc.down, side)
				}
			})
		}
	}
}

// Property: the adapters that refuse every check name themselves and their
// down value: Absent with no reason, Absent for an empty key, and Off.
func TestRefusingAdaptersNameBackendAndDown(t *testing.T) {
	cases := []struct {
		rail    guardrail.Adapter
		backend string
		down    string
	}{
		{guardrail.Absent{}, "absent", "not-configured"},
		{guardrail.Absent{Reason: guardrail.EnvKey + " is empty"}, "absent", "no-key"},
		{guardrail.Off{}, "off", "off"},
	}
	for _, tc := range cases {
		events := downEvents(t, tc.rail)
		g := events[0].Guardrail
		if g == nil || g.Backend != tc.backend || g.Down != tc.down || g.Side != "prompt" || g.Model != "" {
			t.Fatalf("%T: check %#v", tc.rail, g)
		}
	}
}

// Property: a live check that the server refuses with 401 lands as
// "unauthorized" and still names its backend and model.
func TestLiveUnauthorizedNamesBackendAndDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	events := downEvents(t, railMode(t, "server", srv.URL, "unauthorized-key", ""))
	g := events[0].Guardrail
	if g == nil || g.Down != "unauthorized" || g.Backend != "server" || g.Model != "nemotron-3.5-content-safety" {
		t.Fatalf("check %#v", g)
	}
}

// Property: an error outside the closed set lands as the one named value
// "other"; its text stays in the event's text only, never in down. A
// guardrail refusal whose reason is not in the set is the same.
func TestUnknownErrorLandsAsOther(t *testing.T) {
	for _, err := range []error{errors.New("socket on fire"), guardrail.Refuse("socket on fire")} {
		events := downEvents(t, failingRail{err: err})
		last := events[len(events)-1]
		if last.Text != "guardrail: socket on fire" {
			t.Fatalf("text %q", last.Text)
		}
		if g := last.Guardrail; g == nil || g.Down != "other" || g.Side != "prompt" {
			t.Fatalf("check %#v", g)
		}
	}
}
