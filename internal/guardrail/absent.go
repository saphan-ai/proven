package guardrail

import "context"

// Absent is a guardrail that refuses every check and does not dial.
// ADR-0005 point 5: an empty key is not mode off. Wrap writes one trace line
// per refused call, so each model call that was attempted is in the trace.
type Absent struct {
	Reason string
}

// Check refuses. The reason names why the guardrail is down.
func (a Absent) Check(context.Context, string, string) (Verdict, error) {
	reason := a.Reason
	if reason == "" {
		reason = "not configured"
	}
	return Verdict{}, Refuse(reason)
}

// Identity names this adapter "absent": no backend was configured, so no model is.
func (Absent) Identity() (backend, model string) { return "absent", "" }
