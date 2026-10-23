// Package osv is the dependency check of the review. It takes the packages
// a change added or raised (package manifest), queries the OSV package table
// through the query adapter once per package, and compares the version in
// Go, not in SQL. It runs with no model. A matching range is a candidate whose
// evidence is the query and the rows it returned (ADR-0002), and the Judge
// keeps or drops it like any other finding. A version outside every range is
// a trace line, never a finding.
//
// The package also defines the OSV table itself: its columns, the setting
// that names it (PROVEN_OSV_TABLE, ADR-0013), and the converter behind
// cmd/osv-to-csv that flattens OSV advisories into its rows. OSV by package
// cannot be a Zetaris REST table (ADR-0023), so the table is a file an
// administrator puts in Zetaris's storage. docs/OSV_TABLE.md is the contract.
package osv

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

// RuleID is the kind of a dependency finding this check names.
const RuleID = "osv-affected-package"

// KEVView is the known-exploited view an advisory alias is looked up in.
const KEVView = "px_evidence.px_kev_vulns"

// cveID matches a CVE id, the only alias the KEV view is keyed by. Because
// an id must match it, kevSQL can place the id in the SQL with no escaping.
var cveID = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)

// Outcome is what one dependency check found. Notes are the trace lines for
// a module that did not become a finding, and for a check that did not run.
// Down is set when a query named a failed service and the candidates that
// were already proved are still returned.
type Outcome struct {
	Candidates []candidate.Candidate
	Notes      []string
	Down       *service.Down
}

// Check reads Go, npm and PyPI packages the change added or raised, queries
// the OSV table once per package, and compares the version for that ecosystem.
// A lock file contributes the packages it pins. When a manifest and its lock
// file both name a package, the lock file's version is the one compared.
// A lock file that cannot be parsed adds a trace note and no finding.
// A matching range becomes a candidate whose evidence is the query, its rows
// and the comparison. A CVE alias is also looked up in the KEV view.
func Check(ctx context.Context, q query.Adapter, root, table, snapshot string) (Outcome, error) {
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
	// Each ecosystem has a manifest reader and a lock reader. Collect pairs
	// them, so a package both name is queried once, at the lock file's version.
	hits, notes, err := manifest.Collect(files, []manifest.Reader{
		manifest.Go(), manifest.GoSum(),
		manifest.NPM(), manifest.NPMLock(),
		manifest.PyPI(), manifest.Poetry(),
	})
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
		next, note, err := oneModule(ctx, q, table, snapshot, hit)
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
func oneModule(ctx context.Context, q query.Adapter, table, snapshot string, hit manifest.Hit) (*candidate.Candidate, string, error) {
	sqlText, err := moduleSQL(table, hit.Ecosystem, hit.Package)
	if err != nil {
		return nil, "", err
	}
	res, err := q.Query(ctx, sqlText)
	if err != nil {
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
	// returned nothing and carries the SQL, so a reader can run it again.
	if len(rows) == 0 {
		note := fmt.Sprintf("%s %s: no affected row; the query returned no rows; SQL: %s", hit.Package, hit.Version, sqlText)
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
		return nil, missNote(hit, sqlText, missed), nil
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
	}
	// ADR-0002: the evidence is the query and every row it returned, the rows
	// that missed included, so a reader can redo the comparison.
	c.Evidence = append(c.Evidence, queryEvidence(res, snapshot))
	// ADR-0004: a CVE alias of a matching advisory is also looked up in CISA
	// KEV, once per id. A missing KEV row does not drop the finding; the
	// message only says so.
	seen := map[string]struct{}{}
	var kevNotes []string
	for _, item := range matched {
		for _, id := range cveAliases(item.row.Aliases) {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			kres, kerr := q.Query(ctx, kevSQL(id))
			if kerr != nil {
				if d, ok := service.As(kerr); ok {
					kevNotes = append(kevNotes, "alias "+id+" was not looked up in KEV: "+d.Error())
					// The OSV match is already evidence, so the candidate is kept and the
					// KEV service is named as down.
					addNotes(&c, kevNotes)
					return &c, strings.Join(kevNotes, "; "), d
				}
				return nil, "", kerr
			}
			if len(kres.Rows) == 0 {
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
// each row and why it missed.
func missNote(hit manifest.Hit, sqlText string, missed []compared) string {
	parts := make([]string, 0, len(missed))
	for _, item := range missed {
		parts = append(parts, item.row.ID+": "+item.sentence)
	}
	return fmt.Sprintf("%s %s: no affected row; SQL: %s; %s", hit.Package, hit.Version, sqlText, strings.Join(parts, "; "))
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
func queryEvidence(res query.Result, snapshot string) candidate.Evidence {
	snap := res.Snapshot
	if snap == "" {
		snap = snapshot
	}
	return candidate.Evidence{
		Kind: candidate.KindQuery,
		Query: &candidate.QueryEvidence{
			SQL: res.SQL, Engine: res.Engine, Snapshot: snap,
			Columns: append([]string(nil), res.Columns...),
			Rows:    copyRows(res.Rows),
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

// moduleSQL builds the one query per package. The ecosystem and the package
// are quoted literals. The table name cannot be quoted, so it is checked.
func moduleSQL(table, ecosystem, pkg string) (string, error) {
	if !ValidTable(table) {
		return "", fmt.Errorf("osv: table name %q is not allowed", table)
	}
	eco, err := quoteLiteral(ecosystem)
	if err != nil {
		return "", err
	}
	mod, err := quoteLiteral(pkg)
	if err != nil {
		return "", err
	}
	return "SELECT " + strings.Join(Columns, ", ") + " FROM " + table + " WHERE ecosystem = " + eco + " AND package = " + mod, nil
}

// kevSQL builds the KEV lookup for one CVE. id has matched cveID, so it holds
// no quote and needs no escaping.
func kevSQL(id string) string {
	return "SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM " + KEVView + " WHERE cve_id = '" + id + "'"
}

// quoteLiteral makes a SQL string literal by doubling single quotes. A line
// break or a NUL is refused outright; no package name needs one.
func quoteLiteral(s string) (string, error) {
	if strings.ContainsAny(s, "\r\n\x00") {
		return "", fmt.Errorf("osv: value %q cannot be a SQL literal", s)
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'", nil
}

// cveAliases picks the CVE ids out of the space-separated aliases column.
func cveAliases(aliases string) []string {
	var out []string
	for _, part := range strings.Fields(aliases) {
		id := strings.ToUpper(part)
		if cveID.MatchString(id) {
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
