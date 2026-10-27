package guardrail_test

import (
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/service"
)

// With only the key set, the guardrail is on with the server backend and its defaults.
func TestLoadUsesServerDefaults(t *testing.T) {
	seen := map[string]int{}
	s, err := guardrail.Load(func(k string) string {
		seen[k]++
		if k == guardrail.EnvKey {
			return "token-0001"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Mode != guardrail.ModeServer || s.BaseURL != "http://127.0.0.1:8081" || s.Model != "nemotron-3.5-content-safety" {
		t.Fatalf("%#v", s)
	}
	if s.Timeout != 30*time.Second || s.Key != "token-0001" {
		t.Fatalf("%#v", s)
	}
	for k := range seen {
		switch k {
		case guardrail.EnvMode, guardrail.EnvBaseURL, guardrail.EnvModel, guardrail.EnvKey, guardrail.EnvTimeout:
		// ADR-0034: the air-gapped mode is read by every loader that holds an address.
		case network.EnvNetwork, network.EnvAllow:
		default:
			t.Fatalf("Load read %s", k)
		}
	}
	if seen[guardrail.EnvKey] != 1 || seen[guardrail.EnvMode] != 1 {
		t.Fatalf("seen %#v", seen)
	}
}

// The nvidia mode takes NVIDIA's hosted address and model id by default.
func TestLoadNVIDIADefaults(t *testing.T) {
	s, err := guardrail.Load(func(k string) string {
		switch k {
		case guardrail.EnvMode:
			return guardrail.ModeNVIDIA
		case guardrail.EnvKey:
			return "token-0001"
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

// With the guardrail off, a key in the environment is not kept.
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

// An unknown mode is refused by name.
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

// A timeout that is not a Go duration is refused.
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

// A zero timeout is refused: a check must have a bound.
// Mutated field: PROVEN_GUARDRAIL_TIMEOUT, set to zero.
func TestLoadRefusesAZeroTimeout(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvTimeout, "0s"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvTimeout {
		t.Fatalf("%#v", cfg)
	}
}

// An empty key names the guardrail down. It is not mode off, and it is not a
// configuration refusal. Mutated field: PROVEN_GUARDRAIL_KEY, set to empty.
func TestLoadNamesAnEmptyKeyDown(t *testing.T) {
	_, err := guardrail.Load(func(string) string { return "" })
	d, ok := service.As(err)
	if !ok || d.Service != guardrail.Service || !strings.Contains(d.Reason, guardrail.EnvKey) {
		t.Fatalf("%v", err)
	}
	if _, isCfg := err.(*guardrail.ConfigError); isCfg {
		t.Fatal("an empty key was refused as configuration")
	}
}

// A key that contains a line break is still refused, and the refusal prints
// no value. Mutated field: PROVEN_GUARDRAIL_KEY, set to a non-empty value
// with a line break.
func TestLoadRefusesAKeyWithALineBreak(t *testing.T) {
	_, err := guardrail.Load(func(k string) string {
		if k == guardrail.EnvKey {
			return "not-a-secret\n"
		}
		return ""
	})
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvKey || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
	if strings.Contains(err.Error(), "not-a-secret") {
		t.Fatal("the refusal printed the key")
	}
}

// A base that is not an http or https URL is refused.
// Mutated field: PROVEN_GUARDRAIL_BASE_URL.
func TestLoadRefusesABadBaseURL(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvBaseURL, "not a url"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// A base with a path other than /v1 is refused.
// Mutated field: PROVEN_GUARDRAIL_BASE_URL, a path other than empty or /v1.
func TestLoadRefusesABaseURLWithAPath(t *testing.T) {
	_, err := guardrail.Load(envWith(guardrail.EnvBaseURL, "http://127.0.0.1:8081/safety"))
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvBaseURL {
		t.Fatalf("%#v", cfg)
	}
}

// A base URL ending in /v1 is accepted.
func TestLoadAcceptsAnAPIRoot(t *testing.T) {
	s, err := guardrail.Load(envWith(guardrail.EnvBaseURL, "http://127.0.0.1:8081/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if s.BaseURL != "http://127.0.0.1:8081/v1" {
		t.Fatalf("base %s", s.BaseURL)
	}
}

// envWith is an environment with a key and one other setting.
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

// configErr fails the test unless err is a ConfigError.
func configErr(t *testing.T, err error) *guardrail.ConfigError {
	t.Helper()
	cfg, ok := err.(*guardrail.ConfigError)
	if !ok {
		t.Fatalf("got %v", err)
	}
	return cfg
}
