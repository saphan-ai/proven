package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
)

// TestRefusedOrgIDIsNotPrinted proves that livecheck names a refused organisation id's variable and
// never prints its value.
// Mutated field: ZETARIS_ORG_ID, set to a value that is not an organisation id.
func TestRefusedOrgIDIsNotPrinted(t *testing.T) {
	const marker = "1234567890123456789"
	env := func(k string) string {
		switch k {
		case query.EnvRESTURL:
			return "http://127.0.0.1:9"
		case query.EnvAPIKey:
			return "an-api-key-value"
		case query.EnvOrgID:
			return marker
		}
		return ""
	}
	var out, errout bytes.Buffer
	_ = run(nil, &out, &errout, env)
	if strings.Contains(out.String()+errout.String(), marker) {
		t.Fatal("the organisation id reached terminal output")
	}
	if !strings.Contains(out.String(), query.EnvOrgID) {
		t.Fatalf("the refusal does not name the variable:\n%s", out.String())
	}
}

// TestOrgIDEchoedByAServiceIsNotPrinted proves that an organisation id a service echoes in a routed
// model is not printed.
// Mutated field: the routed model of the generator's answer, carrying the organisation id.
func TestOrgIDEchoedByAServiceIsNotPrinted(t *testing.T) {
	const marker = "123456789012345678"
	w := newWorld(t, "ready")
	w.model = "nvidia/" + marker
	var out, errout bytes.Buffer
	_ = run(nil, &out, &errout, w.envWith(map[string]string{query.EnvOrgID: marker}))
	if strings.Contains(out.String()+errout.String(), marker) {
		t.Fatal("the organisation id echoed by a service reached terminal output")
	}
}
