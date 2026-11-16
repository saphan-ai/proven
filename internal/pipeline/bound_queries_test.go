package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

func TestUnrelatedAdvisoryQueryNeverContactsAnEmptyDatabase(t *testing.T) {
	inner := &askedRunner{}
	q := boundQueries{Runner: inner}
	for _, template := range []string{querytpl.OSVByPackage, querytpl.OSVByAdvisory, querytpl.KEVByCVE,
		querytpl.EPSSByCVE, querytpl.MalByPackage, querytpl.ScorecardByPackage, querytpl.CVSSByCVE, querytpl.SPDXByPackage} {
		_, err := q.Run(context.Background(), querytpl.Request{Template: template}, querytpl.Relations{})
		var refused *querytpl.Refusal
		if !errors.As(err, &refused) {
			t.Fatalf("%s: expected refusal, got %v", template, err)
		}
	}
	if len(inner.asked) != 0 {
		t.Fatalf("unrelated queries reached the database: %v", inner.asked)
	}
	if _, err := q.Run(context.Background(), querytpl.Request{Template: "extension-query"}, querytpl.Relations{}); err != nil || len(inner.asked) != 1 {
		t.Fatalf("an extension query was blocked: %v, %v", inner.asked, err)
	}
}

func TestAffectedDependencyStillReachesItsEvidenceDatabase(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	for side, body := range map[string]string{
		"base":   "module example.com/app\n\ngo 1.26\n",
		"change": "module example.com/app\n\ngo 1.26\n\nrequire example.com/widget v1.2.3\n",
	} {
		if err := os.Mkdir(filepath.Join(dir, side), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, side, "go.mod"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "evidence.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Register(ctx, store.Table{Name: querytpl.DefaultOSVTable, Columns: querytpl.OSVPackageColumns,
		Rows: [][]string{{"GHSA-example", "", "Go", "example.com/widget", "1.0.0", "2.0.0", "", "HIGH", "2026-01-01"}}, Snapshot: "fixture"}); err != nil {
		t.Fatal(err)
	}
	inner := storeRunner{s: db}
	checked, err := osv.Check(ctx, inner, dir, "", "fixture")
	if err != nil || checked.Down != nil || len(checked.Candidates) != 1 {
		t.Fatalf("dependency scan: %+v, %v", checked, err)
	}
	q := boundQueries{Runner: inner, facts: checked.Facts}
	res, err := q.Run(ctx, querytpl.Request{Template: querytpl.OSVByPackage,
		Params: map[string]string{querytpl.ParamEcosystem: "Go", querytpl.ParamPackage: "example.com/widget"}}, querytpl.Relations{})
	if err != nil || len(res.Rows) != 1 {
		t.Fatalf("affected dependency query was blocked: %+v, %v", res, err)
	}
}
