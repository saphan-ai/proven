// Package osv is the dependency check of the review. It takes the packages
// a change added or raised (package manifest), queries the OSV package table
// through the evidence seam once per package, and compares the version in
// Go, not in the query. It runs with no model. The seam takes the
// osv-by-package request, and the store builds the form it speaks
// (ADR-0032, ADR-0046, ADR-0047). A matching range
// is a candidate whose evidence is the query and the rows it returned
// (ADR-0002), and the Judge keeps or drops it like any other finding. A version outside every range is
// a trace line, never a finding. The candidate also carries the smallest fixed version that leaves
// every returned range, with the advisory row that shows it (ADR-0042 point 5, fix.go), and the
// ecosystem, so a later step can label reachability. This package does not run govulncheck (ADR-0043).
//
// The package also defines the OSV table itself: its columns, the setting
// that names it (PROVEN_OSV_TABLE, ADR-0013), the setting that names the
// KEV view (PROVEN_KEV_VIEW), and the converter behind cmd/osv-to-csv that
// flattens OSV advisories into its rows. OSV by package cannot be a Zetaris
// REST table (ADR-0023), so the table is a file an administrator puts in
// Zetaris's storage. docs/OSV_TABLE.md is the contract.
package osv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/pep503"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/source"
)

// RuleID is the kind of a dependency finding this check names.
const RuleID = "osv-affected-package"

// KEVView is the known-exploited view used when PROVEN_KEV_VIEW is empty.
// The name is a setting (ADR-0013): a review loads it and Check reads it
// from the context. Sample mode does not set the context, so this default
// is what a sample review queries, and the variable is not read.
const KEVView = querytpl.DefaultKEVView

// kevViewKey is the context key of the view WithKEVView stored.
type kevViewKey struct{}

// WithKEVView returns a context that carries the KEV view Check queries.
// The review sets it from PROVEN_KEV_VIEW after that setting loads.
// Sample mode does not set it, so Check uses KEVView and does not read
// the environment.
func WithKEVView(ctx context.Context, view string) context.Context {
	return context.WithValue(ctx, kevViewKey{}, view)
}

// KEVViewOf is the KEV view WithKEVView stored on ctx, as it was set and not
// yet checked; empty when the context carries none. A caller that goes on to
// build a query gives it to querytpl, which refuses a name that is not plain.
func KEVViewOf(ctx context.Context) string {
	view, _ := ctx.Value(kevViewKey{}).(string)
	return strings.TrimSpace(view)
}

// LoadedKEVView is the view on ctx, or KEVView when the context carries none.
// The name is pasted into the SQL, so a value that is not a view is refused
// here as well as at load. The pipeline reads it too, so the Investigator's
// and the Skeptic's queries are built from the same setting (ADR-0032).
func LoadedKEVView(ctx context.Context) (string, error) {
	view := KEVViewOf(ctx)
	if view == "" {
		view = KEVView
	}
	if !ValidView(view) {
		return "", fmt.Errorf("osv: view name %q is not allowed", view)
	}
	return view, nil
}

// Outcome is what one dependency check found. Notes are the trace lines for
// a module that did not become a finding, and for a check that did not run.
// Down is set when a query named a failed service and the candidates that
// were already proved are still returned.
type Outcome struct {
	Candidates []candidate.Candidate
	Notes      []string
	Down       *service.Down
	// Facts are the packages and advisory ids the change is affected by, as
	// Proven derived them. The pipeline binds a model's query to the change
	// with them.
	Facts Facts
}

// Facts is what the dependency check established about the change: the
// packages the change added or raised, the packages whose changed version
// lies inside an advisory range, and the ids (advisory ids and aliases such
// as CVE ids) of those advisories. A model's words are not in it.
type Facts struct {
	packages map[string]bool
	ids      map[string]bool
	// raised is every package the change added or raised, whether or not an
	// OSV range matched. The ADR-0039 templates bind to it. packages stays
	// the affected set, so a raised package does not become an OSV hit.
	raised map[string]bool
}

func (f *Facts) add(ecosystem, pkg string, row Row) {
	if f.packages == nil {
		f.packages = map[string]bool{}
		f.ids = map[string]bool{}
	}
	f.packages[ecosystem+"\x00"+pkg] = true
	if row.ID != "" {
		f.ids[row.ID] = true
	}
	for _, id := range strings.Fields(row.Aliases) {
		f.ids[id] = true
		f.ids[strings.ToUpper(id)] = true
	}
}

// Binds reports whether a query request is about the change: an OSV package
// lookup of a package the change is affected by, or a lookup by an advisory
// id or CVE id of an advisory that affects it. A malicious-package, scorecard
// or licence lookup binds a package the change added or raised, on the
// ecosystem's normalised name, so two spellings of one PyPI package bind.
// A template or a value that names anything else binds nothing.
func (f Facts) Binds(req querytpl.Request) bool {
	switch req.Template {
	case querytpl.OSVByPackage:
		return f.packages[req.Params[querytpl.ParamEcosystem]+"\x00"+req.Params[querytpl.ParamPackage]]
	case querytpl.OSVByAdvisory:
		return f.ids[req.Params[querytpl.ParamAdvisoryID]]
	case querytpl.KEVByCVE, querytpl.EPSSByCVE:
		return f.ids[req.Params[querytpl.ParamCVEID]]
	case querytpl.MalByPackage, querytpl.ScorecardByPackage, querytpl.SPDXByPackage:
		// These name a package the change added or raised. They do not need
		// an OSV range: a malicious package may have no vulnerability row.
		// The key is raisedKey, the same one noteRaised wrote.
		return f.raised[raisedKey(req.Params[querytpl.ParamEcosystem], req.Params[querytpl.ParamPackage])]
	case querytpl.CVSSByCVE:
		// CVSS is read beside EPSS, for a CVE the dependency check already
		// tied to the change. A CVE only a model typed does not bind.
		return f.ids[req.Params[querytpl.ParamCVEID]]
	}
	return false
}

// noteRaised records a package the change added or raised, under the
// ecosystem's normalised name, the same key Binds reads. Check calls it for
// every hit, including one with no OSV row, so the ADR-0039 templates can
// bind. It does not mark the package affected.
func (f *Facts) noteRaised(ecosystem, pkg string) {
	if f.raised == nil {
		f.raised = map[string]bool{}
	}
	f.raised[raisedKey(ecosystem, pkg)] = true
}

// raisedKey is the key a raised package is stored and bound under. PyPI
// names are the PEP 503 form and Packagist names are lower case, which is
// how those tables are keyed. Go, npm and crates.io keep the spelling: folding
// them would merge packages that are not the same.
func raisedKey(ecosystem, pkg string) string {
	switch ecosystem {
	case "PyPI":
		pkg = pep503.Normalise(pkg)
	case "Packagist":
		pkg = strings.ToLower(pkg)
	}
	return ecosystem + "\x00" + pkg
}

// Check reads the packages the registered adapters report, queries the OSV
// table once per package, and compares the version with that adapter.
// A lock file contributes the packages it pins. When a manifest and its lock
// file both name a package, the lock file's version is the one compared.
// A lock file that cannot be parsed adds a trace note and no finding.
// A matching range becomes a candidate whose evidence is the query, its rows
// and the comparison. A CVE alias is also looked up in the KEV view.
func Check(ctx context.Context, q query.Runner, root, table, snapshot string) (Outcome, error) {
	// An empty name means the default, as the setting documents.
	if table == "" {
		table = DefaultTable
	}
	// The name is written into the SQL as it is, so it is checked here too,
	// not only when the setting is loaded.
	if !ValidTable(table) {
		return Outcome{}, fmt.Errorf("osv: table name %q is not allowed", table)
	}
	files, err := change.Read(root)
	if err != nil {
		return Outcome{}, err
	}
	// ADR-0038: the languages live in the adapter list. Collect pairs a
	// manifest with its lock file, so a package both name is queried once,
	// at the lock file's version.
	hits, notes, err := manifest.Collect(files, manifest.Readers())
	if err != nil {
		return Outcome{}, err
	}
	var out Outcome
	out.Notes = append(out.Notes, notes...)
	// No dependency edit, no query: the trace still says why nothing was checked.
	if len(hits) == 0 {
		out.Notes = append(out.Notes, "no dependency added or raised")
		return out, nil
	}
	// With no query engine nothing can be evidence. The check names the
	// service as down instead of reporting a clean result.
	if q == nil {
		out.Down = &service.Down{Service: "query", Reason: "no query engine"}
		out.Notes = append(out.Notes, "osv check did not run: query: no query engine")
		return out, out.Down
	}
	for _, hit := range hits {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		// Recorded before the OSV query, so a package with no advisory row
		// still binds a malicious-package, scorecard or licence template.
		out.Facts.noteRaised(hit.Ecosystem, hit.Package)
		next, note, err := oneModule(ctx, q, table, snapshot, hit, &out.Facts)
		if note != "" {
			out.Notes = append(out.Notes, note)
		}
		if next != nil {
			out.Candidates = append(out.Candidates, *next)
		}
		if err != nil {
			// A service that went down ends the check but not the review: the
			// candidates already proved are kept, and the review names the service.
			if d, ok := service.As(err); ok {
				out.Down = d
				return out, nil
			}
			return out, err
		}
	}
	return out, nil
}

// oneModule runs the OSV query for one package and compares its version with
// every row. It returns a candidate when at least one row matches, and
// otherwise a trace note that says why.
func oneModule(ctx context.Context, q query.Runner, table, snapshot string, hit manifest.Hit, facts *Facts) (*candidate.Candidate, string, error) {
	// ADR-0032, ADR-0046, ADR-0047: the seam takes the osv-by-package
	// request. The store builds the form. A package name outside its
	// ecosystem's format has no query: the check says so in a trace note
	// and goes on to the next package, as it does for a package with no
	// affected row. The name is the one thing the note must show a reader,
	// and it is quoted with %q so a control byte or invalid UTF-8 in it
	// reaches the trace and the terminal as an escape, not raw.
	res, err := q.Run(ctx, querytpl.Request{
		Template: querytpl.OSVByPackage,
		Params: map[string]string{
			querytpl.ParamEcosystem: hit.Ecosystem,
			querytpl.ParamPackage:   hit.Package,
		},
	}, querytpl.Relations{OSVTable: table})
	if err != nil {
		var refusal *querytpl.Refusal
		if errors.As(err, &refusal) {
			if refusal.Reason == querytpl.ReasonUnknownTemplate {
				return nil, "", fmt.Errorf("osv: template %s is not registered", querytpl.OSVByPackage)
			}
			return nil, fmt.Sprintf("%q %q: osv query refused: %s", hit.Package, hit.Version, refusal.Error()), nil
		}
		if d, ok := service.As(err); ok {
			return nil, fmt.Sprintf("%s %s: osv query did not run: %s", hit.Package, hit.Version, d.Error()), err
		}
		return nil, "", err
	}
	rows, err := readRows(res)
	if err != nil {
		return nil, "", err
	}
	// No row is not proof that the package is safe. The note says the query
	// returned nothing and names the form with the text, so a reader can run
	// it again. CouchDB's text is Mango JSON; a line that always says SQL
	// would tell the reader to run that JSON as SQL (ADR-0047).
	if len(rows) == 0 {
		note := fmt.Sprintf("%s %s: no affected row; the query returned no rows; %s", hit.Package, hit.Version, queryText(res.Form, res.SQL))
		// A version that could not be compared at all is a second reason,
		// and the note names it too.
		if why := Uncompared(hit.Ecosystem, hit.Version); why != "" {
			note += "; " + why
		}
		return nil, note, nil
	}
	var matched []compared
	var missed []compared
	for _, row := range rows {
		ok, sentence := InRangeEcosystem(hit.Ecosystem, hit.Version, row.Introduced, row.Fixed, row.LastAffected)
		item := compared{row: row, sentence: sentence}
		if ok {
			matched = append(matched, item)
		} else {
			missed = append(missed, item)
		}
	}
	if len(matched) == 0 {
		return nil, missNote(hit, res.Form, res.SQL, missed), nil
	}
	// The change's version is inside these rows' ranges, so their package and
	// their advisory ids are facts about the change.
	for _, item := range matched {
		facts.add(hit.Ecosystem, hit.Package, item.row)
	}
	// The row's severity is not mapped: sources write it in different forms,
	// so every match is reported at one level.
	c := candidate.Candidate{
		Risk: candidate.Risk{
			ID:        "osv:" + hit.Package,
			RuleID:    RuleID,
			Title:     fmt.Sprintf("Dependency %s %s is in an OSV affected range", hit.Package, hit.Version),
			Message:   findingMessage(hit, matched, missed),
			Severity:  "ERROR",
			Path:      hit.File,
			StartLine: hit.Line,
			EndLine:   hit.Line,
		},
		// The pipeline reads the ecosystem when it labels reachability (ADR-0043).
		// This check does not run govulncheck.
		Ecosystem: hit.Ecosystem,
	}
	// ADR-0002: the evidence is the query and every row it returned, the rows
	// that missed included, so a reader can redo the comparison.
	c.Evidence = append(c.Evidence, queryEvidence(res, snapshot))
	// ADR-0042 point 5: the fix is the smallest version that leaves the affected range,
	// named with the advisory row that shows it. No test proves a version bump, so the
	// fix is never marked proved.
	c.Fix = dependencyFix(hit, rows)
	// ADR-0004: a CVE alias of a matching advisory is also looked up in CISA
	// KEV, once per id. A missing KEV row does not drop the finding; the
	// message only says so.
	// ADR-0013: the view is a setting. The pipeline call does not take it,
	// so the loaded name rides the context. Empty means KEVView.
	view, verr := LoadedKEVView(ctx)
	if verr != nil {
		return nil, "", verr
	}
	seen := map[string]struct{}{}
	var kevNotes []string
	for _, item := range matched {
		for _, id := range cveAliases(item.row.Aliases) {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			// ADR-0046, ADR-0047: the KEV adapter owns kev-by-cve. The seam
			// passes the request; the store builds the form.
			kres, kerr := q.Run(ctx, querytpl.Request{
				Template: querytpl.KEVByCVE,
				Params:   map[string]string{querytpl.ParamCVEID: id},
			}, querytpl.Relations{KEVView: view})
			if kerr != nil {
				var refusal *querytpl.Refusal
				if errors.As(kerr, &refusal) && refusal.Reason == querytpl.ReasonUnknownTemplate {
					return nil, "", fmt.Errorf("osv: template %s is not registered", querytpl.KEVByCVE)
				}
				if d, ok := service.As(kerr); ok {
					kevNotes = append(kevNotes, "alias "+id+" was not looked up in KEV: "+d.Error())
					// The OSV match is already evidence, so the candidate is kept and the
					// KEV service is named as down.
					addNotes(&c, kevNotes)
					return &c, strings.Join(kevNotes, "; "), d
				}
				return nil, "", kerr
			}
			// The KEV rule is the decision: a row is evidence, no row is not.
			// The evidence attached below stays queryEvidence, which marks
			// the rows Bound. The rule does not know that mark.
			kev, kok := source.ByID("kev")
			if !kok {
				return nil, "", fmt.Errorf("osv: kev adapter is not registered")
			}
			if _, ruled := kev.Rule(source.Evidence{Rows: kres.Rows}); !ruled {
				kevNotes = append(kevNotes, "alias "+id+" is not a KEV row")
				continue
			}
			c.Evidence = append(c.Evidence, queryEvidence(kres, snapshot))
			kevNotes = append(kevNotes, "alias "+id+" is on the KEV list")
		}
	}
	addNotes(&c, kevNotes)
	return &c, "", nil
}

// addNotes appends the KEV sentences to the finding's message.
func addNotes(c *candidate.Candidate, notes []string) {
	if len(notes) == 0 {
		return
	}
	c.Message += " " + strings.Join(notes, " ")
}

// compared is one OSV row and the sentence of its comparison.
type compared struct {
	row      Row
	sentence string
}

// missNote is the trace line for a package whose rows all missed. It names
// each row and why it missed, and the form of the query text.
func missNote(hit manifest.Hit, form, sqlText string, missed []compared) string {
	parts := make([]string, 0, len(missed))
	for _, item := range missed {
		parts = append(parts, item.row.ID+": "+item.sentence)
	}
	return fmt.Sprintf("%s %s: no affected row; %s; %s", hit.Package, hit.Version, queryText(form, sqlText), strings.Join(parts, "; "))
}

// queryText is the query as a trace line names it: the form, then the text.
// form is "sql" or "mango". The text on CouchDB is Mango JSON (ADR-0047).
func queryText(form, text string) string {
	return fmt.Sprintf("query (%s): %s", form, text)
}

// findingMessage names every row that matched and every row that did not,
// so the review shows the whole comparison, not only the hit.
func findingMessage(hit manifest.Hit, matched, missed []compared) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s matches %s: %s.", hit.Package, hit.Version, matched[0].row.ID, matched[0].sentence)
	for _, item := range matched[1:] {
		fmt.Fprintf(&b, " %s matches %s: %s.", hit.Package, item.row.ID, item.sentence)
	}
	for _, item := range missed {
		id := item.row.ID
		if id == "" {
			id = "a row"
		}
		fmt.Fprintf(&b, " %s does not match: %s.", id, item.sentence)
	}
	return b.String()
}

// queryEvidence copies a query result into evidence. The engine's own
// snapshot wins; the review's snapshot is used when the engine gives none.
// The rows are Bound: Check built the query from the change's own manifests
// and a matching range, or from an alias of a matching row.
func queryEvidence(res query.Result, snapshot string) candidate.Evidence {
	snap := res.Snapshot
	if snap == "" {
		snap = snapshot
	}
	return candidate.Evidence{
		Kind: candidate.KindQuery,
		Query: &candidate.QueryEvidence{
			SQL: res.SQL, Form: res.Form, Engine: res.Engine, Snapshot: snap,
			Columns: append([]string(nil), res.Columns...),
			Rows:    copyRows(res.Rows),
			Bound:   true,
		},
	}
}

// readRows maps the result by column name, not by position, so a table
// whose columns come in another order still reads. A missing column
// refuses the result: a row read without it could match by mistake.
func readRows(res query.Result) ([]Row, error) {
	idx := map[string]int{}
	for i, c := range res.Columns {
		idx[strings.ToLower(c)] = i
	}
	for _, name := range Columns {
		if _, ok := idx[name]; !ok {
			return nil, fmt.Errorf("osv: result has no column %s", name)
		}
	}
	out := make([]Row, 0, len(res.Rows))
	for _, raw := range res.Rows {
		cell := func(name string) string {
			i := idx[name]
			if i >= len(raw) {
				return ""
			}
			return raw[i]
		}
		out = append(out, Row{
			ID: cell("id"), Aliases: cell("aliases"), Ecosystem: cell("ecosystem"), Package: cell("package"),
			Introduced: cell("introduced"), Fixed: cell("fixed"), LastAffected: cell("last_affected"),
			Severity: cell("severity"), Published: cell("published"),
		})
	}
	return out, nil
}

// cveAliases picks the CVE ids out of the space-separated aliases column.
func cveAliases(aliases string) []string {
	var out []string
	for _, part := range strings.Fields(aliases) {
		id := strings.ToUpper(part)
		if querytpl.ValidCVE(id) {
			out = append(out, id)
		}
	}
	return out
}

// copyRows copies the rows so the evidence does not share memory with the
// adapter's result. nil becomes an empty list, so evidence always carries one.
func copyRows(rows [][]string) [][]string {
	if rows == nil {
		return [][]string{}
	}
	out := make([][]string, len(rows))
	for i, row := range rows {
		out[i] = append([]string(nil), row...)
	}
	return out
}
