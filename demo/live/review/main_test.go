package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
)

// TestRedactTextHidesAConfiguredSecret proves the text this command prints has each configured
// value replaced by a marker naming its variable (ADR-0005, ADR-0013), in the raw and the %q
// spelling, and that text with no secret in it is unchanged.
func TestRedactTextHidesAConfiguredSecret(t *testing.T) {
	const secret = "demo-\"key\"-0123456789"
	env := func(k string) string {
		if k == generator.EnvToken {
			return secret
		}
		return ""
	}
	for name, text := range map[string]string{
		"raw": "refused for " + secret,
		"%q":  fmt.Sprintf("refused for %q", secret),
	} {
		got := redactText(text, env)
		if strings.Contains(got, "0123456789") || !strings.Contains(got, "[redacted: "+generator.EnvToken+"]") {
			t.Errorf("%s: %q", name, got)
		}
	}
	if got := redactText("nothing to hide", env); got != "nothing to hide" {
		t.Errorf("text with no secret was changed: %q", got)
	}
}
