package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
)

func TestSQLitePersistsRegisteredRowsAndSnapshotIdentity(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "evidence.sqlite")
	db, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	table := Table{Name: querytpl.DefaultEPSSView, Columns: []string{"cve_id", "epss", "percentile", "score_date"}, Rows: [][]string{{"CVE-2021-44228", "0.9", "0.99", "2026-11-16"}}, Snapshot: "2026-11-16"}
	if err = db.Register(ctx, table); err != nil {
		t.Fatal(err)
	}
	req := querytpl.Request{Template: querytpl.EPSSByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}
	a, err := db.Run(ctx, req, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Rows) != 1 || a.Rows[0][1] != "0.9" || !strings.Contains(a.Snapshot, "sha256:") {
		t.Fatalf("%+v", a)
	}
	db.Close()
	db, err = OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	b, err := db.Run(ctx, req, querytpl.Relations{})
	if err != nil || a.Snapshot != b.Snapshot {
		t.Fatalf("restart: %+v %v", b, err)
	}
	if err = db.Register(ctx, table); err == nil {
		t.Fatal("duplicate register accepted")
	}
	table.Rows[0][1] = "0.8"
	if err = db.Replace(ctx, table); err != nil {
		t.Fatal(err)
	}
	c, err := db.Run(ctx, req, querytpl.Relations{})
	if err != nil || c.Snapshot == a.Snapshot || c.Rows[0][1] != "0.8" {
		t.Fatalf("replacement: %+v %v", c, err)
	}
	table.Rows[0] = []string{"broken"}
	if err = db.Replace(ctx, table); err == nil {
		t.Fatal("invalid replacement accepted")
	}
	d, err := db.Run(ctx, req, querytpl.Relations{})
	if err != nil || d.Snapshot != c.Snapshot {
		t.Fatal("failed replacement changed source")
	}
	req.Params[querytpl.ParamCVEID] = "CVE-2021-44228' OR 1=1"
	if _, err = db.Run(ctx, req, querytpl.Relations{}); err == nil {
		t.Fatal("unvalidated query accepted")
	}
}

func TestSQLiteMissingDatasetIsNotClean(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "empty.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Run(context.Background(), querytpl.Request{Template: querytpl.EPSSByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}, querytpl.Relations{})
	if err == nil {
		t.Fatal("absent dataset read as empty answer")
	}
}

func TestPostgresConnectionErrorsDoNotExposeCredentials(t *testing.T) {
	_, err := OpenPostgres("postgres://user:private-password@", nil)
	if err == nil || strings.Contains(err.Error(), "private-password") {
		t.Fatalf("%v", err)
	}
}
