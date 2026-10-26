package store

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// The two accounts of the recorded CouchDB. The writing account is a database
// admin; the read-only account is a member of the database's security
// document. The fake follows CouchDB's own rules: a member reads everything and
// may write any document that is not a design document, and only an admin may
// create the database or write a design document. What stops a member writing
// is therefore the design document's validate_doc_update, which the fake runs
// (in Node) on every write when the stored design document carries one. A test
// that sends a write with the read-only account passes only if that function is
// there and refuses it.
const (
	fakeWriter     = "records-writer"
	fakeWriterPass = "writer-pass-1234"
	fakeReader     = "records-reader"
	fakeReaderPass = "reader-pass-1234"
	fakeDB         = "proven_records"
)

// recordsCouch is a recorded CouchDB: it holds one database in memory, checks the
// account of every request, and logs who asked for what. It is driven through
// a RoundTripper, so no listener is opened and the tests run where a socket
// cannot be bound.
type recordsCouch struct {
	mu         sync.Mutex
	exists     bool
	docs       map[string]map[string]any
	attach     map[string][]byte
	designRev  int
	designPuts int
	log        []string
	down       bool
}

func newRecordsCouch() *recordsCouch {
	return &recordsCouch{docs: map[string]map[string]any{}, attach: map[string][]byte{}}
}

// RoundTrip serves the request in process.
func (f *recordsCouch) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	f.serve(rec, req)
	return rec.Result(), nil
}

func (f *recordsCouch) client() *http.Client { return &http.Client{Transport: f} }

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// serve answers the routes the records need.
func (f *recordsCouch) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	user, pass, _ := r.BasicAuth()
	role := ""
	switch {
	case user == fakeWriter && pass == fakeWriterPass:
		role = "writer"
	case user == fakeReader && pass == fakeReaderPass:
		role = "reader"
	}
	f.log = append(f.log, user+" "+r.Method+" "+r.URL.Path)
	if f.down {
		reply(w, http.StatusInternalServerError, map[string]string{"error": "down"})
		return
	}
	if role == "" {
		reply(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	db, rest, _ := strings.Cut(path, "/")
	if db != fakeDB {
		reply(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	// Only a database admin makes the database or writes a design document.
	if role != "writer" && r.Method != http.MethodGet && r.Method != http.MethodPost {
		reply(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	switch {
	case rest == "" && r.Method == http.MethodPut:
		if f.exists {
			reply(w, http.StatusPreconditionFailed, map[string]string{"error": "file_exists"})
			return
		}
		f.exists = true
		reply(w, http.StatusCreated, map[string]bool{"ok": true})
	case !f.exists:
		reply(w, http.StatusNotFound, map[string]string{"error": "not_found"})
	case rest == "_design/"+designName && r.Method == http.MethodGet:
		d, ok := f.docs["_design/"+designName]
		if !ok {
			reply(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		reply(w, http.StatusOK, d)
	case rest == "_design/"+designName && r.Method == http.MethodPut:
		var d map[string]any
		_ = json.NewDecoder(r.Body).Decode(&d)
		if old, ok := f.docs["_design/"+designName]; ok && d["_rev"] != old["_rev"] {
			reply(w, http.StatusConflict, map[string]string{"error": "conflict"})
			return
		}
		f.designRev++
		f.designPuts++
		d["_rev"] = string(rune('0'+f.designRev)) + "-rev"
		if atts, ok := d["_attachments"].(map[string]any); ok {
			for name, a := range atts {
				raw, _ := base64.StdEncoding.DecodeString(a.(map[string]any)["data"].(string))
				f.attach[name] = raw
			}
		}
		f.docs["_design/"+designName] = d
		reply(w, http.StatusCreated, map[string]bool{"ok": true})
	case rest == "_design/"+designName+"/"+viewerName && r.Method == http.MethodGet:
		page, ok := f.attach[viewerName]
		if !ok {
			reply(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	case rest == "_bulk_docs" && r.Method == http.MethodPost:
		var body struct {
			Docs []map[string]any `json:"docs"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		var out []map[string]any
		for _, d := range body.Docs {
			id := d["_id"].(string)
			if strings.HasPrefix(id, "_design/") && role != "writer" {
				out = append(out, map[string]any{"id": id, "error": "forbidden"})
				continue
			}
			if msg := f.validate(d, user); msg != "" {
				out = append(out, map[string]any{"id": id, "error": "forbidden", "reason": msg})
				continue
			}
			if strings.Contains(id, "refuseme") {
				out = append(out, map[string]any{"id": id, "error": "forbidden"})
				continue
			}
			f.docs[id] = d
			out = append(out, map[string]any{"id": id, "ok": true})
		}
		reply(w, http.StatusCreated, out)
	case r.Method == http.MethodGet:
		d, ok := f.docs[rest]
		if !ok {
			reply(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		reply(w, http.StatusOK, d)
	default:
		reply(w, http.StatusBadRequest, map[string]string{"error": "bad_request"})
	}
}

// securityObject is the database's security document: the writing account is
// a database admin and the read-only account a member.
func securityObject() map[string]any {
	return map[string]any{
		"admins":  map[string]any{"names": []string{fakeWriter}, "roles": []string{}},
		"members": map[string]any{"names": []string{fakeReader}, "roles": []string{}},
	}
}

// validate runs the stored design document's validate_doc_update on a write by
// user, the way CouchDB does, and returns the refusal, or "" when it admits the
// write or the design document has none. Node runs it; without Node the write is
// admitted and the tests that depend on the function skip by name.
func (f *recordsCouch) validate(newDoc map[string]any, user string) string {
	src, _ := f.docs["_design/"+designName]["validate_doc_update"].(string)
	if src == "" {
		return ""
	}
	return runVDU(src, newDoc, map[string]any{"name": user, "roles": []string{}}, securityObject())
}

// runVDU runs one validate_doc_update function in Node over a document, the
// user's context and the security document. It returns "" when the function
// returns and the refusal text when it throws.
func runVDU(src string, newDoc, userCtx, secObj map[string]any) string {
	node, err := exec.LookPath("node")
	if err != nil {
		return ""
	}
	script := `const a = JSON.parse(process.env.ARGS);
const f = eval('(' + process.env.VDU_SRC + ')');
try { f(a.newDoc, a.oldDoc, a.userCtx, a.secObj); console.log(""); }
catch (e) { console.log(JSON.stringify(e)); }`
	args, _ := json.Marshal(map[string]any{"newDoc": newDoc, "oldDoc": nil, "userCtx": userCtx, "secObj": secObj})
	cmd := exec.Command(node, "-e", script)
	cmd.Env = append(os.Environ(), "VDU_SRC="+src, "ARGS="+string(args))
	raw, err := cmd.Output()
	if err != nil {
		return "the function did not run: " + err.Error()
	}
	return strings.TrimSpace(string(raw))
}

// records returns the stored evidence records, sorted by id.
func (f *recordsCouch) records() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []map[string]any
	for id, d := range f.docs {
		if !strings.HasPrefix(id, "_design/") {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["_id"].(string) < out[j]["_id"].(string) })
	return out
}

func writerSettings() RecordsSettings {
	return RecordsSettings{BaseURL: "http://couch.test", Database: fakeDB, User: fakeWriter, Password: fakeWriterPass, Timeout: time.Second}
}

func readerStore(f *recordsCouch) *CouchDB {
	return NewCouchDB(CouchSettings{BaseURL: "http://couch.test", Prefix: "proven_", User: fakeReader, Password: fakeReaderPass, RowLimit: 10, Timeout: time.Second}, f.client())
}

const sampleCommit = "0123456789abcdef0123456789abcdef01234567"

// sampleRecords is the sample scenario's records: a finding proved by the
// known-exploited table's row for Log4Shell, one proved by a package
// advisory row that carries the same CVE as an alias, and one proved by a test.
// The test finding's title carries markup on purpose.
func sampleRecords() []EvidenceRecord {
	kevCols := []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}
	kevRows := [][]string{{"CVE-2021-44228", "Apache", "Log4j2", "2021-12-10", "Known"}}
	osvCols := []string{"id", "aliases", "ecosystem", "package", "introduced", "fixed"}
	osvRows := [][]string{{"GHSA-jfh8-c2jp-5v3q", "CVE-2021-44228", "Maven", "log4j-core", "2.0", "2.15.0"}}
	run := RunStamp(time.Date(2026, 10, 25, 7, 0, 0, 0, time.UTC))
	var out []EvidenceRecord
	for _, q := range []struct {
		id, title string
		cols      []string
		rows      [][]string
		query     string
	}{
		{"f-kev", "Dependency is known exploited", kevCols, kevRows, `{"table":"px_kev_vulns","selector":{"cve_id":"CVE-2021-44228"}}`},
		{"f-osv", "Dependency has an advisory", osvCols, osvRows, `{"table":"px_osv","selector":{"package":"log4j-core"}}`},
	} {
		cves, pkgs := QueryKeys(q.cols, q.rows)
		out = append(out, EvidenceRecord{
			PR: "42", Commit: sampleCommit, Run: run, FindingID: q.id, RuleID: "dependency", Title: q.title,
			Severity: "ERROR", Path: "go.mod", StartLine: 7, ProofLevel: "proved-by-data", CVEs: cves, Packages: pkgs,
			Evidence: []RecordProof{{Kind: "query", Form: "mango", Query: q.query, Engine: "couchdb", Snapshot: "2026-10-24T07:00:00Z", Columns: q.cols, Rows: q.rows}},
		})
	}
	out = append(out, EvidenceRecord{
		PR: "42", Commit: sampleCommit, Run: run, FindingID: "f-test", RuleID: "logic", Title: `Rounds <b onmouseover="x">down</b>`,
		Severity: "WARNING", Path: "price/discount.go", StartLine: 31, ProofLevel: "proved-by-test",
		Evidence: []RecordProof{{Kind: "test", Path: "price/discount_test.go", Name: "TestDiscount", Command: "go test ./price -run TestDiscount", FailsOnChange: true, PassesOnBase: true, Output: "FAIL"}},
	})
	return out
}

type viewRow struct {
	ID    string `json:"id"`
	Key   []any  `json:"key"`
	Value any    `json:"value"`
}

// runMap runs one view's map function in Node over docs, as CouchDB's JavaScript
// query server would, and returns what it emitted. The map functions are
// JavaScript and Go cannot run them, so the test uses the same language.
func runMap(t *testing.T, mapSrc string, docs []map[string]any) []viewRow {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed: the view map functions are UNRUN")
	}
	script := `const docs = JSON.parse(require('fs').readFileSync(0, 'utf8'));
const rows = []; let cur;
function emit(k, v) { rows.push({id: cur._id, key: k, value: v}); }
const f = eval('(' + process.env.MAP_SRC + ')');
for (const d of docs) { cur = d; f(d); }
console.log(JSON.stringify(rows));`
	in, _ := json.Marshal(docs)
	cmd := exec.Command(node, "-e", script)
	cmd.Env = append(os.Environ(), "MAP_SRC="+mapSrc)
	cmd.Stdin = bytes.NewReader(in)
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var rows []viewRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("node output %q: %v", raw, err)
	}
	return rows
}

func publishSample(t *testing.T, f *recordsCouch) map[string]string {
	t.Helper()
	ids, err := NewRecorder(writerSettings(), f.client()).Publish(context.Background(), sampleRecords())
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

// TestViewsReturnTheSampleScenarioRecords proves each of the four views emits
// the sample scenario's records under the keys a reader narrows by: review
// (pull request and commit), finding, CVE with package, and proof level.
// The records are the ones Publish stored, so the whole path from a review's
// record to a view row is run. The mutated field of the negative control is the
// view's map function: a view whose function emits nothing makes it fail.
func TestViewsReturnTheSampleScenarioRecords(t *testing.T) {
	f := newRecordsCouch()
	publishSample(t, f)
	docs := f.records()
	if len(docs) != 3 {
		t.Fatalf("stored %d records, want 3", len(docs))
	}
	keysOf := func(view string) []string {
		var out []string
		for _, r := range runMap(t, ViewMap(view), docs) {
			b, _ := json.Marshal(r.Key)
			out = append(out, string(b))
		}
		sort.Strings(out)
		return out
	}
	run := RunStamp(time.Date(2026, 10, 25, 7, 0, 0, 0, time.UTC))
	cases := map[string][]string{
		"by_review": {
			`["42","` + sampleCommit + `","f-kev","` + run + `"]`,
			`["42","` + sampleCommit + `","f-osv","` + run + `"]`,
			`["42","` + sampleCommit + `","f-test","` + run + `"]`,
		},
		"by_finding": {
			`["f-kev","42","` + sampleCommit + `","` + run + `"]`,
			`["f-osv","42","` + sampleCommit + `","` + run + `"]`,
			`["f-test","42","` + sampleCommit + `","` + run + `"]`,
		},
		// The test finding names no CVE and no package, so it is not in this view.
		"by_cve_package": {
			`["CVE-2021-44228","Log4j2","42","` + sampleCommit + `","f-kev"]`,
			`["CVE-2021-44228","log4j-core","42","` + sampleCommit + `","f-osv"]`,
		},
		"by_proof_level": {
			`["proved-by-data","42","` + sampleCommit + `","f-kev","` + run + `"]`,
			`["proved-by-data","42","` + sampleCommit + `","f-osv","` + run + `"]`,
			`["proved-by-test","42","` + sampleCommit + `","f-test","` + run + `"]`,
		},
	}
	if got := ViewNames(); !reflect.DeepEqual(got, []string{"by_cve_package", "by_finding", "by_proof_level", "by_review"}) {
		t.Fatalf("views %v", got)
	}
	for view, want := range cases {
		t.Run(view, func(t *testing.T) {
			if got := keysOf(view); !reflect.DeepEqual(got, want) {
				t.Fatalf("keys\n got %v\nwant %v", got, want)
			}
		})
	}
	// The design document row is not a record and no view emits it.
	design := f.docs["_design/"+designName]
	if rows := runMap(t, ViewMap("by_review"), []map[string]any{design}); len(rows) != 0 {
		t.Fatalf("the design document was emitted: %v", rows)
	}
}

// TestPublishWritesAsTheWritingAccountAndOnce proves Publish makes the database
// and the design document with the four views and the viewer page as an
// attachment, writes one document per finding, only with the writing account,
// and does not write the design document again when it has not changed.
func TestPublishWritesAsTheWritingAccountAndOnce(t *testing.T) {
	f := newRecordsCouch()
	ids := publishSample(t, f)
	publishSample(t, f)
	if f.designPuts != 1 {
		t.Fatalf("design document written %d times, want 1", f.designPuts)
	}
	for _, line := range f.log {
		if !strings.HasPrefix(line, fakeWriter+" ") {
			t.Fatalf("a request was not made by the writing account: %s", line)
		}
	}
	if !bytes.Equal(f.attach[viewerName], ViewerPage()) {
		t.Fatal("the viewer page attachment is not the page")
	}
	views, _ := f.docs["_design/"+designName]["views"].(map[string]any)
	if len(views) != 4 {
		t.Fatalf("views %v", views)
	}
	if len(ids) != 3 || ids["f-kev"] == "" || !strings.HasPrefix(ids["f-kev"], "42-0123456789ab-") {
		t.Fatalf("ids %v", ids)
	}
	if n := len(f.records()); n != 3 {
		t.Fatalf("stored %d records, want one per finding", n)
	}
}

// TestViewerIsServedToTheReadOnlyAccountAndNeedsNoWrite proves the viewer page
// is served by CouchDB to the read-only account, that the check makes only
// GET requests, and that the read-only account is refused a write. The mutated
// field of the negative control is the same-user refusal in LoadRecords: with
// it removed, TestLoadRecordsKeepsTheWriterApart fails.
func TestViewerIsServedToTheReadOnlyAccountAndNeedsNoWrite(t *testing.T) {
	f := newRecordsCouch()
	publishSample(t, f)
	f.log = nil
	if err := readerStore(f).CheckViewer(context.Background(), fakeDB); err != nil {
		t.Fatal(err)
	}
	if len(f.log) != 1 || f.log[0] != fakeReader+" GET /"+fakeDB+"/_design/evidence/viewer.html" {
		t.Fatalf("the viewer check made: %v", f.log)
	}
}

// bulkAs posts one record to the records database as the given account and
// returns the answer's row.
func bulkAs(t *testing.T, f *recordsCouch, user, pass, id string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"docs": []map[string]any{{"_id": id, "type": "evidence-record", "finding_id": id}}})
	req, _ := http.NewRequest(http.MethodPost, "http://couch.test/"+fakeDB+"/_bulk_docs", bytes.NewReader(body))
	req.SetBasicAuth(user, pass)
	resp, err := f.client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var rows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil || len(rows) != 1 {
		t.Fatalf("answer %v %v", rows, err)
	}
	return rows[0]
}

// TestReadOnlyAccountCannotWriteARecord proves a member of the records database,
// the read-only account, is refused a write by the design document itself, while
// the database admin is admitted. The fake lets a member write any document that
// is not a design document, as CouchDB does, so only the design document's
// validate_doc_update can refuse it. The mutated line is the validate_doc_update
// entry of the design document: with it removed the member's write is admitted
// and the test fails.
func TestReadOnlyAccountCannotWriteARecord(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed: the validate_doc_update function is UNRUN")
	}
	f := newRecordsCouch()
	publishSample(t, f)
	before := len(f.records())
	if row := bulkAs(t, f, fakeReader, fakeReaderPass, "overwrite-1"); row["error"] != "forbidden" {
		t.Fatalf("the read-only account's write was admitted: %v", row)
	}
	if len(f.records()) != before {
		t.Fatal("the read-only account wrote a record")
	}
	if row := bulkAs(t, f, fakeWriter, fakeWriterPass, "admitted-1"); row["ok"] != true {
		t.Fatalf("the writing account was refused: %v", row)
	}
	if len(f.records()) != before+1 {
		t.Fatal("the writing account's record is not stored")
	}
}

// TestValidateDocUpdateAdmitsOnlyAdmins proves the function on its own, over the
// shapes of a security document: a server admin and a database admin, by name
// and by role, are admitted; a member, a member by role, a user in no list, an
// anonymous user and a database with no security document are refused. The
// mutated line is the admins.names check: with it removed the admin by name is
// refused and the test fails.
func TestValidateDocUpdateAdmitsOnlyAdmins(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed: the validate_doc_update function is UNRUN")
	}
	sec := map[string]any{
		"admins":  map[string]any{"names": []string{"dba"}, "roles": []string{"dba-role"}},
		"members": map[string]any{"names": []string{"reader"}, "roles": []string{"reader-role"}},
	}
	doc := map[string]any{"_id": "x", "type": "evidence-record"}
	for _, c := range []struct {
		name  string
		user  any
		roles []string
		sec   map[string]any
		admit bool
	}{
		{"a server admin", "root", []string{"_admin"}, sec, true},
		{"a database admin by name", "dba", []string{}, sec, true},
		{"a database admin by role", "someone", []string{"dba-role"}, sec, true},
		{"a member", "reader", []string{}, sec, false},
		{"a member by role", "other", []string{"reader-role"}, sec, false},
		{"a user in no list", "stranger", []string{}, sec, false},
		{"an anonymous user", nil, []string{}, sec, false},
		{"no security document", "reader", []string{}, map[string]any{}, false},
	} {
		got := runVDU(ViewVDU(), doc, map[string]any{"name": c.user, "roles": c.roles}, c.sec)
		if c.admit && got != "" || !c.admit && !strings.Contains(got, "forbidden") {
			t.Errorf("%s: answer %q, admit want %v", c.name, got, c.admit)
		}
	}
}

// TestCheckViewerNamesAMissingPage proves a records database with no viewer
// page is a named down store, not a silent pass, so no link is shown for it.
func TestCheckViewerNamesAMissingPage(t *testing.T) {
	f := newRecordsCouch()
	f.exists = true
	err := readerStore(f).CheckViewer(context.Background(), fakeDB)
	if d, ok := service.As(err); !ok || d.Service != "couchdb" || !strings.Contains(d.Reason, "404") {
		t.Fatalf("missing page: %v", err)
	}
}

// TestPublishLeavesOutARefusedRecordAndNamesIt proves a record CouchDB refuses
// is not in the returned ids, so no link is made for it, and the error says how
// many were refused.
func TestPublishLeavesOutARefusedRecordAndNamesIt(t *testing.T) {
	f := newRecordsCouch()
	recs := sampleRecords()
	recs[2].FindingID = "refuseme"
	ids, err := NewRecorder(writerSettings(), f.client()).Publish(context.Background(), recs)
	if d, ok := service.As(err); !ok || !strings.Contains(d.Reason, "1 of 3") {
		t.Fatalf("error %v", err)
	}
	if _, has := ids["refuseme"]; has || len(ids) != 2 {
		t.Fatalf("ids %v", ids)
	}
}

// TestPublishOnADownServerIsANamedFailure proves a failing CouchDB is a down
// store named couchdb and that the writing credential is not in the text.
func TestPublishOnADownServerIsANamedFailure(t *testing.T) {
	f := newRecordsCouch()
	f.down = true
	_, err := NewRecorder(writerSettings(), f.client()).Publish(context.Background(), sampleRecords())
	d, ok := service.As(err)
	if !ok || d.Service != "couchdb" || strings.Contains(err.Error(), fakeWriterPass) {
		t.Fatalf("error %v", err)
	}
	if _, err := NewRecorder(writerSettings(), f.client()).Publish(context.Background(), nil); err != nil {
		t.Fatalf("no records is not a failure: %v", err)
	}
}

func recordsEnv(extra map[string]string) func(string) string {
	m := map[string]string{
		EnvCouchURL: "http://couch.test:5984/", EnvCouchUser: fakeReader, EnvCouchPassword: fakeReaderPass,
		EnvRecordsDB: fakeDB, EnvRecordsUser: fakeWriter, EnvRecordsPassword: fakeWriterPass,
	}
	for k, v := range extra {
		m[k] = v
	}
	return func(k string) string { return m[k] }
}

// TestLoadRecordsKeepsTheWriterApart proves the records settings: unset is off,
// a set one needs its address and writing account, and the writing user may not
// be the read-only user. The mutated field is EnvRecordsUser, set equal to the
// read-only user.
func TestLoadRecordsKeepsTheWriterApart(t *testing.T) {
	off, err := LoadRecords(func(string) string { return "" })
	if err != nil || off.Enabled() {
		t.Fatalf("unset: %+v %v", off, err)
	}
	on, err := LoadRecords(recordsEnv(nil))
	if err != nil || !on.Enabled() || on.BaseURL != "http://couch.test:5984" || on.User != fakeWriter {
		t.Fatalf("set: %+v %v", on, err)
	}
	for name, env := range map[string]map[string]string{
		"the writing user is the read-only user": {EnvRecordsUser: fakeReader},
		"no writing user":                        {EnvRecordsUser: ""},
		"no writing password":                    {EnvRecordsPassword: ""},
		"a short writing password":               {EnvRecordsPassword: "short"},
		"no address":                             {EnvCouchURL: ""},
		"a bad database name":                    {EnvRecordsDB: "Not_Valid"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := LoadRecords(recordsEnv(env))
			var ce *ConfigError
			if err == nil || !asConfig(err, &ce) {
				t.Fatalf("accepted: %v", err)
			}
			for _, secret := range []string{fakeWriterPass, fakeReaderPass, fakeWriter} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("the refusal echoes a credential: %v", err)
				}
			}
		})
	}
}

// TestEvidenceURLExistsOnlyWhenTheBaseIsSet proves the link is built from the
// configured base, the records database and the document id, is empty when the
// base is unset, and never goes through the admin console. The mutated line of
// the negative control is the empty-base check in EvidenceURL.
func TestEvidenceURLExistsOnlyWhenTheBaseIsSet(t *testing.T) {
	if got := EvidenceURL("", fakeDB, "42-abc-run-f1"); got != "" {
		t.Fatalf("unset base gave a link: %q", got)
	}
	if got := EvidenceURL("http://couch.test", "", "x"); got != "" {
		t.Fatalf("no database gave a link: %q", got)
	}
	if got := EvidenceURL("http://couch.test", fakeDB, ""); got != "" {
		t.Fatalf("no id gave a link: %q", got)
	}
	got := EvidenceURL("http://couch.test:5984/", fakeDB, "42-abc-run-f1")
	want := "http://couch.test:5984/proven_records/_design/evidence/viewer.html?doc=42-abc-run-f1"
	if got != want {
		t.Fatalf("link %q, want %q", got, want)
	}
	if strings.Contains(got, "_utils") {
		t.Fatal("the link goes through the admin console")
	}
	base, err := LoadLinkBase(func(string) string { return "" })
	if err != nil || base != "" {
		t.Fatalf("unset link base: %q %v", base, err)
	}
	for _, bad := range []string{"couch.test", "ftp://couch.test", "http://u:p@couch.test", "http://couch.test/?a=b", "http://couch.test/#x"} {
		if _, err := LoadLinkBase(func(string) string { return bad }); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

// TestViewerPageIsSelfContainedAndReadOnly proves the page loads nothing from
// outside, never names the admin console, and makes only GET requests.
func TestViewerPageIsSelfContainedAndReadOnly(t *testing.T) {
	page := string(ViewerPage())
	for _, banned := range []string{"http://", "https://", "//cdn", "_utils", "<link", "src=", "@import", `method: "PUT"`, `method: "POST"`, `method: "DELETE"`, "XMLHttpRequest"} {
		if strings.Contains(page, banned) {
			t.Errorf("the page contains %q", banned)
		}
	}
	if n := strings.Count(page, "method:"); n != 1 || !strings.Contains(page, `method: "GET"`) {
		t.Errorf("the page names %d request methods, want one GET", n)
	}
}

// TestViewerEscapesRecordValuesAsText proves the page's escaping function turns
// markup into text: no character that opens a tag or an attribute survives.
// The function is cut out of the page between its markers and run on its own.
func TestViewerEscapesRecordValuesAsText(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed: the escaping function is UNRUN")
	}
	page := string(ViewerPage())
	begin := strings.Index(page, "// esc:begin")
	end := strings.Index(page, "// esc:end")
	if begin < 0 || end < begin {
		t.Fatal("the escaping function's markers are missing")
	}
	script := `const esc = new Function(process.env.CODE + '; return esc;')();
console.log(JSON.stringify([process.env.IN, null, undefined, 42].map(esc)));`
	for _, in := range []string{
		`<img src=x onerror=alert(1)>`,
		`"><script>alert('x')</script>`,
		`a & b`,
	} {
		cmd := exec.Command(node, "-e", script)
		cmd.Env = append(os.Environ(), "CODE="+page[begin:end], "IN="+in)
		raw, err := cmd.Output()
		if err != nil {
			t.Fatalf("node: %v", err)
		}
		var got []string
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if strings.ContainsAny(got[0], `<>"'`) {
			t.Fatalf("%q escaped to %q", in, got[0])
		}
		if got[1] != "" || got[2] != "" || got[3] != "42" {
			t.Fatalf("null, undefined and a number escaped to %q", got[1:])
		}
	}
	cmd := exec.Command(node, "-e", script)
	cmd.Env = append(os.Environ(), "CODE="+page[begin:end], "IN=<b>&")
	raw, _ := cmd.Output()
	if !strings.Contains(string(raw), `&lt;b&gt;&amp;`) {
		t.Fatalf("escape of <b>& is %s", raw)
	}
}

// viewerRender runs the viewer page's own script in Node with a stub document,
// location and fetch, calls showRecord and showReview on markup-bearing data,
// and returns the HTML each one wrote.
func viewerRender(t *testing.T, record map[string]any, rows []map[string]any, pr, commit string) (recordHTML, reviewHTML string) {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed: the page's rendering is UNRUN")
	}
	page := string(ViewerPage())
	begin := strings.Index(page, "<script>")
	end := strings.Index(page, "</script>")
	if begin < 0 || end < begin {
		t.Fatal("the page's script is missing")
	}
	data, _ := json.Marshal(map[string]any{"record": record, "rows": rows, "pr": pr, "commit": commit})
	script := `const data = JSON.parse(process.env.DATA);
const out = { innerHTML: "" };
const document = { getElementById: function () { return out; } };
const location = { pathname: "/proven_records/_design/evidence/viewer.html", search: "" };
const fetch = function (url) {
  const body = url.indexOf("_view") >= 0 ? { rows: data.rows } : data.record;
  return Promise.resolve({ ok: true, json: function () { return Promise.resolve(body); } });
};
const page = new Function("document", "location", "fetch", process.env.CODE + "; return { showRecord: showRecord, showReview: showReview };")(document, location, fetch);
const settle = function () { return new Promise(function (r) { setTimeout(r, 20); }); };
(async function () {
  out.innerHTML = "";
  page.showRecord("rec-1");
  await settle();
  const one = out.innerHTML;
  out.innerHTML = "";
  page.showReview(data.pr, data.commit);
  await settle();
  console.log(JSON.stringify([one, out.innerHTML]));
})();`
	cmd := exec.Command(node, "-e", script)
	cmd.Env = append(os.Environ(), "DATA="+string(data), "CODE="+page[begin+len("<script>"):end])
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var got []string
	if err := json.Unmarshal(raw, &got); err != nil || len(got) != 2 {
		t.Fatalf("node output %q: %v", raw, err)
	}
	return got[0], got[1]
}

// TestViewerShowsEveryRecordValueAsText proves the page, rendered the way a
// browser would run it, writes no markup that came from a record. Every text
// field of a record, of each kind of its evidence and of a review's list rows
// is wrapped in an <i> tag, a tag the page never writes itself, and the title
// also carries an element with an event attribute. The test fails when any <i>
// or the event element is in the written HTML, and when the escaped text of a
// field is not. The mutated line of the negative control is any one esc(...) call
// in showRecord, evidenceHTML or showReview: dropping it lets that field's <i>
// through.
func TestViewerShowsEveryRecordValueAsText(t *testing.T) {
	mark := func(name string) string { return "<i>" + name + "</i>" }
	record := map[string]any{
		"title": `Rounds <b onmouseover="x">down</b>`, "message": mark("message"), "severity": mark("severity"),
		"path": mark("path"), "start_line": mark("line"), "proof_level": mark("level"),
		"pr": mark("pr"), "commit": mark("commit"), "run": mark("run"), "finding_id": mark("finding"),
		"evidence": []map[string]any{
			{"kind": "query", "form": mark("form"), "engine": mark("engine"), "snapshot": mark("snapshot"), "query": mark("query"),
				"columns": []string{mark("column")}, "rows": [][]string{{mark("cell")}}},
			{"kind": "test", "path": mark("testpath"), "name": mark("testname"), "command": mark("testcommand"), "output": mark("testoutput")},
			{"kind": "tool", "tool": mark("tool"), "version": mark("version"), "command": mark("toolcommand"), "output": mark("tooloutput"), "base_output": mark("baseoutput")},
			{"kind": mark("kind")},
		},
	}
	rows := []map[string]any{{"id": mark("rowid"), "value": map[string]any{"title": record["title"], "severity": mark("rowseverity"), "proof_level": mark("rowlevel"), "path": mark("rowpath")}}}
	one, many := viewerRender(t, record, rows, mark("askpr"), mark("askcommit"))
	for name, html := range map[string]string{"showRecord": one, "showReview": many} {
		if html == "" {
			t.Fatalf("%s wrote nothing", name)
		}
		if strings.Contains(html, "<i>") || strings.Contains(html, "<i ") || strings.Contains(html, "<b onmouseover") {
			t.Errorf("%s wrote markup from a record:\n%s", name, html)
		}
		if !strings.Contains(html, "&lt;b onmouseover=&quot;x&quot;&gt;down&lt;/b&gt;") {
			t.Errorf("%s did not show the title as text:\n%s", name, html)
		}
	}
	// Each field reached the page and was shown as its characters.
	for _, field := range []string{"message", "severity", "path", "line", "level", "pr", "commit", "run", "finding", "form", "engine", "snapshot", "query", "column", "cell", "testpath", "testname", "testcommand", "testoutput", "tool", "version", "toolcommand", "tooloutput", "baseoutput", "kind"} {
		if !strings.Contains(one, "&lt;i&gt;"+field+"&lt;/i&gt;") {
			t.Errorf("showRecord did not show %q as text", field)
		}
	}
	// The row id goes only into a link's address, percent-encoded, so it is not shown as text.
	for _, field := range []string{"askpr", "askcommit", "rowseverity", "rowlevel", "rowpath"} {
		if !strings.Contains(many, "&lt;i&gt;"+field+"&lt;/i&gt;") {
			t.Errorf("showReview did not show %q as text", field)
		}
	}
}

// TestQueryKeysPullOfAndRecordID covers the small readers of a record: the
// CVEs and packages a query's rows name, the pull request of a ref, and the
// document id.
func TestQueryKeysPullOfAndRecordID(t *testing.T) {
	cves, pkgs := QueryKeys([]string{"id", "aliases", "package"}, [][]string{{"GHSA-x", "CVE-2021-44228,CVE-2021-45046", "log4j-core"}, {"GHSA-y", "CVE-2021-44228", "log4j-core"}})
	if !reflect.DeepEqual(cves, []string{"CVE-2021-44228", "CVE-2021-45046"}) || !reflect.DeepEqual(pkgs, []string{"log4j-core"}) {
		t.Fatalf("keys %v %v", cves, pkgs)
	}
	if c, p := QueryKeys([]string{"note"}, [][]string{{"CVE-2021-44228"}}); c != nil || p != nil {
		t.Fatalf("an unknown column yielded %v %v", c, p)
	}
	for ref, want := range map[string]string{"refs/pull/42/merge": "42", "refs/heads/main": "", "": "", "refs/pull//merge": ""} {
		if got := PullOf(ref); got != want {
			t.Errorf("PullOf(%q) = %q, want %q", ref, got, want)
		}
	}
	if got := RecordID("", "", "20261025t070000z", "F 1/../x"); !strings.HasPrefix(got, "local-tree-20261025t070000z-f-1-x-") || idUnsafe.MatchString(got) {
		t.Fatalf("id %q", got)
	}
}

// Property: two finding ids that differ only in a character the document id
// cannot hold get two record ids, so neither loses its record in one bulk.
// An id made only of safe characters keeps its plain spelling.
// Mutated field: the finding id.
func TestRecordIDKeepsFoldedFindingIDsApart(t *testing.T) {
	a := RecordID("7", "abc", "run1", "F.1")
	b := RecordID("7", "abc", "run1", "f-1")
	if a == b {
		t.Fatalf("both ids are %s", a)
	}
	if got := RecordID("7", "abc", "run1", "F1"); got != "7-abc-run1-f1" {
		t.Fatalf("a safe id changed: %s", got)
	}
}
