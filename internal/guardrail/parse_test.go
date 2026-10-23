package guardrail_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/guardrail"
)

// The measured unsafe-prompt answer parses to an unsafe user verdict with its categories.
func TestParseMeasuredUnsafePrompt(t *testing.T) {
	v, err := guardrail.Parse("User Safety: unsafe\nSafety Categories: Criminal Planning/Confessions, Other\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if v.User != "unsafe" || v.Response != "" || !v.Unsafe() {
		t.Fatalf("%#v", v)
	}
	if len(v.Categories) != 2 || v.Categories[0] != "Criminal Planning/Confessions" || v.Categories[1] != "Other" {
		t.Fatalf("categories %#v", v.Categories)
	}
}

// The measured unsafe-answer output parses to a safe prompt and an unsafe response.
func TestParseMeasuredUnsafeAnswer(t *testing.T) {
	text := "User Safety: safe\nResponse Safety: unsafe\nSafety Categories: Malware, PII/Privacy\n"
	v, err := guardrail.Parse(text, true)
	if err != nil {
		t.Fatal(err)
	}
	if v.User != "safe" || v.Response != "unsafe" || !v.Unsafe() {
		t.Fatalf("%#v", v)
	}
	if len(v.Categories) != 2 || v.Categories[0] != "Malware" || v.Categories[1] != "PII/Privacy" {
		t.Fatalf("categories %#v", v.Categories)
	}
}

// Verdict words inside a thinking block are not the verdict.
func TestParseSkipsAThinkingBlock(t *testing.T) {
	text := "<think>\nUser Safety: unsafe\nSafety Categories: Other\n</think>\nUser Safety: safe\n"
	v, err := guardrail.Parse(text, false)
	if err != nil {
		t.Fatal(err)
	}
	if v.User != "safe" || v.Unsafe() || len(v.Categories) != 0 {
		t.Fatalf("%#v", v)
	}
}

// A safe verdict carries no categories.
func TestParseDropsCategoriesWhenTheVerdictIsSafe(t *testing.T) {
	v, err := guardrail.Parse("User Safety: safe\nSafety Categories: Other\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if v.Unsafe() || len(v.Categories) != 0 {
		t.Fatalf("%#v", v)
	}
}

// Only the exact measured word is a verdict.
// Mutated field: the User Safety word, set to a capital S.
func TestParseRefusesAVerdictThatIsNotTheMeasuredWord(t *testing.T) {
	_, err := guardrail.Parse("User Safety: Safe\n", false)
	if err == nil {
		t.Fatal("expected a refusal")
	}
}

// An answer check without a Response Safety line is refused.
// Mutated field: the Response Safety line, removed from an answer check.
func TestParseRefusesAnAnswerCheckWithoutResponseSafety(t *testing.T) {
	_, err := guardrail.Parse("User Safety: safe\n", true)
	if err == nil {
		t.Fatal("expected a refusal")
	}
	v, err := guardrail.Parse("User Safety: safe\n", false)
	if err != nil || v.User != "safe" || v.Unsafe() {
		t.Fatalf("prompt-only %#v %v", v, err)
	}
}

// An unsafe verdict stays unsafe when no categories are named.
func TestParseKeepsAnUnsafeVerdictWhenCategoriesAreAbsent(t *testing.T) {
	v, err := guardrail.Parse("User Safety: unsafe\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Unsafe() || len(v.Categories) != 0 {
		t.Fatalf("%#v", v)
	}
	if strings.Contains(errString(err), "secret") {
		t.Fatal(err)
	}
}

// errString is err's text, or empty for nil.
func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
