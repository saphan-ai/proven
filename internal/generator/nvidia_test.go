package generator_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

// These tests cover nvidia mode (ADR-0031). The hosted API is answered by a
// recorded response served through the client's transport, so no port is
// opened and no real key or network is involved.

// hostedKey is a stand-in for an NVIDIA key. No real key is used anywhere.
const hostedKey = "nvapi-test-key-0123456789"

// roundTrip serves one request from a function, as a transport.
type roundTrip func(*http.Request) *http.Response

// RoundTrip lets a roundTrip stand in for the network.
func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r), nil }

// hostedStream is a recorded answer of the hosted API: OpenAI-compatible
// chunks whose own model field names the answering model, with no receipt.
func hostedStream(model string) string {
	chunk := func(content string) string {
		m := ""
		if model != "" {
			m = `"model":"` + model + `",`
		}
		return `data: {"id":"chatcmpl-1",` + m + `"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"` + content + `"}}]}` + "\n\n"
	}
	return chunk("rea") + chunk("dy") + "data: [DONE]\n\n"
}

// answer is a response with a status and a body.
func answer(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// hostedAt builds a nvidia-mode generator over a transport, with extra settings.
func hostedAt(t *testing.T, rt roundTrip, extra map[string]string) *generator.Live {
	t.Helper()
	env := map[string]string{
		generator.EnvMode: "nvidia",
		generator.EnvKey:  hostedKey,
	}
	for k, v := range extra {
		env[k] = v
	}
	cfg, err := generator.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	return generator.NewLive(cfg, &http.Client{Transport: rt})
}

// Proves the hosted answer is kept, is sent to the hosted API's address with
// the key as the only credential, and is recorded as "no receipt".
func TestNVIDIAKeepsAnNVIDIAAnswerWithNoReceipt(t *testing.T) {
	var got *http.Request
	var sent struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	live := hostedAt(t, func(r *http.Request) *http.Response {
		got = r
		_ = json.NewDecoder(r.Body).Decode(&sent)
		return answer(http.StatusOK, hostedStream("nvidia/nemotron-3.5-super-vl-120b-a12b"))
	}, nil)
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil || text != "ready" {
		t.Fatalf("text %q err %v", text, err)
	}
	if got.URL.String() != "https://integrate.api.nvidia.com/v1/chat/completions" {
		t.Errorf("url %s", got.URL)
	}
	if got.Header.Get("Authorization") != "Bearer "+hostedKey {
		t.Errorf("authorization header")
	}
	// Nightshift's contract headers are for Nightshift: the hosted API is not told an app or a scope.
	if got.Header.Get("meterless-app") != "" || got.Header.Get("meterless-scope") != "" {
		t.Errorf("Nightshift headers sent to the hosted API")
	}
	if sent.Model != "nvidia/nemotron-3.5-super-vl-120b-a12b" || !sent.Stream {
		t.Errorf("body %+v", sent)
	}
	u, ok := live.LastUse()
	if !ok || u.Model != "nvidia/nemotron-3.5-super-vl-120b-a12b" || u.ReceiptID != generator.NoReceipt || u.FromReceipt {
		t.Fatalf("use %#v ok %v", u, ok)
	}
}

// Proves the requirement is checked against the response's model field: a
// non-nvidia answer is refused where NVIDIA is required, and still named in the
// trace. Mutated field: the model field of the response chunks.
func TestNVIDIARefusesANonNVIDIAAnswerWhereRequired(t *testing.T) {
	live := hostedAt(t, func(*http.Request) *http.Response {
		return answer(http.StatusOK, hostedStream("openai/gpt-4o"))
	}, nil)
	text, err := live.Complete(context.Background(), "skeptic", "prompt")
	d, ok := service.As(err)
	if text != "" || !ok || d.Service != "nvidia" || !strings.Contains(d.Reason, "openai/gpt-4o") || !strings.Contains(d.Reason, "refused") {
		t.Fatalf("text %q err %v", text, err)
	}
	if u, ok := live.LastUse(); !ok || u.Model != "openai/gpt-4o" || u.ReceiptID != generator.NoReceipt {
		t.Fatalf("use %#v ok %v", u, ok)
	}
}

// Proves the requirement binds only where asked: with it off, or for an agent
// it does not list, a non-nvidia answer is kept. Mutated field:
// PROVEN_GENERATOR_REQUIRE_NVIDIA.
func TestNVIDIAKeepsANonNVIDIAAnswerWhereNotRequired(t *testing.T) {
	serve := func(*http.Request) *http.Response { return answer(http.StatusOK, hostedStream("openai/gpt-4o")) }
	for name, require := range map[string]string{"none": "none", "other agent": "skeptic"} {
		live := hostedAt(t, serve, map[string]string{generator.EnvRequireNVIDIA: require})
		if text, err := live.Complete(context.Background(), "planner", "prompt"); err != nil || text != "ready" {
			t.Errorf("%s: text %q err %v", name, text, err)
		}
	}
}

// Proves a response that names no model cannot satisfy the requirement, and
// that where it is not required the model is recorded as unknown, never
// guessed from the setting. Mutated field: the model field, removed.
func TestNVIDIARefusesAnAnswerThatNamesNoModelWhereRequired(t *testing.T) {
	serve := func(*http.Request) *http.Response { return answer(http.StatusOK, hostedStream("")) }
	live := hostedAt(t, serve, nil)
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err == nil || !strings.Contains(err.Error(), "names no model") {
		t.Fatalf("err %v, want a refusal that names the missing model", err)
	}
	open := hostedAt(t, serve, map[string]string{generator.EnvRequireNVIDIA: "none"})
	if _, err := open.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	if u, _ := open.LastUse(); u.Model != "unknown" || u.FromReceipt || u.ReceiptID != generator.NoReceipt {
		t.Fatalf("use %#v", u)
	}
}

// Proves the Skeptic is sent its own model, another family than the finder's
// (ADR-0006), in nvidia mode too.
func TestNVIDIASendsTheSkepticItsOwnModel(t *testing.T) {
	var models []string
	live := hostedAt(t, func(r *http.Request) *http.Response {
		var sent struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&sent)
		models = append(models, sent.Model)
		return answer(http.StatusOK, hostedStream(sent.Model))
	}, map[string]string{generator.EnvModel: "nvidia/finder-model"})
	// The two defaults are one model since 2026-10-26, so the finder is set to
	// another one: the Skeptic must still be sent its own.
	for _, agent := range []string{"planner", "skeptic"} {
		if _, err := live.Complete(context.Background(), agent, "prompt"); err != nil {
			t.Fatal(err)
		}
	}
	if len(models) != 2 || models[0] == models[1] || models[1] != generator.DefaultSkepticModel {
		t.Fatalf("models %v", models)
	}
}

// Proves the key never appears in any error the generator returns: not for a
// rejected key, a rate limit, a refused answer, nor a refused setting.
// Mutated field: the response body and the model field both carry the key.
func TestNVIDIAKeyIsNeverInAnError(t *testing.T) {
	echo := func(status int) roundTrip {
		return func(*http.Request) *http.Response { return answer(status, hostedKey) }
	}
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError} {
		_, err := hostedAt(t, echo(status), nil).Complete(context.Background(), "planner", "prompt")
		if err == nil || strings.Contains(err.Error(), hostedKey) {
			t.Errorf("status %d: err %v", status, err)
		}
	}
	// A refused answer names the model that answered; a model field is not a place the key may be echoed to by Proven itself.
	keyed := hostedAt(t, func(*http.Request) *http.Response { return answer(http.StatusOK, hostedStream("other/model")) }, nil)
	if _, err := keyed.Complete(context.Background(), "planner", "prompt"); err == nil || strings.Contains(err.Error(), hostedKey) {
		t.Errorf("err %v", err)
	}
	_, err := generator.Load(func(k string) string {
		return map[string]string{generator.EnvMode: "nvidia", generator.EnvKey: hostedKey + "\n"}[k]
	})
	if err == nil || strings.Contains(err.Error(), hostedKey) {
		t.Errorf("refused key: err %v", err)
	}
}

// Proves a rejected key is named as unauthorized and not retried, and a rate
// limit is named as one. Mutated field: the response status.
func TestNVIDIANamesARejectedKeyAndARateLimit(t *testing.T) {
	for status, want := range map[int]string{
		http.StatusUnauthorized:    "unauthorized",
		http.StatusForbidden:       "unauthorized",
		http.StatusTooManyRequests: "rate limited",
	} {
		var posts int
		live := hostedAt(t, func(*http.Request) *http.Response {
			posts++
			return answer(status, "")
		}, nil)
		_, err := live.Complete(context.Background(), "planner", "prompt")
		d, ok := service.As(err)
		if !ok || d.Service != "nvidia" || !strings.Contains(d.Reason, want) {
			t.Errorf("status %d: err %v", status, err)
		}
		if posts != 1 {
			t.Errorf("status %d: posts %d", status, posts)
		}
	}
}

// Proves an empty key is the hosted API down and named, not a refusal and not
// a call (ADR-0013 point 4a), and that the Nightshift token does not stand in
// for it. Mutated field: PROVEN_GENERATOR_KEY, empty.
func TestNVIDIAEmptyKeyIsDownAndNamed(t *testing.T) {
	_, err := generator.Load(func(k string) string {
		return map[string]string{generator.EnvMode: "nvidia", generator.EnvToken: "a-nightshift-token"}[k]
	})
	d, ok := service.As(err)
	if !ok || d.Service != "nvidia" || !strings.Contains(d.Reason, generator.EnvKey) || strings.Contains(d.Reason, "a-nightshift-token") {
		t.Fatalf("err %v", err)
	}
	// The absent generator the review installs names the same service.
	_, err = generator.Absent{Service: d.Service, Reason: d.Reason}.Complete(context.Background(), "planner", "prompt")
	if again, ok := service.As(err); !ok || again.Service != "nvidia" {
		t.Fatalf("absent: %v", err)
	}
}

// Proves a wrong mode is refused by name with the allowed values, and that no
// mode set means Nightshift, as before (ADR-0031 point 2).
func TestLoadRefusesAWrongModeAndDefaultsToNightshift(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvMode, "openai"))
	ce := configErr(t, err)
	if ce.Field != generator.EnvMode || !strings.Contains(err.Error(), "nightshift") || !strings.Contains(err.Error(), "nvidia") {
		t.Fatalf("err %v", err)
	}
	cfg, err := generator.Load(func(k string) string {
		return map[string]string{generator.EnvToken: "test-token"}[k]
	})
	if err != nil || cfg.Mode != generator.ModeNightshift {
		t.Fatalf("mode %q err %v", cfg.Mode, err)
	}
}

// Proves nvidia mode sends the key only to NVIDIA's hosted address: a base set
// for Nightshift (the variable the lab page tells a Nightshift user to set) is
// not read in nvidia mode, so the key and the request never go there
// (ADR-0031 point 3). Mutated field: PROVEN_GENERATOR_BASE_URL, set to a router.
func TestNVIDIAModeIgnoresTheNightshiftBase(t *testing.T) {
	var host string
	live := hostedAt(t, func(r *http.Request) *http.Response {
		host = r.URL.Host
		return answer(http.StatusOK, hostedStream("nvidia/nemotron-3.5-super-vl-120b-a12b"))
	}, map[string]string{generator.EnvBaseURL: "http://127.0.0.1:8088"})
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	if host != "integrate.api.nvidia.com" {
		t.Fatalf("request went to %q, want the hosted address", host)
	}
}

// Proves nvidia mode's defaults: the hosted address and a nvidia/* model.
func TestLoadNVIDIAModeDefaults(t *testing.T) {
	cfg, err := generator.Load(func(k string) string {
		return map[string]string{generator.EnvMode: "nvidia", generator.EnvKey: hostedKey}[k]
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://integrate.api.nvidia.com/v1" || !strings.HasPrefix(cfg.Model, "nvidia/") || cfg.Mode != generator.ModeNVIDIA {
		t.Fatalf("settings %+v", cfg)
	}
}

// Proves the alias is accepted only for the model it stands for: asked for the
// VL model, an answer named with its private alias is kept; asked for another
// model, the same alias is refused. Mutated field: the requested model.
func TestNVIDIAAcceptsAnAliasOnlyForTheModelAskedFor(t *testing.T) {
	const alias = "private/nvidia/nemotron-3.5-super-120b-a12b"
	serve := func(*http.Request) *http.Response { return answer(http.StatusOK, hostedStream(alias)) }
	live := hostedAt(t, serve, nil)
	if text, err := live.Complete(context.Background(), "planner", "prompt"); err != nil || text != "ready" {
		t.Fatalf("asked for the VL model: text %q err %v", text, err)
	}
	if u, ok := live.LastUse(); !ok || u.Model != alias {
		t.Fatalf("the trace must name the alias as answered: %#v", u)
	}
	other := hostedAt(t, serve, map[string]string{generator.EnvModel: "nvidia/other-model"})
	if _, err := other.Complete(context.Background(), "planner", "prompt"); err == nil || !strings.Contains(err.Error(), "not an NVIDIA model") {
		t.Fatalf("asked for another model: err %v, want a refusal", err)
	}
}
