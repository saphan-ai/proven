// Package guardrail checks a prompt, and an answer, before a review uses them.
// An unsafe verdict stops that step. A guardrail that does not answer refuses
// the step by name. It never treats a missing verdict as safe.
//
// ADR-0005: the guardrail is on by default. Every prompt an agent sends to a
// model, and every answer it gets back, passes NVIDIA's Nemotron 3.5 Content
// Safety first. Wrap puts the check around the generator the model-calling
// agents use. One adapter has two backends, our own model server and NVIDIA's
// hosted API; a setting turns the guardrail off, and the trace then says so.
// An empty key is not that setting: Absent refuses every check and does not
// dial (ADR-0005 point 5).
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
// Backend, Model and Down are filled by Wrap, not by an adapter, so every
// check it reports names who answered and, when it could not be made, why.
type Verdict struct {
	User       string
	Response   string
	Categories []string
	Latency    time.Duration
	// Backend is the adapter's name from Identity, or "unknown".
	Backend string
	// Model is the configured model id from Identity; empty when none is configured.
	Model string
	// Down is one value of the closed set in down.go when the check could not
	// be made, and empty when it returned a verdict.
	Down string
}

// Unsafe reports that the prompt or the answer was judged unsafe.
func (v Verdict) Unsafe() bool {
	return v.User == "unsafe" || v.Response == "unsafe"
}

// Adapter is one guardrail backend. answer is empty when only the prompt is judged.
type Adapter interface {
	Check(ctx context.Context, prompt, answer string) (Verdict, error)
}

// Identity is implemented by an adapter that can say what it is: its backend
// name and its configured model id. It is a separate interface, not a method of
// Adapter, so an adapter written elsewhere still satisfies Adapter; Wrap names
// such an adapter "unknown".
type Identity interface {
	Identity() (backend, model string)
}

// BackendUnknown names an adapter that does not implement Identity.
const BackendUnknown = "unknown"

// identify returns the backend and model rail names, or BackendUnknown.
func identify(rail Adapter) (backend, model string) {
	if id, ok := rail.(Identity); ok {
		return id.Identity()
	}
	return BackendUnknown, ""
}

// Refuse is a named guardrail failure. The error text is "guardrail: " plus reason.
func Refuse(reason string) *service.Down {
	return &service.Down{Service: Service, Reason: reason}
}

// unsafeReason is the refusal text for an unsafe verdict, with its categories
// when the guardrail named any.
func unsafeReason(v Verdict) string {
	if len(v.Categories) == 0 {
		return "unsafe"
	}
	return "unsafe: " + strings.Join(v.Categories, ", ")
}
