package review_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/review"
)

// TestZetarisSettingsAreRefusedOnOpen proves that a bad Zetaris setting is refused on Open as a
// *ConfigError naming the variable, and that the refusal never prints the API key.
// Mutated field: ZETARIS_ORG_ID.
func TestZetarisSettingsAreRefusedOnOpen(t *testing.T) {
	const secret = "super-secret-zetaris-key"
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"ZETARIS_API_KEY": secret,
			"ZETARIS_ORG_ID":  "not-numeric",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "ZETARIS_ORG_ID" {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("the refusal printed the API key")
	}
}

// TestZetarisUnauthorizedReviewCompletes proves that a refused Zetaris key degrades the review instead
// of failing it: Zetaris is named once as "unauthorized", the review completes, and the key that
// Zetaris echoed back never reaches the failure reason.
// Mutated field: the HTTP status of the Zetaris response, set to 401.
func TestZetarisUnauthorizedReviewCompletes(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only Zetaris is measured.
	t.Setenv("PATH", t.TempDir())
	const secret = "super-secret-zetaris-key"
	var calls int
	zetaris := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/sql-editor/sqls/run" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"msg":"`+secret+`"}`)
	}))
	defer zetaris.Close()
	gen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE(agentJSON(string(body)), "nvidia/nemotron-3.5-super-vl-120b-a12b", "rcpt_z"))
	}))
	defer gen.Close()
	res := reviewLive(t, gen.URL, map[string]string{
		"ZETARIS_REST_URL": zetaris.URL,
		"ZETARIS_API_KEY":  secret,
		"ZETARIS_TIMEOUT":  "1s",
	})
	// The investigator's KEV query, then the dependency check's OSV query.
	// Both are refused. The review names the service once.
	if calls != 2 {
		t.Fatalf("zetaris calls %d", calls)
	}
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "zetaris" && s.Reason == "unauthorized" {
			found = true
		}
		if strings.Contains(s.Reason, secret) {
			t.Fatal("the failure printed the API key")
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	if res.Partial {
		t.Fatal("partial")
	}
}

// TestZetarisClosedEndpointDegradesTheReview proves that an unreachable Zetaris degrades the review: it
// is named as failed and the review still completes within its deadline.
// Mutated field: the listening socket of the fake Zetaris. It is closed before the review.
func TestZetarisClosedEndpointDegradesTheReview(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only Zetaris is measured.
	t.Setenv("PATH", t.TempDir())
	zetaris := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed endpoint accepted a call")
	}))
	url := zetaris.URL
	zetaris.Close()
	gen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE(agentJSON(string(body)), "nvidia/nemotron-3.5-super-vl-120b-a12b", "rcpt_z"))
	}))
	defer gen.Close()
	res := reviewLive(t, gen.URL, map[string]string{
		"ZETARIS_REST_URL": url,
		"ZETARIS_TIMEOUT":  "1s",
	})
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "zetaris" && (s.Reason == "unreachable" || s.Reason == "timeout") {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	if res.Partial {
		t.Fatal("partial")
	}
}

// reviewLive runs one live review with the generator at base and an 8s deadline.
func reviewLive(t *testing.T, base string, extra map[string]string) review.Result {
	t.Helper()
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv(base, extra),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	res, err := rev.Review(ctx, review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// agentJSON answers each agent's prompt, told apart by its instruction line: one known-exploited risk,
// its KEV query, no test, and no refutation. The answer then needs Zetaris to become a finding.
func agentJSON(prompt string) string {
	switch {
	case strings.Contains(prompt, "Name the risks"):
		return `[{"id":"K1","rule_id":"known-exploited-cve","title":"CVE-2021-44228 is on the KEV list","message":"The advisory CVE is on the known-exploited list.","severity":"ERROR","path":"go.mod","start_line":5,"end_line":5}]`
	case strings.Contains(prompt, "Write one SQL"):
		return `{"sql":"SELECT cve_id FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2021-44228'"}`
	case strings.Contains(prompt, "Write one test"):
		return `{"path":"","name":"","command":""}`
	case strings.Contains(prompt, "Make one attempt"):
		return `{"refuted":false,"reason":"no refutation"}`
	default:
		return `{}`
	}
}
