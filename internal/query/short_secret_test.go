package query_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
)

func shortKeyEnv(key string) func(string) string {
	return func(k string) string {
		switch k {
		case query.EnvRESTURL:
			return "http://127.0.0.1:9/api/v1.0"
		case query.EnvAPIKey:
			return key
		case query.EnvOrgID:
			return "123456789"
		default:
			return ""
		}
	}
}

// TestLoadRefusesAnAPIKeyTooShortToBeRedacted proves the property that every
// configured secret can be redacted: a non-empty API key shorter than the
// minimum secret length is refused by its variable name, and the refusal
// prints no value. Mutated field: ZETARIS_API_KEY, set to 7 bytes.
func TestLoadRefusesAnAPIKeyTooShortToBeRedacted(t *testing.T) {
	const short = "abcdefg"
	_, err := query.Load(shortKeyEnv(short))
	var cfg *query.ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != query.EnvAPIKey || cfg.Value != "" {
		t.Fatalf("%v", err)
	}
	if strings.Contains(err.Error(), short) {
		t.Fatal("the refusal printed the API key")
	}
}

// TestLoadAcceptsAnAPIKeyOfTheMinimumLength is the boundary control of the
// refusal above: the same field at 8 bytes loads.
func TestLoadAcceptsAnAPIKeyOfTheMinimumLength(t *testing.T) {
	if _, err := query.Load(shortKeyEnv("abcdefgh")); err != nil {
		t.Fatal(err)
	}
}
