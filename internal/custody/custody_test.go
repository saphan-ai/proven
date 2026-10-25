package custody

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// The recorded sample is built here with fixed seeds, so every run signs the
// same bytes. No real key or real receipt is used: the sample has the shape
// Nightshift writes (an envelope of body, pubkey_b64 and sig_b64, signed over
// the body as sorted-key compact JSON) and nothing else of it.

var (
	signer = ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	other  = ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, ed25519.SeedSize))
)

const (
	answerModel = "nvidia/nemotron-test"
	day1        = "2026-10-24T10:00:00Z"
	day2        = "2026-10-25T10:00:00Z"
)

// receiptBody is a receipt body for response id, routed to model, at time at.
func receiptBody(id, model, at string) map[string]any {
	t, _ := time.Parse(time.RFC3339, at)
	return map[string]any{
		"signer": "meterlessd",
		"ts_ms":  t.UnixMilli(),
		"receipt": map[string]any{
			"request_id": id, "receipt_id": "rcpt_" + id, "status": "ok", "app": "proven",
			"created_at":    t.UnixMilli(),
			"winning_route": map[string]any{"provider": "nvidia", "model": model},
		},
	}
}

// line signs body with key and writes the envelope line, embedding embed's public key.
func line(t *testing.T, body map[string]any, key, embed ed25519.PrivateKey) string {
	t.Helper()
	raw, err := json.Marshal(body) // sorted keys, compact: the signed form
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]any{
		"body":       json.RawMessage(raw),
		"pubkey_b64": base64.StdEncoding.EncodeToString(embed.Public().(ed25519.PublicKey)),
		"sig_b64":    base64.StdEncoding.EncodeToString(ed25519.Sign(key, raw)),
	}
	out, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func good(t *testing.T, id, model, at string) string {
	return line(t, receiptBody(id, model, at), signer, signer)
}

func pinned() ed25519.PublicKey { return signer.Public().(ed25519.PublicKey) }

// traceLineFor is one trace "model" event.
func traceLineFor(at, requested, routed, receipt, response string) string {
	b, _ := json.Marshal(map[string]any{
		"time": at, "kind": "model", "agent": "planner",
		"model": map[string]any{
			"model": routed, "receipt_id": receipt, "requested": requested,
			"response_id": response, "from_receipt": true,
		},
	})
	return string(b)
}

func audit(t *testing.T, receipts string, traces ...string) Report {
	t.Helper()
	rep, err := Audit(strings.NewReader(receipts), pinned(),
		map[string]io.Reader{"trace.jsonl": strings.NewReader(strings.Join(traces, "\n"))})
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

// Property: a receipt signed by the pinned key over its body verifies, and the
// verifier returns the request id and the signed route.
func TestVerifyGoodReceipt(t *testing.T) {
	r, err := Verify([]byte(good(t, "chatcmpl-1", answerModel, day1)), pinned())
	if err != nil {
		t.Fatal(err)
	}
	if r.RequestID != "chatcmpl-1" || r.Model != answerModel || r.Provider != "nvidia" {
		t.Fatalf("receipt %#v", r)
	}
}

// Property: a body changed after signing does not verify. The mutated field is
// the routed model inside the signed body; the signature is left as it was.
func TestVerifyRefusesATamperedBody(t *testing.T) {
	l := good(t, "chatcmpl-1", answerModel, day1)
	tampered := strings.Replace(l, answerModel, "openai/gpt-tampered", 1)
	if tampered == l {
		t.Fatal("the tamper changed nothing")
	}
	if _, err := Verify([]byte(tampered), pinned()); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("err %v, want ErrBadSignature", err)
	}
}

// Property: a line signed under another key is refused as an unknown key, and
// so is a line whose embedded key is the pinned one but whose signature was made
// by another key (the embedded key is never trusted on its own).
func TestVerifyRefusesAWrongKey(t *testing.T) {
	own := line(t, receiptBody("chatcmpl-1", answerModel, day1), other, other)
	if _, err := Verify([]byte(own), pinned()); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("signed by another key: err %v, want ErrUnknownKey", err)
	}
	forged := line(t, receiptBody("chatcmpl-1", answerModel, day1), other, signer)
	if _, err := Verify([]byte(forged), pinned()); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("pinned key embedded, other key signed: err %v, want ErrBadSignature", err)
	}
}

// Property: a line that is not a receipt is refused by the name ErrMalformed,
// whatever part is wrong.
func TestVerifyRefusesAMalformedReceipt(t *testing.T) {
	ok := good(t, "chatcmpl-1", answerModel, day1)
	noID := line(t, map[string]any{"ts_ms": 1, "receipt": map[string]any{"status": "ok"}}, signer, signer)
	noReceipt := line(t, map[string]any{"ts_ms": 1}, signer, signer)
	cases := map[string]string{
		"not json":          "{nope",
		"empty":             "{}",
		"no signature":      strings.Replace(ok, `"sig_b64"`, `"sig_x"`, 1),
		"short signature":   `{"body":{},"pubkey_b64":"AAAA","sig_b64":"AAAA"}`,
		"signature not b64": strings.Replace(ok, `"sig_b64":"`, `"sig_b64":"!!`, 1),
		"no request id":     noID,
		"no receipt":        noReceipt,
	}
	for name, l := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Verify([]byte(l), pinned()); !errors.Is(err, ErrMalformed) {
				t.Fatalf("err %v, want ErrMalformed", err)
			}
		})
	}
}

// Property: an answer whose receipt verifies, whose response id is the
// receipt's request id, and whose routed model is the requested one is in
// custody, and the numbers are filed under the day each was written.
func TestAuditCountsAnAnswerInCustody(t *testing.T) {
	receipts := good(t, "chatcmpl-1", answerModel, day1) + "\n" + good(t, "chatcmpl-2", answerModel, day2)
	rep := audit(t, receipts,
		traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"),
		traceLineFor(day2, "meterless/auto", answerModel, "rcpt_chatcmpl-2", "chatcmpl-2"))
	if rep.Failed() != 0 {
		t.Fatalf("failures %#v", rep.Days)
	}
	if len(rep.Days) != 2 {
		t.Fatalf("days %d", len(rep.Days))
	}
	d1, d2 := rep.Days[0], rep.Days[1]
	if d1.Date != "2026-10-24" || d1.Receipts != 1 || d1.SignaturesVerified != 1 || d1.Answers != 1 ||
		d1.Matched != 1 || d1.ModelChecked != 1 || d1.ModelMatched != 1 || d1.InCustody != 1 {
		t.Fatalf("day one %#v", d1)
	}
	// An answer that asked for the router's choice has no model to compare;
	// it is counted as auto-routed, not as a match.
	if d2.AutoRouted != 1 || d2.ModelChecked != 0 || d2.InCustody != 1 {
		t.Fatalf("day two %#v", d2)
	}
}

// Property: an answer whose receipt was tampered with is not in custody, and
// the output names the refusal. Mutated field: the receipt body after signing.
func TestAuditNamesATamperedReceipt(t *testing.T) {
	l := good(t, "chatcmpl-1", answerModel, day1)
	tampered := strings.Replace(l, `"status":"ok"`, `"status":"okk"`, 1)
	rep := audit(t, tampered, traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"))
	if rep.Failed() != 1 || rep.Days[0].SignaturesVerified != 0 || rep.Days[0].Receipts != 1 {
		t.Fatalf("report %#v", rep.Days)
	}
	if got := rep.Days[0].Failures[0].Reason; !strings.Contains(got, ErrBadSignature.Error()) {
		t.Fatalf("reason %q", got)
	}
}

// Property: an answer signed under the wrong key is not in custody, and the
// reason is the unknown key.
func TestAuditNamesAReceiptUnderTheWrongKey(t *testing.T) {
	l := line(t, receiptBody("chatcmpl-1", answerModel, day1), other, other)
	rep := audit(t, l, traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"))
	if rep.Failed() != 1 || !strings.Contains(rep.Days[0].Failures[0].Reason, ErrUnknownKey.Error()) {
		t.Fatalf("report %#v", rep.Days)
	}
}

// Property: an answer whose response id matches no receipt is not in custody.
func TestAuditNamesAnAnswerWithNoReceipt(t *testing.T) {
	rep := audit(t, good(t, "chatcmpl-1", answerModel, day1),
		traceLineFor(day1, answerModel, answerModel, "rcpt_x", "chatcmpl-unmatched"))
	if rep.Failed() != 1 || !strings.Contains(rep.Days[0].Failures[0].Reason, "no receipt matches response id chatcmpl-unmatched") {
		t.Fatalf("report %#v", rep.Days)
	}
	if rep.Days[0].Matched != 0 {
		t.Fatalf("matched %d", rep.Days[0].Matched)
	}
}

// Property: an answer whose routed model is not the model requested is not in
// custody, even though its receipt verifies and matches.
func TestAuditNamesAModelMismatch(t *testing.T) {
	rep := audit(t, good(t, "chatcmpl-1", "openai/other", day1),
		traceLineFor(day1, answerModel, "openai/other", "rcpt_chatcmpl-1", "chatcmpl-1"))
	d := rep.Days[0]
	if rep.Failed() != 1 || d.Matched != 1 || d.ModelChecked != 1 || d.ModelMatched != 0 {
		t.Fatalf("report %#v", d)
	}
	if got := d.Failures[0].Reason; !strings.Contains(got, "model mismatch: requested "+answerModel+", answered by openai/other") {
		t.Fatalf("reason %q", got)
	}
}

// Property: when the trace's routed model differs from the model the signed
// receipt names, the answer is not in custody: the receipt is the signed statement.
func TestAuditNamesATraceThatDisagreesWithTheSignedReceipt(t *testing.T) {
	rep := audit(t, good(t, "chatcmpl-1", "openai/signed", day1),
		traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", "chatcmpl-1"))
	if rep.Failed() != 1 || !strings.Contains(rep.Days[0].Failures[0].Reason, "the signed receipt names openai/signed") {
		t.Fatalf("report %#v", rep.Days)
	}
}

// Property: an answer from an API that signs nothing, and an answer recorded
// before the response id was kept, are both refused by name and not counted.
func TestAuditRefusesAnAnswerWithoutAReceiptOrResponseID(t *testing.T) {
	rep := audit(t, good(t, "chatcmpl-1", answerModel, day1),
		traceLineFor(day1, answerModel, answerModel, "no receipt", "chatcmpl-1"),
		traceLineFor(day1, answerModel, answerModel, "rcpt_chatcmpl-1", ""))
	if rep.Failed() != 2 {
		t.Fatalf("report %#v", rep.Days)
	}
	if !strings.Contains(rep.Days[0].Failures[0].Reason, "no receipt") ||
		!strings.Contains(rep.Days[0].Failures[1].Reason, "no response id recorded") {
		t.Fatalf("reasons %#v", rep.Days[0].Failures)
	}
}

// Property: a sample run's model events name no request and no receipt; they
// are not live answers, so they are skipped and counted, never failed.
func TestAuditSkipsSampleCalls(t *testing.T) {
	sample := `{"time":"` + day1 + `","kind":"model","agent":"planner","model":{"model":"unknown","from_receipt":false}}`
	rep := audit(t, good(t, "chatcmpl-1", answerModel, day1), sample)
	if rep.Skipped != 1 || rep.Failed() != 0 {
		t.Fatalf("skipped %d failed %d", rep.Skipped, rep.Failed())
	}
}

// Property: the printed report names every failing answer and says custody
// FAILED; with no failure it says custody holds.
func TestWriteNamesEveryFailure(t *testing.T) {
	rep := audit(t, good(t, "chatcmpl-1", answerModel, day1),
		traceLineFor(day1, answerModel, answerModel, "rcpt_x", "chatcmpl-nope"))
	var out bytes.Buffer
	Write(&out, rep)
	for _, want := range []string{"day 2026-10-24", "receipts: 1", "signatures verified: 1", "FAIL trace.jsonl:1", "custody FAILED for 1 answers"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q:\n%s", want, out.String())
		}
	}
}
