package netguard_test

import (
	"net/http"
	"testing"

	"github.com/saphan-ai/proven/internal/netguard"
)

func TestGuardDialIsVisible(t *testing.T) {
	g := &netguard.Guard{}
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Transport().RoundTrip(req); err == nil {
		t.Fatal("a dial must fail")
	}
	if g.Dials() == 0 {
		t.Fatal("the guard did not see the dial")
	}
}
