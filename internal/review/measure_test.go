package review

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
)

// advisoryQuery answers the OSV query for example.com/lib with one advisory that affects v1.2.0, so the
// dependency check proves a finding from data alone, with no model and no test. Every other query has no rows.
type advisoryQuery struct{ countingQuery }

// Query records sql and answers it.
func (q *advisoryQuery) Query(ctx context.Context, sql string) (query.Result, error) {
	res, err := q.countingQuery.Query(ctx, sql)
	if strings.Contains(sql, "px_osv") && strings.Contains(sql, "example.com/lib") {
		res.Rows = [][]string{{"GO-2026-0001", "CVE-2026-0001", "Go", "example.com/lib", "0", "1.3.0", "", "HIGH", "2026-01-01"}}
	}
	return res, err
}

// measured is one review's cost and output.
type measured struct {
	calls    int
	wall     time.Duration
	findings []string
	class    string
	doubled  bool
}

// measure runs one scenario once and reads its cost: model calls on both generators, wall time, and
// the ids of the findings it kept.
func measure(t *testing.T, cfg fileSettings, target string, withAdvisory bool) measured {
	t.Helper()
	rig := newRig()
	if withAdvisory {
		rig.q = &advisoryQuery{}
	}
	start := time.Now()
	res := rig.reviewWith(t, cfg, target)
	m := measured{calls: rig.calls(), wall: time.Since(start), class: res.Class, doubled: res.Doubled}
	for _, f := range res.Findings {
		m.findings = append(m.findings, f.RuleID+"@"+f.Location.Path)
	}
	sort.Strings(m.findings)
	return m
}

// TestPlanMeasurement is ADR-0049 point 5's measurement on scripted scenarios: for each, the model calls
// and the wall time of one review with the plan and without it, and whether any finding the full plan
// keeps is lost by the plan. The generator is scripted, so the call counts are exact and the times are
// Proven's own overhead with an instant model, never a live model's latency. It fails if a plan loses a
// finding, or costs more calls than the full plan on a change it is meant to cut.
func TestPlanMeasurement(t *testing.T) {
	scenarios := []struct {
		name     string
		target   func(*testing.T) string
		advisory bool
		cut      bool
	}{
		{"documentation only", docsOnly, false, true},
		{"dependency only, one advisory", dependencyOnly, true, true},
		{"security-sensitive session code", sensitive, false, false},
		{"feature: one modified Go file", func(t *testing.T) string {
			return trees(t, map[string][2]string{"internal/x/run.go": {"package x\n", "package x\n\nvar N = 1\n"}})
		}, false, false},
		{"ambiguous: documentation beside code", func(t *testing.T) string {
			return trees(t, map[string][2]string{
				"docs/guide.md":     {"old\n", "new\n"},
				"internal/x/run.go": {"package x\n", "package x\n\nvar N = 1\n"},
			})
		}, false, false},
	}
	var table strings.Builder
	table.WriteString("| scenario (scripted generator) | model calls without plan | model calls with plan | wall without | wall with | findings without | findings with | lost |\n")
	table.WriteString("|---|---|---|---|---|---|---|---|\n")
	for _, s := range scenarios {
		off := defaultSettings()
		on := defaultSettings()
		on.ScoutIntent = true
		without := measure(t, off, s.target(t), s.advisory)
		with := measure(t, on, s.target(t), s.advisory)
		var lost []string
		for _, f := range without.findings {
			if !containsName(with.findings, f) {
				lost = append(lost, f)
			}
		}
		if len(lost) > 0 {
			t.Errorf("%s: the plan lost %v", s.name, lost)
		}
		if s.cut && with.calls >= without.calls {
			t.Errorf("%s: the plan did not cut model calls (%d with, %d without)", s.name, with.calls, without.calls)
		}
		fmt.Fprintf(&table, "| %s | %d | %d | %s | %s | %d | %d | %d |\n", s.name, without.calls, with.calls,
			without.wall.Round(time.Millisecond), with.wall.Round(time.Millisecond), len(without.findings), len(with.findings), len(lost))
	}
	t.Log("\n" + table.String())
}
