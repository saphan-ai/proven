package review

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/store"
)

// recordedCouch is a CouchDB that lives in the test: it answers the requests a
// records writer and the viewer check make, from memory, through a
// RoundTripper, so no listener is opened. It keeps the paths it was asked for.
type recordedCouch struct {
	mu    sync.Mutex
	paths []string
	page  []byte
	bad   bool
}

func (c *recordedCouch) RoundTrip(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths = append(c.paths, req.Method+" "+req.URL.Path)
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	switch {
	case c.bad:
		rec.WriteHeader(http.StatusInternalServerError)
	case req.Method == http.MethodPut && strings.Count(req.URL.Path, "/") == 1:
		rec.WriteHeader(http.StatusCreated)
	case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/_design/evidence"):
		rec.WriteHeader(http.StatusNotFound)
	case req.Method == http.MethodPut:
		rec.WriteHeader(http.StatusCreated)
	case strings.HasSuffix(req.URL.Path, "/_bulk_docs"):
		var body struct {
			Docs []map[string]any `json:"docs"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		var out []map[string]any
		for _, d := range body.Docs {
			out = append(out, map[string]any{"id": d["_id"], "ok": true})
		}
		rec.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rec).Encode(out)
	case strings.HasSuffix(req.URL.Path, "/viewer.html"):
		_, _ = rec.Write(c.page)
	}
	return rec.Result(), nil
}

func sampleFindings() []Finding {
	return []Finding{
		{
			ID: "f1", RuleID: "dependency", Title: "Known exploited", Severity: SeverityError, ProofLevel: ProvedByData,
			Location: Location{Path: "go.mod", StartLine: 3},
			Evidence: []Evidence{{Kind: EvidenceQuery, Query: &QueryEvidence{
				SQL: `{"table":"px_kev_vulns"}`, Form: "mango", Engine: "couchdb",
				Columns: []string{"cve_id", "product"}, Rows: [][]string{{"CVE-2021-44228", "Log4j2"}},
			}}},
		},
		{
			ID: "f2", RuleID: "logic", Title: "Rounds down", Severity: SeverityWarning, ProofLevel: ProvedByTest,
			Evidence: []Evidence{{Kind: EvidenceTest, Test: &TestEvidence{Command: "go test ./...", FailsOnChange: true, PassesOnBase: true}}},
		},
	}
}

func testRecorder(c *recordedCouch, link string) *recorder {
	cfg := store.RecordsSettings{BaseURL: "http://couch.test", Database: "proven_records", User: "writer-user", Password: "writer-pass-1234", Timeout: time.Second}
	client := &http.Client{Transport: c}
	reader := store.NewCouchDB(store.CouchSettings{BaseURL: "http://couch.test", Prefix: "proven_", User: "reader-user", Password: "reader-pass-1234", RowLimit: 10, Timeout: time.Second}, client)
	return &recorder{
		writer: store.NewRecorder(cfg, client), reader: reader, linkBase: link, pr: "42",
		now: func() time.Time { return time.Date(2026, 10, 25, 7, 0, 0, 0, time.UTC) },
	}
}

// TestEvidenceRecordLinkExistsOnlyWhenTheBaseURLIsSet proves a stored finding
// carries the address of its record when the base URL is set, and carries none
// when it is not, while the records are written either way. The mutated line of
// the negative control is the `linkBase == ""` return in publish.
func TestEvidenceRecordLinkExistsOnlyWhenTheBaseURLIsSet(t *testing.T) {
	for _, tc := range []struct {
		name, link string
		wantLink   bool
	}{
		{"set", "http://couch.test:5984", true},
		{"unset", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &recordedCouch{page: store.ViewerPage()}
			res := Result{Findings: sampleFindings()}
			testRecorder(c, tc.link).publish(context.Background(), Request{HeadCommit: "0123456789abcdef"}, &res)
			if len(res.FailedServices) != 0 {
				t.Fatalf("failed services %v", res.FailedServices)
			}
			wrote := false
			for _, p := range c.paths {
				if p == "POST /proven_records/_bulk_docs" {
					wrote = true
				}
			}
			if !wrote {
				t.Fatalf("the records were not written: %v", c.paths)
			}
			for _, f := range res.Findings {
				has := f.EvidenceRecordURL != ""
				if has != tc.wantLink {
					t.Fatalf("finding %s link %q, want a link: %v", f.ID, f.EvidenceRecordURL, tc.wantLink)
				}
				if has && !strings.HasPrefix(f.EvidenceRecordURL, "http://couch.test:5984/proven_records/_design/evidence/viewer.html?doc=42-0123456789ab-") {
					t.Fatalf("link %q", f.EvidenceRecordURL)
				}
				if strings.Contains(f.EvidenceRecordURL, "_utils") {
					t.Fatalf("link goes through the admin console: %q", f.EvidenceRecordURL)
				}
			}
		})
	}
}

// TestNoRecorderWritesNothing proves a review with the records setting unset
// sends nothing to CouchDB and leaves every finding without a link.
func TestNoRecorderWritesNothing(t *testing.T) {
	res := Result{Findings: sampleFindings()}
	var rc *recorder
	rc.publish(context.Background(), Request{}, &res)
	for _, f := range res.Findings {
		if f.EvidenceRecordURL != "" {
			t.Fatalf("link without a recorder: %q", f.EvidenceRecordURL)
		}
	}
	if len(res.FailedServices) != 0 {
		t.Fatalf("failed services %v", res.FailedServices)
	}
}

// TestFailedRecordsAreNamedAndLeaveNoLink proves a CouchDB that fails is a
// failed service named couchdb, the review keeps its findings, and no finding
// gets a link to a record that was not stored.
func TestFailedRecordsAreNamedAndLeaveNoLink(t *testing.T) {
	c := &recordedCouch{page: store.ViewerPage(), bad: true}
	res := Result{Findings: sampleFindings()}
	testRecorder(c, "http://couch.test:5984").publish(context.Background(), Request{}, &res)
	if len(res.FailedServices) == 0 || res.FailedServices[0].Service != "couchdb" {
		t.Fatalf("failed services %v", res.FailedServices)
	}
	for _, f := range res.Findings {
		if f.EvidenceRecordURL != "" {
			t.Fatalf("link for a record that was not stored: %q", f.EvidenceRecordURL)
		}
	}
}

// TestUncheckedViewerIsNamed proves a review with records and a link base but
// no read-only account does not show its links without a word: a failed service
// named couchdb says the link was not checked, and the records and links are
// still made. The mutated line is the `rc.reader == nil` branch in publish:
// removed, the failed service is gone and the test fails.
func TestUncheckedViewerIsNamed(t *testing.T) {
	c := &recordedCouch{page: store.ViewerPage()}
	rc := testRecorder(c, "http://couch.test:5984")
	rc.reader = nil
	res := Result{Findings: sampleFindings()}
	rc.publish(context.Background(), Request{HeadCommit: "0123456789abcdef"}, &res)
	if len(res.FailedServices) != 1 || res.FailedServices[0].Service != "couchdb" || !strings.Contains(res.FailedServices[0].Reason, "not checked") {
		t.Fatalf("failed services %v", res.FailedServices)
	}
	for _, f := range res.Findings {
		if f.EvidenceRecordURL == "" {
			t.Fatalf("finding %s has no link", f.ID)
		}
	}
}

// TestRecordOfCopiesTheEvidenceAndNamesItsKeys proves a finding's record holds
// the query as it ran with its rows, the test command with its base and change
// result, and the CVE and package its rows name.
func TestRecordOfCopiesTheEvidenceAndNamesItsKeys(t *testing.T) {
	fs := sampleFindings()
	q := recordOf(fs[0], "42", "c0ffee", "run1")
	if len(q.CVEs) != 1 || q.CVEs[0] != "CVE-2021-44228" || len(q.Packages) != 1 || q.Packages[0] != "Log4j2" {
		t.Fatalf("keys %v %v", q.CVEs, q.Packages)
	}
	if e := q.Evidence[0]; e.Kind != EvidenceQuery || e.Form != "mango" || e.Query != `{"table":"px_kev_vulns"}` || len(e.Rows) != 1 {
		t.Fatalf("query evidence %+v", e)
	}
	tr := recordOf(fs[1], "42", "c0ffee", "run1")
	if e := tr.Evidence[0]; e.Command != "go test ./..." || !e.FailsOnChange || !e.PassesOnBase {
		t.Fatalf("test evidence %+v", e)
	}
}

// TestAMalformedReadAccountSettingIsNamed proves a read-only account that is
// set but malformed is not reported as an absent one: the viewer reason names
// the setting that kept it from loading. The mutated line is the readerField
// assignment in openRecorder: removed, the reason says only "not configured"
// and the test fails.
func TestAMalformedReadAccountSettingIsNamed(t *testing.T) {
	env := map[string]string{
		store.EnvCouchURL:        "http://couch.test:5984",
		store.EnvRecordsDB:       "proven_records",
		store.EnvRecordsUser:     "records-writer",
		store.EnvRecordsPassword: "writer-secret-1",
		store.EnvEvidenceLink:    "http://couch.test:5984",
		store.EnvCouchUser:       "records-reader",
		store.EnvCouchPassword:   "reader-secret-1",
		store.EnvCouchRowLimit:   "many",
	}
	rc, err := openRecorder(Options{Env: func(k string) string { return env[k] }})
	if err != nil || rc == nil {
		t.Fatalf("recorder %v, err %v", rc, err)
	}
	if rc.reader != nil {
		t.Fatal("a malformed row limit loaded the read-only account")
	}
	rc.writer = testRecorder(&recordedCouch{page: store.ViewerPage()}, "http://couch.test:5984").writer
	res := Result{Findings: sampleFindings()}
	rc.publish(context.Background(), Request{HeadCommit: "0123456789abcdef"}, &res)
	if len(res.FailedServices) != 1 || !strings.Contains(res.FailedServices[0].Reason, store.EnvCouchRowLimit) {
		t.Fatalf("failed services %v", res.FailedServices)
	}
}
