package generator_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
)

// TestLoadRefusesATokenTooShortToBeRedacted proves the property that every
// configured secret can be redacted: a non-empty token shorter than the
// minimum secret length is refused by its variable name, and the refusal
// prints no value. Mutated field: METERLESS_RUNTIME_TOKEN, set to 7 bytes.
func TestLoadRefusesATokenTooShortToBeRedacted(t *testing.T) {
	const short = "abcdefg"
	_, err := generator.Load(func(k string) string {
		if k == generator.EnvToken {
			return short
		}
		return ""
	})
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvToken || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
	if strings.Contains(err.Error(), short) {
		t.Fatal("the refusal printed the token")
	}
}

// TestLoadAcceptsATokenOfTheMinimumLength is the boundary control of the
// refusal above: the same field at 8 bytes loads.
func TestLoadAcceptsATokenOfTheMinimumLength(t *testing.T) {
	if _, err := generator.Load(func(k string) string {
		if k == generator.EnvToken {
			return "abcdefgh"
		}
		return ""
	}); err != nil {
		t.Fatal(err)
	}
}
