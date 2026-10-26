package review

import (
	"errors"
	"strings"
	"testing"
)

// TestLocalRefusesAPublicCouchDBAddress: with PROVEN_NETWORK=local, a
// PROVEN_COUCHDB_URL whose host is outside the organisation's network is refused
// by Open, by name, like ZETARIS_REST_URL.
// Mutated field: PROVEN_COUCHDB_URL (a public host instead of a private one).
func TestLocalRefusesAPublicCouchDBAddress(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	env := map[string]string{
		"PROVEN_NETWORK":          "local",
		"PROVEN_EVIDENCE_STORE":   "couchdb",
		"PROVEN_COUCHDB_URL":      "https://couch.example.com:5984",
		"PROVEN_COUCHDB_USER":     "reviewer",
		"PROVEN_COUCHDB_PASSWORD": "a-password-of-valid-length",
	}
	_, err := Open(Options{LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
	var ce *ConfigError
	if !errors.As(err, &ce) || ce.Field != "PROVEN_COUCHDB_URL" || !strings.Contains(ce.Reason, "couch.example.com") {
		t.Fatalf("Open = %v; want a *ConfigError naming PROVEN_COUCHDB_URL and host couch.example.com", err)
	}
}

// TestLocalRefusesAPublicTelemetryEndpoint: with PROVEN_NETWORK=local, an OTLP
// endpoint outside the organisation's network is refused by Open, by name,
// before any exporter is built. Each of the three endpoint variables is judged.
// Mutated field: the endpoint variable (a public host instead of loopback).
func TestLocalRefusesAPublicTelemetryEndpoint(t *testing.T) {
	for _, field := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
	} {
		t.Run(field, func(t *testing.T) {
			t.Setenv("PATH", t.TempDir())
			env := map[string]string{
				"PROVEN_NETWORK": "local",
				"HOME":           t.TempDir(),
				field:            "https://otel.example.com",
			}
			_, err := Open(Options{LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
			var ce *ConfigError
			if !errors.As(err, &ce) || ce.Field != field {
				t.Fatalf("Open = %v; want a *ConfigError naming %s", err, field)
			}
		})
	}
}

// TestLocalRefusesAPublicTelemetryEndpointInSampleMode: sample mode reads the
// OTEL variables, so it judges them too.
func TestLocalRefusesAPublicTelemetryEndpointInSampleMode(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK":              "local",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "https://otel.example.com",
	}
	_, err := Open(Options{Sample: true, LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
	var ce *ConfigError
	if !errors.As(err, &ce) || ce.Field != "OTEL_EXPORTER_OTLP_ENDPOINT" {
		t.Fatalf("Open = %v; want a *ConfigError naming OTEL_EXPORTER_OTLP_ENDPOINT", err)
	}
}

// Control: the same settings on a private address are accepted, so the tests
// above fail for the host and not for another field.
func TestLocalAcceptsPrivateCouchDBAndTelemetry(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	env := map[string]string{
		"PROVEN_NETWORK":              "local",
		"HOME":                        t.TempDir(),
		"PROVEN_EVIDENCE_STORE":       "couchdb",
		"PROVEN_COUCHDB_URL":          "http://10.0.0.5:5984",
		"PROVEN_COUCHDB_USER":         "reviewer",
		"PROVEN_COUCHDB_PASSWORD":     "a-password-of-valid-length",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4318",
	}
	if _, err := Open(Options{LogDir: t.TempDir(), Env: func(k string) string { return env[k] }}); err != nil {
		t.Fatalf("Open = %v; want nil for private addresses", err)
	}
}
