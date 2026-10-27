package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// htmlGoldenPath names a golden of the HTML report under testdata/golden/html-report.
// go test ./internal/cli -run TestHTMLGolden -update-golden writes it.
func htmlGoldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(moduleRoot(t), "testdata", "golden", "html-report", name)
}

func assertHTMLGolden(t *testing.T, name, got string) {
	t.Helper()
	path := htmlGoldenPath(t, name)
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s: %v (-update-golden writes it)", name, err)
	}
	if string(want) != got {
		t.Fatalf("golden %s differs\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

// sampleHTML is the real sample run on the discount fixture, written with --html.
func sampleHTML(t *testing.T) string {
	t.Helper()
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	env := envOf(map[string]string{"PROVEN_LOG_DIR": t.TempDir()})
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", "--html", target}, env, &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("exit %d\nstderr: %s\nstdout: %s", code, stderr.String(), stdout.String())
	}
	return stdout.String()
}

var (
	htmlHref      = regexp.MustCompile(`(?i)\bhref\s*=\s*("[^"]*"|'[^']*'|[^\s>]*)`)
	htmlTag       = regexp.MustCompile(`<\s*/?\s*([a-zA-Z][a-zA-Z0-9]*)`)
	htmlAttr      = regexp.MustCompile(`\s([a-zA-Z][a-zA-Z0-9-]*)\s*=\s*("[^"]*")`)
	htmlAmpersand = regexp.MustCompile(`&(#[0-9]+;|amp;|lt;|gt;|quot;)?`)
	// htmlAllowedTags are the elements the writer makes, so a tag a value made is named.
	htmlAllowedTags = map[string]bool{
		"html": true, "head": true, "meta": true, "title": true, "style": true, "body": true,
		"main": true, "h1": true, "h2": true, "h3": true, "h4": true, "p": true, "strong": true,
		"span": true, "a": true, "table": true, "thead": true, "tbody": true, "tr": true,
		"th": true, "td": true, "ul": true, "li": true, "section": true, "details": true,
		"summary": true, "pre": true, "code": true, "div": true, "footer": true,
	}
	// htmlAllowedAttrs are the attributes the writer makes; none carries a value from the change.
	htmlAllowedAttrs = map[string]bool{
		"lang": true, "charset": true, "name": true, "content": true, "class": true, "id": true, "href": true,
	}
)

// selfContainedProblems lists every way the body is not one self-contained page:
// a script, a link element, a src, an href that is not an in-page anchor, an
// address, or a stylesheet that fetches. The style element is cut out first only
// for the tag count; the address and src checks read the whole file.
func selfContainedProblems(body string) []string {
	var bad []string
	lower := strings.ToLower(body)
	for _, banned := range []string{"<script", "<link", "src=", "http://", "https://", "@import", "url("} {
		if strings.Contains(lower, banned) {
			bad = append(bad, "the page contains "+banned)
		}
	}
	for _, m := range htmlHref.FindAllStringSubmatch(body, -1) {
		if !strings.HasPrefix(strings.Trim(m[1], `"'`), "#") {
			bad = append(bad, "href that is not an in-page anchor: "+m[0])
		}
	}
	return bad
}

// TestHTMLSampleRunIsSelfContained proves the HTML of the sample run is one file
// with nothing to fetch: no script, no link element, no src, no href but an
// in-page anchor, and no http or https address. The property fails when the
// writer adds any external resource.
func TestHTMLSampleRunIsSelfContained(t *testing.T) {
	body := sampleHTML(t)
	if bad := selfContainedProblems(body); len(bad) != 0 {
		t.Fatalf("the page is not self-contained:\n%s\n--- body\n%s", strings.Join(bad, "\n"), body)
	}
	for _, want := range []string{"<!DOCTYPE html>", `<meta charset="utf-8">`, "<style>", "<details>", `href="#finding-1"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("the page lacks %s\n%s", want, body)
		}
	}
}

// TestHTMLSelfContainedCheckBitesOnOneExternalLink is the negative control of the
// self-contained test (law #27). It mutates exactly one thing: one external
// stylesheet <link> added to the real page. The same check that passes on the
// page must turn red, or the test above proves nothing.
func TestHTMLSelfContainedCheckBitesOnOneExternalLink(t *testing.T) {
	body := sampleHTML(t)
	if bad := selfContainedProblems(body); len(bad) != 0 {
		t.Fatalf("the control starts from a page that is already not self-contained: %v", bad)
	}
	mutated := strings.Replace(body, "<style>", `<link rel="stylesheet" href="https://cdn.example/x.css"><style>`, 1)
	if mutated == body {
		t.Fatal("the mutation changed nothing: no <style> was found")
	}
	if bad := selfContainedProblems(mutated); len(bad) == 0 {
		t.Fatal("one external stylesheet link passed the check")
	}
}

// TestHTMLGoldenRunWithFindings proves the page of the sample run: the verdict,
// the summary, a card per finding with how we know and its evidence folded, the
// opinions section and the foot. The property fails when any part of it moves.
func TestHTMLGoldenRunWithFindings(t *testing.T) {
	assertHTMLGolden(t, "findings.html", sampleHTML(t))
}

// TestHTMLGoldenCleanRun proves a clean run is a page with the checked statement
// and no card. The mutated field is the findings list: empty here.
func TestHTMLGoldenCleanRun(t *testing.T) {
	body, err := render("html", review.Result{
		Class: "docs", ClassRule: "fallback", Reviewers: []review.ReviewerInfo{{Role: review.RoleDefault, Model: "m1", Family: "family acme"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), `class="card`) || strings.Contains(string(body), "<details>") {
		t.Fatalf("a clean page carries a card or a <details>\n%s", body)
	}
	assertHTMLGolden(t, "clean.html", string(body))
}

// htmlHostile is a value from a change that tries to leave the text it is put in.
var htmlHostile = []string{
	"<script>alert(1)</script>",
	`"><img src=x onerror=alert(1)>`,
	"a & b &amp; c &#60;",
	"https://evil.example/x?a=1&b=2",
	"'><svg onload=alert(1)>",
}

// htmlHostileResult puts the whole attack in every field the HTML writer prints
// from the change: title, message, path, query, rows, test output, tool output,
// fix text (proved, a dependency raise and an unproved note), the opinion, the models,
// a failed service, an accepted and an expired exception, and the policy and
// settings statements.
func htmlHostileResult() review.Result {
	attack := strings.Join(htmlHostile, " ")
	multi := strings.Join(htmlHostile, "\n")
	return review.Result{
		Class: attack, ClassRule: attack,
		PolicyStatement: attack, SettingsStatement: attack,
		Expired:        []review.ExpiredException{{ExceptionID: attack, Finding: attack, Expires: attack}},
		ReviewerModels: []string{attack},
		AuthorModels:   []review.AuthorModel{{Model: attack, Family: attack}},
		Reviewers:      []review.ReviewerInfo{{Role: review.RoleDefault, Model: attack, Family: attack}},
		FailedServices: []review.ServiceFailure{{Service: attack, Reason: attack}},
		Opinions:       []review.Opinion{{ID: "O1", Title: attack, Location: review.Location{Path: attack, StartLine: 2}}},
		Accepted:       []review.AcceptedFinding{{ExceptionID: attack, Owner: attack, Expires: attack, Finding: review.Finding{Title: attack, Location: review.Location{Path: attack}}}},
		Findings: []review.Finding{{
			ID: "F2", RuleID: attack, Title: attack, Message: attack, Severity: review.SeverityWarning,
			Location: review.Location{Path: attack},
			Fix:      &review.Fix{Dependency: &review.DependencyFix{Package: attack, From: attack, To: attack}, Why: attack},
		}, {
			ID: "F3", RuleID: attack, Title: attack, Message: attack, Severity: review.SeverityInfo,
			Location: review.Location{Path: attack},
			Fix:      &review.Fix{Note: attack},
		}, {
			ID: "F1", RuleID: attack, Title: attack, Message: attack, Severity: review.SeverityError,
			Location: review.Location{Path: attack, StartLine: 3},
			Evidence: []review.Evidence{
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
					SQL: multi, Engine: attack, Snapshot: attack, Form: attack,
					Columns: []string{attack, "b"}, Rows: [][]string{{multi, attack}},
				}},
				{Kind: review.EvidenceTest, Test: &review.TestEvidence{
					Path: attack, Name: attack, Command: multi, Output: multi,
				}},
				{Kind: review.EvidenceTool, Tool: &review.ToolEvidence{
					Tool: attack, Version: attack, Command: multi, Output: multi, BaseOutput: multi,
				}},
			},
			Fix: &review.Fix{Proved: true, Attempts: 1, What: attack, Why: attack,
				Edits: []review.FixEdit{{Path: attack, Old: multi, New: multi}},
				Runs:  []review.FixRun{{Label: attack, Command: attack, Result: attack}}},
		}},
	}
}

// unsafeHTML lists every way the body lets a value from the change act as markup:
// an element or an attribute the writer does not make, an ampersand that is not
// the start of an entity, a payload in raw form, or anything that makes the page
// reach outside itself. It reads the page itself, a separate reader from the
// writer, so a writer that skips one escape cannot also skip this check.
func unsafeHTML(body string) []string {
	// Not selfContainedProblems: a value shown as text may spell "src=" or "onerror="
	// harmlessly. What matters here is that no element, attribute or address is made.
	var bad []string
	for _, m := range htmlHref.FindAllStringSubmatch(body, -1) {
		if !strings.HasPrefix(strings.Trim(m[1], `"'`), "#") {
			bad = append(bad, "href that is not an in-page anchor: "+m[0])
		}
	}
	if strings.Contains(body, "http://") || strings.Contains(body, "https://") {
		bad = append(bad, "an address survived")
	}
	// The stylesheet is the writer's own fixed text and holds no value from the change.
	page := body
	if start, end := strings.Index(page, "<style>"), strings.Index(page, "</style>"); start >= 0 && end > start {
		page = page[:start] + page[end:]
	}
	for _, m := range htmlTag.FindAllStringSubmatch(page, -1) {
		if !htmlAllowedTags[strings.ToLower(m[1])] {
			bad = append(bad, "an element the writer does not make: "+m[0])
		}
	}
	for _, m := range htmlAttr.FindAllStringSubmatch(page, -1) {
		if !htmlAllowedAttrs[strings.ToLower(m[1])] {
			bad = append(bad, "an attribute the writer does not make: "+m[0])
		}
	}
	for _, m := range htmlAmpersand.FindAllStringSubmatch(page, -1) {
		if m[1] == "" {
			bad = append(bad, "a bare ampersand: "+m[0])
		}
	}
	for _, raw := range []string{`"><img`, "<svg", "<img", "<script"} {
		if strings.Contains(page, raw) {
			bad = append(bad, "a payload in raw form: "+raw)
		}
	}
	return bad
}

// TestHTMLEscapesEveryValueFromTheChange proves a finding whose title, message,
// path, SQL, rows, test output, tool output and fix carry <script>, "><img, an
// ampersand and an https address is shown as text: no element, no attribute and
// no address is made from it, and the text is there, escaped.
func TestHTMLEscapesEveryValueFromTheChange(t *testing.T) {
	body, err := render("html", htmlHostileResult())
	if err != nil {
		t.Fatal(err)
	}
	if bad := unsafeHTML(string(body)); len(bad) != 0 {
		t.Fatalf("the page is not safe:\n%s\n--- body\n%s", strings.Join(bad, "\n"), body)
	}
	for _, want := range []string{"&lt;script&gt;alert(1)&lt;/script&gt;", "&#34;&gt;&lt;img src=x", "a &amp; b &amp;amp; c &amp;#&#8203;60;", "https:&#8203;//evil.example/x?a=1&amp;b=2"} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("the page does not show %q as text\n%s", want, body)
		}
	}
	if n := strings.Count(string(body), `class="card`); n != 3 {
		t.Fatalf("%d cards for three findings\n%s", n, body)
	}
}

// TestHTMLEscapingCheckBitesOnOneUnescapedField is the negative control of the
// escaping test (law #27). It mutates exactly one field: the finding's title in
// the card heading, written raw. The same check that passes on the real page
// must turn red.
func TestHTMLEscapingCheckBitesOnOneUnescapedField(t *testing.T) {
	result := htmlHostileResult()
	body, err := render("html", result)
	if err != nil {
		t.Fatal(err)
	}
	if bad := unsafeHTML(string(body)); len(bad) != 0 {
		t.Fatalf("the control starts from a page that is already unsafe: %v", bad)
	}
	title := result.Findings[0].Title
	mutated := strings.Replace(string(body), "</span> "+htmlText(title)+"</h2>", "</span> "+title+"</h2>", 1)
	if mutated == string(body) {
		t.Fatal("the mutation changed nothing: the heading was not found")
	}
	if bad := unsafeHTML(mutated); len(bad) == 0 {
		t.Fatal("one unescaped title passed the check")
	}
}

// TestHTMLOpinionsAndNotesAreListed proves the opinions are one line each under
// their heading with the sentence that a claim only one model makes is not kept,
// and that a failed service is named in words. The mutated field is Opinions.
func TestHTMLOpinionsAndNotesAreListed(t *testing.T) {
	result := review.Result{
		Opinions:       []review.Opinion{{ID: "O1", Title: "Possible race", Location: review.Location{Path: "a.go", StartLine: 4}}},
		FailedServices: []review.ServiceFailure{{Service: "zetaris", Reason: "ZETARIS_REST_URL is empty"}},
	}
	body, err := render("html", result)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"<h2>Agreed by two models, not proven</h2>",
		"<li>Possible race — <code>a.go:4</code></li>",
		"A claim only one model makes is not kept.",
		"Did not answer: the Zetaris dependency database (not configured for this run)",
		"1 opinion agreed by two models, not proven",
	} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("the page lacks %q\n%s", want, body)
		}
	}
	if strings.Contains(string(body), "ZETARIS_REST_URL") {
		t.Fatalf("a setting name reached the page\n%s", body)
	}
}

// TestHTMLOutputFileIsWrittenBesideTheChosenFormat proves --html-output writes the
// same page the --html standard output prints, beside another chosen format.
func TestHTMLOutputFileIsWrittenBesideTheChosenFormat(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	env := envOf(map[string]string{"PROVEN_LOG_DIR": t.TempDir()})
	out := filepath.Join(t.TempDir(), "report.html")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"scan", "--sample", "--text", "--html-output", out, target}, env, &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("exit %d\nstderr: %s", code, stderr.String())
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(written, []byte("<!DOCTYPE html>")) || strings.Contains(stdout.String(), "<!DOCTYPE") {
		t.Fatalf("the html file or the text output is wrong\nfile: %.80s\nstdout: %.80s", written, stdout.String())
	}
}
