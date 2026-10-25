// Package querytpl is the closed list of questions Proven can ask a store, and
// the only place the built-in query text is written (ADR-0032).
//
// The list is the source-adapter registry (ADR-0046). OSV, KEV and EPSS
// register their templates from this package, and Catalogue and Build read
// that registry, so each template is written once. A model never writes a
// query. The Investigator and the Skeptic name one template and give the
// values of its parameters; Proven checks each value against the strict
// format of that parameter and builds the query itself. A value outside its
// format is refused with a named reason, never escaped, so a quote, a
// backslash, a control byte or invalid UTF-8 cannot reach the text. Each
// template has one form per query language (ADR-0047): sql for Zetaris and
// the local store, and a mango selector for CouchDB. The OSV
// check and the EPSS check build their queries from the same registry, so
// every query Proven sends has one author. docs/QUERIES.md lists the templates.
package querytpl

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/pep503"
	"github.com/saphan-ai/proven/internal/source"
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
	ReasonBuilder         = "query template builder refused the values"
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
// It is the source adapter's parameter type, so a template is not declared twice.
type Param = source.Param

// Template is one question Proven can ask: its name, what it answers, and the
// parameters it takes. Catalogue returns them all. It is the source adapter's
// template type (ADR-0046).
type Template = source.Template

// Catalogue lists every template, in the registry's order. The prompts and
// docs/QUERIES.md are written from it, and a test holds the document to it.
// The list is the source adapters, so a template is not written here again
// (ADR-0046).
func Catalogue() []Template {
	var out []Template
	for _, a := range source.All() {
		for _, spec := range a.Templates() {
			out = append(out, spec.Template)
		}
	}
	return out
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
	ecosystemFormat = regexp.MustCompile(`^(Go|npm|PyPI|Packagist|crates\.io)$`)
	// anyPackageFormat admits every ecosystem's package spelling and nothing
	// more. It is the gate Proven applies to the parameter; the builder then
	// holds the name to its own ecosystem's format.
	anyPackageFormat = regexp.MustCompile(`^@?[A-Za-z0-9][A-Za-z0-9._~/-]{0,299}$`)
	cveFormat        = regexp.MustCompile(`^CVE-[0-9]{4}-[0-9]{4,12}$`)
	advisoryFormat   = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,15}(-[A-Za-z0-9]{1,40}){1,6}$`)
	goFormat         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~/-]{0,299}$`)
	npmFormat        = regexp.MustCompile(`^(@[A-Za-z0-9][A-Za-z0-9._~-]{0,213}/)?[A-Za-z0-9][A-Za-z0-9._~-]{0,213}$`)
	pypiFormat       = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]{0,212}[A-Za-z0-9])?$`)
	packagistFormat  = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]{0,212}[a-z0-9])?/[a-z0-9]([a-z0-9._-]{0,212}[a-z0-9])?$`)
	// A crate name is 1 to 64 characters, starts with an ASCII letter, and
	// otherwise contains ASCII letters, digits, '_' or '-'. The spelling is
	// kept: crates.io does not fold case or turn '_' into '-'.
	crateFormat = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
)

// ValidCVE reports whether s is a CVE id and nothing else.
func ValidCVE(s string) bool { return cveFormat.MatchString(s) }

// ValidRelation reports whether name can be placed in a FROM clause: one
// identifier, or schema.name, of letters, digits and underscores. The name
// goes into the SQL unquoted, so this check is what keeps a setting from
// carrying SQL of its own.
func ValidRelation(name string) bool { return source.ValidRelation(name) }

// checked is the one gate every value passes, whichever query language it
// goes into. It does not escape: a value outside its format is refused, so
// what it returns is exactly the value and holds no quote, no backslash and
// no control byte. The format has already been checked by the caller; the
// last test here refuses any byte an allow-list should never have admitted,
// so a format edited carelessly cannot let one through (ADR-0032 point 3).
// ADR-0047: the sql form and the mango form both take their values from here,
// so the two forms cannot disagree on what a value may be.
func checked(template, param, value string, format *regexp.Regexp) (string, error) {
	if !format.MatchString(value) {
		return "", &Refusal{Reason: ReasonBadValue, Detail: template + " " + param}
	}
	for i := 0; i < len(value); i++ {
		if c := value[i]; c < 0x20 || c > 0x7e || c == '\'' || c == '\\' {
			return "", &Refusal{Reason: ReasonBadValue, Detail: template + " " + param}
		}
	}
	return value, nil
}

// literal is the one builder of a SQL string literal: the checked value
// between two quotes.
func literal(template, param, value string, format *regexp.Regexp) (string, error) {
	v, err := checked(template, param, value, format)
	if err != nil {
		return "", err
	}
	return "'" + v + "'", nil
}

// packageFormat is the format of a package name for one ecosystem.
func packageFormat(ecosystem string) *regexp.Regexp {
	switch ecosystem {
	case "Go":
		return goFormat
	case "npm":
		return npmFormat
	case "Packagist":
		return packagistFormat
	case "crates.io":
		return crateFormat
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
	_, spec, ok := source.Lookup(name)
	if !ok {
		// The name is not repeated: it is model text and may be SQL.
		return "", &Refusal{Reason: ReasonUnknownTemplate}
	}
	if err := checkParams(spec.Template, req.Params); err != nil {
		return "", err
	}
	// ADR-0046: the adapter that owns the template builds it. The text stays
	// in this package; the registry is what names the builder, so the switch
	// that used to live here is not a second copy.
	if spec.Build == nil {
		return "", fmt.Errorf("querytpl: template %s has no builder", spec.Template.Name)
	}
	// Proven holds every value to its parameter's format and to the byte
	// floor before any builder runs, built-in or registered: the adapter's
	// builder is never the only guard (ADR-0032 points 1 and 3).
	for _, p := range spec.Template.Params {
		if !source.CheckValue(p, req.Params[p.Name]) {
			return "", &Refusal{Reason: ReasonBadValue, Detail: spec.Template.Name + " " + p.Name}
		}
	}
	sql, err := spec.Build(req.Params, source.Rel{
		OSVTable:  rel.OSVTable,
		KEVView:   rel.KEVView,
		EPSSView:  rel.EPSSView,
		OSVIDView: rel.OSVIDView,
	})
	if err != nil {
		// A registered builder returns what it likes. Whatever it returns
		// reaches the Investigator as a refusal, so one refused value is a
		// query note and never a failed review. The builder's text is not
		// repeated: it may echo the value.
		var refusal *Refusal
		if errors.As(err, &refusal) {
			return "", err
		}
		return "", &Refusal{Reason: ReasonBuilder, Detail: spec.Template.Name}
	}
	return sql, nil
}

// Validate checks a request without building a query: the Skeptic uses it to
// refuse a bad answer before the pipeline holds the review's settings.
func Validate(req Request) error {
	_, err := Build(req, Relations{})
	return err
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

// osvByAdvisory builds the osv-by-advisory query.
func osvByAdvisory(view, id string) (string, error) {
	if !ValidRelation(view) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: OSVByAdvisory}
	}
	lit, err := literal(OSVByAdvisory, ParamAdvisoryID, id, advisoryFormat)
	if err != nil {
		return "", err
	}
	return "SELECT id, alias, published, severity FROM " + view + " WHERE id = " + lit, nil
}

// OSVPackageSQL is the OSV package query for one ecosystem and package. A PyPI
// name is put in its PEP 503 form first, because the table is keyed by it.
// A Packagist name is put in lower case first, because Packagist names are
// not case-sensitive and the table is keyed by the lower-case form.
// A crates.io name is kept as written.
// The dependency check calls it with names from the change's manifests.
func OSVPackageSQL(table, ecosystem, pkg string) (string, error) {
	return osvByPackage(table, ecosystem, pkg)
}

// osvKey checks an ecosystem and a package name and returns the two values a
// query on the OSV table matches, the name in the spelling the table is keyed
// by. The sql form and the mango form both call it, so they match the same
// rows (ADR-0047).
func osvKey(ecosystem, pkg string) (string, string, error) {
	switch ecosystem {
	case "Go", "npm", "PyPI", "Packagist", "crates.io":
	default:
		return "", "", &Refusal{Reason: ReasonBadValue, Detail: OSVByPackage + " " + ParamEcosystem}
	}
	if ecosystem == "PyPI" {
		pkg = pep503.Normalise(pkg)
	}
	// Lower case is the spelling Packagist publishes. The reader already
	// stores it that way; this is the same fold for a name a model typed.
	if ecosystem == "Packagist" {
		pkg = strings.ToLower(pkg)
	}
	name, err := checked(OSVByPackage, ParamPackage, pkg, packageFormat(ecosystem))
	if err != nil {
		return "", "", err
	}
	return ecosystem, name, nil
}

// osvByPackage builds the osv-by-package query.
func osvByPackage(table, ecosystem, pkg string) (string, error) {
	if !ValidRelation(table) {
		return "", &Refusal{Reason: ReasonBadRelation, Detail: OSVByPackage}
	}
	eco, name, err := osvKey(ecosystem, pkg)
	if err != nil {
		return "", err
	}
	return "SELECT " + strings.Join(OSVPackageColumns, ", ") + " FROM " + table + " WHERE ecosystem = '" + eco + "' AND package = '" + name + "'", nil
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
