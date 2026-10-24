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
// setting and its default.
func TestEnvironmentNamesTheTimeout(t *testing.T) {
	rows := Environment()
	if len(rows) != 1 || rows[0].Name != EnvTimeout || rows[0].Default != "`120s`" {
		t.Fatalf("%#v", rows)
	}
}
