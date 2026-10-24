package testrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A private IP in an ssh:// daemon address does not say where ssh connects: the
// ssh configuration under HOME can send the route through a jump host outside the
// allowed set. Local mode refuses an ssh daemon whatever its host is.
func TestLocalRefusesSSHDaemonWithOutsideProxyJump(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	config := "Host 10.0.0.7\n  ProxyJump jump.example.net\n"
	if err := os.WriteFile(filepath.Join(home, ".ssh", "config"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"HOME":           home,
		"PROVEN_NETWORK": "local",
		"DOCKER_HOST":    "ssh://builder@10.0.0.7",
	}
	_, err := Load(func(k string) string { return env[k] })
	if err == nil || !strings.Contains(err.Error(), "DOCKER_HOST") {
		t.Fatalf("local mode must refuse an SSH daemon whose SSH route leaves the allowed network; got %v", err)
	}
}

// Every way a daemon address reaches ssh is refused in local mode, with the
// setting named: both variables and the docker context. Outside local mode ssh is
// untouched, and tcp:// to a private IP stays allowed.
func TestLocalRefusesEverySSHDaemonRoute(t *testing.T) {
	for _, name := range []string{"DOCKER_HOST", "CONTAINER_HOST"} {
		for _, raw := range []string{"ssh://builder@10.0.0.7", "SSH://10.0.0.7:2222", "ssh://127.0.0.1"} {
			_, err := Load(localEnv(map[string]string{name: raw}))
			if err == nil || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), "ssh") {
				t.Errorf("%s=%s: %v", name, raw, err)
			}
		}
	}
	dir := writeContext(t, "remote", "ssh://builder@10.0.0.7")
	if _, err := Load(localEnv(map[string]string{"DOCKER_CONFIG": dir})); err == nil {
		t.Error("a docker context over ssh was accepted")
	}
	if _, err := Load(func(k string) string {
		if k == "DOCKER_HOST" {
			return "ssh://builder@10.0.0.7"
		}
		return ""
	}); err != nil {
		t.Errorf("public mode refused ssh: %v", err)
	}
	if _, err := Load(localEnv(map[string]string{"DOCKER_HOST": "tcp://10.0.0.7:2376"})); err != nil {
		t.Errorf("tcp to a private IP refused: %v", err)
	}
}
