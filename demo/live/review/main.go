// Command reviewlive runs one live review of a directory with base/ and change/.
// proven scan on this branch accepts only --sample, so the demo calls review.Open.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

func run(args []string, stdout, stderr io.Writer, env func(string) string) int {
	if env == nil {
		env = func(string) string { return "" }
	}
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: reviewlive <demo-dir>")
		return 2
	}
	target, err := filepath.Abs(args[0])
	if err != nil {
		fmt.Fprintln(stderr, redact(err.Error(), env))
		return 2
	}
	logDir := filepath.Join(provenRoot(target), "logs")
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: logDir,
		Env:    env,
	})
	if err != nil {
		return fail(stderr, env, err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: target})
	if err != nil {
		return fail(stderr, env, err)
	}
	var buf bytes.Buffer
	printResult(&buf, res)
	fmt.Fprint(stdout, redact(buf.String(), env))
	return 0
}

func fail(stderr io.Writer, env func(string) string, err error) int {
	fmt.Fprintln(stderr, redact(err.Error(), env))
	var cfg *review.ConfigError
	if errors.As(err, &cfg) {
		return 7
	}
	return 2
}

func provenRoot(target string) string {
	dir := target
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return target
		}
		dir = parent
	}
}

func printResult(w io.Writer, res review.Result) {
	for _, f := range res.Findings {
		fmt.Fprintf(w, "finding %s %s %s\n", f.ID, f.Severity, f.Title)
		if f.Unresolved {
			fmt.Fprintln(w, "  unresolved")
		}
		fmt.Fprintf(w, "  %s:%d-%d\n", f.Location.Path, f.Location.StartLine, f.Location.EndLine)
		fmt.Fprintf(w, "  %s\n", f.Message)
		for _, e := range f.Evidence {
			switch e.Kind {
			case review.EvidenceTest:
				fmt.Fprintf(w, "  evidence test %s %s fails_on_change=%t passes_on_base=%t\n", e.Test.Path, e.Test.Name, e.Test.FailsOnChange, e.Test.PassesOnBase)
				fmt.Fprintf(w, "  command: %s\n", e.Test.Command)
				fmt.Fprintf(w, "  output: %s\n", strings.TrimSpace(e.Test.Output))
			case review.EvidenceQuery:
				fmt.Fprintf(w, "  evidence query %s snapshot=%s\n", e.Query.Engine, e.Query.Snapshot)
				fmt.Fprintf(w, "  SQL: %s\n", e.Query.SQL)
				fmt.Fprintf(w, "  columns: %s\n", strings.Join(e.Query.Columns, ", "))
				for _, row := range e.Query.Rows {
					fmt.Fprintf(w, "  row: %s\n", strings.Join(row, " | "))
				}
			}
		}
	}
	if len(res.FailedServices) == 0 {
		fmt.Fprintln(w, "failed services: none")
	} else {
		for _, s := range res.FailedServices {
			fmt.Fprintf(w, "failed service: %s: %s\n", s.Service, s.Reason)
		}
	}
	if sentence := familySentence(res.TracePath); sentence != "" {
		fmt.Fprintln(w, sentence)
	}
	if res.TracePath != "" {
		fmt.Fprintf(w, "trace: %s\n", res.TracePath)
	}
}

func familySentence(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Kind == "family" && e.Text != "" {
			return e.Text
		}
	}
	return ""
}

func redact(s string, env func(string) string) string {
	var secrets []string
	seen := map[string]bool{}
	for _, key := range []string{"METERLESS_RUNTIME_TOKEN", "PROVEN_GUARDRAIL_KEY", "ZETARIS_API_KEY"} {
		v := env(key)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		secrets = append(secrets, v)
	}
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	for _, sec := range secrets {
		s = strings.ReplaceAll(s, sec, "redacted")
	}
	return s
}
