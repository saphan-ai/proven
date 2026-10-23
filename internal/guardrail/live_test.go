package guardrail_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/service"
)

// The server backend posts the measured request with the key and no
// Nightshift headers, and reads the verdict it gets back.
func TestServerPostsTheMeasuredRequest(t *testing.T) {
	const key = "test-key"
	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer "+key {
			t.Error("authorization missing")
		}
		if r.Header.Get("meterless-app") != "" || r.Header.Get("meterless-scope") != "" {
			t.Error("nightshift headers were sent")
		}
		var err error
		raw, err = io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, verdictJSON("User Safety: unsafe\nSafety Categories: Criminal Planning/Confessions, Other\n"))
	}))
	defer srv.Close()
	rail := liveAt(t, guardrail.ModeServer, srv.URL, key, nil)
	if _, ok := rail.(*guardrail.Server); !ok {
		t.Fatalf("%T", rail)
	}
	v, err := rail.Check(context.Background(), "sort a list", "")
	if err != nil {
		t.Fatal(err)
	}
	if !v.Unsafe() || v.Latency < 0 {
		t.Fatalf("%#v", v)
	}
	if len(v.Categories) != 2 || v.Categories[0] != "Criminal Planning/Confessions" {
		t.Fatalf("categories %#v", v.Categories)
	}
	for _, want := range []string{
		`"model":"nemotron-3.5-content-safety"`,
		`"temperature":0`,
		`"stream":false`,
		`"enable_thinking":false`,
		`"request_categories":"/categories"`,
		`"role":"user"`,
		`"content":"sort a list"`,
	} {
		if !bytes.Contains(raw, []byte(want)) {
			t.Fatalf("body missing %s\n%s", want, raw)
		}
	}
	if bytes.Contains(raw, []byte(`"role":"assistant"`)) {
		t.Fatal("a prompt check sent an assistant turn")
	}
	if bytes.Contains(raw, []byte(key)) {
		t.Fatal("the key was in the body")
	}
}

// An answer check sends the answer as the assistant turn after the prompt.
func TestServerSendsTheAnswerAsAnAssistantTurn(t *testing.T) {
	var raw []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		raw, err = io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			t.Error(err)
		}
		_, _ = io.WriteString(w, verdictJSON("User Safety: safe\nResponse Safety: unsafe\nSafety Categories: Malware, PII/Privacy\n"))
	}))
	defer srv.Close()
	rail := liveAt(t, guardrail.ModeServer, srv.URL, "test-key", nil)
	v, err := rail.Check(context.Background(), "prompt", "the answer")
	if err != nil {
		t.Fatal(err)
	}
	if v.User != "safe" || v.Response != "unsafe" || !v.Unsafe() {
		t.Fatalf("%#v", v)
	}
	if !bytes.Contains(raw, []byte(`"role":"assistant"`)) || !bytes.Contains(raw, []byte(`"content":"the answer"`)) {
		t.Fatalf("body %s", raw)
	}
}

// The nvidia mode builds the NVIDIA backend and sends the hosted model id.
func TestNVIDIAUsesItsModel(t *testing.T) {
	var model string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		model = body.Model
		_, _ = io.WriteString(w, verdictJSON("User Safety: safe\n"))
	}))
	defer srv.Close()
	rail := liveAt(t, guardrail.ModeNVIDIA, srv.URL, "test-key", nil)
	if _, ok := rail.(*guardrail.NVIDIA); !ok {
		t.Fatalf("%T", rail)
	}
	if _, err := rail.Check(context.Background(), "hi", ""); err != nil {
		t.Fatal(err)
	}
	if model != "nvidia/nemotron-3.5-content-safety" {
		t.Fatalf("model %s", model)
	}
}

// A guardrail that cannot be reached refuses the step as unreachable.
// Mutated field: the listening socket. The server is closed before the check.
func TestClosedServerIsRefusedByName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed server accepted a call")
	}))
	url := srv.URL
	srv.Close()
	rail := liveAt(t, guardrail.ModeServer, url, "super-secret", map[string]string{guardrail.EnvTimeout: "1s"})
	_, err := rail.Check(context.Background(), "prompt", "")
	d, ok := service.As(err)
	if !ok || d.Service != guardrail.Service || d.Reason != "unreachable" {
		t.Fatalf("%v", err)
	}
	if err.Error() != "guardrail: unreachable" || strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("error %q", err.Error())
	}
}

// A guardrail that never answers is named as a timeout within its limit.
// Mutated field: the server, which accepts the connection and sends no reply.
func TestSilentServerIsATimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	accepted := make(chan struct{})
	go func() {
		conn, aerr := ln.Accept()
		if aerr != nil {
			return
		}
		close(accepted)
		_, _ = io.Copy(io.Discard, conn)
		_ = conn.Close()
	}()
	rail := liveAt(t, guardrail.ModeServer, "http://"+ln.Addr().String(), "test-key", map[string]string{guardrail.EnvTimeout: "200ms"})
	start := time.Now()
	_, err = rail.Check(context.Background(), "prompt", "")
	if time.Since(start) > 2*time.Second {
		t.Fatal("the timeout did not return")
	}
	d, ok := service.As(err)
	if !ok || d.Service != guardrail.Service || d.Reason != "timeout" {
		t.Fatalf("%v", err)
	}
	if err.Error() != "guardrail: timeout" {
		t.Fatalf("error %q", err.Error())
	}
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("the server did not accept")
	}
}

// A 401 is named as unauthorized, and the key appears in no error text.
// Mutated field: the HTTP status, set to 401.
func TestUnauthorizedDoesNotPrintTheKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	rail := liveAt(t, guardrail.ModeServer, srv.URL, "super-secret", nil)
	_, err := rail.Check(context.Background(), "prompt", "")
	d, ok := service.As(err)
	if !ok || d.Reason != "unauthorized" {
		t.Fatalf("%v", err)
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatal("the refusal printed the key")
	}
}

// An answer that is not the verdict lines is refused, not read as safe.
// Mutated field: the completion body, which is not the verdict lines.
func TestABadVerdictIsRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, verdictJSON("I am not sure"))
	}))
	defer srv.Close()
	rail := liveAt(t, guardrail.ModeServer, srv.URL, "test-key", nil)
	_, err := rail.Check(context.Background(), "prompt", "")
	d, ok := service.As(err)
	if !ok || d.Reason != "bad verdict" {
		t.Fatalf("%v", err)
	}
}

// A redirect is not followed, so the key is not sent on to another address.
func TestRedirectIsNotFollowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			http.Redirect(w, r, "/stolen", http.StatusFound)
			return
		}
		t.Error("followed the redirect")
	}))
	defer srv.Close()
	rail := liveAt(t, guardrail.ModeServer, srv.URL, "test-key", nil)
	_, err := rail.Check(context.Background(), "prompt", "")
	d, ok := service.As(err)
	if !ok || d.Reason != "http 302" {
		t.Fatalf("%v", err)
	}
}

// liveAt builds a live adapter through Load and New, as a review does.
func liveAt(t *testing.T, mode, base, key string, extra map[string]string) guardrail.Adapter {
	t.Helper()
	s, err := guardrail.Load(func(k string) string {
		switch k {
		case guardrail.EnvMode:
			return mode
		case guardrail.EnvBaseURL:
			return base
		case guardrail.EnvKey:
			return key
		default:
			if extra != nil {
				return extra[k]
			}
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	rail, err := guardrail.New(s, nil)
	if err != nil {
		t.Fatal(err)
	}
	return rail
}

// verdictJSON wraps verdict text in a chat completion body.
func verdictJSON(text string) string {
	raw, err := json.Marshal(map[string]any{
		"choices": []any{
			map[string]any{"message": map[string]any{"role": "assistant", "content": text}},
		},
	})
	if err != nil {
		panic(err)
	}
	return string(raw)
}
