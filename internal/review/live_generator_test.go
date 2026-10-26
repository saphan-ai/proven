package review_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/review"
)

// TestGuardrailModeIsRefusedOnOpen proves that a bad guardrail mode is refused when Open runs, before any
// review, and that the refusal never prints the guardrail key.
func TestGuardrailModeIsRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_GUARDRAIL_MODE.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_GUARDRAIL_MODE": "sometimes",
			"PROVEN_GUARDRAIL_KEY":  "super-secret",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_GUARDRAIL_MODE" || cfg.Value != "sometimes" {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatal("the refusal printed the key")
	}
}

// TestLiveSettingsAreRefusedOnOpen proves that a bad generator setting is refused on Open as a
// *ConfigError naming the variable, and that the refusal never prints the token.
func TestLiveSettingsAreRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_GENERATOR_TIMEOUT.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"METERLESS_RUNTIME_TOKEN":  "super-secret",
			"PROVEN_GENERATOR_TIMEOUT": "nope",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_GENERATOR_TIMEOUT" || cfg.Value != "nope" {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatal("the refusal printed the token")
	}
}

// TestEPSSViewNameIsRefusedOnOpen proves that an EPSS view name that is not a plain identifier is refused
// on Open, so it never reaches a query.
func TestEPSSViewNameIsRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_EPSS_VIEW.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_EPSS_VIEW": "px_evidence.bad;drop",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_EPSS_VIEW" || cfg.Value != "px_evidence.bad;drop" {
		t.Fatal(err)
	}
}

// TestOSVTableNameIsRefusedOnOpen proves the same for the OSV table name.
func TestOSVTableNameIsRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_OSV_TABLE.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_OSV_TABLE": "px_osv.pkg;drop",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_OSV_TABLE" || cfg.Value != "px_osv.pkg;drop" {
		t.Fatal(err)
	}
}

// TestOSVIDViewNameIsRefusedOnOpen proves the same for the advisory-id view name.
func TestOSVIDViewNameIsRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_OSV_ID_VIEW.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_OSV_ID_VIEW": "px_usl.bad;drop",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_OSV_ID_VIEW" || cfg.Value != "px_usl.bad;drop" {
		t.Fatal(err)
	}
}

// TestKEVViewNameIsRefusedOnOpen proves that a KEV view name that is not a plain identifier is refused
// on Open, so it never reaches a query.
func TestKEVViewNameIsRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_KEV_VIEW.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_KEV_VIEW": "px_evidence.bad;drop",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_KEV_VIEW" || cfg.Value != "px_evidence.bad;drop" {
		t.Fatal(err)
	}
}

// TestAdapterTimeoutIsRefusedOnOpen proves that an adapter bound that is not a duration is refused
// on Open, before any review.
func TestAdapterTimeoutIsRefusedOnOpen(t *testing.T) {
	// Mutated field: PROVEN_ADAPTER_TIMEOUT.
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_ADAPTER_TIMEOUT": "nope",
		}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "PROVEN_ADAPTER_TIMEOUT" || cfg.Value != "nope" {
		t.Fatal(err)
	}
}

// TestLiveReviewCarriesTheRoutedModel proves that a live review reports the model and receipt the
// provider's stream named (ADR-0019), not the model it asked for.
func TestLiveReviewCarriesTheRoutedModel(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only the generator is measured.
	t.Setenv("PATH", t.TempDir())
	const model = "nvidia/nemotron-3.5-super-vl-120b-a12b"
	const receipt = "rcpt_test_1"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("meterless-scope") != "user" {
			t.Errorf("request %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE("[]", model, receipt))
	}))
	defer srv.Close()
	res := reviewAt(t, srv.URL, nil)
	if len(res.Models) != 1 {
		t.Fatalf("models %#v", res.Models)
	}
	m := res.Models[0]
	if m.Agent != "planner" || m.Model != model || m.ReceiptID != receipt || !m.FromReceipt {
		t.Fatalf("model %#v", m)
	}
	for _, s := range res.FailedServices {
		if s.Service == "nightshift" {
			t.Fatalf("failed %#v", res.FailedServices)
		}
	}
}

// TestLiveReviewRefusesANonNVIDIAModelByName proves that when NVIDIA is required and the provider routes
// to another model, the review keeps no finding from that answer, names the refused model in the failed
// service, and still reports the model the receipt named.
// Mutated field: meterless.routed_model, set to a non-NVIDIA model.
func TestLiveReviewRefusesANonNVIDIAModelByName(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only the generator is measured.
	t.Setenv("PATH", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE("[]", "openai/gpt-4o", "rcpt_test_2"))
	}))
	defer srv.Close()
	res := reviewAt(t, srv.URL, nil)
	if len(res.Findings) != 0 {
		t.Fatalf("findings %#v", res.Findings)
	}
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "nightshift" && strings.Contains(s.Reason, "openai/gpt-4o") && strings.Contains(s.Reason, "refused") {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	if len(res.Models) != 1 || res.Models[0].Model != "openai/gpt-4o" || !res.Models[0].FromReceipt {
		t.Fatalf("models %#v", res.Models)
	}
}

// TestUnauthorizedReviewDegrades proves that a refused token degrades the review instead of failing it:
// Nightshift is named as failed, "unauthorized", and the request is not retried.
// Mutated field: the HTTP status, set to 401.
func TestUnauthorizedReviewDegrades(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only the generator is measured.
	t.Setenv("PATH", t.TempDir())
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	res := reviewAt(t, srv.URL, nil)
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "nightshift" && s.Reason == "unauthorized" {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	if posts != 1 {
		t.Fatalf("posts %d", posts)
	}
}

// TestClosedEndpointDegradesTheReview proves that an unreachable provider degrades the review: Nightshift
// is named as failed, "unreachable", and the review still returns.
// Mutated field: the listening socket. The endpoint is closed before the review.
func TestClosedEndpointDegradesTheReview(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only the generator is measured.
	t.Setenv("PATH", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed endpoint accepted a call")
	}))
	url := srv.URL
	srv.Close()
	res := reviewAt(t, url, map[string]string{"PROVEN_GENERATOR_TIMEOUT": "1s"})
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "nightshift" && s.Reason == "unreachable" {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
}

// reviewAt runs one live review against a generator at base and fails the test on a partial result.
func reviewAt(t *testing.T, base string, extra map[string]string) review.Result {
	t.Helper()
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv(base, extra),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := rev.Review(ctx, review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("partial")
	}
	return res
}

// liveEnv is the live environment of these tests: the generator at base, Zetaris at an address that
// answers nothing, and the guardrail off. extra overrides any key.
func liveEnv(base string, extra map[string]string) func(string) string {
	m := map[string]string{
		"METERLESS_RUNTIME_TOKEN":         "test-token",
		"PROVEN_GENERATOR_BASE_URL":       base,
		"PROVEN_GENERATOR_TIMEOUT":        "2s",
		"PROVEN_GENERATOR_APP":            "proven",
		"PROVEN_GENERATOR_MODEL":          "meterless/auto",
		"PROVEN_GENERATOR_REQUIRE_NVIDIA": "all",
		"ZETARIS_REST_URL":                "http://127.0.0.1:9",
		"ZETARIS_API_KEY":                 "zetaris-test-key",
		"ZETARIS_ORG_ID":                  "1",
		"ZETARIS_TIMEOUT":                 "1s",
		// These reviews measure the generator. The guardrail's own tests cover the on path.
		"PROVEN_GUARDRAIL_MODE": "off",
	}
	for k, v := range extra {
		m[k] = v
	}
	return func(k string) string { return m[k] }
}

// answerSSE is one streamed answer in Nightshift's form: one chunk carrying content and the meterless
// routed model and receipt, then [DONE].
func answerSSE(content, model, receipt string) string {
	chunk := map[string]any{
		"model": "meterless/auto",
		"choices": []any{
			map[string]any{"index": 0, "delta": map[string]any{"content": content}},
		},
		"meterless": map[string]any{"routed_model": model, "receipt_id": receipt},
	}
	raw, err := json.Marshal(chunk)
	if err != nil {
		panic(err)
	}
	return "data: " + string(raw) + "\n\ndata: [DONE]\n"
}
