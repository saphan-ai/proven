package settingspage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/envdoc"
)

func TestConfigurationPageMatchesTheRegistry(t *testing.T) {
	if err := envdoc.Check(Catalog()); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(configPath(t))
	if err != nil {
		t.Fatal(err)
	}
	rendered := envdoc.Render(Catalog())
	interior, err := envdoc.Extract(string(body))
	if err != nil {
		t.Fatalf("docs/CONFIGURATION.md has no generated settings block (%s); run: go run ./cmd/settings-doc", err)
	}
	if interior == rendered {
		return
	}
	name := envdoc.DriftSetting(interior, rendered)
	if name == "" {
		t.Fatalf("docs/CONFIGURATION.md drifted outside a setting row; run: go run ./cmd/settings-doc")
	}
	t.Fatalf("docs/CONFIGURATION.md drifted on %s; run: go run ./cmd/settings-doc", name)
}

func TestSecretFlagMatchesTheDescription(t *testing.T) {
	for _, sec := range Catalog() {
		for _, row := range sec.Rows {
			mentions := strings.Contains(row.Description, "Secret")
			if row.Secret != mentions {
				t.Fatalf("%s secret=%v, description mentions Secret=%v", row.Name, row.Secret, mentions)
			}
		}
	}
}

func configPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "docs", "CONFIGURATION.md")
}
