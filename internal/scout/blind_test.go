package scout

// This file runs Scout Intent once over the frozen blind set (test/conformance/meterless/scout-intent/blind/v1)
// and, on request, writes the report. The set was written without reading Scout Intent and is frozen: the
// round that runs it measures and never tunes, so no test here asserts that a case passes. A miss is a
// line in the report, not a failure.
//
// Without the -blind-report flag, the test only checks that every case runs. With it, the test also writes
// blind/v1/REPORT.md. To regenerate the report:
//
//	go test -count=1 -run TestBlindSet ./internal/scout/ -args -blind-report

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// blindReportFlag asks TestBlindSet to write REPORT.md; without it the test writes nothing.
var blindReportFlag = flag.Bool("blind-report", false, "write blind/v1/REPORT.md")

// blindDir is the frozen set's directory, relative to this package.
const blindDir = "../../test/conformance/meterless/scout-intent/blind/v1"

// blindRunCommit is the commit whose Scout Intent code this report measures; blindSetCommit is the commit
// that froze the set. They are the same commit because the set was frozen on the base this round runs on.
const (
	blindRunCommit = "ec7896fe80ca62bc8b95053728791f68667add49"
	blindSetCommit = "ec7896fe80ca62bc8b95053728791f68667add49"
)

// blindCase is one line of cases.jsonl.
type blindCase struct {
	ID       string   `json:"id"`
	Prompt   string   `json:"prompt"`
	Surface  string   `json:"surface"`
	Tags     []string `json:"tags"`
	Expected struct {
		Intent        string `json:"intent"`
		Clarification bool   `json:"clarification"`
		Reason        string `json:"reason"`
		Why           string `json:"why"`
	} `json:"expected"`
}

// blindAnswer is what the review's call path concludes for one case: a kind, or a clarification.
type blindAnswer struct {
	Kind          string // the kind the review plans, empty when Clarification is set
	Clarification bool
	Detail        string // the band and confidence, for the report
}

func (a blindAnswer) String() string {
	if a.Clarification {
		return "clarification"
	}
	return a.Kind
}

// blindClassify runs one change through the same calls internal/review/plan.go's choosePlan makes to
// pick a kind: the shipped change registry with no extra security patterns, Decide on the formatted
// change, then the review's two rules: a security signal is never outvoted, and an ambiguous or
// not-high-band decision is a clarification (the review runs the full plan and names no kind).
// The review's later steps (prose verification, the plan table) need real files and are not part of
// the kind, so they are not run here.
func blindClassify(t *testing.T, prompt string) blindAnswer {
	t.Helper()
	reg, err := ChangeRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	eng, err := New(reg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	d, err := eng.Decide(Input{Prompt: FormatChange(ParseChange(prompt)), User: User{ID: "proven", Role: "review"}, Surface: "change"})
	if err != nil {
		t.Fatal(err)
	}
	detail := fmt.Sprintf("band %s, confidence %.3f", d.Band, d.Confidence)
	for _, c := range d.Candidates {
		if c.Intent == KindSecurity && c.Score > 0 {
			return blindAnswer{Kind: KindSecurity, Detail: detail}
		}
	}
	if d.Ambiguity.Ambiguous || d.Clarification != nil || d.Band != BandHigh {
		return blindAnswer{Clarification: true, Detail: detail}
	}
	return blindAnswer{Kind: d.Intent, Detail: detail}
}

// blindMatch reports whether the answer is the one the case expects. Two clarifications match: the
// reason is not compared, because Scout does not name one in the set's vocabulary.
func blindMatch(c blindCase, a blindAnswer) bool {
	if c.Expected.Clarification {
		return a.Clarification
	}
	return !a.Clarification && a.Kind == c.Expected.Intent
}

// blindLighter reports whether a miss lets the change in under a lighter plan than its expected answer
// needs: Scout answered docs or dependency (ADR-0049 point 2's lighter plans) and the case expects a
// heavier kind or a clarification. Order of weight: docs, then dependency, then everything else.
func blindLighter(c blindCase, a blindAnswer) bool {
	if a.Clarification || blindMatch(c, a) {
		return false
	}
	weight := func(kind string) int {
		switch kind {
		case KindDocs:
			return 0
		case KindDependency:
			return 1
		}
		return 2
	}
	if a.Kind != KindDocs && a.Kind != KindDependency {
		return false
	}
	want := 2
	if !c.Expected.Clarification {
		want = weight(c.Expected.Intent)
	}
	return want > weight(a.Kind)
}

// blindExpected is the case's expected answer as text.
func blindExpected(c blindCase) string {
	if c.Expected.Clarification {
		return "clarification"
	}
	return c.Expected.Intent
}

// TestBlindSet proves one property: every case of the frozen blind set can be run through the review's
// call path and gets an answer. It never asserts that a case matches; that is the report's business.
func TestBlindSet(t *testing.T) {
	f, err := os.Open(filepath.Join(blindDir, "cases.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var cases []blindCase
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		var c blindCase
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			t.Fatalf("cases.jsonl line %d: %v", len(cases)+1, err)
		}
		if c.Surface != "change" {
			t.Fatalf("%s: surface %q, want change", c.ID, c.Surface)
		}
		cases = append(cases, c)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("the blind set is empty")
	}
	answers := make([]blindAnswer, len(cases))
	for i, c := range cases {
		answers[i] = blindClassify(t, c.Prompt)
		if !answers[i].Clarification && answers[i].Kind == "" {
			t.Errorf("%s: no answer", c.ID)
		}
	}
	if !*blindReportFlag {
		return
	}
	out := filepath.Join(blindDir, "REPORT.md")
	if err := os.WriteFile(out, []byte(blindReport(cases, answers)), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s", out)
}

// blindReport renders the report. It lists every case and every miss; nothing is dropped.
func blindReport(cases []blindCase, answers []blindAnswer) string {
	var b strings.Builder
	matched, lighter := 0, 0
	tagTotal, tagMatch := map[string]int{}, map[string]int{}
	var misses, lights []int
	for i, c := range cases {
		ok := blindMatch(c, answers[i])
		if ok {
			matched++
		} else {
			misses = append(misses, i)
		}
		if blindLighter(c, answers[i]) {
			lighter++
			lights = append(lights, i)
		}
		for _, tg := range c.Tags {
			tagTotal[tg]++
			if ok {
				tagMatch[tg]++
			}
		}
	}
	fmt.Fprintf(&b, "# Scout Intent on the blind set, v1: report\n\n")
	fmt.Fprintf(&b, "Scout Intent ran once over the frozen set. Nothing was changed to improve a result. Every miss is as it came out.\n\n")
	fmt.Fprintf(&b, "- Scout Intent code measured: commit `%s`.\n- Set frozen at: commit `%s`.\n", blindRunCommit, blindSetCommit)
	fmt.Fprintf(&b, "- Call path: `scout.ChangeRegistry(nil)` (the shipped registry, no extra security patterns), `Decide` on the formatted change, then the two kind rules of `choosePlan` in `internal/review/plan.go` (a security candidate wins; an ambiguous or not-high decision is a clarification).\n")
	fmt.Fprintf(&b, "- Regenerate: `go test -count=1 -run TestBlindSet ./internal/scout/ -args -blind-report`.\n")
	fmt.Fprintf(&b, "- Not run: the prose check of `choosePlan` (a docs answer is also turned into a full plan when a file is not verified prose). The set holds file lists, not files, so a docs answer here is the kind Scout chose, not the plan a review would run.\n\n")
	fmt.Fprintf(&b, "## Totals\n\n")
	fmt.Fprintf(&b, "Overall: %d of %d match; %d miss; %d lighter-review misses.\n\n", matched, len(cases), len(cases)-matched, lighter)
	fmt.Fprintf(&b, "| tag | cases | match | miss |\n|---|---|---|---|\n")
	var tags []string
	for tg := range tagTotal {
		tags = append(tags, tg)
	}
	sort.Strings(tags)
	for _, tg := range tags {
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", tg, tagTotal[tg], tagMatch[tg], tagTotal[tg]-tagMatch[tg])
	}
	fmt.Fprintf(&b, "\n## Lighter-review misses\n\nCases Scout answered as docs or dependency (a lighter plan) where the expected answer is a heavier kind or a clarification: %d.\n\n", lighter)
	row := func(i int) {
		c, a := cases[i], answers[i]
		fmt.Fprintf(&b, "- `%s` (%s): expected %s, actual %s (%s). Files: `%s`. Reviewer's reason: %s\n",
			c.ID, strings.Join(c.Tags, ", "), blindExpected(c), a, a.Detail, strings.ReplaceAll(c.Prompt, "\n", "; "), c.Expected.Why)
	}
	for _, i := range lights {
		row(i)
	}
	fmt.Fprintf(&b, "\n## All misses\n\n%d cases.\n\n", len(misses))
	for _, i := range misses {
		row(i)
	}
	fmt.Fprintf(&b, "\n## Every case\n\n| id | tags | expected | actual | match |\n|---|---|---|---|---|\n")
	for i, c := range cases {
		m := "yes"
		if !blindMatch(c, answers[i]) {
			m = "MISS"
			if blindLighter(c, answers[i]) {
				m = "MISS (lighter review)"
			}
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", c.ID, strings.Join(c.Tags, ", "), blindExpected(c), answers[i], m)
	}
	return b.String()
}
