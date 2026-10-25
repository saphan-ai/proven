// Package fixer proposes a fix for a finding that a test proved, and proves the fix with the same test.
//
// ADR-0042: the Fixer asks the model for a patch to the change. The test-run adapter runs the
// proving test on the base, on the change and on the change with the patch applied, and the other
// tests of the package on the patched change. Only that record, never the model's word, makes a fix
// "proved". A patch that fails is tried again by the same model with the failure in hand, up to the
// number of attempts the settings name; after that the review says the fix is not proved and shows
// no patch. The patch is applied to a scratch copy and is a suggestion: nothing here writes to the change.
// With the whole-change attempt on (ADR-0050) this loop is not run; the attempt keeps the rule that judges
// a test-proved finding, through Judge, and has no retry.
// With the bounded-steps engine on (ADR-0048 point 2), a new attempt carries only what failed and why, as the
// engine's bounded state, instead of every earlier attempt whole.
package fixer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/testrun"
)

// DefaultAttempts is the number of attempts when the settings name none (ADR-0013).
// The first try counts: two attempts is one try and one retry with the failure in hand.
const DefaultAttempts = 2

// maxFileBytes bounds the text of the finding's file that the prompt carries, and
// maxOutputBytes bounds each piece of test output it carries, so a prompt stays small.
const (
	maxFileBytes   = 16 << 10
	maxOutputBytes = 2 << 10
	maxProseBytes  = 2 << 10
)

// fixSpec is the JSON the Fixer asks a model to return. It asks for edits and never
// for a command: Proven builds the commands that run the tests (ADR-0025).
var fixSpec = answer.Spec{
	Task: "Write a fix for the change: edits to its files that make the failing test pass again " +
		"without breaking any other test of the package. " +
		"Each edit names a file, the exact old text, which must stand in the file once, and the new text. " +
		"Never edit a test file and never weaken, skip or delete the test. " +
		"Say in what what the fix changes and in why why it closes the problem.",
	Shape:   "one JSON object. It has what (string), why (string) and edits (a list of objects with path, old and new, all strings). Do not add a command: Proven runs the tests.",
	Example: `{"what":"Subtract the discount once.","why":"The price subtracted the percent twice, so the total was too low.","edits":[{"path":"pkg/price.go","old":"return a - b - b","new":"return a - b"}]}`,
}

// Options are the checkouts the proof runs on and the attempts allowed.
// Attempts <= 0 uses DefaultAttempts. The size limit of a patch is the test runner's own.
type Options struct {
	BaseDir   string
	ChangeDir string
	Attempts  int
	// Carry is the bounded-steps engine's carried-state bound, in estimated tokens (ADR-0048 point 2). Zero shows
	// each new attempt every earlier attempt whole, as before; above zero, a new attempt carries only what failed
	// and why, within that bound, and nothing an earlier attempt proposed.
	Carry int
}

// row is the answer as the model returns it. Command is read only so that an answer
// that carries one can be refused.
type row struct {
	What    string `json:"what"`
	Why     string `json:"why"`
	Edits   []edit `json:"edits"`
	Command string `json:"command"`
}

// edit is one edit as the model returns it.
type edit struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

// Attempt is what one try left behind, for the next prompt and for the record.
type Attempt struct {
	What    string
	Edits   []candidate.FixEdit
	Runs    []candidate.FixRun
	Reason  string
	Failing *candidate.FixRun
	// premise is set when the proof of the problem itself did not hold on this run, so
	// another patch could not help: the test did not pass on the base or fail on the change.
	Premise bool
}

// Prompt asks for a patch for one finding. file is the text of the finding's file in the
// change, or "" when it could not be read. previous holds the failed attempts so far; each
// one is shown so the model sees what was tried and why it failed (ADR-0042 point 6).
func Prompt(c candidate.Candidate, test *candidate.TestEvidence, file string, previous []Attempt) string {
	var b strings.Builder
	fmt.Fprintf(&b, "id: %s\ntitle: %s\nmessage: %s\nlocation: %s:%d-%d\n", c.ID, c.Title, c.Message, c.Path, c.StartLine, c.EndLine)
	fmt.Fprintf(&b, "test: %s %s\ntest output on the change:\n%s\n", test.Path, test.Name, clip(test.Output, maxOutputBytes))
	if file != "" {
		fmt.Fprintf(&b, "file %s in the change:\n%s\n", c.Path, file)
	}
	for i, a := range previous {
		fmt.Fprintf(&b, "\nattempt %d did not prove the fix.\nit changed: %s\n", i+1, clip(a.What, maxProseBytes))
		for _, e := range a.Edits {
			fmt.Fprintf(&b, "edit %s\nold: %s\nnew: %s\n", e.Path, e.Old, e.New)
		}
		fmt.Fprintf(&b, "it failed: %s\n", a.Reason)
		if a.Failing != nil {
			fmt.Fprintf(&b, "failing run: %s, command %s, exit code %d\n%s\n", a.Failing.Label, a.Failing.Command, a.Failing.ExitCode, clip(a.Failing.Output, maxOutputBytes))
		}
	}
	return answer.Instruct(fixSpec, b.String())
}

// CarriedPrompt asks for a patch with the earlier attempts carried as the engine's state (ADR-0048 point 2): the
// finding, the test and the file as in Prompt, then only what failed and why, bounded to carry estimated tokens.
// What an earlier attempt changed, and its edits, are not carried: the model starts each attempt from the change.
func CarriedPrompt(c candidate.Candidate, test *candidate.TestEvidence, file string, previous []Attempt, attempts, carry int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "id: %s\ntitle: %s\nmessage: %s\nlocation: %s:%d-%d\n", c.ID, c.Title, c.Message, c.Path, c.StartLine, c.EndLine)
	fmt.Fprintf(&b, "test: %s %s\ntest output on the change:\n%s\n", test.Path, test.Name, clip(test.Output, maxOutputBytes))
	if file != "" {
		fmt.Fprintf(&b, "file %s in the change:\n%s\n", c.Path, file)
	}
	if len(previous) > 0 {
		failures := make([]markov.FixFailure, 0, len(previous))
		for i, a := range previous {
			f := markov.FixFailure{Attempt: i + 1, Reason: a.Reason}
			if a.Failing != nil {
				f.Run = fmt.Sprintf("%s, command %s, exit code %d: %s", a.Failing.Label, a.Failing.Command, a.Failing.ExitCode, clip(a.Failing.Output, maxOutputBytes))
			}
			failures = append(failures, f)
		}
		block, _ := markov.FixBlock(failures, attempts).Fit(carry)
		fmt.Fprintf(&b, "\nthe earlier attempts did not prove the fix; carried state:\n%s\n", block.Render())
	}
	return answer.Instruct(fixSpec, b.String())
}

// Propose asks for a patch for every kind of finding the Fixer can fix: a finding proved by a
// test gets a patch that is proved by that test (ADR-0042 point 2). A finding with no such test
// is returned as it was, and so is a finding when the generator has no answer to give in sample
// mode. The returned candidate is the best result even when err is non-nil, so a service that
// went down after a failed attempt does not hide the attempt.
func Propose(ctx context.Context, c candidate.Candidate, gen generator.Generator, run testrun.FixRunner, opts Options) (candidate.Candidate, error) {
	test := provingTest(c)
	if test == nil {
		return c, nil
	}
	attempts := opts.Attempts
	if attempts <= 0 {
		attempts = DefaultAttempts
	}
	// The model reads the file the finding names, taken from the change and never
	// from a path it chose: the planner's path goes through the same checks as an edit.
	file, _ := testrun.ReadEditable(opts.ChangeDir, c.Path, maxFileBytes)

	var tried []Attempt
	for n := 1; n <= attempts; n++ {
		var got row
		prompt := Prompt(c, test, file, tried)
		// ADR-0048 point 2: with the engine on, a new attempt carries the failures only, through the same engine.
		if opts.Carry > 0 {
			prompt = CarriedPrompt(c, test, file, tried, attempts, opts.Carry)
		}
		err := answer.Read(ctx, gen, "fixer", prompt, fixSpec, &got)
		if err != nil {
			var missing *generator.Missing
			if errors.As(err, &missing) {
				// A sample that recorded no answer made no claim. With nothing tried yet,
				// the finding stays as it was.
				if len(tried) == 0 {
					return c, nil
				}
				c.Fix = notProved(tried, n-1, fmt.Sprintf("no answer for attempt %d", n))
				return c, nil
			}
			if len(tried) > 0 {
				c.Fix = notProved(tried, n-1, "the model stopped answering")
			}
			return c, err
		}
		a := prove(ctx, got, c, test, run, opts, n)
		if a.err != nil {
			if len(tried) > 0 {
				c.Fix = notProved(tried, n-1, "the proof could not run")
			}
			return c, a.err
		}
		if a.ok {
			c.Fix = &candidate.Fix{
				Kind: candidate.FixPatch, Proved: true, Attempts: n,
				What: clip(got.What, maxProseBytes), Why: clip(got.Why, maxProseBytes),
				Edits: a.Edits, Runs: a.Runs,
			}
			return c, nil
		}
		tried = append(tried, a.Attempt)
		// ADR-0042 point 2: a problem that did not show again on this run cannot be fixed by
		// another patch, so the review says so at once instead of spending the attempts.
		if a.Premise {
			break
		}
	}
	c.Fix = notProved(tried, len(tried), "")
	return c, nil
}

// provingTest returns the test that proved the finding, or nil when none did.
// Only a test that fails on the change and passes on the base is a proof (ADR-0002).
func provingTest(c candidate.Candidate) *candidate.TestEvidence {
	for _, e := range c.Evidence {
		if e.Kind == candidate.KindTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase {
			return e.Test
		}
	}
	return nil
}

// result is one attempt's outcome: ok when the fix is proved, and the record otherwise.
type result struct {
	Attempt
	ok  bool
	err error
}

// prove runs one patch through the test runner and judges the record.
// The judgement lives here and not in the runner, so a runner can only report what ran
// and never decide that a fix is proved (ADR-0042 point 2).
func prove(ctx context.Context, got row, c candidate.Candidate, test *candidate.TestEvidence, run testrun.FixRunner, opts Options, n int) result {
	a := Attempt{What: clip(got.What, maxProseBytes)}
	for _, e := range got.Edits {
		a.Edits = append(a.Edits, candidate.FixEdit{Path: e.Path, Old: e.Old, New: e.New})
	}
	// ADR-0025: the answer is a patch and nothing else. A command in it is refused whole.
	if strings.TrimSpace(got.Command) != "" {
		a.Reason = "the patch was refused: " + testrun.ReasonModelCommand
		return result{Attempt: a}
	}
	edits := make([]testrun.Edit, len(got.Edits))
	for i, e := range got.Edits {
		edits[i] = testrun.Edit{Path: e.Path, Old: e.Old, New: e.New}
	}
	out, err := run.RunFix(ctx, testrun.FixRequest{
		FindingID: c.ID, Path: test.Path, Name: test.Name, Content: test.Content,
		BaseDir: opts.BaseDir, ChangeDir: opts.ChangeDir, Edits: edits, Attempt: n,
	})
	if err != nil {
		return result{Attempt: a, err: err}
	}
	if out.Refused != "" {
		a.Reason = "the patch was refused: " + out.Refused
		return result{Attempt: a}
	}
	for _, r := range out.Runs {
		a.Runs = append(a.Runs, candidate.FixRun{
			Label: r.Label, Command: r.Command, Result: r.Result, ExitCode: r.ExitCode, Output: r.Output,
		})
	}
	ok := judge(&a)
	return result{Attempt: a, ok: ok}
}

// Judge decides the runs of one patched change by the rule of ADR-0042 points 2 and 3, which
// the whole-change attempt keeps for a finding proved by a test (ADR-0050 point 2): the proving
// test passes on the patched change and the package's other tests still pass. It returns the
// reason the runs do not prove the fix, or "" when they do. No model is consulted.
func Judge(runs []candidate.FixRun) (ok bool, reason string) {
	a := Attempt{Runs: runs}
	ok = judge(&a)
	return ok, a.Reason
}

// judge decides a record. The fix is proved only when the proving test passes on the base,
// fails on the change and passes on the patched change, and the package's other tests pass on
// the patched change (ADR-0042 points 2 and 3). It fills the reason and the failing run otherwise.
func judge(a *Attempt) bool {
	by := map[string]*candidate.FixRun{}
	for i := range a.Runs {
		by[a.Runs[i].Label] = &a.Runs[i]
	}
	base, change := by[testrun.RunBase], by[testrun.RunChange]
	if base == nil || base.Result != testrun.ResultPassed {
		a.Premise, a.Failing = true, base
		a.Reason = "the proving test did not pass on the base when it was run again"
		return false
	}
	if change == nil || change.Result != testrun.ResultFailed {
		a.Premise, a.Failing = true, change
		a.Reason = "the proving test did not fail on the change when it was run again"
		return false
	}
	patched := by[testrun.RunPatched]
	if patched == nil || patched.Result != testrun.ResultPassed {
		a.Failing = patched
		a.Reason = "the proving test does not pass on the change with the patch applied"
		return false
	}
	pkg := by[testrun.RunPackage]
	if pkg == nil || pkg.Result != testrun.ResultPassed {
		a.Failing = pkg
		a.Reason = "the patch breaks another test of the package"
		return false
	}
	return true
}

// notProved builds the record of a fix that was not proved (ADR-0042 point 3): no patch is
// kept, because the review does not show a patch as a fix unless it was proved. The last
// attempt's runs stay, so a reader can see what failed. why, when given, replaces the last reason.
func notProved(tried []Attempt, attempts int, why string) *candidate.Fix {
	f := &candidate.Fix{Kind: candidate.FixPatch, Attempts: attempts}
	if len(tried) == 0 {
		f.Note = "fix not proved: " + why
		return f
	}
	last := tried[len(tried)-1]
	f.Runs = last.Runs
	reason := last.Reason
	if why != "" {
		reason += "; " + why
	}
	f.Note = fmt.Sprintf("fix not proved after %d %s: %s", attempts, plural(attempts, "attempt"), reason)
	return f
}

// plural adds an s to word unless n is 1.
func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// clip bounds s at max bytes, on a rune boundary, and says so when it cut.
func clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && s[cut]&0xC0 == 0x80 {
		cut--
	}
	return s[:cut] + "\n[truncated]"
}
