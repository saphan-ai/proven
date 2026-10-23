package cli

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestTextCarriesOneRuleLinePerFinding proves the text output carries one
// "  rule <rule-id>" line for each finding whose rule id differs from the
// name in its header. When the two are the same, the line is absent, so the
// id is not printed twice. An empty id is shown as the rule id in the header,
// and the line is then absent for the same reason.
func TestTextCarriesOneRuleLinePerFinding(t *testing.T) {
	result := review.Result{
		Class:     "backend",
		ClassRule: "fallback",
		Findings: []review.Finding{
			{
				ID: "F1", RuleID: "discount.twice", Severity: review.SeverityError,
				Title: "twice", Message: "applied twice",
				Location: review.Location{Path: "billing/discount.go", StartLine: 10, EndLine: 12},
			},
			{
				ID: "F2", RuleID: "leftpad.kev", Severity: review.SeverityError,
				Title: "kev", Message: "known exploited",
				Location: review.Location{Path: "go.mod", StartLine: 5, EndLine: 5},
			},
			{
				ID: "same.rule", RuleID: "same.rule", Severity: review.SeverityWarning,
				Title: "same", Message: "same id",
				Location: review.Location{Path: "a.go", StartLine: 1, EndLine: 1},
			},
			{
				ID: "", RuleID: "header.only", Severity: review.SeverityInfo,
				Title: "fallback", Message: "id empty",
				Location: review.Location{Path: "b.go", StartLine: 2, EndLine: 2},
			},
		},
	}
	body, err := render("text", result)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, line := range []string{"  rule discount.twice\n", "  rule leftpad.kev\n"} {
		if strings.Count(text, line) != 1 {
			t.Fatalf("%q count in\n%s", line, text)
		}
	}
	if strings.Count(text, "\n  rule ") != 2 {
		t.Fatalf("rule lines\n%s", text)
	}
	if strings.Contains(text, "  rule same.rule") || strings.Contains(text, "  rule header.only") {
		t.Fatalf("repeated id\n%s", text)
	}
	if !strings.Contains(text, "finding header.only INFO fallback\n") {
		t.Fatalf("empty id was not shown as the rule id\n%s", text)
	}
}
