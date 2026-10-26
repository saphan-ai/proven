package baseline

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/store"
)

// The fixture below is synthetic. Its packages, advisories and CVE ids are
// made up for the test and name no real advisory: the tests check what the
// harness does with an answer, not what Proven finds in a real snapshot.

const fixtureLock = `{
  "name": "fixture", "version": "0.0.0", "lockfileVersion": 3, "requires": true,
  "packages": {
    "": {"name": "fixture", "version": "0.0.0", "dependencies": {"fixture-a": "1.2.0", "fixture-b": "2.0.0", "fixture-c": "3.1.0"}},
    "node_modules/fixture-a": {"version": "1.2.0"},
    "node_modules/fixture-b": {"version": "2.0.0"},
    "node_modules/fixture-c": {"version": "3.1.0"}
  }
}
`

const fixtureGoMod = "module example.org/svc\n\ngo 1.26\n\nrequire example.org/gomod v0.5.0\n"

const (
	lockPath  = "web/package-lock.json"
	goModPath = "svc/go.mod"
)

// osvRows are the advisories of the fixture's evidence, in the table's columns.
var osvRows = [][]string{
	{"GHSA-fix-aaaa-0001", "CVE-2099-0001", "npm", "fixture-a", "1.0.0", "1.3.0", "", "HIGH", "2099-01-01"},
	{"GHSA-fix-extra-0004", "", "npm", "fixture-a", "1.1.0", "1.2.5", "", "LOW", "2099-01-02"},
	{"GHSA-fix-bbbb-0002", "", "npm", "fixture-b", "0", "2.5.0", "", "HIGH", "2099-01-03"},
	{"GHSA-fix-cccc-0003", "", "npm", "fixture-c", "0", "3.0.0", "", "HIGH", "2099-01-04"},
}

var kevRows = [][]string{{"CVE-2099-0001", "example", "fixture-a", "2099-01-05", "Unknown"}}

func fixtureReference() Reference {
	alert := func(n int, eco, pkg, manifest, ghsa, cve, rng string) Alert {
		return Alert{Number: n, Ecosystem: eco, Package: pkg, Manifest: manifest, GHSA: ghsa, CVE: cve, VulnerableRange: rng, Patched: "x"}
	}
	return Reference{
		Repository: "example/fixture", PinnedCommit: "0123456789abcdef0123456789abcdef01234567",
		Alerts: []Alert{
			alert(1, "npm", "fixture-a", lockPath, "GHSA-fix-aaaa-0001", "CVE-2099-0001", ">= 1.0.0, < 1.3.0"),
			alert(2, "npm", "fixture-b", lockPath, "GHSA-fix-bbbb-9999", "CVE-2099-9999", "< 2.5.0"),
			alert(3, "npm", "fixture-c", lockPath, "GHSA-fix-cccc-0003", "", "< 3.0.0"),
			alert(4, "npm", "fixture-d", lockPath, "GHSA-fix-dddd-0005", "", "< 1.0.0"),
			alert(5, "go", "example.org/gomod", goModPath, "GHSA-fix-gggg-0006", "", "< 0.6.0"),
		},
	}
}

// writeTree writes the change from an empty base: the lock and go.mod copies.
func writeTree(t *testing.T, lock string) (string, []string) {
	t.Helper()
	tree := t.TempDir()
	for path, body := range map[string]string{lockPath: lock, goModPath: fixtureGoMod} {
		dst := filepath.Join(tree, "change", filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(tree, "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	return tree, []string{goModPath, lockPath}
}

// localRunner is a real local evidence file holding the fixture's tables.
func localRunner(t *testing.T, osvRows [][]string) *store.Local {
	t.Helper()
	loc, err := store.OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	loc.SetClock(func() time.Time { return time.Date(2099, 1, 6, 0, 0, 0, 0, time.UTC) })
	ctx := context.Background()
	if err := loc.Register(ctx, store.Table{Name: querytpl.DefaultOSVTable, Columns: querytpl.OSVPackageColumns, Rows: osvRows}); err != nil {
		t.Fatal(err)
	}
	kevCols := []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}
	if err := loc.Register(ctx, store.Table{Name: querytpl.DefaultKEVView, Columns: kevCols, Rows: kevRows}); err != nil {
		t.Fatal(err)
	}
	return loc
}

func analyze(t *testing.T, ref Reference, lock string, rows [][]string) Findings {
	t.Helper()
	tree, manifests := writeTree(t, lock)
	f, err := Analyze(context.Background(), tree, manifests, ref, store.Runner(localRunner(t, rows)),
		Options{Commit: ref.PinnedCommit, Evidence: "fixture OSV table, 2099-01-06", Store: "local"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	return f
}

func verdictOf(t *testing.T, ref Reference, f Findings, number int) Row {
	t.Helper()
	for _, a := range ref.Alerts {
		if a.Number == number {
			return Judge(a, f)
		}
	}
	t.Fatalf("no alert %d", number)
	return Row{}
}

func TestJudgeEachVerdict(t *testing.T) {
	ref := fixtureReference()
	f := analyze(t, ref, fixtureLock, osvRows)
	cases := []struct {
		number  int
		want    Verdict
		contain string
	}{
		{1, Found, "on the KEV list: CVE-2099-0001"},
		{2, FoundDifferently, "GHSA-fix-bbbb-0002"},
		{3, FoundDifferently, "lies outside it"},
		{4, Missed, "does not read this package"},
		{5, Missed, "no row of the evidence names"},
	}
	for _, c := range cases {
		row := verdictOf(t, ref, f, c.number)
		if row.Verdict != c.want || !strings.Contains(row.Detail, c.contain) {
			t.Errorf("alert %d: got %q %q, want %q containing %q", c.number, row.Verdict, row.Detail, c.want, c.contain)
		}
	}
	if row := verdictOf(t, ref, f, 1); !strings.HasPrefix(row.EvidenceID, "ev-") {
		t.Errorf("a found alert carries no evidence id: %q", row.EvidenceID)
	}
}

// Mutated field: the package's entry in the lock file Proven reads. With it
// the row is found; without it the row is missed, because Proven never read
// the package. A harness that says found whatever the manifest holds fails here.
func TestRemovingAPackageFromTheLockTurnsItsRowMissed(t *testing.T) {
	ref := fixtureReference()
	if got := verdictOf(t, ref, analyze(t, ref, fixtureLock, osvRows), 1).Verdict; got != Found {
		t.Fatalf("with the entry, alert 1 is %q, want found", got)
	}
	mutated := strings.ReplaceAll(fixtureLock, `, "fixture-a": "1.2.0"`, "")
	mutated = strings.ReplaceAll(mutated, "    \"node_modules/fixture-a\": {\"version\": \"1.2.0\"},\n", "")
	if mutated == fixtureLock {
		t.Fatal("the mutation changed nothing")
	}
	row := verdictOf(t, ref, analyze(t, ref, mutated, osvRows), 1)
	if row.Verdict != Missed || !strings.Contains(row.Detail, "does not read this package") {
		t.Errorf("without the entry, alert 1 is %q %q, want missed because the package is not read", row.Verdict, row.Detail)
	}
}

// Mutated field: the alert's advisory ids in a copy of the reference. An id
// the evidence does not hold for that package is never found.
func TestChangingAnAlertsAdvisoryIDNeverGivesFound(t *testing.T) {
	ref := fixtureReference()
	f := analyze(t, ref, fixtureLock, osvRows)
	changed := ref.Alerts[0]
	changed.GHSA, changed.CVE = "GHSA-fix-zzzz-0000", "CVE-2099-0000"
	row := Judge(changed, f)
	if row.Verdict == Found {
		t.Fatalf("an advisory id the evidence does not hold was found: %q", row.Detail)
	}
	if row.Verdict != FoundDifferently {
		t.Errorf("alert 1 with a changed id is %q, want found differently (another advisory affects the package)", row.Verdict)
	}
	// With no other advisory affecting the package the row is missed.
	only := analyze(t, ref, fixtureLock, osvRows[2:])
	if got := Judge(changed, only).Verdict; got != Missed {
		t.Errorf("alert 1 with a changed id and no other row is %q, want missed", got)
	}
}

func TestExtrasAreAdvisoriesNoAlertNames(t *testing.T) {
	ref := fixtureReference()
	got := Extras(ref, analyze(t, ref, fixtureLock, osvRows))
	var ids []string
	for _, e := range got {
		ids = append(ids, e.Package+":"+e.Advisory)
	}
	want := "fixture-a:GHSA-fix-extra-0004 fixture-b:GHSA-fix-bbbb-0002"
	if strings.Join(ids, " ") != want {
		t.Errorf("extras = %v, want %s", ids, want)
	}
}

func TestManifestOnlyRunChecksNothing(t *testing.T) {
	ref := fixtureReference()
	tree, manifests := writeTree(t, fixtureLock)
	f, err := Analyze(context.Background(), tree, manifests, ref, nil, Options{Commit: ref.PinnedCommit, Evidence: "none"})
	if err != nil {
		t.Fatal(err)
	}
	for number, want := range map[int]Verdict{1: NotChecked, 4: Missed} {
		if got := verdictOf(t, ref, f, number).Verdict; got != want {
			t.Errorf("manifest-only alert %d is %q, want %q", number, got, want)
		}
	}
	if e := Extras(ref, f); len(e) != 0 {
		t.Errorf("a manifest-only run found %d extras", len(e))
	}
}

// A store that is down is not "missed": the run is refused and says why.
func TestDownStoreRefusesTheRun(t *testing.T) {
	ref := fixtureReference()
	tree, manifests := writeTree(t, fixtureLock)
	_, err := Analyze(context.Background(), tree, manifests, ref, downRunner{}, Options{Commit: ref.PinnedCommit, Evidence: "x"})
	if err == nil || !strings.Contains(err.Error(), "evidence service zetaris is down") {
		t.Fatalf("Analyze with a down store returned %v", err)
	}
}

type downRunner struct{}

func (downRunner) Run(context.Context, querytpl.Request, querytpl.Relations) (query.Result, error) {
	return query.Result{}, &service.Down{Service: "zetaris", Reason: "unreachable"}
}

func TestReportIsByteIdenticalTwice(t *testing.T) {
	ref := fixtureReference()
	one := Report(ref, analyze(t, ref, fixtureLock, osvRows))
	two := Report(ref, analyze(t, ref, fixtureLock, osvRows))
	if one != two {
		t.Fatal("two runs on the same input wrote different reports")
	}
	for _, want := range []string{"## Counts", "| **all** | **5** | **1** | **2** | **2** | **0** |", "evidence: fixture OSV table, 2099-01-06", "ev-"} {
		if !strings.Contains(one, want) {
			t.Errorf("report lacks %q", want)
		}
	}
}

func TestInStatedRange(t *testing.T) {
	r, _ := manifest.ByEcosystem("npm")
	cases := []struct {
		version, rng string
		inside, ok   bool
	}{
		{"15.3.0", ">= 15.3.0, < 15.3.3", true, true},
		{"15.3.3", ">= 15.3.0, < 15.3.3", false, true},
		{"15.3.0", ">= 15.0.0-canary.0, < 15.4.7", true, true},
		{"8.5.31", "<= 8.5.17", false, true},
		{"8.4.31", "<= 8.5.17", true, true},
		{"1.0.0", "> 1.0.0", false, true},
		{"1.0.0", "~1.0", false, false},
		{"not-a-version", "< 1.0.0", false, false},
	}
	for _, c := range cases {
		inside, ok := inStatedRange(r.Compare, c.version, c.rng)
		if inside != c.inside || ok != c.ok {
			t.Errorf("%s in %q = (%v, %v), want (%v, %v)", c.version, c.rng, inside, ok, c.inside, c.ok)
		}
	}
}

func TestExportKeepsOnlyTheManifestsOfTheCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH; CI's go test job has it")
	}
	repo := t.TempDir()
	for path, body := range map[string]string{lockPath: fixtureLock, goModPath: fixtureGoMod, "README.md": "text\n", "svc/main.go": "package main\n"} {
		dst := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@example.org", "commit", "-q", "-m", "[export-test] repo: a fixture commit"}} {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	dir := t.TempDir()
	got, err := Export(repo, "HEAD", dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, " ") != goModPath+" "+lockPath {
		t.Errorf("Export kept %v, want only the go.mod and the lock", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "base")); err != nil {
		t.Errorf("the base is not an (empty) directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "change", "README.md")); err == nil {
		t.Error("Export copied a file that is not a manifest")
	}
}

// The committed report is what the committed inputs give. Regenerate it with
// `make baseline-dependabot-manifests`, or `make baseline-dependabot` with an evidence store.
func TestCommittedReportIsRegenerated(t *testing.T) {
	dir := filepath.Join("..", "..", "test", "corpus", "reference", "dependabot-fbdf7d3")
	ref, err := LoadReference(filepath.Join(dir, "alerts.json"))
	if err != nil {
		t.Fatal(err)
	}
	f, err := LoadFindings(filepath.Join(dir, FindingsFile))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(dir, ReportFile))
	if err != nil {
		t.Fatal(err)
	}
	if got := Report(ref, f); got != string(want) {
		t.Fatalf("%s is not what %s and %s give; run the make target again instead of editing it", ReportFile, "alerts.json", FindingsFile)
	}
	if f.Commit != ref.PinnedCommit {
		t.Errorf("findings are for %s, the reference is pinned at %s", f.Commit, ref.PinnedCommit)
	}
	// Negative control: the same test turns red when one read is dropped.
	dropped := f
	dropped.Reads = f.Reads[1:]
	if Report(ref, dropped) == string(want) {
		t.Error("the report does not change when a read is removed, so this test cannot fail")
	}
	// The reference itself is frozen: 51 alerts, as the master pinned it.
	if len(ref.Alerts) != 51 {
		t.Errorf("the reference holds %d alerts, want the 51 that were frozen", len(ref.Alerts))
	}
}
