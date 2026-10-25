package review

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLiveReviewWithEmptySecretsNamesDownServices proves ADR-0013 point 4a
// and ADR-0005 point 5 together. With no secret set, a live review of a
// scratch git repository completes, names nightshift, zetaris and the
// guardrail, and names a tracked symlink it did not copy.
//
// ZETARIS_REST_URL and ZETARIS_ORG_ID are set: they are not secrets, and an
// empty URL is still a refusal. The three secrets are absent.
// The test runner seam is set so the proof does not start a container.
// The review is empty because the guardrail refuses the planner before the
// generator runs; nightshift is still named, from the empty token recorded
// when the review opened.
func TestLiveReviewWithEmptySecretsNamesDownServices(t *testing.T) {
	repo, base, head, _ := commitsWithSymlink(t)
	rev, err := Open(Options{
		Sample:  false,
		LogDir:  t.TempDir(),
		Env:     emptySecretEnv(),
		testRun: quietRun{},
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := rev.Review(ctx, Request{
		Target: repo, BaselineCommit: base, HeadCommit: head,
	})
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("findings %#v", res.Findings)
	}
	want := map[string]string{
		"nightshift": "METERLESS_RUNTIME_TOKEN",
		"zetaris":    "ZETARIS_API_KEY",
		"guardrail":  "PROVEN_GUARDRAIL_KEY",
	}
	got := map[string]string{}
	for _, s := range res.FailedServices {
		got[s.Service] = s.Reason
	}
	for service, variable := range want {
		reason, ok := got[service]
		if !ok || !strings.Contains(reason, variable) {
			t.Fatalf("failed services %#v", res.FailedServices)
		}
	}
	if len(res.SkippedPaths) != 1 || res.SkippedPaths[0] != "linked.txt" {
		t.Fatalf("skipped %#v", res.SkippedPaths)
	}
}

// TestEmptyGuardrailKeyRefusesEveryModelCall proves ADR-0005 point 5.
// The guardrail key is empty, live mode is on, and every other secret is set.
// The stub generator counts its calls. The count stays zero, each refusal is
// a trace line, and the guardrail is in the failed services.
// One line is written per refused call, not one line for the whole run.
func TestEmptyGuardrailKeyRefusesEveryModelCall(t *testing.T) {
	repo, base, head := twoCommits(t)
	gen := &countGen{}
	logDir := t.TempDir()
	rev, err := Open(Options{
		Sample:    false,
		LogDir:    logDir,
		Env:       liveSecrets(map[string]string{"PROVEN_GUARDRAIL_KEY": ""}),
		gen:       gen,
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
	if gen.calls != 0 {
		t.Fatalf("generator calls %d", gen.calls)
	}
	if !failedHas(res, "guardrail", "PROVEN_GUARDRAIL_KEY") {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	body := traceBody(t, res.TracePath)
	if !strings.Contains(body, "guardrail: PROVEN_GUARDRAIL_KEY is empty") {
		t.Fatalf("trace has no refusal\n%s", body)
	}
}

// TestGuardrailOffWithEmptyKeyLetsModelCallsThrough proves that mode off is
// the only switch. The key is empty, the mode is off, model calls run, and
// the trace says the run had no guardrail.
func TestGuardrailOffWithEmptyKeyLetsModelCallsThrough(t *testing.T) {
	repo, base, head := twoCommits(t)
	gen := &countGen{}
	rev, err := Open(Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveSecrets(map[string]string{
			"PROVEN_GUARDRAIL_KEY":  "",
			"PROVEN_GUARDRAIL_MODE": "off",
		}),
		gen:       gen,
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
	if gen.calls == 0 {
		t.Fatal("no model call")
	}
	if !strings.Contains(traceBody(t, res.TracePath), "no guardrail") {
		t.Fatal("trace did not say the run had no guardrail")
	}
}

// TestWrongSecretsAreRefusedBeforeTheReview proves a non-empty wrong secret
// is still a configuration refusal. The review is not started.
// Mutated field, one case at a time: the secret, set to a value with a line break.
func TestWrongSecretsAreRefusedBeforeTheReview(t *testing.T) {
	cases := []struct {
		name  string
		env   map[string]string
		field string
	}{
		{
			name:  "token",
			env:   map[string]string{"METERLESS_RUNTIME_TOKEN": "not-a-secret\n"},
			field: "METERLESS_RUNTIME_TOKEN",
		},
		{
			name: "guardrail",
			env: liveMap(map[string]string{
				"PROVEN_GUARDRAIL_KEY": "not-a-secret\n",
			}),
			field: "PROVEN_GUARDRAIL_KEY",
		},
		{
			name: "zetaris",
			env: liveMap(map[string]string{
				"ZETARIS_API_KEY": "not-a-secret\n",
			}),
			field: "ZETARIS_API_KEY",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rev, err := Open(Options{
				Sample:  false,
				LogDir:  t.TempDir(),
				Env:     func(k string) string { return tc.env[k] },
				testRun: quietRun{},
			})
			if err == nil {
				t.Fatal("open accepted a broken secret")
			}
			if rev != nil {
				t.Fatal("a reviewer was returned")
			}
			var cfg *ConfigError
			if !errors.As(err, &cfg) || cfg.Field != tc.field {
				t.Fatalf("%v", err)
			}
			if strings.Contains(err.Error(), "not-a-secret") {
				t.Fatal("the refusal printed the secret")
			}
		})
	}
}

// TestEmptyZetarisKeyAddsNoQueryEvidence proves that with Zetaris down the
// investigator's query is unavailable and the service is named. The planner
// runs, so the query is attempted.
func TestEmptyZetarisKeyAddsNoQueryEvidence(t *testing.T) {
	repo, base, head := twoCommits(t)
	gen := &countGen{}
	rev, err := Open(Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveSecrets(map[string]string{
			"ZETARIS_API_KEY":       "",
			"PROVEN_GUARDRAIL_MODE": "off",
		}),
		gen:     gen,
		testRun: quietRun{},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatal(err)
	}
	if gen.agents["investigator"] == 0 {
		t.Fatal("the investigator did not run")
	}
	if !failedHas(res, "zetaris", "ZETARIS_API_KEY") {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	for _, f := range res.Findings {
		for _, e := range f.Evidence {
			if e.Kind == EvidenceQuery {
				t.Fatalf("query evidence %#v", e.Query)
			}
		}
	}
}

// TestEmptyTokenLeavesAnEmptyReviewNamingNightshift proves that a down
// generator, with the guardrail off so the call reaches it, leaves an empty
// review that names Nightshift.
func TestEmptyTokenLeavesAnEmptyReviewNamingNightshift(t *testing.T) {
	repo, base, head := twoCommits(t)
	rev, err := Open(Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveSecrets(map[string]string{
			"METERLESS_RUNTIME_TOKEN": "",
			"PROVEN_GUARDRAIL_MODE":   "off",
		}),
		testRun: quietRun{},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("findings %#v", res.Findings)
	}
	if !failedHas(res, "nightshift", "METERLESS_RUNTIME_TOKEN") {
		t.Fatalf("failed %#v", res.FailedServices)
	}
}

// countGen is the stub generator. It counts every call, so a guardrail that
// lets one through is visible, and it answers the planner and the investigator
// with the JSON those agents ask for.
type countGen struct {
	calls  int
	agents map[string]int
}

// Complete records the call and returns a small answer for that agent.
func (g *countGen) Complete(_ context.Context, agent, _ string) (string, error) {
	g.calls++
	if g.agents == nil {
		g.agents = map[string]int{}
	}
	g.agents[agent]++
	switch agent {
	case "planner":
		return `[{"id":"F1","rule_id":"r","title":"t","message":"m","severity":"ERROR","path":"note.txt","start_line":1,"end_line":1}]`, nil
	case "investigator":
		return `{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`, nil
	default:
		return `{}`, nil
	}
}

// emptySecretEnv leaves the three secrets unset and sets the non-secret
// values a live query load still requires.
func emptySecretEnv() func(string) string {
	set := map[string]string{
		"ZETARIS_REST_URL": "http://127.0.0.1:9",
		"ZETARIS_ORG_ID":   "1",
	}
	return func(k string) string { return set[k] }
}

// liveSecrets is a live environment with every secret set, then extra applied.
// An empty extra value removes that variable.
func liveSecrets(extra map[string]string) func(string) string {
	set := liveMap(extra)
	return func(k string) string { return set[k] }
}

// liveMap is the environment map liveSecrets closes over.
func liveMap(extra map[string]string) map[string]string {
	set := map[string]string{
		"METERLESS_RUNTIME_TOKEN": "test-token",
		"PROVEN_GUARDRAIL_KEY":    "guard-test-key",
		"ZETARIS_REST_URL":        "http://127.0.0.1:9",
		"ZETARIS_API_KEY":         "zetaris-test-key",
		"ZETARIS_ORG_ID":          "1",
	}
	for k, v := range extra {
		if v == "" {
			delete(set, k)
		} else {
			set[k] = v
		}
	}
	return set
}

// reviewWith runs one review under a deadline.
func reviewWith(t *testing.T, rev Reviewer, req Request) (Result, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return rev.Review(ctx, req)
}

// failedHas reports whether a failed service names variable.
func failedHas(res Result, service, variable string) bool {
	for _, s := range res.FailedServices {
		if s.Service == service && strings.Contains(s.Reason, variable) {
			return true
		}
	}
	return false
}

// traceBody is the trace file.
func traceBody(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
