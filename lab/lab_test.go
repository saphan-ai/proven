package lab

// These tests hold the judges' lab together without Docker: the demo pull requests against the sample sets they
// are made from, the start page against the seed that fills it, the compose file against its ports, and the
// closing banner against the owner's block. The half that needs the running lab is `make demo-check`.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const root = ".."

type pullRequest struct {
	branch, fixture, class, verdict, title, says string
}

func read(t *testing.T, parts ...string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{root}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func manifest(t *testing.T) []pullRequest {
	t.Helper()
	f, err := os.Open(filepath.Join(root, "lab", "seed", "prs.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var rows []pullRequest
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		if len(p) != 6 {
			t.Fatalf("a line of prs.tsv has %d fields, not 6: %q", len(p), line)
		}
		rows = append(rows, pullRequest{p[0], p[1], p[2], p[3], p[4], p[5]})
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return rows
}

// checkManifest returns what is wrong with the demo pull requests. It is a function of the rows and the files
// they name, so a test can hand it a broken set and require it to refuse.
func checkManifest(rows []pullRequest, fixtureDir, dockerfile string) []string {
	var bad []string
	branches := map[string]bool{}
	fixtures := map[string]bool{}
	classes := map[string]string{}
	clean, defect := 0, 0
	for _, r := range rows {
		if r.branch == "" || r.fixture == "" || r.class == "" || r.title == "" || r.says == "" {
			bad = append(bad, fmt.Sprintf("%s: an empty field", r.branch))
		}
		if r.branch != "demo-"+r.fixture {
			bad = append(bad, fmt.Sprintf("%s: the workflow reads the fixture from the branch, so it must be demo-%s", r.branch, r.fixture))
		}
		switch r.verdict {
		case "clean":
			clean++
		case "defect":
			defect++
		default:
			bad = append(bad, fmt.Sprintf("%s: verdict %q is not clean or defect", r.branch, r.verdict))
		}
		if branches[r.branch] {
			bad = append(bad, fmt.Sprintf("%s: two pull requests share a branch", r.branch))
		}
		branches[r.branch] = true
		if fixtures[r.fixture] {
			bad = append(bad, fmt.Sprintf("%s: the fixture %s backs two pull requests", r.branch, r.fixture))
		}
		fixtures[r.fixture] = true
		if other, dup := classes[r.class]; dup {
			bad = append(bad, fmt.Sprintf("%s: the class %q is already shown by %s", r.branch, r.class, other))
		}
		classes[r.class] = r.branch
		for _, sub := range []string{"base", "change", "fixture.json"} {
			if _, err := os.Stat(filepath.Join(fixtureDir, r.fixture, sub)); err != nil {
				bad = append(bad, fmt.Sprintf("%s: the fixture %s has no %s", r.branch, r.fixture, sub))
			}
		}
		copyLine := "testdata/fixture/" + r.fixture + " /opt/proven/fixture/" + r.fixture
		if !strings.Contains(dockerfile, copyLine) {
			bad = append(bad, fmt.Sprintf("%s: the image does not copy the fixture (%s)", r.branch, copyLine))
		}
	}
	if clean < 2 {
		bad = append(bad, fmt.Sprintf("%d clean pull requests, the demo needs at least 2", clean))
	}
	if defect < 4 {
		bad = append(bad, fmt.Sprintf("%d defect pull requests, the demo needs at least 4", defect))
	}
	sort.Strings(bad)
	return bad
}

func TestEveryPullRequestMapsToAFixtureAndADistinctClass(t *testing.T) {
	rows := manifest(t)
	dockerfile := read(t, "Dockerfile")
	fixtureDir := filepath.Join(root, "testdata", "fixture")
	if bad := checkManifest(rows, fixtureDir, dockerfile); len(bad) > 0 {
		t.Fatalf("the demo pull requests are not whole:\n%s", strings.Join(bad, "\n"))
	}
}

// Negative control: each line below breaks one field the check guards, and the check must name it.
func TestTheManifestCheckRefusesWhatItGuards(t *testing.T) {
	good := manifest(t)
	dockerfile := read(t, "Dockerfile")
	fixtureDir := filepath.Join(root, "testdata", "fixture")
	mutate := func(f func(rows []pullRequest)) []string {
		rows := append([]pullRequest(nil), good...)
		f(rows)
		return checkManifest(rows, fixtureDir, dockerfile)
	}
	cases := []struct {
		name    string
		mutated []string
		want    string
	}{
		{"a duplicate defect class", mutate(func(r []pullRequest) { r[3].class = r[4].class }), "is already shown by"},
		{"a fixture that does not exist", mutate(func(r []pullRequest) { r[3].fixture = "no-such-fixture"; r[3].branch = "demo-no-such-fixture" }), "has no base"},
		{"a branch that does not name the fixture", mutate(func(r []pullRequest) { r[3].branch = "some-branch" }), "so it must be demo-"},
		{"a verdict that is neither", mutate(func(r []pullRequest) { r[3].verdict = "maybe" }), "is not clean or defect"},
		{"a fixture the image lacks", mutate(func(r []pullRequest) { r[3].fixture = "autofix-fixed"; r[3].branch = "demo-autofix-fixed" }), "does not copy the fixture"},
		{"a single clean pull request", mutate(func(r []pullRequest) { r[0].verdict = "defect" }), "clean pull requests"},
	}
	for _, c := range cases {
		found := false
		for _, line := range c.mutated {
			if strings.Contains(line, c.want) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: the check did not say %q; it said %v", c.name, c.want, c.mutated)
		}
	}
}

func TestEveryDefectPullRequestsFixtureReportsAndEveryCleanOneDoesNot(t *testing.T) {
	// The sample sets record what each review says. A clean pull request's fixture records no finding and
	// changes no dependency; a defect pull request's fixture changes a dependency or records a finding.
	for _, r := range manifest(t) {
		fx := read(t, "testdata", "fixture", r.fixture, "fixture.json")
		recordsFinding := regexp.MustCompile(`"planner_response":\s*\[\s*\{`).MatchString(fx)
		changesDependency := strings.Contains(read(t, "testdata", "fixture", r.fixture, "change", "go.mod"), "\nrequire ")
		switch r.verdict {
		case "clean":
			if recordsFinding || changesDependency {
				t.Errorf("%s is called clean, but fixture %s records a finding or adds a dependency", r.branch, r.fixture)
			}
		case "defect":
			if !recordsFinding && !changesDependency {
				t.Errorf("%s is called a defect, but fixture %s records no finding and adds no dependency", r.branch, r.fixture)
			}
		}
	}
}

func TestTheWorkflowPicksTheFixtureByBranchAndSaysCleanWhenItKeepsNothing(t *testing.T) {
	wf := read(t, "lab", "workflow", "proven.yml")
	for _, want := range []string{
		"demo-*) fixture=",                  // the branch names the fixture
		`-e FIXTURE="$fixture"`,             // and reaches the container
		"|| d=/opt/proven/fixture/discount", // any other branch is the discount sample
		"clean: Proven kept no finding",     // a review with no finding says so
		"grep -q '^finding ' review.txt",
	} {
		if !strings.Contains(wf, want) {
			t.Errorf("lab/workflow/proven.yml lacks %q", want)
		}
	}
	// The seed rewrites the file with sed patterns that must still find their text.
	seed := read(t, "lab", "seed", "seed.sh")
	for _, pattern := range []string{"docker run --rm ", "proven scan --sample "} {
		if strings.Count(wf, pattern) != 1 {
			t.Errorf("the seed rewrites %q in the workflow, and the workflow must hold it exactly once; it holds it %d times", pattern, strings.Count(wf, pattern))
		}
		if !strings.Contains(seed, pattern) {
			t.Errorf("the seed no longer rewrites %q", pattern)
		}
	}
}

func placeholders(s string) map[string]bool {
	out := map[string]bool{}
	for _, m := range regexp.MustCompile(`@@[A-Z_]+@@`).FindAllString(s, -1) {
		out[m] = true
	}
	return out
}

func TestTheStartPageTemplate(t *testing.T) {
	tmpl := read(t, "lab", "page", "lab.html.tmpl")
	render := read(t, "lab", "seed", "render.sh")
	// Every placeholder is one render.sh fills (the rows are filled by the awk in the same file).
	for p := range placeholders(tmpl) {
		if !strings.Contains(render, p) {
			t.Errorf("the template holds %s and render.sh does not fill it", p)
		}
	}
	if !placeholders(tmpl)["@@LAB_PASSWORD@@"] {
		t.Error("the template has no @@LAB_PASSWORD@@ to fill")
	}
	// Every link to another address opens in a new tab and says noopener.
	for _, a := range regexp.MustCompile(`<a [^>]*href="(?:http|@@)[^>]*>`).FindAllString(tmpl, -1) {
		if !strings.Contains(a, `target="_blank"`) || !strings.Contains(a, `rel="noopener"`) {
			t.Errorf("a link that does not open in a new tab with noopener: %s", a)
		}
	}
	// A URL is a setting, so the template names none of the lab's own addresses.
	for _, banned := range []string{"127.0.0.1", "localhost", ":3000", ":3001", ":3002", ":3003"} {
		if strings.Contains(tmpl, banned) {
			t.Errorf("the template hardcodes %q; the address belongs in lab/versions.env", banned)
		}
	}
	// No web fonts and no outside script: the lab works offline.
	for _, banned := range []string{"fonts.googleapis", "<script src", "@import"} {
		if strings.Contains(tmpl, banned) {
			t.Errorf("the template needs the network: %q", banned)
		}
	}
}

func TestTheStartPageIsRenderedWithTheLoginAndNothingLeaks(t *testing.T) {
	dir := t.TempDir()
	rows := filepath.Join(dir, "rows.html")
	row := `    <tr><td class="n">7</td><td><a href="http://127.0.0.1:3000/proven/discount/pulls/7" target="_blank" rel="noopener">A real pull request</a></td><td class="muted">Says something.</td><td><span class="v ok">clean</span></td></tr>` + "\n"
	if err := os.WriteFile(rows, []byte(row), 0o600); err != nil {
		t.Fatal(err)
	}
	pageDir := filepath.Join(dir, "page")
	if err := os.Mkdir(pageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A password with every character the two escapes exist for.
	password := `p&ss|w\rd<b>"q"`
	script := `. "$ROOT/lab/seed/render.sh"; render_page`
	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(),
		"ROOT="+mustAbs(t, root),
		"PAGE_TEMPLATE="+mustAbs(t, filepath.Join(root, "lab", "page", "lab.html.tmpl")),
		"PAGE_DIR="+pageDir,
		"PR_ROWS_FILE="+rows,
		"PAGE_WORK="+filepath.Join(dir, "work.html"),
		"CLEAN_COUNT=2",
		"LAB_URL=http://127.0.0.1:3000/",
		"LAB_ADMIN_USER=proven",
		"LAB_DEMO_REPO=discount",
		"LAB_ADMIN_PASSWORD="+password,
		"LAB_GRAFANA_URL=http://127.0.0.1:3001/d/proven-review",
		"LAB_DOCS_URL=http://127.0.0.1:3003/",
		"LAB_PAGE_URL=http://127.0.0.1:3002/lab",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("render_page: %v\n%s", err, out)
	}
	if strings.Contains(string(out), password) {
		t.Error("render_page printed the password")
	}
	page := string(mustRead(t, filepath.Join(pageDir, "lab.html")))
	wantCode := `<code id="p">p&amp;ss|w\rd&lt;b&gt;&quot;q&quot;</code>`
	if !strings.Contains(page, wantCode) {
		t.Errorf("the page does not carry the escaped password as %s", wantCode)
	}
	for _, want := range []string{
		`>A real pull request</a>`,
		`href="http://127.0.0.1:3000/proven/discount/pulls?state=all"`,
		`href="http://127.0.0.1:3001/d/proven-review"`,
		`href="http://127.0.0.1:3003/"`,
		`served by the lab at http://127.0.0.1:3002/lab`,
		`Two are clean changes`,
		`<code id="u">proven</code>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the rendered page lacks %s", want)
		}
	}
	if strings.Contains(page, "@@") {
		t.Error("the rendered page still holds a placeholder")
	}
	entries, err := os.ReadDir(pageDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "lab.html" {
		t.Errorf("the page directory holds %d entries; only lab.html belongs there", len(entries))
	}
	// The password is in the served copy and in no committed file.
	for _, f := range []string{"lab/page/lab.html.tmpl", "lab/versions.env", "lab/compose.yaml", "lab/seed/seed.sh", "lab/seed/prs.tsv"} {
		if strings.Contains(read(t, f), password) {
			t.Errorf("%s holds the test password", f)
		}
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestEveryPublishedPortBindsToTheLabAddress(t *testing.T) {
	compose := read(t, "lab", "compose.yaml")
	inPorts := false
	published := 0
	for _, line := range strings.Split(compose, "\n") {
		trim := strings.TrimSpace(line)
		switch {
		case trim == "ports:":
			inPorts = true
		case inPorts && strings.HasPrefix(trim, "- "):
			published++
			if !strings.HasPrefix(trim, `- "${LAB_BIND}:`) {
				t.Errorf("a published port is not bound to LAB_BIND: %s", trim)
			}
		case inPorts:
			inPorts = false
		}
	}
	// gitea, lgtm (two), page and docs.
	if published != 5 {
		t.Errorf("the compose file publishes %d ports, expected 5 (gitea, grafana, otlp, page, docs)", published)
	}
	for _, want := range []string{
		`"${LAB_BIND}:${LAB_PAGE_PORT}:${LAB_PAGE_CONTAINER_PORT}"`,
		`"${LAB_BIND}:${LAB_DOCS_PORT}:${LAB_DOCS_CONTAINER_PORT}"`,
		"read_only: true",
		"no-new-privileges:true",
		"lab-page:/usr/share/nginx/html:ro",
		"lab-page:/page",
	} {
		if !strings.Contains(compose, want) {
			t.Errorf("lab/compose.yaml lacks %s", want)
		}
	}
	if !strings.Contains(read(t, "docs-site", "Dockerfile"), "USER 10001:10001") {
		t.Error("the docs image does not run as a user that is not root")
	}
}

func TestTheLabSettingsAreInVersionsEnv(t *testing.T) {
	env := read(t, "lab", "versions.env")
	for _, key := range []string{
		"LAB_PAGE_PORT=3002", "LAB_PAGE_PATH=/lab", "LAB_PAGE_IMAGE=", "LAB_PAGE_IMAGE_VERSION=", "LAB_PAGE_CONTAINER_PORT=",
		"LAB_DOCS_PORT=3003", "LAB_DOCS_CONTAINER_PORT=", "DOCS_IMAGE=", "DOCS_IMAGE_TAG=", "NODE_IMAGE=",
		"GITEA_HTTP_PORT=3000", "LGTM_GRAFANA_PORT=3001", "LGTM_DASHBOARD_UID=proven-review", "LAB_DEMO_HEAD=demo-discount",
	} {
		if !strings.Contains(env, key) {
			t.Errorf("lab/versions.env lacks %s", key)
		}
	}
}

// The banner's own check runs the exact lines, the settings and the redirected case. Its terminal case needs a
// pseudo-terminal; where the host gives none the check names it as not run, and so does this test.
func TestTheBannerCheck(t *testing.T) {
	out, err := exec.Command("sh", filepath.Join(root, "lab", "scripts", "banner-check.sh"), "--no-tty-ok").CombinedOutput()
	if err != nil {
		t.Fatalf("banner-check: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "terminal case NOT RUN") {
		t.Log("the terminal case of the banner check was NOT RUN: this host gives no pseudo-terminal")
	}
}
