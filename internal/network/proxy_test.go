package network_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/network"
)

// TestLocalClientNeverDialsAProxyFromTheEnvironment proves the claim of docs/AIR_GAPPED.md that a proxy
// set in the environment is not used in local mode: the only dial is to the address asked for.
// Mutated field: the Proxy of the local transport (http.ProxyFromEnvironment instead of none).
func TestLocalClientNeverDialsAProxyFromTheEnvironment(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://10.9.9.9:3128")
	t.Setenv("HTTP_PROXY", "http://10.9.9.9:3128")
	t.Setenv("NO_PROXY", "")
	p, err := network.Load(func(k string) string {
		if k == "PROVEN_NETWORK" {
			return "local"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var seen []string
	p = p.WithDial(func(_ context.Context, _, addr string) (net.Conn, error) {
		mu.Lock()
		seen = append(seen, addr)
		mu.Unlock()
		return nil, errors.New("no network in this test")
	})
	c := p.NewClient(2 * time.Second)
	_, _ = c.Get("http://10.0.0.7:8080/v1/models")
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 1 || seen[0] != "10.0.0.7:8080" {
		t.Fatalf("dials %v; want only 10.0.0.7:8080", seen)
	}
}
