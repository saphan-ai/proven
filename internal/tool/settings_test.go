package tool

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestLoadUsesTheDefaultAndRefusesABadTimeout proves an empty setting is two
// minutes and a value that is set wrong is refused by name (ADR-0013).
// Mutated field: PROVEN_TOOL_TIMEOUT.
func TestLoadUsesTheDefaultAndRefusesABadTimeout(t *testing.T) {
	d, err := Load(func(string) string { return "" })
	if err != nil || d != 2*time.Minute {
		t.Fatalf("%s %v", d, err)
	}
	d, err = Load(func(key string) string {
		if key == EnvTimeout {
			return "30s"
		}
		return ""
	})
	if err != nil || d != 30*time.Second {
		t.Fatalf("%s %v", d, err)
	}
	for _, bad := range []string{"0", "-1s", "soon", "0s"} {
		_, err := Load(func(string) string { return bad })
		var cfg *ConfigError
		if !errors.As(err, &cfg) || cfg.Field != EnvTimeout || !strings.Contains(cfg.Error(), bad) {
			t.Fatalf("%q: %v", bad, err)
		}
	}
}

// TestEnvironmentNamesTheTimeout proves the configuration page can list the
// tool settings and their defaults.
func TestEnvironmentNamesTheTimeout(t *testing.T) {
	rows := Environment()
	if len(rows) != 2 || rows[0].Name != EnvTimeout || rows[0].Default != "`120s`" || rows[1].Name != EnvDB || rows[1].Default != "`empty`" {
		t.Fatalf("%#v", rows)
	}
}

// TestLoadDBAcceptsEmptyAndACheckoutFile proves an empty database means the
// tool is not run, a file URL under the checkout is kept, and a network
// address is refused by name (ADR-0013, ADR-0043).
// Mutated field: PROVEN_GOVULNCHECK_DB.
func TestLoadDBAcceptsEmptyAndACheckoutFile(t *testing.T) {
	got, err := LoadDB(func(string) string { return "" })
	if err != nil || got != "" {
		t.Fatalf("%q %v", got, err)
	}
	got, err = LoadDB(func(string) string { return "file:///checkout/vulndb" })
	if err != nil || got != "file:///checkout/vulndb" {
		t.Fatalf("%q %v", got, err)
	}
	for _, bad := range []string{"https://vuln.go.dev", "file:///etc/passwd", "file:///checkout/../etc/passwd", "not-a-url"} {
		_, err := LoadDB(func(string) string { return bad })
		var cfg *ConfigError
		if !errors.As(err, &cfg) || cfg.Field != EnvDB || !strings.Contains(cfg.Error(), bad) {
			t.Fatalf("%q: %v", bad, err)
		}
	}
}
