package store

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
)

// TestLoadRecordsLocalRefusesAnOutsideAddress: the records writer judges the
// CouchDB address under PROVEN_NETWORK=local, as LoadCouch does, so the writing
// credential is never sent outside the organisation's network.
// Mutated field: PROVEN_COUCHDB_URL (203.0.113.9 instead of 10.0.0.5).
func TestLoadRecordsLocalRefusesAnOutsideAddress(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK":                  "local",
		"PROVEN_COUCHDB_URL":              "http://203.0.113.9:5984",
		"PROVEN_COUCHDB_RECORDS":          "proven_records",
		"PROVEN_COUCHDB_RECORDS_USER":     "writer",
		"PROVEN_COUCHDB_RECORDS_PASSWORD": "long-test-password",
	}
	lookup := func(k string) string { return env[k] }
	_, err := LoadRecords(lookup)
	if err == nil || !strings.Contains(err.Error(), "PROVEN_COUCHDB_URL") {
		t.Fatalf("LoadRecords = %v; want a refusal naming PROVEN_COUCHDB_URL", err)
	}
	env["PROVEN_COUCHDB_URL"] = "http://10.0.0.5:5984"
	if _, err := LoadRecords(lookup); err != nil {
		t.Fatalf("control: a private address must pass: %v", err)
	}
}

// TestRecorderLocalClientDialsThroughThePolicy: under PROVEN_NETWORK=local the
// writer opens its connection through the policy's dial, so a client of its own
// (and its proxy from the environment) is not used. The recorded dial is the
// proof. Mutated field: the writer's HTTP client (a plain one instead of the
// policy's), which records no dial.
func TestRecorderLocalClientDialsThroughThePolicy(t *testing.T) {
	var mu sync.Mutex
	var dials []string
	dial := func(_ context.Context, _, addr string) (net.Conn, error) {
		mu.Lock()
		dials = append(dials, addr)
		mu.Unlock()
		client, server := net.Pipe()
		server.Close()
		return client, nil
	}
	env := map[string]string{
		"PROVEN_NETWORK":                  "local",
		"PROVEN_COUCHDB_URL":              "http://10.0.0.5:5984",
		"PROVEN_COUCHDB_RECORDS":          "proven_records",
		"PROVEN_COUCHDB_RECORDS_USER":     "writer",
		"PROVEN_COUCHDB_RECORDS_PASSWORD": "long-test-password",
	}
	cfg, err := LoadRecords(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	cfg.Network = cfg.Network.WithDial(dial)
	_, _ = NewRecorder(cfg, nil).Publish(context.Background(), []EvidenceRecord{{}}) // the pipe answers nothing; the dial is the proof
	mu.Lock()
	defer mu.Unlock()
	if len(dials) == 0 || dials[0] != "10.0.0.5:5984" {
		t.Fatalf("dials = %v; want the records writer to dial 10.0.0.5:5984 through the policy", dials)
	}
}
