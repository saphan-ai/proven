// Package custody checks that a model answer is in Proven's custody: the
// answer counts only when its Nightshift receipt verifies, the receipt names
// the answer's response id, and the model that answered is the model that was
// asked for.
//
// This file verifies one receipt line. audit.go joins the receipts to the
// answers a review recorded in its trace. Nothing here dials a network: the
// receipts file and the trace are the only inputs.
package custody

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// The named refusals of the verifier. Each wraps with the detail of the line
// it refused, so errors.Is names the class and the message names the cause.
var (
	// ErrMalformed is a line that is not a receipt: bad JSON, a missing part,
	// a field that does not decode, or a key or signature of the wrong length.
	ErrMalformed = errors.New("malformed receipt")
	// ErrUnknownKey is a line signed under a key other than the pinned one.
	// The key embedded in a line is never trusted: it is compared with the key
	// the operator pinned.
	ErrUnknownKey = errors.New("unknown key")
	// ErrBadSignature is a line whose Ed25519 signature does not verify over
	// its body under the pinned key.
	ErrBadSignature = errors.New("signature does not verify")
)

// Receipt is the part of a verified receipt the audit reads.
type Receipt struct {
	// RequestID equals the id of the OpenAI-compatible response.
	RequestID string
	// ReceiptID is the receipt's own id.
	ReceiptID string
	// Status is the receipt's status, "ok" or "failed".
	Status string
	// App is the meterless-app header value the request carried.
	App string
	// Provider and Model are the receipt's winning route, "" when it names none.
	Provider string
	Model    string
	// TimeMS is when the receipt was written, in milliseconds since the epoch.
	TimeMS int64
}

// envelope is one line of the receipts file. Body is kept as the bytes that
// were stored, because the signature is over the body and not over the line.
type envelope struct {
	Body      json.RawMessage `json:"body"`
	PubkeyB64 string          `json:"pubkey_b64"`
	SigB64    string          `json:"sig_b64"`
}

// body is the decoded part of the signed body the audit reads.
type body struct {
	TSMS    int64 `json:"ts_ms"`
	Receipt *struct {
		RequestID    string `json:"request_id"`
		ReceiptID    string `json:"receipt_id"`
		CreatedAt    int64  `json:"created_at"`
		Status       string `json:"status"`
		App          string `json:"app"`
		WinningRoute *struct {
			Provider string `json:"provider"`
			Model    string `json:"model"`
		} `json:"winning_route"`
	} `json:"receipt"`
}

// PublicKey decodes a base64 Ed25519 public key. It refuses a key that is not
// base64 or not 32 bytes.
func PublicKey(b64 string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("%w: public key is not base64", ErrMalformed)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: public key is %d bytes, want %d", ErrMalformed, len(raw), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(raw), nil
}

// Verify checks one line of a receipts file against the pinned key and returns
// the receipt it carries. The order of the checks is the order of the named
// refusals: a line must be well formed, then signed under the pinned key, then
// the signature must verify, and only then is its content read.
func Verify(line []byte, pinned ed25519.PublicKey) (Receipt, error) {
	var env envelope
	if err := json.Unmarshal(line, &env); err != nil {
		return Receipt{}, fmt.Errorf("%w: not JSON", ErrMalformed)
	}
	if len(env.Body) == 0 || env.PubkeyB64 == "" || env.SigB64 == "" {
		return Receipt{}, fmt.Errorf("%w: body, pubkey_b64 or sig_b64 is missing", ErrMalformed)
	}
	sig, err := base64.StdEncoding.DecodeString(env.SigB64)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return Receipt{}, fmt.Errorf("%w: signature is not %d bytes of base64", ErrMalformed, ed25519.SignatureSize)
	}
	embedded, err := PublicKey(env.PubkeyB64)
	if err != nil {
		return Receipt{}, err
	}
	if len(pinned) != ed25519.PublicKeySize {
		return Receipt{}, fmt.Errorf("%w: no pinned key", ErrUnknownKey)
	}
	if !embedded.Equal(pinned) {
		return Receipt{}, fmt.Errorf("%w: the line is signed under a key other than the pinned one", ErrUnknownKey)
	}
	if !signatureHolds(env.Body, sig, pinned) {
		return Receipt{}, ErrBadSignature
	}
	var b body
	if err := json.Unmarshal(env.Body, &b); err != nil || b.Receipt == nil {
		return Receipt{}, fmt.Errorf("%w: body has no receipt", ErrMalformed)
	}
	if b.Receipt.RequestID == "" {
		return Receipt{}, fmt.Errorf("%w: receipt has no request_id", ErrMalformed)
	}
	r := Receipt{
		RequestID: b.Receipt.RequestID, ReceiptID: b.Receipt.ReceiptID,
		Status: b.Receipt.Status, App: b.Receipt.App, TimeMS: b.TSMS,
	}
	if r.TimeMS == 0 {
		r.TimeMS = b.Receipt.CreatedAt
	}
	if w := b.Receipt.WinningRoute; w != nil {
		r.Provider, r.Model = w.Provider, w.Model
	}
	return r, nil
}

// signatureHolds reports whether sig is a valid signature by key over the
// body. The signed bytes are the body as sorted-key compact JSON. The stored
// bytes already are that form, and the sorted re-encoding is tried first so a
// body stored with spaces still verifies; the stored bytes, compacted, are the
// second candidate. Both are bytes the key holder could have signed; neither
// weakens the check, because a changed body verifies under neither.
func signatureHolds(raw json.RawMessage, sig []byte, key ed25519.PublicKey) bool {
	if canon, err := canonical(raw); err == nil && ed25519.Verify(key, canon, sig) {
		return true
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return false
	}
	return ed25519.Verify(key, compact.Bytes(), sig)
}

// canonical re-encodes a JSON value with its object keys sorted and no
// spaces. Numbers are kept as written, so a large id does not lose digits.
func canonical(raw []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(out.Bytes(), "\n"), nil
}
