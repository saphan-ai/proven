package review

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// localRuntimeEnv is a local-mode environment with a fake runtime that answers --version.
func localRuntimeEnv(t *testing.T, runtime string) func(string) string {
	t.Helper()
	env := map[string]string{
		"PROVEN_NETWORK":      "local",
		"HOME":                t.TempDir(),
		"PROVEN_TEST_RUNTIME": runtime,
	}
	return func(k string) string { return env[k] }
}

func versionRuntime(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho 'Docker version 29.9.1'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestOpenInLocalModePrintsTheRuntimeItWillRun: at start, local mode prints the absolute path and the
// version of the configured runtime. Mutated field: the mode (public prints nothing).
func TestOpenInLocalModePrintsTheRuntimeItWillRun(t *testing.T) {
	path := versionRuntime(t, "docker")
	var out bytes.Buffer
	if _, err := Open(Options{LogDir: t.TempDir(), Env: localRuntimeEnv(t, path), startLog: &out}); err != nil {
		t.Fatal(err)
	}
	want := "proven: container runtime " + path + ": Docker version 29.9.1\n"
	if out.String() != want {
		t.Fatalf("start line %q, want %q", out.String(), want)
	}

	var pub bytes.Buffer
	env := map[string]string{"PROVEN_TEST_RUNTIME": path, "HOME": t.TempDir()}
	if _, err := Open(Options{LogDir: t.TempDir(), Env: func(k string) string { return env[k] }, startLog: &pub}); err != nil {
		t.Fatal(err)
	}
	if pub.Len() != 0 {
		t.Fatalf("public mode printed %q", pub.String())
	}
}

// TestOpenInLocalModeRefusesAWorldWritableRuntimeByName: Open stops with the seam's ConfigError.
// Mutated field: the mode of the runtime file.
func TestOpenInLocalModeRefusesAWorldWritableRuntimeByName(t *testing.T) {
	path := versionRuntime(t, "docker")
	if err := os.Chmod(path, 0o777); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	_, err := Open(Options{LogDir: t.TempDir(), Env: localRuntimeEnv(t, path), startLog: &out})
	var ce *ConfigError
	if !errors.As(err, &ce) || ce.Field != "PROVEN_TEST_RUNTIME" || !strings.Contains(ce.Reason, "writable by every user") {
		t.Fatalf("not refused by name: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("a refused runtime was described: %q", out.String())
	}
}
