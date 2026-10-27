package generator

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
)

// secondKey is the key the tests configure. It is long enough for the redactor and never a real one.
const secondKey = "sk-second-test-key-0123456789"

// secondEnv builds an environment lookup from the given values.
func secondEnv(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

// TestNoKeyMeansNoSecondReviewer proves ADR-0037 point 1: with no key the second reviewer is not
// connected and nothing is refused, so a user who connects nothing gets one review as before.
// Mutated field: the key, which is empty.
func TestNoKeyMeansNoSecondReviewer(t *testing.T) {
	_, connected, err := LoadSecond(secondEnv(map[string]string{}))
	if err != nil || connected {
		t.Fatalf("connected=%v err=%v", connected, err)
	}
}

// TestKeyConnectsTheSecondReviewer proves the three settings load into one configuration, and that
// the timeout is the generator's own setting.
func TestKeyConnectsTheSecondReviewer(t *testing.T) {
	got, connected, err := LoadSecond(secondEnv(map[string]string{
		EnvSecondKey: secondKey, EnvSecondBaseURL: "https://api.example.test/v1/", EnvSecondModel: "gpt-test",
		EnvTimeout: "9s",
	}))
	if err != nil || !connected {
		t.Fatalf("connected=%v err=%v", connected, err)
	}
	if got.BaseURL != "https://api.example.test/v1" || got.Model != "gpt-test" || got.Key != secondKey || got.Timeout.Seconds() != 9 {
		t.Fatalf("settings %+v", got)
	}
}

// TestAWrongSecondReviewerValueIsRefusedByName proves ADR-0037 point 2: a wrong value is refused
// under its own variable name, with the allowed form, and the refusal never carries the key.
// Each row mutates exactly one field and leaves the others valid.
func TestAWrongSecondReviewerValueIsRefusedByName(t *testing.T) {
	valid := map[string]string{
		EnvSecondKey: secondKey, EnvSecondBaseURL: "https://api.example.test/v1", EnvSecondModel: "gpt-test",
	}
	for _, tc := range []struct {
		field string
		value string
	}{
		{EnvSecondBaseURL, "ftp://api.example.test"},
		{EnvSecondBaseURL, "https://user:pw@api.example.test"},
		{EnvSecondBaseURL, "https://api.example.test/other"},
		{EnvSecondBaseURL, ""},
		{EnvSecondModel, "has a space"},
		{EnvSecondModel, ""},
		{EnvSecondKey, "line\nbreak" + secondKey},
		{EnvTimeout, "soon"},
	} {
		env := map[string]string{}
		for k, v := range valid {
			env[k] = v
		}
		env[tc.field] = tc.value
		_, connected, err := LoadSecond(secondEnv(env))
		var ce *ConfigError
		if err == nil || !errors.As(err, &ce) || ce.Field != tc.field || connected {
			t.Errorf("%s=%q: connected=%v err=%v", tc.field, tc.value, connected, err)
			continue
		}
		if strings.Contains(err.Error(), secondKey) {
			t.Errorf("%s: the refusal carries the key: %v", tc.field, err)
		}
	}
}

// TestAWrongAddressIsRefusedBeforeTheKeyIsAdded proves a typo in the address is found while the
// reviewer is still switched off, so adding the key later does not start a failing review.
func TestAWrongAddressIsRefusedBeforeTheKeyIsAdded(t *testing.T) {
	_, _, err := LoadSecond(secondEnv(map[string]string{EnvSecondBaseURL: "not a url"}))
	var ce *ConfigError
	if !errors.As(err, &ce) || ce.Field != EnvSecondBaseURL {
		t.Fatalf("err=%v", err)
	}
}

// fakeAPI is an OpenAI-compatible server that needs no port: it is the transport of the client under
// test. It records the last request and answers with one streamed event.
type fakeAPI struct {
	status  int
	answer  string
	request *http.Request
	body    chatRequest
}

// RoundTrip records the request and returns the recorded answer.
func (f *fakeAPI) RoundTrip(r *http.Request) (*http.Response, error) {
	f.request = r
	raw, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(raw, &f.body)
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	event, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": f.answer}}}})
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: " + string(event) + "\n\ndata: [DONE]\n\n")),
		Request:    r,
	}, nil
}

// secondOver builds a Second whose transport is api.
func secondOver(api *fakeAPI) *Second {
	return NewSecond(SecondSettings{
		BaseURL: "https://api.example.test/v1", Key: secondKey, Model: "gpt-test", Timeout: 5e9,
	}, &http.Client{Transport: api})
}

// TestSecondSendsOnlyTheKeyAndThePrompt proves the request is a plain OpenAI-compatible chat
// completion: the configured model, one user message, the key as a bearer token, and none of
// Nightshift's routing headers, which would tell a third party our application name.
func TestSecondSendsOnlyTheKeyAndThePrompt(t *testing.T) {
	api := &fakeAPI{answer: "an answer"}
	got, err := secondOver(api).Complete(context.Background(), "planner", "the prompt")
	if err != nil || got != "an answer" {
		t.Fatalf("got %q err %v", got, err)
	}
	if api.request.URL.String() != "https://api.example.test/v1/chat/completions" {
		t.Fatalf("url %s", api.request.URL)
	}
	if api.request.Header.Get("Authorization") != "Bearer "+secondKey {
		t.Fatal("the key was not sent as a bearer token")
	}
	for name := range api.request.Header {
		if strings.HasPrefix(strings.ToLower(name), "meterless") {
			t.Fatalf("a Nightshift header went to the second reviewer: %s", name)
		}
	}
	if api.body.Model != "gpt-test" || len(api.body.Messages) != 1 || api.body.Messages[0].Content != "the prompt" {
		t.Fatalf("body %+v", api.body)
	}
}

// TestSecondNamesTheConfiguredModelAndNeedsNoNVIDIA proves the second reviewer answers for any
// model, not only nvidia/*, and is named by its setting because no receipt is read.
func TestSecondNamesTheConfiguredModelAndNeedsNoNVIDIA(t *testing.T) {
	s := secondOver(&fakeAPI{answer: "ok"})
	if _, err := s.Complete(context.Background(), "skeptic", "p"); err != nil {
		t.Fatal(err)
	}
	use, ok := s.LastUse()
	if !ok || use.Model != "gpt-test" || use.FromReceipt {
		t.Fatalf("use %+v ok %v", use, ok)
	}
}

// TestSecondFailureNamesTheSecondReviewerAndNotTheKey proves a refused call is a named failure of
// the second reviewer, never of Nightshift, and that its text carries no key.
// Mutated field: the HTTP status.
func TestSecondFailureNamesTheSecondReviewerAndNotTheKey(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusInternalServerError} {
		_, err := secondOver(&fakeAPI{status: status}).Complete(context.Background(), "planner", "p")
		var d *service.Down
		if !errors.As(err, &d) || d.Service != SecondService {
			t.Fatalf("status %d: err %v", status, err)
		}
		if strings.Contains(err.Error(), secondKey) {
			t.Fatalf("status %d: the error carries the key", status)
		}
	}
}

// The generator's thinking budget bounds the second reviewer's calls too: LoadSecond reads
// PROVEN_GENERATOR_THINKING_BUDGET and the request carries it; unset, the body has no such field, and a
// value that is not a count is refused by name. Mutated field: PROVEN_GENERATOR_THINKING_BUDGET.
func TestTheSecondReviewerSendsTheGeneratorsThinkingBudget(t *testing.T) {
	base := map[string]string{EnvSecondKey: secondKey, EnvSecondBaseURL: "https://api.example.test/v1", EnvSecondModel: "gpt-test"}
	for setting, want := range map[string]int{"512": 512, "": 0} {
		env := map[string]string{EnvThinkingBudget: setting}
		for k, v := range base {
			env[k] = v
		}
		s, connected, err := LoadSecond(secondEnv(env))
		if err != nil || !connected {
			t.Fatalf("connected %v err %v", connected, err)
		}
		api := &fakeAPI{answer: "an answer"}
		if _, err := NewSecond(s, &http.Client{Transport: api}).Complete(context.Background(), "planner", "p"); err != nil {
			t.Fatal(err)
		}
		if api.body.ThinkingBudget != want {
			t.Fatalf("setting %q: thinking_budget %d", setting, api.body.ThinkingBudget)
		}
	}
	env := map[string]string{EnvThinkingBudget: "lots"}
	for k, v := range base {
		env[k] = v
	}
	_, _, err := LoadSecond(secondEnv(env))
	var cfg *ConfigError
	if !errors.As(err, &cfg) || cfg.Field != EnvThinkingBudget {
		t.Fatalf("a bad budget was not refused by name: %v", err)
	}
}
