package lab

// The six provisioned dashboards (lab/grafana/dashboards, written by lab/grafana/build_dashboards.py). What is held here is
// what can be held without a running Grafana: each file parses, is built to the one style (a title, a time range, at most
// eight panels that do not overlap, a description on every panel), and every query names only a metric and a label that
// Proven emits (internal/telemetry) or that the lab's Nightshift importer writes. How a panel looks is not held here: that
// needs Grafana, and is checked by the screenshots of a live make up.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/telemetry"
)

var dashboardUIDs = []string{"proven-engines", "proven-evidence", "proven-models", "proven-nightshift", "proven-overview", "proven-review"}

type dashPanel struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	GridPos     struct {
		X, Y, W, H int
	} `json:"gridPos"`
	Targets []struct {
		Expr  string `json:"expr"`
		Query string `json:"query"`
	} `json:"targets"`
}

type dashboardDoc struct {
	UID    string `json:"uid"`
	Title  string `json:"title"`
	Tags   []string
	Time   struct{ From, To string }
	Panels []dashPanel `json:"panels"`
}

// emittedPromMetrics is the fixture of what Proven emits, as Prometheus spells it: an OTLP counter gains _total, the
// duration histogram of seconds gains _seconds and its series. Built from the instrument names in internal/telemetry, so a
// renamed instrument turns a dashboard red here.
func emittedPromMetrics() map[string]bool {
	known := map[string]bool{}
	for _, m := range []string{telemetry.MetricFindings, telemetry.MetricDropped, telemetry.MetricVerdicts, telemetry.MetricDisagreements,
		telemetry.MetricModelCalls, telemetry.MetricGuardChecks, telemetry.MetricGuardUnsafe} {
		known[strings.ReplaceAll(m, ".", "_")+"_total"] = true
	}
	base := strings.ReplaceAll(telemetry.MetricModelDuration, ".", "_") + "_seconds"
	for _, s := range []string{"_bucket", "_sum", "_count"} {
		known[base+s] = true
	}
	known[strings.ReplaceAll(telemetry.MetricAgentState, ".", "_")] = true
	return known
}

// emittedLabels are the labels the metrics carry (the attributes in internal/telemetry/telemetry.go, dots as underscores)
// and the two Prometheus adds to every series.
var emittedLabels = map[string]bool{
	"proven_agent": true, "proven_reviewer": true, "proven_proof_level": true, "verdict": true, "kind": true, "reason": true,
	"outcome": true, "side": true, "gen_ai_request_model": true, "gen_ai_response_model": true, "instance": true, "job": true,
	// Made inside the query by label_replace and label_join: the pull request and the change cut out of the instance id.
	"pr": true, "change": true, "who": true, "route": true, "check": true,
}

// importerLabels and importerFields are what lab/seed/nightshift-import.sh writes: the stream labels, and the fields of the
// JSON line a query may unwrap.
var (
	importerLabels = map[string]bool{"__error__": true, "job": true, "app": true, "model": true, "status": true, "reason": true, "verified": true, "provider": true}
	importerFields = map[string]bool{"latency_ms": true, "input_tokens": true, "output_tokens": true}
)

var (
	promMetricRe = regexp.MustCompile(`\b(?:proven|gen_ai)_[a-z_]+`)
	byRe         = regexp.MustCompile(`\b(?:by|without)\s*\(([^)]*)\)`)
	matcherRe    = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*(?:=~|!=|=)\s*"`)
	unwrapRe     = regexp.MustCompile(`unwrap\s+([a-z_]+)`)
)

// lintDashboard returns every way the dashboard breaks the lab's one style, or asks for what nothing emits.
func lintDashboard(raw []byte) []string {
	var bad []string
	add := func(f string, a ...any) { bad = append(bad, fmt.Sprintf(f, a...)) }
	var d dashboardDoc
	if err := json.Unmarshal(raw, &d); err != nil {
		return []string{"does not parse: " + err.Error()}
	}
	if d.UID == "" || d.Title == "" {
		add("no uid or no title")
	}
	if d.Time.From == "" || d.Time.To == "" {
		add("no default time range")
	}
	if len(d.Panels) == 0 || len(d.Panels) > 8 {
		add("%d panels, want 1 to 8", len(d.Panels))
	}
	tagged := false
	for _, tg := range d.Tags {
		tagged = tagged || tg == "proven"
	}
	if !tagged {
		add("not tagged proven, so the other dashboards' links skip it")
	}
	known := emittedPromMetrics()
	ids := map[int]bool{}
	for i, p := range d.Panels {
		if ids[p.ID] {
			add("panel id %d twice", p.ID)
		}
		ids[p.ID] = true
		if p.Title == "" || p.Description == "" {
			add("panel %d has no title or no description", p.ID)
		}
		g := p.GridPos
		if g.W <= 0 || g.H <= 0 || g.X < 0 || g.X+g.W > 24 {
			add("panel %q is outside the 24-column grid", p.Title)
		}
		for _, q := range d.Panels[i+1:] {
			h := q.GridPos
			if g.X < h.X+h.W && h.X < g.X+g.W && g.Y < h.Y+h.H && h.Y < g.Y+g.H {
				add("panels %q and %q overlap", p.Title, q.Title)
			}
		}
		for _, tg := range p.Targets {
			if strings.Contains(tg.Expr, `job="nightshift"`) {
				lintLoki(tg.Expr, p.Title, add)
				continue
			}
			for _, name := range promMetricRe.FindAllString(tg.Expr, -1) {
				if emittedLabels[name] {
					continue
				}
				if !known[name] {
					add("panel %q queries %s, which Proven does not emit", p.Title, name)
				}
			}
			lintLabels(tg.Expr, emittedLabels, p.Title, add)
		}
	}
	return bad
}

func lintLoki(expr, title string, add func(string, ...any)) {
	lintLabels(expr, importerLabels, title, add)
	for _, m := range unwrapRe.FindAllStringSubmatch(expr, -1) {
		if !importerFields[m[1]] {
			add("panel %q unwraps %s, a field the importer does not write", title, m[1])
		}
	}
}

// lintLabels checks the labels a query groups by and matches on. A label nothing carries is a panel that stays empty.
func lintLabels(expr string, allowed map[string]bool, title string, add func(string, ...any)) {
	for _, m := range byRe.FindAllStringSubmatch(expr, -1) {
		for _, l := range strings.Split(m[1], ",") {
			if l = strings.TrimSpace(l); l != "" && !allowed[l] {
				add("panel %q groups by %s, which no series carries", title, l)
			}
		}
	}
	for _, m := range matcherRe.FindAllStringSubmatch(expr, -1) {
		if !allowed[m[1]] {
			add("panel %q matches on %s, which no series carries", title, m[1])
		}
	}
}

func readDashboards(t *testing.T) map[string][]byte {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "lab", "grafana", "dashboards", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		out[strings.TrimSuffix(filepath.Base(f), ".json")] = b
	}
	return out
}

func TestEveryProvisionedDashboardIsBuiltToTheOneStyleAndQueriesWhatIsEmitted(t *testing.T) {
	dashboards := readDashboards(t)
	var got []string
	for name := range dashboards {
		got = append(got, name)
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(dashboardUIDs, ",") {
		t.Fatalf("the dashboards are %v, want the six %v", got, dashboardUIDs)
	}
	for name, raw := range dashboards {
		var d dashboardDoc
		_ = json.Unmarshal(raw, &d)
		if d.UID != name {
			t.Errorf("%s.json holds the uid %q: the file is named for its uid", name, d.UID)
		}
		for _, problem := range lintDashboard(raw) {
			t.Errorf("%s: %s", name, problem)
		}
	}
	// The negative control: a misspelt metric. Mutated field: the metric name in one query of the Overview. It must turn red.
	mut := bytes.Replace(dashboards["proven-overview"], []byte("proven_review_findings_total"), []byte("proven_review_findings_totl"), 1)
	if bytes.Equal(mut, dashboards["proven-overview"]) {
		t.Fatal("the mutation changed nothing")
	}
	if len(lintDashboard(mut)) == 0 {
		t.Error("the control did not turn red: a misspelt metric passed the lint")
	}
	// And a label nothing carries, in Loki.
	mut = bytes.Replace(dashboards["proven-nightshift"], []byte(`sum by (reason) (count_over_time`), []byte(`sum by (reasons) (count_over_time`), 1)
	if bytes.Equal(mut, dashboards["proven-nightshift"]) || len(lintDashboard(mut)) == 0 {
		t.Error("the control did not turn red: a misspelt Loki label passed the lint")
	}
}

func TestTheOverviewHoldsTheDemoManifestToItsCounts(t *testing.T) {
	var clean, defect []string
	for _, r := range manifest(t) {
		if r.verdict == "clean" {
			clean = append(clean, r.fixture)
		} else {
			defect = append(defect, r.fixture)
		}
	}
	raw := string(readDashboards(t)["proven-overview"])
	for _, f := range append(append([]string{}, clean...), defect...) {
		if !strings.Contains(raw, f) {
			t.Errorf("the Overview does not know the demo change %q: lab/seed/prs.tsv changed, run python3 lab/grafana/build_dashboards.py", f)
		}
	}
	if want := fmt.Sprintf("of %d flagged", len(defect)); !strings.Contains(raw, want) {
		t.Errorf("the Overview does not say %q", want)
	}
	if want := fmt.Sprintf("on the %d clean demo pull requests", len(clean)); !strings.Contains(raw, want) {
		t.Errorf("the Overview does not say %q", want)
	}
}

func TestTheDashboardsAreTheGeneratorsOutput(t *testing.T) {
	needTool(t, "python3")
	out := t.TempDir()
	cmd := exec.Command("python3", "-I", filepath.Join(root, "lab", "grafana", "build_dashboards.py"))
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "DASHBOARDS_OUT=" + out}
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the generator: %v\n%s", err, b)
	}
	for name, committed := range readDashboards(t) {
		fresh, err := os.ReadFile(filepath.Join(out, name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fresh, committed) {
			t.Errorf("%s.json is not what lab/grafana/build_dashboards.py writes: run it and commit the result", name)
		}
	}
}

func TestTheDashboardsThatLackMetricsSayWhichOnesAreMissing(t *testing.T) {
	// The Evidence and Engines dashboards bind to what exists and name the rest in a text panel: not a query that stays empty.
	for _, c := range []struct {
		uid  string
		want []string
	}{
		{"proven-evidence", []string{"queries per table (OSV, KEV, EPSS)", "rows returned", "query duration", "queries Zetaris refused", "network off"}},
		{"proven-engines", []string{"plans, by light or full", "guard overrides", "steps taken", "claims", "evidence"}},
	} {
		raw := string(readDashboards(t)[c.uid])
		for _, w := range c.want {
			if !strings.Contains(raw, w) {
				t.Errorf("%s names no missing metric %q", c.uid, w)
			}
		}
		if !strings.Contains(raw, "Not emitted by Proven yet.") {
			t.Errorf("%s does not say its text panels are for metrics Proven does not emit yet", c.uid)
		}
	}
}

func TestTheLabNamesEachReviewByItsPullRequestAndChange(t *testing.T) {
	// The dashboards read a review's pull request and change from its instance id; the seed's workflow sets that id.
	seed := read(t, "lab", "seed", "seed.sh")
	if !strings.Contains(seed, `OTEL_RESOURCE_ATTRIBUTES=service.instance.id=\$(date +%s)-pr\${PR}-\${fixture}`) {
		t.Error("the seed's workflow does not set the review's instance id to <time>-pr<number>-<fixture>")
	}
	wf := read(t, "lab", "workflow", "proven.yml")
	for _, w := range []string{"PR: ${{ github.event.pull_request.number }}", `fixture="${HEAD_REF#demo-}"`} {
		if !strings.Contains(wf, w) {
			t.Errorf("the workflow no longer holds %q, which the instance id reads", w)
		}
	}
	compose := read(t, "lab", "compose.yaml")
	if !strings.Contains(compose, "GF_USERS_DEFAULT_THEME: dark") || !strings.Contains(compose, "proven-overview.json") {
		t.Error("Grafana is not set to the dark theme and the Overview as its home")
	}
}
