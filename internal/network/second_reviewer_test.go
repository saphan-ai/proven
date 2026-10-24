package network_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
)

// TestLocalRefusesAPublicSecondReviewerAddress proves ADR-0034 for the second reviewer of ADR-0037: with
// PROVEN_NETWORK=local, an address off the organisation's network is refused at load by name, with
// the key set and also without it, and the same address loads in public mode.
// Mutated field: the host of PROVEN_SECOND_REVIEWER_BASE_URL.
func TestLocalRefusesAPublicSecondReviewerAddress(t *testing.T) {
	for _, withKey := range []bool{true, false} {
		env := map[string]string{
			"PROVEN_NETWORK":                  "local",
			"PROVEN_SECOND_REVIEWER_BASE_URL": "https://api.second.example.com/v1",
			"PROVEN_SECOND_REVIEWER_MODEL":    "other-model",
		}
		if withKey {
			env["PROVEN_SECOND_REVIEWER_KEY"] = "second-key-ok"
		}
		_, _, err := generator.LoadSecond(func(k string) string { return env[k] })
		field, text := refusedField(t, err)
		if field != "PROVEN_SECOND_REVIEWER_BASE_URL" || !strings.Contains(text, "api.second.example.com") {
			t.Fatalf("key %v: refusal %q", withKey, text)
		}
		env["PROVEN_NETWORK"] = "public"
		if _, _, err := generator.LoadSecond(func(k string) string { return env[k] }); err != nil {
			t.Fatalf("key %v: public mode refused it too: %v", withKey, err)
		}
	}
}

// TestLocalSecondReviewerDialsOnlyItsAllowedAddress proves the dialer covers the second reviewer: its
// client is built from the policy, so its one connection goes through the policy's dial and to the
// address that was loaded.
// Mutated field: the Network of the settings NewSecond builds its client from (nil leaves a plain client).
func TestLocalSecondReviewerDialsOnlyItsAllowedAddress(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK":                  "local",
		"PROVEN_SECOND_REVIEWER_BASE_URL": "http://10.0.0.7:8080/v1",
		"PROVEN_SECOND_REVIEWER_MODEL":    "other-model",
		"PROVEN_SECOND_REVIEWER_KEY":      "second-key-ok",
		"PROVEN_GENERATOR_TIMEOUT":        "2s",
	}
	sec, connected, err := generator.LoadSecond(func(k string) string { return env[k] })
	if err != nil || !connected {
		t.Fatalf("load: connected %v, %v", connected, err)
	}
	var mu sync.Mutex
	var seen []string
	sec.Network = sec.Network.WithDial(func(_ context.Context, _, addr string) (net.Conn, error) {
		mu.Lock()
		seen = append(seen, addr)
		mu.Unlock()
		return nil, errors.New("no network in this test")
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = generator.NewSecond(sec, nil).Complete(ctx, "planner", "hello")
	mu.Lock()
	defer mu.Unlock()
	if len(seen) == 0 || seen[0] != "10.0.0.7:8080" {
		t.Fatalf("dials %v; want the second reviewer's own address through the policy's dial", seen)
	}
	for _, addr := range seen {
		if addr != "10.0.0.7:8080" {
			t.Fatalf("dials %v; want only 10.0.0.7:8080", seen)
		}
	}
}
