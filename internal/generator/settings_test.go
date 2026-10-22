package generator_test

import (
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
)

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

// Mutated field: PROVEN_GENERATOR_TIMEOUT, set to zero.
func TestLoadRefusesAZeroTimeout(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvTimeout, "0s"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvTimeout {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: METERLESS_RUNTIME_TOKEN. Empty, so the error must not invent one.
func TestLoadRefusesAMissingToken(t *testing.T) {
	_, err := generator.Load(func(string) string { return "" })
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvToken || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: PROVEN_GENERATOR_BASE_URL.
func TestLoadRefusesABadBaseURL(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvBaseURL, "not a url"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: PROVEN_GENERATOR_BASE_URL, a path other than empty or /v1.
func TestLoadRefusesABaseURLWithAPath(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvBaseURL, "http://127.0.0.1:8088/nightshift"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: PROVEN_GENERATOR_REQUIRE_NVIDIA.
func TestLoadRefusesAnUnknownAgent(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvRequireNVIDIA, "planner,bob"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvRequireNVIDIA || cfg.Value != "planner,bob" {
		t.Fatalf("%#v", cfg)
	}
}

func TestLoadAcceptsAnotherSkepticModel(t *testing.T) {
	s, err := generator.Load(envWith(generator.EnvSkepticModel, "anthropic/claude-opus"))
	if err != nil {
		t.Fatal(err)
	}
	if s.SkepticModel != "anthropic/claude-opus" || s.ModelFor("skeptic") != "anthropic/claude-opus" {
		t.Fatalf("%s", s.SkepticModel)
	}
}

func TestLoadRefusesASkepticModelWithASpace(t *testing.T) {
	_, err := generator.Load(envWith(generator.EnvSkepticModel, "not a model"))
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvSkepticModel {
		t.Fatalf("%#v", cfg)
	}
}

func TestLoadAcceptsAnAPIRoot(t *testing.T) {
	s, err := generator.Load(envWith(generator.EnvBaseURL, "http://127.0.0.1:8088/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "http://127.0.0.1:8088/v1" {
		t.Fatalf("base %s", s.BaseURL)
	}
}

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

func configErr(t *testing.T, err error) *generator.ConfigError {
	t.Helper()
	cfg, ok := err.(*generator.ConfigError)
	if !ok {
		t.Fatalf("got %v", err)
	}
	return cfg
}
