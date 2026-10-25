package meterless_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMeterlessRunnersLoadTheShim runs the three conformance runners (H-MEM,
// World Model, Markovian) through the shim. The score is recorded in the scorecard files and is not asserted.
// Exit 2 means the runner could not load the shim.
func TestMeterlessRunnersLoadTheShim(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not on PATH")
	}
	// METERLESS_ENGINES names a Meterless engines/ directory; it is not part of this repository, so without it the
	// test is skipped by name.
	ref := os.Getenv("METERLESS_ENGINES")
	if _, err := os.Stat(filepath.Join(ref, "hmem", "conformance", "runner.ts")); err != nil {
		t.Skip("no Meterless runners: set METERLESS_ENGINES to a Meterless engines/ directory")
	}
	script, err := filepath.Abs("run.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", script)
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	exit := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatal(err)
		}
		exit = ee.ExitCode()
	}
	t.Logf("exit %d\n%s", exit, out)
	if exit == 2 {
		t.Fatal("a runner could not load the shim")
	}
	if exit != 0 && exit != 1 {
		t.Fatalf("run.sh exit %d", exit)
	}
	for _, name := range []string{"SCORECARD_hmem.txt", "SCORECARD_world-model.txt", "SCORECARD_markovian.txt"} {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Fatalf("%s is empty", name)
		}
	}
}
