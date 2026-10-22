// Package guardrail checks a prompt, and an answer, before a review uses them.
// An unsafe verdict stops that step. A guardrail that does not answer refuses
// the step by name. It never treats a missing verdict as safe.
package guardrail

import (
	"context"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// Service is the name a refusal and a trace line use.
const Service = "guardrail"

// Verdict is one check. Response is empty when the check had no answer.
// Latency is the time the check took.
type Verdict struct {
	User       string
	Response   string
	Categories []string
	Latency    time.Duration
}

// Unsafe reports that the prompt or the answer was judged unsafe.
func (v Verdict) Unsafe() bool {
	return v.User == "unsafe" || v.Response == "unsafe"
}

// Adapter is one guardrail backend. answer is empty when only the prompt is judged.
type Adapter interface {
	Check(ctx context.Context, prompt, answer string) (Verdict, error)
}

// Refuse is a named guardrail failure. The error text is "guardrail: " plus reason.
func Refuse(reason string) *service.Down {
	return &service.Down{Service: Service, Reason: reason}
}

func unsafeReason(v Verdict) string {
	if len(v.Categories) == 0 {
		return "unsafe"
	}
	return "unsafe: " + strings.Join(v.Categories, ", ")
}
