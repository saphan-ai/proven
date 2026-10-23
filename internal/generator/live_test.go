package generator_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

// The routed model and receipt id in the recorded stream, and a model outside the NVIDIA family.
const (
	routedNVIDIA = "nvidia/nemotron-3.5-super-vl-120b-a12b"
	receiptID    = "rcpt_req_k67ooqsx"
	otherModel   = "openai/gpt-4o"
)

// A recorded Nightshift stream gives its text, one POST with the contract
// headers, and the routed model and receipt id from the receipt.
func TestLiveReadsTheProvisionedStream(t *testing.T) {
	body := streamFixture(t)
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		assertContract(t, r, "meterless/auto", "proven")
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	text, err := live.Complete(context.Background(), "planner", "name the risks\n")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ready" {
		t.Fatalf("text %q", text)
	}
	if posts != 1 {
		t.Fatalf("posts %d", posts)
	}
	u, ok := live.LastUse()
	if !ok || u.Model != routedNVIDIA || u.ReceiptID != receiptID || !u.FromReceipt || u.Route != "" {
		t.Fatalf("use %#v ok %v", u, ok)
	}
}

// The configured model is sent, but the reported model comes from the receipt, not from the setting.
// Mutated field: PROVEN_GENERATOR_MODEL. The stream still names the NVIDIA model.
func TestLiveSendsTheConfiguredModelAndDoesNotTrustIt(t *testing.T) {
	sse := streamFixture(t)
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model  string `json:"model"`
			Stream bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		got = body.Model
		if !body.Stream {
			t.Error("stream false")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, sse)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, map[string]string{generator.EnvModel: "meterless/fast"})
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ready" || got != "meterless/fast" {
		t.Fatalf("text %q sent %q", text, got)
	}
	u, _ := live.LastUse()
	if u.Model != routedNVIDIA || !u.FromReceipt {
		t.Fatalf("use %#v", u)
	}
}

// The Skeptic sends Nemotron by default and the configured Skeptic model when
// one is set (ADR-0006).
func TestLiveSendsNemotronForTheSkeptic(t *testing.T) {
	sse := streamFixture(t)
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		got = body.Model
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, sse)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	if _, err := live.Complete(context.Background(), "skeptic", "prompt"); err != nil {
		t.Fatal(err)
	}
	if got != generator.DefaultSkepticModel {
		t.Fatalf("sent %s", got)
	}
	live = liveAt(t, srv.URL, map[string]string{generator.EnvSkepticModel: "anthropic/claude-opus"})
	if _, err := live.Complete(context.Background(), "skeptic", "prompt"); err != nil {
		t.Fatal(err)
	}
	if got != "anthropic/claude-opus" {
		t.Fatalf("sent %s", got)
	}
}

// A receipt naming a non-NVIDIA model is refused by name (ADR-0019 point 4).
// Mutated field: meterless.routed_model. The chunk model field stays nvidia/*.
func TestLiveRefusesANonNVIDIARoutedModel(t *testing.T) {
	body := strings.Replace(streamFixture(t), `"routed_model":"`+routedNVIDIA+`"`, `"routed_model":"`+otherModel+`"`, 1)
	if strings.Contains(body, `"routed_model":"`+routedNVIDIA+`"`) {
		t.Fatal("routed model was not mutated")
	}
	srv := httptest.NewServer(replay(body))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	text, err := live.Complete(context.Background(), "skeptic", "prompt")
	if text != "" {
		t.Fatalf("text %q", text)
	}
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || !strings.Contains(d.Reason, otherModel) || !strings.Contains(d.Reason, "refused") {
		t.Fatalf("got %v", err)
	}
	u, ok := live.LastUse()
	if !ok || u.Model != otherModel || !u.FromReceipt {
		t.Fatalf("use %#v ok %v", u, ok)
	}
}

// With the requirement off, a non-NVIDIA answer is kept.
// Mutated field: PROVEN_GENERATOR_REQUIRE_NVIDIA, set to none.
func TestLiveAcceptsANonNVIDIAModelWhenTheAgentIsNotRequired(t *testing.T) {
	body := strings.Replace(streamFixture(t), `"routed_model":"`+routedNVIDIA+`"`, `"routed_model":"`+otherModel+`"`, 1)
	srv := httptest.NewServer(replay(body))
	defer srv.Close()
	live := liveAt(t, srv.URL, map[string]string{generator.EnvRequireNVIDIA: "none"})
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ready" {
		t.Fatalf("text %q", text)
	}
	u, _ := live.LastUse()
	if u.Model != otherModel || !u.FromReceipt {
		t.Fatalf("use %#v", u)
	}
}

// The requirement applies only to the agents it lists.
// Mutated field: PROVEN_GENERATOR_REQUIRE_NVIDIA, set to skeptic only.
// The planner is outside that list, so the same non-NVIDIA answer is kept.
func TestLiveKeepsANonNVIDIAAnswerForAnAgentThatIsNotListed(t *testing.T) {
	body := strings.Replace(streamFixture(t), `"routed_model":"`+routedNVIDIA+`"`, `"routed_model":"`+otherModel+`"`, 1)
	srv := httptest.NewServer(replay(body))
	defer srv.Close()
	live := liveAt(t, srv.URL, map[string]string{generator.EnvRequireNVIDIA: "skeptic"})
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ready" {
		t.Fatalf("text %q", text)
	}
	if _, err := live.Complete(context.Background(), "skeptic", "prompt"); err == nil {
		t.Fatal("the skeptic is on the list and must be refused")
	}
}

// Without the receipt the model is reported as unknown; no other field stands in for it.
// Mutated field: the top-level meterless object, removed.
// The chunk model field and meterless_deliberation.receipt_id stay, and neither is the receipt.
func TestLiveReportsModelUnknownWhenMeterlessIsRemoved(t *testing.T) {
	body := withoutMeterless(t, streamFixture(t))
	srv := httptest.NewServer(replay(body))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ready" {
		t.Fatalf("text %q", text)
	}
	u, ok := live.LastUse()
	if !ok || u.Model != "unknown" || u.FromReceipt || u.ReceiptID != "" || u.Route != "" {
		t.Fatalf("use %#v ok %v", u, ok)
	}
}

// The winning provider in the receipt is reported as the route.
func TestLiveReadsTheProviderWhenTheStreamNamesIt(t *testing.T) {
	body := strings.Replace(streamFixture(t), `"routed_model":"`+routedNVIDIA+`"`, `"routed_model":"`+routedNVIDIA+`","winning_route":{"provider":"nvidia"}`, 1)
	srv := httptest.NewServer(replay(body))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	u, _ := live.LastUse()
	if u.Route != "nvidia" || u.Model != routedNVIDIA {
		t.Fatalf("use %#v", u)
	}
}

// A 401 is named as unauthorized after one request; it is not retried.
// Mutated field: the HTTP status, set to 401.
func TestLiveNamesUnauthorizedAndDoesNotRetry(t *testing.T) {
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		assertContract(t, r, "meterless/auto", "proven")
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	_, err := live.Complete(context.Background(), "planner", "prompt")
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "unauthorized" {
		t.Fatalf("got %v", err)
	}
	if posts != 1 {
		t.Fatalf("posts %d", posts)
	}
	if _, ok := live.LastUse(); ok {
		t.Fatal("a 401 has no model")
	}
}

// A server slower than the timeout is named as a timeout.
// Mutated field: the server does not answer before the timeout.
func TestLiveNamesATimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer func() {
		close(release)
		srv.Close()
	}()
	live := liveAt(t, srv.URL, map[string]string{generator.EnvTimeout: "200ms"})
	start := time.Now()
	_, err := live.Complete(context.Background(), "planner", "prompt")
	if time.Since(start) > 3*time.Second {
		t.Fatal("the call hung past its timeout")
	}
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "timeout" {
		t.Fatalf("got %v", err)
	}
}

// An empty answer is retried once, and a second empty answer is a named failure.
// Mutated field: the stream body, replaced with an empty event stream.
func TestLiveRetriesAnEmptyStreamOnce(t *testing.T) {
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: [DONE]\n")
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	_, err := live.Complete(context.Background(), "planner", "prompt")
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "empty answer" {
		t.Fatalf("got %v", err)
	}
	if posts != 2 {
		t.Fatalf("posts %d", posts)
	}
}

// When the retry answers, its text and its receipt are kept.
func TestLiveRetriesAnEmptyStreamThenKeepsTheSecond(t *testing.T) {
	var posts int
	body := streamFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		w.Header().Set("Content-Type", "text/event-stream")
		if posts == 1 {
			_, _ = io.WriteString(w, "data: [DONE]\n")
			return
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ready" || posts != 2 {
		t.Fatalf("text %q posts %d", text, posts)
	}
	u, _ := live.LastUse()
	if u.Model != routedNVIDIA || u.ReceiptID != receiptID || !u.FromReceipt {
		t.Fatalf("use %#v", u)
	}
}

// The models call is GET /v1/models with the same contract headers as a completion.
func TestLiveDiscoverySendsTheContract(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Method + " " + r.URL.Path
		if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("meterless-app") != "proven" || r.Header.Get("meterless-scope") != "user" {
			t.Errorf("headers")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	if err := live.Models(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got != "GET /v1/models" {
		t.Fatalf("got %s", got)
	}
}

// A base URL ending in /v1 does not get a second /v1.
// Mutated field: the base URL path, set to /v1. The call stays /v1/chat/completions.
func TestLivePostsOnceWhenTheBaseIsTheAPIRoot(t *testing.T) {
	body := streamFixture(t)
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL+"/v1", nil)
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	if path != "/v1/chat/completions" {
		t.Fatalf("path %s", path)
	}
}

// A closed port is named as unreachable, not as a timeout.
func TestClosedPortIsUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	live := liveAt(t, url, map[string]string{generator.EnvTimeout: "1s"})
	_, err := live.Complete(context.Background(), "planner", "prompt")
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "unreachable" {
		t.Fatalf("got %v", err)
	}
}

// WithModel sends its own model and leaves the original generator's model and last use alone.
func TestLiveWithModelSendsThatModelAndKeepsTheOriginal(t *testing.T) {
	sse := streamFixture(t)
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		got = append(got, body.Model)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, sse)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	pointed := live.WithModel("qwen-special")
	if _, err := pointed.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "qwen-special" || got[1] != "meterless/auto" {
		t.Fatalf("sent %#v", got)
	}
	su, ok := pointed.LastUse()
	du, dok := live.LastUse()
	if !ok || !dok || su.Model != routedNVIDIA || du.Model != routedNVIDIA {
		t.Fatalf("special %#v %v default %#v %v", su, ok, du, dok)
	}
}

// liveAt builds a live generator for base with a test token and any extra settings.
func liveAt(t *testing.T, base string, extra map[string]string) *generator.Live {
	t.Helper()
	env := map[string]string{
		generator.EnvToken:         "test-token",
		generator.EnvBaseURL:       base,
		generator.EnvTimeout:       "2s",
		generator.EnvApp:           "proven",
		generator.EnvModel:         "meterless/auto",
		generator.EnvRequireNVIDIA: "all",
	}
	for k, v := range extra {
		env[k] = v
	}
	cfg, err := generator.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	return generator.NewLive(cfg, nil)
}

// assertContract checks the method, path, headers and body of one completion request.
func assertContract(t *testing.T, r *http.Request, model, app string) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
		t.Errorf("request %s %s", r.Method, r.URL.Path)
	}
	if r.Header.Get("Authorization") != "Bearer test-token" {
		t.Errorf("authorization header")
	}
	if r.Header.Get("meterless-app") != app {
		t.Errorf("app %s", r.Header.Get("meterless-app"))
	}
	if r.Header.Get("meterless-scope") != "user" {
		t.Errorf("scope %s", r.Header.Get("meterless-scope"))
	}
	var body struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Stream || body.Model != model {
		t.Errorf("body model %q stream %v", body.Model, body.Stream)
	}
}

// replay answers every request with body as an event stream.
func replay(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}
}

// withoutMeterless removes the top-level meterless object from every event of a stream.
func withoutMeterless(t *testing.T, sse string) string {
	t.Helper()
	var b strings.Builder
	removed := false
	for _, line := range strings.Split(sse, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "data:") {
			b.WriteString(line)
			if line != "" || !strings.HasSuffix(b.String(), "\n") {
				b.WriteByte('\n')
			}
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(trim, "data:"))
		if data == "[DONE]" || !strings.Contains(data, `"meterless"`) {
			b.WriteString(line)
			b.WriteByte('\n')
			continue
		}
		var obj map[string]json.RawMessage
		if err := json.Unmarshal([]byte(data), &obj); err != nil {
			t.Fatal(err)
		}
		if _, ok := obj["meterless"]; !ok {
			b.WriteString(line)
			b.WriteByte('\n')
			continue
		}
		delete(obj, "meterless")
		removed = true
		raw, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		b.WriteString("data: ")
		b.Write(raw)
		b.WriteByte('\n')
	}
	if !removed {
		t.Fatal("the fixture has no meterless object")
	}
	return b.String()
}

// streamFixture reads the recorded Nightshift stream from the module's test data.
func streamFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(moduleRoot(t), "testdata", "generator", "nightshift-stream.sse"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// moduleRoot walks up from the test's directory to the module's go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(b), "module github.com/saphan-ai/proven") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
