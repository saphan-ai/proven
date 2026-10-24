package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// agreedOpinion is the shape the recorded scenario puts on the result: both
// reasons word for word, level note, and a place that is the shared lines.
func agreedOpinion() review.Opinion {
	return review.Opinion{
		ID: "O1", Title: "bounds check", Level: review.OpinionLevel,
		Location: review.Location{Path: "billing/discount.go", StartLine: 6, EndLine: 8},
		Reasons: []review.ReviewerVerdict{
			{Role: review.RoleDefault, Model: "NVIDIA", Verdict: review.VerdictReported, Reason: "default: the bounds check is missing"},
			{Role: review.RoleSpecial, Model: "qwen-special", Verdict: review.VerdictReported, Reason: "special: the bounds check is missing"},
		},
	}
}

// TestTheOpinionSectionCarriesBothReasons proves the text section names itself
// "Agreed by two models, not proven" and prints both reviewers' reasons as they
// wrote them. The opinion is not a "finding " line.
func TestTheOpinionSectionCarriesBothReasons(t *testing.T) {
	body, err := render("text", review.Result{
		Class: "security", ClassRule: "go.mod",
		Reviewers: []review.ReviewerInfo{{Role: review.RoleDefault}, {Role: review.RoleSpecial}},
		Opinions:  []review.Opinion{agreedOpinion()},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		"Agreed by two models, not proven\n",
		"  opinion O1 billing/discount.go:6-8\n",
		"  default NVIDIA: default: the bounds check is missing\n",
		"  special qwen-special: special: the bounds check is missing\n",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "\nfinding ") {
		t.Fatalf("the opinion was printed as a finding\n%s", text)
	}
}

// TestOneReviewerSaysWhyTheSectionIsEmpty proves ADR-0041's consequence: with
// one reviewer the section is empty and says why.
func TestOneReviewerSaysWhyTheSectionIsEmpty(t *testing.T) {
	body, err := render("text", review.Result{
		Class: "backend", ClassRule: "fallback",
		Reviewers: []review.ReviewerInfo{{Role: review.RoleDefault, Model: "NVIDIA"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "Agreed by two models, not proven\n  none: one reviewer; a claim only one model makes stays dropped\n") {
		t.Fatalf("%s", body)
	}
}

// TestAnOpinionDoesNotChangeTheExitCode proves an opinion is not a finding for
// the exit code. With --error, an opinion alone exits 0. One finding exits 1,
// and the same finding plus the opinion still exits 1.
func TestAnOpinionDoesNotChangeTheExitCode(t *testing.T) {
	opinionOnly := &captureReviewer{res: review.Result{
		Reviewers: []review.ReviewerInfo{{Role: review.RoleDefault}, {Role: review.RoleSpecial}},
		Opinions:  []review.Opinion{agreedOpinion()},
	}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--error"}, emptyEnv, &stdout, &stderr, openCapture(opinionOnly))
	if code != exitOK {
		t.Fatalf("opinion alone: exit %d\n%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "default: the bounds check is missing") || !strings.Contains(stdout.String(), "special: the bounds check is missing") {
		t.Fatalf("reasons missing\n%s", stdout.String())
	}

	finding := review.Finding{
		ID: "F1", RuleID: "sample.finding", Severity: review.SeverityError,
		Title: "a proved one", Message: "proved",
		Location: review.Location{Path: "a.go", StartLine: 1, EndLine: 1},
	}
	findingOnly := &captureReviewer{res: review.Result{Findings: []review.Finding{finding}}}
	stdout.Reset()
	stderr.Reset()
	without := Run([]string{"scan", "--error"}, emptyEnv, &stdout, &stderr, openCapture(findingOnly))
	if without != exitFindings {
		t.Fatalf("finding alone: exit %d", without)
	}

	withOpinion := &captureReviewer{res: review.Result{
		Findings: []review.Finding{finding},
		Opinions: []review.Opinion{agreedOpinion()},
	}}
	stdout.Reset()
	stderr.Reset()
	with := Run([]string{"scan", "--error"}, emptyEnv, &stdout, &stderr, openCapture(withOpinion))
	if with != without {
		t.Fatalf("exit with the opinion %d, without it %d", with, without)
	}
}
