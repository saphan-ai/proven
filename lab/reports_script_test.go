package lab

// make reports (lab/scripts/reports.sh): for each pull request the seed stored, Proven's image writes the review as
// Markdown, with no network and no key, and the seed image's report.sh turns it into the page. The stand-in docker answers
// each of the three calls the script makes, runs the real report.sh for the last one, and records every call.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const reportsDocker = `#!/bin/sh
echo "$*" >> "$FAKE_DIR/docker.calls"
case "$1" in
  run)
    cat "$MD_FILE"
    exit 0
    ;;
esac
case " $* " in
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

func TestReportsBuildsOnePagePerPullRequestWithNoNetworkAndNoKey(t *testing.T) {
	tree := mustAbs(t, root)
	out, calls, page, err := runReports(t, tree)
	if err != nil {
		t.Fatalf("make reports failed: %v\n%s", err, out)
	}
	if !strings.Contains(page, `<span class="badge bad">Blocked</span>`) || !strings.Contains(out, "reports: 2 report pages built") {
		t.Errorf("the page was not built from the review:\n%s\n%s", out, page)
	}
	if !strings.Contains(calls, "run --rm --network none --cap-drop ALL") || !strings.Contains(calls, "proven scan --md ") {
		t.Errorf("Proven was not run with no network and with --md:\n%s", calls)
	}
	for _, secret := range []string{reportTokenSentinel, reportPasswordSentinel} {
		if strings.Contains(calls, secret) || strings.Contains(out, secret) {
			t.Errorf("a credential reached a docker call or the output")
		}
	}
	if strings.Contains(calls, "PROVEN_GENERATOR_KEY") || strings.Contains(calls, "METERLESS_RUNTIME_TOKEN") {
		t.Errorf("a key was named to a container:\n%s", calls)
	}
	// The negative control: the run gets the network back. Mutated field: --network none in reports.sh.
	mut := copyLab(t)
	copyInto(t, mut, "lab/seed", "lab/page")
	mutate(t, mut, "lab/scripts/reports.sh", "--network none", "--network bridge")
	_, mcalls, _, _ := runReports(t, mut)
	if strings.Contains(mcalls, "--network none") {
		t.Error("the control did not turn red: with the flag changed the run still had no network")
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
