package generator_test

import (
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
)

// With only the token set, every other setting takes its documented default.
func TestLoadUsesDefaults(t *testing.T) {
	s, err := generator.Load(func(k string) string {
		if k == generator.EnvToken {
			return "tok"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "http://127.0.0.1:8088" || s.App != "proven" || s.Model != "meterless/auto" {
		t.Fatalf("%#v", s)
	}
	if s.SkepticModel != generator.DefaultSkepticModel || s.ModelFor("skeptic") != generator.DefaultSkepticModel {
		t.Fatalf("skeptic model %s", s.SkepticModel)
	}
	if s.ModelFor("planner") != "meterless/auto" {
		t.Fatalf("planner model %s", s.ModelFor("planner"))
	}
	if s.Timeout != 60*time.Second {
		t.Fatalf("timeout %s", s.Timeout)
	}
	if s.Token != "tok" {
		t.Fatal("token")
	}
}

// A timeout that is not a Go duration is refused by name.
// Mutated field: PROVEN_GENERATOR_TIMEOUT.
func TestLoadRefusesABadTimeout(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvTimeout, "nope"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvTimeout || cfg.Value != "nope" {
		t.Fatalf("%#v", cfg)
	}
	if len(cfg.Allowed) != 1 || !strings.Contains(cfg.Allowed[0], "Go duration") {
		t.Fatalf("allowed %#v", cfg.Allowed)
	}
}

// A zero timeout is refused: a call must have a bound.
// Mutated field: PROVEN_GENERATOR_TIMEOUT, set to zero.
func TestLoadRefusesAZeroTimeout(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvTimeout, "0s"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvTimeout {
		t.Fatalf("%#v", cfg)
	}
}

// A live generator without a token is refused, and the refusal prints no value.
// Mutated field: METERLESS_RUNTIME_TOKEN. Empty, so the error must not invent one.
func TestLoadRefusesAMissingToken(t *testing.T) {
	_, err := generator.Load(func(string) string { return "" })
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvToken || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
}

// A base that is not an http or https URL is refused.
// Mutated field: PROVEN_GENERATOR_BASE_URL.
func TestLoadRefusesABadBaseURL(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvBaseURL, "not a url"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// A base with a path other than /v1 is refused.
// Mutated field: PROVEN_GENERATOR_BASE_URL, a path other than empty or /v1.
func TestLoadRefusesABaseURLWithAPath(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvBaseURL, "http://127.0.0.1:8088/nightshift"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// An unknown agent in the require-NVIDIA list is refused, not ignored.
// Mutated field: PROVEN_GENERATOR_REQUIRE_NVIDIA.
func TestLoadRefusesAnUnknownAgent(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvRequireNVIDIA, "planner,bob"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvRequireNVIDIA || cfg.Value != "planner,bob" {
		t.Fatalf("%#v", cfg)
	}
}

// The Skeptic's model can be set to another family's model.
func TestLoadAcceptsAnotherSkepticModel(t *testing.T) {
	s, err := generator.Load(envWith(generator.EnvSkepticModel, "anthropic/claude-opus"))
	if err != nil {
		t.Fatal(err)
	}
	if s.SkepticModel != "anthropic/claude-opus" || s.ModelFor("skeptic") != "anthropic/claude-opus" {
		t.Fatalf("%s", s.SkepticModel)
	}
}

// A Skeptic model with a space is refused.
func TestLoadRefusesASkepticModelWithASpace(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvSkepticModel, "not a model"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvSkepticModel {
		t.Fatalf("%#v", cfg)
	}
}

// A base URL ending in /v1 is accepted and kept as given.
func TestLoadAcceptsAnAPIRoot(t *testing.T) {
	s, err := generator.Load(envWith(generator.EnvBaseURL, "http://127.0.0.1:8088/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "http://127.0.0.1:8088/v1" {
		t.Fatalf("base %s", s.BaseURL)
	}
}

// envWith is an environment with a token and one other setting.
func envWith(key, value string) func(string) string {
	return func(k string) string {
		if k == generator.EnvToken {
			return "tok"
		}
		if k == key {
			return value
		}
		return ""
	}
}

// configErr fails the test unless err is a ConfigError.
func configErr(t *testing.T, err error) *generator.ConfigError {
	t.Helper()
	cfg, ok := err.(*generator.ConfigError)
	if !ok {
		t.Fatalf("got %v", err)
	}
	return cfg
}
