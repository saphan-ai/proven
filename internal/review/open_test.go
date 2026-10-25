package review_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/store"
)

// TestOpenSampleReviewKeepsF1AndF2 proves the sample review end to end through the seam: F1 is kept on a
// test that fails on the change and passes on the base, F2 and F4 on query rows with their EPSS rows,
// F3 is dropped for no evidence with its full trail (claim, evidence, refutation, verdict) in the trace,
// and no model is named as if from a receipt, because the sample has none.
func TestOpenSampleReviewKeepsF1AndF2(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("partial")
	}
	if len(res.FailedServices) != 0 {
		t.Fatalf("failed services %#v", res.FailedServices)
	}
	f1, ok := find(res, "F1")
	if !ok {
		t.Fatal("F1 missing")
	}
	if len(f1.Evidence) != 1 || f1.Evidence[0].Kind != review.EvidenceTest || f1.Evidence[0].Test == nil {
		t.Fatalf("F1 evidence %#v", f1.Evidence)
	}
	if !f1.Evidence[0].Test.FailsOnChange || !f1.Evidence[0].Test.PassesOnBase {
		t.Fatalf("F1 test %#v", f1.Evidence[0].Test)
	}
	if f1.Evidence[0].Test.Name != "TestDiscounted" {
		t.Fatalf("test %s", f1.Evidence[0].Test.Name)
	}
	f2, ok := find(res, "F2")
	if !ok {
		t.Fatal("F2 missing")
	}
	if len(f2.Evidence) != 2 || f2.Evidence[0].Kind != review.EvidenceQuery || f2.Evidence[0].Query == nil {
		t.Fatalf("F2 evidence %#v", f2.Evidence)
	}
	q := f2.Evidence[0].Query
	// ADR-0032: F2's query is the osv-by-package template, which returns the OSV rows of the package.
	if q.Engine != "local-sample" || len(q.Rows) != 2 || q.Rows[0][3] != "evil.example/leftpad" {
		t.Fatalf("F2 query %#v", q)
	}
	if f2.Evidence[1].Query == nil || len(f2.Evidence[1].Query.Rows) != 1 || f2.Evidence[1].Query.Rows[0][0] != "CVE-2026-0001" || f2.Evidence[1].Query.Rows[0][1] != "0.50000" {
		t.Fatalf("F2 EPSS %#v", f2.Evidence[1])
	}
	f4, ok := find(res, "F4")
	if !ok {
		t.Fatal("F4 missing")
	}
	if len(f4.Evidence) != 2 || f4.Evidence[0].Query == nil {
		t.Fatalf("F4 evidence %#v", f4.Evidence)
	}
	kq := f4.Evidence[0].Query
	if kq.Engine != "local-sample" || kq.SQL != "SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'" ||
		len(kq.Rows) != 1 || kq.Rows[0][0] != "CVE-2026-0001" {
		t.Fatalf("F4 query %#v", kq)
	}
	eq := f4.Evidence[1].Query
	if eq == nil || len(eq.Rows) != 1 || eq.Rows[0][0] != "CVE-2026-0001" || eq.Rows[0][1] != "0.50000" || eq.Rows[0][2] != "0.90000" || eq.Rows[0][3] != "2026-10-20" {
		t.Fatalf("F4 EPSS %#v", eq)
	}
	if _, ok := find(res, "F3"); ok {
		t.Fatal("F3 must be dropped")
	}
	for _, m := range res.Models {
		if m.Model != "unknown" || m.FromReceipt {
			t.Fatalf("model %#v", m)
		}
	}
	f3 := traceEvents(t, res.TracePath, "F3")
	kinds := make([]string, 0, len(f3))
	for _, e := range f3 {
		kinds = append(kinds, e.Kind)
	}
	if len(kinds) != 4 || kinds[0] != "claim" || kinds[1] != "evidence" || kinds[2] != "refutation" || kinds[3] != "verdict" {
		t.Fatalf("F3 events %#v", f3)
	}
	if f3[3].Text != "dropped: no evidence" {
		t.Fatalf("verdict %q", f3[3].Text)
	}
}

// TestSampleReviewDoesNotReadTheEnvironment proves that a sample review reads no environment variable, so
// it needs no key and reaches no service (ADR-0001).
func TestSampleReviewDoesNotReadTheEnvironment(t *testing.T) {
	rev, err := review.Open(review.Options{
		Sample: true,
		LogDir: t.TempDir(),
		Env: func(k string) string {
			// ADR-0045: the OTEL_* variables are the one family sample mode reads, because the lab's
			// pull-request review is a sample review and the dashboard shows it. They name no key of
			// Proven's, and a review with none of them set exports nothing.
			// ADR-0047 point 7: the evidence-record database and link base are read in sample mode too.
			if !strings.HasPrefix(k, "OTEL_") && !(k == store.EnvRecordsDB || k == store.EnvEvidenceLink) {
				t.Fatalf("sample mode read %s", k)
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)}); err != nil {
		t.Fatal(err)
	}
}

// TestCancelledContextIsPartial proves the seam's deadline contract: a context that has already ended
// gives a Partial result and a nil error, not a failure.
func TestCancelledContextIsPartial(t *testing.T) {
	rev := openSample(t, t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := rev.Review(ctx, review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Partial {
		t.Fatal("want partial")
	}
}

// TestLiveModeWithoutRuntimeNamesProverOff proves ADR-0025's fallback: with no container runtime the
// Prover is named off in the failed services, the review still returns, and no finding claims a test
// that ran.
func TestLiveModeWithoutRuntimeNamesProverOff(t *testing.T) {
	// An empty PATH finds no container runtime: the case this test is about.
	t.Setenv("PATH", t.TempDir())
	// The endpoint is closed so this test cannot reach a real Nightshift.
	// Mutated field: the listening socket. It is closed before Open returns.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed endpoint accepted a call")
	}))
	srv.Close()
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv(srv.URL, map[string]string{"PROVEN_GENERATOR_TIMEOUT": "1s"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("partial")
	}
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "prover" && strings.Contains(s.Reason, "Prover off") {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	for _, f := range res.Findings {
		for _, e := range f.Evidence {
			if e.Kind == review.EvidenceTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase {
				t.Fatalf("finding %s proved by a test %#v", f.ID, e.Test)
			}
		}
	}
}

// TestLiveModeRefusesABadContainerSetting proves that a bad container setting is a refusal, not a Prover
// that is quietly off: Open returns a *ConfigError that names the field and the allowed values.
// Mutated field: PROVEN_TEST_MEMORY. A value Validate refuses stops Open.
func TestLiveModeRefusesABadContainerSetting(t *testing.T) {
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv("http://127.0.0.1:9", map[string]string{"PROVEN_TEST_MEMORY": "lots"}),
	})
	var cfg *review.ConfigError
	if !errorAs(err, &cfg) || cfg.Field != "PROVEN_TEST_MEMORY" {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "allowed") || !strings.Contains(err.Error(), "512m") {
		t.Fatal(err)
	}
}

// TestEmptyLogDirIsRefused proves that Open refuses an empty LogDir, since the review must write its trace.
func TestEmptyLogDirIsRefused(t *testing.T) {
	_, err := review.Open(review.Options{Sample: true, Env: os.Getenv})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "LogDir" {
		t.Fatal(err)
	}
}

// TestCouchDBIsAccepted proves PROVEN_EVIDENCE_STORE=couchdb opens a review
// with the CouchDB settings given, builds no Zetaris client, and reads no
// Zetaris variable (ADR-0047). The mutated field is PROVEN_EVIDENCE_STORE;
// the Zetaris key is left empty, which would be a down Zetaris on that path.
func TestCouchDBIsAccepted(t *testing.T) {
	before := query.LiveBuilt()
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_EVIDENCE_STORE":   "couchdb",
			"PROVEN_COUCHDB_URL":      "http://127.0.0.1:9",
			"PROVEN_COUCHDB_USER":     "proven",
			"PROVEN_COUCHDB_PASSWORD": "couch-test-password",
			"ZETARIS_API_KEY":         "",
		}),
	})
	if err != nil || rev == nil {
		t.Fatalf("%v", err)
	}
	if query.LiveBuilt() != before {
		t.Fatalf("couchdb store built %d Zetaris clients", query.LiveBuilt()-before)
	}
}

// TestCouchDBWrongValueIsRefusedByName proves a CouchDB setting that is set
// wrongly stops Open before a review and the refusal names the variable and
// does not print the password (ADR-0013). The mutated field is
// PROVEN_COUCHDB_URL, then PROVEN_COUCHDB_PASSWORD (too short).
func TestCouchDBWrongValueIsRefusedByName(t *testing.T) {
	for field, extra := range map[string]map[string]string{
		store.EnvCouchURL:      {"PROVEN_COUCHDB_URL": "ftp://host"},
		store.EnvCouchPassword: {"PROVEN_COUCHDB_URL": "http://127.0.0.1:9", "PROVEN_COUCHDB_PASSWORD": "short"},
	} {
		extra["PROVEN_EVIDENCE_STORE"] = "couchdb"
		_, err := review.Open(review.Options{Sample: false, LogDir: t.TempDir(), Env: liveEnv("http://127.0.0.1:9", extra)})
		var cfg *review.ConfigError
		if err == nil || !errorAs(err, &cfg) || cfg.Field != field {
			t.Fatalf("%s: %v", field, err)
		}
		if strings.Contains(err.Error(), "short") {
			t.Fatalf("%s: the refusal printed the value: %v", field, err)
		}
	}
}

// TestCouchDBEmptyAddressIsADownStore proves an empty CouchDB address does
// not stop Open: the review names couchdb down (ADR-0013 point 4a). The
// mutated field is PROVEN_COUCHDB_URL, left empty.
func TestCouchDBEmptyAddressIsADownStore(t *testing.T) {
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv("http://127.0.0.1:9", map[string]string{"PROVEN_EVIDENCE_STORE": "couchdb"}),
	})
	if err != nil {
		t.Fatalf("an empty CouchDB address stopped Open: %v", err)
	}
}

// TestLocalStoreNeedsNoZetaris proves a local store does not build a Zetaris
// client and does not require the Zetaris settings. The mutated field is
// ZETARIS_API_KEY, left empty, with the store set to local.
func TestLocalStoreNeedsNoZetaris(t *testing.T) {
	dir := t.TempDir()
	before := query.LiveBuilt()
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: dir,
		Env: liveEnv("http://127.0.0.1:9", map[string]string{
			"PROVEN_EVIDENCE_STORE": "local",
			"PROVEN_EVIDENCE_FILE":  filepath.Join(dir, "evidence.sql"),
			"ZETARIS_REST_URL":      "",
			"ZETARIS_API_KEY":       "",
			"ZETARIS_ORG_ID":        "",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if rev == nil {
		t.Fatal("no reviewer")
	}
	if query.LiveBuilt() != before {
		t.Fatalf("local store built %d Zetaris clients", query.LiveBuilt()-before)
	}
}

// openSample opens a sample reviewer with the default settings and logs under logDir.
func openSample(t *testing.T, logDir string) review.Reviewer {
	t.Helper()
	rev, err := review.Open(review.Options{Sample: true, LogDir: logDir, Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

// find returns the finding with id.
func find(res review.Result, id string) (review.Finding, bool) {
	for _, f := range res.Findings {
		if f.ID == id {
			return f, true
		}
	}
	return review.Finding{}, false
}

// traceEvent is the part of one trace line these tests read.
type traceEvent struct {
	FindingID string `json:"finding_id"`
	Kind      string `json:"kind"`
	Text      string `json:"text"`
}

// traceEvents returns, in order, the trace lines about finding id.
func traceEvents(t *testing.T, path, id string) []traceEvent {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []traceEvent
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e traceEvent
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		if e.FindingID == id {
			out = append(out, e)
		}
	}
	return out
}

// fixturePath returns the discount fixture, the change every sample review reads.
func fixturePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "fixture", "discount")
}

// repoRoot walks up from the test's directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

// errorAs finds a *review.ConfigError in err's chain of single wraps.
func errorAs(err error, target **review.ConfigError) bool {
	if err == nil {
		return false
	}
	type unwrapper interface{ Unwrap() error }
	for err != nil {
		if cfg, ok := err.(*review.ConfigError); ok {
			*target = cfg
			return true
		}
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
