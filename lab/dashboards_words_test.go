package lab

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The owner's rule of 27 Oct: no raw field or enum name reaches a screen a visitor sees. A label value Proven emits is
// shown on a dashboard only through the words of HUMAN in lab/grafana/build_dashboards.py, either rewritten in the query
// (label_replace) or mapped in a table column (value mappings).

// emittedValues reads, from Proven's own sources, every raw value of each label a dashboard shows, so the table cannot
// fall behind a new value: the drop reasons and reviewers of internal/telemetry, the agents, reviewer verdicts and
// disagreement kinds of internal/pipeline, and the proof levels of internal/review.
func emittedValues(t *testing.T) map[string][]string {
	t.Helper()
	tel := read(t, "internal", "telemetry", "telemetry.go")
	pipe := read(t, "internal", "pipeline", "pipeline.go")
	rev := read(t, "internal", "review", "review.go")
	// returned reads only the values a function returns, not the strings it compares against.
	returned := func(src, from string) []string {
		i := strings.Index(src, from)
		if i < 0 {
			t.Fatalf("source anchor %q is gone; read the label values again", from)
		}
		body := src[i:]
		if j := strings.Index(body, "\n}\n"); j > 0 {
			body = body[:j]
		}
		var out []string
		for _, m := range regexp.MustCompile(`return "([a-z][a-z_]*)"`).FindAllStringSubmatch(body, -1) {
			out = append(out, m[1])
		}
		return out
	}
	quoted := func(src, from, to string) []string {
		i := strings.Index(src, from)
		if i < 0 {
			t.Fatalf("source anchor %q is gone; read the label values again", from)
		}
		body := src[i:]
		if j := strings.Index(body, to); j > 0 {
			body = body[:j]
		}
		var out []string
		for _, m := range regexp.MustCompile(`"([a-z][a-z_-]*)"`).FindAllStringSubmatch(body, -1) {
			out = append(out, m[1])
		}
		return out
	}
	vals := map[string][]string{
		"reason":             append(returned(tel, "func dropReason("), "suppressed"),
		"proven_reviewer":    quoted(tel, "ReviewerDefault =", ")\n"),
		"proven_agent":       quoted(pipe, "AgentPlanner      =", "ToolGovulncheck"),
		"verdict":            returned(pipe, "func verdictOf("),
		"proven_proof_level": quoted(rev, "ProvedByTest ProofLevel =", ")\n"),
	}
	for _, m := range regexp.MustCompile(`tel\.Disagreement\(context\.Background\(\), "([a-z_]+)"\)`).FindAllStringSubmatch(pipe, -1) {
		vals["kind"] = append(vals["kind"], m[1])
	}
	for k, v := range vals {
		if len(v) == 0 {
			t.Fatalf("no emitted value found for %s: the source moved, read it again", k)
		}
	}
	return vals
}

// dashboardText is every provisioned dashboard's JSON, as one string.
func dashboardText(t *testing.T, dir string) string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "lab", "grafana", "dashboards", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no dashboards under %s: %v", dir, err)
	}
	sort.Strings(files)
	var b strings.Builder
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(data)
	}
	return b.String()
}

// unmapped names every emitted value a dashboard shows raw: a label in a legend or a table column with no words for one
// of its values.
func unmapped(t *testing.T, text string, vals map[string][]string) []string {
	var miss []string
	for label, raws := range vals {
		shown := strings.Contains(text, "{{"+label+"}}") || strings.Contains(text, `\"`+label+`\"`) ||
			strings.Contains(text, `"`+label+`"`)
		if !shown {
			continue
		}
		for _, raw := range raws {
			inQuery := strings.Contains(text, `\"`+label+`\", \"`+raw+`\")`)
			inTable := strings.Contains(text, `"`+raw+`": {`)
			if !inQuery && !inTable {
				miss = append(miss, label+"="+raw)
			}
		}
	}
	sort.Strings(miss)
	return miss
}

func TestNoRawValueReachesADashboard(t *testing.T) {
	vals := emittedValues(t)
	if miss := unmapped(t, dashboardText(t, root), vals); len(miss) > 0 {
		t.Errorf("raw values a visitor would read on a dashboard (give each its words in HUMAN, lab/grafana/build_dashboards.py): %v", miss)
	}
	// No visible text of a dashboard is snake_case: titles, descriptions, units, the words of a mapping.
	snake := regexp.MustCompile(`"(title|description|noValue|text|unit|content)": "[^"]*\b[a-z][a-z0-9]*_[a-z0-9_]+\b`)
	if m := snake.FindString(dashboardText(t, root)); m != "" {
		t.Errorf("a visible text is snake_case: %s", m)
	}
}

// The control: the table without the words for no_evidence, regenerated, must turn the test red.
func TestNoRawValueTurnsRedWithoutItsWords(t *testing.T) {
	needTool(t, "python3")
	tree := t.TempDir()
	for _, rel := range []string{"lab/grafana/build_dashboards.py", "lab/seed/prs.tsv",
		"internal/telemetry/telemetry.go", "internal/pipeline/pipeline.go", "internal/review/review.go"} {
		dst := filepath.Join(tree, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, []byte(read(t, strings.Split(rel, "/")...)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mutate(t, tree, "lab/grafana/build_dashboards.py", `"no_evidence": "No evidence", `, ``)
	out := filepath.Join(tree, "lab", "grafana", "dashboards")
	cmd := exec.Command("python3", filepath.Join(tree, "lab", "grafana", "build_dashboards.py"))
	cmd.Env = append(os.Environ(), "DASHBOARDS_OUT="+out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("regenerate: %v\n%s", err, b)
	}
	miss := unmapped(t, dashboardText(t, tree), emittedValues(t))
	if len(miss) == 0 || miss[0] != "reason=no_evidence" {
		t.Errorf("the control did not turn red on no_evidence: %v", miss)
	}
}
