package review_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestRuntimeTokenEchoedInTheReceiptIsNotWritten proves that the runtime token never reaches the
// trace or the review result, even when the stream's receipt fields echo it.
// Mutated field: meterless.routed_model and meterless.receipt_id, each carrying the token's value.
func TestRuntimeTokenEchoedInTheReceiptIsNotWritten(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only the generator is measured.
	t.Setenv("PATH", t.TempDir())
	const marker = "marker-token-7d41c0"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE("[]", "openai/"+marker, "rcpt_"+marker))
	}))
	defer srv.Close()
	res := reviewAt(t, srv.URL, map[string]string{"METERLESS_RUNTIME_TOKEN": marker})
	for _, m := range res.Models {
		if strings.Contains(m.Model, marker) || strings.Contains(m.ReceiptID, marker) || strings.Contains(m.Route, marker) {
			t.Errorf("the result's model use carries the token: %#v", m)
		}
	}
	for _, s := range res.FailedServices {
		if strings.Contains(s.Reason, marker) {
			t.Errorf("a failed service's reason carries the token: %#v", s)
		}
	}
	assertTraceLacks(t, res.TracePath, marker)
}

// TestGuardrailKeyEchoedInTheCategoriesIsNotWritten proves that the guardrail key never reaches the
// trace or the review result, even when the guardrail's answer echoes it in its categories line.
// Mutated field: the Safety Categories line of the guardrail's answer, carrying the key's value.
func TestGuardrailKeyEchoedInTheCategoriesIsNotWritten(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	const marker = "marker-guardkey-5e92ab"
	rail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"User Safety: unsafe\nResponse Safety: unsafe\nSafety Categories: `+marker+`"}}]}`)
	}))
	defer rail.Close()
	gen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE("[]", "nvidia/nemotron-3.5-super-vl-120b-a12b", "rcpt_g"))
	}))
	defer gen.Close()
	res := reviewAt(t, gen.URL, map[string]string{
		"PROVEN_GUARDRAIL_MODE":     "server",
		"PROVEN_GUARDRAIL_BASE_URL": rail.URL,
		"PROVEN_GUARDRAIL_KEY":      marker,
		"PROVEN_GUARDRAIL_TIMEOUT":  "1s",
	})
	for _, s := range res.FailedServices {
		if strings.Contains(s.Reason, marker) {
			t.Errorf("a failed service's reason carries the key: %#v", s)
		}
	}
	assertTraceLacks(t, res.TracePath, marker)
}

// TestARefusedSecretIsNamedNotPrinted proves that a secret refused for its form is named by its
// variable and its value never appears in the refusal.
// Mutated field: each secret in turn, set to a marker with a line break, so it is present and refused.
func TestARefusedSecretIsNamedNotPrinted(t *testing.T) {
	const marker = "marker-refused-3b7f"
	for _, name := range []string{"METERLESS_RUNTIME_TOKEN", "PROVEN_GUARDRAIL_KEY", "ZETARIS_API_KEY"} {
		t.Run(name, func(t *testing.T) {
			extra := map[string]string{name: marker + "\n"}
			if name == "PROVEN_GUARDRAIL_KEY" {
				extra["PROVEN_GUARDRAIL_MODE"] = "server"
			}
			_, err := review.Open(review.Options{
				Sample: false,
				LogDir: t.TempDir(),
				Env:    liveEnv("http://127.0.0.1:9", extra),
			})
			var cfg *review.ConfigError
			if err == nil || !errorAs(err, &cfg) || cfg.Field != name {
				t.Fatalf("refusal %v", err)
			}
			if strings.Contains(err.Error(), marker) || strings.Contains(cfg.Value, marker) {
				t.Fatalf("the refusal printed the secret: %s", err)
			}
		})
	}
}

// TestARefusedOrgIDIsNamedNotPrinted proves that a refused organisation id is named by its variable
// and by the rule it broke, and that its value never appears in the refusal.
// Mutated field: ZETARIS_ORG_ID, set to a value that is not digits only.
func TestARefusedOrgIDIsNamedNotPrinted(t *testing.T) {
	const marker = "org-marker-81c4"
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv("http://127.0.0.1:9", map[string]string{"ZETARIS_ORG_ID": marker}),
	})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "ZETARIS_ORG_ID" {
		t.Fatalf("refusal %v", err)
	}
	if strings.Contains(err.Error(), marker) || strings.Contains(cfg.Value, marker) {
		t.Fatalf("the refusal printed the organisation id: %s", err)
	}
	if !strings.Contains(err.Error(), "digits") {
		t.Fatalf("the refusal does not name the rule: %s", err)
	}
}

// TestOrgIDEchoedInTheReceiptIsNotWritten proves that the organisation id, which Proven sends with
// every query, does not reach the trace or the result when a service echoes it back.
// Mutated field: meterless.routed_model, carrying the organisation id's value.
func TestOrgIDEchoedInTheReceiptIsNotWritten(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	const org = "918273645501"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, answerSSE("[]", "openai/"+org, "rcpt_x"))
	}))
	defer srv.Close()
	res := reviewAt(t, srv.URL, map[string]string{"ZETARIS_ORG_ID": org})
	for _, m := range res.Models {
		if strings.Contains(m.Model, org) {
			t.Errorf("the result's model use carries the organisation id: %#v", m)
		}
	}
	for _, s := range res.FailedServices {
		if strings.Contains(s.Reason, org) {
			t.Errorf("a failed service's reason carries the organisation id: %#v", s)
		}
	}
	assertTraceLacks(t, res.TracePath, org)
}

// assertTraceLacks fails when the trace file holds marker.
func assertTraceLacks(t *testing.T, path, marker string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), marker) {
		t.Errorf("the trace carries %q:\n%s", marker, b)
	}
}
