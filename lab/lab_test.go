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
	// The seed's sed writes the pattern with its quote escaped, so the two spellings are listed apart.
	for _, p := range []struct{ inWorkflow, inSeed string }{
		{"docker run --rm ", "docker run --rm "},
		{`proven scan "`, `proven scan \"`},
		{"GATEWAY_ZETARIS_PATH: /zetaris/api/v1.0", "GATEWAY_ZETARIS_PATH: /zetaris/api/v1.0"},
	} {
		if strings.Count(wf, p.inWorkflow) != 1 {
			t.Errorf("the seed rewrites %q in the workflow, and the workflow must hold it exactly once; it holds it %d times", p.inWorkflow, strings.Count(wf, p.inWorkflow))
		}
		if !strings.Contains(seed, p.inSeed) {
			t.Errorf("the seed no longer rewrites %q", p.inSeed)
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
	for _, banned := range []string{"127.0.0.1", "localhost", ":3300", ":3301", ":3302", ":3303"} {
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

// writePRNumbers writes the seed's list of pull requests (number, branch, fixture, verdict) into dir, numbering the manifest
// from 10, and returns its path: the start page's scenarios link each fixture to its number.
func writePRNumbers(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "pr-numbers.tsv")
	var b strings.Builder
	for i, r := range manifest(t) {
		fmt.Fprintf(&b, "%d\t%s\t%s\t%s\n", 10+i, r.branch, r.fixture, r.verdict)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTheStartPageIsRenderedWithTheLoginAndNothingLeaks(t *testing.T) {
	dir := t.TempDir()
	rows := filepath.Join(dir, "rows.html")
	row := `    <tr><td class="n">7</td><td><a href="http://127.0.0.1:3300/proven/discount/pulls/7" target="_blank" rel="noopener">A real pull request</a></td><td class="muted">Says something.</td><td><span class="v ok">clean</span></td></tr>` + "\n"
	if err := os.WriteFile(rows, []byte(row), 0o600); err != nil {
		t.Fatal(err)
	}
	// The seed's list of pull requests (number, branch, fixture, verdict): the scenarios link each fixture to its number.
	numbers := writePRNumbers(t, dir)
	pageDir := filepath.Join(dir, "page")
	tmp := filepath.Join(dir, "tmp")
	for _, d := range []string{pageDir, tmp} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A password with every character the two escapes exist for.
	password := `p&ss|w\rd<b>"q"`
	script := `. "$ROOT/lab/seed/render.sh"; render_page`
	// A stand-in awk, first on the path, looks at where the page is being staged at the moment the password is about
	// to flow through the pipe, then runs the real awk. The review's finding: no staging file outside the volume.
	probe := filepath.Join(dir, "probe")
	if err := os.Mkdir(probe, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeAwk := "#!/bin/sh\nls -A \"$PAGE_DIR\" > \"$PROBE/during.page\"\nls -lA \"$PAGE_DIR\" > \"$PROBE/during.page.long\"\nls -A \"$TMPDIR\" > \"$PROBE/during.tmp\"\nPATH=\"$REAL_PATH\" exec awk \"$@\"\n"
	if err := os.Mkdir(filepath.Join(dir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "awk"), []byte(fakeAwk), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(),
		"ROOT="+mustAbs(t, root),
		"PAGE_TEMPLATE="+mustAbs(t, filepath.Join(root, "lab", "page", "lab.html.tmpl")),
		"PAGE_DIR="+pageDir,
		"PR_ROWS_FILE="+rows, "PR_NUMBERS_FILE="+numbers,
		// Every temp file the renderer or its tools make lands in tmp, so a copy outside the page volume is seen.
		"TMPDIR="+tmp,
		"PATH="+filepath.Join(dir, "bin")+":"+os.Getenv("PATH"), "REAL_PATH="+os.Getenv("PATH"), "PROBE="+probe,
		"LAB_LIVE_STATE=keys stored: PROVEN_GENERATOR_KEY",
		"LAB_MODE=live via our gateway",
		"CLEAN_COUNT=2",
		"LAB_URL=http://127.0.0.1:3300/",
		"LAB_ADMIN_USER=proven",
		"LAB_DEMO_REPO=discount",
		"LAB_ADMIN_PASSWORD="+password,
		"LAB_GRAFANA_URL=http://127.0.0.1:3301/d/proven-review",
		"LAB_DOCS_URL=http://127.0.0.1:3303/",
		"LAB_HELP_URL=https://example.test/help",
		"LAB_PAGE_URL=http://127.0.0.1:3302/lab",
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
		`href="http://127.0.0.1:3300/proven/discount/pulls?state=all"`,
		`href="http://127.0.0.1:3301/d/proven-review"`,
		`href="http://127.0.0.1:3303/"`,
		`href="https://example.test/help"`,
		`href="http://127.0.0.1:3301/d/proven-evidence"`,
		`make -C &lt;path to the repository&gt; demo-check`,
		`served by the lab at http://127.0.0.1:3302/lab`,
		`Two are clean changes`,
		`Mode: <b>live via our gateway</b>.`,
		`<code id="u">proven</code>`,
		// A scenario button opens the report of the pull request that carries its change.
		`href="http://127.0.0.1:3302/lab/report/10.html"`,
		`make try SCENARIO=forged`,
		`make try SCENARIO=prover`,
		`https://hub.docker.com/r/saphan/proven`,
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
	// While the password was flowing, the page was being staged inside the volume, readable by its owner only, and
	// the temp directory was empty.
	if during := string(mustRead(t, filepath.Join(probe, "during.page"))); !strings.Contains(during, ".lab.html.new") {
		t.Errorf("the page was not staged in the page volume; the volume held: %q", during)
	}
	if long := string(mustRead(t, filepath.Join(probe, "during.page.long"))); !strings.Contains(long, "-rw-------") {
		t.Errorf("the staging file was not mode 0600: %s", long)
	}
	if during := strings.TrimSpace(string(mustRead(t, filepath.Join(probe, "during.tmp")))); during != "" {
		t.Errorf("the temp directory held %q while the page was rendered", during)
	}
	// The password-bearing page exists in the page volume and nowhere else: not in the temp directory the renderer and
	// its tools used, not beside the volume. (Review of round 1, BLOCKING 2.)
	for _, needle := range []string{password, `p&amp;ss|w\rd&lt;b&gt;&quot;q&quot;`} {
		for _, hit := range walkFor(t, dir, needle) {
			if !strings.HasPrefix(hit, pageDir+string(filepath.Separator)) {
				t.Errorf("a copy of the page's password is outside the page volume: %s", hit)
			}
		}
	}
	if info, err := os.Stat(filepath.Join(pageDir, "lab.html")); err != nil || info.Mode().Perm() != 0o644 {
		t.Errorf("the served page is not mode 0644 (the page service reads it as another user): %v %v", info, err)
	}
	// The password is in the served copy and in no committed file.
	for _, f := range []string{"lab/page/lab.html.tmpl", "lab/versions.env", "lab/compose.yaml", "lab/seed/seed.sh", "lab/seed/prs.tsv"} {
		if strings.Contains(read(t, f), password) {
			t.Errorf("%s holds the test password", f)
		}
	}
}

// A render that stops between its stages leaves no page with the password in it: the staging file is removed, and
// nothing outside the page volume ever held one. The failure is made by one change: a placeholder the renderer
// does not fill, found only after the password was written into the staging file.
func TestTheStartPageRenderThatStopsLeavesNoCopyOfThePassword(t *testing.T) {
	dir := t.TempDir()
	pageDir := filepath.Join(dir, "page")
	tmp := filepath.Join(dir, "tmp")
	for _, d := range []string{pageDir, tmp} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	rows := filepath.Join(dir, "rows.html")
	if err := os.WriteFile(rows, []byte("<tr></tr>\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(dir, "broken.tmpl")
	tmpl := string(mustRead(t, filepath.Join(root, "lab", "page", "lab.html.tmpl"))) + "\n@@NOT_FILLED@@\n"
	if err := os.WriteFile(broken, []byte(tmpl), 0o600); err != nil {
		t.Fatal(err)
	}
	const password = "stop-secret-0123456789"
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/render.sh"; render_page`)
	cmd.Env = append(os.Environ(),
		"ROOT="+mustAbs(t, root), "PAGE_TEMPLATE="+broken, "PAGE_DIR="+pageDir, "PR_ROWS_FILE="+rows, "TMPDIR="+tmp,
		"CLEAN_COUNT=2", "LAB_URL=http://127.0.0.1:3300/", "LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount",
		"LAB_ADMIN_PASSWORD="+password)
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "the start page still holds a placeholder") {
		t.Fatalf("the render did not stop by name: %v\n%s", err, out)
	}
	if hits := walkFor(t, dir, password); len(hits) > 0 {
		t.Errorf("a stopped render left a copy of the password: %v", hits)
	}
	entries, _ := os.ReadDir(pageDir)
	if len(entries) != 0 {
		t.Errorf("a stopped render left %d entries in the page volume", len(entries))
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
			// Every port, the page's (it carries the lab's password) included, names 127.0.0.1 itself: no setting moves it.
			if !strings.HasPrefix(trim, `- "127.0.0.1:`) {
				t.Errorf("a published port is not bound to 127.0.0.1: %s", trim)
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
		`"127.0.0.1:${LAB_PAGE_PORT}:${LAB_PAGE_CONTAINER_PORT}"`,
		`"127.0.0.1:${LAB_DOCS_PORT}:${LAB_DOCS_CONTAINER_PORT}"`,
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
		"LAB_PAGE_PORT=3302", "LAB_PAGE_PATH=/lab", "LAB_PAGE_IMAGE=", "LAB_PAGE_IMAGE_VERSION=", "LAB_PAGE_CONTAINER_PORT=",
		"LAB_DOCS_PORT=3303", "LAB_DOCS_CONTAINER_PORT=", "DOCS_IMAGE=", "DOCS_IMAGE_TAG=", "NODE_IMAGE=",
		"GITEA_HTTP_PORT=3300", "LGTM_GRAFANA_PORT=3301", "LGTM_DASHBOARD_UID=proven-overview", "LAB_DEMO_HEAD=demo-discount",
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
