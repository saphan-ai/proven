package answer

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
)

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
