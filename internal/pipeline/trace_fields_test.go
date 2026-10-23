package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
)

// traced is the JSONL shape a reader needs to follow a finding.
// It is local so a missing field fails this test instead of failing to compile.
type traced struct {
	FindingID  string        `json:"finding_id"`
	Kind       string        `json:"kind"`
	Evidence   []tracedProof `json:"evidence"`
	Refutation *tracedRefute `json:"refutation"`
}

type tracedProof struct {
	Kind  string       `json:"kind"`
	Test  *tracedTest  `json:"test"`
	Query *tracedQuery `json:"query"`
}

type tracedTest struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	FailsOnChange bool   `json:"fails_on_change"`
	PassesOnBase  bool   `json:"passes_on_base"`
	Output        string `json:"output"`
}

type tracedQuery struct {
	SQL      string     `json:"sql"`
	Engine   string     `json:"engine"`
	Snapshot string     `json:"snapshot"`
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
}

type tracedRefute struct {
	Attempted bool   `json:"attempted"`
	Refuted   bool   `json:"refuted"`
	Reason    string `json:"reason"`
}

// The trace alone lets a reader follow a finding: a test's path, name, command,
// outcomes and output, a query's SQL, engine, snapshot, columns and rows, and
// each refutation attempt. A finding with no proof stores none.
func TestTraceCarriesEvidenceAndRefutation(t *testing.T) {
	p := buildSample(t, t.TempDir(), nil)
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(d.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	var events []traced
	for _, line := range bytes.Split(raw, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e traced
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		events = append(events, e)
	}

	f1 := oneProof(t, events, "F1")
	if f1.Test == nil || f1.Kind != "test" {
		t.Fatalf("F1 evidence %#v", f1)
	}
	te := f1.Test
	if te.Path != "billing/discount_test.go" || te.Name != "TestDiscounted" ||
		te.Command != "go test -run '^TestDiscounted$' ./billing" ||
		!te.FailsOnChange || !te.PassesOnBase ||
		te.Output != "base: PASS\nchange: FAIL got 60 want 80" {
		t.Fatalf("F1 test %#v", te)
	}

	f2 := oneProof(t, events, "F2")
	if f2.Query == nil || f2.Kind != "query" {
		t.Fatalf("F2 evidence %#v", f2)
	}
	q := f2.Query
	if q.SQL != "SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'" ||
		q.Engine != "local-sample" || q.Snapshot != "sample-2026-10-22" {
		t.Fatalf("F2 query %#v", q)
	}
	if len(q.Columns) != 4 || q.Columns[0] != "package" || q.Columns[1] != "version" || q.Columns[2] != "cve" || q.Columns[3] != "kev" {
		t.Fatalf("columns %#v", q.Columns)
	}
	if len(q.Rows) != 1 || len(q.Rows[0]) != 4 ||
		q.Rows[0][0] != "evil.example/leftpad" || q.Rows[0][1] != "1.2.3" ||
		q.Rows[0][2] != "CVE-2026-0001" || q.Rows[0][3] != "true" {
		t.Fatalf("rows %#v", q.Rows)
	}

	f4 := oneProof(t, events, "F4")
	if f4.Query == nil || f4.Kind != "query" {
		t.Fatalf("F4 evidence %#v", f4)
	}
	kq := f4.Query
	if kq.SQL != "SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'" ||
		kq.Engine != "local-sample" || kq.Snapshot != "sample-2026-10-22" {
		t.Fatalf("F4 query %#v", kq)
	}
	if len(kq.Rows) != 1 || kq.Rows[0][0] != "CVE-2026-0001" || kq.Rows[0][1] != "evil.example" || kq.Rows[0][2] != "leftpad" {
		t.Fatalf("F4 rows %#v", kq.Rows)
	}

	for _, id := range []string{"F1", "F2", "F3", "F4"} {
		line := rawEvent(t, raw, id, "refutation")
		if !bytes.Contains(line, []byte(`"attempted":true`)) || !bytes.Contains(line, []byte(`"refuted":false`)) || !bytes.Contains(line, []byte(`"reason":"no refutation"`)) {
			t.Fatalf("%s refutation %s", id, line)
		}
	}
	f3 := findEvent(t, events, "F3", "evidence")
	if len(f3.Evidence) != 0 {
		t.Fatalf("F3 has no proof to store, got %#v", f3.Evidence)
	}
}

func oneProof(t *testing.T, events []traced, id string) tracedProof {
	t.Helper()
	e := findEvent(t, events, id, "evidence")
	if len(e.Evidence) != 1 {
		t.Fatalf("%s evidence count %d", id, len(e.Evidence))
	}
	return e.Evidence[0]
}

func findEvent(t *testing.T, events []traced, id, kind string) traced {
	t.Helper()
	for _, e := range events {
		if e.FindingID == id && e.Kind == kind {
			return e
		}
	}
	t.Fatalf("no %s event for %s", kind, id)
	return traced{}
}

func rawEvent(t *testing.T, raw []byte, id, kind string) []byte {
	t.Helper()
	for _, line := range bytes.Split(raw, []byte("\n")) {
		if bytes.Contains(line, []byte(`"finding_id":"`+id+`"`)) && bytes.Contains(line, []byte(`"kind":"`+kind+`"`)) {
			return line
		}
	}
	t.Fatalf("no %s event for %s", kind, id)
	return nil
}
