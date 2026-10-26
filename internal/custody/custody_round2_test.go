package custody

import (
	"strings"
	"testing"
)

// Property: a receipt that verifies but whose signed status is not "ok" never
// counts an answer in custody. The mutated field is `status`, inside the signed
// body, signed again so the signature still verifies.
func TestAuditFailsAnAnswerWhoseReceiptStatusIsFailed(t *testing.T) {
	body := receiptBody("chatcmpl-1", answerModel, day1)
	body["receipt"].(map[string]any)["status"] = "failed"
	rep := audit(t, line(t, body, signer, signer),
		traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"))
	d := rep.Days[0]
	if rep.Failed() != 1 || d.InCustody != 0 || d.Matched != 0 || d.SignaturesVerified != 1 {
		t.Fatalf("report %#v", d)
	}
	if got := d.Failures[0].Reason; !strings.Contains(got, `receipt status is "failed", not "ok"`) {
		t.Fatalf("reason %q", got)
	}
}

// Property: one receipt vouches for one answer. Two answers that name the same
// response id are not both in custody; the second fails and names the source
// of the first. The mutated input is the second trace line's repeated id.
func TestAuditCountsOneReceiptForOneAnswer(t *testing.T) {
	rep := audit(t, good(t, "chatcmpl-1", answerModel, day1),
		traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"),
		traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"))
	d := rep.Days[0]
	if rep.Failed() != 1 || d.Matched != 1 || d.InCustody != 1 {
		t.Fatalf("report %#v", d)
	}
	if got := d.Failures[0].Reason; got != "response id already matched at trace.jsonl:1" {
		t.Fatalf("reason %q", got)
	}
}

// Property: an answer that asked for meterless/auto still needs a receipt that
// verifies. With no receipt, or with a receipt that is refused, it fails by
// name and is not counted as auto-routed.
func TestAuditFailsAnAutoAnswerWithoutAValidReceipt(t *testing.T) {
	tampered := strings.Replace(good(t, "chatcmpl-2", answerModel, day1), `"status":"ok"`, `"status":"okk"`, 1)
	cases := map[string]struct {
		receipts string
		response string
		want     string
	}{
		"no receipt":      {good(t, "chatcmpl-1", answerModel, day1), "chatcmpl-unmatched", "no receipt matches response id"},
		"refused receipt": {tampered, "chatcmpl-2", "receipt refused: " + ErrBadSignature.Error()},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			rep := audit(t, c.receipts, traceLineFor(day1, AutoModel, answerModel, "rcpt_x", c.response))
			d := rep.Days[0]
			if rep.Failed() != 1 || d.AutoRouted != 0 || d.InCustody != 0 {
				t.Fatalf("report %#v", d)
			}
			if got := d.Failures[0].Reason; !strings.Contains(got, c.want) {
				t.Fatalf("reason %q, want %q", got, c.want)
			}
		})
	}
}
