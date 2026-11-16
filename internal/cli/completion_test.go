package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

type incompleteReviewer struct{ result review.Result }

func (r incompleteReviewer) Review(context.Context, review.Request) (review.Result, error) {
	return r.result, nil
}

func TestUnresolvedPlanningHypothesisIsReportedAndFailsRequiredCompletion(t *testing.T) {
	h := review.UnresolvedHypothesis{Reviewer: "default", AfterStep: 33, Description: `a missing caller needs inspection <script>alert(1)</script>`}
	result := review.Result{UnresolvedHypotheses: []review.UnresolvedHypothesis{h}}
	if result.CompletionStatus() != "incomplete" || result.Partial || len(result.Findings) != 0 {
		t.Fatal("hypothesis was confused with a finding or deadline")
	}
	for _, format := range []string{"text", "md", "html", "json", "sarif"} {
		b, err := render(format, result)
		if err != nil {
			t.Fatal(err)
		}
		if format == "json" || format == "sarif" {
			var doc map[string]any
			if err := json.Unmarshal(b, &doc); err != nil {
				t.Fatal(err)
			}
			if format == "sarif" {
				doc = doc["runs"].([]any)[0].(map[string]any)["properties"].(map[string]any)
			}
			if doc["completion_status"] != "incomplete" || len(doc["unresolved_hypotheses"].([]any)) != 1 {
				t.Fatal(string(b))
			}
		} else if !strings.Contains(string(b), "missing caller needs inspection") || strings.Contains(string(b), "Proven: clean") {
			t.Fatal(string(b))
		}
		if format == "html" && strings.Contains(string(b), "<script>alert(1)</script>") {
			t.Fatal("unescaped hypothesis")
		}
	}
	for _, required := range []string{"true", "false"} {
		var out, log bytes.Buffer
		code := Run([]string{"scan", "--json", t.TempDir()}, func(k string) string {
			if k == "PROVEN_REQUIRE_COMPLETE" {
				return required
			}
			return ""
		}, &out, &log, func(review.Options) (review.Reviewer, error) { return incompleteReviewer{result}, nil })
		want := 0
		if required == "true" {
			want = 3
		}
		if code != want || !strings.Contains(out.String(), "unresolved_hypotheses") {
			t.Fatalf("code=%d output=%s log=%s", code, &out, &log)
		}
	}
}

func TestCompletionPolicyIsIndependentOfFindings(t *testing.T) {
	for _, partial := range []bool{false, true} {
		result := review.Result{Partial: partial}
		if !partial {
			result.FailedServices = []review.ServiceFailure{{Service: "generator", Reason: "timeout"}}
		}
		for _, required := range []string{"true", "false"} {
			dir := t.TempDir()
			logDir := t.TempDir()
			var out, errout bytes.Buffer
			code := Run([]string{"scan", "--json", dir}, func(k string) string {
				switch k {
				case "PROVEN_REQUIRE_COMPLETE":
					return required
				case "PROVEN_LOG_DIR":
					return logDir
				}
				return ""
			}, &out, &errout, func(review.Options) (review.Reviewer, error) { return incompleteReviewer{result}, nil })
			want := 0
			if required == "true" {
				want = 3
			}
			if code != want {
				t.Fatalf("partial=%v required=%s: %d %s", partial, required, code, errout.String())
			}
			if !bytes.Contains(out.Bytes(), []byte(`"completion_status": "incomplete"`)) {
				t.Fatal(out.String())
			}
		}
	}
}
