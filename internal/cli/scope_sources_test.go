package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/source"
)

func TestScopeSourcesAndFreshnessSurviveEveryReportAndMarkdownBounds(t *testing.T) {
	r := review.Result{Delivery: &delivery.Result{Mode: "model", Complete: true, Claims: []delivery.Claim{}}, SourceChecks: []source.Observation{{Adapter: "spdx", Package: "example", Version: "1.0.0", Complete: true, Matched: true, SQL: "registered query", Snapshot: "content-id"}}, SourceRefresh: []source.Refresh{{ID: "spdx", Origin: "https://api.deps.dev/v3", CheckedAt: "2026-11-17T00:00:00Z", Status: "partial", Rows: 1, Issue: "One mapping unavailable"}}}
	for i := 0; i < 150; i++ {
		c := delivery.Claim{ID: fmt.Sprintf("D%d.1", i+1), SourceKind: "pr_comment", SourceIndex: i, Requirement: strings.Repeat("A scope requirement ", 50), Status: "unclear", Reason: strings.Repeat("Bounded evidence ", 80)}
		r.Delivery.Claims = append(r.Delivery.Claims, c)
		r.Delivery.Assessments = append(r.Delivery.Assessments, delivery.Assessment{Source: delivery.Source{Kind: "pr_comment", CommentID: int64(i + 1), Author: "<script>@someone", URL: "javascript:alert(1)", Text: "Keep this source separate from its neighboring comments"}, Complete: true, SemanticStatus: "assessed", Description: delivery.CheckDescription("Keep all declarations separate"), Claims: []delivery.Claim{c}})
	}
	var out bytes.Buffer
	if err := (jsonFormatter{}).Write(&out, r); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc["source_checks"] == nil || doc["source_refresh"] == nil || len(doc["delivery"].(map[string]any)["assessments"].([]any)) != 150 {
		t.Fatal("lost separate sources")
	}
	md := mdFit(r)
	if len(md) > mdMaxComment || !strings.Contains(md, "more source assessments") || strings.Contains(md, "javascript:") {
		t.Fatal("unsafe, unannounced or oversized Markdown")
	}
	html := htmlPage(r)
	if !strings.Contains(html, htmlText("PR comment #150")) || !strings.Contains(html, "Deterministic data checks") || !strings.Contains(html, "One mapping unavailable") || strings.Contains(html, "<script>@someone") {
		t.Fatal("HTML lost or evaluated source data")
	}
	if log := sarifFrom(r); log.Runs[0].Properties["source_checks"] == nil || log.Runs[0].Properties["source_refresh"] == nil {
		t.Fatal("SARIF lost source observations")
	}
}

func TestRefreshRefusesWritesThroughLinksIntoTarget(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(outside, "cache")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	if err := outsideRefreshRoot(root, filepath.Join(link, "nested", "db.sqlite")); err == nil {
		t.Fatal("followed cache link into target")
	}
	if err := outsideRefreshRoot(root, filepath.Join(outside, "new", "cache")); err != nil {
		t.Fatal(err)
	}
}
