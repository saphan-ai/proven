package querytpl

import (
	"encoding/json"
	"strings"
)

// The columns each template returns, in the order its sql form selects them.
// ADR-0047: the mango form lists the same fields in the same order, so a row
// from CouchDB lines up with a row from Zetaris. A test holds each list to the
// SELECT list of its sql form.
var (
	osvAdvisoryColumns = []string{"id", "alias", "published", "severity"}
	kevColumns         = []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}
	epssColumns        = []string{"cve_id", "epss", "percentile", "score_date"}
	malColumns         = []string{"id", "ecosystem", "package", "summary"}
	scorecardColumns   = []string{"ecosystem", "package", "maintained", "signed", "release_age_days"}
	cvssColumns        = []string{"cve_id", "cvss_score", "severity", "source"}
	spdxColumns        = []string{"ecosystem", "package", "licence"}
)

// Find is one template built as a CouchDB Mango query (ADR-0047).
// Table is the registered table the query reads; the store maps it to one
// database. Selector holds one equality per parameter, as the strings that
// passed the same check the sql form uses (ADR-0032). Fields is the column
// order of the answer. A store sends Selector and Fields to the database and
// never reads them back out of the Mango text.
type Find struct {
	Table    string            `json:"table"`
	Selector map[string]string `json:"selector"`
	Fields   []string          `json:"fields"`
}

// buildFind validates the request the way Build does and returns the Mango
// query for it. rel is already resolved. The relation names are checked
// again: a table name becomes a database name, and only a plain name may.
func buildFind(name string, params map[string]string, rel Relations) (Find, error) {
	switch name {
	case OSVByPackage:
		if !ValidRelation(rel.OSVTable) {
			return Find{}, &Refusal{Reason: ReasonBadRelation, Detail: name}
		}
		eco, pkg, err := osvKey(params[ParamEcosystem], params[ParamPackage])
		if err != nil {
			return Find{}, err
		}
		return Find{
			Table:    rel.OSVTable,
			Selector: map[string]string{"ecosystem": eco, "package": pkg},
			Fields:   OSVPackageColumns,
		}, nil
	case OSVByAdvisory:
		if !ValidRelation(rel.OSVIDView) {
			return Find{}, &Refusal{Reason: ReasonBadRelation, Detail: name}
		}
		id, err := checked(name, ParamAdvisoryID, params[ParamAdvisoryID], advisoryFormat)
		if err != nil {
			return Find{}, err
		}
		return Find{Table: rel.OSVIDView, Selector: map[string]string{"id": id}, Fields: osvAdvisoryColumns}, nil
	case KEVByCVE:
		return cveFind(name, rel.KEVView, params[ParamCVEID], kevColumns)
	case EPSSByCVE:
		return cveFind(name, rel.EPSSView, params[ParamCVEID], epssColumns)
	case MalByPackage:
		return packageFind(name, DefaultMalTable, params, malColumns)
	case ScorecardByPackage:
		return packageFind(name, DefaultScorecardTable, params, scorecardColumns)
	case CVSSByCVE:
		return cveFind(name, DefaultCVSSTable, params[ParamCVEID], cvssColumns)
	case SPDXByPackage:
		return packageFind(name, DefaultSPDXTable, params, spdxColumns)
	default:
		// ADR-0046: a template a source adapter registered carries the sql
		// form only, so it has no Mango form and the CouchDB store names it.
		return Find{}, nil
	}
}

// packageFind builds the query of a template that reads one package.
// The ecosystem and the package are held to the same spelling osvKey uses,
// so a PyPI name and a Packagist name match the sql form.
func packageFind(name, table string, params map[string]string, fields []string) (Find, error) {
	if !ValidRelation(table) {
		return Find{}, &Refusal{Reason: ReasonBadRelation, Detail: name}
	}
	eco, pkg, err := ecosystemKey(name, params[ParamEcosystem], params[ParamPackage])
	if err != nil {
		return Find{}, err
	}
	return Find{
		Table:    table,
		Selector: map[string]string{"ecosystem": eco, "package": pkg},
		Fields:   fields,
	}, nil
}

// cveFind builds the query of a template that reads one row by CVE id.
func cveFind(name, view, cve string, fields []string) (Find, error) {
	if !ValidRelation(view) {
		return Find{}, &Refusal{Reason: ReasonBadRelation, Detail: name}
	}
	id, err := checked(name, ParamCVEID, cve, cveFormat)
	if err != nil {
		return Find{}, err
	}
	return Find{Table: view, Selector: map[string]string{"cve_id": id}, Fields: fields}, nil
}

// text is the Mango form as the evidence shows it: the JSON of the query.
// Keys of the selector are sorted by the encoder, so the same request always
// gives the same text.
func (f Find) text() string {
	// Marshal cannot fail on strings and a map of strings.
	b, _ := json.Marshal(f)
	return strings.TrimSpace(string(b))
}
