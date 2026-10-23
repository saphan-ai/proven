package review

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/osv"
)

// dialRecord is the probe of ADR-0034 point 3. It stands where the network would be: it records every
// address asked of it and answers the three local services from memory, so no socket is opened and
// nothing can leave the host. It refuses an address that is not allowed, and records that too.
type dialRecord struct {
	mu      sync.Mutex
	allowed map[string]bool
	dialed  []string
	refused []string
}

func (d *dialRecord) dial(_ context.Context, _, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	d.mu.Lock()
	defer d.mu.Unlock()
	if err != nil || !d.allowed[host] {
		d.refused = append(d.refused, addr)
		return nil, io.ErrClosedPipe
	}
	d.dialed = append(d.dialed, addr)
	client, server := net.Pipe()
	go serveMemory(server, memoryService(port))
	return client, nil
}

// memoryService is the answer of the local service at one port: Nightshift 8088, the safety model 8081,
// Zetaris 9.
func memoryService(port string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch port {
		case "8088":
			w.Header().Set("Content-Type", "text/event-stream")
			chunk, _ := json.Marshal(map[string]any{
				"model":     "meterless/auto",
				"choices":   []any{map[string]any{"index": 0, "delta": map[string]any{"content": "[]"}}},
				"meterless": map[string]any{"routed_model": "nvidia/nemotron-3.5-super-vl-120b-a12b", "receipt_id": "rcpt_air_1"},
			})
			_, _ = io.WriteString(w, "data: "+string(chunk)+"\n\ndata: [DONE]\n")
		case "8081":
			verdict, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": "User Safety: safe\nResponse Safety: safe\n"}}}})
			_, _ = w.Write(verdict)
		default:
			// Zetaris answers the dependency check with the OSV columns and no advisory.
			table, _ := json.Marshal(map[string]any{"headers": osv.Columns, "data": [][]string{}})
			_, _ = w.Write(table)
		}
	})
}

// serveMemory answers the requests that arrive on conn until it closes.
func serveMemory(conn net.Conn, h http.Handler) {
	defer conn.Close()
	br := bufio.NewReader(conn)
	for {
		req, err := http.ReadRequest(br)
		if err != nil {
			return
		}
		_, _ = io.Copy(io.Discard, req.Body)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		// The length is set so the client knows where the answer ends and the connection stays open.
		resp := rec.Result()
		resp.ContentLength = int64(rec.Body.Len())
		if err := resp.Write(conn); err != nil {
			return
		}
	}
}

// TestAirGappedReviewCompletesAndNoConnectionLeavesTheAllowedHosts proves ADR-0034 point 3. With
// PROVEN_NETWORK=local a live review of the discount fixture completes, every dial it made went to a
// host the setting allows, and the dialer refused nothing because nothing asked for another host. Then a
// probe asks the same clients for a public address that no setting named, as a redirect or a later
// address would, and the dial is refused before it reaches the network.
// Mutated field: the dial check of network.Policy; with it removed the probe's request reaches the recorder.
func TestAirGappedReviewCompletesAndNoConnectionLeavesTheAllowedHosts(t *testing.T) {
	// An empty PATH finds no container runtime, so the Prover is off and only the outbound calls are measured.
	t.Setenv("PATH", t.TempDir())
	rec := &dialRecord{allowed: map[string]bool{"127.0.0.1": true, "nightshift.corp.example": true}}
	env := map[string]string{
		"PROVEN_NETWORK":            "local",
		"PROVEN_NETWORK_ALLOW":      "nightshift.corp.example",
		"METERLESS_RUNTIME_TOKEN":   "test-token",
		"PROVEN_GENERATOR_BASE_URL": "http://nightshift.corp.example:8088",
		"PROVEN_GUARDRAIL_BASE_URL": "http://127.0.0.1:8081",
		"PROVEN_GUARDRAIL_KEY":      "guard-key",
		"ZETARIS_REST_URL":          "http://127.0.0.1:9",
		"ZETARIS_API_KEY":           "zetaris-key",
		"ZETARIS_ORG_ID":            "1",
		"PROVEN_GENERATOR_TIMEOUT":  "5s",
	}
	lookup := func(k string) string { return env[k] }
	rev, err := Open(Options{LogDir: t.TempDir(), Env: lookup, dial: rec.dial})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := rev.Review(ctx, Request{Target: filepath.Join("..", "..", "testdata", "fixture", "discount")})
	if err != nil || res.Partial {
		t.Fatalf("review: partial %v err %v", res.Partial, err)
	}
	for _, s := range res.FailedServices {
		if s.Service == "nightshift" || s.Service == "guardrail" {
			t.Fatalf("a local service failed: %#v (dialed %v refused %v)", res.FailedServices, rec.dialed, rec.refused)
		}
	}
	rec.mu.Lock()
	dialed := append([]string(nil), rec.dialed...)
	refused := append([]string(nil), rec.refused...)
	rec.mu.Unlock()
	if len(dialed) == 0 {
		t.Fatal("the review made no call at all, so the probe measured nothing")
	}
	sawGenerator, sawGuardrail := false, false
	for _, addr := range dialed {
		host, _, _ := net.SplitHostPort(addr)
		if !rec.allowed[host] {
			t.Errorf("a connection went to %s", addr)
		}
		sawGenerator = sawGenerator || addr == "nightshift.corp.example:8088"
		sawGuardrail = sawGuardrail || addr == "127.0.0.1:8081"
	}
	if !sawGenerator || !sawGuardrail {
		t.Errorf("the review did not reach the generator and the guardrail: %v", dialed)
	}
	if len(refused) != 0 {
		t.Errorf("the recording dialer refused %v: something asked for a host off the network", refused)
	}

	// The probe: the generator's own client, loaded in local mode, is pointed at the hosted address by
	// hand, past the load check. The dial check is what stops it.
	pol, err := network.Load(lookup)
	if err != nil {
		t.Fatal(err)
	}
	pol = pol.WithDial(rec.dial)
	cfg := generator.Settings{
		BaseURL: "https://integrate.api.nvidia.com/v1", App: "proven", Model: "meterless/auto", Token: "test-token",
		Timeout: 5 * time.Second, Network: pol,
	}
	_, perr := generator.NewLive(cfg, nil).Complete(ctx, "planner", "probe")
	if perr == nil {
		t.Fatal("a call to a public address succeeded")
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.refused) != len(refused) || len(rec.dialed) != len(dialed) {
		t.Fatalf("the probe's address reached the dialler: dialed %v refused %v", rec.dialed[len(dialed):], rec.refused[len(refused):])
	}
	if !strings.Contains(perr.Error(), "nightshift") {
		t.Errorf("the failure does not name the service: %v", perr)
	}
}
