package network

import "testing"

// TestLocalRefusesAZoneOnAnIPv4Address: a zone belongs to an IPv6 address only,
// so in local mode 10.1.2.3 followed by a zone is not the private address it
// starts with, and is refused; an IPv6 link-local address with a zone is still
// allowed. Mutated field: the URL's host (a zone after an IPv4 address).
func TestLocalRefusesAZoneOnAnIPv4Address(t *testing.T) {
	pol, err := Load(func(k string) string {
		if k == EnvNetwork {
			return ModeLocal
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"http://10.1.2.3%25ghe.example.com/", "http://10.1.2.3%25eth0/"} {
		if err := pol.CheckURL("GITHUB_API_URL", raw); err == nil {
			t.Errorf("CheckURL(%s) allowed an IPv4 address with a zone", raw)
		}
		if err := pol.CheckProgramAddress("GITHUB_API_URL", raw); err == nil {
			t.Errorf("CheckProgramAddress(%s) allowed an IPv4 address with a zone", raw)
		}
	}
	if err := pol.CheckProgramAddress("GITHUB_API_URL", "http://[fe80::1%25eth0]/"); err != nil {
		t.Errorf("control: an IPv6 link-local address with a zone must be allowed: %v", err)
	}
}
