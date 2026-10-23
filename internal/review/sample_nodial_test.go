package review_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/review"
)

// TestSampleReviewDialsNothing proves a full sample review opens no network
// connection and builds no live client.
//
// It can catch a dial made through http.DefaultTransport or through
// net.DefaultResolver, and a call that constructs a live generator, query
// or guardrail client. It cannot catch a dial on a private Transport or
// Resolver, a direct net.Dial, or a send that never asks those two.
func TestSampleReviewDialsNothing(t *testing.T) {
	prevTransport := http.DefaultTransport
	prevResolver := net.DefaultResolver
	http.DefaultTransport = dialFail{t: t}
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(context.Context, string, string) (net.Conn, error) {
			t.Error("sample review resolved a name")
			return nil, errors.New("dial")
		},
	}
	t.Cleanup(func() {
		http.DefaultTransport = prevTransport
		net.DefaultResolver = prevResolver
	})
	beforeGen := generator.LiveBuilt()
	beforeQuery := query.LiveBuilt()
	beforeRail := guardrail.LiveBuilt()

	rev := openSample(t, t.TempDir())
	if _, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)}); err != nil {
		t.Fatal(err)
	}
	if generator.LiveBuilt() != beforeGen || query.LiveBuilt() != beforeQuery || guardrail.LiveBuilt() != beforeRail {
		t.Fatalf("live clients built: generator %d query %d guardrail %d",
			generator.LiveBuilt()-beforeGen, query.LiveBuilt()-beforeQuery, guardrail.LiveBuilt()-beforeRail)
	}
}

// dialFail is an HTTP transport whose every request fails the test.
type dialFail struct{ t *testing.T }

// RoundTrip records that a sample review tried to dial.
func (d dialFail) RoundTrip(*http.Request) (*http.Response, error) {
	d.t.Error("sample review dialed")
	return nil, errors.New("dial")
}
