package childenv

import (
	"slices"
	"testing"
)

// TestMinimalFromPassesOnlyNamedVariables proves the child environment holds
// the base variables and the extras the caller names, and nothing else, so a
// secret in the lookup is absent unless a caller names it.
func TestMinimalFromPassesOnlyNamedVariables(t *testing.T) {
	all := map[string]string{
		"PATH": "/bin", "HOME": "/home/x", "DOCKER_HOST": "unix:///d.sock",
		"METERLESS_RUNTIME_TOKEN": "secret",
	}
	lookup := func(k string) (string, bool) { v, ok := all[k]; return v, ok }
	got := MinimalFrom(lookup, "DOCKER_HOST", "NOT_SET")
	want := []string{"PATH=/bin", "HOME=/home/x", "DOCKER_HOST=unix:///d.sock"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
