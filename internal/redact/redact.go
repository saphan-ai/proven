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
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/store"
)

// Names lists every variable whose value Proven redacts: the keys it sends (the
// generator's two, the second reviewer's (ADR-0037), the guardrail's, Zetaris's and the CouchDB store's two passwords), and the
// organisation id it identifies itself with. The generator's NVIDIA key is here so a hosted
// answer that echoes it never reaches the trace (ADR-0031 point 3). The organisation id is
// an identifier, not a credential: an id shorter than service.MinSecretBytes is NOT redacted
// (FromEnv leaves it out), so a short id can appear in a trace.
var Names = []string{generator.EnvToken, generator.EnvKey, generator.EnvSecondKey, guardrail.EnvKey, query.EnvAPIKey, query.EnvOrgID, store.EnvCouchPassword, store.EnvRecordsPassword}

// Redactor replaces configured values in text. The zero value and a nil
// *Redactor leave text unchanged.
type Redactor struct {
	// needles is ordered longest value first, so a value that contains another
	// is replaced whole.
	needles []needle
	// lengths is the trimmed length of each configured value, by variable name.
	lengths map[string]int
}

// MarkerPrefix begins every marker a redacted value is replaced by; the marker is the prefix, the
// variable's name and a closing bracket. Text that holds it was redacted.
const MarkerPrefix = "[redacted: "

type needle struct {
	value, marker string
}

// New builds a Redactor from variable name to value. A value that is empty
// after trimming white space is not redacted; every other value is, whatever
// its length. A secret too short to redact without damaging ordinary text is
// refused by the settings loaders (service.MinSecretBytes), so none reaches
// here from the environment.
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
	r := &Redactor{lengths: map[string]int{}}
	for _, name := range names {
		v := strings.TrimSpace(values[name])
		if v == "" {
			continue
		}
		r.lengths[name] = len(v)
		if seen[v] {
			continue
		}
		// ADR-0005: an error that quotes a value with %q, or a JSON line that carries it, writes
		// it escaped, so each escaped spelling is a needle of its own under the same marker.
		for _, form := range spellings(v) {
			if seen[form] {
				continue
			}
			seen[form] = true
			r.needles = append(r.needles, needle{value: form, marker: MarkerPrefix + name + "]"})
		}
	}
	sort.SliceStable(r.needles, func(i, j int) bool { return len(r.needles[i].value) > len(r.needles[j].value) })
	return r
}

// Shorter returns the names of the configured values shorter than min bytes, sorted.
// A pack writer uses it to refuse a secret too short to redact without damage.
func (r *Redactor) Shorter(min int) []string {
	if r == nil {
		return nil
	}
	var out []string
	for name, n := range r.lengths {
		if n < min {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
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
	add(inner(strconv.Quote(v)))
	add(inner(strconv.QuoteToASCII(v)))
	if b, err := json.Marshal(v); err == nil {
		add(inner(string(b)))
	}
	// An encoder with HTML escaping off (the evidence pack's) keeps <, > and &
	// as written and still leaves what %q escapes, so it spells a value a
	// fourth way.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err == nil {
		add(inner(strings.TrimSuffix(buf.String(), "\n")))
	}
	return out
}

// inner drops the one quote at each end of an encoded string. strings.Trim
// would drop every quote, and a value that ends in a quote ends in an escaped
// one that must stay.
func inner(quoted string) string {
	if len(quoted) < 2 {
		return quoted
	}
	return quoted[1 : len(quoted)-1]
}

// FromEnv builds a Redactor from the variables in Names, read through env.
// A nil env reads nothing.
//
// The five credentials (the generator's two keys, the second reviewer's, the
// guardrail's and Zetaris's) are always handed to the redactor: a short one that
// is in use is refused by its settings loader, so none reaches here. A credential
// set but unused by the active mode is not checked by a loader and is redacted
// as it is. The organisation id is
// an identifier, not a credential, and the loaders accept a short numeric one;
// handed over as it is, an id of one digit would replace that digit in every
// commit id and line number. So an id shorter than service.MinSecretBytes is
// left out here, and New still redacts any value it is given directly.
func FromEnv(env func(string) string) *Redactor {
	values := map[string]string{}
	if env != nil {
		for _, name := range Names {
			v := env(name)
			if name == query.EnvOrgID && service.TooShortSecret(v) {
				continue
			}
			values[name] = v
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
