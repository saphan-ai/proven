package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
)

type callCount struct {
	inner generator.Generator
	calls int
	last  string
}

func (g *callCount) Complete(ctx context.Context, agent, prompt string) (string, error) {
	g.calls++
	text, err := g.inner.Complete(ctx, agent, prompt)
	g.last = text
	return text, err
}

func TestLivePipelineWithoutAGuardrailIsRefused(t *testing.T) {
	_, err := Build(Options{LogDir: t.TempDir()})
	if err == nil {
		t.Fatal("a live pipeline with no guardrail was accepted")
	}
}

func TestUnsafePromptStopsTheStepAndNamesCategories(t *testing.T) {
	fx := loadFixture(t)
	prompt := fx.Answers[0].Prompt
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		body := readBody(t, r)
		if !bytes.Contains(body, []byte(`"temperature":0`)) || !bytes.Contains(body, []byte(`"request_categories":"/categories"`)) {
			t.Errorf("body missing the measured fields")
		}
		msgs := messagesOf(t, body)
		text := "User Safety: safe\nResponse Safety: safe\n"
		if len(msgs) == 1 && msgs[0] == prompt {
			text = "User Safety: unsafe\nSafety Categories: Criminal Planning/Confessions, Other\n"
		}
		_, _ = io.WriteString(w, verdictJSON(text))
	}))
	defer srv.Close()
	d, gen := reviewWithRail(t, fx, railAt(t, srv.URL), false)
	if gen.calls != 0 {
		t.Fatalf("the model ran %d times", gen.calls)
	}
	if posts != 1 {
		t.Fatalf("guardrail posts %d", posts)
	}
	if id := modelFinding(d); id != "" {
		t.Fatalf("the model produced %s", id)
	}
	assertGuardrail(t, d, "unsafe: Criminal Planning/Confessions, Other", "Criminal Planning/Confessions", "Other")
}

func TestUnsafeAnswerIsNotUsed(t *testing.T) {
	fx := loadFixture(t)
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		text := "User Safety: safe\n"
		if messageCount(t, readBody(t, r)) >= 2 {
			text = "User Safety: safe\nResponse Safety: unsafe\nSafety Categories: Malware, PII/Privacy\n"
		}
		_, _ = io.WriteString(w, verdictJSON(text))
	}))
	defer srv.Close()
	d, gen := reviewWithRail(t, fx, railAt(t, srv.URL), false)
	if gen.calls != 1 || gen.last == "" {
		t.Fatalf("calls %d last %q", gen.calls, gen.last)
	}
	if posts != 2 {
		t.Fatalf("guardrail posts %d", posts)
	}
	if id := modelFinding(d); id != "" {
		t.Fatalf("the answer was used: %s", id)
	}
	raw := traceRaw(t, d.TracePath)
	if bytes.Contains(raw, []byte(gen.last)) {
		t.Fatal("the trace stored the answer")
	}
	assertGuardrail(t, d, "unsafe: Malware, PII/Privacy", "Malware", "PII/Privacy")
	if !bytes.Contains(raw, []byte("answer unsafe")) {
		t.Fatalf("trace %s", raw)
	}
}

// Mutated field: PROVEN_GUARDRAIL_MODE, set to off.
// The same server marks the planner prompt unsafe. With the guardrail off
// the review completes, and the trace says no guardrail.
func TestGuardrailOffCompletesAndSaysNoGuardrail(t *testing.T) {
	fx := loadFixture(t)
	prompt := fx.Answers[0].Prompt
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		msgs := messagesOf(t, readBody(t, r))
		text := "User Safety: safe\nResponse Safety: safe\n"
		if len(msgs) > 0 && msgs[0] == prompt {
			text = "User Safety: unsafe\nSafety Categories: Criminal Planning/Confessions, Other\n"
		}
		_, _ = io.WriteString(w, verdictJSON(text))
	}))
	defer srv.Close()
	d, gen := reviewWithRail(t, fx, nil, true)
	if posts != 0 {
		t.Fatalf("the off switch called the guardrail %d times", posts)
	}
	if gen.calls == 0 {
		t.Fatal("the model did not run")
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("F1 missing")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 missing")
	}
	if _, ok := kept(d, "F3"); ok {
		t.Fatal("F3 was kept")
	}
	if len(d.Failed) != 0 {
		t.Fatalf("failed %#v", d.Failed)
	}
	raw := traceRaw(t, d.TracePath)
	if bytes.Count(raw, []byte("no guardrail")) != 1 {
		t.Fatalf("trace %s", raw)
	}
}

// Mutated field: the guardrail server, closed before the review.
func TestClosedGuardrailRefusesTheStepByName(t *testing.T) {
	fx := loadFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed server accepted a call")
	}))
	url := srv.URL
	srv.Close()
	d, gen := reviewWithRail(t, fx, railAt(t, url), false)
	if gen.calls != 0 {
		t.Fatalf("the model ran %d times", gen.calls)
	}
	if id := modelFinding(d); id != "" {
		t.Fatalf("the model produced %s", id)
	}
	assertGuardrail(t, d, "unreachable")
	raw := traceRaw(t, d.TracePath)
	if !bytes.Contains(raw, []byte("guardrail: unreachable")) {
		t.Fatalf("trace %s", raw)
	}
}

func TestRecordedFixturePromptStopsTheStep(t *testing.T) {
	fx := loadFixture(t)
	rec := guardrail.RecordedUnsafePrompt()
	rec.Prompt = fx.Answers[0].Prompt
	sample, err := guardrail.NewSample([]guardrail.Record{rec})
	if err != nil {
		t.Fatal(err)
	}
	d, gen := reviewWithRail(t, fx, sample, false)
	if gen.calls != 0 {
		t.Fatalf("the model ran %d times", gen.calls)
	}
	assertGuardrail(t, d, "unsafe: Criminal Planning/Confessions, Other", "Criminal Planning/Confessions")
}

func TestSafeGuardrailKeepsTheSampleFindings(t *testing.T) {
	fx := loadFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		text := "User Safety: safe\n"
		if messageCount(t, readBody(t, r)) >= 2 {
			text = "User Safety: safe\nResponse Safety: safe\n"
		}
		_, _ = io.WriteString(w, verdictJSON(text))
	}))
	defer srv.Close()
	d, gen := reviewWithRail(t, fx, railAt(t, srv.URL), false)
	if gen.calls == 0 {
		t.Fatal("the model did not run")
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("F1 missing")
	}
	if _, ok := kept(d, "F2"); !ok {
		t.Fatal("F2 missing")
	}
	if _, ok := kept(d, "F3"); ok {
		t.Fatal("F3 was kept")
	}
	if len(d.Failed) != 0 {
		t.Fatalf("failed %#v", d.Failed)
	}
}

func modelFinding(d candidate.Draft) string {
	for _, f := range d.Findings {
		if !strings.HasPrefix(f.ID, "osv:") {
			return f.ID
		}
	}
	return ""
}

func reviewWithRail(t *testing.T, fx *fixture.Fixture, rail guardrail.Adapter, off bool) (candidate.Draft, *callCount) {
	t.Helper()
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	gen := &callCount{inner: inner}
	p, err := Build(Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: gen,
		Guardrail: rail,
		GuardOff:  off,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	return d, gen
}

func loadFixture(t *testing.T) *fixture.Fixture {
	t.Helper()
	fx, err := fixture.Load(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	return fx
}

func railAt(t *testing.T, base string) guardrail.Adapter {
	t.Helper()
	s, err := guardrail.Load(func(k string) string {
		switch k {
		case guardrail.EnvBaseURL:
			return base
		case guardrail.EnvKey:
			return "test-key"
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

func assertGuardrail(t *testing.T, d candidate.Draft, reason string, categories ...string) {
	t.Helper()
	found := false
	for _, f := range d.Failed {
		if f.Service == guardrail.Service && f.Reason == reason {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v, want %s", d.Failed, reason)
	}
	raw := traceRaw(t, d.TracePath)
	for _, c := range categories {
		if !bytes.Contains(raw, []byte(c)) {
			t.Fatalf("trace missing %s\n%s", c, raw)
		}
	}
}

func traceRaw(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readBody(t *testing.T, r *http.Request) []byte {
	t.Helper()
	b, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func messageCount(t *testing.T, body []byte) int {
	t.Helper()
	return len(messagesOf(t, body))
}

func messagesOf(t *testing.T, body []byte) []string {
	t.Helper()
	var decoded struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	out := make([]string, len(decoded.Messages))
	for i, m := range decoded.Messages {
		out[i] = m.Content
	}
	return out
}

func verdictJSON(text string) string {
	raw, err := json.Marshal(map[string]any{
		"choices": []any{
			map[string]any{"message": map[string]any{"role": "assistant", "content": text}},
		},
	})
	if err != nil {
		panic(err)
	}
	return string(raw)
}
