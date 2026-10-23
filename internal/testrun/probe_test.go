package testrun

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// TestProbeAsksTheRuntimeAndInspectsTheImage proves the probe makes exactly
// two calls, version then image inspect, and never pulls.
func TestProbeAsksTheRuntimeAndInspectsTheImage(t *testing.T) {
	c := probedContainer(t)
	var got [][]string
	c.run = func(ctx context.Context, argv []string) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		got = append(got, append([]string(nil), argv...))
		return "ok\n", nil
	}
	if err := c.Probe(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("calls %d: %#v", len(got), got)
	}
	if strings.Join(got[0], " ") != "/usr/bin/docker version" {
		t.Fatalf("version %#v", got[0])
	}
	if strings.Join(got[1], " ") != "/usr/bin/docker image inspect golang:1.26" {
		t.Fatalf("inspect %#v", got[1])
	}
	for _, argv := range got {
		for _, a := range argv {
			if a == "pull" {
				t.Fatal("probe pulled the image")
			}
		}
	}
}

// Mutated field: the runtime's exit status. version fails, so inspect is not run.
// It proves a silent runtime is named down with its last output line as the reason.
func TestProbeStopsWhenTheRuntimeDoesNotAnswer(t *testing.T) {
	c := probedContainer(t)
	var got []string
	c.run = func(context.Context, []string) (string, error) {
		got = append(got, "called")
		return "Client: Docker Engine\nCannot connect to the Docker daemon\n", errors.New("exit status 1")
	}
	err := c.Probe(context.Background())
	d, ok := service.As(err)
	if !ok || d.Service != "prover" || !strings.Contains(d.Reason, "runtime did not answer") || !strings.Contains(d.Reason, "Cannot connect to the Docker daemon") || strings.Contains(d.Reason, "Client:") {
		t.Fatalf("%v", err)
	}
	if len(got) != 1 {
		t.Fatalf("calls %d", len(got))
	}
}

// Mutated field: image inspect's exit status. The image is absent, and pull is not a call.
// It proves a missing image is named down and is not fetched.
func TestProbeNamesAMissingImage(t *testing.T) {
	c := probedContainer(t)
	var got [][]string
	c.run = func(_ context.Context, argv []string) (string, error) {
		got = append(got, append([]string(nil), argv...))
		if len(argv) > 1 && argv[1] == "version" {
			return "ok\n", nil
		}
		return "", errors.New("exit status 1")
	}
	err := c.Probe(context.Background())
	d, ok := service.As(err)
	if !ok || d.Service != "prover" || !strings.Contains(d.Reason, "golang:1.26") || !strings.Contains(d.Reason, "not present") {
		t.Fatalf("%v", err)
	}
	if strings.Contains(d.Reason, "pull") {
		t.Fatalf("reason mentions pull: %s", d.Reason)
	}
	if len(got) != 2 || got[1][1] != "image" || got[1][2] != "inspect" {
		t.Fatalf("calls %#v", got)
	}
}

// Mutated field: the probe context deadline. A hung runtime is a timeout.
// It proves the reason is "timeout", not a claim that the image is absent.
func TestProbeTimesOut(t *testing.T) {
	c := probedContainer(t)
	c.cfg.Timeout = 20 * time.Millisecond
	c.run = func(ctx context.Context, _ []string) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	}
	err := c.Probe(context.Background())
	d, ok := service.As(err)
	if !ok || d.Reason != "timeout" {
		t.Fatalf("%v", err)
	}
}

// TestProbeHonoursACancelledContext proves a cancelled caller calls no runtime
// and gets its own cancellation back.
func TestProbeHonoursACancelledContext(t *testing.T) {
	c := probedContainer(t)
	c.run = func(context.Context, []string) (string, error) {
		t.Fatal("runtime was called")
		return "", nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.Probe(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("%v", err)
	}
}

// probedContainer builds a Container with the defaults and a found runtime.
func probedContainer(t *testing.T) *Container {
	t.Helper()
	c, err := NewContainer(DefaultContainerConfig(), func(string) (string, error) {
		return "/usr/bin/docker", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}
