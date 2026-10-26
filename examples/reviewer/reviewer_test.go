package reviewer_test

import (
	"context"
	"errors"
	"testing"

	eg "github.com/saphan-ai/proven/examples/reviewer"
	ir "github.com/saphan-ai/proven/internal/review"
)

// The assignment proves Table compiles against the real interface:
// examples/reviewer.Table and internal/review.Reviewer agree on one
// method set.
var _ ir.Reviewer = eg.Table{}

// fixedResult is the verdict the table returns for its target: one finding,
// proved by a test that fails on the change and passes on the base.
func fixedResult() ir.Result {
	return ir.Result{
		Findings: []ir.Finding{
			{
				ID:       "f1",
				RuleID:   "example-rule",
				Title:    "the cap is ignored",
				Message:  "discount.go raises the cap without a proving test",
				Severity: ir.SeverityError,
				Location: ir.Location{Path: "discount.go", StartLine: 12, EndLine: 12},
				Evidence: []ir.Evidence{
					{
						Kind: ir.EvidenceTest,
						Test: &ir.TestEvidence{
							Path:          "discount_test.go",
							Name:          "TestCap",
							Command:       "go test -run TestCap ./...",
							FailsOnChange: true,
							PassesOnBase:  true,
							Output:        "discount_test.go:12: cap exceeded",
						},
					},
				},
				ProofLevel: ir.ProvedByTest,
				Models:     []string{"default"},
			},
		},
		ProofCounts: map[ir.ProofLevel]int{
			ir.ProvedByTest:    1,
			ir.ProvedByTool:    0,
			ir.ProvedByData:    0,
			ir.AgreedNotProven: 0,
		},
	}
}

// TestTableReviewsItsTarget proves one answer: a known target returns its
// recorded result, unchanged, with no error.
func TestTableReviewsItsTarget(t *testing.T) {
	table, err := eg.NewTable([]eg.Row{
		{Target: "discount.go", Result: fixedResult()},
	})
	if err != nil {
		t.Fatalf("NewTable: %v", err)
	}
	got, err := table.Review(context.Background(), ir.Request{Target: "discount.go"})
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if len(got.Findings) != 1 || got.Findings[0].ID != "f1" {
		t.Fatalf("got %d findings, want one named f1: %+v", len(got.Findings), got.Findings)
	}
	if got.ProofCounts[ir.ProvedByTest] != 1 {
		t.Fatalf("got proof counts %v, want one proved-by-test", got.ProofCounts)
	}
}

// TestTableRefusesAnUnknownTarget proves the refusal: a target the table has
// not recorded returns ErrUnknownTarget and no result.
func TestTableRefusesAnUnknownTarget(t *testing.T) {
	table, err := eg.NewTable([]eg.Row{
		{Target: "discount.go", Result: fixedResult()},
	})
	if err != nil {
		t.Fatalf("NewTable: %v", err)
	}
	got, err := table.Review(context.Background(), ir.Request{Target: "other.go"})
	if err == nil {
		t.Fatal("an unknown target must be refused")
	}
	if !errors.Is(err, eg.ErrUnknownTarget) {
		t.Fatalf("got %v, want %v", err, eg.ErrUnknownTarget)
	}
	if len(got.Findings) != 0 {
		t.Fatalf("got %d findings, want none", len(got.Findings))
	}
}

// TestTableRefusesACanceledContext proves the interface's error path: a
// canceled context returns an error and no result.
func TestTableRefusesACanceledContext(t *testing.T) {
	table, err := eg.NewTable([]eg.Row{
		{Target: "discount.go", Result: fixedResult()},
	})
	if err != nil {
		t.Fatalf("NewTable: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = table.Review(ctx, ir.Request{Target: "discount.go"})
	if err == nil {
		t.Fatal("a canceled context must be refused")
	}
}

// TestTableRefusesADuplicateTarget proves the builder refuses two rows for one target.
func TestTableRefusesADuplicateTarget(t *testing.T) {
	_, err := eg.NewTable([]eg.Row{
		{Target: "discount.go", Result: fixedResult()},
		{Target: "discount.go", Result: fixedResult()},
	})
	if err == nil {
		t.Fatal("a duplicate target must be refused")
	}
}
