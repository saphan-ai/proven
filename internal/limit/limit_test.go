package limit

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestLoadEmptyIsTheDefaultAndABadValueIsRefused proves an empty setting is
// Adapter, a real duration is kept, and a value that is not a duration
// greater than zero is refused by name.
// Mutated field: PROVEN_ADAPTER_TIMEOUT. "nope" and "0s" are not a bound.
func TestLoadEmptyIsTheDefaultAndABadValueIsRefused(t *testing.T) {
	got, err := Load(nil)
	if err != nil || got != Adapter {
		t.Fatalf("nil %s %v", got, err)
	}
	got, err = Load(func(string) string { return "  " })
	if err != nil || got != Adapter {
		t.Fatalf("blank %s %v", got, err)
	}
	got, err = Load(func(string) string { return "2s" })
	if err != nil || got != 2*time.Second {
		t.Fatalf("set %s %v", got, err)
	}
	for _, bad := range []string{"nope", "0s", "-1s"} {
		_, err = Load(func(string) string { return bad })
		var cfg *ConfigError
		if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvAdapter || cfg.Value != bad {
			t.Fatalf("%s: %v", bad, err)
		}
		if !strings.Contains(err.Error(), "allowed") || strings.Contains(err.Error(), "\n") {
			t.Fatal(err)
		}
	}
}

// TestBoundUsesTheCarriedDuration proves Bound ends at the duration the
// review loaded, and at Adapter when the context carries none.
// Mutated field: the duration on the context. A nanosecond bound that is
// ignored leaves the context unexpired.
func TestBoundUsesTheCarriedDuration(t *testing.T) {
	ctx := WithBound(context.Background(), time.Nanosecond)
	ctx, cancel := Bound(ctx)
	defer cancel()
	if ctx.Err() == nil {
		t.Fatal("the carried bound was not applied")
	}
	plain, cancelPlain := Bound(context.Background())
	defer cancelPlain()
	deadline, ok := plain.Deadline()
	if !ok {
		t.Fatal("Adapter bound set no deadline")
	}
	left := time.Until(deadline)
	if left < 4*time.Second || left > 5*time.Second+time.Second {
		t.Fatalf("default bound %s", left)
	}
}

func asConfig(err error, dst **ConfigError) bool {
	cfg, ok := err.(*ConfigError)
	if ok {
		*dst = cfg
	}
	return ok
}
