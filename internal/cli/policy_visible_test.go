package cli

import (
	"strings"
	"testing"
)

// A control byte in a policy exception's reason reaches the text output
// visibly, never raw, for an accepted and for an expired exception, so a
// policy file cannot clear or rewrite the review the user reads.
func TestPolicyExceptionFromFileCannotClearReviewText(t *testing.T) {
	for _, expires := range []string{"2999-12-31", "2000-01-01"} {
		t.Run(expires, func(t *testing.T) {
			body := policyException("a", "discount-applied-twice", "billing/discount.go", expires)
			body = strings.Replace(body, "reviewed by the platform team", "reviewed\x1b[2Jhidden", 1)
			_, stdout, stderr := runWithPolicy(t, body)
			if stderr != "" {
				t.Fatalf("policy was refused: %s", stderr)
			}
			if strings.Contains(stdout, "\x1b") {
				t.Fatalf("policy text can clear the review")
			}
			if !strings.Contains(stdout, `\x1b`) {
				t.Fatalf("control byte was not shown visibly")
			}
		})
	}
}
