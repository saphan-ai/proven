package scout

// This file is the execution contract (the spec's docs/execution-contract.md): the one artifact a
// downstream engine accepts. It is signed with HMAC-SHA256 over its canonical JSON, expires after a
// time to live, and is verified before anything acts on it.

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Contract is a signed execution contract.
type Contract struct {
	ContractID        string         `json:"contractId"`
	ParentID          string         `json:"parentId,omitempty"`
	TraceID           string         `json:"traceId"`
	Intent            ContractIntent `json:"intent"`
	Risk              Risk           `json:"risk"`
	ToolPlan          []ToolStep     `json:"toolPlan"`
	ModelProfile      string         `json:"modelProfile"`
	Scope             Scope          `json:"scope"`
	IssuedAt          string         `json:"issuedAt"`
	ExpiresAt         string         `json:"expiresAt"`
	RegistryVersion   int            `json:"registryVersion"`
	PolicyPackVersion string         `json:"policyPackVersion"`
	Signature         string         `json:"signature"`
	SignedBy          string         `json:"signedBy"`
}

// ContractIntent is the intent part of a contract: the primary intent with its bound parameters and
// references, the confidence it was chosen with, and the alternates a surface may offer.
type ContractIntent struct {
	Primary    IntentRef      `json:"primary"`
	Secondary  []IntentRef    `json:"secondary,omitempty"`
	Parameters map[string]any `json:"parameters"`
	References map[string]any `json:"references"`
	Confidence float64        `json:"confidence"`
	Alternates []Candidate    `json:"alternates"`
}

// IntentRef names one intent by id.
type IntentRef struct {
	ID string `json:"id"`
}

// ToolStep is one step of the tool plan. Capability is the intent's abstract verb; Tool is what it
// resolves to. With no capability graph configured, each capability resolves to the tool of its own name.
type ToolStep struct {
	Tool       string         `json:"tool"`
	Capability string         `json:"capability"`
	Input      map[string]any `json:"input"`
}

// Scope is what the contract lets a downstream engine do: who asked, where, and the capabilities it
// may use. An engine refuses a tool outside Capabilities.
type Scope struct {
	User         string   `json:"user"`
	Surface      string   `json:"surface"`
	Capabilities []string `json:"capabilities"`
}

// Signer signs and checks contracts. The spec makes the signer pluggable; this one is HMAC-SHA256.
type Signer struct {
	key []byte
}

// NewSigner returns a signer with key. An empty key gets a random one, which signs and checks within
// one process: the spec makes a signature mandatory at every boundary, including development.
func NewSigner(key []byte) (*Signer, error) {
	if len(key) == 0 {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
	}
	return &Signer{key: append([]byte(nil), key...)}, nil
}

// Sign fills the signature of c over its canonical encoding without the signature.
func (s *Signer) Sign(c *Contract) error {
	c.Signature = ""
	c.SignedBy = "scout"
	b, err := canonical(c)
	if err != nil {
		return err
	}
	c.Signature = s.mac(b)
	return nil
}

// ErrUnsigned, ErrTampered and ErrExpired are the three refusals of Verify.
var (
	ErrUnsigned = errors.New("scout: contract is not signed")
	ErrTampered = errors.New("scout: contract signature does not match")
	ErrExpired  = errors.New("scout: contract has expired")
)

// Verify checks that c carries this signer's signature over its content and has not expired at now.
// Spec: a downstream engine refuses an unsigned, tampered or expired contract.
func (s *Signer) Verify(c Contract, now time.Time) error {
	if c.Signature == "" {
		return ErrUnsigned
	}
	got := c.Signature
	c.Signature = ""
	b, err := canonical(&c)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(got), []byte(s.mac(b))) {
		return ErrTampered
	}
	exp, err := time.Parse(time.RFC3339Nano, c.ExpiresAt)
	if err != nil {
		return fmt.Errorf("scout: contract expiry %q: %w", c.ExpiresAt, err)
	}
	if !now.Before(exp) {
		return ErrExpired
	}
	return nil
}

// mac is the hex HMAC-SHA256 of b.
func (s *Signer) mac(b []byte) string {
	h := hmac.New(sha256.New, s.key)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// canonical is the contract's canonical JSON: encoding/json writes struct fields in declaration order
// and map keys sorted, so the same contract always encodes to the same bytes.
func canonical(c *Contract) ([]byte, error) {
	return json.Marshal(c)
}

// crockford is the ULID alphabet.
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// newULID returns a ULID for t: 48 bits of milliseconds and 80 random bits, in Crockford base 32.
func newULID(t time.Time) (string, error) {
	var b [16]byte
	ms := uint64(t.UnixMilli())
	for i := 5; i >= 0; i-- {
		b[i] = byte(ms)
		ms >>= 8
	}
	if _, err := rand.Read(b[6:]); err != nil {
		return "", err
	}
	// 128 bits are 26 characters of 5 bits; the first character carries the top 3 bits.
	var sb strings.Builder
	var acc uint32
	bits := 2
	for _, x := range b {
		acc = acc<<8 | uint32(x)
		bits += 8
		for bits >= 5 {
			bits -= 5
			sb.WriteByte(crockford[(acc>>uint(bits))&31])
		}
	}
	return sb.String(), nil
}
