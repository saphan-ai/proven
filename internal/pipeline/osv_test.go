package pipeline

import (
	"context"
	"strings"
	"testing"
)

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
	if !strings.Contains(c.Message, "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4") {
		t.Fatalf("message %s", c.Message)
	}
	if !strings.Contains(c.Message, "GHSA-px-sample-0002 does not match") {
		t.Fatalf("message %s", c.Message)
	}
	if len(c.Evidence) != 2 || c.Evidence[0].Query == nil || c.Evidence[1].Query == nil {
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
	raw := traceRaw(t, d.TracePath)
	if !strings.Contains(string(raw), "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4") {
		t.Fatal("trace has no comparison")
	}
}
