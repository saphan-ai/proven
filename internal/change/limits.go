package change

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// Environment variables that bound a read of repository content (ADR-0013).
// The set is closed: LoadLimits reads these names and no others. An empty
// value means the default from DefaultLimits, which is the one list of them.
const (
	EnvMaxArchiveBytes = "PROVEN_MAX_ARCHIVE_BYTES"
	EnvMaxFileBytes    = "PROVEN_MAX_FILE_BYTES"
	EnvMaxConfigBytes  = "PROVEN_MAX_CONFIG_BYTES"
	EnvMaxRecordBytes  = "PROVEN_MAX_RECORD_BYTES"
)

// Limits is one size bound per kind of read of repository content. The
// repository under review is not trusted to be small, so no read of it is
// without a bound (ADR-0013, ADR-0025). A value is a number of bytes.
type Limits struct {
	// Archive bounds everything one export reads: the git archive stream, or
	// the sum of the tracked files copied from a working tree.
	Archive int64
	// File bounds one file read from a tree, one git blob and one commit log.
	File int64
	// Config bounds proven.yaml, whether read from the target or a commit.
	Config int64
	// Record bounds the author record, summed over the files its setting names.
	Record int64
}

// DefaultLimits is what Proven uses when no setting names a bound.
func DefaultLimits() Limits {
	return Limits{Archive: 128 << 20, File: 8 << 20, Config: 1 << 20, Record: 1 << 20}
}

// limitsInUse is the bounds every reader of this package applies. Nil means
// DefaultLimits. A review sets it once, when its settings are loaded.
var limitsInUse atomic.Pointer[Limits]

// Bounds returns the limits in use now.
func Bounds() Limits {
	if l := limitsInUse.Load(); l != nil {
		return *l
	}
	return DefaultLimits()
}

// UseLimits makes l the limits every reader of this package applies and returns
// the ones it replaced, so a caller can restore them. A field that is zero or
// below keeps its default, so a partly filled Limits is never an unbounded one.
func UseLimits(l Limits) (previous Limits) {
	previous = Bounds()
	d := DefaultLimits()
	if l.Archive <= 0 {
		l.Archive = d.Archive
	}
	if l.File <= 0 {
		l.File = d.File
	}
	if l.Config <= 0 {
		l.Config = d.Config
	}
	if l.Record <= 0 {
		l.Record = d.Record
	}
	limitsInUse.Store(&l)
	return previous
}

// LimitError names a refused setting and the values that are allowed.
type LimitError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the setting, the refused value, why, and what is allowed.
func (e *LimitError) Error() string {
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

// LoadLimits reads the four bounds from env. An empty value is the default. A
// value that is set and is not a positive number of bytes is refused and is
// never replaced by the default (ADR-0013).
// maxLimitBytes is the largest bound a setting may name: 1 TiB, far above any
// repository read and far below the int64 maximum, so limit+1 never overflows.
const maxLimitBytes int64 = 1 << 40

func LoadLimits(env func(string) string) (Limits, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	l := DefaultLimits()
	for _, f := range []struct {
		name string
		dst  *int64
	}{
		{EnvMaxArchiveBytes, &l.Archive},
		{EnvMaxFileBytes, &l.File},
		{EnvMaxConfigBytes, &l.Config},
		{EnvMaxRecordBytes, &l.Record},
	} {
		raw := strings.TrimSpace(env(f.name))
		if raw == "" {
			continue
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 {
			return Limits{}, &LimitError{
				Field: f.name, Value: clipValue(raw), Reason: "not a positive number of bytes",
				Allowed: []string{"a positive integer; empty means the default"},
			}
		}
		// Every reader reads limit+1 bytes to tell "at the bound" from "over it",
		// so a bound near the int64 maximum would overflow and read nothing. A
		// bound above maxLimitBytes is refused rather than clipped (ADR-0013).
		if n > maxLimitBytes {
			return Limits{}, &LimitError{
				Field: f.name, Value: clipValue(raw), Reason: "more than the largest bound Proven reads",
				Allowed: []string{"a positive integer up to " + strconv.FormatInt(maxLimitBytes, 10) + "; empty means the default"},
			}
		}
		*f.dst = n
	}
	return l, nil
}

// clipValue keeps a refused value short, so a long one is not echoed whole.
func clipValue(s string) string {
	const max = 40
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Environment lists the read bounds for the configuration page.
func Environment() []envdoc.Setting {
	d := DefaultLimits()
	return []envdoc.Setting{
		{Name: EnvMaxArchiveBytes, Default: envdoc.Tick(strconv.FormatInt(d.Archive, 10)), Description: "the most one export reads from the repository under review: a git archive, or the tracked files of a working tree"},
		{Name: EnvMaxFileBytes, Default: envdoc.Tick(strconv.FormatInt(d.File, 10)), Description: "the largest single file, git blob or commit log read from the repository under review"},
		{Name: EnvMaxConfigBytes, Default: envdoc.Tick(strconv.FormatInt(d.Config, 10)), Description: "the largest `proven.yaml` Proven reads"},
		{Name: EnvMaxRecordBytes, Default: envdoc.Tick(strconv.FormatInt(d.Record, 10)), Description: "the most the author record may hold, summed over the files `author_record` names"},
	}
}

// OverBound is the refusal for a read that went over a bound. It names the
// thing read, the bound and the setting that sets it, and never the content.
func OverBound(what string, bound int64, env string) error {
	return fmt.Errorf("change: %s is larger than the %d-byte bound (setting %s)", what, bound, env)
}
