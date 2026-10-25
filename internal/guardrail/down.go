package guardrail

import (
	"strconv"
	"strings"
)

// Down values: why a check could not be made. The set is closed, so a count
// of down checks never groups free text. Each value is one token with no
// space, so a table that splits on spaces keeps its columns.
const (
	DownNotConfigured     = "not-configured"
	DownNoKey             = "no-key"
	DownOff               = "off"
	DownBadVerdict        = "bad-verdict"
	DownTimeout           = "timeout"
	DownUnauthorized      = "unauthorized"
	DownEmptyAnswer       = "empty-answer"
	DownUnreachable       = "unreachable"
	DownNoRecordedVerdict = "no-recorded-verdict"
	// DownOther is every reason outside the set. The reason's text stays in
	// the trace line's text, never in its down value.
	DownOther = "other"
)

// downCodes maps each refusal reason this package makes to its down value.
// An HTTP status is not listed: downCode names it "http-<code>".
var downCodes = map[string]string{
	"not configured": DownNotConfigured,
	// ADR-0005 point 5: an empty key is down, not off. Load names it this way.
	EnvKey + " is empty":  DownNoKey,
	"off":                 DownOff,
	"bad verdict":         DownBadVerdict,
	"timeout":             DownTimeout,
	"unauthorized":        DownUnauthorized,
	"empty answer":        DownEmptyAnswer,
	"unreachable":         DownUnreachable,
	"no recorded verdict": DownNoRecordedVerdict,
}

// downCode is the down value of a refusal reason. httpReason's "http <code>"
// becomes "http-<code>" only for a three-digit status, so the set stays
// bounded; anything else is DownOther.
func downCode(reason string) string {
	if c, ok := downCodes[reason]; ok {
		return c
	}
	if code, ok := strings.CutPrefix(reason, "http "); ok && len(code) == 3 {
		if n, err := strconv.Atoi(code); err == nil && n >= 100 && n <= 599 {
			return "http-" + code
		}
	}
	return DownOther
}
