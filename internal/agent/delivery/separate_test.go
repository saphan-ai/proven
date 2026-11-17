package delivery

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	data "github.com/saphan-ai/proven/internal/delivery"
)

func TestEachMessageIsAssessedAloneAndFailureDoesNotDiscardOtherSources(t *testing.T) {
	m := data.Metadata{Description: "Preserve existing behavior Alpha", Commits: []data.Commit{{SHA: "a", Message: "Preserve existing behavior Bravo"}, {SHA: "b", Message: "Preserve existing behavior Charlie"}}, Comments: []data.Comment{{Kind: "pr_comment", ID: 7, Body: "Preserve existing behavior Delta"}, {Kind: "review_comment", ID: 8, Body: "Preserve existing behavior Echo"}}}
	calls := 0
	gen := modelFunc(func(_ context.Context, _, prompt string) (string, error) {
		calls++
		var request struct {
			Source data.Source `json:"source"`
		}
		if err := json.Unmarshal([]byte(strings.SplitN(prompt, "DATA:\n", 2)[1]), &request); err != nil {
			t.Fatal(err)
		}
		for _, other := range data.Sources(m) {
			if other.ID != request.Source.ID && strings.Contains(prompt, other.Text) {
				t.Fatal("another source leaked into this assessment")
			}
		}
		claim := data.Claim{Requirement: "Preserve existing behavior", SourceKind: request.Source.Kind, SourceIndex: request.Source.Index, SourceQuote: request.Source.Text, Status: "unclear", Reason: "This source needs acceptance coverage", SourceSHA: "invented", ID: "invented", SourceID: "invented"}
		if request.Source.Kind == "commit" && request.Source.Index == 0 {
			claim.SourceIndex = 1
		}
		b, _ := json.Marshal(struct {
			Claims []data.Claim `json:"claims"`
		}{[]data.Claim{claim}})
		return string(b), nil
	})
	r, err := Assess(context.Background(), gen, m, nil, "snapshot", nil)
	if err == nil || r.Complete || calls != 5 || len(r.Assessments) != 5 || len(r.Claims) != 4 {
		t.Fatalf("%+v err=%v calls=%d", r, err, calls)
	}
	if r.Assessments[1].Complete || !r.Assessments[2].Complete || !r.Assessments[4].Complete {
		t.Fatal("source failure was hidden or stopped unrelated sources")
	}
	seen := map[string]bool{}
	for _, c := range r.Claims {
		if seen[c.ID] || c.ID == "invented" || c.SourceID == "invented" || c.SourceSHA == "invented" {
			t.Fatal("untrusted or merged identity")
		}
		seen[c.ID] = true
	}
}

func TestDeadlineKeepsEveryUnassessedSourceVisible(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := data.Metadata{Description: "Describe the requested behavior", Commits: []data.Commit{{SHA: "a", Message: "Describe the requested behavior"}}}
	r, err := Assess(ctx, modelFunc(func(context.Context, string, string) (string, error) {
		t.Fatal("called model after cancellation")
		return "", nil
	}), m, nil, "snapshot", nil)
	if err == nil || r.Complete || len(r.Assessments) != 2 || r.Assessments[0].Complete || r.Assessments[1].Issue == "" {
		t.Fatalf("%+v %v", r, err)
	}
}
