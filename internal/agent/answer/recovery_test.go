package answer_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/generator"
)

func TestEmptyAndSchemaRecoveryShareOneBudget(t *testing.T) {
	for _, replies := range [][]string{{"", "invalid"}, {"invalid", ""}, {"invalid", `{"sql":"SELECT 1"}`}} {
		var calls int
		var prompts []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Messages []struct {
					Content string `json:"content"`
				} `json:"messages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Error(err)
			}
			prompts = append(prompts, req.Messages[0].Content)
			content := "unexpected third call"
			if calls < len(replies) {
				content = replies[calls]
			}
			calls++
			b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": content}}}})
			io.WriteString(w, "data: "+string(b)+"\n\ndata: [DONE]\n\n")
		}))
		cfg, err := generator.Load(func(k string) string {
			return map[string]string{generator.EnvMode: "openai", generator.EnvModel: "test/model", generator.EnvBaseURL: srv.URL}[k]
		})
		if err != nil {
			t.Fatal(err)
		}
		var dst struct {
			SQL string `json:"sql"`
		}
		err = answer.Read(context.Background(), generator.NewLive(cfg, nil), "investigator", "ORIGINAL_LARGE_CONTEXT", answer.Spec{Shape: `{"sql":string}`}, &dst)
		srv.Close()
		wantOK := strings.HasPrefix(replies[1], "{")
		if calls != 2 || (err == nil) != wantOK {
			t.Fatalf("replies=%q calls=%d err=%v", replies, calls, err)
		}
		if replies[0] != "" && strings.Contains(prompts[1], "ORIGINAL_LARGE_CONTEXT") {
			t.Fatal("repair repeated original context")
		}
	}
}

func TestDecoderRejectsVendorEnvelopeButKeepsTolerantAnswers(t *testing.T) {
	for _, tc := range []struct {
		body string
		ok   bool
	}{
		{`{"type":"result","result":"{\"sql\":\"SELECT 1\"}","usage":{}}`, false},
		{`{"assistant":{"message":{"content":"answer"}}}`, false},
		{`null`, false}, {`{}`, true},
		{`{"SQL":"SELECT 1","explanation":"extra keys are tolerated",}`, true},
	} {
		var dst struct {
			SQL string `json:"sql"`
		}
		if err := answer.Decode(tc.body, &dst); (err == nil) != tc.ok {
			t.Fatalf("%s: %v", tc.body, err)
		}
	}
}
