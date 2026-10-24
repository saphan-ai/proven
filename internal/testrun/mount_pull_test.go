package testrun

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
)

// TestRunLineNeverPullsTheImage proves the run line carries --pull never, so an absent image is not
// fetched at test time (ADR-0025).
func TestRunLineNeverPullsTheImage(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	assertPair(t, c.Argv(t.TempDir(), goTest), "--pull", "never")
}

// TestAbsentImageIsTheProverDownAndNamesThePull proves that when the runtime says the image is not on
// the host, the run reports the Prover down with the pull to do, not a test that failed.
func TestAbsentImageIsTheProverDownAndNamesThePull(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	c.run = func(_ context.Context, argv []string) (string, error) {
		if argv[1] == "rm" {
			return "", nil
		}
		return "Unable to find image 'golang:1.26' locally", exitCode(t, 125)
	}
	_, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: t.TempDir(),
	})
	d, ok := service.As(err)
	if !ok {
		t.Fatalf("an absent image was not the Prover down: %v", err)
	}
	if !strings.Contains(d.Reason, "next: docker pull golang:1.26") {
		t.Fatalf("the reason does not name the pull: %s", d.Reason)
	}
}

// TestLinkedCheckoutIsNotMounted proves a checkout directory that is a symbolic link is refused as a
// mount source and no container is started (ADR-0025). Mutated field: only the type of ChangeDir.
func TestLinkedCheckoutIsNotMounted(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "change")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	c := mustContainer(t, DefaultContainerConfig())
	c.run = func(context.Context, []string) (string, error) {
		t.Fatal("a container was started with a linked checkout")
		return "", nil
	}
	out, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: link,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Found || !strings.Contains(out.Output, "symbolic link") {
		t.Fatalf("%#v", out)
	}
}
