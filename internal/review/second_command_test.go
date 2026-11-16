package review

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/trace"
)

// Exercise application wiring, not just the transport constructor: two local
// text commands must both review a real Git comparison without HTTP or a key.
func TestOpenWiresTwoCommandReviewers(t *testing.T) {
	repo, base, head := twoCommits(t)
	config := filepath.Join(t.TempDir(), "proven.yaml")
	if err := os.WriteFile(config, []byte("scout_intent: off\nfixes: off\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		generator.EnvMode:          generator.ModeCommand,
		generator.EnvCommand:       `["/usr/bin/printf","%s","[]"]`,
		generator.EnvModel:         "first-cli",
		generator.EnvSecondMode:    generator.ModeCommand,
		generator.EnvSecondCommand: `["/usr/bin/printf","%s","[]"]`,
		generator.EnvSecondModel:   "second-cli",
		"PROVEN_GUARDRAIL_MODE":    "off",
		"PROVEN_MARKOVIAN":         "off",
		"PROVEN_TEST_RUNTIME":      "proven-no-container-runtime",
	}
	r, err := Open(Options{ConfigPath: config, LogDir: t.TempDir(), Env: func(k string) string { return env[k] }, testRun: quietRun{}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Review(context.Background(), Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FailedServices) != 0 {
		t.Fatalf("failed: %+v", result.FailedServices)
	}
	f, err := os.Open(result.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	seen := map[string]bool{}
	for decoder.More() {
		var event trace.Event
		if err := decoder.Decode(&event); err != nil {
			t.Fatal(err)
		}
		if event.Kind == "model" && event.Agent == "planner" && event.Model != nil {
			if event.Model.Route != "command" || event.Model.Model != "unknown" {
				t.Fatalf("invented identity: %+v", event.Model)
			}
			seen[event.Model.Requested] = true
		}
	}
	if !seen["first-cli"] || !seen["second-cli"] {
		t.Fatalf("both CLI planners did not run: %+v", seen)
	}
}
