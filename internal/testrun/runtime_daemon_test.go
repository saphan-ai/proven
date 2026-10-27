package testrun

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// podmanHome writes a containers.conf under a new home and returns the home.
func podmanHome(t *testing.T, uri string) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	conf := "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\nuri = \"" + uri + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "containers.conf"), []byte(conf), 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}

// TestLocalRefusesAPodmanRemoteServiceThatIsPublic proves ADR-0034 for a runtime other than docker. The
// runtime is any command (PROVEN_TEST_RUNTIME), and podman, which the air-gapped page names, finds a
// remote service in its own configuration under HOME (containers.conf: [engine] remote and
// service_destinations), not in DOCKER_HOST or a docker context. HOME reaches the runtime process.
// Load in local mode refuses, by name, a runtime whose remote service is off the organisation's network.
// Mutated field: the uri of the service destination (a public host here; 10.0.0.5 would pass).
func TestLocalRefusesAPodmanRemoteServiceThatIsPublic(t *testing.T) {
	home := podmanHome(t, "ssh://core@builder.example.com:22/run/podman/podman.sock")
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": home}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("local mode loaded a podman runtime whose remote service is builder.example.com")
	}
}

// TestLocalAcceptsAPodmanRemoteServiceOnTheNetwork proves the podman check judges the address and does
// not refuse podman: an address on the organisation's network loads, and public mode loads the public one.
// Mutated field: the uri of the service destination.
func TestLocalAcceptsAPodmanRemoteServiceOnTheNetwork(t *testing.T) {
	near := podmanHome(t, "tcp://10.0.0.5:2375")
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": near}
	if _, err := Load(func(k string) string { return env[k] }); err != nil {
		t.Fatalf("a podman service on 10.0.0.5 was refused: %v", err)
	}
	// An ssh destination is refused even on a private IP: its route is ssh's to choose.
	viaSSH := podmanHome(t, "ssh://core@10.0.0.5:22/run/podman/podman.sock")
	env = map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": viaSSH}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("a podman ssh service on 10.0.0.5 was accepted")
	}
	far := podmanHome(t, "ssh://core@builder.example.com:22/run/podman/podman.sock")
	env = map[string]string{"PROVEN_NETWORK": "public", "PROVEN_TEST_RUNTIME": "podman", "HOME": far}
	if _, err := Load(func(k string) string { return env[k] }); err != nil {
		t.Fatalf("public mode refused podman: %v", err)
	}
}

// TestLocalRefusesAPodmanConnectionsFileThatIsPublic proves the connections file, the other place podman
// keeps a remote address, is judged too.
// Mutated field: the URI of the connection.
func TestLocalRefusesAPodmanConnectionsFileThatIsPublic(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := `{"Connection":{"Default":"far","Connections":{"far":{"URI":"ssh://core@builder.example.com:22/run/podman/podman.sock"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "podman-connections.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": home}
	_, err := Load(func(k string) string { return env[k] })
	var ce *ConfigError
	if !errors.As(err, &ce) || !strings.Contains(ce.Error(), "builder.example.com") {
		t.Fatalf("local mode loaded a podman connection to builder.example.com: %v", err)
	}
}

// TestLocalRefusesARuntimeItCannotJudge proves a runtime that is neither docker nor podman is refused by
// name in local mode, because Proven cannot tell where its daemon is, and runs in public mode.
// Mutated field: PROVEN_TEST_RUNTIME.
func TestLocalRefusesARuntimeItCannotJudge(t *testing.T) {
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "nerdctl", "HOME": t.TempDir()}
	_, err := Load(func(k string) string { return env[k] })
	var ce *ConfigError
	if !errors.As(err, &ce) || ce.Field != EnvRuntime || !strings.Contains(ce.Error(), "nerdctl") {
		t.Fatalf("local mode loaded a runtime it cannot judge: %v", err)
	}
	env["PROVEN_NETWORK"] = "public"
	if _, err := Load(func(k string) string { return env[k] }); err != nil {
		t.Fatalf("public mode refused the runtime: %v", err)
	}
}
