package change

import (
	"strings"
	"testing"
)

// TestGitEnvironmentOmitsSecrets proves a secret set in Proven's environment
// is absent from the environment git starts with (ADR-0025). Mutated call:
// GitCommand's cmd.Env; setting it to os.Environ() turns this red.
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
	// Positive control: git must have printed the base variables, so an empty
	// output cannot pass for a clean one. git adds its own GIT_* names, so the
	// check is on Proven's names, not on the whole set.
	names := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		if name, _, ok := strings.Cut(line, "="); ok {
			names[name] = true
		}
	}
	for _, want := range []string{"PATH", "HOME"} {
		if !names[want] {
			t.Errorf("git's environment lacks %s:\n%s", want, text)
		}
	}
	for _, banned := range []string{"METERLESS_RUNTIME_TOKEN", "PROVEN_GUARDRAIL_KEY"} {
		if names[banned] {
			t.Errorf("git's environment carries %s", banned)
		}
	}
}
