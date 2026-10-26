package gateway

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HeaderAssertion is the header Cloudflare Access adds to every request it lets
// through. It holds a JWT signed with the team's key.
const HeaderAssertion = "Cf-Access-Jwt-Assertion"

// The refusal codes of the Access assertion. They are what a caller sees.
const (
	CodeAssertionMissing   = "missing_assertion"
	CodeAssertionMalformed = "assertion_malformed"
	CodeAssertionKey       = "assertion_unknown_key"
	CodeAssertionSignature = "assertion_bad_signature"
	CodeAssertionExpired   = "assertion_expired"
	CodeAssertionAudience  = "assertion_wrong_audience"
	CodeAssertionIssuer    = "assertion_wrong_issuer"
	CodeAssertionNoSubject = "assertion_without_identity"
	CodeKeysUnavailable    = "access_keys_unavailable"
)

// The bounds of the key set: how long it is trusted, how soon an unknown key id
// may trigger another fetch, and how many bytes of it are read.
const (
	keysMaxAge     = 10 * time.Minute
	keysMinRefresh = time.Minute
	keysMaxBytes   = 1 << 20
	assertionMax   = 8 << 10
	clockLeeway    = 30 * time.Second
)

// Subject is who a valid assertion names: a judge by e-mail, or a service token
// by its client id. Exactly one of the two is set.
type Subject struct {
	Email    string
	ClientID string
}

// Name is the identity the settings file is matched against, lower case.
func (s Subject) Name() string {
	if s.Email != "" {
		return strings.ToLower(s.Email)
	}
	return s.ClientID
}

// Verifier checks an Access assertion with the standard library only. It
// trusts a token when its signature verifies against a key the team publishes,
// its issuer is the team domain, its audience is our application, and it has not
// expired.
type Verifier struct {
	Issuer   string // https://<team>.cloudflareaccess.com
	Audience string // the application's AUD tag
	KeysURL  string // Issuer + /cdn-cgi/access/certs
	Client   *http.Client
	Now      func() time.Time

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time // last fetch that succeeded
	triedAt   time.Time // last fetch of any kind
}

// NewVerifier builds a verifier for a team domain such as
// "team.cloudflareaccess.com". The scheme is added here.
func NewVerifier(teamDomain, audience string, client *http.Client, now func() time.Time) *Verifier {
	issuer := "https://" + strings.TrimSuffix(teamDomain, "/")
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if now == nil {
		now = time.Now
	}
	return &Verifier{Issuer: issuer, Audience: audience, KeysURL: issuer + "/cdn-cgi/access/certs", Client: client, Now: now}
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type jwtClaims struct {
	Iss        string          `json:"iss"`
	Aud        json.RawMessage `json:"aud"`
	Exp        float64         `json:"exp"`
	Nbf        float64         `json:"nbf"`
	Email      string          `json:"email"`
	CommonName string          `json:"common_name"`
}

// Verify returns the subject of a valid assertion, or a *Refusal naming what is
// wrong. The signature is checked before any claim is believed.
func (v *Verifier) Verify(ctx context.Context, assertion string) (Subject, error) {
	bad := func(code, detail string) (Subject, error) {
		return Subject{}, &Refusal{Status: http.StatusUnauthorized, Code: code, Detail: detail}
	}
	if assertion == "" {
		return bad(CodeAssertionMissing, "this gateway accepts only requests that passed Cloudflare Access; sign in through Access")
	}
	if len(assertion) > assertionMax {
		return bad(CodeAssertionMalformed, "the Access assertion is not a JWT")
	}
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		return bad(CodeAssertionMalformed, "the Access assertion is not a JWT")
	}
	var h jwtHeader
	if !decodePart(parts[0], &h) {
		return bad(CodeAssertionMalformed, "the Access assertion is not a JWT")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) == 0 {
		return bad(CodeAssertionMalformed, "the Access assertion is not a JWT")
	}
	if h.Alg != "RS256" {
		return bad(CodeAssertionKey, "the Access assertion is not signed with RS256")
	}
	key, kerr := v.key(ctx, h.Kid)
	if kerr != nil {
		return Subject{}, kerr
	}
	if key == nil {
		return bad(CodeAssertionKey, "the Access assertion is signed with a key the team does not publish")
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], sig) != nil {
		return bad(CodeAssertionSignature, "the signature of the Access assertion does not verify")
	}

	var c jwtClaims
	if !decodePart(parts[1], &c) {
		return bad(CodeAssertionMalformed, "the Access assertion is not a JWT")
	}
	now := v.Now()
	if c.Exp == 0 || !now.Before(time.Unix(int64(c.Exp), 0).Add(clockLeeway)) {
		return bad(CodeAssertionExpired, "the Access assertion has expired; sign in again")
	}
	if c.Nbf != 0 && now.Add(clockLeeway).Before(time.Unix(int64(c.Nbf), 0)) {
		return bad(CodeAssertionExpired, "the Access assertion is not valid yet")
	}
	if c.Iss != v.Issuer {
		return bad(CodeAssertionIssuer, "the Access assertion was issued by another team")
	}
	if !audienceHas(c.Aud, v.Audience) {
		return bad(CodeAssertionAudience, "the Access assertion is for another application")
	}
	switch {
	case c.Email != "":
		return Subject{Email: c.Email}, nil
	case c.CommonName != "":
		return Subject{ClientID: c.CommonName}, nil
	}
	return bad(CodeAssertionNoSubject, "the Access assertion names no e-mail and no service token")
}

func decodePart(part string, into any) bool {
	raw, err := base64.RawURLEncoding.DecodeString(part)
	return err == nil && json.Unmarshal(raw, into) == nil
}

// audienceHas reads the aud claim, which is a string or a list of strings.
func audienceHas(raw json.RawMessage, want string) bool {
	if want == "" {
		return false
	}
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == want
	}
	var many []string
	if json.Unmarshal(raw, &many) == nil {
		for _, a := range many {
			if a == want {
				return true
			}
		}
	}
	return false
}

// key returns the public key for a key id. A set older than keysMaxAge is
// fetched again; an id not in the set triggers one fetch, at most once a
// keysMinRefresh, so a caller who invents key ids cannot make the gateway call
// Cloudflare on every request. When a fetch fails and the old set still has the
// id, the old key is used; when it has not, the answer is a 503 naming the keys.
func (v *Verifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	now := v.Now()
	k := v.keys[kid]
	stale := v.keys == nil || now.Sub(v.fetchedAt) >= keysMaxAge
	unknown := k == nil && now.Sub(v.triedAt) >= keysMinRefresh
	if (stale || unknown) && now.Sub(v.triedAt) >= keysMinRefresh {
		v.triedAt = now
		if keys, err := v.fetch(ctx); err == nil {
			v.keys, v.fetchedAt = keys, now
			k = keys[kid]
		}
	}
	if v.keys == nil {
		return nil, &Refusal{Status: http.StatusServiceUnavailable, Code: CodeKeysUnavailable, Detail: "the gateway cannot read the Access signing keys; tell the owner"}
	}
	return k, nil
}

// fetch reads the team's key set: a JSON object with a "keys" list of RSA keys
// in JWK form.
func (v *Verifier) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.KeysURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.Client.Do(req)
	if err != nil {
		return nil, errors.New("keys unreachable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("keys not served")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, keysMaxBytes+1))
	if err != nil || len(raw) > keysMaxBytes {
		return nil, errors.New("keys unreadable")
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(raw, &set); err != nil {
		return nil, errors.New("keys unreadable")
	}
	out := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		n, nerr := base64.RawURLEncoding.DecodeString(k.N)
		e, eerr := base64.RawURLEncoding.DecodeString(k.E)
		if nerr != nil || eerr != nil || len(n) < 256 || len(e) == 0 || len(e) > 4 {
			continue // a key under 2048 bits is not accepted
		}
		exp := 0
		for _, b := range e {
			exp = exp<<8 | int(b)
		}
		if exp < 3 {
			continue
		}
		out[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exp}
	}
	if len(out) == 0 {
		return nil, errors.New("keys empty")
	}
	return out, nil
}
