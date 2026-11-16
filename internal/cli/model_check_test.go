package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelCheckSupportsCommandWithoutInventingHTTPMeasurements(t *testing.T) {
	for _, tc := range []struct {
		response string
		want     int
	}{
		{`{"ok":true}`, 0}, {`{"ok":false}`, 3}, {`not JSON`, 3},
	} {
		t.Run(tc.response, func(t *testing.T) {
			dir := t.TempDir()
			command := filepath.Join(dir, "model")
			// A fixed local fixture; no vendor account or runtime is used.
			if err := os.WriteFile(command, []byte("#!/bin/sh\nprintf '%s' '"+tc.response+"'\n"), 0700); err != nil {
				t.Fatal(err)
			}
			argv, _ := json.Marshal([]string{command})
			env := map[string]string{"PROVEN_GENERATOR_MODE": "command", "PROVEN_GENERATOR_MODEL": "chosen", "PROVEN_GENERATOR_COMMAND": string(argv)}
			var out, log bytes.Buffer
			code := Run([]string{"model", "check"}, func(k string) string { return env[k] }, &out, &log, nil)
			var result struct {
				OK       bool  `json:"ok"`
				Attempts []any `json:"attempts"`
			}
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatalf("%v: %s; %s", err, &out, &log)
			}
			if code != tc.want || result.OK != (tc.want == 0) || len(result.Attempts) != 0 {
				t.Fatalf("code=%d output=%s log=%s", code, &out, &log)
			}
		})
	}
}

func TestModelCheckMeasuresOneRequestEvenWhenRetriesConfigured(t *testing.T) {
	for _, status := range []int{200, 400, 429} {
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(status)
			fmt.Fprint(w, "data: {\"model\":\"secret-api-key-value\",\"choices\":[{\"delta\":{\"content\":\"{\\\"ok\\\":true}\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
		}))
		env := map[string]string{"PROVEN_GENERATOR_MODE": "openai", "PROVEN_GENERATOR_MODEL": "model", "PROVEN_GENERATOR_BASE_URL": srv.URL,
			"PROVEN_GENERATOR_KEY": "secret-api-key-value", "PROVEN_GENERATOR_RATE_LIMIT_RETRIES": "3", "PROVEN_GENERATOR_TRANSPORT_RETRIES": "3"}
		var out, log bytes.Buffer
		code := Run([]string{"model", "check"}, func(k string) string { return env[k] }, &out, &log, nil)
		srv.Close()
		var result struct {
			OK       bool  `json:"ok"`
			Attempts []any `json:"attempts"`
		}
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if calls != 1 || result.OK != (status == 200) || len(result.Attempts) != 1 || (code == 0) != (status == 200) || strings.Contains(out.String(), "secret-api-key-value") {
			t.Fatalf("calls=%d code=%d output=%s log=%s", calls, code, &out, &log)
		}
	}
}
