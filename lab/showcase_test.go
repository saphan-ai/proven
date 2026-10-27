package lab

// The start page as a showcase (the owner, 27 Oct): scenarios each with a "Try yourself", and a security section whose claims
// link to the live proof. What is held here: every button names something that exists, every outside link is a public page of
// ours, nothing private is in the page, and the two scenarios that run a command say what they run. A scenario's command is
// run for real only on a machine with Docker; make try is run there, not here.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/testrun"
)

var (
	prTokenRe   = regexp.MustCompile(`@@PR_([A-Za-z0-9._-]+)@@`)
	hrefRe      = regexp.MustCompile(`href="(https?://[^"]+)"`)
	publicHosts = []string{"https://github.com/saphan-ai/proven", "https://hub.docker.com/r/saphan/proven", "https://build.nvidia.com", "@@"}
)

func TestEveryScenarioButtonOpensAPullRequestTheLabSeeds(t *testing.T) {
	tmpl := read(t, "lab", "page", "lab.html.tmpl")
	fixtures := map[string]bool{}
	for _, r := range manifest(t) {
		fixtures[r.fixture] = true
	}
	var tokens [][]string
	for _, m := range prTokenRe.FindAllStringSubmatch(tmpl, -1) {
		if m[1] != "ROWS" { // the table's rows, not a button
			tokens = append(tokens, m)
		}
	}
	if len(tokens) < 3 {
		t.Fatalf("the page has %d scenario buttons that open a pull request, want the three the owner named", len(tokens))
	}
	for _, m := range tokens {
		if !fixtures[m[1]] {
			t.Errorf("a scenario button opens the pull request of %q, which lab/seed/prs.tsv does not seed", m[1])
		}
	}
	// The negative control: a token for a fixture the lab does not seed is refused when the page is rendered (mutated field:
	// the fixture name in a token).
	tree := copyLab(t)
	copyInto(t, tree, "lab/seed")
	copyInto(t, tree, "lab/page")
	mutate(t, tree, "lab/page/lab.html.tmpl", "@@PR_discount@@", "@@PR_no-such-change@@")
	dir := t.TempDir()
	pageDir := filepath.Join(dir, "page")
	if err := os.Mkdir(pageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	rows := filepath.Join(dir, "rows.html")
	if err := os.WriteFile(rows, []byte("<tr></tr>\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/render.sh"; render_page`)
	cmd.Env = append(os.Environ(), "ROOT="+tree, "PAGE_TEMPLATE="+filepath.Join(tree, "lab", "page", "lab.html.tmpl"), "PAGE_DIR="+pageDir,
		"PR_ROWS_FILE="+rows, "PR_NUMBERS_FILE="+writePRNumbers(t, dir), "TMPDIR="+dir, "CLEAN_COUNT=2", "LAB_URL=http://127.0.0.1:3300/",
		"LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "LAB_ADMIN_PASSWORD=pw-0123456789")
	if out, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(out), "the start page still holds a placeholder") {
		t.Errorf("the control did not turn red: a button for a change nobody seeds was rendered: %v\n%s", err, out)
	}
}

func TestEveryOutsideLinkOnThePagesIsAPublicPageOfOurs(t *testing.T) {
	for _, f := range []string{"lab/page/lab.html.tmpl", "lab/page/report.html.tmpl"} {
		tmpl := read(t, strings.Split(f, "/")...)
		for _, m := range hrefRe.FindAllStringSubmatch(tmpl, -1) {
			ok := false
			for _, h := range publicHosts {
				ok = ok || strings.HasPrefix(m[1], h)
			}
			if !ok {
				t.Errorf("%s links %s, which is not a public page of ours", f, m[1])
			}
		}
	}
}

// The blackout: no page, and no image beside it, carries a path, an address, a login or a machine name of ours.
func TestThePagesCarryNoPrivatePathAddressLoginOrMachineName(t *testing.T) {
	private := regexp.MustCompile(`(/home/|/Users/|/root/|/var/folders|C:\\Users|\b(?:10|192\.168|172\.(?:1[6-9]|2\d|3[01]))\.\d+\.\d+(?:\.\d+)?\b|\.local\b|\.saphan|macm5|code-fleet)`)
	files, err := filepath.Glob(filepath.Join(root, "lab", "page", "*"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no page files: %v", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if m := private.Find(b); m != nil {
			t.Errorf("%s carries %q", filepath.Base(f), m)
		}
		// An image, if one is added, is held to the same list by its name: a screenshot is a file in this folder and is
		// opened full size or links to the live page it shows; its name must not say where it was taken.
		if ext := strings.ToLower(filepath.Ext(f)); ext == ".png" || ext == ".jpg" {
			if private.MatchString(filepath.Base(f)) {
				t.Errorf("the image name %s names a private place", filepath.Base(f))
			}
		}
	}
	// The negative control: a path in the page is seen (mutated field: a line of the template).
	if !private.MatchString("see /home/someone/work/proven") {
		t.Error("the control did not turn red: a home path passed the blackout check")
	}
}

func TestTryNamesTheScenariosItRunsAndRefusesAnyOther(t *testing.T) {
	cmd := exec.Command("sh", filepath.Join(root, "lab", "scripts", "try.sh"), "nonsense")
	cmd.Env = append(os.Environ(), "ROOT="+mustAbs(t, root))
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "it must be forged or prover. NEXT") {
		t.Errorf("try.sh did not refuse a scenario it does not know, by name: %v\n%s", err, out)
	}
	script := read(t, "lab", "scripts", "try.sh")
	// The page names exactly the two commands the script runs.
	page := read(t, "lab", "page", "lab.html.tmpl")
	for _, s := range []string{"forged", "prover"} {
		if !strings.Contains(page, "make try SCENARIO="+s) || !strings.Contains(script, "try_"+s+"()") {
			t.Errorf("the page and the script do not agree on the scenario %q", s)
		}
	}
}

func TestTheProverScenarioRunsWithTheFlagsTheLiveProverUses(t *testing.T) {
	script := read(t, "lab", "scripts", "try.sh")
	cfg := testrun.DefaultContainerConfig()
	for _, want := range []string{"--network none", "--cap-drop ALL", "--read-only", "--cpus " + cfg.CPUs, "--memory " + cfg.Memory, "no-new-privileges"} {
		if !strings.Contains(script, want) {
			t.Errorf("the Prover scenario lacks %q, which the live Prover runs with", want)
		}
	}
	// No secret goes in, and the script does not carry one: it passes no -e of its own in the Prover scenario.
	prover := script[strings.Index(script, "try_prover()"):]
	if strings.Contains(prover, " -e ") {
		t.Error("the Prover scenario passes an environment variable in")
	}
}

func TestTheForgedScenarioChangesOneByteAndExpectsARefusal(t *testing.T) {
	script := read(t, "lab", "scripts", "try.sh")
	for _, want := range []string{"proven scan --evidence-pack", "proven verify /pack/review.pack", "conv=notrunc", "proven verify /pack/forged.pack", "the forged pack was accepted"} {
		if !strings.Contains(script, want) {
			t.Errorf("the forged scenario lacks %q", want)
		}
	}
}
