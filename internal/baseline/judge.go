package baseline

import (
	"fmt"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/manifest"
)

// Verdict is what Proven's run says about one alert.
type Verdict string

// The four verdicts. NotChecked is a manifest-only run that read the package,
// or an evidence run whose store named no advisory for any package of the
// alert's ecosystem: nothing was found and nothing can be called missed.
const (
	Found            Verdict = "found"
	FoundDifferently Verdict = "found differently"
	Missed           Verdict = "missed"
	NotChecked       Verdict = "not checked"
)

// Row is one alert and Proven's result for it.
type Row struct {
	Alert      Alert
	Verdict    Verdict
	Detail     string
	EvidenceID string
	Reads      string
	// RangeCheck says whether the version Proven read lies inside the range
	// the alert states, by Proven's own version order.
	RangeCheck string
}

// Judge compares one alert with the findings.
//
//   - missed: Proven does not read the package in the alert's manifest, or it
//     reads it and no row of the evidence names the alert's advisory.
//   - found: a row for the alert's GHSA or CVE id affects a version Proven read.
//   - found differently: Proven reads the package and either another advisory
//     affects it, or the alert's advisory is in the evidence but the version
//     lies outside its range.
//   - not checked: a manifest-only run read the package, or the evidence holds
//     no advisory for any package of the alert's ecosystem that Proven read.
func Judge(a Alert, f Findings) Row {
	row := Row{Alert: a}
	var versions []string
	for _, r := range f.Reads {
		if r.File == a.Manifest && r.Package == a.Package && sameEcosystem(r.Ecosystem, a.Ecosystem) {
			versions = append(versions, r.Version)
		}
	}
	sort.Strings(versions)
	if len(versions) == 0 {
		row.Verdict = Missed
		row.Reads = "—"
		row.RangeCheck = "—"
		row.Detail = "Proven does not read this package in this manifest"
		return row
	}
	row.Reads = a.Package + "@" + strings.Join(versions, ", ")
	row.RangeCheck = rangeCheck(a, versions)
	if f.Mode != ModeEvidence {
		row.Verdict = NotChecked
		row.Detail = "manifest-only run: Proven read the package and asked no store"
		return row
	}
	var matches []Match
	for _, m := range f.Matches {
		if m.File == a.Manifest && m.Package == a.Package && sameEcosystem(m.Ecosystem, a.Ecosystem) {
			matches = append(matches, m)
		}
	}
	if !ecosystemAnswered(f, a.Ecosystem) {
		row.Verdict = NotChecked
		row.Detail = fmt.Sprintf("the evidence names no advisory for any %s package Proven read, so it cannot tell a miss from a store without %s advisories", a.Ecosystem, a.Ecosystem)
		return row
	}
	return judgeEvidence(row, a, matches, versions)
}

// ecosystemAnswered reports whether any row of the evidence names an advisory
// for a package of this ecosystem.
func ecosystemAnswered(f Findings, eco string) bool {
	for _, m := range f.Matches {
		if sameEcosystem(m.Ecosystem, eco) && len(m.Advisories) > 0 {
			return true
		}
	}
	return false
}

// kevOf returns the CVEs of this advisory, by its id or an alias, that the KEV
// view holds; a KEV CVE of another advisory of the same version is not one.
func kevOf(adv Advisory, kev []string) []string {
	var out []string
	for _, id := range append([]string{adv.ID}, strings.Fields(adv.Aliases)...) {
		for _, k := range kev {
			if strings.EqualFold(id, k) {
				out = append(out, k)
			}
		}
	}
	return out
}

// judgeEvidence decides an alert whose package Proven read, on the rows it got.
func judgeEvidence(row Row, a Alert, matches []Match, versions []string) Row {
	for _, m := range matches {
		for _, adv := range m.Advisories {
			if adv.Affected && advisoryIs(adv, a) {
				row.Verdict = Found
				row.EvidenceID = m.EvidenceID
				row.Detail = fmt.Sprintf("%s affects %s@%s (%s)", adv.ID, a.Package, m.Version, rangeText(adv))
				if k := kevOf(adv, m.KEV); len(k) > 0 {
					row.Detail += "; on the KEV list: " + strings.Join(k, ", ")
				}
				return row
			}
		}
	}
	for _, m := range matches {
		var other []string
		for _, adv := range m.Advisories {
			if adv.Affected {
				other = append(other, adv.ID)
			}
		}
		if len(other) > 0 {
			row.Verdict = FoundDifferently
			row.EvidenceID = m.EvidenceID
			row.Detail = fmt.Sprintf("%s@%s is affected by %s, and no affecting row names %s", a.Package, m.Version, strings.Join(other, ", "), strings.Join(a.ids(), " or "))
			return row
		}
	}
	for _, m := range matches {
		for _, adv := range m.Advisories {
			if advisoryIs(adv, a) {
				row.Verdict = FoundDifferently
				row.EvidenceID = m.EvidenceID
				row.Detail = fmt.Sprintf("%s is in the evidence (%s), but %s@%s lies outside it", adv.ID, rangeText(adv), a.Package, m.Version)
				return row
			}
		}
	}
	row.Verdict = Missed
	row.Detail = fmt.Sprintf("Proven read %s@%s, and no row of the evidence names %s", a.Package, strings.Join(versions, ", "), strings.Join(a.ids(), " or "))
	return row
}

// advisoryIs reports whether the row is the alert's advisory, by its id or by
// an alias, in either spelling of case.
func advisoryIs(adv Advisory, a Alert) bool {
	have := append([]string{adv.ID}, strings.Fields(adv.Aliases)...)
	for _, id := range a.ids() {
		for _, h := range have {
			if strings.EqualFold(h, id) {
				return true
			}
		}
	}
	return false
}

// rangeText writes a row's range as the evidence states it.
func rangeText(adv Advisory) string {
	intro := adv.Introduced
	if intro == "" {
		intro = "0"
	}
	switch {
	case adv.Fixed != "":
		return ">= " + intro + ", < " + adv.Fixed
	case adv.LastAffected != "":
		return ">= " + intro + ", <= " + adv.LastAffected
	}
	return ">= " + intro
}

// rangeCheck says whether each version Proven read lies inside the alert's
// stated range, by the ecosystem's own version order. A version or a clause
// that cannot be ordered is reported as such and never guessed.
func rangeCheck(a Alert, versions []string) string {
	r, ok := manifest.ByEcosystem(ecosystemName(a.Ecosystem))
	if !ok {
		return "no version order for " + a.Ecosystem
	}
	var parts []string
	for _, v := range versions {
		in, ordered := inStatedRange(r.Compare, v, a.VulnerableRange)
		switch {
		case !ordered:
			parts = append(parts, v+": cannot be ordered")
		case in:
			parts = append(parts, v+": inside")
		default:
			parts = append(parts, v+": outside")
		}
	}
	return strings.Join(parts, "; ")
}

// ecosystemName maps the reference's word to the adapter's.
func ecosystemName(eco string) string {
	for _, r := range manifest.Readers() {
		if sameEcosystem(r.Ecosystem, eco) {
			return r.Ecosystem
		}
	}
	return eco
}

// inStatedRange evaluates a Dependabot range such as ">= 13.0.0, < 15.5.21".
// The bool pair is (inside, ordered); ordered is false when a clause or the
// version does not parse or does not order.
func inStatedRange(cmp func(a, b string) (int, bool), version, rng string) (bool, bool) {
	for _, clause := range strings.Split(rng, ",") {
		clause = strings.TrimSpace(clause)
		var op string
		for _, o := range []string{">=", "<=", ">", "<", "="} {
			if strings.HasPrefix(clause, o) {
				op = o
				break
			}
		}
		if op == "" {
			return false, false
		}
		c, ok := cmp(version, strings.TrimSpace(strings.TrimPrefix(clause, op)))
		if !ok {
			return false, false
		}
		var pass bool
		switch op {
		case ">=":
			pass = c >= 0
		case "<=":
			pass = c <= 0
		case ">":
			pass = c > 0
		case "<":
			pass = c < 0
		default:
			pass = c == 0
		}
		if !pass {
			return false, true
		}
	}
	return true, true
}

// Extra is a package version Proven found affected by an advisory that no
// alert of the reference names for that package in that file.
type Extra struct {
	File       string
	Package    string
	Version    string
	Advisory   string
	Aliases    string
	KEV        bool
	EvidenceID string
}

// Extras lists what Proven found that the reference does not have. A
// manifest-only run found nothing, so it has none.
func Extras(ref Reference, f Findings) []Extra {
	var out []Extra
	seen := map[string]bool{}
	for _, m := range f.Matches {
		for _, adv := range m.Advisories {
			if !adv.Affected || coveredByAlert(ref, m, adv) {
				continue
			}
			key := m.File + "|" + m.Package + "|" + m.Version + "|" + adv.ID
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, Extra{
				File: m.File, Package: m.Package, Version: m.Version, Advisory: adv.ID,
				Aliases: adv.Aliases, KEV: len(kevOf(adv, m.KEV)) > 0, EvidenceID: m.EvidenceID,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		if a.Version != b.Version {
			return a.Version < b.Version
		}
		return a.Advisory < b.Advisory
	})
	return out
}

// coveredByAlert reports whether an alert for the same package in the same
// file names the advisory.
func coveredByAlert(ref Reference, m Match, adv Advisory) bool {
	for _, a := range ref.Alerts {
		if a.Manifest == m.File && a.Package == m.Package && advisoryIs(adv, a) {
			return true
		}
	}
	return false
}
