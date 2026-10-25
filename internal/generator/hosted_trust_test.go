package generator_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
)

// These tests cover what nvidia mode trusts and where it sends the key
// (ADR-0031 point 3). Every answer is served through the client's transport, so
// no port is opened.

// Proves nvidia mode refuses a base that is not https, by the name of its own
// setting, and never echoes the key. Mutated field: PROVEN_NVIDIA_BASE_URL,
// set to an http address.
func TestNVIDIABaseMustBeHTTPSAndIsRefusedByName(t *testing.T) {
	_, err := generator.Load(func(k string) string {
		return map[string]string{
			generator.EnvMode:          "nvidia",
			generator.EnvKey:           hostedKey,
			generator.EnvNVIDIABaseURL: "http://example.test/v1",
		}[k]
	})
	ce := configErr(t, err)
	if ce.Field != generator.EnvNVIDIABaseURL || !strings.Contains(err.Error(), "https") || strings.Contains(err.Error(), hostedKey) {
		t.Fatalf("err %v", err)
	}
}

// Proves nvidia mode's own base setting is used, and that Nightshift mode still
// reads only PROVEN_GENERATOR_BASE_URL and allows http there. Mutated field:
// PROVEN_NVIDIA_BASE_URL in nvidia mode; PROVEN_GENERATOR_BASE_URL in nightshift mode.
func TestEachModeReadsItsOwnBase(t *testing.T) {
	cfg, err := generator.Load(func(k string) string {
		return map[string]string{
			generator.EnvMode:          "nvidia",
			generator.EnvKey:           hostedKey,
			generator.EnvNVIDIABaseURL: "https://example.test/v1",
		}[k]
	})
	if err != nil || cfg.BaseURL != "https://example.test/v1" || cfg.BaseHost() != "example.test" {
		t.Fatalf("base %q err %v", cfg.BaseURL, err)
	}
	cfg, err = generator.Load(func(k string) string {
		return map[string]string{
			generator.EnvToken:         "test-token",
			generator.EnvBaseURL:       "http://127.0.0.1:8088",
			generator.EnvNVIDIABaseURL: "https://example.test/v1",
		}[k]
	})
	if err != nil || cfg.BaseURL != "http://127.0.0.1:8088" {
		t.Fatalf("nightshift base %q err %v", cfg.BaseURL, err)
	}
}

// Proves a hosted answer names the host it came from as its route, so the
// trace's model event says where the trusted model field was read.
func TestNVIDIAUseNamesTheHost(t *testing.T) {
	live := hostedAt(t, func(*http.Request) *http.Response {
		return answer(http.StatusOK, hostedStream("nvidia/a"))
	}, nil)
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	if u, _ := live.LastUse(); u.Route != "integrate.api.nvidia.com" {
		t.Fatalf("route %q", u.Route)
	}
}

// Proves a meterless object in a hosted stream is ignored: the answering model
// is the chunk's own field, the receipt stays "no receipt", and the route is
// the host. Mutated field: a meterless object with another routed model and a receipt.
func TestNVIDIAIgnoresAMeterlessObject(t *testing.T) {
	body := `data: {"model":"nvidia/own","choices":[{"delta":{"content":"ready"}}],"meterless":{"receipt_id":"rcpt-1","routed_model":"openai/other","winning_route":{"provider":"p"}}}` + "\n\n" + "data: [DONE]\n\n"
	live := hostedAt(t, func(*http.Request) *http.Response { return answer(http.StatusOK, body) }, nil)
	if _, err := live.Complete(context.Background(), "planner", "prompt"); err != nil {
		t.Fatal(err)
	}
	u, _ := live.LastUse()
	if u.Model != "nvidia/own" || u.ReceiptID != generator.NoReceipt || u.FromReceipt || u.Route != "integrate.api.nvidia.com" {
		t.Fatalf("use %#v", u)
	}
}

// Proves the first chunk that names a model decides who answered: a later chunk
// cannot replace it. Mutated field: the model field of the second chunk.
func TestNVIDIAFirstChunkModelWins(t *testing.T) {
	body := `data: {"model":"nvidia/a","choices":[{"delta":{"content":"rea"}}]}` + "\n\n" +
		`data: {"model":"openai/b","choices":[{"delta":{"content":"dy"}}]}` + "\n\n" + "data: [DONE]\n\n"
	live := hostedAt(t, func(*http.Request) *http.Response { return answer(http.StatusOK, body) }, nil)
	text, err := live.Complete(context.Background(), "planner", "prompt")
	if err != nil || text != "ready" {
		t.Fatalf("text %q err %v", text, err)
	}
	if u, _ := live.LastUse(); u.Model != "nvidia/a" {
		t.Fatalf("model %q", u.Model)
	}
}

// nightshiftAt builds a nightshift-mode generator over a transport.
func nightshiftAt(t *testing.T, rt roundTrip, extra map[string]string) *generator.Live {
	t.Helper()
	env := map[string]string{generator.EnvToken: "a-nightshift-token"}
	for k, v := range extra {
		env[k] = v
	}
	cfg, err := generator.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	return generator.NewLive(cfg, &http.Client{Transport: rt})
}

// Proves a 403 is a rejected key only in nvidia mode: in nightshift mode it stays
// "http 403", as it was before nvidia mode existed. Mutated field: the response status.
func TestNightshiftKeeps403AsAnHTTPStatus(t *testing.T) {
	live := nightshiftAt(t, func(*http.Request) *http.Response { return answer(http.StatusForbidden, "") }, nil)
	_, err := live.Complete(context.Background(), "planner", "prompt")
	d, ok := service.As(err)
	if !ok || d.Service != "nightshift" || d.Reason != "http 403" {
		t.Fatalf("err %v", err)
	}
}

// stallThenBad is a body that gives nothing until the request is cancelled and
// then yields one event that is not JSON.
type stallThenBad struct {
	ctx  context.Context
	done bool
}

// Read blocks until the request's context ends, then returns one bad event.
func (b *stallThenBad) Read(p []byte) (int, error) {
	if b.done {
		return 0, io.EOF
	}
	<-b.ctx.Done()
	b.done = true
	return copy(p, "data: not json\n\n"), nil
}

// Close ends the body.
func (b *stallThenBad) Close() error { return nil }

// Proves the order in which a failed stream is named: an event that is not JSON
// and that arrives after this call's deadline is a timeout, not a bad stream,
// because the deadline is the cause. Nightshift and nvidia modes alike.
// Mutated field: PROVEN_GENERATOR_TIMEOUT, short, against a stalled body.
func TestABadEventAfterTheDeadlineIsATimeout(t *testing.T) {
	serve := func(r *http.Request) *http.Response {
		return &http.Response{StatusCode: http.StatusOK, Body: &stallThenBad{ctx: r.Context()}}
	}
	short := map[string]string{generator.EnvTimeout: "50ms"}
	for name, live := range map[string]*generator.Live{
		"nightshift": nightshiftAt(t, serve, short),
		"nvidia":     hostedAt(t, serve, short),
	} {
		_, err := live.Complete(context.Background(), "planner", "prompt")
		if d, ok := service.As(err); !ok || d.Reason != "timeout" {
			t.Errorf("%s: err %v", name, err)
		}
	}
}
