package tool

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/testrun"
)

// TestRaceToolRunInTheContainerProvesTheChange proves a real run of the race
// detector, in the Prover's locked-down container, is a finding proved by a
// tool when the change adds a data race and the base has none (ADR-0040).
// The recorded command and the tool version are filled.
// It runs only with PROVEN_TEST_DOCKER=1.
func TestRaceToolRunInTheContainerProvesTheChange(t *testing.T) {
	if os.Getenv("PROVEN_TEST_DOCKER") != "1" {
		t.Skip("PROVEN_TEST_DOCKER=1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker")
	}
	base := t.TempDir()
	change := t.TempDir()
	const (
		mod = "module demo\n\ngo 1.26\n"
		// The module imports only the standard library, so the run needs no
		// network. The container has none (ADR-0025).
		testSrc = "package demo\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif got := Add(); got < 1 {\n\t\tt.Fatalf(\"Add() = %d\", got)\n\t}\n}\n"
		// The receive happens before the second write, so the writes are ordered.
		baseSrc = "package demo\n\nfunc Add() int {\n\tn := 0\n\tdone := make(chan struct{})\n\tgo func() {\n\t\tn++\n\t\tclose(done)\n\t}()\n\t<-done\n\tn++\n\treturn n\n}\n"
		// The second write does not wait for the first, so the two writes race.
		changeSrc = "package demo\n\nfunc Add() int {\n\tn := 0\n\tdone := make(chan struct{})\n\tgo func() {\n\t\tn++\n\t\tclose(done)\n\t}()\n\tn++\n\t<-done\n\treturn n\n}\n"
	)
	for _, dir := range []string{base, change} {
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "add_test.go"), []byte(testSrc), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(base, "add.go"), []byte(baseSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(change, "add.go"), []byte(changeSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	// ADR-0025: the container runs as a user that does not own these trees.
	// Without this grant it cannot read the checkout the run mounts.
	for _, dir := range []string{base, change} {
		if err := testrun.MakeReadable(dir); err != nil {
			t.Fatal(err)
		}
	}
	box, err := testrun.NewContainer(testrun.DefaultContainerConfig(), exec.LookPath)
	if err != nil {
		t.Fatal(err)
	}
	limit, err := Load(os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	out, err := NewRunner(Race{}, ContainerHost{Box: box, Timeout: limit}).Run(context.Background(), Request{
		Paths: []string{"add.go"}, BaseDir: base, ChangeDir: change,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Proves() || !out.PresentOnChange || out.PresentOnBase {
		t.Fatalf("not a race only on the change: answered=%v change=%v base=%v reason=%s\nchange output:\n%s\nbase output:\n%s",
			out.Answered, out.PresentOnChange, out.PresentOnBase, out.Reason, out.ChangeOutput, out.BaseOutput)
	}
	if out.Command == "" || out.Version == "" {
		t.Fatalf("command %q version %q", out.Command, out.Version)
	}
	if out.Command != "go test -race ." {
		t.Fatalf("command %q", out.Command)
	}
	found, ok := out.Candidate("add.go")
	if !ok {
		t.Fatal("the run reported no finding")
	}
	kept := judge.Decide(found)
	if kept.Dropped || len(kept.Evidence) == 0 || kept.Evidence[0].Kind != candidate.KindTool {
		t.Fatalf("%#v", kept)
	}
	ev := kept.Evidence[0].Tool
	if ev == nil || ev.Command == "" || ev.Version == "" || !ev.PresentOnChange || ev.PresentOnBase {
		t.Fatalf("%#v", ev)
	}
	level, err := review.ProofLevelOf([]review.Evidence{{Kind: review.EvidenceTool}})
	if err != nil || level != review.ProvedByTool {
		t.Fatalf("%s %v", level, err)
	}
}
