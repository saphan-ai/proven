package generator_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

func TestThreeClocksAndNoRetryOnTimeout(t *testing.T) {
	for _, tc := range []struct {
		name, reason       string
		headers, heartbeat bool
		options            map[string]string
	}{
		{"headers", "response headers timeout", false, false, map[string]string{generator.EnvHeaderTimeout: "60ms"}},
		{"idle", "stream idle timeout", true, false, map[string]string{generator.EnvIdleTimeout: "60ms"}},
		{"total", "timeout", true, true, map[string]string{generator.EnvTimeout: "180ms", generator.EnvIdleTimeout: "80ms"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				io.Copy(io.Discard, r.Body)
				if tc.headers {
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
				}
				ticker := time.NewTicker(15 * time.Millisecond)
				defer ticker.Stop()
				end := time.NewTimer(time.Second)
				defer end.Stop()
				for {
					select {
					case <-end.C:
						return
					case <-r.Context().Done():
						return
					case <-ticker.C:
						if tc.heartbeat {
							io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"thinking\"}}]}\n\n")
							w.(http.Flusher).Flush()
						}
					}
				}
			}))
			defer srv.Close()
			tc.options[generator.EnvTransportRetries] = "3"
			g := configured(t, srv.URL, tc.options)
			_, err := g.Complete(context.Background(), "planner", "prompt")
			d, ok := service.As(err)
			if !ok || d.Reason != tc.reason || calls.Load() != 1 {
				t.Fatalf("calls=%d err=%v", calls.Load(), err)
			}
			a := g.LastAttempts()
			if len(a) != 1 || a[0].Error != tc.reason {
				t.Fatalf("measurements: %#v", a)
			}
			if tc.heartbeat && (a[0].ReasoningBytes == 0 || a[0].AnswerBytes != 0 || a[0].Reasoning != "") {
				t.Fatalf("reasoning: %#v", a[0])
			}
		})
	}
}

func TestReasoningAndHeartbeatKeepStreamAliveWithoutEnteringAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		for i := 0; i < 6; i++ {
			io.WriteString(w, ": heartbeat\n\ndata: {\"choices\":[{\"delta\":{\"reasoning\":\"thought\"}}]}\n\n")
			w.(http.Flusher).Flush()
			select {
			case <-r.Context().Done():
				return
			case <-time.After(15 * time.Millisecond):
			}
		}
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"{\\\"ok\\\":true}\"},\"finish_reason\":\"stop\"}]}\n\ndata: {\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":20,\"total_tokens\":30,\"completion_tokens_details\":{\"reasoning_tokens\":12}}}\n\ndata: [DONE]\n\n")
	}))
	defer srv.Close()
	g := configured(t, srv.URL, map[string]string{generator.EnvIdleTimeout: "80ms", generator.EnvTraceReasonBytes: "8"})
	text, err := g.Complete(context.Background(), "planner", "prompt")
	if err != nil || text != `{"ok":true}` {
		t.Fatalf("%q %v", text, err)
	}
	a := g.LastAttempts()[0]
	if a.ReasoningBytes != 42 || a.Reasoning != "thoughtt" || !a.ReasoningClipped || a.FinishReason != "stop" || a.FirstByteMS == nil || a.HeadersMS == nil {
		t.Fatalf("%#v", a)
	}
	if a.Usage == nil || *a.Usage.TotalTokens != 30 || *a.Usage.ReasoningTokens != 12 {
		t.Fatalf("usage: %#v", a.Usage)
	}
}

func TestHTTPRetryClasses(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 408, 429, 500, 502, 503, 504, 307} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					w.Header().Set("Retry-After", "0")
					w.Header().Set("Location", "/other")
					w.WriteHeader(status)
					return
				}
				reply(w, "ok")
			}))
			defer srv.Close()
			g := configured(t, srv.URL, map[string]string{generator.EnvRateLimitRetries: "1", generator.EnvTransportRetries: "1", generator.EnvRetryBackoff: "1ms"})
			text, err := g.Complete(context.Background(), "planner", "prompt")
			retry := status == 429 || status == 500 || status == 502 || status == 503
			if retry && (err != nil || text != "ok" || calls.Load() != 2) || !retry && (err == nil || calls.Load() != 1) {
				t.Fatalf("calls=%d text=%q err=%v", calls.Load(), text, err)
			}
			if len(g.LastAttempts()) != int(calls.Load()) {
				t.Fatal("lost retry measurements")
			}
		})
	}
}

func TestRetryAfterCapAndCancellation(t *testing.T) {
	for _, hint := range []string{"3600", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)} {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.Header().Set("Retry-After", hint)
			w.WriteHeader(429)
		}))
		g := configured(t, srv.URL, map[string]string{generator.EnvRateLimitRetries: "3"})
		_, err := g.Complete(context.Background(), "planner", "prompt")
		srv.Close()
		if err == nil || calls.Load() != 1 {
			t.Fatalf("hint=%s calls=%d err=%v", hint, calls.Load(), err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(429)
		time.AfterFunc(20*time.Millisecond, cancel)
	}))
	defer srv.Close()
	g := configured(t, srv.URL, map[string]string{generator.EnvRateLimitRetries: "3"})
	_, err := g.Complete(ctx, "planner", "prompt")
	if !errors.Is(err, context.Canceled) || len(g.LastAttempts()) != 1 {
		t.Fatalf("%v", err)
	}
}

func TestNoRetryAfterPartialStreamOrTokenLimit(t *testing.T) {
	for _, sse := range []string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\ndata: broken\n\n",
		"data: {\"error\":{\"message\":\"private provider error\"}}\n\n",
		"data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"thought\"},\"finish_reason\":\"length\"}]}\n\n",
	} {
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; io.WriteString(w, sse) }))
		g := configured(t, srv.URL, map[string]string{generator.EnvTransportRetries: "3"})
		_, err := g.Complete(context.Background(), "prover", "prompt")
		srv.Close()
		if err == nil || calls != 1 || len(g.LastAttempts()) != 1 {
			t.Fatalf("calls=%d err=%v", calls, err)
		}
	}
}

func TestReserveStopsBeforeStartingModelAndCapsAtTenthOfReview(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; reply(w, "ok") }))
	defer srv.Close()
	g := configured(t, srv.URL, map[string]string{generator.EnvTimeout: "500ms", generator.EnvDeadlineReserve: "60s"})
	ctx, cancel := context.WithTimeout(context.Background(), 450*time.Millisecond)
	defer cancel()
	_, err := g.Complete(ctx, "planner", "prompt")
	if !errors.Is(err, generator.ErrReviewBudget) || calls != 0 || ctx.Err() != nil {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	g = configured(t, srv.URL, map[string]string{generator.EnvTimeout: "100ms", generator.EnvDeadlineReserve: "60s"})
	if _, err := g.Complete(ctx, "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
}
