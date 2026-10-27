package lab

// The report directory is read by the page's nginx, a different user from the seed's. Round 2 of the lab found it 700
// (a 403 on every Report link) and the page rendered from a CLI that had no --md. These tests hold both.

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runReportUnderUmask runs the report script of tree under the given umask and returns the reports directory it wrote.
func runReportUnderUmask(t *testing.T, tree, umask, md string) string {
	t.Helper()
	dir := t.TempDir()
	state, page := filepath.Join(dir, "state"), filepath.Join(dir, "page")
	for _, d := range []string{state, page} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(state, "prs.tsv"), []byte("3\tdemo-discount\tdiscount\tdefect\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", "umask "+umask+`; exec sh "$0" render 3`, filepath.Join(tree, "lab", "seed", "report.sh"))
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "STATE_DIR=" + state, "PAGE_DIR=" + page,
		"PRS_FILE=" + filepath.Join(tree, "lab", "seed", "prs.tsv"),
		"REPORT_TEMPLATE=" + filepath.Join(tree, "lab", "page", "report.html.tmpl"), "LAB_URL=http://127.0.0.1:3300/"}
	cmd.Stdin = strings.NewReader(md)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("report.sh under umask %s: %v\n%s", umask, err, out)
	}
	return filepath.Join(page, "reports")
}

// worldReadable says whether a user other than the owner can list the directory and read the page: what the page's nginx
// (another uid) needs, and what a 700 directory refused with a 403.
func worldReadable(t *testing.T, reports string) bool {
	t.Helper()
	d, err := os.Stat(reports)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Stat(filepath.Join(reports, "3.html"))
	if err != nil {
		t.Fatal(err)
	}
	return d.Mode().Perm()&0o005 == 0o005 && f.Mode().Perm()&0o004 == 0o004
}

// reportTreeWithout copies the report script and what it reads into a temporary tree, dropping every line of the script
// that contains one of the given words.
func reportTreeWithout(t *testing.T, words ...string) string {
	t.Helper()
	tree := t.TempDir()
	for _, f := range []string{"seed/report.sh", "seed/render.sh", "seed/md2html.awk", "seed/prs.tsv", "page/report.html.tmpl"} {
		b, err := os.ReadFile(filepath.Join(root, "lab", filepath.FromSlash(f)))
		if err != nil {
			t.Fatal(err)
		}
		if f == "seed/report.sh" {
			var keep []string
			for _, l := range strings.Split(string(b), "\n") {
				drop := false
				for _, w := range words {
					drop = drop || strings.Contains(l, w)
				}
				if !drop {
					keep = append(keep, l)
				}
			}
			b = []byte(strings.Join(keep, "\n"))
		}
		dst := filepath.Join(tree, "lab", filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return tree
}

func TestReportsDirectoryIsWorldReadableWhateverTheUmask(t *testing.T) {
	for _, umask := range []string{"022", "077"} {
		if !worldReadable(t, runReportUnderUmask(t, root, umask, goldenMD(t, "clean.md"))) {
			t.Errorf("under umask %s the reports directory or the page cannot be read by another user: nginx would answer 403", umask)
		}
	}
	// The negative control. Mutated field: the umask and the chmod of the reports directory, removed from the script. Under
	// umask 077 the directory is 700 again and the check must turn red.
	mut := reportTreeWithout(t, "umask 022", "chmod 755")
	if worldReadable(t, runReportUnderUmask(t, mut, "077", goldenMD(t, "clean.md"))) {
		t.Error("the control did not turn red: a 700 reports directory passed as readable")
	}
}

func TestReportPageRendersTheReviewTheCLIWritesNow(t *testing.T) {
	// One review's real --md output (the sample change, recorded answers, no network) goes through the page: the verdict
	// badge, the findings table and the folded details must come out of what the CLI writes today, not of a golden copy.
	needTool(t, "go")
	cmd := exec.Command("go", "run", "./cmd/proven", "scan", "--sample", "testdata/fixture/discount", "--md")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PROVEN_LOG_DIR="+t.TempDir())
	md, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("proven scan --md: %v", err)
	}
	if !strings.Contains(string(md), "Proven: ") {
		t.Fatalf("proven scan --md wrote no review (%v): %q", err, md)
	}
	page := renderReport(t, root, "render", string(md))
	for _, w := range []string{`<span class="badge bad">Blocked</span>`, `<th>Severity</th>`, `<details><summary>`, `class="markdown-body"`} {
		if !strings.Contains(page, w) {
			t.Errorf("the page lacks %q for the review the CLI writes now", w)
		}
	}
	if strings.Contains(page, "@@") {
		t.Error("a template field was left unfilled")
	}
}
