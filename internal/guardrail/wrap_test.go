package guardrail_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/service"
)

type scriptGen struct {
	text  string
	calls int
}

func (g *scriptGen) Complete(ctx context.Context, agent, prompt string) (string, error) {
	g.calls++
	return g.text, nil
}

type scriptRail struct {
	rows []guardrail.Verdict
	i    int
}

func (s *scriptRail) Check(ctx context.Context, prompt, answer string) (guardrail.Verdict, error) {
	v := s.rows[s.i]
	s.i++
	return v, nil
}

func TestWrapStopsBeforeTheModelWhenThePromptIsUnsafe(t *testing.T) {
	gen := &scriptGen{text: "must-not-leak"}
	rail := &scriptRail{rows: []guardrail.Verdict{{
		User: "unsafe", Categories: []string{"Criminal Planning/Confessions", "Other"},
	}}}
	var lines []string
	w, err := guardrail.Wrap(gen, rail, func(agent, side string, v guardrail.Verdict, refused string) error {
		lines = append(lines, side+":"+v.User+":"+refused)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	text, err := w.Complete(context.Background(), "planner", "bad prompt")
	if text != "" {
		t.Fatalf("text %q", text)
	}
	d, ok := service.As(err)
	if !ok || d.Service != guardrail.Service || d.Reason != "unsafe: Criminal Planning/Confessions, Other" {
		t.Fatalf("%v", err)
	}
	if gen.calls != 0 {
		t.Fatalf("model calls %d", gen.calls)
	}
	if len(lines) != 1 || lines[0] != "prompt:unsafe:" {
		t.Fatalf("observed %#v", lines)
	}
}

func TestWrapDoesNotReturnAnUnsafeAnswer(t *testing.T) {
	gen := &scriptGen{text: "must-not-leak"}
	rail := &scriptRail{rows: []guardrail.Verdict{
		{User: "safe"},
		{User: "safe", Response: "unsafe", Categories: []string{"Malware", "PII/Privacy"}},
	}}
	w, err := guardrail.Wrap(gen, rail, nil)
	if err != nil {
		t.Fatal(err)
	}
	text, err := w.Complete(context.Background(), "planner", "prompt")
	if text != "" {
		t.Fatalf("text %q", text)
	}
	d, ok := service.As(err)
	if !ok || d.Reason != "unsafe: Malware, PII/Privacy" {
		t.Fatalf("%v", err)
	}
	if gen.calls != 1 {
		t.Fatalf("model calls %d", gen.calls)
	}
}

func TestRecordedFixtureVerdictStopsTheCall(t *testing.T) {
	rec := guardrail.RecordedUnsafePrompt()
	rec.Prompt = "one fixture prompt"
	sample, err := guardrail.NewSample([]guardrail.Record{rec})
	if err != nil {
		t.Fatal(err)
	}
	gen := &scriptGen{text: "must-not-leak"}
	w, err := guardrail.Wrap(gen, sample, nil)
	if err != nil {
		t.Fatal(err)
	}
	text, err := w.Complete(context.Background(), "planner", rec.Prompt)
	if text != "" || gen.calls != 0 {
		t.Fatalf("text %q calls %d err %v", text, gen.calls, err)
	}
	d, ok := service.As(err)
	if !ok || d.Reason != "unsafe: Criminal Planning/Confessions, Other" {
		t.Fatalf("%v", err)
	}
}

func TestSafeSampleAnswersSafe(t *testing.T) {
	v, err := guardrail.Safe{}.Check(context.Background(), "prompt", "answer")
	if err != nil || v.Unsafe() || v.User != "safe" || v.Response != "safe" {
		t.Fatalf("%#v %v", v, err)
	}
}
