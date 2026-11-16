package delivery

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
	data "github.com/saphan-ai/proven/internal/delivery"
)

type modelFunc func(context.Context, string, string) (string, error)

func (f modelFunc) Complete(c context.Context, a, p string) (string, error) { return f(c, a, p) }

func TestAssessmentValidatesMetadataAndCodeRatherThanModelConfidence(t *testing.T) {
	m := data.Metadata{Description: "Reject negative amounts", Commits: []data.Commit{{SHA: "abc", Message: "Reject negative amounts"}}}
	files := []change.File{{Path: "price.py", Change: []byte("def price(amount):\n    if amount < 0:\n        raise ValueError('negative')\n")}}
	valid := data.Claim{Requirement: "Reject negative amounts", SourceKind: "description", SourceQuote: m.Description, Status: "code_indication", Reason: "A guard is present; not executed here.", Citations: []data.Citation{{Path: "price.py", StartLine: 2, EndLine: 3, Quote: "    if amount < 0:\n        raise ValueError('negative')"}}}
	for _, tc := range []struct {
		name   string
		mutate func(*data.Claim)
		want   bool
	}{
		{"valid", func(*data.Claim) {}, true},
		{"invented requirement", func(c *data.Claim) { c.SourceQuote = "Support every currency" }, false},
		{"wrong line", func(c *data.Claim) { c.Citations[0].StartLine = 1 }, false},
		{"wrong quote", func(c *data.Claim) { c.Citations[0].Quote = "return true" }, false},
		{"outside snapshot", func(c *data.Claim) { c.Citations[0].Path = "../secret" }, false},
		{"model proof claim", func(c *data.Claim) { c.Status = "proved" }, false},
		{"empty citation", func(c *data.Claim) { c.Citations = nil }, false},
		{"invented finding", func(c *data.Claim) { c.Status = "contradiction"; c.FindingIDs = []string{"invented"} }, false},
		{"real counterexample", func(c *data.Claim) { c.Status = "contradiction"; c.FindingIDs = []string{"F1"} }, true},
		{"wrong commit", func(c *data.Claim) { c.SourceKind = "commit"; c.SourceIndex = 5 }, false},
		{"commit", func(c *data.Claim) { c.SourceKind = "commit" }, true},
		{"unclear", func(c *data.Claim) { c.Status = "unclear"; c.Citations = nil }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := valid
			c.Citations = append([]data.Citation(nil), valid.Citations...)
			tc.mutate(&c)
			body, _ := json.Marshal(struct {
				Claims []data.Claim `json:"claims"`
			}{[]data.Claim{c}})
			calls := 0
			gen := modelFunc(func(_ context.Context, a, p string) (string, error) {
				calls++
				if a != "delivery" || !strings.Contains(p, "proved_counterexamples") {
					t.Fatal("not routed through delivery adapter")
				}
				return string(body), nil
			})
			r, err := Assess(context.Background(), gen, m, files, "snapshot", []data.Counterexample{{ID: "F1", Path: "price.py"}})
			if (err == nil) != tc.want || r.Complete != tc.want || calls != 1 {
				t.Fatalf("result=%+v err=%v calls=%d", r, err, calls)
			}
			if len(r.MetadataSHA256) != 64 {
				t.Fatal("missing context identity")
			}
		})
	}
	for _, body := range []string{`{}`, `{"result":"text envelope"}`} {
		r, err := Assess(context.Background(), modelFunc(func(context.Context, string, string) (string, error) { return body, nil }), m, files, "snapshot", nil)
		if err == nil || r.Complete {
			t.Fatal("accepted an empty vendor envelope")
		}
	}
}

func TestAssessmentMarksBoundedContextAndRedactsWithoutMutating(t *testing.T) {
	files := []change.File{{Path: "big.py", Change: []byte(strings.Repeat("a line\n", 3000))}}
	m := data.Metadata{Description: "Keep subscription-oauth-value private"}
	c := data.Claim{Requirement: m.Description, SourceKind: "description", SourceQuote: m.Description, Status: "unclear", Reason: "subscription-oauth-value is not a guarantee"}
	b, _ := json.Marshal(struct {
		Claims []data.Claim `json:"claims"`
	}{[]data.Claim{c}})
	r, err := Assess(context.Background(), modelFunc(func(context.Context, string, string) (string, error) { return string(b), nil }), m, files, "snapshot", nil)
	if err != nil || !r.ContextTruncated {
		t.Fatalf("%+v %v", r, err)
	}
	clean := r.Redacted(func(s string) string { return strings.ReplaceAll(s, "subscription-oauth-value", "[redacted]") })
	if strings.Contains(clean.Claims[0].Reason, "subscription-oauth-value") || !strings.Contains(r.Claims[0].Reason, "subscription-oauth-value") {
		t.Fatal("redaction leaked or mutated its input")
	}
}
