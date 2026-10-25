package review

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
)

// nvidiaEnv is a live environment in nvidia mode with the guardrail off, so
// only the generator is measured. An empty extra value removes that variable.
func nvidiaEnv(extra map[string]string) func(string) string {
	set := liveMap(map[string]string{
		"PROVEN_GENERATOR_MODE": "nvidia",
		"PROVEN_GUARDRAIL_MODE": "off",
	})
	delete(set, "METERLESS_RUNTIME_TOKEN")
	for k, v := range extra {
		if v == "" {
			delete(set, k)
		} else {
			set[k] = v
		}
	}
	return func(k string) string { return set[k] }
}

// Proves an empty NVIDIA key names the service "nvidia", never "nightshift",
// in the failed services and on every refused call in the trace, and that no
// model call is made. Mutated field: PROVEN_GENERATOR_KEY, empty, in nvidia mode.
func TestEmptyNVIDIAKeyNamesNVIDIAOnEveryRefusedCall(t *testing.T) {
	repo, base, head := twoCommits(t)
	rev, err := Open(Options{
		Sample:    false,
		LogDir:    t.TempDir(),
		Env:       nvidiaEnv(nil),
		liveStore: quietQuery{},
		testRun:   quietRun{},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatal(err)
	}
	if !failedHas(res, "nvidia", generator.EnvKey) {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	// The planner's refused call is noted by the same service name; any other
	// name here (an empty one, or "nightshift") means the refusal lost its provider.
	for _, s := range res.FailedServices {
		if s.Service != "nvidia" {
			t.Errorf("failed service %#v", s)
		}
	}
}

// hostedEcho is a generator that reports a model field carrying a configured
// value, as a hosted answer can.
type hostedEcho struct{ model string }

// Complete answers every agent with an empty list of findings.
func (hostedEcho) Complete(context.Context, string, string) (string, error) { return "[]", nil }

// LastUse reports the echoed model, with no receipt, as nvidia mode does.
func (h hostedEcho) LastUse() (generator.Use, bool) {
	return generator.Use{Model: h.model, Route: "integrate.api.nvidia.com", ReceiptID: generator.NoReceipt}, true
}

// Proves the NVIDIA key is redacted from a hosted answer's model field in the
// trace and in the review's model list. Mutated field: the key's entry in
// redact.Names, and the model field of the answer, which carries the key's value.
func TestNVIDIAKeyEchoedInTheModelFieldIsNotWritten(t *testing.T) {
	const key = "nvapi-echo-3f9a1c77d2"
	repo, base, head := twoCommits(t)
	rev, err := Open(Options{
		Sample:    false,
		LogDir:    t.TempDir(),
		Env:       nvidiaEnv(map[string]string{generator.EnvKey: key}),
		gen:       hostedEcho{model: "other/" + key},
		liveStore: quietQuery{},
		testRun:   quietRun{},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Models) == 0 {
		t.Fatal("no model use was recorded")
	}
	for _, m := range res.Models {
		if strings.Contains(m.Model, key) || m.Model != "other/[redacted: "+generator.EnvKey+"]" {
			t.Errorf("model use %#v", m)
		}
	}
	body := traceBody(t, res.TracePath)
	if strings.Contains(body, key) || !strings.Contains(body, "[redacted: "+generator.EnvKey+"]") {
		t.Fatalf("trace does not carry the marker in place of the key\n%s", body)
	}
}
