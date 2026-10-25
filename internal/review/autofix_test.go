package review

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// reviewFixture runs one sample review of a named fixture with body as an explicit proven.yaml.
func reviewFixture(t *testing.T, fixture, body string) Result {
	t.Helper()
	opts := Options{Sample: true, LogDir: t.TempDir(), Env: os.Getenv}
	if body != "" {
		path := filepath.Join(t.TempDir(), "proven.yaml")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		opts.ConfigPath = path
	}
	rev, err := Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(t.Context(), Request{Target: filepath.Join(moduleRoot(t), "testdata", "fixture", fixture)})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// finding returns the verdict row of the finding with this id.
func autofixFinding(t *testing.T, a *Autofix, id string) AutofixFinding {
	t.Helper()
	for _, f := range a.Findings {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("no verdict on %s: %#v", id, a.Findings)
	return AutofixFinding{}
}

// TestAutofixScenarioEndsFixedAutomatically proves ADR-0050 points 1 to 3 end to end in sample mode: one
// patch for both proved findings, each judged by its own test run again on the patched change and the tests
// of its package, ends "fixed automatically, waiting for a person" with the patch as a diff to suggest.
func TestAutofixScenarioEndsFixedAutomatically(t *testing.T) {
	res := reviewFixture(t, "autofix-fixed", "autofix: on\n")
	a := res.Autofix
	if a == nil || a.Verdict != AutofixFixed || !a.Suggested() || a.Refused != "" {
		t.Fatalf("%#v", a)
	}
	if len(a.Findings) != 2 || len(a.Edits) != 2 {
		t.Fatalf("findings %d edits %d", len(a.Findings), len(a.Edits))
	}
	for _, id := range []string{"F1", "F2"} {
		f := autofixFinding(t, a, id)
		if !f.Closed || len(f.Proofs) != 1 || f.Proofs[0].Kind != AutofixProofTest || !f.Proofs[0].Closed {
			t.Fatalf("%s: %#v", id, f)
		}
		runs := f.Proofs[0].Runs
		if len(runs) != 4 || runs[2].Label != "change+patch" || runs[2].ExitCode != 0 || runs[3].Label != "package on change+patch" || runs[3].ExitCode != 0 {
			t.Fatalf("%s: the re-run evidence is not shown with exit codes: %#v", id, runs)
		}
	}
	for _, want := range []string{
		"--- a/billing/discount.go\n+++ b/billing/discount.go\n",
		"-\treturn cents - cents*percent/100 - cents*percent/100\n+\treturn cents - cents*percent/100\n",
		"--- a/shipping/shipping.go\n+++ b/shipping/shipping.go\n",
		"+\tif cents >= 5000 {\n+\t\treturn 0\n+\t}\n+\treturn 500\n",
	} {
		if !strings.Contains(a.Diff, want) {
			t.Fatalf("the diff lacks %q:\n%s", want, a.Diff)
		}
	}
	if a.What == "" || a.Why == "" {
		t.Fatalf("the explanation was lost: %#v", a)
	}
	// ADR-0050 point 4: the patch is a suggestion. Proven wrote nothing to the reviewed trees.
	for _, side := range []string{"base", "change"} {
		b, err := os.ReadFile(filepath.Join(moduleRoot(t), "testdata", "fixture", "autofix-fixed", side, "billing", "discount.go"))
		if err != nil || (side == "change" && !strings.Contains(string(b), "- cents*percent/100\n")) {
			t.Fatalf("%s tree changed: %v %s", side, err, b)
		}
	}
}

// TestAutofixScenarioEndsRejected proves ADR-0050 point 3 end to end: the same two findings, but the patch
// returns 500 for every order, so F2's own test passes again and the tests of its package break. F1 closes,
// F2 stays open with the failing run and its exit code, and the verdict is "rejected", not a suggestion.
func TestAutofixScenarioEndsRejected(t *testing.T) {
	res := reviewFixture(t, "autofix-rejected", "autofix: on\n")
	a := res.Autofix
	if a == nil || a.Verdict != AutofixRejected || a.Suggested() {
		t.Fatalf("%#v", a)
	}
	f1, f2 := autofixFinding(t, a, "F1"), autofixFinding(t, a, "F2")
	if !f1.Closed || f2.Closed {
		t.Fatalf("closures: F1 %t F2 %t", f1.Closed, f2.Closed)
	}
	open := f2.Proofs[0]
	if open.Closed || !strings.Contains(open.Reason, "the patch breaks another test of the package") {
		t.Fatalf("%#v", open)
	}
	last := open.Runs[len(open.Runs)-1]
	if last.Label != "package on change+patch" || last.ExitCode != 1 || last.Result != "failed" || !strings.Contains(last.Output, "TestFreeShipping") {
		t.Fatalf("the failing run is not shown: %#v", last)
	}
	if a.Diff == "" || len(a.Edits) != 2 {
		t.Fatalf("what was tried is not kept for the record: %#v", a)
	}
}

// TestAutofixOffMakesNoAttempt proves ADR-0050 point 6: with no setting the review makes no attempt, even
// on a change that has proved findings and a recorded answer for the attempt, and the findings are the same
// as with it on.
func TestAutofixOffMakesNoAttempt(t *testing.T) {
	off := reviewFixture(t, "autofix-fixed", "")
	if off.Autofix != nil {
		t.Fatalf("an attempt was made with the setting off: %#v", off.Autofix)
	}
	on := reviewFixture(t, "autofix-fixed", "autofix: on\n")
	if len(on.Findings) != len(off.Findings) || len(off.Findings) != 2 {
		t.Fatalf("findings: on %d off %d", len(on.Findings), len(off.Findings))
	}
	// With the attempt on, it is the one loop: the per-finding suggestions of ADR-0042 are not made.
	for _, f := range on.Findings {
		if f.Fix != nil {
			t.Fatalf("%s carries a per-finding fix beside the whole-change attempt", f.ID)
		}
	}
}

// TestAutofixWithNoProvedFindingMakesNoAttempt proves a review with nothing proved has no verdict: a change
// with no finding that a test, a query or a tool proved is left as it was even with the setting on.
func TestAutofixWithNoProvedFindingMakesNoAttempt(t *testing.T) {
	res := reviewFixture(t, "refuted", "autofix: on\n")
	if res.Autofix != nil {
		t.Fatalf("%#v", res.Autofix)
	}
}

// TestAutofixSettings proves ADR-0050 point 6: autofix is off by default, takes on or off; the generator is
// the review's own unless the second one is named; the time limit is a duration above zero. Each bad value is
// refused by its key with the allowed values.
func TestAutofixSettings(t *testing.T) {
	def := defaultSettings()
	if def.Autofix || def.AutofixGenerator != "default" || def.AutofixTimeout != 10*time.Minute {
		t.Fatalf("defaults: %t %q %s", def.Autofix, def.AutofixGenerator, def.AutofixTimeout)
	}
	set := mustParse(t, "autofix: on\nautofix_generator: special\nautofix_timeout: 90s\n")
	if !set.Autofix || set.AutofixGenerator != "special" || set.AutofixTimeout != 90*time.Second {
		t.Fatalf("%t %q %s", set.Autofix, set.AutofixGenerator, set.AutofixTimeout)
	}
	if off := mustParse(t, "autofix: off\n"); off.Autofix {
		t.Fatal("autofix: off turned it on")
	}
	mustRefuse(t, "autofix: maybe\n", "autofix", "not on or off")
	mustRefuse(t, "autofix_generator: gpt\n", "autofix_generator", "not a generator name")
	mustRefuse(t, "autofix_timeout: 0s\n", "autofix_timeout", "greater than zero")
	mustRefuse(t, "autofix_timeout: -5m\n", "autofix_timeout", "greater than zero")
	mustRefuse(t, "autofix_timeout: soon\n", "autofix_timeout", "greater than zero")
}

// TestAutofixSettingsAreDocumented proves every autofix key has its row in docs/CONFIGURATION.md, so a setting
// cannot ship without the page that tells a person what it does.
func TestAutofixSettingsAreDocumented(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(moduleRoot(t), "docs", "CONFIGURATION.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"autofix", "autofix_generator", "autofix_timeout"} {
		if !strings.Contains(string(b), "| `"+key+"` |") {
			t.Fatalf("docs/CONFIGURATION.md has no row for %s", key)
		}
	}
}

// TestASecondGeneratorThatIsNotConnectedIsNamedNotGuessed proves ADR-0050 point 6: naming the second
// reviewer's generator for the patch with none connected makes no attempt and names the service as failed,
// rather than falling back to the review's own generator without saying so.
func TestASecondGeneratorThatIsNotConnectedIsNamedNotGuessed(t *testing.T) {
	res := reviewFixture(t, "autofix-fixed", "autofix: on\nautofix_generator: special\n")
	if res.Autofix != nil {
		t.Fatalf("an attempt was made without the generator: %#v", res.Autofix)
	}
	named := false
	for _, f := range res.FailedServices {
		if strings.Contains(f.Service, "autofix generator") && strings.Contains(f.Reason, "no second reviewer") {
			named = true
		}
	}
	if !named {
		t.Fatalf("the missing generator is not named: %#v", res.FailedServices)
	}
}

// TestAResultWithNoAttemptEncodesAsItDidBefore proves the evidence pack, which encodes the Result as it is,
// is byte-identical for a review that made no attempt: the field is left out, and appears only when set.
func TestAResultWithNoAttemptEncodesAsItDidBefore(t *testing.T) {
	none, err := json.Marshal(Result{})
	if err != nil || strings.Contains(string(none), "Autofix") {
		t.Fatalf("%v %s", err, none)
	}
	some, err := json.Marshal(Result{Autofix: &Autofix{Verdict: AutofixRejected}})
	if err != nil || !strings.Contains(string(some), `"Autofix":{"Verdict":"rejected"`) {
		t.Fatalf("%v %s", err, some)
	}
}
