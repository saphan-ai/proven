package change

import (
	"strings"
	"testing"
)

// TestGitEnvironmentOmitsSecrets proves a secret set in Proven's environment
// is absent from the environment git starts with (ADR-0025). Mutated call:
// gitCommand's cmd.Env; setting it to os.Environ() turns this red.
func TestGitEnvironmentOmitsSecrets(t *testing.T) {
	const marker = "hardening-marker-5b1e"
	t.Setenv("METERLESS_RUNTIME_TOKEN", marker)
	t.Setenv("PROVEN_GUARDRAIL_KEY", marker)
	text, err := gitZText([]string{"-c", "alias.dumpenv=!env", "dumpenv"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, marker) {
		t.Fatalf("a secret reached git:\n%s", text)
	}
}
