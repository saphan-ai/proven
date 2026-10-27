package testrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The runtime reads the whole configuration file, so local mode never judges a
// prefix of it: a destination past the read limit would reach the runtime unchecked.
func TestLocalRefusesUnreadPodmanConfigTail(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "containers")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	config := "[engine]\nremote = true\nactive_service = \"far\"\n" + strings.Repeat("# padding\n", 120000) +
		"[engine.service_destinations.far]\nuri = \"ssh://builder@203.0.113.8/run/podman/podman.sock\"\n"
	if err := os.WriteFile(filepath.Join(dir, "containers.conf"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman", "HOME": home}
	if _, err := Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("local mode accepted a podman configuration whose remote destination was beyond the bytes it checked")
	}
}

// A file at the limit is still read and judged; only a larger one is refused.
func TestReadSmallKeepsAFileAtTheLimit(t *testing.T) {
	dir := t.TempDir()
	at := filepath.Join(dir, "at")
	over := filepath.Join(dir, "over")
	if err := os.WriteFile(at, make([]byte, maxDaemonFile), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(over, make([]byte, maxDaemonFile+1), 0600); err != nil {
		t.Fatal(err)
	}
	if b, err := readSmall(at); err != nil || len(b) != maxDaemonFile {
		t.Fatalf("a file of exactly %d bytes: %d bytes, %v", maxDaemonFile, len(b), err)
	}
	if _, err := readSmall(over); err == nil {
		t.Fatalf("a file of %d bytes was read as a prefix instead of refused", maxDaemonFile+1)
	}
}
