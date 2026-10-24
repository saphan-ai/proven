// Package redact removes the values of Proven's configured secrets from text
// before that text is written or printed.
//
// ADR-0005 and ADR-0013 point 4: no key reaches the trace, the result, a
// refusal reason or an error message. A service can echo a value it was sent,
// in a routed model, a receipt id or a safety category, so the settings
// loaders never print a secret and every writer that carries a service's text
// passes it through one Redactor built from the values actually configured.
// A value is replaced by a fixed marker that names its variable, never by a
// part of the value.
package redact

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/query"
)

// Names lists every variable whose value Proven must never write: the three
// keys it sends, and the organisation id it identifies itself with.
var Names = []string{generator.EnvToken, guardrail.EnvKey, query.EnvAPIKey, query.EnvOrgID}

// MinLength is the shortest value that is redacted as a substring.
// A value shorter than this would be found inside ordinary words, numbers and
// line counts, and replacing it would destroy the text it sits in; no key a
// service issues is this short. A shorter value is left alone here, and the
// settings loaders still never print it in a refusal.
const MinLength = 6

// Redactor replaces configured values in text. The zero value and a nil
// *Redactor leave text unchanged.
type Redactor struct {
	// needles is ordered longest value first, so a value that contains another
	// is replaced whole.
	needles []needle
}

type needle struct {
	value, marker string
}

// New builds a Redactor from variable name to value. An empty value, and a
// value shorter than MinLength after trimming white space, is not redacted.
// A value is redacted in its raw form and in the escaped forms %q, %+q and JSON give it.
// When two variables hold the same value the name that sorts first is used,
// so the marker does not depend on map order.
func New(values map[string]string) *Redactor {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	seen := map[string]bool{}
	r := &Redactor{}
	for _, name := range names {
		v := strings.TrimSpace(values[name])
		if len(v) < MinLength || seen[v] {
			continue
		}
		// ADR-0005: an error that quotes a value with %q, or a JSON line that carries it, writes
		// it escaped, so each escaped spelling is a needle of its own under the same marker.
		for _, form := range spellings(v) {
			if seen[form] {
				continue
			}
			seen[form] = true
			r.needles = append(r.needles, needle{value: form, marker: "[redacted: " + name + "]"})
		}
	}
	sort.SliceStable(r.needles, func(i, j int) bool { return len(r.needles[i].value) > len(r.needles[j].value) })
	return r
}

// spellings returns v and the ways Go's %q, %+q and a JSON encoder write it inside quotes, without
// the quotes, each once. A value with nothing to escape has one spelling.
func spellings(v string) []string {
	out := []string{v}
	add := func(s string) {
		for _, have := range out {
			if have == s {
				return
			}
		}
		out = append(out, s)
	}
	add(strings.Trim(strconv.Quote(v), `"`))
	add(strings.Trim(strconv.QuoteToASCII(v), `"`))
	if b, err := json.Marshal(v); err == nil {
		add(strings.Trim(string(b), `"`))
	}
	return out
}

// FromEnv builds a Redactor from the variables in Names, read through env.
// A nil env reads nothing.
func FromEnv(env func(string) string) *Redactor {
	values := map[string]string{}
	if env != nil {
		for _, name := range Names {
			values[name] = env(name)
		}
	}
	return New(values)
}

// Text returns s with every configured value replaced by its marker.
func (r *Redactor) Text(s string) string {
	if r == nil {
		return s
	}
	for _, n := range r.needles {
		s = strings.ReplaceAll(s, n.value, n.marker)
	}
	return s
}

// Texts returns a copy of ss with Text applied to each element. A nil slice
// stays nil, so "none" and "empty" stay distinct.
func (r *Redactor) Texts(ss []string) []string {
	if ss == nil {
		return nil
	}
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = r.Text(s)
	}
	return out
}
