// Command livecheck names each live service a review needs, before a demo.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/livecheck"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
)

const (
	exitReady    = 0
	exitNotReady = 1
	exitUsage    = 2
)

const helpText = `livecheck checks every live service a review needs.

One line per service, then a summary.
A line is "ready" or "not ready: <reason>".
The generator line names the routed model and the receipt id, or "none".
The guardrail line names the verdict and the latency.
Mode off is ready and prints "off by setting".
The query line names the row count of one SELECT over the evidence view.
The container line checks that the runtime answers and the image is present.
It does not pull.

Exit codes:
  0  every service is ready
  1  at least one service is not ready
  2  the arguments were refused

Settings are the ones a live review loads, from the environment.
Each check is bounded by that service's own timeout.
` + livecheck.EnvEvidenceView + ` names the evidence view. Empty means ` + livecheck.DefaultEvidenceView + `.
A secret value is never printed.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

type check struct {
	line  string
	ready bool
}

func run(args []string, stdout, stderr io.Writer, env func(string) string) int {
	if env == nil {
		env = func(string) string { return "" }
	}
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(stdout, helpText)
		return exitReady
	}
	if len(args) > 0 {
		fmt.Fprintf(stderr, "unknown argument %s\n%s", args[0], helpText)
		return exitUsage
	}
	secrets := secretValues(env)
	checks := []check{
		checkGenerator(env),
		checkGuardrail(env),
		checkQuery(env),
		checkContainer(env),
	}
	readyN, notN := 0, 0
	for _, c := range checks {
		fmt.Fprintln(stdout, redact(c.line, secrets))
		if c.ready {
			readyN++
		} else {
			notN++
		}
	}
	fmt.Fprintln(stdout, redact(fmt.Sprintf("summary: %d ready, %d not ready", readyN, notN), secrets))
	if notN > 0 {
		return exitNotReady
	}
	return exitReady
}

func checkGenerator(env func(string) string) check {
	cfg, err := generator.Load(env)
	if err != nil {
		return notReady("generator", err.Error())
	}
	live := generator.NewLive(cfg, nil)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if _, err := live.Complete(ctx, "planner", "hello"); err != nil {
		return notReady("generator", err.Error())
	}
	model, receipt := "unknown", "none"
	if use, ok := live.LastUse(); ok {
		if use.Model != "" {
			model = use.Model
		}
		if use.ReceiptID != "" {
			receipt = use.ReceiptID
		}
	}
	return readyLine("generator", fmt.Sprintf("model %s, receipt %s", model, receipt))
}

func checkGuardrail(env func(string) string) check {
	cfg, err := guardrail.Load(env)
	if err != nil {
		return notReady("guardrail", err.Error())
	}
	if cfg.Mode == guardrail.ModeOff {
		return readyLine("guardrail", "off by setting")
	}
	rail, err := guardrail.New(cfg, nil)
	if err != nil {
		return notReady("guardrail", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	v, err := rail.Check(ctx, "hello", "")
	if err != nil {
		return notReady("guardrail", err.Error())
	}
	return readyLine("guardrail", fmt.Sprintf("verdict %s, latency %s", v.User, v.Latency.Round(time.Millisecond)))
}

func checkQuery(env func(string) string) check {
	cfg, err := query.Load(env)
	if err != nil {
		return notReady("query", err.Error())
	}
	view, err := evidenceView(env(livecheck.EnvEvidenceView))
	if err != nil {
		return notReady("query", err.Error())
	}
	live := query.NewLive(cfg, nil)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	res, err := live.Query(ctx, "SELECT 1 FROM "+view+" LIMIT 1")
	if err != nil {
		return notReady("query", err.Error())
	}
	return readyLine("query", fmt.Sprintf("rows %d", len(res.Rows)))
}

func checkContainer(env func(string) string) check {
	cfg, err := testrun.Load(env)
	if err != nil {
		return notReady("container", err.Error())
	}
	box, err := testrun.NewContainer(cfg, nil)
	if err != nil {
		return notReady("container", err.Error())
	}
	if err := box.Probe(context.Background()); err != nil {
		return notReady("container", err.Error())
	}
	return readyLine("container", "")
}

func readyLine(name, detail string) check {
	if detail == "" {
		return check{line: name + ": ready", ready: true}
	}
	return check{line: name + ": ready: " + detail, ready: true}
}

func notReady(name, reason string) check {
	return check{line: name + ": not ready: " + reason, ready: false}
}

func evidenceView(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = livecheck.DefaultEvidenceView
	}
	if !validView(raw) {
		return "", fmt.Errorf("configuration refused: %s=%q: not an evidence view; allowed: a schema.table of letters, digits and underscores, at most 128 characters", livecheck.EnvEvidenceView, clip(raw))
	}
	return raw, nil
}

func validView(s string) bool {
	if len(s) > 128 {
		return false
	}
	parts := strings.Split(s, ".")
	if len(parts) != 1 && len(parts) != 2 {
		return false
	}
	for _, p := range parts {
		if !viewPart(p) {
			return false
		}
	}
	return true
}

func viewPart(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}

func secretValues(env func(string) string) []string {
	seen := map[string]bool{}
	var out []string
	for _, key := range []string{generator.EnvToken, guardrail.EnvKey, query.EnvAPIKey} {
		v := env(key)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}

// redact strips secret values before a line is printed. A routed model, a
// receipt id, or a refusal can echo one.
func redact(s string, secrets []string) string {
	for _, sec := range secrets {
		if sec == "" {
			continue
		}
		s = strings.ReplaceAll(s, sec, "redacted")
	}
	return s
}
