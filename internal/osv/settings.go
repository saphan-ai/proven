package osv

import (
	"strings"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// EnvTable is the environment variable that names the OSV package table.
// Sample mode does not read it.
const EnvTable = "PROVEN_OSV_TABLE"

// EnvKEVView is the environment variable that names the KEV view.
// Sample mode does not read it.
const EnvKEVView = "PROVEN_KEV_VIEW"

// DefaultTable is the table a review queries when EnvTable is empty.
const DefaultTable = "px_osvfile.px_osv_pkg"

// ConfigError names a refused OSV setting and the values that are allowed.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error says which setting was refused, why, and what is allowed, on one
// line (ADR-0013, point 4a).
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := "configuration refused: " + e.Field + "=" + quote(e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// quote shows the value in double quotes, so an empty or padded value is visible.
func quote(s string) string {
	return `"` + s + `"`
}

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
// from carrying SQL of its own.
func ValidTable(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) < 1 || len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if !ident(p) {
			return false
		}
	}
	return true
}

// ident reports whether s is one SQL identifier: a letter or underscore,
// then letters, digits and underscores.
func ident(s string) bool {
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

// Environment lists the OSV settings for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{Name: EnvTable, Default: envdoc.Tick(DefaultTable), Description: "the OSV package table the dependency check queries; see [OSV_TABLE.md](OSV_TABLE.md). Sample mode uses the default and does not read this variable"},
		{Name: EnvKEVView, Default: envdoc.Tick(KEVView), Description: "the KEV view an advisory's CVE alias is looked up in. Sample mode uses the default and does not read this variable"},
	}
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
