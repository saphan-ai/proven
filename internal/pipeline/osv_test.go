package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// A dependency the change adds is reported from the OSV package table with the
// range it falls in and the advisory it does not match, its KEV and EPSS rows
// as further query evidence, and the version comparison on the trace.
func TestSampleReviewReportsTheOSVRange(t *testing.T) {
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	c, ok := kept(d, "osv:evil.example/leftpad")
	if !ok {
		t.Fatal("osv finding missing")
	}
	if c.Path != "go.mod" || c.StartLine != 5 || c.RuleID != "osv-affected-package" {
		t.Fatalf("%#v", c.Risk)
	}
	// Sample mode starts no container, so the Go finding is not measured and
	// is never called unreached (ADR-0043).
	if c.Ecosystem != "Go" || c.ReachLabel != candidate.ReachNotMeasured || c.CallPath != "" {
		t.Fatalf("reach %#v", c)
	}
	if !strings.Contains(c.Message, "Reachability: not measured") || strings.Contains(c.Message, "imported but not reached") {
		t.Fatalf("message %s", c.Message)
	}
	if !strings.Contains(c.Message, "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4") {
		t.Fatalf("message %s", c.Message)
	}
	if !strings.Contains(c.Message, "GHSA-px-sample-0002 does not match") {
		t.Fatalf("message %s", c.Message)
	}
	if len(c.Evidence) != 3 || c.Evidence[0].Query == nil || c.Evidence[1].Query == nil || c.Evidence[2].Query == nil {
		t.Fatalf("evidence %#v", c.Evidence)
	}
	q := c.Evidence[0].Query
	if q.Engine != "local-sample" || q.Snapshot != "sample-2026-10-22" || len(q.Rows) != 2 {
		t.Fatalf("query %#v", q)
	}
	if !strings.Contains(q.SQL, "px_osvfile.px_osv_pkg") || !strings.Contains(q.SQL, "ecosystem = 'Go'") || !strings.Contains(q.SQL, "package = 'evil.example/leftpad'") {
		t.Fatalf("sql %s", q.SQL)
	}
	if c.Evidence[1].Query.Rows[0][0] != "CVE-2026-0001" {
		t.Fatalf("kev %#v", c.Evidence[1].Query)
	}
	ep := c.Evidence[2].Query
	if len(ep.Rows) != 1 || ep.Rows[0][0] != "CVE-2026-0001" || ep.Rows[0][1] != "0.50000" {
		t.Fatalf("epss %#v", ep)
	}
	raw := traceRaw(t, d.TracePath)
	if !strings.Contains(string(raw), "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4") {
		t.Fatal("trace has no comparison")
	}
	if !strings.Contains(string(raw), "govulncheck was not run: sample mode starts no container") {
		t.Fatal("trace does not name the unrun tool")
	}
}
