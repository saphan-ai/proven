package testrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The mutated field is the location of a configured Podman service: a
// system drop-in alongside a system containers.conf rather than under HOME.
func TestLocalRefusesSystemPodmanDropIn(t *testing.T) {
	system := t.TempDir()
	dropins := filepath.Join(system, "containers.conf.d")
	if err := os.Mkdir(dropins, 0o700); err != nil {
		t.Fatal(err)
	}
	conf := filepath.Join(system, "containers.conf")
	remote := "[engine.service_destinations.far]\nuri = 'tcp://203.0.113.8:2375'\n"
	if err := os.WriteFile(filepath.Join(dropins, "10-remote.conf"), []byte(remote), 0o600); err != nil {
		t.Fatal(err)
	}
	saved := podmanSystemConfs
	podmanSystemConfs = []string{conf}
	defer func() { podmanSystemConfs = saved }()
	values := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": t.TempDir()}
	if err := os.WriteFile(conf, []byte(remote), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(func(key string) string { return values[key] }); err == nil {
		t.Fatal("positive control: local mode did not read the system main file")
	}
	if err := os.WriteFile(conf, []byte("[engine]\nremote = true\nactive_service = 'far'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(func(key string) string { return values[key] })
	if err == nil || !strings.Contains(err.Error(), "203.0.113.8") {
		t.Fatalf("local mode did not refuse a system Podman drop-in: %v", err)
	}
}
