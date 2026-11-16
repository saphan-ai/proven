package generator

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNeutralProviderNeedsNoSponsorHeadersOrReceipt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("meterless-app") != "" {
			t.Error("unexpected credentials or sponsor header")
		}
		fmt.Fprint(w, "data: {\"id\":\"response-1\",\"model\":\"client-model\",\"choices\":[{\"delta\":{\"content\":\"answer\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer srv.Close()
	s, err := Load(func(k string) string {
		switch k {
		case EnvMode:
			return ModeOpenAI
		case EnvModel:
			return "client-model"
		case EnvBaseURL:
			return srv.URL
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	g := NewLive(s, nil)
	answer, err := g.Complete(context.Background(), "planner", "prompt")
	if err != nil || answer != "answer" {
		t.Fatalf("%q %v", answer, err)
	}
	use, ok := g.LastUse()
	if !ok || use.Model != "client-model" || use.FromReceipt {
		t.Fatalf("%+v", use)
	}
}

func TestNoProviderConfigurationDisablesModels(t *testing.T) {
	s, err := Load(nil)
	if err != nil || s.Mode != ModeOff {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestCommandTransportPassesPromptAndCancels(t *testing.T) {
	cat := NewCommand(Settings{Mode: ModeCommand, Command: []string{"/bin/cat"}, Model: "client-model", Timeout: time.Second})
	input := "prompt; $(printf should-not-execute)"
	if answer, err := cat.Complete(context.Background(), "planner", input); err != nil || answer != input {
		t.Fatalf("stdin changed: %q %v", answer, err)
	}
	// A text-only executable exercises argv substitution without a shell or account.
	g := NewCommand(Settings{Mode: ModeCommand, Command: []string{"/usr/bin/printf", "answer-%s", "{model}"}, Model: "client-model", Timeout: time.Second})
	answer, err := g.Complete(context.Background(), "planner", "untrusted prompt; $(false)")
	if err != nil || answer != "answer-client-model" {
		t.Fatalf("%q %v", answer, err)
	}
	g = NewCommand(Settings{Mode: ModeCommand, Command: []string{"/bin/sleep", "5"}, Timeout: 20 * time.Millisecond})
	start := time.Now()
	_, err = g.Complete(context.Background(), "planner", "prompt")
	if err == nil || !strings.Contains(err.Error(), "timeout") || time.Since(start) > time.Second {
		t.Fatalf("cancellation: %v", err)
	}
}
