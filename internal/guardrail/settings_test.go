package guardrail_test

import (
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/guardrail"
)

func TestLoadUsesServerDefaults(t *testing.T) {
	seen := map[string]int{}
	s, err := guardrail.Load(func(k string) string {
		seen[k]++
		if k == guardrail.EnvKey {
			return "tok"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Mode != guardrail.ModeServer || s.BaseURL != "http://127.0.0.1:8081" || s.Model != "nemotron-3.5-content-safety" {
		t.Fatalf("%#v", s)
	}
	if s.Timeout != 5*time.Second || s.Key != "tok" {
		t.Fatalf("%#v", s)
	}
	for k := range seen {
		switch k {
		case guardrail.EnvMode, guardrail.EnvBaseURL, guardrail.EnvModel, guardrail.EnvKey, guardrail.EnvTimeout:
		default:
			t.Fatalf("Load read %s", k)
		}
	}
	if seen[guardrail.EnvKey] != 1 || seen[guardrail.EnvMode] != 1 {
		t.Fatalf("seen %#v", seen)
	}
}

func TestLoadNVIDIADefaults(t *testing.T) {
	s, err := guardrail.Load(func(k string) string {
		switch k {
		case guardrail.EnvMode:
			return guardrail.ModeNVIDIA
		case guardrail.EnvKey:
			return "tok"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Mode != guardrail.ModeNVIDIA || s.BaseURL != "https://integrate.api.nvidia.com/v1" || s.Model != "nvidia/nemotron-3.5-content-safety" {
		t.Fatalf("%#v", s)
	}
}

func TestLoadOffDoesNotKeepTheKey(t *testing.T) {
	s, err := guardrail.Load(func(k string) string {
		switch k {
		case guardrail.EnvMode:
			return guardrail.ModeOff
		case guardrail.EnvKey:
			return "super-secret"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Mode != guardrail.ModeOff || s.Key != "" || s.BaseURL != "" || s.Model != "" {
		t.Fatalf("%#v", s)
	}
}

// Mutated field: PROVEN_GUARDRAIL_MODE.
func TestLoadRefusesAnUnknownMode(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvMode, "sometimes"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvMode || cfg.Value != "sometimes" {
		t.Fatalf("%#v", cfg)
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatal("the refusal printed the key")
	}
}

// Mutated field: PROVEN_GUARDRAIL_TIMEOUT.
func TestLoadRefusesABadTimeout(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvTimeout, "nope"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvTimeout || cfg.Value != "nope" {
		t.Fatalf("%#v", cfg)
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatal("the refusal printed the key")
	}
}

// Mutated field: PROVEN_GUARDRAIL_TIMEOUT, set to zero.
func TestLoadRefusesAZeroTimeout(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvTimeout, "0s"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvTimeout {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: PROVEN_GUARDRAIL_KEY. Empty, so the error must not invent one.
func TestLoadRefusesAMissingKey(t *testing.T) {
	_, err := guardrail.Load(func(string) string { return "" })
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvKey || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: PROVEN_GUARDRAIL_BASE_URL.
func TestLoadRefusesABadBaseURL(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvBaseURL, "not a url"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// Mutated field: PROVEN_GUARDRAIL_BASE_URL, a path other than empty or /v1.
func TestLoadRefusesABaseURLWithAPath(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvBaseURL, "http://127.0.0.1:8081/safety"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

func TestLoadAcceptsAnAPIRoot(t *testing.T) {
	s, err := guardrail.Load(envWith(guardrail.EnvBaseURL, "http://127.0.0.1:8081/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "http://127.0.0.1:8081/v1" {
		t.Fatalf("base %s", s.BaseURL)
	}
}

func envWith(key, value string) func(string) string {
	return func(k string) string {
		if k == guardrail.EnvKey {
			return "super-secret"
		}
		if k == key {
			return value
		}
		return ""
	}
}

func configErr(t *testing.T, err error) *guardrail.ConfigError {
	t.Helper()
	cfg, ok := err.(*guardrail.ConfigError)
	if !ok {
		t.Fatalf("got %v", err)
	}
	return cfg
}
