package testrun

import (
	"os"
	"path/filepath"
	"testing"
)

// TestR13LocalJudgesPodmanBehindADockerName: a runtime named docker can be podman (the podman-docker
// package installs a docker command that runs podman). Podman then reads its own containers.conf, not
// DOCKER_HOST. In local mode, Load must refuse a docker-named runtime whose podman configuration names a
// remote service outside the organisation's network, or refuse a docker command it cannot tell is docker.
// Mutated field: the runtime (a docker command that runs podman) and the uri of the service destination.
func TestR13LocalJudgesPodmanBehindADockerName(t *testing.T) {
	home := podmanHome(t, "tcp://203.0.113.8:2375")
	bin := t.TempDir()
	shim := filepath.Join(bin, "docker")
	if err := os.WriteFile(shim, []byte("#!/bin/sh\nexec podman \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": shim, "HOME": home}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("local mode loaded a docker command that runs podman, whose remote service is 203.0.113.8")
	}
}

// TestR13LocalJudgesAQuotedUriKey: TOML allows a quoted key, so "uri" = "..." is the same setting podman
// reads as uri = "...". In local mode, Load must judge that destination and refuse it when it is outside.
// Mutated field: the spelling of the uri key (quoted).
func TestR13LocalJudgesAQuotedUriKey(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	conf := "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\n\"uri\" = \"tcp://203.0.113.8:2375\"\n"
	if err := os.WriteFile(filepath.Join(dir, "containers.conf"), []byte(conf), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": home}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("local mode loaded a podman service whose quoted uri key names 203.0.113.8")
	}
}

// TestReviewLocalRefusesQuotedPodmanURIKey is the codex reviewer's test of aa2503e.
// Podman reads the quoted key as the same uri field and selects that service.
// Mutated field: the spelling of the uri key (quoted).
func TestReviewLocalRefusesQuotedPodmanURIKey(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	conf := "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\n\"uri\" = \"tcp://203.0.113.8:2375\"\n"
	if err := os.WriteFile(filepath.Join(dir, "containers.conf"), []byte(conf), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": home}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("local mode accepted a podman destination outside the allowed network")
	}
}
