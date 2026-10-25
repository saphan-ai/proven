package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed" // the viewer page is embedded so one binary carries what it publishes
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// designName is the name of the design document that holds the views and the
// viewer page, and viewerName the attachment that is the page. CouchDB serves
// an attachment at /<database>/_design/<designName>/<viewerName>, so the page
// needs no server of its own.
const (
	designName = "evidence"
	viewerName = "viewer.html"
)

// viewerPage is the one static HTML page the design document carries. It has
// no external resource, never names the admin console, and escapes every value
// it shows (the esc function between its markers is tested on its own).
//
//go:embed records_viewer.html
var viewerPage []byte

// ViewerPage returns the viewer page as it is published.
func ViewerPage() []byte { return append([]byte(nil), viewerPage...) }

// The map functions of the four views, run by CouchDB. A key starts with the
// fields a reader narrows by, so a range over the leading fields lists a
// review, a finding, a CVE with its package, or a proof level.
var designViews = map[string]string{
	// by_review lists a review's findings: the key is the pull request, the
	// commit, the finding and the run, and the value is what a list shows.
	"by_review": `function (doc) {
  if (doc.type === "evidence-record") {
    emit([doc.pr, doc.commit, doc.finding_id, doc.run],
      {title: doc.title, severity: doc.severity, proof_level: doc.proof_level, path: doc.path});
  }
}`,
	// by_finding finds every record of one finding id across reviews.
	"by_finding": `function (doc) {
  if (doc.type === "evidence-record") {
    emit([doc.finding_id, doc.pr, doc.commit, doc.run], {title: doc.title, proof_level: doc.proof_level});
  }
}`,
	// by_cve_package pairs each CVE a record names with each package it names.
	// A record with a CVE and no package, or the reverse, is keyed with an
	// empty text on the missing side, so neither is lost.
	"by_cve_package": `function (doc) {
  if (doc.type !== "evidence-record") { return; }
  var cves = doc.cves && doc.cves.length ? doc.cves : [""];
  var pkgs = doc.packages && doc.packages.length ? doc.packages : [""];
  if (cves[0] === "" && pkgs[0] === "") { return; }
  for (var i = 0; i < cves.length; i++) {
    for (var j = 0; j < pkgs.length; j++) {
      emit([cves[i], pkgs[j], doc.pr, doc.commit, doc.finding_id], {title: doc.title, proof_level: doc.proof_level});
    }
  }
}`,
	// by_proof_level lists records by how far a reader may trust them.
	"by_proof_level": `function (doc) {
  if (doc.type === "evidence-record") {
    emit([doc.proof_level, doc.pr, doc.commit, doc.finding_id, doc.run], {title: doc.title, severity: doc.severity});
  }
}`,
}

// designVDU is the design document's validate_doc_update, run by CouchDB on
// every write to the records database. A member of the database's security
// document may read it and may write any document that is not a design
// document, so membership alone does not make an account read-only. This
// function admits a write only from a server admin or a database admin (named
// in the security document's admins, by name or by role) and refuses every
// other account, so the read-only account, a member, cannot write a record.
const designVDU = `function (newDoc, oldDoc, userCtx, secObj) {
  var roles = userCtx.roles || [];
  if (roles.indexOf("_admin") !== -1) { return; }
  var admins = (secObj && secObj.admins) || {};
  var names = admins.names || [];
  var adminRoles = admins.roles || [];
  if (userCtx.name && names.indexOf(userCtx.name) !== -1) { return; }
  for (var i = 0; i < roles.length; i++) {
    if (adminRoles.indexOf(roles[i]) !== -1) { return; }
  }
  throw({forbidden: "Only a database admin writes evidence records."});
}`

// ViewVDU returns the validate_doc_update function the design document carries.
func ViewVDU() string { return designVDU }

// ViewNames lists the design document's views in a fixed order.
func ViewNames() []string {
	names := make([]string, 0, len(designViews))
	for n := range designViews {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ViewMap returns the map function of a view, or "" for a name that is not one.
func ViewMap(name string) string { return designViews[name] }

// EvidenceRecord is one finding's record: one CouchDB document. The fields are
// the ones the views and the viewer read.
type EvidenceRecord struct {
	PR         string        `json:"pr"`
	Commit     string        `json:"commit"`
	Run        string        `json:"run"`
	FindingID  string        `json:"finding_id"`
	RuleID     string        `json:"rule_id"`
	Title      string        `json:"title"`
	Message    string        `json:"message"`
	Severity   string        `json:"severity"`
	Path       string        `json:"path"`
	StartLine  int           `json:"start_line"`
	EndLine    int           `json:"end_line"`
	ProofLevel string        `json:"proof_level"`
	Unresolved bool          `json:"unresolved"`
	CVEs       []string      `json:"cves"`
	Packages   []string      `json:"packages"`
	Evidence   []RecordProof `json:"evidence"`
}

// RecordProof is one piece of a finding's evidence. Kind is query, test or
// tool, and only the fields of that kind are set.
type RecordProof struct {
	Kind string `json:"kind"`
	// A query: the form that ran, its text, and the rows it returned.
	Form     string     `json:"form,omitempty"`
	Query    string     `json:"query,omitempty"`
	Engine   string     `json:"engine,omitempty"`
	Snapshot string     `json:"snapshot,omitempty"`
	Columns  []string   `json:"columns,omitempty"`
	Rows     [][]string `json:"rows,omitempty"`
	// A test: its command and how it fared on the base and on the change.
	Path          string `json:"path,omitempty"`
	Name          string `json:"name,omitempty"`
	Command       string `json:"command,omitempty"`
	FailsOnChange bool   `json:"fails_on_change,omitempty"`
	PassesOnBase  bool   `json:"passes_on_base,omitempty"`
	Output        string `json:"output,omitempty"`
	// A tool: the same command run on both sides.
	Tool            string `json:"tool,omitempty"`
	Version         string `json:"version,omitempty"`
	BaseOutput      string `json:"base_output,omitempty"`
	PresentOnChange bool   `json:"present_on_change,omitempty"`
	PresentOnBase   bool   `json:"present_on_base,omitempty"`
}

var (
	cveFormat = regexp.MustCompile(`CVE-[0-9]{4}-[0-9]{4,}`)
	// idUnsafe is what a document id part may not hold, so the id needs no
	// escaping in a path and reads the same everywhere. A dot is unsafe too, so
	// no part can be `..`.
	idUnsafe = regexp.MustCompile(`[^a-z0-9_-]+`)
)

// queryKeyColumns are the columns of the registered tables that name a CVE,
// and queryPackageColumns those that name a package.
var (
	queryKeyColumns     = map[string]bool{"cve_id": true, "cve": true, "id": true, "aliases": true, "advisory_id": true}
	queryPackageColumns = map[string]bool{"package": true, "product": true}
)

// QueryKeys reads the CVE ids and the package names a query's rows hold, for
// the by_cve_package view. A column is read by its name, so a row of a table
// this function does not know simply yields nothing, and the record is still
// found by its other views.
func QueryKeys(columns []string, rows [][]string) (cves, packages []string) {
	seenC, seenP := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		for i, name := range columns {
			if i >= len(row) {
				break
			}
			name = strings.ToLower(name)
			switch {
			case queryKeyColumns[name]:
				for _, c := range cveFormat.FindAllString(row[i], -1) {
					if !seenC[c] {
						seenC[c] = true
						cves = append(cves, c)
					}
				}
			case queryPackageColumns[name]:
				if p := strings.TrimSpace(row[i]); p != "" && !seenP[p] {
					seenP[p] = true
					packages = append(packages, p)
				}
			}
		}
	}
	sort.Strings(cves)
	sort.Strings(packages)
	return cves, packages
}

// PullOf returns the pull request number a GITHUB_REF names (refs/pull/<n>/merge),
// or "" when the ref is not a pull request's.
func PullOf(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) >= 3 && parts[0] == "refs" && parts[1] == "pull" && parts[2] != "" {
		return parts[2]
	}
	return ""
}

// RecordID is the document id of one finding's record in one run. The run
// stamp is part of it, so a second review of the same commit adds records and
// never conflicts with the first: the records are a history, not one row.
func RecordID(pr, commit, run, findingID string) string {
	part := func(s string) string {
		s = idUnsafe.ReplaceAllString(strings.ToLower(s), "-")
		return strings.Trim(s, "-")
	}
	if len(commit) > 12 {
		commit = commit[:12]
	}
	return strings.Join([]string{part(orDefault(pr, "local")), part(orDefault(commit, "tree")), part(run), part(findingID)}, "-")
}

func orDefault(s, d string) string {
	if strings.TrimSpace(s) == "" {
		return d
	}
	return s
}

// RunStamp is the compact UTC time a run's records carry.
func RunStamp(t time.Time) string { return t.UTC().Format("20060102t150405z") }

// Recorder writes evidence records to CouchDB with the writing account. It is
// a different value from CouchDB, the read store: the two hold different
// credentials, and the viewer is opened with neither.
type Recorder struct {
	cfg    RecordsSettings
	client *http.Client
}

// NewRecorder builds the recorder on validated settings. A nil client gets one
// with the settings' timeout and no redirects, so the credential goes only to
// the configured address; a test may pass its own.
func NewRecorder(cfg RecordsSettings, client *http.Client) *Recorder {
	if client == nil {
		client = &http.Client{
			Timeout:       cfg.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}
	return &Recorder{cfg: cfg, client: client}
}

// Database is the records database's name.
func (r *Recorder) Database() string { return r.cfg.Database }

// Publish makes the database and the design document exist, then writes the
// records, and returns each record's document id keyed by finding id. A record
// that CouchDB refuses is left out of the map, so a caller links only what was
// stored. It does nothing for an empty list of records, so a review with no
// finding writes no document.
func (r *Recorder) Publish(ctx context.Context, recs []EvidenceRecord) (map[string]string, error) {
	if r == nil || r.client == nil {
		return nil, couchFailed("not configured")
	}
	if len(recs) == 0 {
		return map[string]string{}, nil
	}
	if err := r.ensureDatabase(ctx); err != nil {
		return nil, err
	}
	if err := r.ensureDesign(ctx); err != nil {
		return nil, err
	}
	docs := make([]map[string]any, len(recs))
	ids := make([]string, len(recs))
	for i, rec := range recs {
		raw, err := json.Marshal(rec)
		if err != nil {
			return nil, couchFailed("not configured")
		}
		doc := map[string]any{}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, couchFailed("not configured")
		}
		ids[i] = RecordID(rec.PR, rec.Commit, rec.Run, rec.FindingID)
		doc["_id"] = ids[i]
		doc["type"] = "evidence-record"
		docs[i] = doc
	}
	body, err := json.Marshal(map[string]any{"docs": docs})
	if err != nil {
		return nil, couchFailed("not configured")
	}
	status, raw, err := r.do(ctx, http.MethodPost, "/"+r.cfg.Database+"/_bulk_docs", body)
	if err != nil {
		return nil, err
	}
	if status != http.StatusCreated && status != http.StatusAccepted {
		return nil, couchDown(status)
	}
	// _bulk_docs answers 201 even when single documents are refused; each
	// answer names its document. A document with an error is not linked.
	var answers []struct {
		ID    string `json:"id"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &answers); err != nil {
		return nil, couchFailed("bad response")
	}
	refused := map[string]bool{}
	for _, a := range answers {
		if a.Error != "" {
			refused[a.ID] = true
		}
	}
	out := map[string]string{}
	for i, rec := range recs {
		if !refused[ids[i]] {
			out[rec.FindingID] = ids[i]
		}
	}
	if len(refused) > 0 {
		return out, couchFailed(fmt.Sprintf("%d of %d records refused", len(refused), len(recs)))
	}
	return out, nil
}

// CheckViewer fetches the viewer page of the records database as the
// read-only account, which is the account a reader opens the link with, and
// names a down store when the page is not served to it or is not this build's.
// It only reads (GET): a link is shown only for a page the reader's account is
// seen to get, and the check itself needs no write right.
func (c *CouchDB) CheckViewer(ctx context.Context, database string) error {
	if c == nil || c.client == nil {
		return couchFailed("not configured")
	}
	status, raw, err := c.do(ctx, http.MethodGet, "/"+database+"/_design/"+designName+"/"+viewerName, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return couchDown(status)
	}
	if !bytes.Equal(raw, viewerPage) {
		return couchFailed("the viewer page served is not this build's")
	}
	return nil
}

// ensureDatabase creates the records database. 412 means it exists, which is
// the normal case after the first review.
func (r *Recorder) ensureDatabase(ctx context.Context) error {
	status, _, err := r.do(ctx, http.MethodPut, "/"+r.cfg.Database, nil)
	if err != nil {
		return err
	}
	switch status {
	case http.StatusCreated, http.StatusAccepted, http.StatusPreconditionFailed:
		return nil
	}
	return couchDown(status)
}

// designVersion is a digest of everything the design document holds, kept in
// the document so an unchanged one is not written again: every write is a new
// revision CouchDB keeps.
func designVersion() string {
	h := sha256.New()
	for _, n := range ViewNames() {
		fmt.Fprintf(h, "%s\x00%s\x00", n, designViews[n])
	}
	fmt.Fprintf(h, "vdu\x00%s\x00", designVDU)
	h.Write(viewerPage)
	return hex.EncodeToString(h.Sum(nil))
}

// ensureDesign writes the design document, with the viewer page as its
// attachment, unless the stored one already carries this build's version.
func (r *Recorder) ensureDesign(ctx context.Context) error {
	path := "/" + r.cfg.Database + "/_design/" + designName
	status, raw, err := r.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	rev := ""
	switch status {
	case http.StatusOK:
		var cur struct {
			Rev     string `json:"_rev"`
			Version string `json:"proven_version"`
		}
		if err := json.Unmarshal(raw, &cur); err != nil {
			return couchFailed("bad response")
		}
		if cur.Version == designVersion() {
			return nil
		}
		rev = cur.Rev
	case http.StatusNotFound:
	default:
		return couchDown(status)
	}
	views := map[string]any{}
	for n, m := range designViews {
		views[n] = map[string]string{"map": m}
	}
	doc := map[string]any{
		"_id":                 "_design/" + designName,
		"language":            "javascript",
		"views":               views,
		"validate_doc_update": designVDU,
		"proven_version":      designVersion(),
		"_attachments": map[string]any{
			viewerName: map[string]string{
				"content_type": "text/html; charset=utf-8",
				"data":         base64.StdEncoding.EncodeToString(viewerPage),
			},
		},
	}
	if rev != "" {
		doc["_rev"] = rev
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return couchFailed("not configured")
	}
	status, _, err = r.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusAccepted {
		return couchDown(status)
	}
	return nil
}

// do sends one request as the writing account and returns the status and the
// answer, read up to maxAnswer. A transport failure is a down store named
// couchdb; the credential never appears in it.
func (r *Recorder) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, r.cfg.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, r.cfg.BaseURL+path, rd)
	if err != nil {
		return 0, nil, couchFailed("not configured")
	}
	req.SetBasicAuth(r.cfg.User, r.cfg.Password)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := r.client.Do(req)
	if err != nil {
		if perr := parent.Err(); perr != nil {
			return 0, nil, perr
		}
		return 0, nil, couchFailed("unreachable")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxAnswer))
	if err != nil {
		return 0, nil, couchFailed("unreachable")
	}
	return resp.StatusCode, raw, nil
}
