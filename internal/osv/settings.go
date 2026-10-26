package osv

import (
	"strings"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// EnvTable is the environment variable that names the OSV package table.
// Sample mode does not read it.
const EnvTable = "PROVEN_OSV_TABLE"

// EnvKEVView is the environment variable that names the KEV view.
// Sample mode does not read it.
const EnvKEVView = "PROVEN_KEV_VIEW"

// EnvIDView is the environment variable that names the OSV advisory-id view,
// the one the osv-by-advisory template reads. Sample mode does not read it.
const EnvIDView = "PROVEN_OSV_ID_VIEW"

// IDView is the advisory-id view a review queries when EnvIDView is empty.
const IDView = querytpl.DefaultOSVIDView

// DefaultTable is the table a review queries when EnvTable is empty.
const DefaultTable = querytpl.DefaultOSVTable

// ConfigError names a refused OSV setting and the values that are allowed.
// The sentence is one line: which setting was refused, why, and what is
// allowed (ADR-0013, point 4a).
type ConfigError = service.ConfigError

// LoadTable reads EnvTable. An empty value is DefaultTable.
// ADR-0013: the table name is a setting with a documented default. A value
// that is set and wrong refuses the run; it never falls back to the default.
func LoadTable(env func(string) string) (string, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvTable))
	if raw == "" {
		return DefaultTable, nil
	}
	if !ValidTable(raw) {
		return "", &ConfigError{
			Field: EnvTable, Value: raw, Reason: "not a table name",
			Allowed: []string{"a table or schema.table of letters, digits and underscores; empty means " + DefaultTable},
		}
	}
	return raw, nil
}

// ValidTable reports whether name can be placed in a FROM clause.
// One identifier, or schema.table. Nothing else.
// The name goes into the SQL unquoted, so this check is what keeps a setting
// from carrying SQL of its own. ADR-0032: the one rule lives in querytpl,
// which builds the query.
func ValidTable(name string) bool {
	return querytpl.ValidRelation(name)
}

// Environment lists the OSV settings for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{Name: EnvTable, Default: envdoc.Tick(DefaultTable), Description: "the OSV package table the dependency check queries; see [OSV_TABLE.md](OSV_TABLE.md). Sample mode uses the default and does not read this variable"},
		{Name: EnvKEVView, Default: envdoc.Tick(KEVView), Description: "the KEV view an advisory's CVE alias is looked up in. Sample mode uses the default and does not read this variable"},
		{Name: EnvIDView, Default: envdoc.Tick(IDView), Description: "the OSV advisory-id view the `osv-by-advisory` template reads; see [ZETARIS_USL.md](ZETARIS_USL.md). Sample mode uses the default and does not read this variable"},
	}
}

// LoadIDView reads EnvIDView. An empty value is IDView.
// ADR-0013: a value that is set and wrong refuses the run; it never falls
// back to the default.
func LoadIDView(env func(string) string) (string, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvIDView))
	if raw == "" {
		return IDView, nil
	}
	if !ValidView(raw) {
		return "", &ConfigError{
			Field: EnvIDView, Value: raw, Reason: "not a view name",
			Allowed: []string{"a view or schema.view of letters, digits and underscores; empty means " + IDView},
		}
	}
	return raw, nil
}

// LoadView reads EnvKEVView. An empty value is KEVView.
// ADR-0013: the view name is a setting with a documented default. A value
// that is set and wrong refuses the run; it never falls back to the default.
func LoadView(env func(string) string) (string, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvKEVView))
	if raw == "" {
		return KEVView, nil
	}
	if !ValidView(raw) {
		return "", &ConfigError{
			Field: EnvKEVView, Value: raw, Reason: "not a view name",
			Allowed: []string{"a view or schema.view of letters, digits and underscores; empty means " + KEVView},
		}
	}
	return raw, nil
}

// ValidView reports whether name can be placed in a FROM clause.
// One identifier, or schema.view. Nothing else.
// The view is pasted into the KEV query the same way the table name is
// pasted into the package query, so it has to pass the same name check.
func ValidView(name string) bool {
	return ValidTable(name)
}
