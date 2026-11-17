// Command scope renders a recorded teaching scenario through the real scope
// validator and report writers. No live model, network or project test runs.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	agent "github.com/saphan-ai/proven/internal/agent/delivery"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/cli"
	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/review"
)

type recorded struct{}

func (recorded) Complete(_ context.Context, _ string, prompt string) (string, error) {
	var input struct {
		Source delivery.Source `json:"source"`
	}
	parts := strings.SplitN(prompt, "DATA:\n", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("missing source")
	}
	if err := json.Unmarshal([]byte(parts[1]), &input); err != nil {
		return "", err
	}
	c := delivery.Claim{SourceKind: input.Source.Kind, SourceIndex: input.Source.Index, SourceQuote: input.Source.Text, Requirement: input.Source.Text, Status: "unclear", Reason: "The bounded code does not establish an empty-input contract; no acceptance test ran."}
	switch input.Source.Kind {
	case "description":
		c.Status = "code_indication"
		c.Reason = "A negative-amount guard is present. This assessment did not execute it."
		c.Citations = []delivery.Citation{{Path: "payment.go", StartLine: 3, EndLine: 5, Quote: "    if amount < 0 {\n        return 0\n    }"}}
	case "commit":
		c.Status = "contradiction"
		c.Reason = "The code records the key but does not check for a previous request before returning the charge amount."
		c.Citations = []delivery.Citation{{Path: "payment.go", StartLine: 6, EndLine: 7, Quote: "    seen[key] = true\n    return amount"}}
	}
	b, err := json.Marshal(struct {
		Claims []delivery.Claim `json:"claims"`
	}{[]delivery.Claim{c}})
	return string(b), err
}

type replay struct{ result review.Result }

func (r replay) Review(context.Context, review.Request) (review.Result, error) { return r.result, nil }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: go run ./examples/scope OUTPUT_DIRECTORY")
		return 2
	}
	if err := os.MkdirAll(args[0], 0755); err != nil {
		return 2
	}
	m := delivery.Metadata{URL: "https://github.com/example/project/pull/42", Description: "Reject negative amounts before charging.", Commits: []delivery.Commit{{SHA: "abcdef0123456789abcdef0123456789abcdef01", Message: "Keep repeated requests idempotent.", URL: "https://github.com/example/project/commit/abcdef0123456789abcdef0123456789abcdef01"}}, Comments: []delivery.Comment{{Kind: "pr_comment", ID: 42, Author: "example-reviewer", Body: "Handle empty amounts without throwing.", URL: "https://github.com/example/project/pull/42#issuecomment-42"}, {Kind: "pr_comment", ID: 43, Author: "example-author", Body: "fix", URL: "https://github.com/example/project/pull/42#issuecomment-43"}}}
	code := "package payment\nfunc Charge(amount int, seen map[string]bool, key string) int {\n    if amount < 0 {\n        return 0\n    }\n    seen[key] = true\n    return amount\n}\n"
	d, err := agent.Assess(context.Background(), recorded{}, m, []change.File{{Path: "payment.go", Change: []byte(code)}}, "recorded-teaching-scenario", nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	r := review.Result{Delivery: &d, Class: "recorded example", ClassRule: "recorded scope scenario", ReviewerStatements: []string{"Recorded teaching scenario: synthetic declarations and code, recorded model answers; no live model or acceptance tests."}}
	base := filepath.Join(args[0], "scope-review")
	return cli.Run([]string{"scan", "--no-error", "--html-output", base + ".html", "--json-output", base + ".json", "--md-output", base + ".md"}, func(string) string { return "" }, stdout, stderr, func(review.Options) (review.Reviewer, error) { return replay{r}, nil })
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
