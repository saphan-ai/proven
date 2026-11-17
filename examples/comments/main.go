// Command comments renders real Go AST checks and a labelled semantic replay.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agent "github.com/saphan-ai/proven/internal/agent/codecomments"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/cli"
	"github.com/saphan-ai/proven/internal/codecomments"
	"github.com/saphan-ai/proven/internal/review"
)

const code = `package payments

// Charge rejects negative amounts before charging.
// @param previous obsolete parameter name
func Charge(amount int) int {
    return amount * 2
}

func Total(amount int) int {
    return Charge(amount)
}

// TODO: document behavior
func helper() bool { return true }
`

type recorded struct{}

func (recorded) Complete(_ context.Context, _ string, prompt string) (string, error) {
	parts := strings.SplitN(prompt, "DATA:\n", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("missing unit")
	}
	var unit codecomments.Unit
	if err := json.Unmarshal([]byte(parts[1]), &unit); err != nil {
		return "", err
	}
	a := codecomments.Assessment{UnitID: unit.ID, Status: "unclear", Reason: "A placeholder does not state assessable behavior; the structural warning remains separate."}
	if unit.Symbol == "Charge" {
		a.Status = "needs_attention"
		a.Reason = "The comment promises rejection of negative amounts, but the supplied function doubles and returns the amount without a guard. This is a recorded assessment; no behavior was executed."
		a.Quote = "    return amount * 2"
	}
	b, err := json.Marshal(a)
	return string(b), err
}

type reviewer struct{ result review.Result }

func (r reviewer) Review(context.Context, review.Request) (review.Result, error) {
	return r.result, nil
}
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/comments OUTPUT_DIRECTORY")
		os.Exit(2)
	}
	if err := os.MkdirAll(os.Args[1], 0755); err != nil {
		panic(err)
	}
	deterministic := codecomments.Check([]change.File{{Path: "billing/payment.go", Base: []byte("package payments\n"), Change: []byte(code)}})
	semantic := agent.Assess(context.Background(), recorded{}, deterministic)
	for _, example := range []struct {
		name, mode string
		data       codecomments.Result
	}{{"comments-deterministic", "deterministic", deterministic}, {"comments-linguistic", "recorded", semantic}} {
		r := review.Result{AnalysisMode: example.mode, Class: "documentation", ClassRule: "controlled Go comment example", CodeComments: &example.data}
		out := filepath.Join(os.Args[1], example.name)
		if code := cli.Run([]string{"scan", "--html-output", out + ".html", "--json-output", out + ".json", "--md-output", out + ".md"}, func(key string) string {
			if key == "PROVEN_LOG_DIR" {
				return filepath.Join(os.Args[1], "logs")
			}
			return ""
		}, os.Stdout, os.Stderr, func(review.Options) (review.Reviewer, error) { return reviewer{r}, nil }); code != 0 {
			os.Exit(code)
		}
	}
}
