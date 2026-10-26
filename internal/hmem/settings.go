package hmem

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// Environment variables of H-MEM (ADR-0013). The set is closed: LoadLimits
// reads these names and no others. An empty value means the default from
// DefaultLimits, which is the one list of them.
const (
	EnvMaxStateBytes = "PROVEN_HMEM_MAX_STATE_BYTES"
	EnvQueryTokens   = "PROVEN_HMEM_QUERY_TOKENS"
)

// DefaultMaxStateBytes is the default cap on state.json: 64 MiB. H-MEM reads
// the whole file at open and rewrites it on every write, so an unbounded file
// would be an unbounded read and an unbounded rewrite. The cap is far above
// what the conformance corpus and a review's refutations need (a few
// kilobytes), and far below a size at which a whole-file read is a hazard.
const DefaultMaxStateBytes int64 = 64 << 20

// DefaultQueryTokens is how many query tokens count in a Query (ADR-0013).
const DefaultQueryTokens = 8

// maxSettingBytes is the largest cap a setting may name: 1 TiB, so the
// reader's limit+1 never overflows. A larger value is refused, not clipped.
const maxSettingBytes int64 = 1 << 40

// maxQueryTokens is the largest token limit a setting may name. It keeps the
// per-record scoring cost of one query bounded by a number a person can read.
const maxQueryTokens = 1024

// ErrStateTooLarge is the refusal of a state file above the cap. Open wraps it
// with the cap and, when the file reports its size, that size. A file that
// reports a size over the cap is not read at all; a source that under-reports
// its size is read up to the cap plus one byte, and the refusal names the cap.
var ErrStateTooLarge = errors.New("hmem: state file is over the size cap")

// Limits are the two bounds H-MEM runs under. A zero field means the default,
// so a partly filled Limits is never an unbounded one.
type Limits struct {
	// MaxStateBytes is the largest state.json Open will read.
	MaxStateBytes int64
	// QueryTokens is how many leading query tokens count in a Query.
	QueryTokens int
}

// DefaultLimits is what H-MEM uses when no setting names a bound.
func DefaultLimits() Limits {
	return Limits{MaxStateBytes: DefaultMaxStateBytes, QueryTokens: DefaultQueryTokens}
}

// withDefaults fills each zero or negative field with its default.
func (l Limits) withDefaults() Limits {
	d := DefaultLimits()
	if l.MaxStateBytes <= 0 {
		l.MaxStateBytes = d.MaxStateBytes
	}
	if l.QueryTokens <= 0 {
		l.QueryTokens = d.QueryTokens
	}
	return l
}

// ConfigError names a refused H-MEM setting and the values that are allowed.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the setting, the refused value, why, and what is allowed.
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := fmt.Sprintf("configuration refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// LoadLimits reads the two bounds from env. An empty value is the default. A
// value that is set and is not a positive number in range is refused and is
// never replaced by the default (ADR-0013).
func LoadLimits(env func(string) string) (Limits, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	l := DefaultLimits()
	if raw := strings.TrimSpace(env(EnvMaxStateBytes)); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 || n > maxSettingBytes {
			return Limits{}, &ConfigError{
				Field: EnvMaxStateBytes, Value: clip(raw), Reason: "not a positive number of bytes in range",
				Allowed: []string{"a positive integer up to " + strconv.FormatInt(maxSettingBytes, 10) + "; empty means the default"},
			}
		}
		l.MaxStateBytes = n
	}
	if raw := strings.TrimSpace(env(EnvQueryTokens)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 || n > maxQueryTokens {
			return Limits{}, &ConfigError{
				Field: EnvQueryTokens, Value: clip(raw), Reason: "not a positive number of tokens in range",
				Allowed: []string{"a positive integer up to " + strconv.Itoa(maxQueryTokens) + "; empty means the default"},
			}
		}
		l.QueryTokens = n
	}
	return l, nil
}

// clip keeps a refused value short, so a long one is not echoed whole.
func clip(s string) string {
	const max = 40
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Environment lists the H-MEM settings for the configuration page.
func Environment() []envdoc.Setting {
	d := DefaultLimits()
	return []envdoc.Setting{
		{Name: EnvMaxStateBytes, Default: envdoc.Tick(strconv.FormatInt(d.MaxStateBytes, 10)), Description: "the largest H-MEM `state.json` that opens; above it the store refuses by name and reads nothing more (`ErrStateTooLarge`). A positive number of bytes, at most 1 TiB (1099511627776); a larger value is refused. Sample mode uses the default and does not read this variable"},
		{Name: EnvQueryTokens, Default: envdoc.Tick(strconv.Itoa(d.QueryTokens)), Description: "how many leading query tokens count when H-MEM scores a query; a positive integer up to 1024. Sample mode uses the default and does not read this variable"},
	}
}
