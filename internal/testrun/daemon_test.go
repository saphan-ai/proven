package testrun

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func localEnv(extra map[string]string) func(string) string {
	m := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "docker.corp.example"}
	for k, v := range extra {
		m[k] = v
	}
	return func(k string) string { return m[k] }
}

// TestLocalRefusesAPublicDaemonAddressAtLoad proves ADR-0034 for the container runtime: the address of
// its daemon, from DOCKER_HOST or CONTAINER_HOST, is a local socket or a host on the network, or Load
// refuses it by name. The same addresses load in public mode.
// Mutated field: the daemon address of the one variable under test.
func TestLocalRefusesAPublicDaemonAddressAtLoad(t *testing.T) {
	for _, name := range []string{"DOCKER_HOST", "CONTAINER_HOST"} {
		for _, bad := range []string{"tcp://203.0.113.5:2375", "tcp://docker.example.com:2376", "ssh://builder@public.example.com", "public.example.com:2375"} {
			_, err := Load(localEnv(map[string]string{name: bad}))
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != name || !strings.Contains(ce.Error(), "allowed:") {
				t.Errorf("%s=%s: %v", name, bad, err)
			}
			if _, err := Load(func(k string) string {
				if k == name {
					return bad
				}
				return ""
			}); err != nil {
				t.Errorf("public mode refused %s=%s: %v", name, bad, err)
			}
		}
		for _, good := range []string{"unix:///var/run/docker.sock", "npipe:////./pipe/docker_engine", "tcp://127.0.0.1:2375", "tcp://10.0.0.7:2376", "tcp://docker.corp.example:2376", "ssh://builder@docker.corp.example"} {
			if _, err := Load(localEnv(map[string]string{name: good})); err != nil {
				t.Errorf("%s=%s refused: %v", name, good, err)
			}
		}
	}
}

// writeContext writes a docker config whose current context is name and points at host.
func writeContext(t *testing.T, name, host string) string {
	t.Helper()
	dir := t.TempDir()
	sum := sha256.Sum256([]byte(name))
	meta := filepath.Join(dir, "contexts", "meta", hex.EncodeToString(sum[:]))
	if err := os.MkdirAll(meta, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(meta, "meta.json"), []byte(`{"Name":"`+name+`","Endpoints":{"docker":{"Host":"`+host+`"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"currentContext":"`+name+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestLocalRefusesADockerContextWhoseDaemonIsPublic proves the context in force is read and judged:
// a context named by the config file, or by DOCKER_CONTEXT, that points at a public daemon is refused,
// and a context that points at a local one loads. A context that cannot be read is refused by name.
// Mutated field: the Host of the context's docker endpoint.
func TestLocalRefusesADockerContextWhoseDaemonIsPublic(t *testing.T) {
	remote := writeContext(t, "remote", "tcp://203.0.113.5:2376")
	_, err := Load(localEnv(map[string]string{"DOCKER_CONFIG": remote}))
	ce, ok := err.(*ConfigError)
	if !ok || ce.Field != "DOCKER_CONTEXT" || !strings.Contains(ce.Error(), "203.0.113.5") {
		t.Fatalf("a public context from the config file: %v", err)
	}
	if _, err := Load(func(k string) string {
		if k == "DOCKER_CONFIG" {
			return remote
		}
		return ""
	}); err != nil {
		t.Fatalf("public mode refused the context: %v", err)
	}
	local := writeContext(t, "lab", "unix:///run/user/1000/docker.sock")
	if _, err := Load(localEnv(map[string]string{"DOCKER_CONFIG": local})); err != nil {
		t.Fatalf("a local context refused: %v", err)
	}
	// DOCKER_CONTEXT names a context the config directory does not hold: it cannot be judged.
	_, err = Load(localEnv(map[string]string{"DOCKER_CONFIG": local, "DOCKER_CONTEXT": "elsewhere"}))
	if ce, ok := err.(*ConfigError); !ok || ce.Field != "DOCKER_CONTEXT" || !strings.Contains(ce.Error(), "cannot be read") {
		t.Fatalf("an unreadable context: %v", err)
	}
	if _, err := Load(localEnv(map[string]string{"DOCKER_CONFIG": t.TempDir(), "DOCKER_CONTEXT": "default"})); err != nil {
		t.Fatalf("the default context refused: %v", err)
	}
}
