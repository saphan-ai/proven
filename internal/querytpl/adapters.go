package querytpl

import "github.com/saphan-ai/proven/internal/source"

// builtinErr is set if the built-in adapters could not be registered.
// A review's catalogue would then be empty, and the tests name this error.
var builtinErr error

func init() {
	// ADR-0046: the registry lives in source, and the query text lives here,
	// next to the literal check. source cannot import this package, because
	// Catalogue reads the registry. The registration therefore happens here.
	builtinErr = source.SetBuiltins([]source.Adapter{
		osvAdapter{},
		kevAdapter{},
		epssAdapter{},
		malAdapter{},
		scorecardAdapter{},
		cvssAdapter{},
		spdxAdapter{},
	})
}

// osvAdapter is the OSV source (ADR-0046). It owns both OSV templates.
// The package table is the load's table. The advisory template reads the
// id view, which is a relation name, not a second source.
type osvAdapter struct{}

func (osvAdapter) ID() string { return "osv" }

func (osvAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://osv.dev",
		Refresh: "an administrator replaces the CSV; a review reads the table as it stands",
		Table:   DefaultOSVTable,
	}
}

func (osvAdapter) Templates() []source.Spec {
	return []source.Spec{
		{
			Template: source.Template{
				Name:   OSVByPackage,
				Answer: "the OSV advisory rows of one package",
				Params: []source.Param{
					{Name: ParamEcosystem, Format: "one of Go, npm, PyPI, Packagist, crates.io", Match: ecosystemFormat},
					{Name: ParamPackage, Format: "a package name in the ecosystem's own spelling", Match: anyPackageFormat},
				},
			},
			Build: func(params map[string]string, rel source.Rel) (string, error) {
				return OSVPackageSQL(rel.OSVTable, params[ParamEcosystem], params[ParamPackage])
			},
		},
		{
			Template: source.Template{
				Name:   OSVByAdvisory,
				Answer: "the OSV advisory with one advisory id",
				Params: []source.Param{
					{Name: ParamAdvisoryID, Format: "an advisory id such as GHSA-xxxx-xxxx-xxxx or GO-2021-0113", Match: advisoryFormat},
				},
			},
			Build: func(params map[string]string, rel source.Rel) (string, error) {
				return osvByAdvisory(rel.OSVIDView, params[ParamAdvisoryID])
			},
		},
	}
}

func (osvAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// No row means the table has no advisory for that package. That is
	// nothing to attach, not a statement that the package is safe.
	// Rows are the advisory ranges. They show that the table named the
	// package. They are not yet a finding: the dependency check compares
	// the change's version, and a version outside every range is a trace
	// line. This rule does not see the version, so that check does not
	// use the rule as its finding decision (ADR-0046).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	return copyEvidence(in), true
}

// kevAdapter is the CISA known-exploited source (ADR-0046).
type kevAdapter struct{}

func (kevAdapter) ID() string { return "kev" }

func (kevAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://www.cisa.gov/known-exploited-vulnerabilities-catalog",
		Refresh: "an administrator refreshes the view; Proven does not fetch it",
		Table:   DefaultKEVView,
	}
}

func (kevAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   KEVByCVE,
			Answer: "the CISA known-exploited entry of one CVE",
			Params: []source.Param{
				{Name: ParamCVEID, Format: "a CVE id such as CVE-2021-44228", Match: cveFormat},
			},
		},
		Build: func(params map[string]string, rel source.Rel) (string, error) {
			return KEVSQL(rel.KEVView, params[ParamCVEID])
		},
	}}
}

func (kevAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// A row means the CVE is on the known-exploited list. No row means it
	// is not on the list, and a missing row is not evidence (ADR-0046).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	return copyEvidence(in), true
}

// epssAdapter is the FIRST EPSS source (ADR-0046).
type epssAdapter struct{}

func (epssAdapter) ID() string { return "epss" }

func (epssAdapter) Load() source.Load {
	return source.Load{
		Origin:  "https://www.first.org/epss/",
		Refresh: "FIRST publishes a daily score; the view is the administrator's copy",
		Table:   DefaultEPSSView,
	}
}

func (epssAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   EPSSByCVE,
			Answer: "the EPSS score of one CVE",
			Params: []source.Param{
				{Name: ParamCVEID, Format: "a CVE id such as CVE-2021-44228", Match: cveFormat},
			},
		},
		Build: func(params map[string]string, rel source.Rel) (string, error) {
			return EPSSSQL(rel.EPSSView, params[ParamCVEID])
		},
	}}
}

func (epssAdapter) Rule(in source.Evidence) (source.Evidence, bool) {
	// A row is a score, and the score is evidence. No row means the view
	// has no score. A missing score says nothing about the CVE, so it is
	// not evidence (ADR-0046).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	return copyEvidence(in), true
}

// copyEvidence copies the rows so a rule's result does not share memory
// with the store's answer.
func copyEvidence(in source.Evidence) source.Evidence {
	out := in
	out.Columns = append([]string(nil), in.Columns...)
	if in.Rows == nil {
		out.Rows = [][]string{}
		return out
	}
	out.Rows = make([][]string, len(in.Rows))
	for i, row := range in.Rows {
		out.Rows[i] = append([]string(nil), row...)
	}
	return out
}
