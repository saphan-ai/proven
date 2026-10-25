package guardrail_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/guardrail"
)

// TestLoadRefusesAKeyTooShortToBeRedacted proves the property that every
// configured secret can be redacted: a non-empty key shorter than the minimum
// secret length is refused by its variable name, and the refusal prints no
// value. Mutated field: PROVEN_GUARDRAIL_KEY, set to 7 bytes.
func TestLoadRefusesAKeyTooShortToBeRedacted(t *testing.T) {
	const short = "abcdefg"
	_, err := guardrail.Load(func(k string) string {
		if k == guardrail.EnvKey {
			return short
		}
		return ""
	})
	cfg := configErr(t, err)
	if cfg.Field != guardrail.EnvKey || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
	if strings.Contains(err.Error(), short) {
		t.Fatal("the refusal printed the key")
	}
}

// TestLoadAcceptsAKeyOfTheMinimumLength is the boundary control of the refusal
// above: the same field at 8 bytes loads.
func TestLoadAcceptsAKeyOfTheMinimumLength(t *testing.T) {
	if _, err := guardrail.Load(func(k string) string {
		if k == guardrail.EnvKey {
			return "abcdefgh"
		}
		return ""
	}); err != nil {
		t.Fatal(err)
	}
}
