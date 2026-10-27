package cloudflare

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// stubToken is the fake API token. Every test greps for it.
const stubToken = "STUBTOKEN-abcDEF0123456789"

type obj = map[string]any

// logged is one request the stub received.
type logged struct {
	Method string
	Path   string // with the query
	Body   string
	Auth   string
}

func (l logged) write() bool { return l.Method != http.MethodGet }

// stub is an in-memory Cloudflare API v4: just the parts the scripts use, with
// the envelope, the filters and the error shape of the real one.
type stub struct {
	t   *testing.T
	srv *httptest.Server
	mu  sync.Mutex

	n     int
	reqs  []logged
	fail  map[string]int // "METHOD /path-suffix" -> Cloudflare error code to answer
	delay time.Duration
	// dnsFirst, when set, is listed first on every DNS listing, whatever its name: the API is not bound to list the
	// asked-for record first, and the scripts must read the record they selected, never the first one.
	dnsFirst  obj
	connected map[string]bool // tunnel id -> has live connections

	Zones     []obj
	Org       obj
	Tunnels   map[string]obj
	Ingress   map[string]any
	DNS       map[string]obj
	IdPs      map[string]obj
	Apps      map[string]obj
	Policies  map[string]map[string]obj
	STokens   map[string]obj
	secretOf  map[string]string
	rotations int
}

func newStub(t *testing.T) *stub {
	t.Helper()
	s := &stub{
		t: t, fail: map[string]int{}, connected: map[string]bool{}, delay: 25 * time.Millisecond,
		Zones:   []obj{{"id": "zone-1", "name": "example.org"}},
		Org:     obj{"auth_domain": "jury.cloudflareaccess.com", "name": "jury"},
		Tunnels: map[string]obj{}, Ingress: map[string]any{}, DNS: map[string]obj{}, IdPs: map[string]obj{},
		Apps: map[string]obj{}, Policies: map[string]map[string]obj{}, STokens: map[string]obj{}, secretOf: map[string]string{},
	}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *stub) base() string { return s.srv.URL + "/client/v4" }

// settle orders the test's own reads and writes of the stub's maps against the
// handler goroutines. The scripts have exited when it is called, so nothing runs
// concurrently, but the race detector cannot see through a subprocess: the lock
// and unlock are the synchronisation.
func (s *stub) settle() {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *stub) id(kind string) string {
	s.n++
	return fmt.Sprintf("%s-%d", kind, s.n)
}

func (s *stub) log() []logged {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]logged(nil), s.reqs...)
}

func (s *stub) writes() []logged {
	var out []logged
	for _, l := range s.log() {
		if l.write() {
			out = append(out, l)
		}
	}
	return out
}

// snapshot is the whole state as JSON, to compare before and after a run.
func (s *stub) snapshot() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.Marshal(obj{"t": s.Tunnels, "i": s.Ingress, "d": s.DNS, "p": s.IdPs, "a": s.Apps, "o": s.Policies, "s": s.STokens})
	return string(b)
}

func reply(w http.ResponseWriter, status int, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(obj{"success": true, "errors": []any{}, "messages": []any{}, "result": result})
}

func refuse(w http.ResponseWriter, status, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(obj{"success": false, "errors": []obj{{"code": code, "message": msg}}, "messages": []any{}, "result": nil})
}

func list[T any](m map[string]T, keep func(T) bool) []T {
	out := []T{}
	for _, v := range m {
		if keep == nil || keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func (s *stub) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reqs = append(s.reqs, logged{r.Method, r.URL.RequestURI(), string(body), r.Header.Get("Authorization")})
	time.Sleep(s.delay) // keeps the request open long enough for a process listing to catch curl
	if r.Header.Get("Authorization") != "Bearer "+stubToken {
		refuse(w, 403, 10000, "Authentication error")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/client/v4")
	for key, code := range s.fail {
		m, suffix, _ := strings.Cut(key, " ")
		if m == r.Method && strings.HasSuffix(path, suffix) {
			refuse(w, 400, code, "injected failure for the test")
			return
		}
	}
	seg := strings.Split(strings.Trim(path, "/"), "/")
	var in obj
	if len(body) > 0 {
		if err := json.Unmarshal(body, &in); err != nil {
			refuse(w, 400, 1004, "the body is not JSON")
			return
		}
	}
	q := r.URL.Query()
	switch {
	case path == "/user/tokens/verify":
		reply(w, 200, obj{"id": "tok", "status": "active"})
	case path == "/zones" && r.Method == "GET":
		reply(w, 200, list(toMap(s.Zones), func(z obj) bool { return q.Get("name") == "" || z["name"] == q.Get("name") }))
	case len(seg) == 3 && seg[0] == "zones" && seg[2] == "dns_records":
		s.dnsCollection(w, r, in, q.Get("name"))
	case len(seg) == 4 && seg[0] == "zones" && seg[2] == "dns_records":
		s.dnsItem(w, r, in, seg[3])
	case len(seg) >= 3 && seg[0] == "accounts":
		s.account(w, r, in, seg[2:], q)
	default:
		refuse(w, 404, 7003, "no route "+path)
	}
}

func toMap(zs []obj) map[string]obj {
	m := map[string]obj{}
	for i, z := range zs {
		m[fmt.Sprint(i)] = z
	}
	return m
}

func (s *stub) dnsCollection(w http.ResponseWriter, r *http.Request, in obj, name string) {
	switch r.Method {
	case "GET":
		found := list(s.DNS, func(d obj) bool { return name == "" || d["name"] == name })
		if s.dnsFirst != nil {
			found = append([]obj{s.dnsFirst}, found...)
		}
		reply(w, 200, found)
	case "POST":
		for _, d := range s.DNS {
			if d["name"] == in["name"] {
				refuse(w, 400, 81053, "A record with that host already exists")
				return
			}
		}
		id := s.id("dns")
		in["id"] = id
		s.DNS[id] = in
		reply(w, 200, in)
	}
}

func (s *stub) dnsItem(w http.ResponseWriter, r *http.Request, in obj, id string) {
	d, ok := s.DNS[id]
	if !ok {
		refuse(w, 404, 81044, "Record does not exist")
		return
	}
	switch r.Method {
	case "PATCH":
		for k, v := range in {
			d[k] = v
		}
		reply(w, 200, d)
	case "DELETE":
		delete(s.DNS, id)
		reply(w, 200, obj{"id": id})
	}
}

func aud(id string) string {
	sum := sha256.Sum256([]byte("aud:" + id))
	return hex.EncodeToString(sum[:])
}

func (s *stub) account(w http.ResponseWriter, r *http.Request, in obj, seg []string, q map[string][]string) {
	get := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	switch {
	case len(seg) == 2 && seg[0] == "tokens" && seg[1] == "verify":
		reply(w, 200, obj{"id": "tok", "status": "active"})

	case len(seg) == 2 && seg[0] == "access" && seg[1] == "organizations":
		if r.Method == "POST" {
			s.Org = obj{"auth_domain": in["auth_domain"], "name": in["name"]}
		}
		if s.Org == nil {
			refuse(w, 400, 12001, "access.api.error.not_enabled")
			return
		}
		reply(w, 200, s.Org)

	// ---- tunnels ----
	case seg[0] == "cfd_tunnel":
		s.tunnel(w, r, in, seg[1:], get)

	// ---- identity providers ----
	case len(seg) >= 2 && seg[0] == "access" && seg[1] == "identity_providers":
		switch {
		case len(seg) == 2 && r.Method == "GET":
			reply(w, 200, list(s.IdPs, nil))
		case len(seg) == 2 && r.Method == "POST":
			id := s.id("idp")
			in["id"] = id
			s.IdPs[id] = in
			reply(w, 200, in)
		case len(seg) == 3 && r.Method == "DELETE":
			delete(s.IdPs, seg[2])
			reply(w, 200, obj{"id": seg[2]})
		}

	// ---- applications and their policies ----
	case len(seg) >= 2 && seg[0] == "access" && seg[1] == "apps":
		s.app(w, r, in, seg[2:])

	// ---- service tokens ----
	case len(seg) >= 2 && seg[0] == "access" && seg[1] == "service_tokens":
		s.serviceToken(w, r, in, seg[2:])

	default:
		refuse(w, 404, 7003, "no route")
	}
}

func (s *stub) tunnel(w http.ResponseWriter, r *http.Request, in obj, seg []string, get func(string) string) {
	switch {
	case len(seg) == 0 && r.Method == "GET":
		reply(w, 200, list(s.Tunnels, func(t obj) bool { return get("name") == "" || t["name"] == get("name") }))
	case len(seg) == 0 && r.Method == "POST":
		id := s.id("tunnel")
		in["id"] = id
		s.Tunnels[id] = in
		reply(w, 200, in)
	case len(seg) == 1 && r.Method == "DELETE":
		if s.connected[seg[0]] {
			refuse(w, 400, 1022, "Cannot delete a tunnel with active connections")
			return
		}
		delete(s.Tunnels, seg[0])
		delete(s.Ingress, seg[0])
		reply(w, 200, obj{"id": seg[0]})
	case len(seg) == 2 && seg[1] == "connections" && r.Method == "DELETE":
		delete(s.connected, seg[0])
		reply(w, 200, nil)
	case len(seg) == 2 && seg[1] == "configurations" && r.Method == "GET":
		reply(w, 200, obj{"tunnel_id": seg[0], "config": s.Ingress[seg[0]]})
	case len(seg) == 2 && seg[1] == "configurations" && r.Method == "PUT":
		s.Ingress[seg[0]] = in["config"]
		reply(w, 200, obj{"tunnel_id": seg[0], "config": in["config"]})
	case len(seg) == 2 && seg[1] == "token" && r.Method == "GET":
		reply(w, 200, "TUNNELRUNTOKEN-"+seg[0])
	default:
		refuse(w, 404, 7003, "no route")
	}
}

func (s *stub) app(w http.ResponseWriter, r *http.Request, in obj, seg []string) {
	switch {
	case len(seg) == 0 && r.Method == "GET":
		reply(w, 200, list(s.Apps, nil))
	case len(seg) == 0 && r.Method == "POST":
		id := s.id("app")
		in["id"], in["aud"] = id, aud(id)
		s.Apps[id] = in
		s.Policies[id] = map[string]obj{}
		reply(w, 200, in)
	case len(seg) == 1 && r.Method == "PUT":
		a, ok := s.Apps[seg[0]]
		if !ok {
			refuse(w, 404, 12006, "application not found")
			return
		}
		in["id"], in["aud"] = seg[0], a["aud"]
		s.Apps[seg[0]] = in
		reply(w, 200, in)
	case len(seg) == 1 && r.Method == "DELETE":
		delete(s.Apps, seg[0])
		delete(s.Policies, seg[0])
		reply(w, 200, obj{"id": seg[0]})
	case len(seg) >= 2 && seg[1] == "policies":
		ps, ok := s.Policies[seg[0]]
		if !ok {
			refuse(w, 404, 12006, "application not found")
			return
		}
		switch {
		case len(seg) == 2 && r.Method == "GET":
			reply(w, 200, list(ps, nil))
		case len(seg) == 2 && r.Method == "POST":
			id := s.id("policy")
			in["id"] = id
			ps[id] = in
			reply(w, 200, in)
		case len(seg) == 3 && r.Method == "PUT":
			in["id"] = seg[2]
			ps[seg[2]] = in
			reply(w, 200, in)
		case len(seg) == 3 && r.Method == "DELETE":
			delete(ps, seg[2])
			reply(w, 200, obj{"id": seg[2]})
		}
	}
}

func (s *stub) serviceToken(w http.ResponseWriter, r *http.Request, in obj, seg []string) {
	switch {
	case len(seg) == 0 && r.Method == "GET":
		out := []obj{}
		for _, t := range s.STokens {
			out = append(out, obj{"id": t["id"], "name": t["name"], "client_id": t["client_id"], "duration": t["duration"]})
		}
		reply(w, 200, out)
	case len(seg) == 0 && r.Method == "POST":
		id := s.id("st")
		in["id"], in["client_id"] = id, "client"+id+".access"
		s.secretOf[id] = "SECRET-" + id + "-0"
		s.STokens[id] = in
		reply(w, 200, obj{"id": id, "name": in["name"], "client_id": in["client_id"], "client_secret": s.secretOf[id]})
	case len(seg) == 1 && r.Method == "DELETE":
		delete(s.STokens, seg[0])
		reply(w, 200, obj{"id": seg[0]})
	case len(seg) == 2 && seg[1] == "rotate" && r.Method == "POST":
		t, ok := s.STokens[seg[0]]
		if !ok {
			refuse(w, 404, 12007, "service token not found")
			return
		}
		s.rotations++
		s.secretOf[seg[0]] = fmt.Sprintf("SECRET-%s-%d", seg[0], s.rotations)
		reply(w, 200, obj{"id": seg[0], "client_id": t["client_id"], "client_secret": s.secretOf[seg[0]]})
	}
}
