package gateway

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testTeam = "jury.cloudflareaccess.com"
	testAud  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testKid  = "kid-one"
)

var (
	keyOnce sync.Once
	keyA    *rsa.PrivateKey // the team's key, published
	keyB    *rsa.PrivateKey // a key the team never published
)

func testKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PrivateKey) {
	t.Helper()
	keyOnce.Do(func() {
		var err error
		if keyA, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
			panic(err)
		}
		if keyB, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
			panic(err)
		}
	})
	return keyA, keyB
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// jwk is the public half of a key as the team's certs endpoint serves it.
func jwk(kid string, k *rsa.PrivateKey) map[string]string {
	e := []byte{0, 0, 0, 0}
	e[0], e[1], e[2], e[3] = byte(k.E>>24), byte(k.E>>16), byte(k.E>>8), byte(k.E)
	for len(e) > 1 && e[0] == 0 {
		e = e[1:]
	}
	return map[string]string{"kid": kid, "kty": "RSA", "alg": "RS256", "use": "sig", "n": b64(k.N.Bytes()), "e": b64(e)}
}

// sign makes a JWT with the given header and claims, signed with RS256.
func sign(t *testing.T, key *rsa.PrivateKey, header, claims map[string]any) string {
	t.Helper()
	h, _ := json.Marshal(header)
	c, _ := json.Marshal(claims)
	signing := b64(h) + "." + b64(c)
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	return signing + "." + b64(sig)
}

// claimsFor are the claims of a good judge assertion at the time t0.
func claimsFor(email string) map[string]any {
	return map[string]any{
		"iss": "https://" + testTeam, "aud": []string{testAud}, "email": email,
		"iat": t0.Unix(), "exp": t0.Add(time.Hour).Unix(),
	}
}

func goodHeader() map[string]any { return map[string]any{"alg": "RS256", "kid": testKid, "typ": "JWT"} }

// jwks is a fake certs endpoint that counts its calls.
type jwks struct {
	mu     sync.Mutex
	hits   int
	status int
	keys   []map[string]string
	srv    *httptest.Server
}

func newJWKS(t *testing.T, keys ...map[string]string) *jwks {
	t.Helper()
	j := &jwks{status: 200, keys: keys}
	j.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		j.mu.Lock()
		j.hits++
		status, ks := j.status, j.keys
		j.mu.Unlock()
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": ks})
	}))
	t.Cleanup(j.srv.Close)
	return j
}

func (j *jwks) count() int { j.mu.Lock(); defer j.mu.Unlock(); return j.hits }
func (j *jwks) set(status int, keys ...map[string]string) {
	j.mu.Lock()
	j.status, j.keys = status, keys
	j.mu.Unlock()
}

// clock is a clock a test moves.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func verifierFor(j *jwks, c *clock) *Verifier {
	v := NewVerifier(testTeam, testAud, j.srv.Client(), c.now)
	v.KeysURL = j.srv.URL
	return v
}

func refusal(t *testing.T, err error) *Refusal {
	t.Helper()
	var ref *Refusal
	if !errors.As(err, &ref) {
		t.Fatalf("not a refusal: %v", err)
	}
	return ref
}

func TestAssertionIsVerifiedOrRefusedByName(t *testing.T) {
	a, b := testKeys(t)
	j := newJWKS(t, jwk(testKid, a))
	v := verifierFor(j, &clock{t: t0})

	with := func(mut func(c map[string]any)) map[string]any {
		c := claimsFor("ada@jury.example")
		mut(c)
		return c
	}
	cases := map[string]struct {
		token string
		code  string // empty means accepted
		want  Subject
	}{
		"a judge by e-mail":     {sign(t, a, goodHeader(), claimsFor("ada@jury.example")), "", Subject{Email: "ada@jury.example"}},
		"a service token":       {sign(t, a, goodHeader(), with(func(c map[string]any) { delete(c, "email"); c["common_name"] = "abc.access" })), "", Subject{ClientID: "abc.access"}},
		"aud as a plain string": {sign(t, a, goodHeader(), with(func(c map[string]any) { c["aud"] = testAud })), "", Subject{Email: "ada@jury.example"}},
		"aud among several":     {sign(t, a, goodHeader(), with(func(c map[string]any) { c["aud"] = []string{"other", testAud} })), "", Subject{Email: "ada@jury.example"}},

		"no header":            {"", CodeAssertionMissing, Subject{}},
		"not a jwt":            {"abc", CodeAssertionMalformed, Subject{}},
		"four parts":           {"a.b.c.d", CodeAssertionMalformed, Subject{}},
		"unknown key id":       {sign(t, b, map[string]any{"alg": "RS256", "kid": "kid-evil"}, claimsFor("ada@jury.example")), CodeAssertionKey, Subject{}},
		"known id, wrong key":  {sign(t, b, goodHeader(), claimsFor("ada@jury.example")), CodeAssertionSignature, Subject{}},
		"alg none":             {b64([]byte(`{"alg":"none","kid":"kid-one"}`)) + "." + b64([]byte(`{"email":"ada@jury.example"}`)) + ".AAAA", CodeAssertionKey, Subject{}},
		"alg HS256":            {sign(t, a, map[string]any{"alg": "HS256", "kid": testKid}, claimsFor("ada@jury.example")), CodeAssertionKey, Subject{}},
		"expired":              {sign(t, a, goodHeader(), with(func(c map[string]any) { c["exp"] = t0.Add(-time.Minute).Unix() })), CodeAssertionExpired, Subject{}},
		"no expiry":            {sign(t, a, goodHeader(), with(func(c map[string]any) { delete(c, "exp") })), CodeAssertionExpired, Subject{}},
		"not yet valid":        {sign(t, a, goodHeader(), with(func(c map[string]any) { c["nbf"] = t0.Add(time.Hour).Unix() })), CodeAssertionExpired, Subject{}},
		"wrong audience":       {sign(t, a, goodHeader(), with(func(c map[string]any) { c["aud"] = []string{strings.Repeat("b", 64)} })), CodeAssertionAudience, Subject{}},
		"no audience":          {sign(t, a, goodHeader(), with(func(c map[string]any) { delete(c, "aud") })), CodeAssertionAudience, Subject{}},
		"wrong issuer":         {sign(t, a, goodHeader(), with(func(c map[string]any) { c["iss"] = "https://other.cloudflareaccess.com" })), CodeAssertionIssuer, Subject{}},
		"no e-mail and no cn":  {sign(t, a, goodHeader(), with(func(c map[string]any) { delete(c, "email") })), CodeAssertionNoSubject, Subject{}},
		"claims tampered":      tamper(t, a),
		"signature cut to nil": {strings.Join(strings.Split(sign(t, a, goodHeader(), claimsFor("ada@jury.example")), ".")[:2], ".") + ".", CodeAssertionMalformed, Subject{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := v.Verify(context.Background(), tc.token)
			if tc.code == "" {
				if err != nil || got != tc.want {
					t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("accepted %+v; want %s", got, tc.code)
			}
			if ref := refusal(t, err); ref.Code != tc.code || ref.Status != 401 {
				t.Fatalf("refused as %s (%d); want %s", ref.Code, ref.Status, tc.code)
			}
		})
	}
}

// tamper changes the claims of a signed assertion and keeps the old signature.
func tamper(t *testing.T, key *rsa.PrivateKey) struct {
	token string
	code  string
	want  Subject
} {
	parts := strings.Split(sign(t, key, goodHeader(), claimsFor("ada@jury.example")), ".")
	forged, _ := json.Marshal(claimsFor("root@jury.example"))
	parts[1] = b64(forged)
	return struct {
		token string
		code  string
		want  Subject
	}{strings.Join(parts, "."), CodeAssertionSignature, Subject{}}
}

func TestSigningKeysAreCachedAndFetchedWithinBounds(t *testing.T) {
	a, b := testKeys(t)
	j := newJWKS(t, jwk(testKid, a))
	c := &clock{t: t0}
	v := verifierFor(j, c)
	good := sign(t, a, goodHeader(), claimsFor("ada@jury.example"))

	for i := 0; i < 5; i++ {
		if _, err := v.Verify(context.Background(), good); err != nil {
			t.Fatal(err)
		}
	}
	if j.count() != 1 {
		t.Fatalf("the key set was fetched %d times for 5 calls; want 1", j.count())
	}

	// A caller who invents key ids cannot make the gateway fetch on every call.
	evil := sign(t, b, map[string]any{"alg": "RS256", "kid": "kid-evil"}, claimsFor("ada@jury.example"))
	c.add(2 * time.Minute)
	for i := 0; i < 20; i++ {
		if _, err := v.Verify(context.Background(), evil); refusal(t, err).Code != CodeAssertionKey {
			t.Fatal(err)
		}
	}
	if j.count() != 2 {
		t.Fatalf("20 unknown key ids caused %d fetches; want exactly 1 more", j.count()-1)
	}

	// A rotated key is picked up by the next unknown id after the bound.
	j.set(200, jwk(testKid, a), jwk("kid-two", b))
	rotated := sign(t, b, map[string]any{"alg": "RS256", "kid": "kid-two"}, claimsFor("ada@jury.example"))
	if _, err := v.Verify(context.Background(), rotated); refusal(t, err).Code != CodeAssertionKey {
		t.Fatal("a rotated key was trusted before the refresh bound")
	}
	c.add(2 * time.Minute)
	if _, err := v.Verify(context.Background(), rotated); err != nil {
		t.Fatalf("a rotated key was not picked up: %v", err)
	}
}

func TestKeysThatCannotBeFetchedAreNamed(t *testing.T) {
	a, _ := testKeys(t)
	j := newJWKS(t, jwk(testKid, a))
	j.set(500)
	c := &clock{t: t0}
	v := verifierFor(j, c)
	good := sign(t, a, goodHeader(), claimsFor("ada@jury.example"))

	_, err := v.Verify(context.Background(), good)
	if ref := refusal(t, err); ref.Code != CodeKeysUnavailable || ref.Status != http.StatusServiceUnavailable {
		t.Fatalf("no keys: %+v", ref)
	}
	// Every call in the bound after that is refused without another fetch.
	for i := 0; i < 5; i++ {
		_, _ = v.Verify(context.Background(), good)
	}
	if j.count() != 1 {
		t.Fatalf("a down key endpoint was called %d times in a minute", j.count())
	}
	// Once it answers again, the next call after the bound works.
	j.set(200, jwk(testKid, a))
	c.add(2 * time.Minute)
	if _, err := v.Verify(context.Background(), good); err != nil {
		t.Fatalf("after recovery: %v", err)
	}
	// A stale set stays in use while a refresh fails.
	j.set(500)
	c.add(11 * time.Minute)
	late := claimsFor("ada@jury.example")
	late["exp"] = t0.Add(24 * time.Hour).Unix()
	if _, err := v.Verify(context.Background(), sign(t, a, goodHeader(), late)); err != nil {
		t.Fatalf("a stale key set was dropped while its refresh failed: %v", err)
	}
}

func TestKeyOfTooFewBitsIsNotTrusted(t *testing.T) {
	small, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	j := newJWKS(t, jwk(testKid, small))
	v := verifierFor(j, &clock{t: t0})
	_, verr := v.Verify(context.Background(), sign(t, small, goodHeader(), claimsFor("ada@jury.example")))
	if ref := refusal(t, verr); ref.Status != http.StatusServiceUnavailable {
		t.Fatalf("a 1024-bit key: %+v", ref)
	}
}
