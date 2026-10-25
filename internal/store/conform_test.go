package store

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestStoresAgreeOnOneScenario proves one template, run on every store this
// wagon builds, returns the same form, the same query text, the same columns
// and the same rows (ADR-0047). Zetaris is a recorded HTTP server. The local
// store is the real file. The mutated field of the negative control is the
// local store's result rows: dropping one row makes this comparison fail.
func TestStoresAgreeOnOneScenario(t *testing.T) {
	when := time.Date(2026, 10, 24, 7, 0, 0, 0, time.UTC)
	cols := []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}
	row := []string{"CVE-2021-44228", "Apache", "Log4j2", "2021-12-10", "Known"}
	table := Table{
		Name:     "px_evidence.px_kev_vulns",
		Columns:  cols,
		Rows:     [][]string{row},
		Snapshot: when.UTC().Format(time.RFC3339),
	}
	req := querytpl.Request{
		Template: querytpl.KEVByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"},
	}
	ctx := context.Background()

	local, err := OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	local.SetClock(func() time.Time { return when })
	if err := local.Register(ctx, table); err != nil {
		t.Fatal(err)
	}

	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		if r.URL.Path != "/sql-editor/sqls/run" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"headers": cols,
			"data":    [][]string{row},
		})
	}))
	defer srv.Close()
	z := NewZetaris(query.Settings{
		BaseURL: srv.URL, APIKey: "conform-key", OrgID: "1",
		RowLimit: 100, Timeout: time.Second,
	})
	z.SetClock(func() time.Time { return when })
	// Register records the name and must not dial. ADR-0023: Proven's key only reads.
	if err := z.Register(ctx, table); err != nil {
		t.Fatal(err)
	}
	if posts != 0 {
		t.Fatalf("register sent %d requests", posts)
	}
	if err := z.Register(ctx, table); err == nil {
		t.Fatal("a second register of the same table was accepted")
	}

	gotLocal, err := local.Run(ctx, req, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	gotZ, err := z.Run(ctx, req, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	if posts != 1 {
		t.Fatalf("zetaris calls %d, want the one query", posts)
	}
	if gotLocal.Engine != "local" || gotZ.Engine != "zetaris" {
		t.Fatalf("engines %s %s", gotLocal.Engine, gotZ.Engine)
	}
	if gotLocal.Form != querytpl.FormSQL || gotZ.Form != querytpl.FormSQL {
		t.Fatalf("forms %s %s", gotLocal.Form, gotZ.Form)
	}
	if gotLocal.Query != gotZ.Query || gotLocal.Query == "" {
		t.Fatalf("queries %q %q", gotLocal.Query, gotZ.Query)
	}
	if !reflect.DeepEqual(gotLocal.Columns, gotZ.Columns) || !reflect.DeepEqual(gotLocal.Rows, gotZ.Rows) {
		t.Fatalf("local %#v zetaris %#v", gotLocal, gotZ)
	}
	if len(gotLocal.Rows) != 1 || gotLocal.Rows[0][0] != "CVE-2021-44228" {
		t.Fatalf("rows %#v", gotLocal.Rows)
	}
	if gotLocal.Snapshot != gotZ.Snapshot || gotLocal.Snapshot != "2026-10-24T07:00:00Z" {
		t.Fatalf("snapshots %s %s", gotLocal.Snapshot, gotZ.Snapshot)
	}
}

// TestLocalRunDoesNotDial proves the local store answers with the server
// closed. The mutated field is the server: it is closed before Run.
func TestLocalRunDoesNotDial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the local store dialed")
	}))
	srv.Close()
	local, err := OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	table := Table{
		Name:    "px_evidence.px_kev_vulns",
		Columns: []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"},
		Rows:    [][]string{{"CVE-2021-44228", "Apache", "Log4j2", "2021-12-10", "Known"}},
	}
	if err := local.Register(context.Background(), table); err != nil {
		t.Fatal(err)
	}
	got, err := local.Run(context.Background(), querytpl.Request{
		Template: querytpl.KEVByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"},
	}, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Form != querytpl.FormSQL || len(got.Rows) != 1 {
		t.Fatalf("%#v", got)
	}
}
