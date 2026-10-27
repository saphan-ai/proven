package telemetry

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

// TestSetupLocalDialsBothExportersThroughThePolicy: under PROVEN_NETWORK=local
// the trace and the metric exporter each open their connection through the
// policy's dial, so a client of their own (and its proxy from the environment)
// is not used. The recorded dial is the proof.
// Mutated field: the exporter's HTTP client (the SDK's default instead of the
// policy's), which records no dial.
func TestSetupLocalDialsBothExportersThroughThePolicy(t *testing.T) {
	var mu sync.Mutex
	var dials []string
	dial := func(_ context.Context, _, addr string) (net.Conn, error) {
		mu.Lock()
		dials = append(dials, addr)
		mu.Unlock()
		client, server := net.Pipe()
		server.Close()
		return client, nil
	}
	env := map[string]string{
		"PROVEN_NETWORK":                      "local",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT":  "http://127.0.0.1:4318/v1/traces",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": "http://127.0.0.1:4319/v1/metrics",
		"OTEL_METRIC_EXPORT_INTERVAL":         "1000",
	}
	tel, err := SetupWithDial(context.Background(), func(k string) string { return env[k] }, dial)
	if err != nil {
		t.Fatal(err)
	}
	ctx, rev := tel.StartReview(context.Background())
	rev.End(nil)
	flush, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = tel.Flush(flush) // the pipe answers nothing, so an error is expected; the dials are the proof
	mu.Lock()
	defer mu.Unlock()
	seen := map[string]bool{}
	for _, d := range dials {
		seen[d] = true
	}
	if !seen["127.0.0.1:4318"] || !seen["127.0.0.1:4319"] {
		t.Fatalf("dials = %v; want the trace endpoint 127.0.0.1:4318 and the metric endpoint 127.0.0.1:4319", dials)
	}
}
