package generator_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
)

func configured(t *testing.T, base string, overrides map[string]string) *generator.Live {
	t.Helper()
	env := map[string]string{generator.EnvMode: "openai", generator.EnvBaseURL: base, generator.EnvModel: "test/model", generator.EnvTimeout: "2s"}
	for k, v := range overrides {
		env[k] = v
	}
	cfg, err := generator.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	return generator.NewLive(cfg, nil)
}

func reply(w http.ResponseWriter, content string) {
	b, _ := json.Marshal(map[string]any{"model": "test/model", "choices": []any{map[string]any{"delta": map[string]string{"content": content}, "finish_reason": "stop"}}})
	io.WriteString(w, "data: "+string(b)+"\n\ndata: [DONE]\n\n")
}

func TestOptionalRequestFieldsAndAgentOverrides(t *testing.T) {
	var bodies []map[string]json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			t.Error(err)
		}
		bodies = append(bodies, b)
		reply(w, "{}")
	}))
	defer srv.Close()
	gen := configured(t, srv.URL, nil)
	if _, err := gen.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"max_tokens", "max_completion_tokens", "stream_options", "thinking_budget", "chat_template_kwargs"} {
		if _, ok := bodies[0][k]; ok {
			t.Fatalf("optional field %s sent without configuration", k)
		}
	}
	gen = configured(t, srv.URL, map[string]string{
		generator.EnvMaxTokens: "2048", generator.EnvMaxTokensField: "max_completion_tokens",
		generator.EnvAgentMaxTokens: `{"prover":8192,"skeptic":0}`, generator.EnvIncludeUsage: "true",
		generator.EnvExtraBody:      `{"temperature":0,"chat_template_kwargs":{"enable_thinking":true}}`,
		generator.EnvAgentExtraBody: `{"prover":{"chat_template_kwargs":{"enable_thinking":false}}}`,
	})
	for _, a := range []string{"planner", "prover", "skeptic"} {
		if _, err := gen.Complete(context.Background(), a, "prompt"); err != nil {
			t.Fatal(err)
		}
	}
	if string(bodies[1]["max_completion_tokens"]) != "2048" || string(bodies[2]["max_completion_tokens"]) != "8192" || bodies[3]["max_completion_tokens"] != nil {
		t.Fatalf("token limits: %#v", bodies)
	}
	if string(bodies[2]["chat_template_kwargs"]) != `{"enable_thinking":false}` || string(bodies[1]["stream_options"]) != `{"include_usage":true}` {
		t.Fatalf("body: %#v", bodies[2])
	}
}

func TestCallOptionValidationDoesNotEchoExtraBody(t *testing.T) {
	for _, entry := range []struct{ k, v string }{
		{generator.EnvExtraBody, `{"model":"private-secret-value"}`},
		{generator.EnvExtraBody, `{"Stream":false}`},
		{generator.EnvExtraBody, `[]`},
		{generator.EnvExtraBody, `null`},
		{generator.EnvExtraBody, `not-json-private-secret-value`},
		{generator.EnvAgentExtraBody, `{"planner":{"messages":[]}}`},
		{generator.EnvAgentExtraBody, `{"unknown":{}}`},
		{generator.EnvMaxTokens, "-1"}, {generator.EnvAgentMaxTokens, `{"planner":-1}`},
		{generator.EnvHeaderTimeout, "-1s"}, {generator.EnvIncludeUsage, "yes"},
		{generator.EnvRecoveryRetries, "9"}, {generator.EnvMaxTokensField, "tokens"},
	} {
		t.Run(entry.k+entry.v, func(t *testing.T) {
			_, err := generator.Load(func(k string) string {
				if k == entry.k {
					return entry.v
				}
				return ""
			})
			if err == nil || !strings.Contains(err.Error(), entry.k) || strings.Contains(err.Error(), "private-secret-value") {
				t.Fatalf("error: %v", err)
			}
		})
	}
}

func TestSecondReviewerCanOverrideEveryProviderControl(t *testing.T) {
	env := map[string]string{
		generator.EnvSecondBaseURL: "http://localhost:9999/v1", generator.EnvSecondModel: "other/model",
		generator.EnvTimeout: "300s", generator.EnvMaxTokens: "8192", generator.EnvExtraBody: `{"vendor_primary":true}`,
		generator.EnvSecondTimeout: "90s", "PROVEN_SECOND_REVIEWER_MAX_TOKENS": "1024",
		"PROVEN_SECOND_REVIEWER_EXTRA_BODY": `{}`, "PROVEN_SECOND_REVIEWER_IDLE_TIMEOUT": "10s",
	}
	s, ok, err := generator.LoadSecond(func(k string) string { return env[k] })
	if err != nil || !ok {
		t.Fatalf("%v %v", ok, err)
	}
	if s.Timeout != 90*time.Second || s.Options.MaxTokens != 1024 || s.Options.IdleTimeout != 10*time.Second || len(s.Options.ExtraBody) != 0 {
		t.Fatalf("settings: %#v %#v", s, s.Options)
	}
	env["PROVEN_SECOND_REVIEWER_EXTRA_BODY"] = `{"model":"wrong"}`
	_, _, err = generator.LoadSecond(func(k string) string { return env[k] })
	if err == nil || !strings.Contains(err.Error(), "PROVEN_SECOND_REVIEWER_EXTRA_BODY") {
		t.Fatalf("%v", err)
	}
}
