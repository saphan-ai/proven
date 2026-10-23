package answer

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
)

// The live prompt carries the task, the shape, the example and the data, and asks for JSON only.
func TestInstructAsksForTheJSONOnly(t *testing.T) {
	got := Instruct(Spec{
		Task:    "Name the risks.",
		Shape:   "a JSON array of objects.",
		Example: `[]`,
	}, "FILE a.go\n")
	for _, want := range []string{
		"Task: Name the risks.",
		"Reply with JSON only.",
		"The JSON is a JSON array of objects.",
		"Example:\n[]",
		"FILE a.go\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q\n%s", want, got)
		}
	}
}

// Decode finds the JSON in a fence, in prose, and with a trailing comma.
func TestDecodeTakesJSONFromAFenceFromProseAndFromATrailingComma(t *testing.T) {
	type row struct {
		SQL string `json:"sql"`
	}
	fence := "The query follows.\n```json\n{\"sql\":\"SELECT 1\"}\n```\nThat is all.\n"
	prose := "Use this: {\"sql\":\"SELECT 1\"} and nothing else."
	comma := "{\"sql\":\"SELECT 1\",}"
	commaFence := "Here:\n```json\n{\"sql\":\"SELECT 1\",}\n```\nThanks.\n"
	nested := "{\n  \"sql\": \"SELECT 1\",\n}\n"
	for _, raw := range []string{fence, prose, comma, commaFence, nested} {
		var got row
		if err := Decode(raw, &got); err != nil {
			t.Fatalf("%q: %v", raw, err)
		}
		if got.SQL != "SELECT 1" {
			t.Fatalf("%q decoded %#v", raw, got)
		}
	}
}

// A comma before a bracket inside a string is part of the value and is kept.
func TestDecodeKeepsACommaInsideAString(t *testing.T) {
	type row struct {
		SQL string `json:"sql"`
	}
	var got row
	if err := Decode("{\"sql\":\"SELECT 1,\"}", &got); err != nil {
		t.Fatal(err)
	}
	if got.SQL != "SELECT 1," {
		t.Fatalf("%q", got.SQL)
	}
}

// When a reply has a fence, JSON in the prose around it is ignored.
func TestDecodePrefersTheFencedJSON(t *testing.T) {
	type row struct {
		SQL string `json:"sql"`
	}
	raw := "Ignore {\"sql\":\"SELECT 9\"}.\n```json\n{\"sql\":\"SELECT 1\"}\n```\n"
	var got row
	if err := Decode(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.SQL != "SELECT 1" {
		t.Fatalf("%q", got.SQL)
	}
}

// A numbered list is not read as a number; the decode error is returned.
func TestDecodeNumberedProseIsNotJSON(t *testing.T) {
	var got map[string]any
	err := Decode("1. this is not json", &got)
	if err == nil {
		t.Fatal("prose decoded")
	}
	if !strings.Contains(err.Error(), "after decimal point") {
		t.Fatalf("%v", err)
	}
}

// A reply that stays unreadable after one repair degrades by the agent's name,
// and the repair prompt carries the previous answer.
func TestReadRepairsOnceThenNamesTheAgent(t *testing.T) {
	g := &script{replies: []string{"1. this is not json", "still prose"}}
	var got map[string]any
	err := Read(context.Background(), g, "planner", "the prompt", Spec{Task: "Name the risks."}, &got)
	d, ok := service.As(err)
	if !ok || d.Service != "planner" || d.Reason != FailReason {
		t.Fatalf("%v", err)
	}
	if len(g.prompts) != 2 {
		t.Fatalf("calls %d", len(g.prompts))
	}
	if !strings.Contains(g.prompts[1], "Reply with JSON only") || !strings.Contains(g.prompts[1], "1. this is not json") {
		t.Fatalf("repair %q", g.prompts[1])
	}
}

// A repaired answer is used, and no field from the failed first decode survives.
func TestReadUsesARepairedAnswer(t *testing.T) {
	type row struct {
		SQL string `json:"sql"`
		N   int    `json:"n"`
	}
	g := &script{replies: []string{`{"sql":"KEEP","n":"bad"}`, `{"n":2}`}}
	var got row
	if err := Read(context.Background(), g, "investigator", "prompt", Spec{Task: "Query."}, &got); err != nil {
		t.Fatal(err)
	}
	if got.SQL != "" || got.N != 2 {
		t.Fatalf("%#v", got)
	}
	if len(g.prompts) != 2 {
		t.Fatalf("calls %d", len(g.prompts))
	}
}

// A reply that decodes is used at once; no repair call is made.
func TestReadDoesNotRepairADecodedAnswer(t *testing.T) {
	g := &script{replies: []string{"Here:\n```json\n{\"sql\":\"SELECT 1\",}\n```\n", "unused"}}
	var got struct {
		SQL string `json:"sql"`
	}
	if err := Read(context.Background(), g, "investigator", "prompt", Spec{}, &got); err != nil {
		t.Fatal(err)
	}
	if got.SQL != "SELECT 1" || len(g.prompts) != 1 {
		t.Fatalf("%#v calls %d", got, len(g.prompts))
	}
}

// A generator failure is returned as it is, with no repair call.
func TestReadReturnsAGeneratorFailureUnchanged(t *testing.T) {
	g := &script{err: &service.Down{Service: "nightshift", Reason: "timeout"}}
	var got map[string]any
	err := Read(context.Background(), g, "planner", "prompt", Spec{}, &got)
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "timeout" {
		t.Fatalf("%v", err)
	}
	if len(g.prompts) != 1 {
		t.Fatalf("calls %d", len(g.prompts))
	}
}

// script is a generator that returns its replies in order and records each prompt.
type script struct {
	replies []string
	prompts []string
	err     error
}

func (s *script) Complete(ctx context.Context, agent, prompt string) (string, error) {
	s.prompts = append(s.prompts, prompt)
	if s.err != nil {
		return "", s.err
	}
	if len(s.replies) == 0 {
		return "", context.Canceled
	}
	text := s.replies[0]
	s.replies = s.replies[1:]
	return text, nil
}
