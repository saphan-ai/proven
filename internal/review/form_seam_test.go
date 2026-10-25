package review

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/store"
)

// The recorded CouchDB answers with this credential. It is a test value, not
// a secret of any deployment.
const (
	seamCouchUser = "proven-reader"
	seamCouchPass = "couch-test-password"
)

// TestReviewEvidenceNamesTheFormTheStoreRan proves a live review's query
// evidence names the form the configured store ran, with the rows of that
// query (ADR-0047, ADR-0032). CouchDB names mango. The same change on the
// local store names sql. The mutated field is the form the CouchDB store
// reports for the query it ran: reporting sql turns the couchdb case red.
func TestReviewEvidenceNamesTheFormTheStoreRan(t *testing.T) {
	repo, base, head := moduleCommits(t)
	t.Run("couchdb", func(t *testing.T) {
		srv := newSeamCouch(t)
		loadSeamTables(t, srv.URL)
		env := liveSecrets(map[string]string{
			"PROVEN_EVIDENCE_STORE":   "couchdb",
			"PROVEN_GUARDRAIL_MODE":   "off",
			"PROVEN_COUCHDB_URL":      srv.URL,
			"PROVEN_COUCHDB_USER":     seamCouchUser,
			"PROVEN_COUCHDB_PASSWORD": seamCouchPass,
		})
		res := reviewSeam(t, env, repo, base, head)
		assertForm(t, res, "mango", "couchdb")
	})
	t.Run("local", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "evidence.sql")
		if err := os.WriteFile(file, []byte(seamTablesSQL), 0o600); err != nil {
			t.Fatal(err)
		}
		env := liveSecrets(map[string]string{
			"PROVEN_EVIDENCE_STORE": "local",
			"PROVEN_EVIDENCE_FILE":  file,
			"PROVEN_GUARDRAIL_MODE": "off",
		})
		res := reviewSeam(t, env, repo, base, head)
		assertForm(t, res, "sql", "local")
	})
}

// reviewSeam opens a live review on the stub generator and runs it.
func reviewSeam(t *testing.T, env func(string) string, repo, base, head string) Result {
	t.Helper()
	rev, err := Open(Options{LogDir: t.TempDir(), Env: env, gen: &countGen{}, testRun: quietRun{}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatalf("the review ended: %v", err)
	}
	return res
}

// assertForm requires the dependency finding's query evidence to name form
// and engine, and to carry the advisory row the store was given.
func assertForm(t *testing.T, res Result, form, engine string) {
	t.Helper()
	var found *Finding
	for i := range res.Findings {
		if res.Findings[i].ID == "osv:example.com/mod" {
			found = &res.Findings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("no dependency finding; failed %#v findings %d", res.FailedServices, len(res.Findings))
	}
	var sawRow bool
	var queries int
	for _, e := range found.Evidence {
		if e.Query == nil {
			continue
		}
		queries++
		if e.Query.Form != form || e.Query.Engine != engine {
			t.Fatalf("evidence form %q engine %q, want %q %q; query %s", e.Query.Form, e.Query.Engine, form, engine, e.Query.SQL)
		}
		for _, row := range e.Query.Rows {
			if len(row) > 0 && row[0] == "GHSA-1" {
				sawRow = true
			}
		}
	}
	if queries == 0 || !sawRow {
		t.Fatalf("want the mango or sql row GHSA-1, queries %d evidence %#v", queries, found.Evidence)
	}
}

// moduleCommits is a git repository whose second commit adds one Go module.
func moduleCommits(t *testing.T) (dir, base, head string) {
	t.Helper()
	dir = t.TempDir()
	gitCmd(t, dir, "init")
	writeMod(t, dir, "module m\n\ngo 1.26\n")
	gitCmd(t, dir, "add", "go.mod")
	gitCmd(t, dir, "commit", "-m", "[fixture] gomod: base")
	base = strings.TrimSpace(string(gitCmd(t, dir, "rev-parse", "HEAD")))
	writeMod(t, dir, "module m\n\ngo 1.26\n\nrequire example.com/mod v1.2.3\n")
	gitCmd(t, dir, "add", "go.mod")
	gitCmd(t, dir, "commit", "-m", "[fixture] gomod: add module")
	head = strings.TrimSpace(string(gitCmd(t, dir, "rev-parse", "HEAD")))
	return dir, base, head
}

func writeMod(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// seamTablesSQL is the same rows the recorded CouchDB is given, in the
// file the local store loads.
const seamTablesSQL = `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', 'CVE-2026-0001', 'Go', 'example.com/mod', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001', 'example', 'mod', '2026-01-15', 'Known');
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2021-44228', 'example', 'log4j', '2021-12-10', 'Known');
CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
`

// loadSeamTables registers the three tables on the recorded server, the way
// an administrator would, so the review only reads.
func loadSeamTables(t *testing.T, rawURL string) {
	t.Helper()
	c := store.NewCouchDB(store.CouchSettings{
		BaseURL: rawURL, Prefix: "proven_", User: seamCouchUser, Password: seamCouchPass,
		RowLimit: 100, Timeout: 2 * time.Second,
	}, nil)
	ctx := context.Background()
	tables := []store.Table{
		{
			Name:    "px_osvfile.px_osv_pkg",
			Columns: []string{"id", "aliases", "ecosystem", "package", "introduced", "fixed", "last_affected", "severity", "published"},
			Rows:    [][]string{{"GHSA-1", "CVE-2026-0001", "Go", "example.com/mod", "1.0.0", "1.2.4", "", "HIGH", "2026-01-15"}},
		},
		{
			Name:    "px_evidence.px_kev_vulns",
			Columns: []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"},
			Rows: [][]string{
				{"CVE-2026-0001", "example", "mod", "2026-01-15", "Known"},
				{"CVE-2021-44228", "example", "log4j", "2021-12-10", "Known"},
			},
		},
		{
			Name:    "px_evidence.px_epss_cve_rows",
			Columns: []string{"cve_id", "epss", "percentile", "score_date"},
		},
	}
	for _, table := range tables {
		if err := c.Register(ctx, table); err != nil {
			t.Fatal(err)
		}
	}
}

// seamCouch is a recorded CouchDB: it keeps documents in memory and answers
// the calls the store makes. It is not a network service.
type seamCouch struct {
	mu  sync.Mutex
	dbs map[string][]map[string]any
}

func newSeamCouch(t *testing.T) *httptest.Server {
	t.Helper()
	f := &seamCouch{dbs: map[string][]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	return srv
}

func (f *seamCouch) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, p, ok := r.BasicAuth(); !ok || u != seamCouchUser || p != seamCouchPass {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	db := parts[0]
	w.Header().Set("Content-Type", "application/json")
	switch {
	case len(parts) == 1 && r.Method == http.MethodPut:
		if _, ok := f.dbs[db]; ok {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		f.dbs[db] = nil
		w.WriteHeader(http.StatusCreated)
	case len(parts) == 2 && parts[1] == "_bulk_docs":
		var in struct {
			Docs []map[string]any `json:"docs"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if _, ok := f.dbs[db]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.dbs[db] = append(f.dbs[db], in.Docs...)
		w.WriteHeader(http.StatusCreated)
	case len(parts) == 2 && parts[1] == "_find":
		docs, ok := f.dbs[db]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not_found"}`))
			return
		}
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		sel, _ := in["selector"].(map[string]any)
		out := []map[string]any{}
		for _, d := range docs {
			match := true
			for k, v := range sel {
				if d[k] != v {
					match = false
				}
			}
			if match {
				out = append(out, d)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"docs": out})
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
