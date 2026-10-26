// Package cloudflare tests the scripts in this directory against a stub of the
// Cloudflare API. There is no network and no real credential: the stub answers
// on a loopback port and the token is a fake.
package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const accountID = "0123456789abcdef0123456789abcdef"

// rig is a stub, a settings file and a place to run the scripts from.
type rig struct {
	t        *testing.T
	stub     *stub
	dir      string
	settings string
	state    string
	extra    string // lines added to the settings file
	hits     atomic.Int32
	sawCurl  atomic.Bool
}

func newRig(t *testing.T) *rig {
	t.Helper()
	r := &rig{t: t, stub: newStub(t), dir: t.TempDir()}
	r.state = filepath.Join(r.dir, "state")
	r.settings = filepath.Join(r.dir, "settings")
	r.writeSettings("judges = Ada@Jury.Example, *@panel.example")
	return r
}

func (r *rig) writeSettings(judges string) {
	r.t.Helper()
	text := "account_id = " + accountID + "\nzone = example.org\nhostname = gw.example.org\n" + judges + "\nstate_dir = " + r.state + "\n" + r.extra
	if err := os.WriteFile(r.settings, []byte(text), 0o600); err != nil {
		r.t.Fatal(err)
	}
}

type result struct {
	out, err string
	code     int
}

func (x result) all() string { return x.out + x.err }

// sh runs one of the scripts and watches the process list while it runs: any
// command line that holds the token is counted.
func (r *rig) sh(script string, args ...string) result {
	r.t.Helper()
	return r.shWith(stubToken, script, args...)
}

func (r *rig) shWith(token, script string, args ...string) result {
	r.t.Helper()
	r.stub.settle()
	defer r.stub.settle()
	cmd := exec.Command("sh", append([]string{script}, args...)...)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + r.dir, "TMPDIR=" + os.TempDir(), "CLOUDFLARE_API_BASE=" + r.stub.base()}
	if token != "" {
		cmd.Env = append(cmd.Env, "CLOUDFLARE_API_TOKEN="+token)
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	ctx, stop := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.watch(ctx)
	}()
	runErr := cmd.Run()
	stop()
	<-done
	res := result{out: out.String(), err: errb.String()}
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		res.code = ee.ExitCode()
	} else if runErr != nil {
		r.t.Fatal(runErr)
	}
	return res
}

// watch reads every process's command line again and again until stopped.
func (r *rig) watch(ctx context.Context) {
	for ctx.Err() == nil {
		entries, _ := os.ReadDir("/proc")
		for _, e := range entries {
			if e.Name()[0] < '0' || e.Name()[0] > '9' {
				continue
			}
			raw, err := os.ReadFile("/proc/" + e.Name() + "/cmdline")
			if err != nil {
				continue
			}
			line := string(raw)
			if strings.Contains(line, "curl") {
				r.sawCurl.Store(true)
			}
			if strings.Contains(line, stubToken) || strings.Contains(line, "SECRET-") {
				r.hits.Add(1)
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func script(name string) string { return name + ".sh" }

func (r *rig) up(mode string) result { return r.sh(script("up"), mode, r.settings) }
func (r *rig) down(mode string, extra ...string) result {
	return r.sh(script("down"), append([]string{mode, r.settings}, extra...)...)
}

func mustOK(t *testing.T, x result, what string) {
	t.Helper()
	if x.code != 0 {
		t.Fatalf("%s exited %d\nstdout:\n%s\nstderr:\n%s", what, x.code, x.out, x.err)
	}
}

func (r *rig) read(name string) string {
	b, err := os.ReadFile(filepath.Join(r.state, name))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(b)
}

func (r *rig) mode(name string) os.FileMode {
	st, err := os.Stat(filepath.Join(r.state, name))
	if err != nil {
		r.t.Fatal(err)
	}
	return st.Mode().Perm()
}

func TestUpCreatesEverythingByNameAndASecondRunChangesNothing(t *testing.T) {
	r := newRig(t)
	x := r.up("apply")
	mustOK(t, x, "up apply")

	s := r.stub
	if len(s.Tunnels) != 1 || len(s.DNS) != 1 || len(s.Apps) != 1 || len(s.STokens) != 1 || len(s.IdPs) != 1 {
		t.Fatalf("objects: tunnels %d dns %d apps %d tokens %d idps %d", len(s.Tunnels), len(s.DNS), len(s.Apps), len(s.STokens), len(s.IdPs))
	}
	var tid, appID, stID string
	for id := range s.Tunnels {
		tid = id
	}
	for id := range s.Apps {
		appID = id
	}
	for id := range s.STokens {
		stID = id
	}
	// the ingress: the one host to the loopback port, everything else 404
	want, _ := json.Marshal([]obj{{"hostname": "gw.example.org", "service": "http://127.0.0.1:8089"}, {"service": "http_status:404"}})
	got, _ := json.Marshal(s.Ingress[tid].(obj)["ingress"])
	if string(got) != string(want) {
		t.Errorf("ingress %s, want %s", got, want)
	}
	// the DNS record: a proxied CNAME to the tunnel, tagged as ours
	for _, d := range s.DNS {
		if d["type"] != "CNAME" || d["content"] != tid+".cfargotunnel.com" || d["proxied"] != true || d["comment"] != "managed-by:proven-judges-gateway" {
			t.Errorf("dns record %v", d)
		}
	}
	// the application and its two policies
	app := s.Apps[appID]
	if app["domain"] != "gw.example.org" || app["type"] != "self_hosted" || app["session_duration"] != "24h" {
		t.Errorf("app %v", app)
	}
	var judges, runner obj
	for _, p := range s.Policies[appID] {
		switch p["name"] {
		case "proven-judges-gateway judges":
			judges = p
		case "proven-judges-gateway runner":
			runner = p
		}
	}
	jInc, _ := json.Marshal(judges["include"])
	if judges["decision"] != "allow" || string(jInc) != `[{"email":{"email":"ada@jury.example"}},{"email_domain":{"domain":"panel.example"}}]` {
		t.Errorf("judges policy %v", judges)
	}
	rInc, _ := json.Marshal(runner["include"])
	if runner["decision"] != "non_identity" || string(rInc) != `[{"service_token":{"token_id":"`+stID+`"}}]` {
		t.Errorf("runner policy %v", runner)
	}
	// the files: the gateway's settings, the runner's secret once, the connector token
	env := r.read("gateway-access.env")
	for _, line := range []string{"JUDGES_GATEWAY_AUTH=access", "JUDGES_GATEWAY_ACCESS_TEAM=jury.cloudflareaccess.com", "JUDGES_GATEWAY_ACCESS_AUD=" + aud(appID)} {
		if !strings.Contains(env, line+"\n") {
			t.Errorf("the gateway's settings lack %s:\n%s", line, env)
		}
	}
	secret := r.read("runner-service-token.env")
	if !strings.Contains(secret, "CF_ACCESS_CLIENT_ID=client"+stID+".access\n") || !strings.Contains(secret, "CF_ACCESS_CLIENT_SECRET="+s.secretOf[stID]+"\n") {
		t.Errorf("the runner's secret file: %q", secret)
	}
	for _, f := range []string{"runner-service-token.env", "gateway-access.env", "tunnel.token"} {
		if m := r.mode(f); m != 0o600 {
			t.Errorf("%s has mode %o, want 600", f, m)
		}
	}
	// the secret is shown nowhere
	if strings.Contains(x.all(), s.secretOf[stID]) {
		t.Error("the runner's secret was printed")
	}

	// A second run reads, and writes nothing.
	before, writes := s.snapshot(), len(s.writes())
	y := r.up("apply")
	mustOK(t, y, "second up apply")
	if after := s.snapshot(); after != before {
		t.Errorf("the second run changed the stub:\n%s\n%s", before, after)
	}
	if n := len(s.writes()) - writes; n != 0 {
		t.Errorf("the second run made %d write requests", n)
	}
	if strings.Contains(y.out, "do:") || strings.Contains(y.out, "plan:") {
		t.Errorf("the second run reported a change:\n%s", y.out)
	}
}

func TestPlanMakesNoWriteRequest(t *testing.T) {
	r := newRig(t)
	x := r.up("plan")
	mustOK(t, x, "up plan on an empty account")
	if w := r.stub.writes(); len(w) != 0 {
		t.Fatalf("plan made %d write requests, first %s %s", len(w), w[0].Method, w[0].Path)
	}
	for _, want := range []string{"would create the tunnel", "would create the DNS CNAME", "would create the Access application", "would create the service token", "would create the policy proven-judges-gateway judges", "would create the policy proven-judges-gateway runner"} {
		if !strings.Contains(x.out, want) {
			t.Errorf("the plan lacks %q:\n%s", want, x.out)
		}
	}
	if _, err := os.Stat(r.state); err == nil {
		t.Error("plan made the state directory")
	}

	// A plan against a built account names only what differs.
	mustOK(t, r.up("apply"), "apply")
	r.writeSettings("judges = Ada@Jury.Example, *@panel.example, new@judge.example")
	writes := len(r.stub.writes())
	y := r.up("plan")
	mustOK(t, y, "plan after a change")
	if n := len(r.stub.writes()) - writes; n != 0 {
		t.Errorf("plan after a change made %d writes", n)
	}
	if !strings.Contains(y.out, "plan: would update the policy proven-judges-gateway judges") || strings.Count(y.out, "plan:") != 1 {
		t.Errorf("the plan should name the one policy:\n%s", y.out)
	}
}

func TestUpUpdatesOnlyWhatDiffers(t *testing.T) {
	r := newRig(t)
	mustOK(t, r.up("apply"), "apply")
	r.extra = "gateway_port = 9100\nsession_duration = 12h\n"
	r.writeSettings("judges = Ada@Jury.Example, *@panel.example, new@judge.example")
	before := len(r.stub.writes())
	mustOK(t, r.up("apply"), "apply with changes")
	var got []string
	for _, w := range r.stub.writes()[before:] {
		got = append(got, w.Method+" "+regexp.MustCompile(`/(cfd_tunnel|apps|policies)/[^/]+`).ReplaceAllString(strings.TrimPrefix(w.Path, "/client/v4/accounts/"+accountID), "/$1/ID"))
	}
	want := []string{"PUT /cfd_tunnel/ID/configurations", "PUT /access/apps/ID", "PUT /access/apps/ID/policies/ID"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("writes %v, want %v", got, want)
	}
}

func TestDownRemovesOnlyWhatUpMade(t *testing.T) {
	r := newRig(t)
	s := r.stub
	// Things that were there before: an identity provider, a tunnel, an app, a record.
	s.IdPs["idp-old"] = obj{"id": "idp-old", "type": "onetimepin", "name": "One-time PIN"}
	s.Tunnels["tunnel-old"] = obj{"id": "tunnel-old", "name": "other-tunnel"}
	s.Apps["app-old"] = obj{"id": "app-old", "name": "other-app", "domain": "other.example.org", "aud": aud("app-old")}
	s.Policies["app-old"] = map[string]obj{}
	s.DNS["dns-old"] = obj{"id": "dns-old", "type": "CNAME", "name": "other.example.org", "content": "x.example.net"}
	mustOK(t, r.up("apply"), "up")
	if len(s.IdPs) != 1 {
		t.Fatalf("up made a second identity provider")
	}
	for id := range s.Tunnels {
		s.connected[id] = true // the connector is running
	}

	p := r.down("plan")
	mustOK(t, p, "down plan")
	if len(s.writes()) == 0 {
		t.Fatal("the stub saw no writes at all; the test is blind")
	}
	snap := s.snapshot()
	mustOK(t, r.down("plan"), "down plan again")
	if s.snapshot() != snap || strings.Count(p.out, "plan: would delete") != 6 {
		t.Errorf("down plan:\n%s", p.out)
	}

	x := r.down("apply")
	mustOK(t, x, "down apply")
	if len(s.Tunnels) != 1 || s.Tunnels["tunnel-old"] == nil {
		t.Errorf("tunnels left: %v", s.Tunnels)
	}
	if len(s.Apps) != 1 || s.Apps["app-old"] == nil {
		t.Errorf("apps left: %v", s.Apps)
	}
	if len(s.DNS) != 1 || s.DNS["dns-old"] == nil {
		t.Errorf("dns left: %v", s.DNS)
	}
	if len(s.STokens) != 0 {
		t.Errorf("service tokens left: %v", s.STokens)
	}
	if len(s.IdPs) != 1 || s.IdPs["idp-old"] == nil {
		t.Errorf("the identity provider was not ours and must stay: %v", s.IdPs)
	}
	if !strings.Contains(x.out, "keep: the e-mail one-time-PIN identity provider was not made by up.sh") {
		t.Errorf("down did not say it left the provider:\n%s", x.out)
	}
	for _, f := range []string{"runner-service-token.env", "gateway-access.env", "tunnel.token"} {
		if _, err := os.Stat(filepath.Join(r.state, f)); err == nil {
			t.Errorf("down left %s", f)
		}
	}
	if st := strings.TrimSpace(r.read("state.env")); st != "" {
		t.Errorf("the state file still lists %q", st)
	}
	// And a second down has nothing to do.
	n := len(s.writes())
	mustOK(t, r.down("apply"), "second down")
	if len(s.writes()) != n {
		t.Error("the second down wrote")
	}
}

func TestDownLeavesAnObjectWithOurNameThatUpDidNotMake(t *testing.T) {
	r := newRig(t)
	s := r.stub
	// An app with our name, made by hand before up.sh ever ran, and a record
	// at our host without our tag.
	s.Apps["app-hand"] = obj{"id": "app-hand", "name": "proven-judges-gateway", "domain": "gw.example.org", "aud": aud("app-hand"), "session_duration": "24h", "allowed_idps": []string{"idp-old"}}
	s.Policies["app-hand"] = map[string]obj{}
	s.IdPs["idp-old"] = obj{"id": "idp-old", "type": "onetimepin"}
	s.DNS["dns-hand"] = obj{"id": "dns-hand", "type": "CNAME", "name": "gw.example.org", "content": "x.example.net", "comment": "mine"}

	x := r.up("apply")
	if x.code == 0 || !strings.Contains(x.err, "did not make") {
		t.Fatalf("up over a record that is not ours: %d\n%s%s", x.code, x.out, x.err)
	}
	if s.DNS["dns-hand"]["content"] != "x.example.net" {
		t.Error("up changed a record that is not ours")
	}
	// The failed run made the tunnel before it stopped, so the state lists it;
	// the app and the record carry our names but are not in the state: kept.
	d := r.down("apply")
	mustOK(t, d, "down after a failed up")
	if len(s.Tunnels) != 0 {
		t.Error("down left the tunnel that up made")
	}
	if len(s.Apps) != 1 || s.DNS["dns-hand"] == nil || !strings.Contains(d.out, "keep: Access application proven-judges-gateway was not made by up.sh") {
		t.Errorf("down touched what up did not make:\n%s", d.out)
	}
	// Without a state file down refuses; with the name alone it deletes the
	// app by name, and still keeps a record that lacks the tag.
	if err := os.Remove(filepath.Join(r.state, "state.env")); err != nil {
		t.Fatal(err)
	}
	if n := r.down("apply"); n.code == 0 || !strings.Contains(n.err, "no state file") {
		t.Errorf("down without state: %d %s", n.code, n.err)
	}
	byName := r.down("apply", "--by-name")
	mustOK(t, byName, "down --by-name")
	if len(s.Apps) != 0 {
		t.Error("--by-name did not take the app that carries our name")
	}
	if s.DNS["dns-hand"] == nil || !strings.Contains(byName.out, "keep: DNS record gw.example.org was not made by up.sh") {
		t.Errorf("the untagged record was not kept:\n%s", byName.out)
	}
}

func TestDownRunnerRemovesOnlyTheRunner(t *testing.T) {
	r := newRig(t)
	mustOK(t, r.up("apply"), "up")
	mustOK(t, r.down("runner"), "down runner")
	s := r.stub
	if len(s.STokens) != 0 {
		t.Error("the service token is still there")
	}
	var names []string
	for _, ps := range s.Policies {
		for _, p := range ps {
			names = append(names, p["name"].(string))
		}
	}
	if strings.Join(names, ",") != "proven-judges-gateway judges" || len(s.Apps) != 1 || len(s.Tunnels) != 1 || len(s.DNS) != 1 {
		t.Errorf("left: policies %v apps %d tunnels %d dns %d", names, len(s.Apps), len(s.Tunnels), len(s.DNS))
	}
	if _, err := os.Stat(filepath.Join(r.state, "runner-service-token.env")); err == nil {
		t.Error("the runner's secret file is still there")
	}
}

func TestRotateRunnerIssuesANewSecretToTheFileOnly(t *testing.T) {
	r := newRig(t)
	mustOK(t, r.up("apply"), "up")
	first := r.read("runner-service-token.env")
	x := r.sh("up.sh", "rotate-runner", r.settings)
	mustOK(t, x, "rotate")
	second := r.read("runner-service-token.env")
	if first == second || !strings.Contains(second, "-1\n") {
		t.Errorf("the secret did not change: %q -> %q", first, second)
	}
	if m := r.mode("runner-service-token.env"); m != 0o600 {
		t.Errorf("mode %o", m)
	}
	if strings.Contains(x.all(), "SECRET-") {
		t.Error("the new secret was printed")
	}
}

// TestTheTokenAndTheSecretAreNeverShown is the control for the credential: it
// greps every output, every file the scripts wrote and a sample of the process
// list for the stub token, and shows the sampler works by seeing curl.
func TestTheTokenAndTheSecretAreNeverShown(t *testing.T) {
	r := newRig(t)
	var outputs []string
	for _, x := range []result{r.up("plan"), r.up("apply"), r.up("apply"), r.sh("up.sh", "rotate-runner", r.settings), r.down("plan"), r.down("runner"), r.down("apply")} {
		mustOK(t, x, "a script")
		outputs = append(outputs, x.all())
	}
	// an error run too
	r.stub.fail["POST /access/apps"] = 12130
	failed := r.up("apply")
	_ = failed
	r.stub.fail = map[string]int{}
	outputs = append(outputs, failed.all())

	all := strings.Join(outputs, "\n")
	if strings.Contains(all, stubToken) {
		t.Error("the API token appeared in a script's output")
	}
	if strings.Contains(all, "SECRET-") {
		t.Error("a service-token secret appeared in a script's output")
	}
	if !r.sawCurl.Load() {
		t.Error("the process sampler never saw curl: its zero proves nothing")
	}
	if n := r.hits.Load(); n != 0 {
		t.Errorf("the token or a secret was on a command line in %d samples", n)
	}
	// The stub did receive the token on every request, in the header.
	reqs := r.stub.log()
	if len(reqs) == 0 {
		t.Fatal("no requests")
	}
	for _, l := range reqs {
		if l.Auth != "Bearer "+stubToken {
			t.Fatalf("a request without the token: %s %s", l.Method, l.Path)
		}
		if strings.Contains(l.Path, stubToken) {
			t.Fatalf("the token is in a URL: %s", l.Path)
		}
	}
	// And no file under the state directory holds the API token.
	_ = filepath.Walk(r.dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			if b, _ := os.ReadFile(p); strings.Contains(string(b), stubToken) {
				t.Errorf("%s holds the API token", p)
			}
		}
		return nil
	})
}

func TestEveryAPIErrorIsPrintedByCodeAndMessage(t *testing.T) {
	r := newRig(t)
	r.stub.fail["POST /access/apps"] = 12130
	x := r.up("apply")
	if x.code == 0 {
		t.Fatal("up went on after an API error")
	}
	if !strings.Contains(x.err, "Cloudflare error 12130: injected failure for the test") || !strings.Contains(x.err, "POST /accounts/"+accountID+"/access/apps") {
		t.Errorf("the error is not named:\n%s", x.err)
	}
	if strings.Contains(x.all(), stubToken) {
		t.Error("the token is in the error text")
	}
	// The objects made before the failure are recorded, so down can take them away.
	if st := r.read("state.env"); !strings.Contains(st, "tunnel=") || !strings.Contains(st, "dns=") {
		t.Errorf("state after a failed run: %q", st)
	}

	// A wrong token is named by its Cloudflare code, and never echoed.
	y := r.shWith("WRONGTOKEN-0123", "up.sh", "plan", r.settings)
	if y.code == 0 || !strings.Contains(y.err, "Cloudflare error 10000: Authentication error") || strings.Contains(y.all(), "WRONGTOKEN") {
		t.Errorf("wrong token: %d\n%s", y.code, y.err)
	}
	z := r.shWith("", "up.sh", "plan", r.settings)
	if z.code == 0 || !strings.Contains(z.err, "no API token") {
		t.Errorf("no token: %d %s", z.code, z.err)
	}
}

func TestTheTokenMayComeFromAFile(t *testing.T) {
	r := newRig(t)
	tokenFile := filepath.Join(r.dir, "token")
	if err := os.WriteFile(tokenFile, []byte(stubToken+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r.extra = "token_file = " + tokenFile + "\n"
	r.writeSettings("judges = ada@jury.example")
	x := r.shWith("", "up.sh", "plan", r.settings)
	mustOK(t, x, "plan with a token file")
	if strings.Contains(x.all(), stubToken) {
		t.Error("the token was printed")
	}
}

func TestSettingsAreRefusedByNameBeforeAnyRequest(t *testing.T) {
	for name, tc := range map[string]struct{ text, want string }{
		"unknown key":      {"colour = red\n", "unknown key colour"},
		"empty judges":     {"judges =\n", "judges is empty"},
		"judge not mail":   {"judges = ada\n", "not an e-mail"},
		"host outside":     {"hostname = gw.other.org\n", "inside it"},
		"bad account":      {"account_id = abc\n", "account_id"},
		"bad port":         {"gateway_port = 99999\n", "gateway_port"},
		"bad duration":     {"session_duration = 1d\n", "duration"},
		"bad name":         {"name = Has Space\n", "name"},
		"two at-signs":     {"judges = a@b@c.org\n", "two @"},
		"a line without =": {"just words\n", "without '='"},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			base, _ := os.ReadFile(r.settings)
			// the later line wins, so the bad value follows the good file
			if err := os.WriteFile(r.settings, append(base, []byte(tc.text)...), 0o600); err != nil {
				t.Fatal(err)
			}
			x := r.up("apply")
			if x.code == 0 || !strings.Contains(x.err, tc.want) {
				t.Errorf("exit %d, stderr %q; want it to name %q", x.code, x.err, tc.want)
			}
			if n := len(r.stub.log()); n != 0 {
				t.Errorf("%d requests were made before the settings were refused", n)
			}
		})
	}
}

func TestAMissingOrganisationIsCreatedOnlyWhenTheTeamIsNamed(t *testing.T) {
	r := newRig(t)
	r.stub.Org = nil
	x := r.up("apply")
	if x.code == 0 || !strings.Contains(x.err, "no Zero Trust organisation") || len(r.stub.writes()) != 0 {
		t.Fatalf("without team_name: %d %s (writes %d)", x.code, x.err, len(r.stub.writes()))
	}
	r.extra = "team_name = jury\n"
	r.writeSettings("judges = ada@jury.example")
	mustOK(t, r.up("apply"), "with team_name")
	if !strings.Contains(r.read("gateway-access.env"), "JUDGES_GATEWAY_ACCESS_TEAM=jury.cloudflareaccess.com") {
		t.Error("the new team domain was not written")
	}
}

func TestVerifyShowsAccessInFrontAndTheRunnerThrough(t *testing.T) {
	r := newRig(t)
	mustOK(t, r.up("apply"), "up")
	var id, secret string
	for _, l := range strings.Split(r.read("runner-service-token.env"), "\n") {
		if v, ok := strings.CutPrefix(l, "CF_ACCESS_CLIENT_ID="); ok {
			id = v
		}
		if v, ok := strings.CutPrefix(l, "CF_ACCESS_CLIENT_SECRET="); ok {
			secret = v
		}
	}
	// A fake front door: Access redirects a caller with no credential, and the
	// gateway behind it answers the runner's.
	var open atomic.Bool
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !open.Load() && (req.Header.Get("CF-Access-Client-Id") != id || req.Header.Get("CF-Access-Client-Secret") != secret) {
			w.Header().Set("Location", "https://jury.cloudflareaccess.com/login")
			w.WriteHeader(302)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"refused":"not_a_template_request","detail":"x"}`))
	}))
	defer front.Close()

	run := func() result {
		cmd := exec.Command("sh", "verify.sh", r.settings)
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + r.dir, "TMPDIR=" + os.TempDir(), "VERIFY_BASE=" + front.URL}
		var o, e bytes.Buffer
		cmd.Stdout, cmd.Stderr = &o, &e
		err := cmd.Run()
		res := result{out: o.String(), err: e.String()}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			res.code = ee.ExitCode()
		}
		return res
	}
	x := run()
	mustOK(t, x, "verify")
	if !strings.Contains(x.out, "stopped by Access (HTTP 302)") || !strings.Contains(x.out, "passes Access") || strings.Contains(x.all(), secret) {
		t.Errorf("verify:\n%s%s", x.out, x.err)
	}
	// If Access were not there, an anonymous call is answered: verify fails.
	open.Store(true)
	y := run()
	if y.code == 0 || !strings.Contains(y.out, "FAIL: an anonymous call got HTTP 400") {
		t.Errorf("verify with the door open: %d\n%s", y.code, y.out)
	}
}

func TestUpAddsTheGatewayIdentityRowsOnce(t *testing.T) {
	r := newRig(t)
	bin := filepath.Join(r.dir, "judges-gateway")
	build := exec.Command("go", "build", "-o", bin, "github.com/saphan-ai/proven/cmd/judges-gateway")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	settings := filepath.Join(r.dir, "gateway.json")
	r.extra = "gateway_bin = " + bin + "\ngateway_settings_file = " + settings + "\nidentities_expire = 2099-01-01T00:00:00Z\n"
	r.writeSettings("judges = Ada@Jury.Example, *@panel.example")
	mustOK(t, r.up("apply"), "up")

	list := func() string {
		cmd := exec.Command(bin, "identity", "list")
		cmd.Env = []string{"JUDGES_GATEWAY_TOKENS_FILE=" + settings}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("identity list: %v\n%s", err, out)
		}
		return string(out)
	}
	rows := list()
	class := map[string]string{}
	for _, l := range strings.Split(strings.TrimSpace(rows), "\n")[1:] {
		f := strings.Fields(l)
		class[f[0]] = f[1]
	}
	if len(class) != 3 || class["ada@jury.example"] != "judge" || class["*@panel.example"] != "judge" {
		t.Errorf("the gateway's rows:\n%s", rows)
	}
	runners := 0
	for _, c := range class {
		if c == "runner" {
			runners++
		}
	}
	if runners != 1 {
		t.Errorf("runner rows: %d\n%s", runners, rows)
	}
	y := r.up("apply")
	mustOK(t, y, "second up")
	if list() != rows || strings.Contains(y.out, "do:") {
		t.Errorf("the second run changed the gateway's rows:\n%s", y.out)
	}
}
