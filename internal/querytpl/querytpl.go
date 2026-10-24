// Package querytpl is the closed list of questions Proven can ask Zetaris, and
// the only place their SQL is written (ADR-0032).
//
// A model never writes a query. The Investigator and the Skeptic name one
// template and give the values of its parameters; Proven checks each value
// against the strict format of that parameter and builds the SQL itself. A
// value outside its format is refused with a named reason, never escaped, so
// a quote, a backslash, a control byte or invalid UTF-8 cannot reach the SQL.
// The OSV check and the EPSS check build their queries here too, so every
// query Proven sends has one author. docs/QUERIES.md lists the templates.
package querytpl

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/pep503"
)

// The defaults of the three settings that name a table or a view (ADR-0013).
// osv, epss and the review settings refer to these, so a name is written once.
const (
	DefaultOSVTable  = "px_osvfile.px_osv_pkg"
	DefaultKEVView   = "px_evidence.px_kev_vulns"
	DefaultEPSSView  = "px_evidence.px_epss_cve_rows"
	DefaultOSVIDView = "px_evidence.px_osv_id"
)

// The names of the templates. The list is closed: a name not here is refused.
const (
	OSVByPackage  = "osv-by-package"
	OSVByAdvisory = "osv-by-advisory"
	KEVByCVE      = "kev-by-cve"
	EPSSByCVE     = "epss-by-cve"
)

// The names of the parameters.
const (
	ParamEcosystem  = "ecosystem"
	ParamPackage    = "package"
	ParamCVEID      = "cve_id"
	ParamAdvisoryID = "advisory_id"
)

// OSVPackageColumns are the columns the OSV package query selects and the
// dependency check reads by name. docs/OSV_TABLE.md says what each holds.
var OSVPackageColumns = []string{
	"id", "aliases", "ecosystem", "package", "introduced", "fixed", "last_affected", "severity", "published",
}

// The refusal reasons. A refusal names one of them first, so a reader of the
// trace sees why a finding has no query evidence.
const (
	ReasonUnknownTemplate = "query template is not on the list"
	ReasonMissingParam    = "query template parameter is missing"
	ReasonUnknownParam    = "query template parameter is not one the template takes"
	ReasonBadValue        = "query parameter value is outside its format"
	ReasonBadRelation     = "table or view name is not a plain name"
	ReasonModelSQL        = "a model's answer carries SQL text"
)

// Refusal is a template or a value Proven will not build a query from.
type Refusal struct {
	Reason string
	// Detail names the template and the parameter. It never repeats the
	// offending value, which may hold bytes a terminal or a log would act on.
	Detail string
}

// Error reads "<reason>: <detail>".
func (r *Refusal) Error() string {
	if r.Detail == "" {
		return r.Reason
	}
	return r.Reason + ": " + r.Detail
}

// Request is what a model gives: a template name and the values of its
// parameters. It is the JSON the Investigator and the Skeptic ask for.
type Request struct {
	Template string            `json:"template"`
	Params   map[string]string `json:"params"`
}

// Empty reports whether the request names no template, which means the model
// proposed no query.
func (r Request) Empty() bool {
	return strings.TrimSpace(r.Template) == "" && len(r.Params) == 0
}

// Relations are the table and view names a review's settings give (ADR-0013).
// A zero field means the default. Build checks each name again, because it is
// the one thing that cannot be a quoted literal.
type Relations struct {
	OSVTable  string
	KEVView   string
	EPSSView  string
	OSVIDView string
}

// resolved fills each empty name with its default.
func (r Relations) resolved() Relations {
	fill := func(v *string, def string) {
		if *v == "" {
			*v = def
		}
	}
	fill(&r.OSVTable, DefaultOSVTable)
	fill(&r.KEVView, DefaultKEVView)
	fill(&r.EPSSView, DefaultEPSSView)
	fill(&r.OSVIDView, DefaultOSVIDView)
	return r
}

// Param is one parameter of a template and the format its value must have.
type Param struct {
	Name   string
	Format string
}

// Template is one question Proven can ask: its name, what it answers, and the
// parameters it takes. Catalogue returns them all.
type Template struct {
	Name   string
	Answer string
	Params []Param
}

// Catalogue lists every template, in a fixed order. The prompts and
// docs/QUERIES.md are written from it, and a test holds the document to it.
func Catalogue() []Template {
	return []Template{
		{
			Name:   OSVByPackage,
			Answer: "the OSV advisory rows of one package",
			Params: []Param{
				{ParamEcosystem, "one of Go, npm, PyPI"},
				{ParamPackage, "a package name in the ecosystem's own spelling"},
			},
		},
		{
			Name:   OSVByAdvisory,
			Answer: "the OSV advisory with one advisory id",
			Params: []Param{{ParamAdvisoryID, "an advisory id such as GHSA-xxxx-xxxx-xxxx or GO-2021-0113"}},
		},
		{
			Name:   KEVByCVE,
			Answer: "the CISA known-exploited entry of one CVE",
			Params: []Param{{ParamCVEID, "a CVE id such as CVE-2021-44228"}},
		},
		{
			Name:   EPSSByCVE,
			Answer: "the EPSS score of one CVE",
			Params: []Param{{ParamCVEID, "a CVE id such as CVE-2021-44228"}},
		},
	}
}

// Describe is the list of templates as a prompt gives it to a model: one line
// each, with its parameters. It is written from Catalogue, so what a model is
// told and what Build accepts cannot part.
func Describe() string {
	var b strings.Builder
	for i, t := range Catalogue() {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(t.Name + " (")
		for j, p := range t.Params {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p.Name + ": " + p.Format)
		}
		b.WriteString(") gives " + t.Answer + ".")
	}
	return b.String()
}

// The strict formats. Each is an allow-list of ASCII, so none admits a quote,
// a backslash, a control byte or a byte that is not UTF-8. They are anchored:
// in RE2 the dollar sign matches the end of the text only, so a trailing line
// break is not admitted.
var (
	cveFormat      = regexp.MustCompile(`^CVE-[0-9]{4}-[0-9]{4,12}$`)
	advisoryFormat = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,15}(-[A-Za-z0-9]{1,40}){1,6}$`)
	goFormat       = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~/-]{0,299}$`)
	npmFormat      = regexp.MustCompile(`^(@[A-Za-z0-9][A-Za-z0-9._~-]{0,213}/)?[A-Za-z0-9][A-Za-z0-9._~-]{0,213}$`)
	pypiFormat     = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]{0,212}[A-Za-z0-9])?$`)
)

// ValidCVE reports whether s is a CVE id and nothing else.
func ValidCVE(s string) bool { return cveFormat.MatchString(s) }

// ValidRelation reports whether name can be placed in a FROM clause: one
// identifier, or schema.name, of letters, digits and underscores. The name
// goes into the SQL unquoted, so this check is what keeps a setting from
// carrying SQL of its own.
func ValidRelation(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if !identifier(p) {
			return false
		}
	}
	return true
}

// identifier reports whether s is a letter or underscore followed by letters,
// digits and underscores.
func identifier(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c == '_':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

// literal is the one builder of a SQL string literal. It does not escape: a
// value outside its format is refused, so the text between the two quotes is
// exactly the value and holds no quote, no backslash and no control byte.
// The format has already been checked by the caller; the last test here
// refuses any byte an allow-list should never have admitted, so a format
// edited carelessly cannot let one through (ADR-0032 point 3).
func literal(template, param, value string, format *regexp.Regexp) (string, error) {
	if !format.MatchString(value) {
		return "", &Refusal{Reason: ReasonBadValue, Detail: template + " " + param}
	}
	for i := 0; i < len(value); i++ {
		if c := value[i]; c < 0x20 || c > 0x7e || c == '\'' || c == '\\' {
			return "", &Refusal{Reason: ReasonBadValue, Detail: template + " " + param}
		}
	}
	return "'" + value + "'", nil
}

// packageFormat is the format of a package name for one ecosystem.
func packageFormat(ecosystem string) *regexp.Regexp {
	switch ecosystem {
	case "Go":
		return goFormat
	case "npm":
		return npmFormat
	default:
		return pypiFormat
	}
}

// Build checks the request and returns the SQL Proven built from it. A name
// that is not on the list, a missing or unexpected parameter, a value outside
// its format and a table or view name that is not plain are each a *Refusal.
// The model's text is never part of the SQL beyond the validated values.
func Build(req Request, rel Relations) (string, error) {
	rel = rel.resolved()
	name := strings.TrimSpace(req.Template)
	tpl, ok := lookup(name)
	if !ok {
		// The name is not repeated: it is model text and may be SQL.
		return "", &Refusal{Reason: ReasonUnknownTemplate}
	}
	if err := checkParams(tpl, req.Params); err != nil {
		return "", err
	}
	switch name {
	case OSVByPackage:
		return osvByPackage(rel.OSVTable, req.Params[ParamEcosystem], req.Params[ParamPackage])
	case OSVByAdvisory:
		if !ValidRelation(rel.OSVIDView) {
			return "", &Refusal{Reason: ReasonBadRelation, Detail: name}
		}
		id, err := literal(name, ParamAdvisoryID, req.Params[ParamAdvisoryID], advisoryFormat)
		if err != nil {
			return "", err
		}
		return "SELECT id, alias, published, severity FROM " + rel.OSVIDView + " WHERE id = " + id, nil
	case KEVByCVE:
		return KEVSQL(rel.KEVView, req.Params[ParamCVEID])
	default:
		return EPSSSQL(rel.EPSSView, req.Params[ParamCVEID])
	}
}

// Validate checks a request without building a query: the Skeptic uses it to
// refuse a bad answer before the pipeline holds the review's settings.
func Validate(req Request) error {
	_, err := Build(req, Relations{})
	return err
}

// lookup finds a template by name.
func lookup(name string) (Template, bool) {
	for _, t := range Catalogue() {
		if t.Name == name {
			return t, true
		}
	}
	return Template{}, false
}

// checkParams holds the request to exactly the template's parameters: each one
// present, and no other.
func checkParams(tpl Template, params map[string]string) error {
	want := map[string]bool{}
	for _, p := range tpl.Params {
		want[p.Name] = true
		if _, ok := params[p.Name]; !ok {
			return &Refusal{Reason: ReasonMissingParam, Detail: tpl.Name + " " + p.Name}
		}
	}
	var extra []string
	for k := range params {
		if !want[k] {
			extra = append(extra, k)
		}
	}
	if len(extra) > 0 {
		// The key is model text; only how many are said, never what they are.
		sort.Strings(extra)
		return &Refusal{Reason: ReasonUnknownParam, Detail: fmt.Sprintf("%s takes %d parameters and %d others were given", tpl.Name, len(tpl.Params), len(extra))}
	}
	return nil
}

// OSVPackageSQL is the OSV package query for one ecosystem and package. A PyPI
// name is put in its PEP 503 form first, because the table is keyed by it.
// The dependency check calls it with names from the change's manifests.
func OSVPackageSQL(table, ecosystem, pkg string) (string, error) {
	return osvByPackage(table, ecosystem, pkg)
}

// osvByPackage builds the osv-by-package query.
func osvByPackage(table, ecosystem, pkg string) (string, error) {
	if !ValidRelation(table) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: OSVByPackage}
	}
	switch ecosystem {
	case "Go", "npm", "PyPI":
	default:
		return "", &Refusal{Reason: ReasonBadValue, Detail: OSVByPackage + " " + ParamEcosystem}
	}
	if ecosystem == "PyPI" {
		pkg = pep503.Normalise(pkg)
	}
	name, err := literal(OSVByPackage, ParamPackage, pkg, packageFormat(ecosystem))
	if err != nil {
		return "", err
	}
	return "SELECT " + strings.Join(OSVPackageColumns, ", ") + " FROM " + table + " WHERE ecosystem = '" + ecosystem + "' AND package = " + name, nil
}

// KEVSQL is the known-exploited query for one CVE id.
func KEVSQL(view, cve string) (string, error) {
	if !ValidRelation(view) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: KEVByCVE}
	}
	id, err := literal(KEVByCVE, ParamCVEID, cve, cveFormat)
	if err != nil {
		return "", err
	}
	return "SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM " + view + " WHERE cve_id = " + id, nil
}

// EPSSSQL is the EPSS score query for one CVE id.
func EPSSSQL(view, cve string) (string, error) {
	if !ValidRelation(view) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: EPSSByCVE}
	}
	id, err := literal(EPSSByCVE, ParamCVEID, cve, cveFormat)
	if err != nil {
		return "", err
	}
	return "SELECT cve_id, epss, percentile, score_date FROM " + view + " WHERE cve_id = " + id, nil
}
