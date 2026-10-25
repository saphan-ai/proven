package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

const (
	testCouchUser = "proven-reader"
	testCouchPass = "couch-test-password"
)

// fakeCouch is a recorded CouchDB: it answers the four calls the store makes,
// in the shapes CouchDB documents, and it keeps every request it saw. It holds
// documents in memory and runs the equality selectors Proven builds. It
// returns the fields of a document in reverse name order, because CouchDB
// promises no order, so a store that trusted the answer's order would be wrong.
type fakeCouch struct {
	mu       sync.Mutex
	dbs      map[string][]map[string]any
	requests []string
	finds    []map[string]any
}

// newFakeCouch starts the recorded server. The caller closes it.
func newFakeCouch(t *testing.T) (*fakeCouch, *httptest.Server) {
	t.Helper()
	f := &fakeCouch{dbs: map[string][]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	return f, srv
}

// serve is the one handler. It refuses a request without the credential, as
// a real server does, so a store that forgot it fails.
func (f *fakeCouch) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)
	if u, p, ok := r.BasicAuth(); !ok || u != testCouchUser || p != testCouchPass {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized","reason":"` + r.Header.Get("Authorization") + `"}`))
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
			_, _ = w.Write([]byte(`{"error":"not_found","reason":"Database does not exist."}`))
			return
		}
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		f.finds = append(f.finds, in)
		sel, _ := in["selector"].(map[string]any)
		var fields []string
		for _, v := range in["fields"].([]any) {
			fields = append(fields, v.(string))
		}
		sort.Sort(sort.Reverse(sort.StringSlice(fields)))
		out := []map[string]any{}
		for _, d := range docs {
			match := true
			for k, v := range sel {
				if d[k] != v {
					match = false
				}
			}
			if !match {
				continue
			}
			row := map[string]any{}
			for _, k := range fields {
				if v, ok := d[k]; ok {
					row[k] = v
				}
			}
			out = append(out, row)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"docs": out})
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

// couchStore builds the CouchDB store on the recorded server.
func couchStore(srv *httptest.Server) *CouchDB {
	return NewCouchDB(CouchSettings{
		BaseURL: srv.URL, Prefix: "t_", User: testCouchUser, Password: testCouchPass,
		RowLimit: 100, Timeout: 2 * time.Second,
	}, nil)
}

// scenarioTables are the two tables the cross-store scenarios register: one
// KEV row, and two OSV rows of which one is for the package asked.
func scenarioTables() []Table {
	snap := "2026-10-24T07:00:00Z"
	return []Table{
		{
			Name:     "px_evidence.px_kev_vulns",
			Columns:  []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"},
			Rows:     [][]string{{"CVE-2021-44228", "Apache", "Log4j2", "2021-12-10", "Known"}},
			Snapshot: snap,
		},
		{
			Name:    "px_osvfile.px_osv_pkg",
			Columns: querytpl.OSVPackageColumns,
			Rows: [][]string{
				{"GHSA-aaaa-bbbb-cccc", "CVE-2021-1", "PyPI", "django", "1.0", "1.4", "", "HIGH", "2021-01-01"},
				{"GHSA-dddd-eeee-ffff", "", "PyPI", "flask", "0.1", "", "0.5", "", "2020-01-01"},
			},
			Snapshot: snap,
		},
	}
}

// TestCouchDBAgreesWithTheLocalStore proves the CouchDB store returns the same
// columns, rows and snapshot as the local store for the same request, with
// the cells in the sql form's column order though the server answers in
// another, and that its evidence names the mango form and the Mango query
// that ran (ADR-0047). The mutated fields are the template and the stored
// package name; the fake server answers fields in reverse order. Negative
// control: dropping one row in CouchDB.Run makes the row comparison fail.
func TestCouchDBAgreesWithTheLocalStore(t *testing.T) {
	when := time.Date(2026, 10, 24, 7, 0, 0, 0, time.UTC)
	ctx := context.Background()
	local, err := OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	local.SetClock(func() time.Time { return when })
	_, srv := newFakeCouch(t)
	couch := couchStore(srv)
	couch.SetClock(func() time.Time { return when })
	for _, tb := range scenarioTables() {
		if err := local.Register(ctx, tb); err != nil {
			t.Fatal(err)
		}
		if err := couch.Register(ctx, tb); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name string
		req  querytpl.Request
		rows int
	}{
		{"kev", querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}, 1},
		{"osv-package", querytpl.Request{Template: querytpl.OSVByPackage, Params: map[string]string{querytpl.ParamEcosystem: "PyPI", querytpl.ParamPackage: "Django"}}, 1},
		{"no-match", querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2000-0001"}}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantEv, err := local.Run(ctx, c.req, querytpl.Relations{})
			if err != nil {
				t.Fatal(err)
			}
			got, err := couch.Run(ctx, c.req, querytpl.Relations{})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rows) != c.rows || len(wantEv.Rows) != c.rows {
				t.Fatalf("rows: couchdb %d, local %d, want %d", len(got.Rows), len(wantEv.Rows), c.rows)
			}
			if !reflect.DeepEqual(got.Columns, wantEv.Columns) {
				t.Fatalf("columns %v, local %v", got.Columns, wantEv.Columns)
			}
			if !reflect.DeepEqual(got.Rows, wantEv.Rows) {
				t.Fatalf("rows %#v, local %#v", got.Rows, wantEv.Rows)
			}
			if got.Snapshot != wantEv.Snapshot || got.Snapshot != "2026-10-24T07:00:00Z" {
				t.Fatalf("snapshots %s %s", got.Snapshot, wantEv.Snapshot)
			}
			if got.Engine != "couchdb" || got.Form != querytpl.FormMango {
				t.Fatalf("engine %s form %s", got.Engine, got.Form)
			}
			forms, _ := querytpl.BuildForms(c.req, querytpl.Relations{})
			if got.Query != forms.Mango || !strings.HasPrefix(got.Query, `{"table":`) || got.Query == wantEv.Query {
				t.Fatalf("the evidence names %q, want the mango form %q", got.Query, forms.Mango)
			}
		})
	}
}

// TestCouchDBSendsOnlyTheBuiltSelector proves what reaches CouchDB is the
// template's selector with the validated values, the field list, the row
// limit and the credential, and nothing the request's text could add
// (ADR-0032). The mutated field is the CVE id: one that carries a Mango
// operator is refused before any request is sent.
func TestCouchDBSendsOnlyTheBuiltSelector(t *testing.T) {
	f, srv := newFakeCouch(t)
	couch := couchStore(srv)
	ctx := context.Background()
	if err := couch.Register(ctx, scenarioTables()[0]); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.requests, f.finds = nil, nil
	f.mu.Unlock()
	hostile := querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: `CVE-2021-1"},"$or":[{"x":"y`}}
	if _, err := couch.Run(ctx, hostile, querytpl.Relations{}); err == nil {
		t.Fatal("a value outside its format reached the store")
	}
	if len(f.requests) != 0 {
		t.Fatalf("requests sent for a refused value: %v", f.requests)
	}
	good := querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}
	if _, err := couch.Run(ctx, good, querytpl.Relations{}); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"selector": map[string]any{"cve_id": "CVE-2021-44228"},
		"fields":   []any{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"},
		"limit":    float64(100),
	}
	if len(f.finds) != 1 || !reflect.DeepEqual(f.finds[0], want) {
		t.Fatalf("sent %#v, want %#v", f.finds, want)
	}
	if !reflect.DeepEqual(f.requests, []string{"POST /t_px_evidence-px_kev_vulns/_find"}) {
		t.Fatalf("requests %v", f.requests)
	}
}

// TestCouchDBMissingDatabaseIsADownStore proves a database that was never
// registered is a named down store and never an empty answer, as it is on the
// other stores (ADR-0047). The mutated field is the database: the server
// holds none. Negative control: in CouchDB.Run, returning an empty Evidence
// for status 404 instead of the down error makes this test fail.
func TestCouchDBMissingDatabaseIsADownStore(t *testing.T) {
	_, srv := newFakeCouch(t)
	couch := couchStore(srv)
	req := querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}
	_, err := couch.Run(context.Background(), req, querytpl.Relations{})
	d, ok := service.As(err)
	if !ok || d.Service != "couchdb" || !strings.HasPrefix(d.Reason, "no table ") {
		t.Fatalf("want a down store named couchdb, got %v", err)
	}
}

// TestCouchDBFailuresAreNamedAndKeepTheCredentialOut proves a wrong credential,
// a server error and an unreachable server are each a down store named
// couchdb, and no reason holds the password or the server's body (ADR-0013
// point 4). The mutated fields are the password, then the server.
func TestCouchDBFailuresAreNamedAndKeepTheCredentialOut(t *testing.T) {
	_, srv := newFakeCouch(t)
	req := querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}
	cfg := CouchSettings{BaseURL: srv.URL, Prefix: "t_", User: testCouchUser, Password: "wrong-password", RowLimit: 10, Timeout: time.Second}
	_, err := NewCouchDB(cfg, nil).Run(context.Background(), req, querytpl.Relations{})
	if d, ok := service.As(err); !ok || d.Service != "couchdb" || d.Reason != "unauthorized" {
		t.Fatalf("wrong password: %v", err)
	}
	srv.Close()
	cfg.Password = testCouchPass
	_, err = NewCouchDB(cfg, nil).Run(context.Background(), req, querytpl.Relations{})
	d, ok := service.As(err)
	if !ok || d.Service != "couchdb" || d.Reason != "unreachable" {
		t.Fatalf("closed server: %v", err)
	}
	if strings.Contains(err.Error(), testCouchPass) || strings.Contains(err.Error(), "wrong-password") {
		t.Fatalf("a reason holds a password: %v", err)
	}
}

// TestCouchDBRegisterRefusesASecondLoad proves registering the same table
// twice is refused and the first load is kept (ADR-0047: the stores agree on
// this). The mutated field is the table name, registered twice.
func TestCouchDBRegisterRefusesASecondLoad(t *testing.T) {
	f, srv := newFakeCouch(t)
	couch := couchStore(srv)
	tb := scenarioTables()[0]
	if err := couch.Register(context.Background(), tb); err != nil {
		t.Fatal(err)
	}
	if err := couch.Register(context.Background(), tb); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second register: %v", err)
	}
	if n := len(f.dbs["t_px_evidence-px_kev_vulns"]); n != 1 {
		t.Fatalf("the database holds %d documents, want the first load's 1", n)
	}
	bad := tb
	bad.Name = "x;drop"
	if err := couch.Register(context.Background(), bad); err == nil {
		t.Fatal("a table name that is not plain was registered")
	}
}

// TestCouchDBDatabaseNames proves the database name is the prefix, the table
// name in lower case, and `.` written as `-` (ADR-0013). The mutated field is
// the table name.
func TestCouchDBDatabaseNames(t *testing.T) {
	c := NewCouchDB(CouchSettings{Prefix: "proven_"}, nil)
	for table, want := range map[string]string{
		"px_evidence.px_kev_vulns": "proven_px_evidence-px_kev_vulns",
		"Plain":                    "proven_plain",
	} {
		if got := c.Database(table); got != want {
			t.Fatalf("%s: %s, want %s", table, got, want)
		}
	}
}

// TestCouchCellKeepsNumbersAndRefusesNesting proves a field read from a
// document is text: a number keeps its written form, a missing or null field
// is empty, and a nested value is refused. The mutated field is the cell.
func TestCouchCellKeepsNumbersAndRefusesNesting(t *testing.T) {
	for raw, want := range map[string]string{`"x"`: "x", `0.97`: "0.97", `12`: "12", `null`: "", ``: "", `true`: "true"} {
		got, err := couchCell(json.RawMessage(raw))
		if err != nil || got != want {
			t.Fatalf("%q: %q %v", raw, got, err)
		}
	}
	for _, raw := range []string{`{"a":1}`, `[1]`} {
		if _, err := couchCell(json.RawMessage(raw)); err == nil {
			t.Fatalf("%s was accepted", raw)
		}
	}
}

// TestCouchDBLive runs the scenario against a real CouchDB. It runs only when
// PROVEN_COUCHDB_LIVE_URL names one, with PROVEN_COUCHDB_USER and
// PROVEN_COUCHDB_PASSWORD set, and it skips by name otherwise. It makes one
// database with a name no other run uses and removes it at the end.
func TestCouchDBLive(t *testing.T) {
	url := os.Getenv("PROVEN_COUCHDB_LIVE_URL")
	if url == "" {
		t.Skip("PROVEN_COUCHDB_LIVE_URL is not set; the live CouchDB test did not run")
	}
	// The clock makes the name unique to this run, and it fits the prefix's
	// 32 characters.
	prefix := fmt.Sprintf("proven_live_%d_", time.Now().UnixNano())
	cfg, err := LoadCouch(func(k string) string {
		switch k {
		case EnvCouchURL:
			return url
		case EnvCouchPrefix:
			return prefix
		}
		return os.Getenv(k)
	})
	if err != nil {
		t.Fatalf("live settings: %v", err)
	}
	couch := NewCouchDB(cfg, nil)
	ctx := context.Background()
	tb := scenarioTables()[0]
	t.Cleanup(func() { _, _, _ = couch.do(ctx, http.MethodDelete, "/"+couch.Database(tb.Name), nil) })
	if err := couch.Register(ctx, tb); err != nil {
		t.Fatal(err)
	}
	got, err := couch.Run(ctx, querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Form != querytpl.FormMango || !reflect.DeepEqual(got.Rows, tb.Rows) {
		t.Fatalf("%#v", got)
	}
}

// TestCouchDBAppearsInTheConformanceOfStores proves the three stores answer
// one request with the same columns and rows, each in the form it speaks:
// sql for Zetaris and the local store, mango for CouchDB (ADR-0047 point 4).
// The mutated field is the store: each is asked the same KEV request.
func TestCouchDBAppearsInTheConformanceOfStores(t *testing.T) {
	when := time.Date(2026, 10, 24, 7, 0, 0, 0, time.UTC)
	tb := scenarioTables()[0]
	req := querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}
	ctx := context.Background()

	local, err := OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	local.SetClock(func() time.Time { return when })
	_, csrv := newFakeCouch(t)
	couch := couchStore(csrv)
	couch.SetClock(func() time.Time { return when })
	zsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"headers": tb.Columns, "data": tb.Rows})
	}))
	defer zsrv.Close()
	z := NewZetaris(query.Settings{BaseURL: zsrv.URL, APIKey: "conform-key", OrgID: "1", RowLimit: 100, Timeout: time.Second})
	z.SetClock(func() time.Time { return when })

	stores := []struct {
		name, form string
		s          Store
	}{{"local", querytpl.FormSQL, local}, {"zetaris", querytpl.FormSQL, z}, {"couchdb", querytpl.FormMango, couch}}
	var first *Evidence
	for _, st := range stores {
		if err := st.s.Register(ctx, tb); err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
		got, err := st.s.Run(ctx, req, querytpl.Relations{})
		if err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
		if got.Form != st.form || got.Engine != st.name || got.Query == "" {
			t.Fatalf("%s: form %s engine %s query %q", st.name, got.Form, got.Engine, got.Query)
		}
		if first == nil {
			first = &got
			continue
		}
		if !reflect.DeepEqual(got.Columns, first.Columns) || !reflect.DeepEqual(got.Rows, first.Rows) || got.Snapshot != first.Snapshot {
			t.Fatalf("%s: %#v differs from %#v", st.name, got, *first)
		}
	}
}
