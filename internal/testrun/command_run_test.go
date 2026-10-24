package testrun

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRunCommandPassesTheWordsAndNamesATimeout proves a tool command is the
// words after the image, with no shell, and that a side past its limit is a
// timeout result rather than a pass (ADR-0040).
func TestRunCommandPassesTheWordsAndNamesATimeout(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultContainerConfig()
	cfg.Timeout = time.Minute
	c := mustContainer(t, cfg)
	fast, err := c.WithTimeout(20 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if c.cfg.Timeout != time.Minute {
		t.Fatal("WithTimeout changed the original container")
	}
	var runs [][]string
	fast.run = func(ctx context.Context, argv []string) (string, error) {
		runs = append(runs, append([]string(nil), argv...))
		if len(argv) > 1 && argv[1] == "rm" {
			return "", nil
		}
		<-ctx.Done()
		return "WARNING: DATA RACE", ctx.Err()
	}
	got, err := fast.RunCommand(context.Background(), dir, []string{"go", "test", "-race", "./counter"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != CommandTimeout {
		t.Fatalf("status %s", got.Status)
	}
	if len(runs) == 0 {
		t.Fatal("nothing ran")
	}
	line := runs[0]
	tail := strings.Join(line[len(line)-4:], " ")
	if tail != "go test -race ./counter" {
		t.Fatalf("tail %q in %q", tail, line)
	}
	for _, word := range line {
		if word == "sh" || word == "-c" {
			t.Fatalf("a shell in %q", line)
		}
	}
}

// TestWithTimeoutRefusesZero proves a non-positive limit is refused and does
// not become a container that times out at once.
func TestWithTimeoutRefusesZero(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	if _, err := c.WithTimeout(0); err == nil {
		t.Fatal("zero was accepted")
	}
}

// TestRunCommandDoesNotMountASymlink proves a checkout that is a symbolic
// link is not a mount source (ADR-0025).
func TestRunCommandDoesNotMountASymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	c := mustContainer(t, DefaultContainerConfig())
	started := false
	c.run = func(context.Context, []string) (string, error) {
		started = true
		return "", nil
	}
	got, err := c.RunCommand(context.Background(), link, []string{"go", "test", "-race", "."})
	if err != nil {
		t.Fatal(err)
	}
	if started || got.Status != CommandNotRun || !strings.Contains(got.Output, "symbolic link") {
		t.Fatalf("started %v %#v", started, got)
	}
}
