package review

import (
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/markov"
)

// TestOpenRefusesABadBoundedStepsSettingByName proves the bounded-steps settings are validated on load (ADR-0013,
// ADR-0048): a live Open with each setting at a value it does not allow is refused before any review, naming the
// variable and the allowed values; the same Open with allowed values succeeds. Mutated field: one variable at a time.
func TestOpenRefusesABadBoundedStepsSettingByName(t *testing.T) {
	good := map[string]string{markov.EnvSwitch: "on", markov.EnvThreshold: "12000", markov.EnvUnit: "hunk", markov.EnvCarry: "512"}
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if _, err := Open(Options{LogDir: t.TempDir(), Env: env(good)}); err != nil {
		t.Fatalf("allowed values were refused: %v", err)
	}
	for key, bad := range map[string]string{
		markov.EnvSwitch: "yes", markov.EnvThreshold: "none", markov.EnvUnit: "line", markov.EnvCarry: "64",
	} {
		m := map[string]string{}
		for k, v := range good {
			m[k] = v
		}
		m[key] = bad
		_, err := Open(Options{LogDir: t.TempDir(), Env: env(m)})
		var cfg *ConfigError
		if !errors.As(err, &cfg) || cfg.Field != key || cfg.Value != bad || len(cfg.Allowed) == 0 {
			t.Errorf("%s=%s: %v", key, bad, err)
			continue
		}
		if !strings.Contains(err.Error(), key) {
			t.Errorf("the refusal does not name %s: %v", key, err)
		}
	}
}
