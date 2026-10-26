package gateway

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// accessRig is the rig with the Access method on: the gateway trusts the
// assertion, which a test signs with the team's key.
type accessRig struct {
	*rig
	key  *jwks
	team string
}

func newAccessRig(t *testing.T) *accessRig {
	t.Helper()
	a, _ := testKeys(t)
	j := newJWKS(t, jwk(testKid, a))
	var r *rig
	r = newRigWith(t, func(cfg *Config) {
		v := NewVerifier(testTeam, testAud, j.srv.Client(), func() time.Time { r.mu.Lock(); defer r.mu.Unlock(); return r.now })
		v.KeysURL = j.srv.URL
		cfg.Access = v
	})
	return &accessRig{rig: r, key: j, team: testTeam}
}

// via sends a call carrying an assertion and no Authorization header.
func (r *accessRig) via(path, assertion, body string) reply {
	r.t.Helper()
	req, _ := http.NewRequest(http.MethodPost, r.gw.URL+path, strings.NewReader(body))
	if assertion != "" {
		req.Header.Set(HeaderAssertion, assertion)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return reply{resp.StatusCode, string(b), resp.Header}
}

func (r *accessRig) judge(email string) string {
	a, _ := testKeys(r.t)
	return sign(r.t, a, goodHeader(), claimsFor(email))
}

func (r *accessRig) runner(clientID string) string {
	a, _ := testKeys(r.t)
	c := claimsFor("")
	delete(c, "email")
	c["common_name"] = clientID
	return sign(r.t, a, goodHeader(), c)
}

func (r *accessRig) allow(match string, perMin, perDay int) {
	r.t.Helper()
	if err := AddIdentity(r.path, match, t0.Add(24*time.Hour), perMin, perDay, t0); err != nil {
		r.t.Fatal(err)
	}
}

func TestAListedIdentityIsServedAndAnUnlistedOneIsRefused(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 10, 100)

	if rep := r.via(PathModel, r.judge("ada@jury.example"), chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("a listed judge: %d %s", rep.status, rep.body)
	}
	// The mutated field is the e-mail claim: valid assertion, address not on the list.
	rep := r.via(PathModel, r.judge("mallory@jury.example"), chat("meterless/auto"))
	if rep.status != 403 || refusedCode(t, rep) != CodeIdentityNotListed {
		t.Fatalf("an unlisted judge: %d %s", rep.status, rep.body)
	}
	if r.ns.count() != 1 {
		t.Errorf("an unlisted identity reached the model upstream: %d calls", r.ns.count())
	}
	if rep := r.via(PathEvidence, r.judge("mallory@jury.example"), evidence("kev_lookup", kevParams)); rep.status != 403 || r.zt.count() != 0 {
		t.Errorf("an unlisted identity on the evidence route: %d, upstream calls %d", rep.status, r.zt.count())
	}
}

func TestTheEmailIsMatchedWithoutRegardToCase(t *testing.T) {
	r := newAccessRig(t)
	r.allow("Ada@Jury.Example", 10, 100)
	if rep := r.via(PathModel, r.judge("ADA@jury.example"), chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("case: %d %s", rep.status, rep.body)
	}
}

func TestADomainRowAdmitsOnlyThatDomain(t *testing.T) {
	r := newAccessRig(t)
	r.allow("*@jury.example", 10, 100)
	for email, want := range map[string]int{
		"ada@jury.example":          200,
		"bob@jury.example":          200,
		"eve@evil.jury.example":     403, // a subdomain is another domain
		"eve@jury.example.evil.com": 403,
		"eve@xjury.example":         403,
		"eve@example.com":           403,
		"jury.example":              403,
		"a@b@jury.example":          403,
	} {
		rep := r.via(PathModel, r.judge(email), chat("meterless/auto"))
		if rep.status != want {
			t.Errorf("%s: %d %s; want %d", email, rep.status, rep.body, want)
		}
	}
}

func TestADomainRowLimitsEachAddressOnItsOwn(t *testing.T) {
	r := newAccessRig(t)
	r.allow("*@jury.example", 2, 100)
	ada, bob := r.judge("ada@jury.example"), r.judge("bob@jury.example")
	for i := 0; i < 2; i++ {
		if rep := r.via(PathModel, ada, chat("meterless/auto")); rep.status != 200 {
			t.Fatal(rep.status)
		}
	}
	rep := r.via(PathModel, ada, chat("meterless/auto"))
	if rep.status != 429 || refusedCode(t, rep) != CodeRateMinute {
		t.Fatalf("ada's third call: %d %s", rep.status, rep.body)
	}
	if rep := r.via(PathModel, bob, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("bob was limited by ada's use: %d %s", rep.status, rep.body)
	}
}

func TestAnExactRevokedRowCutsOneAddressOfADomain(t *testing.T) {
	r := newAccessRig(t)
	r.allow("*@jury.example", 10, 100)
	r.allow("eve@jury.example", 10, 100)
	if err := RevokeIdentity(r.path, "eve@jury.example", t0); err != nil {
		t.Fatal(err)
	}
	rep := r.via(PathModel, r.judge("eve@jury.example"), chat("meterless/auto"))
	if rep.status != 401 || refusedCode(t, rep) != CodeIdentityRevoked {
		t.Fatalf("eve: %d %s", rep.status, rep.body)
	}
	if rep := r.via(PathModel, r.judge("ada@jury.example"), chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("ada stopped with eve: %d %s", rep.status, rep.body)
	}
}

func TestARevokedRunnerIsRefusedWhileAJudgeKeepsWorking(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 10, 100)
	r.allow("abc123.access", 10, 100)
	runner, judge := r.runner("abc123.access"), r.judge("ada@jury.example")
	for _, tok := range []string{runner, judge} {
		if rep := r.via(PathModel, tok, chat("meterless/auto")); rep.status != 200 {
			t.Fatalf("before revoke: %d %s", rep.status, rep.body)
		}
	}
	// The mutated field is the runner row's revoked_at.
	if err := RevokeIdentity(r.path, "abc123.access", t0); err != nil {
		t.Fatal(err)
	}
	rep := r.via(PathModel, runner, chat("meterless/auto"))
	if rep.status != 401 || refusedCode(t, rep) != CodeIdentityRevoked {
		t.Fatalf("the revoked runner: %d %s", rep.status, rep.body)
	}
	if rep := r.via(PathModel, judge, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("the judge stopped with the runner: %d %s", rep.status, rep.body)
	}
}

func TestAnIdentityExpires(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 10, 100)
	tok := r.judge("ada@jury.example")
	// The assertion is good for an hour; the row for a day. Move past the row only.
	a, _ := testKeys(t)
	late := claimsFor("ada@jury.example")
	late["exp"] = t0.Add(48 * time.Hour).Unix()
	tok = sign(t, a, goodHeader(), late)
	r.advance(24 * time.Hour)
	rep := r.via(PathModel, tok, chat("meterless/auto"))
	if rep.status != 401 || refusedCode(t, rep) != CodeIdentityExpired {
		t.Fatalf("after the row's expiry: %d %s", rep.status, rep.body)
	}
}

func TestARunnerClientIdIsNotAJudgeAndAJudgeAddressIsNotARunner(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 10, 100)
	r.allow("abc123.access", 10, 100)
	a, _ := testKeys(t)
	// A service token whose common name is a listed judge address is not that judge.
	c := claimsFor("")
	delete(c, "email")
	c["common_name"] = "ada@jury.example"
	if rep := r.via(PathModel, sign(t, a, goodHeader(), c), chat("meterless/auto")); rep.status != 403 {
		t.Errorf("a client id shaped like an address: %d %s", rep.status, rep.body)
	}
	// A judge whose e-mail claim is a listed client id is not the runner.
	if rep := r.via(PathModel, r.judge("abc123.access"), chat("meterless/auto")); rep.status != 403 {
		t.Errorf("an e-mail claim shaped like a client id: %d %s", rep.status, rep.body)
	}
}

func TestAccessRefusalsAreNamedAndNothingGoesUpstream(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 10, 100)
	a, b := testKeys(t)
	bad := func(mut func(map[string]any)) string {
		c := claimsFor("ada@jury.example")
		mut(c)
		return sign(t, a, goodHeader(), c)
	}
	for name, tc := range map[string]struct{ token, code string }{
		"no assertion":      {"", CodeAssertionMissing},
		"unknown key":       {sign(t, b, map[string]any{"alg": "RS256", "kid": "kid-evil"}, claimsFor("ada@jury.example")), CodeAssertionKey},
		"expired":           {bad(func(c map[string]any) { c["exp"] = t0.Add(-time.Hour).Unix() }), CodeAssertionExpired},
		"wrong audience":    {bad(func(c map[string]any) { c["aud"] = []string{strings.Repeat("c", 64)} }), CodeAssertionAudience},
		"wrong issuer":      {bad(func(c map[string]any) { c["iss"] = "https://evil.cloudflareaccess.com" }), CodeAssertionIssuer},
		"unknown id, valid": {sign(t, b, goodHeader(), claimsFor("ada@jury.example")), CodeAssertionSignature},
	} {
		for _, path := range []string{PathModel, PathEvidence} {
			rep := r.via(path, tc.token, chat("meterless/auto"))
			if rep.status != 401 || refusedCode(t, rep) != tc.code {
				t.Errorf("%s on %s: %d %s; want %s", name, path, rep.status, rep.body, tc.code)
			}
		}
	}
	if r.ns.count()+r.zt.count() != 0 {
		t.Error("a refused assertion reached an upstream")
	}
}

func TestInAccessModeAJudgeTokenIsNotAWayIn(t *testing.T) {
	r := newAccessRig(t)
	tok := r.issue("judge-ada", 10, 100) // a token row exists, as in a mixed settings file
	rep := r.post(PathModel, tok, chat("meterless/auto"))
	if rep.status != 401 || refusedCode(t, rep) != CodeAssertionMissing {
		t.Fatalf("a token in access mode: %d %s", rep.status, rep.body)
	}
	if r.ns.count() != 0 {
		t.Error("a token reached the upstream in access mode")
	}
}

func TestTheLogNamesTheIdentityAndClassAndNeverTheAssertion(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 10, 100)
	r.allow("abc123.access", 10, 100)
	ada, runner := r.judge("ada@jury.example"), r.runner("abc123.access")
	r.via(PathModel, ada, chat("meterless/auto"))
	r.via(PathEvidence, runner, evidence("kev_lookup", kevParams))
	r.via(PathModel, r.judge("mallory@jury.example"), chat("meterless/auto"))

	out := r.log.String()
	for _, want := range []string{`"identity":"ada@jury.example"`, `"class":"judge"`, `"identity":"abc123.access"`, `"class":"runner"`, `"identity":"mallory@jury.example"`, `"refused":"identity_not_listed"`} {
		if !strings.Contains(out, want) {
			t.Errorf("the log lacks %s:\n%s", want, out)
		}
	}
	for _, secret := range []string{ada, runner, nightshiftSecret, zetarisSecret, codeMarker, strings.Split(ada, ".")[2]} {
		if strings.Contains(out, secret) {
			t.Errorf("the log holds a secret or a body")
		}
	}
	if strings.Contains(out, `"token"`) {
		t.Errorf("the log names a token in the Access method:\n%s", out)
	}
}

func TestAnAlarmNamesTheIdentity(t *testing.T) {
	r := newAccessRig(t)
	r.allow("ada@jury.example", 5, 100)
	ada := r.judge("ada@jury.example")
	for i := 0; i < 4; i++ {
		r.via(PathModel, ada, chat("meterless/auto"))
	}
	out := r.alarm.String()
	if n := strings.Count(out, AlarmPrefix+" identity=ada@jury.example window=minute used=4 limit=5"); n != 1 {
		t.Fatalf("alarm lines: %d\n%s", n, out)
	}
}

func TestIdentityRowsAreHeldToTheirRules(t *testing.T) {
	path := t.TempDir() + "/s.json"
	exp := t0.Add(time.Hour)
	for name, tc := range map[string]struct {
		match      string
		perMin     int
		perDay     int
		expires    time.Time
		wantErrHas string
	}{
		"no shape":                 {"not valid!", 1, 1, exp, "match"},
		"an upper-case domain row": {"*@Jury.Example", 1, 1, exp, ""},
		"no limit":                 {"ada@jury.example", 0, 1, exp, "per_minute"},
		"in the past":              {"ada@jury.example", 1, 1, t0.Add(-time.Hour), "future"},
		"bare wildcard":            {"*", 1, 1, exp, "match"},
		"wildcard local":           {"a*@jury.example", 1, 1, exp, "match"},
	} {
		err := AddIdentity(path, tc.match, tc.expires, tc.perMin, tc.perDay, t0)
		if tc.wantErrHas == "" {
			if err != nil {
				t.Errorf("%s: %v", name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.wantErrHas) {
			t.Errorf("%s: %v; want an error naming %q", name, err, tc.wantErrHas)
		}
	}
	if err := AddIdentity(path, "*@jury.example", exp, 1, 1, t0); err == nil || !strings.Contains(err.Error(), "twice") {
		t.Errorf("a duplicate row: %v", err)
	}
	if err := RevokeIdentity(path, "nobody@jury.example", t0); err == nil || !strings.Contains(err.Error(), "no identity") {
		t.Errorf("revoke of an unknown row: %v", err)
	}
	// A class that does not fit the match is refused when the file is read.
	bad := `{"version":1,"tokens":[],"identities":[{"match":"abc.access","class":"judge","expires":"2026-10-28T00:00:00Z","per_minute":1,"per_day":1}]}`
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ListIdentities(path); err == nil || !strings.Contains(err.Error(), "class") {
		t.Errorf("a runner row written as a judge: %v", err)
	}
}

func TestLoadDefaultsToAccessAndRefusesItsMissingSettingsByName(t *testing.T) {
	file := t.TempDir() + "/s.json"
	if err := os.WriteFile(file, []byte(`{"version":1,"tokens":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	base := map[string]string{
		EnvTokensFile: file, EnvNightshiftToken: "test-token-0001",
		"ZETARIS_REST_URL": "http://127.0.0.1:1", "ZETARIS_API_KEY": "test-key-00002", "ZETARIS_ORG_ID": "42",
	}
	load := func(extra map[string]string) (Settings, error) {
		return Load(func(k string) string {
			if v, ok := extra[k]; ok {
				return v
			}
			return base[k]
		})
	}
	var ce *ConfigError
	if _, err := load(nil); !errors.As(err, &ce) || ce.Field != EnvAccessTeam {
		t.Errorf("the default method without a team: %v", err)
	}
	if _, err := load(map[string]string{EnvAccessTeam: "jury.cloudflareaccess.com"}); !errors.As(err, &ce) || ce.Field != EnvAccessAud {
		t.Errorf("without an AUD tag: %v", err)
	}
	if _, err := load(map[string]string{EnvAccessTeam: "evil.example.com", EnvAccessAud: testAud}); !errors.As(err, &ce) || ce.Field != EnvAccessTeam {
		t.Errorf("a team outside cloudflareaccess.com: %v", err)
	}
	if _, err := load(map[string]string{EnvAuth: "none"}); !errors.As(err, &ce) || ce.Field != EnvAuth {
		t.Errorf("an unknown method: %v", err)
	}
	s, err := load(map[string]string{EnvAccessTeam: "jury.cloudflareaccess.com", EnvAccessAud: testAud})
	if err != nil || s.Config.Access == nil {
		t.Fatalf("access settings: %v", err)
	}
	if s.Config.Access.KeysURL != "https://jury.cloudflareaccess.com/cdn-cgi/access/certs" || s.Config.Access.Issuer != "https://jury.cloudflareaccess.com" {
		t.Errorf("the verifier points at %s, issuer %s", s.Config.Access.KeysURL, s.Config.Access.Issuer)
	}
	s, err = load(map[string]string{EnvAuth: AuthToken})
	if err != nil || s.Config.Access != nil {
		t.Fatalf("the token method: %v, access %v", err, s.Config.Access)
	}
}
