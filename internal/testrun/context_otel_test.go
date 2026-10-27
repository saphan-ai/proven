package testrun

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// otelContext writes a docker context "lab" whose daemon is a local socket and
// whose metadata names endpoint for the docker CLI's telemetry, makes it the
// current one, and returns the config directory.
func otelContext(t *testing.T, endpoint string) string {
	t.Helper()
	dir := t.TempDir()
	sum := sha256.Sum256([]byte("lab"))
	meta := filepath.Join(dir, "contexts", "meta", hex.EncodeToString(sum[:]))
	if err := os.MkdirAll(meta, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"Name":"lab","Metadata":{"otel":{"OTEL_EXPORTER_OTLP_ENDPOINT":"` + endpoint + `"}},` +
		`"Endpoints":{"docker":{"Host":"unix:///run/user/1000/docker.sock"}}}`
	if err := os.WriteFile(filepath.Join(meta, "meta.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"currentContext":"lab"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestLocalRefusesAContextTelemetryEndpoint: the docker CLI exports its own
// telemetry to the OTLP endpoint its context names, so in local mode a context
// naming one outside the network, or by a name, is refused naming
// DOCKER_CONTEXT; a socket or a private IP loads, and public mode is unchanged.
// Mutated field: Metadata.otel.OTEL_EXPORTER_OTLP_ENDPOINT of the context.
func TestLocalRefusesAContextTelemetryEndpoint(t *testing.T) {
	for _, r := range []struct {
		endpoint string
		refused  bool
	}{
		{"https://203.0.113.7:4318", true},
		{"https://otel.example:4318", true},
		{"http://10.0.0.5:4318", false},
		{"unix:///run/user/1000/otel.sock", false},
	} {
		dir := otelContext(t, r.endpoint)
		_, err := Load(localEnv(map[string]string{"DOCKER_CONFIG": dir}))
		ce, ok := err.(*ConfigError)
		if r.refused && (!ok || ce.Field != "DOCKER_CONTEXT" || !strings.Contains(ce.Error(), "otel")) {
			t.Errorf("%s: local mode loaded a context whose telemetry endpoint is outside the network or a name: %v", r.endpoint, err)
		}
		if !r.refused && err != nil {
			t.Errorf("%s: control: a local socket or a private address must load: %v", r.endpoint, err)
		}
		if _, err := Load(func(k string) string {
			if k == "DOCKER_CONFIG" {
				return dir
			}
			return ""
		}); err != nil {
			t.Errorf("%s: public mode refused the context: %v", r.endpoint, err)
		}
	}
}
