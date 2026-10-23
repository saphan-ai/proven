// Command livecheck names each live service a review needs, before a demo.
//
// It checks the generator (Meterless Nightshift), the guardrail, Zetaris and
// the Prover's container runtime. It reads the same settings a live review
// loads, sends one small call to each service, and prints one line for each.
// A secret value is never printed.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/livecheck"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/testrun"
)

const (
	// Exit codes, as the help text lists them.
	exitReady    = 0
	exitNotReady = 1
	exitUsage    = 2
)

// helpText is printed by -h and --help, and after a refused argument.
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

// check is one service's printed line and whether it is ready.
type check struct {
	line  string
	ready bool
}

// run checks every service and returns the exit code. The environment is
// passed in, so a test sets it without touching the process.
func run(args []string, stdout, stderr io.Writer, env func(string) string) int {
	if env == nil {
		env = func(string) string { return "" }
	}
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(stdout, helpText)
		return exitReady
	}
	// The configured values are read once, so every printed line, the summary
	// and a usage refusal included, is redacted by the one shared redactor.
	rd := redact.FromEnv(env)
	if len(args) > 0 {
		fmt.Fprintf(stderr, "unknown argument %s\n%s", rd.Text(args[0]), helpText)
		return exitUsage
	}
	// Every service is checked even after one fails, so one run names all that
	// are missing.
	checks := []check{
		checkGenerator(env),
		checkGuardrail(env),
		checkQuery(env),
		checkContainer(env),
	}
	readyN, notN := 0, 0
	for _, c := range checks {
		fmt.Fprintln(stdout, rd.Text(c.line))
		if c.ready {
			readyN++
		} else {
			notN++
		}
	}
	fmt.Fprintln(stdout, rd.Text(fmt.Sprintf("summary: %d ready, %d not ready", readyN, notN)))
	if notN > 0 {
		return exitNotReady
	}
	return exitReady
}

// checkGenerator sends one short prompt and names the routed model and the
// receipt id the answer carried.
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

// checkGuardrail checks one short prompt. Mode off is a setting, not a
// failure, so it is ready with no call.
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

// checkQuery runs one SELECT on the evidence view and names the row count.
// A view name that is not a plain identifier is refused before anything is
// sent, because it is pasted into the SQL.
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

// checkContainer asks the runtime for its version and inspects the image. It
// never pulls: a missing image is not ready. Probe bounds itself with the
// container timeout.
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

// readyLine is a ready line with an optional detail.
func readyLine(name, detail string) check {
	if detail == "" {
		return check{line: name + ": ready", ready: true}
	}
	return check{line: name + ": ready: " + detail, ready: true}
}

// notReady is a not-ready line with its reason.
func notReady(name, reason string) check {
	return check{line: name + ": not ready: " + reason, ready: false}
}

// evidenceView reads the evidence view setting. Empty means the default view.
func evidenceView(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = livecheck.DefaultEvidenceView
	}
	// The view is pasted into SQL. epss.ValidView is the one check for that
	// shape, so this command cannot grow a second length limit of its own.
	if !epss.ValidView(raw) {
		return "", fmt.Errorf("configuration refused: %s=%q: not an evidence view; allowed: a view or schema.view of letters, digits and underscores; empty means %s", livecheck.EnvEvidenceView, clip(raw), livecheck.DefaultEvidenceView)
	}
	return raw, nil
}

// clip shortens a refused value before it is printed.
func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
