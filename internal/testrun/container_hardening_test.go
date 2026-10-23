package testrun

import (
	"context"
	"strings"
	"testing"
)

// secretMarker stands for the value of a configured secret in Proven's own
// environment. It is never a real credential.
const secretMarker = "hardening-marker-5b1e"

// TestArgvHardensTheContainer proves the container runs as a fixed
// unprivileged user, with a process limit, every capability dropped, no
// privilege gain and a read-only root filesystem, beside the existing
// no-network flag (ADR-0025). Mutated flag: --cap-drop; removing it from argv
// turns the capability assertion red.
func TestArgvHardensTheContainer(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	argv := c.Argv(t.TempDir(), "go test ./...")
	assertPair(t, argv, "--user", "65534:65534")
	assertPair(t, argv, "--pids-limit", "256")
	assertPair(t, argv, "--cap-drop", "ALL")
	assertPair(t, argv, "--security-opt", "no-new-privileges")
	assertFlag(t, argv, "--read-only")
	assertPair(t, argv, "--network", "none")
}

// TestArgvScratchIsBoundedTmpfs proves the only writable places beside the
// checkout are size-bounded tmpfs mounts, so a read-only root still lets a
// Go test build.
func TestArgvScratchIsBoundedTmpfs(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	argv := c.Argv(t.TempDir(), "go test ./...")
	found := false
	for i, a := range argv {
		if a != "--tmpfs" {
			continue
		}
		found = true
		if !strings.Contains(argv[i+1], "size=64m") {
			t.Errorf("tmpfs %q has no size bound", argv[i+1])
		}
	}
	if !found {
		t.Fatalf("no --tmpfs scratch in %v", argv)
	}
}

// TestRuntimeProcessEnvironmentOmitsSecrets proves a secret set in Proven's
// environment is absent from the environment of the runtime process, so the
// runtime's output cannot carry it into a trace (ADR-0025). Mutated call:
// commandOutput's cmd.Env; setting it to os.Environ() turns this red.
func TestRuntimeProcessEnvironmentOmitsSecrets(t *testing.T) {
	t.Setenv("METERLESS_RUNTIME_TOKEN", secretMarker)
	t.Setenv("PROVEN_GUARDRAIL_KEY", secretMarker)
	text, err := commandOutput(context.Background(), []string{"/usr/bin/env"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, secretMarker) {
		t.Fatalf("a secret reached the runtime process:\n%s", text)
	}
}
