package store

import (
	"strings"
	"testing"
)

// TestLoadCouchLocalRefusesAnOutsideAddress: LoadCouch itself judges the address
// under PROVEN_NETWORK=local, so a caller that skips Open still cannot build a
// store for a host outside the organisation's network.
// Mutated field: PROVEN_COUCHDB_URL (203.0.113.9 instead of 10.0.0.5).
func TestLoadCouchLocalRefusesAnOutsideAddress(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK":          "local",
		"PROVEN_COUCHDB_URL":      "http://203.0.113.9:5984",
		"PROVEN_COUCHDB_USER":     "reviewer",
		"PROVEN_COUCHDB_PASSWORD": "long-test-password",
	}
	lookup := func(k string) string { return env[k] }
	_, err := LoadCouch(lookup)
	if err == nil || !strings.Contains(err.Error(), "PROVEN_COUCHDB_URL") {
		t.Fatalf("LoadCouch = %v; want a refusal naming PROVEN_COUCHDB_URL", err)
	}
	env["PROVEN_COUCHDB_URL"] = "http://10.0.0.5:5984"
	if _, err := LoadCouch(lookup); err != nil {
		t.Fatalf("control: a private address must pass: %v", err)
	}
}
