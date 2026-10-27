package network_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/saphan-ai/proven/internal/network"
)

// queryCounter is a resolver that sends nothing: every query the Go resolver would write to a name
// server first calls its Dial, which counts and refuses. A count above zero is a DNS query attempted.
func queryCounter() (*net.Resolver, *atomic.Int32) {
	var n atomic.Int32
	return &net.Resolver{PreferGo: true, Dial: func(context.Context, string, string) (net.Conn, error) {
		n.Add(1)
		return nil, errors.New("no network in this test")
	}}, &n
}

func localPolicy(t *testing.T, r *net.Resolver) *network.Policy {
	t.Helper()
	p, err := network.Load(envOf(map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "nightshift.corp.example"}))
	if err != nil {
		t.Fatal(err)
	}
	return p.WithResolver(r)
}

// TestLocalNeverQueriesDNSForAPublicName proves ADR-0034 point 1: a name off the allow list is refused
// before any DNS query, by the lookup and by the dialer alike. The same policy asked for an allowed
// name does send a query, so the counter can see one.
// Mutated field: the name check of Lookup; with it removed the public names reach the resolver.
func TestLocalNeverQueriesDNSForAPublicName(t *testing.T) {
	r, queries := queryCounter()
	p := localPolicy(t, r)
	ctx := context.Background()
	for _, name := range []string{"integrate.api.nvidia.com", "localhost.evil.example", "intranet.corp.example"} {
		if _, err := p.Lookup(ctx, name); err == nil || !strings.Contains(err.Error(), "outside the organisation's network") {
			t.Errorf("lookup of %s: %v", name, err)
		}
		if _, err := p.DialContext(ctx, "tcp", net.JoinHostPort(name, "443")); err == nil || !strings.Contains(err.Error(), "outside the organisation's network") {
			t.Errorf("dial of %s: %v", name, err)
		}
	}
	if n := queries.Load(); n != 0 {
		t.Fatalf("%d DNS queries were attempted for names off the list", n)
	}
	// The false world: an allowed name is looked up, so a counter of zero above means something.
	_, _ = p.Lookup(ctx, "nightshift.corp.example")
	if queries.Load() == 0 {
		t.Fatal("the counter saw no query for an allowed name, so it measures nothing")
	}
}

// TestLocalAnswersLocalhostAndLiteralsWithoutAQuery proves the names that need no resolver get none:
// "localhost" (also with a trailing dot) answers loopback, and an IP literal answers itself.
func TestLocalAnswersLocalhostAndLiteralsWithoutAQuery(t *testing.T) {
	r, queries := queryCounter()
	p := localPolicy(t, r)
	for host, want := range map[string]string{"localhost": "127.0.0.1", "LOCALHOST.": "127.0.0.1", "10.1.2.3": "10.1.2.3", "::1": "::1"} {
		ips, err := p.Lookup(context.Background(), host)
		if err != nil || len(ips) == 0 || ips[0].String() != want {
			t.Errorf("%s: %v %v", host, ips, err)
		}
	}
	if n := queries.Load(); n != 0 {
		t.Fatalf("%d queries for localhost and literals", n)
	}
	for _, public := range []string{"8.8.8.8", "2001:4860:4860::8888"} {
		if _, err := p.Lookup(context.Background(), public); err == nil {
			t.Errorf("a public literal %s was looked up", public)
		}
	}
}
