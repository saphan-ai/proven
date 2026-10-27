package network

import (
	"context"
	"net"
	"strings"
	"testing"
)

func localPolicyAllowing(t *testing.T, allow string) *Policy {
	t.Helper()
	p, err := Load(func(k string) string {
		switch k {
		case EnvNetwork:
			return ModeLocal
		case EnvAllow:
			return allow
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// recordDial returns a policy whose every connection is recorded, none opened.
func recordDial(p *Policy) (*Policy, *[]string) {
	var seen []string
	return p.WithDial(func(_ context.Context, _, addr string) (net.Conn, error) {
		seen = append(seen, addr)
		client, server := net.Pipe()
		server.Close()
		return client, nil
	}), &seen
}

// A name server outside the allowed set is refused by name and address, and no
// connection is made to it; a loopback, a private and an allow-listed one are used.
func TestLocalResolverDialsOnlyAnAllowedNameServer(t *testing.T) {
	for _, tc := range []struct {
		name, server string
		allowed      bool
	}{
		{"public", "8.8.8.8:53", false},
		{"public ipv6", "[2001:4860:4860::8888]:53", false},
		{"public name", "dns.example.net:53", false},
		{"loopback", "127.0.0.53:53", true},
		{"private", "10.1.2.3:53", true},
		{"allow listed name", "dns.corp.example:53", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, seen := recordDial(localPolicyAllowing(t, "dns.corp.example"))
			conn, err := p.resolverDial(context.Background(), "udp", tc.server)
			if conn != nil {
				conn.Close()
			}
			if tc.allowed {
				if err != nil {
					t.Fatalf("allowed name server refused: %v", err)
				}
				if len(*seen) != 1 || (*seen)[0] != tc.server {
					t.Fatalf("dials = %v, want [%s]", *seen, tc.server)
				}
				return
			}
			if err == nil {
				t.Fatalf("outside name server %s was dialled", tc.server)
			}
			if !strings.Contains(err.Error(), tc.server) || !strings.Contains(err.Error(), "name server is outside") {
				t.Fatalf("refusal does not name the resolver: %v", err)
			}
			if len(*seen) != 0 {
				t.Fatalf("a connection was opened to %v", *seen)
			}
		})
	}
}

// The resolver that local mode builds is the pure-Go one and dials through the
// check: its Dial refuses a public server.
func TestLocalLookupUsesTheCheckedResolver(t *testing.T) {
	r := localPolicyAllowing(t, "").checkedResolver()
	if !r.PreferGo || r.Dial == nil {
		t.Fatalf("local resolver is not the checked pure-Go one: PreferGo=%v Dial nil=%v", r.PreferGo, r.Dial == nil)
	}
	if _, err := r.Dial(context.Background(), "udp", "9.9.9.9:53"); err == nil {
		t.Fatal("the local resolver dialled a public name server")
	}
}
