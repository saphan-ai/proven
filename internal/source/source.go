// Package source is the data-source adapter (ADR-0046).
//
// A source declares how it loads, the query templates the Investigator may
// pick, and the rule that turns returned rows into finding evidence or
// nothing. OSV, KEV and EPSS are the first three. An organisation adds its
// own the same way: one adapter, registered beside them. The adapter names
// no store (ADR-0047). The configured evidence store runs whatever form the
// template carries, and a model never writes the query (ADR-0032).
//
// The built-in adapters are registered by internal/querytpl, which holds
// their query text. Importing that package fills this registry. This
// package does not import it: the catalogue reads the registry, so the
// import goes the other way.
package source

// Load is how a source is loaded into the evidence store (ADR-0046).
type Load struct {
	// Origin is where the data comes from.
	Origin string
	// Refresh is how often the data is refreshed.
	Refresh string
	// Table is the table name the source is registered as.
	Table string
	// Snapshot is the date of this load.
	// Empty means this process did not load the table. The query evidence
	// then carries the date the store stamped on the read.
	Snapshot string
}

// Template is one question a source answers.
// querytpl.Template is this type, so the catalogue and the adapter name
// one struct (ADR-0046).
type Template struct {
	// Name is the template's stable name, for example "osv-by-package".
	Name string
	// Answer is what the template returns, in one line a prompt can quote.
	Answer string
	// Params are the parameters the template takes, in the order a prompt lists them.
	Params []Param
}

// Param is one parameter of a template and the format its value must have.
type Param struct {
	// Name is the parameter's name, for example "cve_id".
	Name string
	// Format is the format a prompt shows. The builder refuses a value outside it.
	Format string
}

// Spec is one template and the builder that writes its query.
// The builder is the only place that query's text is written (ADR-0032, ADR-0046).
type Spec struct {
	// Template is the question. It is a querytpl.Template.
	Template Template
	// Build writes the query.
	// params are the template's parameter values. rel carries the review's
	// table and view names (ADR-0013). A value outside its format is refused
	// and never escaped. An organisation's adapter may ignore rel and use
	// Load.Table when the table name is the adapter's own.
	Build func(params map[string]string, rel Rel) (string, error)
}

// Rel is the table and view names a review's settings give.
// A builder places them in the query unquoted, after the same plain-name
// check the template already applies. Empty means the caller did not fill
// them; querytpl.Build fills the built-in defaults before it calls Build.
type Rel struct {
	// OSVTable is the OSV package table.
	OSVTable string
	// KEVView is the known-exploited view.
	KEVView string
	// EPSSView is the EPSS view.
	EPSSView string
	// OSVIDView is the OSV advisory-id view.
	OSVIDView string
}

// Evidence is the query and the rows a rule keeps.
// The finding's evidence is that query and those rows (ADR-0002). The Judge
// decides whether to keep the finding.
type Evidence struct {
	// SQL is the query that returned the rows.
	SQL string
	// Form is the query language that ran, for example "sql".
	Form string
	// Engine is the store that ran it, for example "local".
	Engine string
	// Snapshot is the date of the read.
	Snapshot string
	// Columns are the result's column names.
	Columns []string
	// Rows are the result's rows. Empty means the query returned nothing.
	Rows [][]string
}

// Adapter is one data source (ADR-0046).
// It declares how the source loads, the templates the Investigator may pick,
// and the rule that turns returned rows into finding evidence or nothing.
type Adapter interface {
	// ID is the source's stable name, for example "osv".
	ID() string
	// Load is where the data comes from, how often it is refreshed, the
	// table it is registered as, and the snapshot date of that load.
	Load() Load
	// Templates is this source's questions, each with the builder that
	// writes its query. A model never writes that query (ADR-0032).
	Templates() []Spec
	// Rule returns finding evidence when the rows mean a finding, and ok
	// false when they mean nothing. The evidence is the query and the rows.
	Rule(in Evidence) (Evidence, bool)
}
