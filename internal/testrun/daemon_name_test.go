package testrun

import "testing"

// TestLocalDaemonNameIsNotResolvedByTheRuntime: in PROVEN_NETWORK=local a daemon
// address whose host is a name, even a listed one, is not handed to the runtime
// as a name, because the runtime would resolve it with its own resolver, through
// name servers and to answers Proven never judged.
// Mutated field: DOCKER_HOST (a listed name instead of a private IP literal).
func TestLocalDaemonNameIsNotResolvedByTheRuntime(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK":       "local",
		"PROVEN_NETWORK_ALLOW": "docker.corp.example",
		"DOCKER_HOST":          "tcp://docker.corp.example:2376",
		"HOME":                 t.TempDir(),
	}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("Load accepted a daemon host name that the runtime resolves itself; want a refusal naming DOCKER_HOST")
	}
	env["DOCKER_HOST"] = "tcp://10.0.0.5:2376"
	if _, err := Load(func(k string) string { return env[k] }); err != nil {
		t.Fatalf("control: a private IP literal must pass: %v", err)
	}
}
