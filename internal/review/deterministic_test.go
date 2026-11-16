package review

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

func TestDeterministicReviewNeedsNoModelOrSponsorAccount(t *testing.T) {
	repo, base, head := twoCommits(t)
	dir := t.TempDir()
	built := generator.LiveBuilt()
	r, err := Open(Options{LogDir: filepath.Join(dir, "logs"), Env: func(k string) string {
		if k == "PROVEN_EVIDENCE_FILE" {
			return filepath.Join(dir, "evidence.sqlite")
		}
		return ""
	}, testRun: quietRun{}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := r.Review(context.Background(), Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil || len(res.Models) != 0 || len(res.FailedServices) != 0 || res.Doubled || generator.LiveBuilt() != built {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestDeterministicSQLiteReviewFindsDependencyRegression(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	file := filepath.Join(dir, "evidence.sqlite")
	db, err := store.OpenSQLite(file)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, table := range []store.Table{
		{Name: querytpl.DefaultKEVView, Columns: []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}, Rows: [][]string{{"CVE-2026-4242", "example.com", "widget", "2026-04-01", "Unknown"}}, Snapshot: "fixture"},
		{Name: querytpl.DefaultOSVTable, Columns: []string{"id", "aliases", "ecosystem", "package", "introduced", "fixed", "last_affected", "severity", "published"}, Rows: [][]string{{"GHSA-w19-sample-0001", "CVE-2026-4242", "Go", "example.com/widget", "1.3.0", "1.5.0", "", "HIGH", "2026-04-01"}}, Snapshot: "fixture"},
		{Name: querytpl.DefaultEPSSView, Columns: []string{"cve_id", "epss", "percentile", "score_date"}, Rows: [][]string{{"CVE-2026-4242", "0.42", "0.91", "2026-10-21"}}, Snapshot: "fixture"},
	} {
		if err := db.Register(ctx, table); err != nil {
			t.Fatal(err)
		}
	}
	r, err := Open(Options{LogDir: filepath.Join(dir, "logs"), Env: func(k string) string {
		if k == "PROVEN_EVIDENCE_FILE" {
			return file
		}
		return ""
	}, testRun: quietRun{}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := r.Review(ctx, Request{Target: filepath.Join(moduleRoot(t), "testdata", "fixture", "osv-go-module")})
	if err != nil || len(res.Findings) != 1 || len(res.Models) != 0 || len(res.FailedServices) != 0 {
		t.Fatalf("%+v %v", res, err)
	}
	if res.Findings[0].Evidence[0].Query.Engine != "sqlite" || !res.Findings[0].Evidence[0].Query.Bound {
		t.Fatalf("%+v", res.Findings[0])
	}
}
