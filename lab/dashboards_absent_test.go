package lab

// A number on a dashboard is a measurement or it is nothing. A stat whose query has no answer says so in words (its noValue
// sentence); it never turns the absence into a 0 that reads as "none, good". A zero is shown only where a series proves the
// source reported: a review finished, a model was called, a receipt was loaded.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// absentProblems names every stat panel of the dashboards in dir whose absence could read as a measured zero.
func absentProblems(t *testing.T, dashboards map[string][]byte) []string {
	t.Helper()
	var bad []string
	for uid, raw := range dashboards {
		var d struct {
			Panels []struct {
				Type        string `json:"type"`
				Title       string `json:"title"`
				FieldConfig struct {
					Defaults struct {
						NoValue string `json:"noValue"`
					} `json:"defaults"`
				} `json:"fieldConfig"`
				Targets []struct {
					Expr string `json:"expr"`
				} `json:"targets"`
			} `json:"panels"`
		}
		if err := json.Unmarshal(raw, &d); err != nil {
			t.Fatalf("%s: %v", uid, err)
		}
		for _, p := range d.Panels {
			for _, tg := range p.Targets {
				if strings.Contains(tg.Expr, "vector(") {
					bad = append(bad, uid+" / "+p.Title+": a constant answers when the metric is absent")
				}
			}
			if p.Type == "table" {
				for _, tg := range p.Targets {
					if strings.Contains(tg.Expr, "* 0)") {
						bad = append(bad, uid+" / "+p.Title+": a table cell gets a substituted zero when its metric is absent")
					}
				}
			}
			if p.Type != "stat" {
				continue
			}
			nv := strings.TrimSpace(p.FieldConfig.Defaults.NoValue)
			if nv == "" || strings.HasPrefix(nv, "0") || strings.Contains(nv, "good") {
				bad = append(bad, uid+" / "+p.Title+": no data does not say it is not measured ("+nv+")")
			}
		}
	}
	return bad
}

func TestNoAbsentMetricReadsAsAMeasuredZero(t *testing.T) {
	for _, b := range absentProblems(t, readDashboards(t)) {
		t.Error(b)
	}
	// The negative control: the generator's measured zero goes back to a constant 0. Mutated field: measured() in
	// lab/grafana/build_dashboards.py.
	needTool(t, "python3")
	mut := copyLab(t)
	copyInto(t, mut, "lab/grafana", "lab/seed")
	mutate(t, mut, "lab/grafana/build_dashboards.py", `return "%s or (%s * 0)" % (expr, presence)`, `return "%s or vector(0)" % expr`)
	out := t.TempDir()
	cmd := exec.Command("python3", "-I", filepath.Join(mut, "lab", "grafana", "build_dashboards.py"))
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "DASHBOARDS_OUT=" + out}
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the mutated generator: %v\n%s", err, b)
	}
	mutated := map[string][]byte{}
	files, _ := filepath.Glob(filepath.Join(out, "*.json"))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		mutated[strings.TrimSuffix(filepath.Base(f), ".json")] = b
	}
	if len(absentProblems(t, mutated)) == 0 {
		t.Error("the control did not turn red: with absence answered by vector(0) no panel was named")
	}
	// The second control: a table cell answered by another series times zero (review r5). Mutated field: the overview's
	// Claims dropped query.
	raw := readDashboards(t)["proven-overview"]
	var d map[string]any
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	for _, p := range d["panels"].([]any) {
		pm := p.(map[string]any)
		if pm["type"] != "table" {
			continue
		}
		for _, tg := range pm["targets"].([]any) {
			tm := tg.(map[string]any)
			if e, _ := tm["expr"].(string); strings.Contains(e, "claims_dropped") {
				tm["expr"] = "(" + e + ") or (count by (instance) (proven_agent_state) * 0)"
			}
		}
	}
	b, _ := json.Marshal(d)
	if len(absentProblems(t, map[string][]byte{"proven-overview": b})) == 0 {
		t.Error("the control did not turn red: a table zero substituted for an absent metric was not named")
	}
}
