package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestCheckAddressJudgesTheForgeByTheNetworkMode proves the command the action asks before reviewdog posts
// (ADR-0034): in local mode an address outside the organisation's network, or a name that reviewdog would
// resolve with its own resolver, is refused with exit 7 and names the setting; loopback, private and the
// allow list pass; in public mode every address passes.
// Mutated field: Policy.CheckProgramAddress (return nil and every refusal row turns red).
func TestCheckAddressJudgesTheForgeByTheNetworkMode(t *testing.T) {
	rows := []struct {
		name    string
		env     map[string]string
		setting string
		address string
		code    int
	}{
		{"public mode, any address", nil, "GITHUB_API_URL", "https://api.github.com", exitOK},
		{"local, public IP", map[string]string{"PROVEN_NETWORK": "local"}, "GITHUB_API_URL", "https://203.0.113.7/api/v3", exitConfig},
		{"local, the default forge name", map[string]string{"PROVEN_NETWORK": "local"}, "GITHUB_API_URL", "https://api.github.com", exitConfig},
		{"local, empty address", map[string]string{"PROVEN_NETWORK": "local"}, "GITEA_ADDRESS", "", exitConfig},
		{"local, private IP", map[string]string{"PROVEN_NETWORK": "local"}, "GITHUB_API_URL", "http://10.1.2.3/api/v3", exitOK},
		{"local, loopback", map[string]string{"PROVEN_NETWORK": "local"}, "GITEA_ADDRESS", "http://127.0.0.1:3000", exitOK},
		{"local, listed IP", map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "203.0.113.7"}, "GITHUB_API_URL", "https://203.0.113.7", exitOK},
		{"local, listed name is still a name", map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "ghe.example"}, "GITHUB_API_URL", "https://ghe.example", exitConfig},
		{"a wrong mode", map[string]string{"PROVEN_NETWORK": "offline"}, "GITHUB_API_URL", "http://10.1.2.3", exitConfig},
	}
	for _, r := range rows {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"check-address", r.setting, r.address}, func(k string) string { return r.env[k] }, &stdout, &stderr, nil)
		if code != r.code {
			t.Errorf("%s: exit %d, want %d (stderr %q)", r.name, code, r.code, stderr.String())
			continue
		}
		if r.code == exitConfig && !strings.Contains(stderr.String(), r.setting) && r.env["PROVEN_NETWORK"] == "local" {
			t.Errorf("%s: the refusal does not name %s: %q", r.name, r.setting, stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"check-address", "only-one"}, nil, &stdout, &stderr, nil); code != exitFatal {
		t.Errorf("a wrong command line: exit %d, want %d", code, exitFatal)
	}
}
