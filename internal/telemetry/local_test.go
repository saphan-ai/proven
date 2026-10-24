package telemetry

import (
	"context"
	"strings"
	"testing"
)

// TestSetupLocalRefusesAnOutsideEndpoint: Setup itself judges the endpoint under
// PROVEN_NETWORK=local, before an exporter is built.
// Mutated field: the endpoint (203.0.113.9 instead of 127.0.0.1).
func TestSetupLocalRefusesAnOutsideEndpoint(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK":              "local",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://203.0.113.9:4318",
	}
	lookup := func(k string) string { return env[k] }
	_, err := Setup(context.Background(), lookup)
	if err == nil || !strings.Contains(err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT") {
		t.Fatalf("Setup = %v; want a refusal naming OTEL_EXPORTER_OTLP_ENDPOINT", err)
	}
	env["OTEL_EXPORTER_OTLP_ENDPOINT"] = "http://127.0.0.1:4318"
	tel, err := Setup(context.Background(), lookup)
	if err != nil {
		t.Fatalf("control: a loopback endpoint must pass: %v", err)
	}
	_ = tel.Flush(context.Background())
}
