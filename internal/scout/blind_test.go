package scout_test

// This file runs Scout Intent once over the frozen blind set (test/conformance/meterless/scout-intent/blind/v1)
// and, on request, writes the report. The set was written without reading Scout Intent and is frozen: the
// round that runs it measures and never tunes, so no test here asserts that a case passes. A miss is a
// line in the report, not a failure.
//
// The test is an external test package (scout_test) because it calls the review's exported entry point,
// and the review package imports scout: an internal test could not import it.
//
// Two things are measured, and the report keeps them apart:
//
//  1. The kind Scout Intent chose for each case (blindClassify), the same calls the review makes to pick a kind.
//  2. For every case whose kind is a lighter one (docs or dependency) and is not the expected answer, the plan
//     the live review path actually runs for that change (blindLivePlan): review.Open with scout_intent on,
//     everything else at its shipped default, on a real git change built from the case's file list.
//
// Without the -blind-report flag, the test only checks that every case runs. With it, the test also writes
// blind/v1/REPORT.md. To regenerate the report:
//
//	go test -count=1 -run TestBlindSet ./internal/scout/ -args -blind-report

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/scout"
)

// blindReportFlag asks TestBlindSet to write REPORT.md; without it the test writes nothing.
var blindReportFlag = flag.Bool("blind-report", false, "write blind/v1/REPORT.md")

// blindDir is the frozen set's directory, relative to this package.
const blindDir = "../../test/conformance/meterless/scout-intent/blind/v1"

// blindRunCommit is the commit whose Scout Intent code the kind measurement ran on; blindSetCommit is the
// commit that froze the set. They are the same commit because the set was frozen on the base the first run used.
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
// This is the kind only. The plan the review then runs is measured separately, by blindLivePlan.
func blindClassify(t *testing.T, prompt string) blindAnswer {
	t.Helper()
	reg, err := scout.ChangeRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	eng, err := scout.New(reg, scout.Options{})
	if err != nil {
		t.Fatal(err)
	}
	d, err := eng.Decide(scout.Input{Prompt: scout.FormatChange(scout.ParseChange(prompt)), User: scout.User{ID: "proven", Role: "review"}, Surface: "change"})
	if err != nil {
		t.Fatal(err)
	}
	detail := fmt.Sprintf("band %s, confidence %.3f", d.Band, d.Confidence)
	for _, c := range d.Candidates {
		if c.Intent == scout.KindSecurity && c.Score > 0 {
			return blindAnswer{Kind: scout.KindSecurity, Detail: detail}
		}
	}
	if d.Ambiguity.Ambiguous || d.Clarification != nil || d.Band != scout.BandHigh {
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

// blindLighterKind reports whether a miss is a lighter-kind answer: Scout answered docs or dependency (the
// two kinds ADR-0049 point 2 gives a plan with less work) and the case expects a heavier kind or a
// clarification. It says what kind Scout chose. It does not say what review ran: the live plan can still be
// full (a docs answer for a source file is turned into a full plan), and that is measured by blindLivePlan.
// Order of weight: docs, then dependency, then everything else.
func blindLighterKind(c blindCase, a blindAnswer) bool {
	if a.Clarification || blindMatch(c, a) {
		return false
	}
	if a.Kind != scout.KindDocs && a.Kind != scout.KindDependency {
		return false
	}
	return blindExpectedRank(c) > blindKindRank(a.Kind)
}

// blindKindRank is how much work the default plan table gives a kind (ADR-0049 point 2): docs runs
// nothing (0), dependency skips the Prover (1), every other kind runs in full (2).
func blindKindRank(kind string) int {
	switch kind {
	case scout.KindDocs:
		return 0
	case scout.KindDependency:
		return 1
	}
	return 2
}

// blindExpectedRank is the work the case's expected answer would plan: a clarification runs the full plan.
func blindExpectedRank(c blindCase) int {
	if c.Expected.Clarification {
		return 2
	}
	return blindKindRank(c.Expected.Intent)
}

// blindExpected is the case's expected answer as text.
func blindExpected(c blindCase) string {
	if c.Expected.Clarification {
		return "clarification"
	}
	return c.Expected.Intent
}

// livePlan is the plan the review's live path ran for one case. The agents, tools and sources come from the
// plan line the review writes on its trace before any agent runs; the class and the doubling are what the
// review's own result states for that review (its Class and Doubled fields), so the report quotes the
// review and not a reading of its prose.
type livePlan struct {
	Text    string // the plan in words, for the report
	Rank    int    // 0 runs no agent, tool or source; 1 runs some; 2 runs all of them
	Class   string // the review class the review ran in
	Doubled bool   // whether the review owes a second review; it is run only when a special reviewer is configured
	Trace   string // the plan line as the review wrote it, quoted in the report beside the verdict
}

// blindClassOrder is the closed list of review classes, highest risk first (ADR-0014 point 7). The review
// package keeps it unexported, so the harness holds its own copy. The live class is quoted from the review
// in every report row, so a drift in this copy shows as a wrong lighter-by-class verdict beside the quote.
var blindClassOrder = []string{"security", "backend", "mobile", "user interface", "docs"}

// blindClassRank is a class's place in blindClassOrder: 0 is the highest risk. An unknown class ranks as the
// highest risk, so a name this copy does not know is never read as a lighter review.
func blindClassRank(class string) int {
	for i, c := range blindClassOrder {
		if c == class {
			return i
		}
	}
	return 0
}

// expectedPlan is the plan the case's expected answer would get from the default plan table (ADR-0049
// point 2), in the same terms as a live plan.
type expectedPlan struct {
	Rank    int
	Class   string
	Doubled bool
}

// blindExpectedPlan states the plan the expected answer would run for a case whose live plan is live.
// A security answer runs in full in class security, doubled (ADR-0049 point 2). Every other expected answer,
// a clarification included, runs the plan its kind's rank gives in the class the path rules choose for these
// files, and that is the class the live review states when it runs a full plan: the path rules do not depend
// on what Scout chose. The default doubles only class security (ADR-0009). Where the live review ran a cut
// plan in another class, the files' own class is not known here, so the live class is used: the class cannot
// make that plan look lighter, and the work rank still can.
func blindExpectedPlan(c blindCase, live livePlan) expectedPlan {
	if !c.Expected.Clarification && c.Expected.Intent == scout.KindSecurity {
		return expectedPlan{Rank: 2, Class: "security", Doubled: true}
	}
	return expectedPlan{Rank: blindExpectedRank(c), Class: live.Class, Doubled: live.Class == "security"}
}

// blindLighterPlan says why the live plan is lighter than the expected answer's plan, or returns "" when it
// is not. ADR-0049 point 3: class and doubling are obligations of a plan, not labels, so two plans that run
// the same agents are not equal when their classes differ. A plan is lighter when it runs less work, runs
// in a lower-risk class, or the expected plan owes a second review and the live one does not.
func blindLighterPlan(c blindCase, live livePlan) string {
	exp := blindExpectedPlan(c, live)
	var why []string
	if live.Rank < exp.Rank {
		why = append(why, "less work")
	}
	if blindClassRank(live.Class) > blindClassRank(exp.Class) {
		why = append(why, "lower class")
	}
	if exp.Doubled && !live.Doubled {
		why = append(why, "no second review owed")
	}
	return strings.Join(why, " and ")
}

// blindDoubling is a plan's doubling as the report writes it.
func blindDoubling(doubled bool) string {
	if doubled {
		return "doubled"
	}
	return "not doubled"
}

// blindPlanLine finds the plan line in the review's trace: the first event of kind plan that names the
// class Scout chose. The trace is the only place a review states the plan it chose.
func blindPlanLine(tracePath string) (string, error) {
	f, err := os.Open(tracePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var ev struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			return "", err
		}
		if ev.Kind == "plan" && strings.HasPrefix(ev.Text, "scout-intent: class ") {
			return ev.Text, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no plan line in %s", tracePath)
}

// planParts reads the agents, tools and sources a chosen plan runs, as "name+name" lists.
var planParts = regexp.MustCompile(`agents ([^,]+), tools ([^,]+), sources ([^,]+), doubled`)

// blindParsePlan turns a plan line into the words and the work rank the report shows.
func blindParsePlan(line string) livePlan {
	if m := planParts.FindStringSubmatch(line); m != nil {
		n := 0
		for _, part := range m[1:4] {
			if part != "none" {
				n += len(strings.Split(part, "+"))
			}
		}
		all := len(pipeline.PlanAgents) + len(pipeline.PlanTools) + len(pipeline.PlanSources)
		rank := 1
		switch n {
		case 0:
			rank = 0
		case all:
			rank = 2
		}
		i := strings.Index(line, "plan: class ")
		return livePlan{Text: line[i+len("plan: "):], Rank: rank}
	}
	// A full plan says why in the last clause of the line, after the Scout signals.
	reason := line
	if j := strings.LastIndex(line, "; "); j >= 0 {
		reason = line[j+2:]
	}
	return livePlan{Text: reason, Rank: 2}
}

// blindGit runs one git command in dir and fails the test with its output when it fails.
func blindGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=blind", "-c", "user.email=blind@example.invalid",
		"-c", "commit.gpgsign=false", "-c", "core.quotepath=off", "-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// blindWrite writes one file under root, creating its folders. Content is plain text with no shebang and no
// executable bit, so the review's prose check judges the file by its name and its real git facts.
func blindWrite(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// blindBuildChange builds a git repository whose change from base to head has the case's file list: the
// status letter and path of every line. It returns the root and the two commits. A renamed file keeps its
// content, so git reports a rename; every other file differs between base and head.
func blindBuildChange(t *testing.T, prompt string) (root, base, head string) {
	t.Helper()
	root = t.TempDir()
	blindGit(t, root, "init", "-q")
	// A file the change never touches, so a change made only of additions still has a base commit.
	blindWrite(t, root, "untouched.txt", "untouched\n")
	type step struct{ status, path, from string }
	var steps []step
	for _, line := range strings.Split(prompt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		status, rest, _ := strings.Cut(line, " ")
		if from, to, ok := strings.Cut(rest, " -> "); ok && status == "R" {
			steps = append(steps, step{status, to, from})
			continue
		}
		steps = append(steps, step{status, rest, ""})
	}
	for _, s := range steps {
		switch s.status {
		case "M", "D":
			blindWrite(t, root, s.path, "base "+s.path+"\n")
		case "R":
			blindWrite(t, root, s.from, "base "+s.from+"\n")
		}
	}
	blindGit(t, root, "add", "-A")
	blindGit(t, root, "commit", "-q", "-m", "[blind] fixture: base")
	base = blindGit(t, root, "rev-parse", "HEAD")
	for _, s := range steps {
		switch s.status {
		case "A":
			blindWrite(t, root, s.path, "head "+s.path+"\n")
		case "M":
			blindWrite(t, root, s.path, "head "+s.path+"\nchanged\n")
		case "D":
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(s.path))); err != nil {
				t.Fatal(err)
			}
		case "R":
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(s.from))); err != nil {
				t.Fatal(err)
			}
			blindWrite(t, root, s.path, "base "+s.from+"\n")
		default:
			t.Fatalf("unknown status %q in %q", s.status, prompt)
		}
	}
	blindGit(t, root, "add", "-A")
	blindGit(t, root, "commit", "-q", "-m", "[blind] fixture: head")
	head = blindGit(t, root, "rev-parse", "HEAD")
	return root, base, head
}

// blindLivePlan runs the review's exported entry point on a git change with the case's file list and reads
// back the plan it chose. The only setting changed from the shipped defaults is scout_intent: on, because
// the shipped default is off and then every review runs the full plan with no plan chosen. No model is called:
// the plan is chosen before any agent runs, from the real git status and file modes.
func blindLivePlan(t *testing.T, c blindCase) livePlan {
	t.Helper()
	root, base, head := blindBuildChange(t, c.Prompt)
	work := t.TempDir()
	cfg := filepath.Join(work, "proven.yaml")
	if err := os.WriteFile(cfg, []byte("scout_intent: on\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Sample mode reads a fixture.json that an arbitrary git change does not have, so the review opens in live
	// mode. Its generator is a local stub that refuses every call: the plan is chosen and traced before any
	// agent runs, so no model answer is needed, and the review ends degraded, which is not read here.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "blind set: no model is called", http.StatusServiceUnavailable)
	}))
	defer stub.Close()
	env := func(key string) string {
		switch {
		case key == "PROVEN_GENERATOR_BASE_URL":
			return stub.URL
		case strings.HasSuffix(key, "_TOKEN"):
			return "placeholder" // the live adapter will not open without a bearer value; the stub ignores it
		}
		return ""
	}
	r, err := review.Open(review.Options{ConfigPath: cfg, LogDir: filepath.Join(work, "logs"), Env: env})
	if err != nil {
		t.Fatalf("%s: open: %v", c.ID, err)
	}
	res, err := r.Review(context.Background(), review.Request{Target: root, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatalf("%s: review: %v", c.ID, err)
	}
	line, err := blindPlanLine(res.TracePath)
	if err != nil {
		t.Fatalf("%s: %v", c.ID, err)
	}
	p := blindParsePlan(line)
	p.Class, p.Doubled, p.Trace = res.Class, res.Doubled, line
	return p
}

// blindLoad reads the frozen set.
func blindLoad(t *testing.T) []blindCase {
	t.Helper()
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
	return cases
}

// TestBlindSet proves one property: every case of the frozen blind set can be run through the review's
// call path and gets an answer, and every lighter-kind answer gets a live plan read back from a review.
// It never asserts that a case matches; that is the report's business.
func TestBlindSet(t *testing.T) {
	cases := blindLoad(t)
	answers := make([]blindAnswer, len(cases))
	plans := make([]*livePlan, len(cases))
	for i, c := range cases {
		answers[i] = blindClassify(t, c.Prompt)
		if !answers[i].Clarification && answers[i].Kind == "" {
			t.Errorf("%s: no answer", c.ID)
		}
		if blindLighterKind(c, answers[i]) {
			p := blindLivePlan(t, c)
			plans[i] = &p
		}
	}
	if !*blindReportFlag {
		return
	}
	out := filepath.Join(blindDir, "REPORT.md")
	if err := os.WriteFile(out, []byte(blindReport(cases, answers, plans)), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s", out)
}

// blindReport renders the report. It lists every case and every miss; nothing is dropped.
func blindReport(cases []blindCase, answers []blindAnswer, plans []*livePlan) string {
	var b strings.Builder
	matched, lighterKind, lighterPlan := 0, 0, 0
	tagTotal, tagMatch := map[string]int{}, map[string]int{}
	var misses, lights []int
	for i, c := range cases {
		ok := blindMatch(c, answers[i])
		if ok {
			matched++
		} else {
			misses = append(misses, i)
		}
		if blindLighterKind(c, answers[i]) {
			lighterKind++
			lights = append(lights, i)
			if blindLighterPlan(c, *plans[i]) != "" {
				lighterPlan++
			}
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
	fmt.Fprintf(&b, "- Kind measured by: `scout.ChangeRegistry(nil)` (the shipped registry, no extra security patterns), `Decide` on the formatted change, then the two kind rules of `choosePlan` in `internal/review/plan.go` (a security candidate wins; an ambiguous or not-high decision is a clarification). This is the kind Scout Intent chose.\n")
	fmt.Fprintf(&b, "- Plan measured by: `review.Open` and `Review` (live mode, against a local stub that refuses every model call, so no model runs) on a real git change built from the case's file list, with `scout_intent: on` and every other setting at its shipped default. The plan is read from the plan line the review writes on its trace before any agent runs. It is measured only for the lighter-kind answers below. `scout_intent` is off in the shipped default, and then every review runs the full plan.\n")
	fmt.Fprintf(&b, "- Regenerate: `go test -count=1 -run TestBlindSet ./internal/scout/ -args -blind-report`.\n")
	fmt.Fprintf(&b, "- Not measured: the plan for any case whose kind is not a lighter-kind answer; the file contents (each file in the built change holds a one-line placeholder, no shebang, not executable, so the prose check judges the name and the git mode only); any model verdict; the second review itself. A doubled plan owes a second review, and it is run only when a special reviewer is configured (none is configured by default), so this report compares the obligation the review states (its class and its doubling), not a second review that ran.\n\n")
	fmt.Fprintf(&b, "## Totals\n\n")
	fmt.Fprintf(&b, "Overall: %d of %d match; %d miss; %d lighter-kind answers; %d of those run a lighter plan than the expected answer would (compared on agents, tools and sources, on class, and on doubling).\n\n", matched, len(cases), len(cases)-matched, lighterKind, lighterPlan)
	fmt.Fprintf(&b, "| tag | cases | match | miss |\n|---|---|---|---|\n")
	var tags []string
	for tg := range tagTotal {
		tags = append(tags, tg)
	}
	sort.Strings(tags)
	for _, tg := range tags {
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", tg, tagTotal[tg], tagMatch[tg], tagTotal[tg]-tagMatch[tg])
	}
	fmt.Fprintf(&b, "\n## Lighter-kind answers\n\nCases Scout answered with a lighter kind (docs or dependency) where the expected answer is a heavier kind or a clarification: %d. This lists the kind Scout chose, then the plan the live path ran for that change. A plan is compared on three things: the agents, tools and sources it runs; the review class it runs in; and whether it owes a second review. A plan in class docs is lighter than one in class security even when the same agents run. %d of the %d ran a lighter plan than the expected answer would.\n\nFor each: the expected plan's class and doubling, the live plan's class and doubling, and whether the live plan is lighter. The live class and doubling are the review's own statement; the trace line is quoted as the review wrote it. The expected plan is the default plan table's plan for the expected answer, in the class the path rules give the same files.\n\n", lighterKind, lighterPlan, lighterKind)
	for _, i := range lights {
		c, a := cases[i], answers[i]
		exp := blindExpectedPlan(c, *plans[i])
		verdict := "not lighter than expected"
		if why := blindLighterPlan(c, *plans[i]); why != "" {
			verdict = "LIGHTER than expected (" + why + ")"
		}
		fmt.Fprintf(&b, "- `%s` (%s): expected %s, actual kind %s (%s). Files: `%s`. Expected plan: class %s, %s. Live plan: class %s, %s. This plan is %s. Trace: %q. Reviewer's reason: %s\n",
			c.ID, strings.Join(c.Tags, ", "), blindExpected(c), a, a.Detail, strings.ReplaceAll(c.Prompt, "\n", "; "),
			exp.Class, blindDoubling(exp.Doubled), plans[i].Class, blindDoubling(plans[i].Doubled), verdict, plans[i].Trace, c.Expected.Why)
	}
	row := func(i int) {
		c, a := cases[i], answers[i]
		fmt.Fprintf(&b, "- `%s` (%s): expected %s, actual %s (%s). Files: `%s`. Reviewer's reason: %s\n",
			c.ID, strings.Join(c.Tags, ", "), blindExpected(c), a, a.Detail, strings.ReplaceAll(c.Prompt, "\n", "; "), c.Expected.Why)
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
			if blindLighterKind(c, answers[i]) {
				m = "MISS (lighter kind)"
			}
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", c.ID, strings.Join(c.Tags, ", "), blindExpected(c), answers[i], m)
	}
	return b.String()
}
