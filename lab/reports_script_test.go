package lab

// make reports (lab/scripts/reports.sh): for each pull request the seed stored, the seed image reads back the review Proven
// posted on it (report.sh comment) and turns it into the page (report.sh render). No review is made for the page. The
// stand-in docker answers each of the three calls the script makes, runs the real report.sh for the last one, and records
// every call. COMMENT_STATUS makes the comment call answer as Gitea would with no review posted yet (3).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const reportsDocker = `#!/bin/sh
echo "$*" >> "$FAKE_DIR/docker.calls"
case " $* " in
  *"report.sh comment"*)
    # A refusing reader that still printed something: only reports.sh's status check keeps it off a page.
    cat "$MD_FILE"
    exit "${COMMENT_STATUS:-0}"
    ;;
  *" seed /state/prs.tsv "*)
    printf '3\tdemo-discount\tdiscount\tdefect\n4\tdemo-discount-clean\tdiscount-clean\tclean\n'
    exit 0
    ;;
  *"report.sh render"*)
    for last; do :; done
    exec sh "$REPORT_SH" render "$last"
    ;;
esac
exit 0
`

func runReports(t *testing.T, tree string, extra ...string) (out string, calls string, page string, err error) {
	t.Helper()
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	pageDir := filepath.Join(dir, "page")
	for _, d := range []string{state, pageDir} {
		if e := os.MkdirAll(d, 0o755); e != nil {
			t.Fatal(e)
		}
	}
	if e := os.WriteFile(filepath.Join(state, "prs.tsv"), []byte("3\tdemo-discount\tdiscount\tdefect\n4\tdemo-discount-clean\tdiscount-clean\tclean\n"), 0o600); e != nil {
		t.Fatal(e)
	}
	bin := filepath.Join(dir, "bin")
	writeExec(t, bin, "docker", reportsDocker)
	cmd := exec.Command("sh", filepath.Join(tree, "lab", "scripts", "reports.sh"))
	cmd.Env = append([]string{
		"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + dir, "TMPDIR=" + dir, "FAKE_DIR=" + dir,
		"MD_FILE=" + mustAbs(t, filepath.Join("testdata", "report", "findings.md")),
		"REPORT_SH=" + filepath.Join(tree, "lab", "seed", "report.sh"),
		"STATE_DIR=" + state, "PAGE_DIR=" + pageDir, "PRS_FILE=" + filepath.Join(tree, "lab", "seed", "prs.tsv"),
		"REPORT_TEMPLATE=" + filepath.Join(tree, "lab", "page", "report.html.tmpl"),
		"LAB_URL=http://127.0.0.1:3300/", "LAB_PAGE_URL=http://127.0.0.1:3302/lab", "LAB_MODE=sample",
		// A credential in the environment of the run: it must reach no docker call.
		"PROVEN_GENERATOR_KEY=" + reportTokenSentinel, "LAB_ADMIN_PASSWORD=" + reportPasswordSentinel,
	}, extra...)
	b, err := cmd.CombinedOutput()
	c, _ := os.ReadFile(filepath.Join(dir, "docker.calls"))
	p, _ := os.ReadFile(filepath.Join(pageDir, "reports", "3.html"))
	return string(b), string(c), string(p), err
}

func TestReportsBuildsEachPageFromTheCommentPostedOnThePullRequest(t *testing.T) {
	tree := mustAbs(t, root)
	out, calls, page, err := runReports(t, tree)
	if err != nil {
		t.Fatalf("make reports failed: %v\n%s", err, out)
	}
	if !strings.Contains(page, `<span class="badge bad">Blocked</span>`) || !strings.Contains(out, "reports: 2 report pages built") {
		t.Errorf("the page was not built from the comment:\n%s\n%s", out, page)
	}
	if !strings.Contains(calls, "report.sh comment") {
		t.Errorf("the comment posted on the pull request was not read:\n%s", calls)
	}
	// The page is the comment, never a review made apart from it: no Proven run at all.
	if strings.Contains(calls, "proven scan") || strings.Contains(calls, "SAMPLE_MODE") {
		t.Errorf("a review was made for the page instead of reading the comment:\n%s", calls)
	}
	for _, secret := range []string{reportTokenSentinel, reportPasswordSentinel} {
		if strings.Contains(calls, secret) || strings.Contains(out, secret) {
			t.Errorf("a credential reached a docker call or the output")
		}
	}
}

func TestReportsWaitsForAReviewNotPostedYetAndNamesIt(t *testing.T) {
	// The unplug of the review: the job has not posted its comment. No page may be built in its place.
	tree := mustAbs(t, root)
	out, _, page, err := runReports(t, tree, "COMMENT_STATUS=3", "LAB_REPORT_WAIT_SECONDS=0")
	if err == nil || !strings.Contains(out, "Proven has not posted its review yet") || page != "" {
		t.Errorf("a review not posted was not named, or a page was built without it: %v\n%s\n%s", err, out, page)
	}
	// The negative control: reports.sh renders whatever the comment call answered. Mutated field: the status check.
	mut := copyLab(t)
	copyInto(t, mut, "lab/seed", "lab/page")
	mutate(t, mut, "lab/scripts/reports.sh", `if [ "$status" -eq 0 ] &&`, `if [ "$status" -ne 99 ] &&`)
	_, _, mpage, _ := runReports(t, mut, "COMMENT_STATUS=3", "LAB_REPORT_WAIT_SECONDS=0")
	if mpage == "" {
		t.Error("the control did not turn red: with the status unchecked no page was built either")
	}
}

// commentCurl answers report.sh's one read, the pull request's comments, as Gitea does: JSON, with < and > escaped.
const commentCurl = `#!/bin/sh
echo "$*" >> "$FAKE_DIR/curl.calls"
cat "$COMMENTS_JSON"
`

func TestReportCommentIsTheBotsMarkedCommentDecoded(t *testing.T) {
	needTool(t, "jq")
	dir := t.TempDir()
	comments := `[{"id":1,"user":{"login":"proven"},"body":"<!-- proven-review run=1 sha=a -->\nwritten by a person"},` +
		`{"id":2,"user":{"login":"proven-bot"},"body":"a note without the marker"},` +
		`{"id":3,"user":{"login":"proven-bot"},"body":"<!-- proven-review run=4 sha=b -->\n**Proven: blocked**\n\u003cdetails\u003e 1\\.2.3 \"q\"\u003c/details\u003e"}]`
	cj := filepath.Join(dir, "comments.json")
	if err := os.WriteFile(cj, []byte(comments), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "bin")
	writeExec(t, bin, "curl", commentCurl)
	run := func(bot string) (string, error) {
		cmd := exec.Command("sh", filepath.Join(mustAbs(t, root), "lab", "seed", "report.sh"), "comment", "3")
		cmd.Env = []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + dir, "TMPDIR=" + dir, "FAKE_DIR=" + dir,
			"COMMENTS_JSON=" + cj, "GITEA_INTERNAL_URL=http://gitea:3000", "LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount",
			"LAB_BOT_USER=" + bot, "STATE_DIR=" + dir, "PAGE_DIR=" + dir}
		b, err := cmd.Output()
		return string(b), err
	}
	got, err := run("proven-bot")
	want := "**Proven: blocked**\n<details> 1\\.2.3 \"q\"</details>\n"
	if err != nil || got != want {
		t.Errorf("the comment read is %q (%v), want %q", got, err, want)
	}
	calls, _ := os.ReadFile(filepath.Join(dir, "curl.calls"))
	if !strings.Contains(string(calls), "/api/v1/repos/proven/discount/issues/3/comments") {
		t.Errorf("not the pull request's comments: %s", calls)
	}
	// The control: a bot that wrote no marked comment gets no review, and says so with exit 3.
	if out, err := run("someone-else"); out != "" || err == nil {
		t.Errorf("the control did not turn red: another author's review was taken: %q %v", out, err)
	} else if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 3 {
		t.Errorf("no review posted must exit 3: %v", err)
	}
}

func TestReportsNamesAPageThatWasNotBuiltAndFails(t *testing.T) {
	// A report.sh that refuses every pull request: the mutated field is the lookup of the branch in the seed's list.
	bad := copyLab(t)
	copyInto(t, bad, "lab/seed", "lab/page")
	mutate(t, bad, "lab/seed/report.sh", `[ -n "$branch" ] || fail "pull request $num is not in $STATE_DIR/prs.tsv"`, `fail "pull request $num is not in $STATE_DIR/prs.tsv"`)
	out, _, _, err := runReports(t, bad)
	if err == nil || !strings.Contains(out, "the report page was not built") {
		t.Errorf("a refused page was not named, or the run did not fail: %v\n%s", err, out)
	}
}
