package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// promName is how Prometheus spells an OTLP instrument: dots become underscores. A counter gains
// _total, a histogram of seconds gains _seconds and its _bucket, _sum and _count series.
func promNames(instrument string) []string {
	base := strings.ReplaceAll(instrument, ".", "_")
	switch instrument {
	case MetricModelDuration:
		return []string{base + "_seconds_bucket", base + "_seconds_sum", base + "_seconds_count"}
	case MetricAgentState:
		return []string{base}
	}
	return []string{base + "_total"}
}

// TestTheDashboardQueriesOnlyWhatProvenEmits holds lab/grafana/dashboards/proven-review.json to the
// instruments in this package. A metric the dashboard asks for that no instrument makes is a panel that
// stays empty for ever; a renamed instrument must change the dashboard in the same commit.
func TestTheDashboardQueriesOnlyWhatProvenEmits(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "lab", "grafana", "dashboards", "proven-review.json"))
	if err != nil {
		t.Fatal(err)
	}
	var dash struct {
		UID    string `json:"uid"`
		Panels []struct {
			Title   string `json:"title"`
			Targets []struct {
				Expr string `json:"expr"`
			} `json:"targets"`
		} `json:"panels"`
	}
	if err := json.Unmarshal(raw, &dash); err != nil {
		t.Fatalf("the dashboard is not JSON: %v", err)
	}
	if dash.UID == "" || len(dash.Panels) == 0 {
		t.Fatalf("the dashboard has uid %q and %d panels", dash.UID, len(dash.Panels))
	}
	known := map[string]bool{}
	for _, m := range []string{MetricFindings, MetricDropped, MetricVerdicts, MetricDisagreements, MetricModelCalls,
		MetricModelDuration, MetricGuardChecks, MetricGuardUnsafe, MetricAgentState} {
		for _, n := range promNames(m) {
			known[n] = true
		}
	}
	ident := regexp.MustCompile(`\b(?:proven|gen_ai)_[a-z_]+`)
	used := map[string]bool{}
	for _, p := range dash.Panels {
		for _, tg := range p.Targets {
			for _, name := range ident.FindAllString(tg.Expr, -1) {
				// proven_reviewer, proven_agent and proven_proof_level are labels, not metrics.
				if name == "proven_reviewer" || name == "proven_agent" || name == "proven_proof_level" {
					continue
				}
				used[name] = true
				if !known[name] {
					t.Errorf("panel %q queries %s, which no instrument makes", p.Title, name)
				}
			}
		}
	}
	// Every instrument is on the dashboard, so none is emitted for nobody.
	var missing []string
	for name := range known {
		if strings.HasSuffix(name, "_bucket") {
			continue // the dashboard reads the histogram through its sum and count
		}
		if !used[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("no panel shows %v", missing)
	}
}
