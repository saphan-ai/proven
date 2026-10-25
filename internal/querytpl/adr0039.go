package querytpl

import (
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/source"
)

// youngReleaseDays is the age under which a release counts as only days old
// (ADR-0039). Fourteen days is not that: a release of that age has left the
// first fortnight. The number is the rule's, written once so the test and
// the sentence name the same bound.
const youngReleaseDays = 14

// permissiveLicences is the closed set a licence cell may name and still
// not be evidence. It stands in for an organisation policy (ADR-0036): that
// policy is not this program's store, and the cell is compared whole, so
// "MIT OR Apache-2.0" is not two allowed identifiers.
var permissiveLicences = map[string]struct{}{
	"0BSD":         {},
	"Apache-2.0":   {},
	"BSD-2-Clause": {},
	"BSD-3-Clause": {},
	"ISC":          {},
	"MIT":          {},
	"Unlicense":    {},
}

// packageSQL builds a query that reads one package from table.
// columns is the select list, and it is the same list the mango form returns.
func packageSQL(template, table string, columns []string, ecosystem, pkg string) (string, error) {
	if !ValidRelation(table) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: template}
	}
	eco, name, err := ecosystemKey(template, ecosystem, pkg)
	if err != nil {
		return "", err
	}
	return "SELECT " + strings.Join(columns, ", ") + " FROM " + table +
		" WHERE ecosystem = '" + eco + "' AND package = '" + name + "'", nil
}

// cvssSQL is the CVSS query for one CVE id. The table is the adapter's own.
func cvssSQL(cve string) (string, error) {
	if !ValidRelation(DefaultCVSSTable) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: CVSSByCVE}
	}
	id, err := literal(CVSSByCVE, ParamCVEID, cve, cveFormat)
	if err != nil {
		return "", err
	}
	return "SELECT " + strings.Join(cvssColumns, ", ") + " FROM " + DefaultCVSSTable + " WHERE cve_id = " + id, nil
}

// malAdapter is the OpenSSF malicious-package source (ADR-0039, ADR-0046).
// A row is an OSV advisory whose id starts with MAL-.
type malAdapter struct{}

func (malAdapter) ID() string { return "mal" }

func (malAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://osv.dev",
		Refresh: "an administrator replaces the table; a review reads it as it stands",
		Table:   DefaultMalTable,
	}
}

func (malAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   MalByPackage,
			Answer: "the malicious-package advisory of one package",
			Params: []source.Param{
				{Name: ParamEcosystem, Format: "one of Go, npm, PyPI, Packagist, crates.io", Match: ecosystemFormat},
				{Name: ParamPackage, Format: "a package name in the ecosystem's own spelling", Match: anyPackageFormat},
			},
		},
		Build: func(params map[string]string, _ source.Rel) (string, error) {
			return packageSQL(MalByPackage, DefaultMalTable, malColumns, params[ParamEcosystem], params[ParamPackage])
		},
	}}
}

func (malAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// An empty row list is not evidence. The table not naming the package
	// does not say the package was published to do harm (ADR-0039, ADR-0046).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	idAt := columnIndex(in.Columns, "id")
	var kept [][]string
	for _, row := range in.Rows {
		// ADR-0039: the signal is the MAL- prefix on the advisory id.
		// A row with another id is not this source's finding.
		if strings.HasPrefix(cell(row, idAt), "MAL-") {
			kept = append(kept, append([]string(nil), row...))
		}
	}
	return keepRows(in, kept)
}

// scorecardAdapter is project health from OpenSSF Scorecard through deps.dev
// (ADR-0039, ADR-0046).
type scorecardAdapter struct{}

func (scorecardAdapter) ID() string { return "scorecard" }

func (scorecardAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://deps.dev",
		Refresh: "an administrator replaces the table; a review reads it as it stands",
		Table:   DefaultScorecardTable,
	}
}

func (scorecardAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   ScorecardByPackage,
			Answer: "the project-health scores of one package",
			Params: []source.Param{
				{Name: ParamEcosystem, Format: "one of Go, npm, PyPI, Packagist, crates.io", Match: ecosystemFormat},
				{Name: ParamPackage, Format: "a package name in the ecosystem's own spelling", Match: anyPackageFormat},
			},
		},
		Build: func(params map[string]string, _ source.Rel) (string, error) {
			return packageSQL(ScorecardByPackage, DefaultScorecardTable, scorecardColumns, params[ParamEcosystem], params[ParamPackage])
		},
	}}
}

func (scorecardAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// No row means the table has no scores. Missing scores are not a
	// statement that the project is healthy (ADR-0039).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	maintainedAt := columnIndex(in.Columns, "maintained")
	signedAt := columnIndex(in.Columns, "signed")
	ageAt := columnIndex(in.Columns, "release_age_days")
	var kept [][]string
	for _, row := range in.Rows {
		if unhealthy(cell(row, maintainedAt), cell(row, signedAt), cell(row, ageAt)) {
			kept = append(kept, append([]string(nil), row...))
		}
	}
	return keepRows(in, kept)
}

// unhealthy reports the three signals ADR-0039 names: no maintainers, an
// unsigned release, or a release only days old. A cell that is not the
// failing spelling is not that signal. Empty is not a guess.
func unhealthy(maintained, signed, age string) bool {
	if maintained == "0" {
		return true
	}
	if signed == "false" {
		return true
	}
	n, err := strconv.Atoi(age)
	if err != nil {
		return false
	}
	return n >= 0 && n < youngReleaseDays
}

// cvssAdapter is the NVD CVSS source, read beside EPSS (ADR-0039, ADR-0046).
type cvssAdapter struct{}

func (cvssAdapter) ID() string { return "cvss" }

func (cvssAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://nvd.nist.gov",
		Refresh: "an administrator replaces the table; a review reads it as it stands",
		Table:   DefaultCVSSTable,
	}
}

func (cvssAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   CVSSByCVE,
			Answer: "the CVSS score of one CVE",
			Params: []source.Param{
				{Name: ParamCVEID, Format: "a CVE id such as CVE-2021-44228", Match: cveFormat},
			},
		},
		Build: func(params map[string]string, _ source.Rel) (string, error) {
			return cvssSQL(params[ParamCVEID])
		},
	}}
}

func (cvssAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// A row is a severity score, and the score is evidence beside EPSS.
	// No row means the table has no score for that CVE (ADR-0039).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	return copyEvidence(in), true
}

// spdxAdapter is the SPDX licence source (ADR-0039, ADR-0046).
type spdxAdapter struct{}

func (spdxAdapter) ID() string { return "spdx" }

func (spdxAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://spdx.org",
		Refresh: "an administrator replaces the table; a review reads it as it stands",
		Table:   DefaultSPDXTable,
	}
}

func (spdxAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   SPDXByPackage,
			Answer: "the SPDX licence of one package",
			Params: []source.Param{
				{Name: ParamEcosystem, Format: "one of Go, npm, PyPI, Packagist, crates.io", Match: ecosystemFormat},
				{Name: ParamPackage, Format: "a package name in the ecosystem's own spelling", Match: anyPackageFormat},
			},
		},
		Build: func(params map[string]string, _ source.Rel) (string, error) {
			return packageSQL(SPDXByPackage, DefaultSPDXTable, spdxColumns, params[ParamEcosystem], params[ParamPackage])
		},
	}}
}

func (spdxAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// No row means the table does not name a licence. That is not a
	// statement that the licence is allowed (ADR-0039).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	at := columnIndex(in.Columns, "licence")
	var kept [][]string
	for _, row := range in.Rows {
		licence := cell(row, at)
		if licence == "" {
			continue
		}
		if _, ok := permissiveLicences[licence]; ok {
			continue
		}
		kept = append(kept, append([]string(nil), row...))
	}
	return keepRows(in, kept)
}

// columnIndex is the index of name in columns, or -1 when the result has
// no such column. A rule that cannot see its column does not guess.
func columnIndex(columns []string, name string) int {
	for i, c := range columns {
		if c == name {
			return i
		}
	}
	return -1
}

// cell is the trimmed cell at i, or empty when i is not a cell of the row.
func cell(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

// keepRows returns evidence holding only rows. No row left is not evidence.
func keepRows(in source.Evidence, rows [][]string) (source.Evidence, bool) {
	if len(rows) == 0 {
		return source.Evidence{}, false
	}
	out := copyEvidence(in)
	out.Rows = rows
	return out, true
}
