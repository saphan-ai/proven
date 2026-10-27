package testrun

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// confHome writes body as the user's containers.conf under a new home and returns the home.
func confHome(t *testing.T, body string) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "containers.conf"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}

func loadPodman(home string) error {
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": home}
	_, err := Load(func(k string) string { return env[k] })
	return err
}

// tomlForms are spellings of one service destination that TOML makes the same document. Each is
// written once with an address outside the network and once with an address inside it.
var tomlForms = map[string]string{
	"dotted key":        "[engine]\nremote = true\nactive_service = \"far\"\nservice_destinations.far.uri = \"tcp://%s:2375\"\n",
	"fully dotted key":  "engine.remote = true\nengine.active_service = \"far\"\nengine.service_destinations.far.uri = \"tcp://%s:2375\"\n",
	"quoted table name": "[engine]\nremote = true\nactive_service = \"far\"\n[engine.\"service_destinations\".far]\nuri = \"tcp://%s:2375\"\n",
	"quoted key":        "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\n\"uri\" = \"tcp://%s:2375\"\n",
	"inline table":      "[engine]\nremote = true\nactive_service = \"far\"\nservice_destinations = { far = { uri = \"tcp://%s:2375\" } }\n",
	"literal string":    "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\nuri = 'tcp://%s:2375'\n",
	"escape in key":     "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\n\"\\u0075ri\" = \"tcp://%s:2375\"\n",
	"other case":        "[engine]\nremote = true\nactive_service = \"far\"\n[engine.service_destinations.far]\nURI = \"tcp://%s:2375\"\n",
	"string with uri=":  "[engine]\nremote = true\nactive_service = \"uri=tcp://10.0.0.5:2375\"\n[engine.service_destinations.far]\nuri = \"tcp://%s:2375\"\n",
}

// TestLocalRefusesAPodmanDestinationInEveryTOMLForm: podman reads containers.conf as TOML, so a
// destination outside the network is refused however the document spells it (dotted key, quoted table
// name, quoted key, inline table, literal string, an escape in a key). Mutated field: the spelling of
// the destination; the address is the same in each. The same spelling with 10.0.0.5 must load, so the
// test shows the form is read and not refused for being unusual.
func TestLocalRefusesAPodmanDestinationInEveryTOMLForm(t *testing.T) {
	for name, form := range tomlForms {
		t.Run(name, func(t *testing.T) {
			far := strings.ReplaceAll(form, "%s", "203.0.113.8")
			err := loadPodman(confHome(t, far))
			var ce *ConfigError
			if !errors.As(err, &ce) || !strings.Contains(ce.Error(), "203.0.113.8") {
				t.Fatalf("local mode loaded a podman destination 203.0.113.8 written as %s: %v", name, err)
			}
			near := strings.ReplaceAll(form, "%s", "10.0.0.5")
			if err := loadPodman(confHome(t, near)); err != nil {
				t.Fatalf("a podman destination 10.0.0.5 written as %s was refused: %v", name, err)
			}
		})
	}
}

// TestLocalRefusesAPodmanConfigThatIsNotTOML: a file Proven cannot parse is a file whose destinations it
// cannot judge, so Load refuses it by name. Mutated field: the syntax of the file.
func TestLocalRefusesAPodmanConfigThatIsNotTOML(t *testing.T) {
	for name, body := range map[string]string{
		"unterminated string": "[engine]\nactive_service = \"far\n",
		"missing equals":      "[engine]\nremote true\n",
		"duplicate key":       "[engine]\nremote = true\nremote = false\n",
		"duplicate table":     "[engine]\nremote = true\n[engine]\nremote = false\n",
	} {
		t.Run(name, func(t *testing.T) {
			err := loadPodman(confHome(t, body))
			var ce *ConfigError
			if !errors.As(err, &ce) || !strings.Contains(ce.Error(), "TOML") {
				t.Fatalf("local mode loaded a podman configuration that is %s: %v", name, err)
			}
		})
	}
}

// TestLocalRefusesAPodmanDestinationOfAnUnreadableShape: a destination the parsed document gives in a
// shape podman does not read as an address is refused, not skipped. Mutated field: the type of the value.
func TestLocalRefusesAPodmanDestinationOfAnUnreadableShape(t *testing.T) {
	for name, body := range map[string]string{
		"uri is a number":        "[engine.service_destinations.far]\nuri = 5\n",
		"uri is an array":        "[engine.service_destinations.far]\nuri = [\"tcp://203.0.113.8:2375\"]\n",
		"destination is a value": "[engine]\nservice_destinations = \"tcp://203.0.113.8:2375\"\n",
		"remote is a string":     "[engine]\nremote = \"yes\"\n",
		"engine is a value":      "engine = \"tcp://203.0.113.8:2375\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			if err := loadPodman(confHome(t, body)); err == nil {
				t.Fatalf("local mode loaded a podman configuration where %s", name)
			}
		})
	}
}

// TestLocalReadsEveryDropInAsTOML: a drop-in file is parsed the same way as the main file.
// Mutated field: the file that holds the quoted key (a drop-in, not containers.conf).
func TestLocalReadsEveryDropInAsTOML(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers", "containers.conf.d")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "[engine.service_destinations.far]\n\"uri\" = \"tcp://203.0.113.8:2375\"\n"
	if err := os.WriteFile(filepath.Join(dir, "10-far.conf"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadPodman(home); err == nil {
		t.Fatal("local mode loaded a drop-in whose quoted uri key names 203.0.113.8")
	}
}
