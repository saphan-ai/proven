package lab

// The visual bar of the six dashboards (the owner's, 26 Oct: a five-second read, colour always with a word, no empty panel
// without its reason in words, no cut label), held as far as it can be held without a browser. Each rule is one a live run of
// round 1 broke: a duration 10^6 too large, a "no data" sentence in every table cell, a bar chart that was an empty box, a
// table that hid a row, a time range that put every review in one bar. How a panel looks is still checked by the screenshots
// of a live make up; what is held here is the JSON that makes it look so.

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"testing"
)

// The fixture of the emitted telemetry: a lab's reviews and its replayed receipts all happen within one wait of make up,
// which is bounded by DEMO_WAIT_SECONDS (lab/versions.env, 1200 s). A time range must fit it: not so long that the whole run
// is a sliver at the edge, and a bar chart must step finely enough to spread it over several bars.
const (
	fixtureSpanMinutes = 20
	maxRangeMinutes    = 3 * fixtureSpanMinutes
	maxStepMinutes     = fixtureSpanMinutes / 5
)

type stylePanel struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	GridPos struct{ W, H int }
	Options struct {
		CellHeight    string `json:"cellHeight"`
		NamePlacement string `json:"namePlacement"`
	} `json:"options"`
	FieldConfig struct {
		Defaults struct {
			Unit        string `json:"unit"`
			DisplayName string `json:"displayName"`
			NoValue     string `json:"noValue"`
			Thresholds  struct {
				Steps []struct{ Color string }
			} `json:"thresholds"`
			Mappings []json.RawMessage `json:"mappings"`
		} `json:"defaults"`
		Overrides []struct {
			Matcher struct{ Options string }
			Props   []struct {
				ID    string          `json:"id"`
				Value json.RawMessage `json:"value"`
			} `json:"properties"`
		} `json:"overrides"`
	} `json:"fieldConfig"`
	Targets []struct {
		Expr         string `json:"expr"`
		LegendFormat string `json:"legendFormat"`
		Interval     string `json:"interval"`
		Limit        int    `json:"limit"`
		Datasource   struct{ Type string }
	} `json:"targets"`
}

type styleDoc struct {
	Time   struct{ From string }
	Panels []stylePanel `json:"panels"`
}

var (
	labelTplRe  = regexp.MustCompile(`\$\{__field\.labels\.([a-z_]+)\}`)
	minutesRe   = regexp.MustCompile(`^(?:now-)?(\d+)([mhd])$`)
	pluralWords = []string{"findings", "reviews", "claims", "calls", "agents", "receipts", "alarms", "models", "runs", "steps"}
	badStatBase = map[string]bool{"red": true, "yellow": true, "orange": true}
)

func minutesOf(s string) (int, bool) {
	m := minutesRe.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	var n int
	_, _ = fmt.Sscan(m[1], &n)
	return n * map[string]int{"m": 1, "h": 60, "d": 1440}[m[2]], true
}

// tableRowsFit is the least panel height, in grid units of about 38 px, that shows rows rows at the small cell height with
// a header and the title: the same sum as table_h in lab/grafana/build_dashboards.py.
func tableRowsFit(rows int) int { return int(math.Ceil(float64(rows+1)*30/38)) + 2 }

// lintVisualBar returns every way a dashboard breaks the visual bar.
func lintVisualBar(raw []byte) []string {
	var bad []string
	add := func(f string, a ...any) { bad = append(bad, fmt.Sprintf(f, a...)) }
	var d styleDoc
	if err := json.Unmarshal(raw, &d); err != nil {
		return []string{"does not parse: " + err.Error()}
	}
	if m, ok := minutesOf(d.Time.From); !ok || m > maxRangeMinutes {
		add("the default time range %q does not fit a run of %d minutes (at most %d)", d.Time.From, fixtureSpanMinutes, maxRangeMinutes)
	}
	for _, p := range d.Panels {
		if n := len([]rune(p.Title)); n > p.GridPos.W*6 {
			add("panel %q has %d characters, which a %d-column panel cuts at 1920 px (at most %d)", p.Title, n, p.GridPos.W, p.GridPos.W*6)
		}
		def := p.FieldConfig.Defaults
		for _, tg := range p.Targets {
			if (strings.Contains(tg.Expr, "_seconds_sum") || strings.Contains(tg.Expr, "_seconds_bucket")) && def.Unit != "s" {
				add("panel %q reads a duration in seconds and says its unit is %q", p.Title, def.Unit)
			}
			if strings.Contains(tg.Expr, "latency_ms") && def.Unit != "ms" {
				add("panel %q reads a duration in milliseconds and says its unit is %q", p.Title, def.Unit)
			}
		}
		switch p.Type {
		case "stat":
			unit := strings.TrimSpace(strings.TrimPrefix(def.Unit, "suffix:"))
			if !strings.HasPrefix(def.Unit, "suffix:") || unit == "" {
				if len(def.Mappings) == 0 {
					add("stat %q has colours but no word beside them", p.Title)
				}
			}
			if st := def.Thresholds.Steps; len(st) > 0 && badStatBase[st[0].Color] {
				add("stat %q is %s at its base step, so a zero that means nothing yet reads as an alarm", p.Title, st[0].Color)
			}
			for _, w := range pluralWords {
				if regexp.MustCompile(`\b` + w + `\b`).MatchString(unit) {
					add("stat %q puts %q after a number, which reads \"1 %s\"", p.Title, w, w)
				}
			}
		case "bargauge":
			m := labelTplRe.FindStringSubmatch(def.DisplayName)
			if m == nil {
				add("bar gauge %q has no display name template naming its label", p.Title)
				break
			}
			if p.Options.NamePlacement != "top" {
				add("bar gauge %q puts its names at the %q, where a long label is cut", p.Title, p.Options.NamePlacement)
			}
			for _, tg := range p.Targets {
				if !strings.Contains(tg.LegendFormat, "{{"+m[1]+"}}") {
					add("bar gauge %q names %s in its display name but its legend does not", p.Title, m[1])
				}
				if !strings.Contains(tg.Expr, "or on() label_replace(vector(0)") {
					add("bar gauge %q draws an empty box when there is no data: its query has no row that says why", p.Title)
				}
			}
		case "table":
			if def.NoValue != "" {
				add("table %q has a noValue sentence, which Grafana prints in every empty cell", p.Title)
			}
			if p.Options.CellHeight != "sm" {
				add("table %q has cell height %q, not sm", p.Title, p.Options.CellHeight)
			}
			rows := 0
			for _, tg := range p.Targets {
				if tg.Datasource.Type == "tempo" {
					rows = tg.Limit
				}
			}
			if rows > 0 && !strings.Contains(p.Title, fmt.Sprintf("(the latest %d)", rows)) {
				add("table %q shows only the latest %d rows and does not say so", p.Title, rows)
			}
			if need := tableRowsFit(rows); rows > 0 && p.GridPos.H < need {
				add("table %q is %d high and hides rows: %d rows need %d", p.Title, p.GridPos.H, rows, need)
			}
			if rows == 0 && p.GridPos.H < tableRowsFit(4) {
				add("table %q is %d high, too small for its rows", p.Title, p.GridPos.H)
			}
			for _, o := range p.FieldConfig.Overrides {
				for _, pr := range o.Props {
					if pr.ID == "noValue" {
						add("table %q has a noValue sentence on %q", p.Title, o.Matcher.Options)
					}
					// Tempo returns a duration in nanoseconds.
					if o.Matcher.Options == "Duration" && pr.ID == "unit" && string(pr.Value) != `"ns"` {
						add("table %q says its duration is %s: Tempo returns nanoseconds", p.Title, pr.Value)
					}
				}
			}
		case "timeseries":
			for _, tg := range p.Targets {
				if m, ok := minutesOf(tg.Interval); !ok || m > maxStepMinutes {
					add("chart %q steps by %q: a run of %d minutes must spread over bars of at most %d minutes", p.Title, tg.Interval, fixtureSpanMinutes, maxStepMinutes)
				}
			}
		}
	}
	return bad
}

// mutatePanel changes one panel of the dashboard (found by title prefix) through fn and returns the new JSON.
func mutatePanel(t *testing.T, raw []byte, titlePrefix string, fn func(p map[string]any)) []byte {
	t.Helper()
	var d map[string]any
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	done := false
	for _, x := range d["panels"].([]any) {
		p := x.(map[string]any)
		if strings.HasPrefix(p["title"].(string), titlePrefix) {
			fn(p)
			done = true
		}
	}
	if !done {
		t.Fatalf("no panel starts with %q", titlePrefix)
	}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func sub(m map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		m = m[k].(map[string]any)
	}
	return m
}

func TestEveryDashboardHoldsTheVisualBar(t *testing.T) {
	for name, raw := range readDashboards(t) {
		for _, problem := range lintVisualBar(raw) {
			t.Errorf("%s: %s", name, problem)
		}
	}
}

func TestTheVisualBarTurnsRedWhenEachRuleIsBroken(t *testing.T) {
	ds := readDashboards(t)
	for _, c := range []struct {
		name, uid, panel, want string
		mutate                 func(p map[string]any)
	}{
		{"a duration unit that is not the source's (mutated field: the Duration unit of a Tempo table)", "proven-overview", "Reviews with a failed service", "Tempo returns nanoseconds",
			func(p map[string]any) {
				for _, o := range sub(p, "fieldConfig")["overrides"].([]any) {
					for _, pr := range o.(map[string]any)["properties"].([]any) {
						if m := pr.(map[string]any); m["id"] == "unit" {
							m["value"] = "ms"
						}
					}
				}
			}},
		{"a noValue sentence on a table (mutated field: defaults.noValue)", "proven-engines", "What each review ran as", "noValue sentence",
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["noValue"] = "No review has run yet." }},
		{"a bar gauge with no display name (mutated field: defaults.displayName)", "proven-models", "What Proven refused", "display name template",
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["displayName"] = "" }},
		{"a bar gauge that can be an empty box (mutated field: the target's expr)", "proven-nightshift", "Why calls failed", "empty box",
			func(p map[string]any) {
				tg := p["targets"].([]any)[0].(map[string]any)
				tg["expr"] = strings.Split(tg["expr"].(string), " or on()")[0]
			}},
		{"a stat that is red at zero (mutated field: the base threshold colour)", "proven-nightshift", "Receipts loaded", "alarm",
			func(p map[string]any) {
				sub(p, "fieldConfig", "defaults", "thresholds")["steps"].([]any)[0].(map[string]any)["color"] = "red"
			}},
		{"a stat with no word (mutated field: unit and mappings)", "proven-models", "Model calls that returned", "no word beside them",
			func(p map[string]any) {
				d := sub(p, "fieldConfig", "defaults")
				d["unit"], d["mappings"] = "short", []any{}
			}},
		{"a plural after a number (mutated field: the stat's unit)", "proven-review", "Findings kept", `reads "1 findings"`,
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["unit"] = "suffix: findings kept" }},
		{"a table too short for its rows (mutated field: gridPos.h)", "proven-review", "The agents, in order", "hides rows",
			func(p map[string]any) { p["gridPos"].(map[string]any)["h"] = 8 }},
		{"a table that does not say it is cut (mutated field: the title)", "proven-evidence", "Investigator and Prover runs", "does not say so",
			func(p map[string]any) { p["title"] = "Investigator and Prover runs" }},
		{"a bar chart that steps too coarsely (mutated field: the target interval)", "proven-overview", "Reviews started over time", "must spread over bars",
			func(p map[string]any) { p["targets"].([]any)[0].(map[string]any)["interval"] = "10m" }},
		{"a title that 1920 px cuts (mutated field: the title)", "proven-overview", "Reviews finished", "cuts at 1920",
			func(p map[string]any) { p["title"] = "Reviews finished, counted by the Writer of every review" }},
		{"a duration in the wrong unit (mutated field: the bar gauge's unit)", "proven-models", "Model latency", "duration in seconds",
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["unit"] = "ms" }},
	} {
		mut := mutatePanel(t, ds[c.uid], c.panel, c.mutate)
		var found bool
		for _, problem := range lintVisualBar(mut) {
			found = found || strings.Contains(problem, c.want)
		}
		if !found {
			t.Errorf("the control did not turn red for %s: want a problem containing %q, got %v", c.name, c.want, lintVisualBar(mut))
		}
	}
	// A time range that does not fit the run. Mutated field: the default range of the Nightshift dashboard.
	var d map[string]any
	_ = json.Unmarshal(ds["proven-nightshift"], &d)
	d["time"].(map[string]any)["from"] = "now-6h"
	mut, _ := json.Marshal(d)
	red := false
	for _, problem := range lintVisualBar(mut) {
		red = red || strings.Contains(problem, "does not fit")
	}
	if !red {
		t.Error("the control did not turn red: a six-hour range passed for a run of twenty minutes")
	}
}
