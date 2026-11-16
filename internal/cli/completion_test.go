package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

type incompleteReviewer struct{ result review.Result }

func (r incompleteReviewer) Review(context.Context, review.Request) (review.Result, error) {
	return r.result, nil
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
