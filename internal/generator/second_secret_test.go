package generator_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
)

// TestLoadSecondRefusesAKeyTooShortToBeRedacted proves the second reviewer's key
// meets the same length floor as every other secret: a key shorter than the
// minimum is refused by its variable name and the refusal prints no value.
// Mutated field: the second reviewer's key, set to 3 bytes.
func TestLoadSecondRefusesAKeyTooShortToBeRedacted(t *testing.T) {
	const short = "abc"
	env := map[string]string{
		generator.EnvSecondKey:     short,
		generator.EnvSecondBaseURL: "https://api.example.com/v1",
		generator.EnvSecondModel:   "some-model",
	}
	_, _, err := generator.LoadSecond(func(k string) string { return env[k] })
	if err == nil {
		t.Fatal("a 3-byte second-reviewer key was accepted")
	}
	cfg := configErr(t, err)
	if cfg.Field != generator.EnvSecondKey || cfg.Value != "" {
		t.Fatalf("%#v", cfg)
	}
	if strings.Contains(err.Error(), short) {
		t.Fatal("the refusal printed the key")
	}
}

// TestLoadSecondAcceptsAKeyOfTheMinimumLength is the boundary control of the
// refusal above: the same field at 8 bytes loads.
func TestLoadSecondAcceptsAKeyOfTheMinimumLength(t *testing.T) {
	env := map[string]string{
		generator.EnvSecondKey:     "abcdefgh",
		generator.EnvSecondBaseURL: "https://api.example.com/v1",
		generator.EnvSecondModel:   "some-model",
	}
	if _, connected, err := generator.LoadSecond(func(k string) string { return env[k] }); err != nil || !connected {
		t.Fatalf("connected = %v, err = %v", connected, err)
	}
}
