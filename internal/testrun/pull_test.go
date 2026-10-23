package testrun

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
)

// TestLocalNetworkNeverPullsTheImage proves the run line carries --pull never
// in every network mode. The flag is unconditional (ADR-0025); local mode does
// not add a second switch.
// Mutated field: PROVEN_NETWORK.
func TestLocalNetworkNeverPullsTheImage(t *testing.T) {
	for _, mode := range []string{"", "public", "local"} {
		cfg, err := Load(func(k string) string {
			if k == "PROVEN_NETWORK" {
				return mode
			}
			return ""
		})
		if err != nil {
			t.Fatalf("%q: %v", mode, err)
		}
		argv := mustContainer(t, cfg).Argv(t.TempDir(), goTest)
		if !strings.Contains(strings.Join(argv, " "), "--pull never") {
			t.Errorf("%q: run line %v", mode, argv)
		}
	}
}

// TestLocalMissingImageIsTheProverDownByName proves a missing image is the
// Prover down in every network mode, because the run never pulls. The refusal
// names the image and the pull.
// Mutated field: PROVEN_NETWORK.
func TestLocalMissingImageIsTheProverDownByName(t *testing.T) {
	code := exitCode(t, 125)
	missing := func(context.Context, []string) (string, error) {
		return "Unable to find image 'golang:1.26' locally\ndocker: Error response from daemon: pull access denied.", code
	}
	for _, mode := range []string{"", "public", "local"} {
		cfg, err := Load(func(k string) string {
			if k == "PROVEN_NETWORK" {
				return mode
			}
			return ""
		})
		if err != nil {
			t.Fatalf("%q: %v", mode, err)
		}
		c := mustContainer(t, cfg)
		c.run = missing
		_, err = c.Run(context.Background(), Request{
			Path: "billing/discount_test.go", Name: "TestDiscounted", BaseDir: t.TempDir(), ChangeDir: t.TempDir(),
		})
		var down *service.Down
		if !errors.As(err, &down) || down.Service != "prover" || !strings.Contains(down.Reason, "golang:1.26") || !strings.Contains(down.Reason, "next:") {
			t.Fatalf("%q: error %v", mode, err)
		}
	}
}
