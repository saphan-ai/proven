package lab

// The visual bar of the six dashboards (the owner's, 26-27 Oct: a five-second read, one sentence on top, colour always with a
// word, numbers that agree, no empty panel without its reason in words, no cut label), held as far as it can be held without
// a browser. Each rule is one a live run broke: a duration 10^6 too large, a "no data" sentence in every table cell, a bar
// chart that was an empty box or a grey bar with a zero, "Value #A" for a name, "8 of 5 flagged", a table that hid a row, a
// time range that put every review in one bar. How a panel looks is still checked by the screenshots of a live make up; what
// is held here is the JSON that makes it look so.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
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
	minBarTextSize     = 16
)

type stylePanel struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Desc    string `json:"description"`
	GridPos struct{ W, H int }
	Options struct {
		CellHeight    string `json:"cellHeight"`
		NamePlacement string `json:"namePlacement"`
		Text          struct {
			TitleSize int `json:"titleSize"`
			ValueSize int `json:"valueSize"`
		} `json:"text"`
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
	Transformations []struct {
		ID      string          `json:"id"`
		Options json.RawMessage `json:"options"`
	} `json:"transformations"`
}

type styleDoc struct {
	Time   struct{ From string }
	Panels []stylePanel `json:"panels"`
}

var (
	minutesRe   = regexp.MustCompile(`^(?:now-)?(\d+)([mhd])$`)
	ofNRe       = regexp.MustCompile(`^suffix: of (\d+) `)
	pluralWords = []string{"findings", "reviews", "claims", "calls", "agents", "receipts", "alarms", "models", "runs", "steps"}
	badStatBase = map[string]bool{"red": true, "yellow": true, "orange": true}
	// A review is named by its pull request: one instance id each, "pr<number>-<fixture>" (lab/seed/seed.sh).
	perPullRequestRe = regexp.MustCompile(`instance=~"pr\[0-9\]\+-`)
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

// tableRowsFit is the least panel height, in grid units (30 px and an 8 px gap), that shows rows rows, the header and the
// title with room to spare: the same sum as table_h in lab/grafana/build_dashboards.py.
func tableRowsFit(rows int) int { return int(math.Ceil(float64((rows+1)*34+58) / 38)) }

// demoPullRequests is the number of demo pull requests in the manifest: what "of N" on a dashboard must say.
func demoPullRequests() int {
	f, err := os.Open(filepath.Join(root, "lab", "seed", "prs.tsv"))
	if err != nil {
		return -1
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			n++
		}
	}
	return n
}

func hasMapping(p stylePanel, key string) bool {
	for _, m := range p.FieldConfig.Defaults.Mappings {
		if strings.Contains(string(m), `"`+key+`":`) {
			return true
		}
	}
	return false
}

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
	prs := demoPullRequests()
	for i, p := range d.Panels {
		if n := len([]rune(p.Title)); n > p.GridPos.W*6 {
			add("panel %q has %d characters, which a %d-column panel cuts at 1920 px (at most %d)", p.Title, n, p.GridPos.W, p.GridPos.W*6)
		}
		def := p.FieldConfig.Defaults
		// No panel shows a raw series name as its own: "Value #A" is what Grafana calls a series it was not told how to name.
		for _, s := range []string{p.Title, p.Desc, def.DisplayName} {
			if strings.Contains(s, "Value #") {
				add("panel %q shows the name %q, which means nothing to a reader", p.Title, "Value #…")
			}
		}
		for _, tg := range p.Targets {
			if strings.Contains(tg.LegendFormat, "Value #") {
				add("panel %q names a series %q, which means nothing to a reader", p.Title, "Value #…")
			}
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
				if regexp.MustCompile(`\b`+w+`\b`).MatchString(unit) && !hasMapping(p, "1") {
					add("stat %q puts %q after a number and has no sentence for 1, which reads \"1 %s\"", p.Title, w, w)
				}
			}
			// "N of M": M is the number of demo pull requests, and N counts pull requests, one per instance id, so N cannot pass M.
			if m := ofNRe.FindStringSubmatch(def.Unit); m != nil {
				if m[1] != fmt.Sprint(prs) {
					add("stat %q says \"of %s\" but the lab seeds %d pull requests: its denominator is not what it counts", p.Title, m[1], prs)
				}
				for _, tg := range p.Targets {
					if !strings.Contains(tg.Expr, "count(") || !perPullRequestRe.MatchString(tg.Expr) {
						add("stat %q does not count one review per pull request (count over instance ids pr<number>-<fixture>), so its numerator can pass its denominator", p.Title)
					}
				}
			}
		case "bargauge":
			if def.NoValue == "" {
				add("bar gauge %q says nothing when it has no data", p.Title)
			}
			if p.Options.NamePlacement != "top" {
				add("bar gauge %q puts its names at the %q, where a long label is cut", p.Title, p.Options.NamePlacement)
			}
			if p.Options.Text.TitleSize < minBarTextSize || p.Options.Text.ValueSize < minBarTextSize {
				add("bar gauge %q has labels of %d and values of %d px, too small to read in five seconds (at least %d)", p.Title, p.Options.Text.TitleSize, p.Options.Text.ValueSize, minBarTextSize)
			}
			for _, tg := range p.Targets {
				if !strings.Contains(tg.LegendFormat, "{{") {
					add("bar gauge %q names its bars by nothing: its legend has no label", p.Title)
				}
				if strings.Contains(tg.Expr, "vector(0)") {
					add("bar gauge %q draws a zero bar when there is no data: it says why with noValue instead", p.Title)
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
					kept := false
					for _, tr := range p.Transformations {
						kept = kept || tr.ID == "filterFieldsByName"
					}
					if !kept {
						add("table %q carries Tempo's raw columns to the reader: it keeps only the columns it names", p.Title)
					}
				}
			}
			if need := tableRowsFit(rows); rows > 0 && p.GridPos.H < need {
				add("table %q is %d high and hides rows: %d rows need %d", p.Title, p.GridPos.H, rows, need)
			}
			if rows == 0 && p.GridPos.H < tableRowsFit(demoPullRequests()) {
				add("table %q is %d high, too small for its %d rows", p.Title, p.GridPos.H, demoPullRequests())
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
		// The first panel is the sentence the dashboard opens with.
		if i == 0 && (p.Type != "stat" || p.GridPos.H < 4) {
			add("the first panel %q is not the one-sentence stat the dashboard opens with", p.Title)
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
		{"a duration unit that is not the source's (mutated field: the Duration unit of a Tempo table)", "proven-evidence", "Investigator and Prover runs", "Tempo returns nanoseconds",
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
		{"a bar gauge that is silent without data (mutated field: defaults.noValue)", "proven-models", "What Proven refused", "says nothing when it has no data",
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["noValue"] = "" }},
		{"a bar gauge that draws a zero bar (mutated field: the target's expr)", "proven-nightshift", "Why calls failed", "draws a zero bar",
			func(p map[string]any) {
				tg := p["targets"].([]any)[0].(map[string]any)
				tg["expr"] = tg["expr"].(string) + ` or on() label_replace(vector(0), "reason", "none", "", "")`
			}},
		{"a bar named \"Value #A\" (mutated field: the target's legendFormat)", "proven-models", "Calls, by agent and reviewer", "Value #",
			func(p map[string]any) { p["targets"].([]any)[0].(map[string]any)["legendFormat"] = "Value #A" }},
		{"a bar gauge whose labels are too small (mutated field: options.text.titleSize)", "proven-models", "Calls, by agent and reviewer", "too small to read",
			func(p map[string]any) { sub(p, "options", "text")["titleSize"] = 9 }},
		{"a stat that is red at zero (mutated field: the base threshold colour)", "proven-nightshift", "Calls made", "alarm",
			func(p map[string]any) {
				sub(p, "fieldConfig", "defaults", "thresholds")["steps"].([]any)[0].(map[string]any)["color"] = "red"
			}},
		{"a stat with no word (mutated field: unit and mappings)", "proven-models", "Model calls that returned", "no word beside them",
			func(p map[string]any) {
				d := sub(p, "fieldConfig", "defaults")
				d["unit"], d["mappings"] = "short", []any{}
			}},
		{"a plural with no sentence for 1 (mutated field: the stat's mappings)", "proven-review", "Findings this review kept", `reads "1 findings"`,
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["mappings"] = []any{} }},
		{"\"8 of 5\": a denominator that is not the pull requests (mutated field: the stat's unit)", "proven-overview", "Pull requests blocked", "its denominator is not what it counts",
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["unit"] = "suffix: of 5 pull requests blocked" }},
		{"a count that is not one per pull request (mutated field: the target's expr)", "proven-overview", "Reviews finished", "one review per pull request",
			func(p map[string]any) {
				p["targets"].([]any)[0].(map[string]any)["expr"] = `count(proven_agent_state{proven_agent="writer"} == 2)`
			}},
		{"a table too short for its rows (mutated field: gridPos.h)", "proven-review", "The agents, in order", "hides rows",
			func(p map[string]any) { p["gridPos"].(map[string]any)["h"] = 8 }},
		{"a Tempo table with raw columns (mutated field: transformations)", "proven-engines", "What each review ran as", "raw columns",
			func(p map[string]any) { p["transformations"] = []any{} }},
		{"a bar chart that steps too coarsely (mutated field: the target interval)", "proven-nightshift", "Calls over time", "must spread over bars",
			func(p map[string]any) { p["targets"].([]any)[0].(map[string]any)["interval"] = "10m" }},
		{"a title that 1920 px cuts (mutated field: the title)", "proven-overview", "Reviews finished", "cuts at 1920",
			func(p map[string]any) { p["title"] = "Reviews finished, counted by the Writer of every review" }},
		{"a duration in the wrong unit (mutated field: the bar gauge's unit)", "proven-models", "Model latency", "duration in seconds",
			func(p map[string]any) { sub(p, "fieldConfig", "defaults")["unit"] = "ms" }},
		{"a dashboard that does not open with its sentence (mutated field: the first panel's height)", "proven-engines", "H-MEM", "opens with",
			func(p map[string]any) { p["gridPos"].(map[string]any)["h"] = 2 }},
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
