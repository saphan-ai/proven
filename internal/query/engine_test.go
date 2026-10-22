package query_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

const sampleSQL = `
CREATE TABLE advisories (package TEXT, version TEXT, cve TEXT, kev TEXT);
INSERT INTO advisories VALUES ('evil.example/leftpad', '1.2.3', 'CVE-2026-0001', 'true');
`

func TestLocalEngineReturnsTheSampleRow(t *testing.T) {
	db, err := query.LoadSQL(sampleSQL)
	if err != nil {
		t.Fatal(err)
	}
	cols, rows, err := db.Query("SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0][0] != "evil.example/leftpad" || rows[0][2] != "CVE-2026-0001" || cols[3] != "kev" {
		t.Fatalf("got %#v %#v", cols, rows)
	}
	_, empty, err := db.Query("SELECT package FROM advisories WHERE package = 'billing/note' AND version = 'change'")
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("want no rows, got %#v", empty)
	}
}

func TestLiveQueryIsDown(t *testing.T) {
	_, err := query.Live{}.Query(context.Background(), "SELECT package FROM advisories")
	d, ok := service.As(err)
	if !ok || d.Service != "query" {
		t.Fatalf("got %v", err)
	}
}
