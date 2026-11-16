package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// An explicitly supplied disposable database enables the real PostgreSQL contract.
func TestPostgresIntegration(t *testing.T) {
	url := os.Getenv("PROVEN_TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("set PROVEN_TEST_POSTGRES_URL to a disposable PostgreSQL database")
	}
	db, err := OpenPostgres(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	schema := fmt.Sprintf("proven_test_%d", time.Now().UnixNano())
	defer db.db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+db.quote(schema)+" CASCADE")
	table := Table{Name: schema + ".epss", Columns: []string{"cve_id", "epss", "percentile", "score_date"}, Rows: [][]string{{"CVE-2021-44228", "0.9", "0.99", "2026-11-16"}}, Snapshot: "integration fixture"}
	if err := db.Register(ctx, table); err != nil {
		t.Fatal(err)
	}
	result, err := db.Run(ctx, querytpl.Request{Template: querytpl.EPSSByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}}, querytpl.Relations{EPSSView: table.Name})
	if err != nil || len(result.Rows) != 1 || result.Rows[0][1] != "0.9" || result.Snapshot == "" {
		t.Fatalf("%+v %v", result, err)
	}
}
