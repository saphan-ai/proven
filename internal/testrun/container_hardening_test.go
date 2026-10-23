package testrun

import (
	"context"
	"os"
	"reflect"
	"sort"
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

// TestArgvScratchIsBoundedTmpfs proves the only writable place beside the
// checkout is a size-bounded tmpfs at /tmp that allows exec (the runtimes mount
// a tmpfs noexec by default, and a Go test runs the binary it builds there) and
// refuses setuid files and devices. Mutated option: exec; putting noexec back
// turns this red.
func TestArgvScratchIsBoundedTmpfs(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	argv := c.Argv(t.TempDir(), "go test ./...")
	mounts := 0
	for i, a := range argv {
		if a != "--tmpfs" {
			continue
		}
		mounts++
		opts := strings.Split(argv[i+1], ":")
		if len(opts) != 2 || opts[0] != "/tmp" {
			t.Fatalf("tmpfs %q is not the /tmp scratch", argv[i+1])
		}
		have := map[string]bool{}
		for _, o := range strings.Split(opts[1], ",") {
			have[o] = true
		}
		for _, want := range []string{"rw", "exec", "nosuid", "nodev", "size=64m"} {
			if !have[want] {
				t.Errorf("tmpfs %q lacks option %q", argv[i+1], want)
			}
		}
		if have["noexec"] {
			t.Errorf("tmpfs %q is noexec; a test binary could not run", argv[i+1])
		}
	}
	if mounts != 1 {
		t.Fatalf("%d --tmpfs mounts in %v, want exactly the /tmp scratch", mounts, argv)
	}
}

// TestArgvPassesOnlyTheFixedEnvironment proves the container receives exactly
// HOME=/tmp and GOCACHE=/tmp/go-build, so the build cache lands on scratch, and
// no other -e value, whatever Proven's own environment holds (ADR-0025).
// Mutated field: Proven's environment; a planted secret must not appear.
func TestArgvPassesOnlyTheFixedEnvironment(t *testing.T) {
	t.Setenv("METERLESS_RUNTIME_TOKEN", secretMarker)
	t.Setenv("GOCACHE", "/elsewhere")
	c := mustContainer(t, DefaultContainerConfig())
	argv := c.Argv(t.TempDir(), "go test ./...")
	forbidEnv(t, argv)
	for _, a := range argv {
		if strings.Contains(a, secretMarker) || strings.Contains(a, "/elsewhere") {
			t.Fatalf("Proven's environment reached argv: %q", a)
		}
	}
}

// TestRuntimeProcessEnvironmentIsExactlyTheNamedOnes proves the runtime process
// starts with PATH, HOME and the daemon-locating names that are set, and with
// nothing else, so a secret cannot reach its output (ADR-0025). Positive
// control: the expected names must be present, so a child that printed nothing
// fails. Mutated call: commandOutput's cmd.Env; os.Environ() turns this red.
func TestRuntimeProcessEnvironmentIsExactlyTheNamedOnes(t *testing.T) {
	t.Setenv("METERLESS_RUNTIME_TOKEN", secretMarker)
	t.Setenv("PROVEN_GUARDRAIL_KEY", secretMarker)
	t.Setenv("PATH", "/usr/bin:/bin")
	t.Setenv("HOME", "/home/prover")
	t.Setenv("DOCKER_HOST", "unix:///run/test.sock")
	for _, name := range []string{"DOCKER_CONTEXT", "DOCKER_CONFIG", "XDG_RUNTIME_DIR"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	text, err := commandOutput(context.Background(), []string{"/usr/bin/env"})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Fields(text)
	sort.Strings(got)
	want := []string{"DOCKER_HOST=unix:///run/test.sock", "HOME=/home/prover", "PATH=/usr/bin:/bin"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("runtime process environment %q, want exactly %q", got, want)
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
