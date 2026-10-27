package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestHTMLVerdictHasASpaceBeforeTheDash proves the verdict reads "Proven: blocked —
// 4 findings", not "Proven: blocked— 4 findings", on a blocked, a clean and a
// not-reviewed page and in the goldens. The mutated line is the format string in
// htmlVerdict: without its space after </strong> every row is red.
func TestHTMLVerdictHasASpaceBeforeTheDash(t *testing.T) {
	blocked := review.Result{Findings: []review.Finding{{ID: "F1", Title: "t", Severity: review.SeverityError}}}
	notReviewed := review.Result{FailedServices: []review.ServiceFailure{{Service: "nightshift", Reason: "timeout"}}}
	for name, tc := range map[string]struct {
		result review.Result
		want   string
	}{
		"blocked":      {blocked, "Proven: blocked</strong> — 1 finding"},
		"clean":        {review.Result{}, "Proven: clean</strong> — 0 findings"},
		"not reviewed": {notReviewed, "Proven: not reviewed</strong> — "},
	} {
		body, err := render("html", tc.result)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), tc.want) {
			t.Errorf("%s: the verdict lacks %q\n%s", name, tc.want, body)
		}
	}
	for _, golden := range []string{"findings.html", "clean.html"} {
		raw, err := os.ReadFile(htmlGoldenPath(t, golden))
		if err != nil {
			t.Fatal(err)
		}
		if want := string(raw); !strings.Contains(want, "</strong> —") || strings.Contains(want, "</strong>—") {
			t.Errorf("golden %s does not hold the space before the dash", golden)
		}
	}
}

// longHTMLResult is a result whose fix explanation, notes and exception lists are
// longer than every Markdown bound: each value ends with a marker no bounded
// writer can reach. Accepted and expired hold more than mdMaxNotes entries.
func longHTMLResult() (review.Result, []string) {
	long := strings.Repeat("x", mdMaxMessage+50)
	longNote := strings.Repeat("n", mdMaxNote+50)
	longTitle := strings.Repeat("t", mdMaxTitle+50)
	r := review.Result{
		PolicyStatement: longNote + "_POLICY_END",
		Findings: []review.Finding{{
			ID: "F1", Title: longTitle + "_TITLE_END", Severity: review.SeverityError,
			Location: review.Location{Path: "a.go", StartLine: 1},
			Fix:      &review.Fix{Proved: true, Attempts: 1, What: long + "_WHAT_END", Why: long + "_WHY_END"},
		}, {
			ID: "F2", Title: "second", Severity: review.SeverityWarning,
			Location: review.Location{Path: "b.go", StartLine: 1},
			Fix:      &review.Fix{Note: long + "_NOTE_END"},
		}},
	}
	want := []string{"_POLICY_END", "_TITLE_END", "_WHAT_END", "_WHY_END", "_NOTE_END"}
	for i := 0; i < mdMaxNotes+5; i++ {
		id := fmt.Sprintf("ACCEPT-%02d", i)
		r.Accepted = append(r.Accepted, review.AcceptedFinding{ExceptionID: id, Owner: "o", Expires: "2027-01-01",
			Finding: review.Finding{Title: "accepted", Location: review.Location{Path: "c.go"}}})
		want = append(want, id)
		id = fmt.Sprintf("EXPIRE-%02d", i)
		r.Expired = append(r.Expired, review.ExpiredException{ExceptionID: id, Finding: "f", Expires: "2026-01-01"})
		want = append(want, id)
	}
	return r, want
}

// missingFrom names every marker that is not in body.
func missingFrom(body string, markers []string) []string {
	var missing []string
	for _, m := range markers {
		if !strings.Contains(body, m) {
			missing = append(missing, m)
		}
	}
	return missing
}

// TestHTMLCarriesTheFullFixNotesAndExceptionLists proves the HTML page is never
// cut: a fix explanation past mdMaxMessage, a note past mdMaxNote, a title past
// mdMaxTitle and more than mdMaxNotes accepted and expired exceptions all appear
// whole (ADR-0033 point 8), and says no "more, not listed". The Markdown comment
// keeps its bounds for the platform's size limit, which is the same result read
// by the other writer. The mutated fields are the lengths and counts above.
func TestHTMLCarriesTheFullFixNotesAndExceptionLists(t *testing.T) {
	result, markers := longHTMLResult()
	html, err := render("html", result)
	if err != nil {
		t.Fatal(err)
	}
	if missing := missingFrom(string(html), markers); len(missing) != 0 {
		t.Errorf("the HTML page cuts %v", missing)
	}
	if strings.Contains(string(html), "not listed") {
		t.Error("the HTML page says an exception is not listed")
	}
	md, err := render("md", result)
	if err != nil {
		t.Fatal(err)
	}
	if missing := missingFrom(string(md), markers); len(missing) == 0 {
		t.Error("the Markdown comment lost its bounds: nothing is cut")
	}
	if !strings.Contains(string(md), "more accepted by exception, not listed") {
		t.Error("the Markdown comment no longer cuts the accepted list with its count")
	}
}

// TestHTMLCompletenessCheckBitesWhenTheMarkdownBoundIsPutBack is the negative
// control of the test above (law #27, #41). It puts the Markdown bound back on
// the HTML path by clearing the one field that lifts it, htmlStyle.unbounded, and
// the same check must turn red.
func TestHTMLCompletenessCheckBitesWhenTheMarkdownBoundIsPutBack(t *testing.T) {
	result, markers := longHTMLResult()
	saved := htmlStyle
	t.Cleanup(func() { htmlStyle = saved })
	htmlStyle.unbounded = false
	body, err := render("html", result)
	if err != nil {
		t.Fatal(err)
	}
	if len(missingFrom(string(body), markers)) == 0 {
		t.Fatal("with the Markdown bound put back on the HTML path the check found nothing cut")
	}
}

// TestHTMLCarriesTheIncompleteHeading proves a review whose model answered only
// some calls is headed "Proven: incomplete" on the HTML page, not "clean" and not
// "not reviewed", and that with no answered call it is "not reviewed". The mutated
// field is Result.Models. Negative control (run on the writer, not in this test):
// replacing the verdictLine call in format_html.go turns this test red.
func TestHTMLCarriesTheIncompleteHeading(t *testing.T) {
	partial := review.Result{Class: "backend", ClassRule: "fallback",
		Models:         []review.ModelUse{{Agent: "planner", Model: "glm-5.3"}},
		FailedServices: []review.ServiceFailure{{Service: "nightshift", Reason: "timeout"}}}
	body, err := render("html", partial)
	if err != nil {
		t.Fatal(err)
	}
	const want = "<strong>Proven: incomplete</strong> — the model reviewer did not answer every call"
	if !strings.Contains(string(body), want) {
		t.Fatalf("the page lacks %q\n%s", want, body)
	}
	for _, bad := range []string{"Proven: clean", "Proven: not reviewed"} {
		if strings.Contains(string(body), bad) {
			t.Errorf("a partly answered review says %q", bad)
		}
	}
	partial.Models = nil
	none, _ := render("html", partial)
	if !strings.Contains(string(none), "<strong>Proven: not reviewed</strong>") {
		t.Errorf("with no model answered the heading is not 'not reviewed'\n%s", none)
	}
}

// TestTwoStandardOutputFormatsAreRefusedWithHTML proves --md with --html is
// refused before the review runs, with the sentence naming every choice. The
// mutated field is the --html flag beside another format; with the html count
// removed from the format check the pair is accepted and the row is red.
func TestTwoStandardOutputFormatsAreRefusedWithHTML(t *testing.T) {
	for _, pair := range [][]string{{"--md", "--html"}, {"--html", "--text"}, {"--html", "--json"}, {"--html", "--sarif"}} {
		rev := &captureReviewer{}
		var stdout, stderr bytes.Buffer
		code := Run(append([]string{"scan"}, pair...), emptyEnv, &stdout, &stderr, openCapture(rev))
		if code == exitOK || !strings.Contains(stderr.String(), "choose one of --text, --json, --sarif, --md, --html") {
			t.Errorf("%v: exit %d, stderr %q", pair, code, stderr.String())
		}
		if rev.req.Target != "" {
			t.Errorf("%v: the review ran despite the refusal", pair)
		}
	}
}

// TestHTMLFootCarriesTheGuardrailLine proves the foot of the HTML page ends with
// the counted guardrail line, the same words as the text output. The mutated field
// is the Result's Guardrail counts. Negative control (run on the writer, not in this
// test): removing the footIn call in format_html.go turns this test red.
func TestHTMLFootCarriesTheGuardrailLine(t *testing.T) {
	const want = "Guardrail: NVIDIA nvidia/test-guard-model (NVIDIA hosted API), checked 2 prompts and 1 answer, 1 blocked"
	body, err := render("html", checkedResult(0))
	if err != nil {
		t.Fatal(err)
	}
	page := string(body)
	foot := page[strings.Index(page, "<footer>"):]
	if !strings.Contains(foot, want+"</footer>") {
		t.Fatalf("the foot does not end with the guardrail line\n%s", foot)
	}
}

// TestHTMLNamesTheEvidenceStoreInWords proves the HTML page of the sample run names
// the evidence store and the query language in words ("SQL on the bundled sample
// data"), never the adapter's id or the form's enum (the owner's rule of 27 Oct).
// Mutated line for the negative control: the store name in the How we know sentence.
func TestHTMLNamesTheEvidenceStoreInWords(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	env := envOf(map[string]string{"PROVEN_LOG_DIR": t.TempDir()})
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"scan", "--sample", "--html", target}, env, &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("exit %d\nstderr: %s", code, stderr.String())
	}
	page := stdout.String()
	for _, raw := range []string{"local-sample", "form sql"} {
		if strings.Contains(page, raw) {
			t.Errorf("the page shows the internal name %q", raw)
		}
	}
	if !strings.Contains(page, "the bundled sample data") {
		t.Error("the page does not name the evidence store in words")
	}
}
