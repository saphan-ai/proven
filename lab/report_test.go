package lab

// The report page of a demo pull request (lab/seed/report.sh, lab/seed/md2html.awk, lab/page/report.html.tmpl): the review
// Proven wrote with `proven scan --md`, set in the way GitHub shows a Markdown comment. The inputs are the two Markdown
// files the CLI's own golden tests hold for a review with findings and for a clean one (lab/testdata/report).

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	reportPasswordSentinel = "report-password-sentinel-9f3a"
	reportTokenSentinel    = "report-token-sentinel-71bc"
)

// renderReport runs report.sh as the seed container does, against a state directory that names pull request 3 as the
// defect demo, and returns the page. mode is "render" or "stub"; md is the Markdown on standard input.
func renderReport(t *testing.T, tree, mode, md string, env ...string) string {
	t.Helper()
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	page := filepath.Join(dir, "page")
	for _, d := range []string{state, page} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(state, "prs.tsv"), []byte("3\tdemo-discount\tdiscount\tdefect\n4\tdemo-discount-clean\tdiscount-clean\tclean\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", filepath.Join(tree, "lab", "seed", "report.sh"), mode, "3")
	cmd.Env = append([]string{
		"PATH=" + os.Getenv("PATH"), "STATE_DIR=" + state, "PAGE_DIR=" + page,
		"PRS_FILE=" + filepath.Join(tree, "lab", "seed", "prs.tsv"),
		"REPORT_TEMPLATE=" + filepath.Join(tree, "lab", "page", "report.html.tmpl"),
		"LAB_URL=http://127.0.0.1:3300/", "LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount",
		"LAB_PAGE_URL=http://127.0.0.1:3302/lab", "LAB_MODE=sample",
		// What must never reach a page: the lab's password and a token, present in the environment of the run.
		"LAB_ADMIN_PASSWORD=" + reportPasswordSentinel, "PROVEN_GENERATOR_KEY=" + reportTokenSentinel, "METERLESS_RUNTIME_TOKEN=" + reportTokenSentinel,
	}, env...)
	cmd.Stdin = strings.NewReader(md)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("report.sh %s: %v\n%s", mode, err, out)
	}
	b, err := os.ReadFile(filepath.Join(page, "reports", "3.html"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func goldenMD(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "report", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestReportPageSetsTheReviewLikeAGitHubComment(t *testing.T) {
	page := renderReport(t, root, "render", goldenMD(t, "findings.md"))
	for _, w := range []string{
		`<span class="badge bad">Blocked</span>`,
		`Apply a discount (it is taken off twice) <span class="num">#3</span>`,
		`href="http://127.0.0.1:3300/proven/discount/pulls/3"`,
		`<strong>Proven: blocked</strong> — 4 findings: 4 error, 0 warning, 0 info`,
		// The findings table, with a coloured severity.
		`<th>Severity</th>`, `<td class="sev sev-error">ERROR</td>`, `<code>go.mod:5</code>`, `<code>known-exploited-dependency</code>`,
		// Each finding folded in a details block, closed.
		`<details><summary>ERROR F1 — Discount is applied twice</summary>`, `</details>`,
		// Evidence: a query block, a result table, a failing test with its output. Backslash escapes are gone, angle brackets are not tags.
		`<pre class="fence"><code class="lang-sql">`, `SELECT cve_id, epss, percentile, score_date FROM`, `<td>1.0.0</td>`, `introduced 1.0.0 &lt;= v1.2.3 &lt; fixed 1.2.4`,
		`<code class="lang-text">`, `change: FAIL got 60 want 80`, `<strong>Failing test</strong>`,
		// The foot.
		`<p class="foot"><sub>reviewed by <code>NVIDIA</code>`,
		// Light and dark.
		`prefers-color-scheme: dark`,
	} {
		if !strings.Contains(page, w) {
			t.Errorf("the page lacks %q", w)
		}
	}
	if strings.Contains(page, "<details open") {
		t.Error("a finding is open: the details must be folded")
	}
	if strings.Count(page, "<details>") != strings.Count(page, "</details>") || strings.Count(page, "<details>") != 4 {
		t.Errorf("details blocks: %d opened, %d closed, want 4 and 4", strings.Count(page, "<details>"), strings.Count(page, "</details>"))
	}
	if strings.Count(page, "<table>") != strings.Count(page, "</table>") {
		t.Error("a table is not closed")
	}
	clean := renderReport(t, root, "render", goldenMD(t, "clean.md"))
	if !strings.Contains(clean, `<span class="badge ok">Clean</span>`) || !strings.Contains(clean, "<strong>Proven: clean</strong> — 0 findings") || strings.Contains(clean, "<details>") {
		t.Errorf("the clean page is wrong:\n%s", clean)
	}
}

func TestReportPageCarriesNoPasswordTokenOrPrivateName(t *testing.T) {
	pages := map[string]string{
		"findings": renderReport(t, root, "render", goldenMD(t, "findings.md")),
		"clean":    renderReport(t, root, "render", goldenMD(t, "clean.md")),
		"stub":     renderReport(t, root, "stub", ""),
	}
	private := []string{reportPasswordSentinel, reportTokenSentinel, "/home/", "/Users/", "saphan-studio", "Bearer ", "METERLESS_", "PROVEN_GENERATOR_KEY", "JUDGES_GATEWAY", "password"}
	for name, page := range pages {
		for _, p := range private {
			if strings.Contains(page, p) {
				t.Errorf("%s page carries %q", name, p)
			}
		}
		if regexp.MustCompile(`@@[A-Z_]+@@`).MatchString(page) {
			t.Errorf("%s page still holds a placeholder", name)
		}
		// No script, and no address but the lab's own: the page works offline.
		if strings.Contains(page, "<script") || regexp.MustCompile(`(src|href)="https?://([^1]|1[^2])`).MatchString(page) {
			t.Errorf("%s page holds a script or a remote address", name)
		}
	}
	// The negative control: the template gains the password. The same assertion must fail.
	tree := copyLab(t)
	for _, rel := range []string{"lab/seed", "lab/page"} {
		copyInto(t, tree, rel)
	}
	mutate(t, tree, "lab/page/report.html.tmpl", "@@NOTE@@", "@@NOTE@@ "+"@@LAB_PASSWORD_LEAK@@")
	mutate(t, tree, "lab/seed/report.sh", `s = fill(s, "NOTE", ENVIRON["R_NOTE"])`, `s = fill(s, "NOTE", ENVIRON["R_NOTE"]); s = fill(s, "LAB_PASSWORD_LEAK", ENVIRON["LAB_ADMIN_PASSWORD"])`)
	if leaked := renderReport(t, tree, "render", goldenMD(t, "clean.md")); !strings.Contains(leaked, reportPasswordSentinel) {
		t.Error("the control did not turn red: a password placed in the page was not seen by the check")
	}
}

func TestReportPageWritesNoTagTheReviewDidNotMean(t *testing.T) {
	hostile := strings.Join([]string{
		"**Proven: blocked** — 1 finding: 1 error, 0 warning, 0 info",
		"",
		"<script>alert(1)</script>",
		"<img src=x onerror=alert(2)>",
		"a [link](javascript:alert(3)) and <b onclick=x>bold</b> and & a bare ampersand",
		"",
		"| Severity | Finding | Location | Rule |",
		"| --- | --- | --- | --- |",
		"| ERROR | <iframe src=//evil.example></iframe> | `a<b>.go:1` | `r` |",
		"",
		"<details><summary>ERROR F1 — <script>x</script></summary>",
		"",
		"```html",
		"<script>fenced()</script>",
		"```",
		"",
		"</details>",
	}, "\n") + "\n"
	page := renderReport(t, root, "render", hostile)
	for _, bad := range []string{"<script>alert", "<img", "<iframe", "<b onclick", "<script>x", "<script>fenced"} {
		if strings.Contains(page, bad) {
			t.Errorf("the page writes %q", bad)
		}
	}
	for _, want := range []string{"&lt;script&gt;alert(1)&lt;/script&gt;", "&lt;img src=x", "&lt;script&gt;fenced()&lt;/script&gt;", "a bare ampersand", "&amp; a bare"} {
		if !strings.Contains(page, want) {
			t.Errorf("the page lacks the escaped form %q", want)
		}
	}
	// The page's own script count is zero: the template and the review write none.
	if n := strings.Count(page, "<script"); n != 0 {
		t.Errorf("%d script tags", n)
	}
}

func TestReportStubSaysTheReportIsNotReadyAndNamesTheCommand(t *testing.T) {
	page := renderReport(t, root, "stub", "")
	for _, w := range []string{"Report not ready", "make reports", "the comment on the pull request in Gitea"} {
		if !strings.Contains(page, w) {
			t.Errorf("the stub lacks %q", w)
		}
	}
}

func TestReportPageNamesTheModeItWasMadeIn(t *testing.T) {
	if p := renderReport(t, root, "render", goldenMD(t, "clean.md")); !strings.Contains(p, "Sample mode: recorded answers") {
		t.Error("a sample lab's report does not say it is a sample")
	}
	if p := renderReport(t, root, "render", goldenMD(t, "clean.md"), "LAB_MODE=live via our gateway"); !strings.Contains(p, "the live review is the comment on the pull request in Gitea") {
		t.Error("a live lab's report does not say the live review is the comment")
	}
}

func TestReportRefusesWhatItCannotName(t *testing.T) {
	for _, c := range []struct{ name, mode, num string }{
		{"no mode", "", "3"},
		{"a number that is not one", "render", "3x"},
		{"a pull request the seed did not store", "render", "99"},
	} {
		cmd := exec.Command("sh", filepath.Join(root, "lab", "seed", "report.sh"), c.mode, c.num)
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "prs.tsv"), []byte("3\tdemo-discount\tdiscount\tdefect\n"), 0o600)
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "STATE_DIR=" + dir, "PAGE_DIR=" + dir, "PRS_FILE=" + filepath.Join(root, "lab", "seed", "prs.tsv"), "REPORT_TEMPLATE=" + filepath.Join(root, "lab", "page", "report.html.tmpl")}
		cmd.Stdin = strings.NewReader("**Proven: clean** — 0 findings\n")
		if out, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(out), "report: ") {
			t.Errorf("%s: not refused by name: %v %s", c.name, err, out)
		}
	}
}
