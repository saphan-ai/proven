package network_test

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/network"
)

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// TestLoadRefusesAWrongModeByName proves a mode that is neither public nor local is refused,
// naming the setting, the value and what is allowed (ADR-0013).
func TestLoadRefusesAWrongModeByName(t *testing.T) {
	// Mutated field: PROVEN_NETWORK.
	_, err := network.Load(envOf(map[string]string{"PROVEN_NETWORK": "offline"}))
	ce, ok := network.As(err)
	if !ok || ce.Field != "PROVEN_NETWORK" || ce.Value != "offline" || len(ce.Allowed) != 1 || !strings.Contains(ce.Allowed[0], "public") || !strings.Contains(ce.Allowed[0], "local") {
		t.Fatalf("refusal %v", err)
	}
}

// TestLoadDefaultsToPublic proves an empty lookup is today's behaviour: no policy, nothing refused.
func TestLoadDefaultsToPublic(t *testing.T) {
	for _, env := range []func(string) string{nil, envOf(nil), envOf(map[string]string{"PROVEN_NETWORK": "public"})} {
		p, err := network.Load(env)
		if err != nil || p.Local() {
			t.Fatalf("policy %v err %v", p, err)
		}
		if err := p.CheckURL("ZETARIS_REST_URL", "https://integrate.api.nvidia.com/v1"); err != nil {
			t.Fatalf("public mode refused an address: %v", err)
		}
	}
}

// TestAllowListIsValidatedAtLoad proves a malformed allow list is refused by name, also in public mode,
// so a typo is caught before the mode is switched on.
func TestAllowListIsValidatedAtLoad(t *testing.T) {
	for _, bad := range []string{"https://a.example", "a.example:8080", "a.example/path", "a b", "a.example,,b.example", "[::1]"} {
		for _, mode := range []string{"public", "local"} {
			// Mutated field: PROVEN_NETWORK_ALLOW.
			_, err := network.Load(envOf(map[string]string{"PROVEN_NETWORK": mode, "PROVEN_NETWORK_ALLOW": bad}))
			ce, ok := network.As(err)
			if !ok || ce.Field != "PROVEN_NETWORK_ALLOW" {
				t.Errorf("%s %q: %v", mode, bad, err)
			}
		}
	}
}

// TestCheckURLJudgesTheHostByItsText proves which hosts are inside the network in local mode and that
// a name is never resolved: a name off the allow list is refused whatever it would resolve to.
func TestCheckURLJudgesTheHostByItsText(t *testing.T) {
	p, err := network.Load(envOf(map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "Nightshift.Corp.Example, 203.0.113.9"}))
	if err != nil || !p.Local() {
		t.Fatal(err)
	}
	cases := []struct {
		url   string
		allow bool
	}{
		{"http://127.0.0.1:8088", true},
		{"http://127.1.2.3", true},
		{"http://localhost:8081", true},
		{"http://LOCALHOST:8081", true},
		{"http://[::1]:8081", true},
		{"http://10.0.0.5", true},
		{"http://172.16.9.1", true},
		{"http://172.31.255.255", true},
		{"http://192.168.1.1", true},
		{"http://[fd12:3456::1]", true},
		{"http://169.254.10.10", true},
		{"http://[fe80::1]", true},
		{"http://[::ffff:10.0.0.5]", true},
		{"http://nightshift.corp.example:8088", true},
		{"http://203.0.113.9", true},
		{"http://172.32.0.1", false},
		{"http://172.15.0.1", false},
		{"http://8.8.8.8", false},
		{"http://[2606:4700::1111]", false},
		{"http://[::ffff:8.8.8.8]", false},
		{"https://integrate.api.nvidia.com/v1", false},
		// A name that is not on the list is refused even if it points at a private address.
		{"http://intranet.corp.example", false},
		{"http://localhost.evil.example", false},
	}
	for _, tc := range cases {
		err := p.CheckURL("ZETARIS_REST_URL", tc.url)
		if tc.allow && err != nil {
			t.Errorf("%s refused: %v", tc.url, err)
		}
		if !tc.allow {
			ce, ok := network.As(err)
			u, _ := url.Parse(tc.url)
			if !ok || ce.Field != "ZETARIS_REST_URL" || !strings.Contains(ce.Reason, u.Hostname()) {
				t.Errorf("%s: refusal %v", tc.url, err)
			}
		}
	}
}

// TestRefuseNamesTheValueInLocalOnly proves the hosted guardrail mode is refused in local mode and
// left alone in public mode.
func TestRefuseNamesTheValueInLocalOnly(t *testing.T) {
	local, _ := network.Load(envOf(map[string]string{"PROVEN_NETWORK": "local"}))
	if ce, ok := network.As(local.Refuse("PROVEN_GUARDRAIL_MODE", "nvidia", "hosted")); !ok || ce.Field != "PROVEN_GUARDRAIL_MODE" || ce.Value != "nvidia" {
		t.Fatalf("local did not refuse: %v", ce)
	}
	var public *network.Policy
	if err := public.Refuse("PROVEN_GUARDRAIL_MODE", "nvidia", "hosted"); err != nil {
		t.Fatal(err)
	}
}

// TestDialRefusesWhatTheLoadCheckWouldRefuse proves the dial check on its own: a public address, or a
// name off the allow list, never reaches the dialler, however the address came to be asked for.
func TestDialRefusesWhatTheLoadCheckWouldRefuse(t *testing.T) {
	var seen []string
	p, err := network.Load(envOf(map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "nightshift.corp.example"}))
	if err != nil {
		t.Fatal(err)
	}
	p = p.WithDial(func(_ context.Context, _, addr string) (net.Conn, error) {
		seen = append(seen, addr)
		return nil, errors.New("reached the dialler")
	})
	for _, addr := range []string{"8.8.8.8:443", "integrate.api.nvidia.com:443", "[2606:4700::1111]:443", "intranet.corp.example:80", "no-port"} {
		_, err := p.DialContext(context.Background(), "tcp", addr)
		if err == nil || strings.Contains(err.Error(), "reached the dialler") || !strings.Contains(err.Error(), "refused") {
			t.Errorf("%s: %v", addr, err)
		}
	}
	if len(seen) != 0 {
		t.Fatalf("a refused address reached the dialler: %v", seen)
	}
	for _, addr := range []string{"127.0.0.1:8088", "10.1.2.3:443", "nightshift.corp.example:8088", "localhost:8081", "[fd00::1]:80"} {
		_, err := p.DialContext(context.Background(), "tcp", addr)
		if err == nil || !strings.Contains(err.Error(), "reached the dialler") {
			t.Errorf("%s: %v", addr, err)
		}
	}
	if len(seen) != 5 {
		t.Fatalf("allowed dials %v", seen)
	}
}

// TestNilPolicyIsPublic proves a Settings built by hand, with no policy, keeps today's client.
func TestNilPolicyIsPublic(t *testing.T) {
	var p *network.Policy
	c := p.NewClient(0)
	if c.Transport != nil {
		t.Fatal("public mode installed a transport")
	}
	if p.WithDial(nil) != nil {
		t.Fatal("WithDial invented a policy")
	}
}
