package review

import (
	"context"
	"strings"
	"testing"
)

// TestCouchDBLiveReviewNamesCouchDBDown proves the live query path with the
// couchdb store never answers from a store and never reads as "no such row":
// the review that asks a query names couchdb down with the seam reason
// (ADR-0047, ADR-0002). The mutated field is the installed adapter, whose
// Query returns an empty result instead of the down error; that turns this
// red. No network is used: the adapter does not dial.
func TestCouchDBLiveReviewNamesCouchDBDown(t *testing.T) {
	repo, base, head := twoCommits(t)
	env := liveSecrets(map[string]string{
		"PROVEN_EVIDENCE_STORE":   "couchdb",
		"PROVEN_GUARDRAIL_MODE":   "off",
		"PROVEN_COUCHDB_URL":      "http://127.0.0.1:9",
		"PROVEN_COUCHDB_USER":     "proven",
		"PROVEN_COUCHDB_PASSWORD": "couch-test-password",
	})
	rev, err := Open(Options{LogDir: t.TempDir(), Env: env, gen: &countGen{}, testRun: quietRun{}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatalf("the review ended: %v", err)
	}
	for _, s := range res.FailedServices {
		if s.Service == "couchdb" && s.Reason == couchdbSQLReason {
			return
		}
	}
	t.Fatalf("want couchdb down with the seam reason, got %#v", res.FailedServices)
}

// TestCouchDBSQLAdapterNamesCouchDBDown proves the adapter itself returns a
// down couchdb for any SQL, so the review test above fails for the right
// cause (ADR-0047). The mutated field is the adapter's return value.
func TestCouchDBSQLAdapterNamesCouchDBDown(t *testing.T) {
	_, err := couchdbSQL{reason: couchdbSQLReason}.Query(context.Background(), "SELECT 1")
	if err == nil || !strings.Contains(err.Error(), "couchdb") {
		t.Fatalf("want a down couchdb, got %v", err)
	}
}
