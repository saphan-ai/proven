package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// The two routes. The model route has the path Proven's generator already
// calls, so a lab points its base address at the tunnel and changes nothing
// else.
const (
	RouteModel    = "model"
	RouteEvidence = "evidence"
	PathModel     = "/v1/chat/completions"
	PathEvidence  = "/v1/evidence/query"
	PathHealth    = "/healthz"
)

// The refusal codes of the routes.
const (
	CodeNoRoute          = "no_such_route"
	CodeMethod           = "method_not_allowed"
	CodeBodyTooLarge     = "body_too_large"
	CodeNotATemplate     = "not_a_template_request"
	CodeUnknownTemplate  = "unknown_template"
	CodeMissingParameter = "missing_parameter"
	CodeUnknownParameter = "unknown_parameter"
	CodeBadParameter     = "bad_parameter"
	CodeTemplateRefused  = "template_refused"
	CodeBadRequest       = "bad_request"
	CodeModelNotAllowed  = "model_not_allowed"
	CodeUpstreamDown     = "upstream_unavailable"
	CodeUpstreamError    = "upstream_error"
	CodeLogUnavailable   = "log_unavailable"
	CodeResponseTooLarge = "response_over_limit"
)

// Querier runs one SELECT on the evidence store. *query.Live is the one used.
type Querier interface {
	Query(ctx context.Context, sql string) (query.Result, error)
}

// Config is everything a Gateway needs. The two credentials are fields here
// and nowhere else: nothing in this package prints, logs or returns them.
type Config struct {
	Tokens  *Store // the settings file: token rows and identity rows
	Limiter *Limiter

	// Access, when it is set, is the only way in: a call must carry a valid
	// Cloudflare Access assertion for a listed identity. When it is nil the
	// judge's own token is the way in (the fallback method).
	Access *Verifier

	NightshiftURL   string // origin, no path
	NightshiftToken string
	NightshiftApp   string
	Models          []string // models a judge may name; the model route refuses any other
	Upstream        *http.Client

	Zetaris   Querier
	Relations querytpl.Relations

	MaxModelBody    int64 // bytes of one chat request
	MaxEvidenceBody int64 // bytes of one template request
	MaxModelReply   int64 // bytes relayed from the model service in one call; zero means the default

	Log   io.Writer // one JSON line per call
	Alarm io.Writer // alarm lines
	Fault io.Writer // the second channel: a sink that failed or recovered (standard error)
	Now   func() time.Time
}

// Gateway is the http.Handler of both routes.
type Gateway struct {
	cfg       Config
	logMu     sync.Mutex // guards the two sinks and unlogged
	logSink   sink
	alarmSink sink
	unlogged  int // refusals not queued because the sink was down
}

// New builds the gateway. A nil Now means the wall clock.
func New(cfg Config) *Gateway {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Log == nil {
		cfg.Log = io.Discard
	}
	if cfg.Alarm == nil {
		cfg.Alarm = io.Discard
	}
	if cfg.Fault == nil {
		cfg.Fault = io.Discard
	}
	if cfg.MaxModelReply <= 0 {
		cfg.MaxModelReply = MaxModelReply
	}
	return &Gateway{cfg: cfg, logSink: sink{name: "log", w: cfg.Log}, alarmSink: sink{name: "alarm", w: cfg.Alarm}}
}

// entry is one log line. It has no field that can hold a body or a
// credential: Model and Template are written only when the value is on the
// gateway's own list, never as the caller sent them.
type entry struct {
	Time     string `json:"time"`
	Token    string `json:"token,omitempty"`    // the token id, token method only
	Identity string `json:"identity,omitempty"` // the e-mail or client id, Access method only
	Class    string `json:"class,omitempty"`    // judge or runner
	Route    string `json:"route"`
	Model    string `json:"model,omitempty"`
	Template string `json:"template,omitempty"`
	Status   int    `json:"status"`
	Bytes    int64  `json:"bytes"`
	Refused  string `json:"refused,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// recorder counts what the judge receives.
type recorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (r *recorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytes += int64(n)
	return n, err
}

// Unwrap lets the response controller flush a streamed answer.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// ServeHTTP checks the token, the route and the limits, in that order, then
// serves the call. Every call, served or refused, leaves one log line.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == PathHealth {
		g.health(w)
		return
	}
	rec := &recorder{ResponseWriter: w}
	e := entry{Time: g.cfg.Now().UTC().Format(time.RFC3339), Route: "-"}
	defer func() {
		e.Status, e.Bytes = rec.status, rec.bytes
		if e.Refused == CodeLogUnavailable {
			// The sink is down: this refusal is counted, not queued, so the
			// queue holds only calls that were served.
			g.logMu.Lock()
			g.unlogged++
			g.logMu.Unlock()
			return
		}
		g.write(e)
	}()
	refuse := func(ref *Refusal) {
		e.Refused = ref.Code
		writeRefusal(rec, ref)
	}

	// A call the gateway cannot log is not served: if the log or the alarm
	// cannot be written, refuse before anything else happens.
	g.logMu.Lock()
	up := g.sinksUp()
	g.logMu.Unlock()
	if !up {
		refuse(&Refusal{Status: http.StatusServiceUnavailable, Code: CodeLogUnavailable, Detail: "the gateway cannot write its call log; tell the owner"})
		return
	}

	// Who is calling: the Access assertion, or the judge's own token.
	var who limited
	if g.cfg.Access != nil {
		e.Identity = "-"
		sub, err := g.cfg.Access.Verify(r.Context(), r.Header.Get(HeaderAssertion))
		if err != nil {
			refuseWith(refuse, err)
			return
		}
		res, err := g.cfg.Tokens.LookupIdentity(sub, g.cfg.Now())
		if res.Name != "" {
			e.Identity, e.Class = res.Name, res.Class
		}
		if err != nil {
			refuseWith(refuse, err)
			return
		}
		who = limited{kind: "identity", name: res.Name, perMinute: res.PerMinute, perDay: res.PerDay}
	} else {
		e.Token = "-"
		token := bearer(r.Header.Get("Authorization"))
		if token == "" {
			refuse(&Refusal{Status: http.StatusUnauthorized, Code: CodeMissingToken, Detail: "send the judge token as 'Authorization: Bearer <token>'"})
			return
		}
		judge, err := g.cfg.Tokens.Lookup(token, g.cfg.Now())
		if judge.ID != "" {
			e.Token = judge.ID
		}
		if err != nil {
			refuseWith(refuse, err)
			return
		}
		who = limited{kind: "token", name: judge.ID, perMinute: judge.PerMinute, perDay: judge.PerDay}
	}

	var serve func(http.ResponseWriter, *http.Request, *entry) *Refusal
	switch r.URL.Path {
	case PathModel:
		e.Route, serve = RouteModel, g.model
	case PathEvidence:
		e.Route, serve = RouteEvidence, g.evidence
	default:
		refuse(&Refusal{Status: http.StatusNotFound, Code: CodeNoRoute, Detail: "the gateway serves " + PathModel + " and " + PathEvidence + " only"})
		return
	}
	if r.Method != http.MethodPost {
		refuse(&Refusal{Status: http.StatusMethodNotAllowed, Code: CodeMethod, Detail: "use POST"})
		return
	}

	alarms, ref := g.cfg.Limiter.AllowKey(who.kind, who.name, who.perMinute, who.perDay, g.cfg.Now())
	for _, a := range alarms {
		g.alarm(a)
	}
	if ref != nil {
		refuse(ref)
		return
	}
	if ref := serve(rec, r, &e); ref != nil {
		refuse(ref)
	}
}

// limited is the caller as the limiter sees it.
type limited struct {
	kind, name        string
	perMinute, perDay int
}

// refuseWith turns the error of an authentication step into its refusal. An
// error that is not a refusal is a failure of the gateway, not of the caller.
func refuseWith(refuse func(*Refusal), err error) {
	var ref *Refusal
	if !errors.As(err, &ref) {
		ref = &Refusal{Status: http.StatusServiceUnavailable, Code: CodeSettingsDown, Detail: "the gateway cannot check who you are; tell the owner"}
	}
	refuse(ref)
}

// bearer returns the token of an Authorization header, or "".
func bearer(h string) string {
	const p = "bearer "
	if len(h) > len(p) && strings.EqualFold(h[:len(p)], p) {
		return strings.TrimSpace(h[len(p):])
	}
	return ""
}

func (g *Gateway) write(e entry) {
	line, err := json.Marshal(e)
	if err != nil {
		g.fault("%s sink=log error=%q", FaultPrefix, "a log line could not be encoded")
		return
	}
	g.logMu.Lock()
	defer g.logMu.Unlock()
	g.put(&g.logSink, append(line, '\n'))
}

func (g *Gateway) alarm(line string) {
	g.logMu.Lock()
	defer g.logMu.Unlock()
	g.put(&g.alarmSink, []byte(line+"\n"))
}

// health answers the owner's probe: 200 when both sinks write, 503 when not.
// It needs no token and says nothing else.
func (g *Gateway) health(w http.ResponseWriter) {
	if g.Healthy() {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "{\"status\":\"ok\"}\n")
		return
	}
	writeRefusal(w, &Refusal{Status: http.StatusServiceUnavailable, Code: CodeLogUnavailable, Detail: "the gateway cannot write its call log or alarm; tell the owner"})
}

// writeRefusal sends the refusal as JSON. The body holds the code and a fixed
// sentence, never the caller's text.
func writeRefusal(w http.ResponseWriter, ref *Refusal) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ref.Status)
	_ = json.NewEncoder(w).Encode(map[string]string{"refused": ref.Code, "detail": ref.Detail})
}

// readBody reads at most max bytes; one more byte is the refusal.
func readBody(w http.ResponseWriter, r *http.Request, max int64) ([]byte, *Refusal) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, max))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return nil, &Refusal{Status: http.StatusRequestEntityTooLarge, Code: CodeBodyTooLarge, Detail: "the request is larger than the gateway accepts"}
		}
		return nil, &Refusal{Status: http.StatusBadRequest, Code: CodeBadRequest, Detail: "the request body could not be read"}
	}
	return body, nil
}

// evidence serves a Proven query-template request and nothing else. The
// body must be exactly {"template": ..., "params": {...}}: a raw statement,
// an extra field or any other shape is refused by name. The SQL is built here,
// by the same code Proven uses, from values that code has held to their format.
func (g *Gateway) evidence(w http.ResponseWriter, r *http.Request, e *entry) *Refusal {
	body, ref := readBody(w, r, g.cfg.MaxEvidenceBody)
	if ref != nil {
		return ref
	}
	notTemplate := &Refusal{Status: http.StatusBadRequest, Code: CodeNotATemplate, Detail: `send {"template": <name>, "params": {...}}; the gateway runs query templates only, never SQL`}
	var req querytpl.Request
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || dec.More() || req.Empty() {
		return notTemplate
	}
	sql, err := querytpl.Build(req, g.cfg.Relations)
	if err != nil {
		var tr *querytpl.Refusal
		if !errors.As(err, &tr) {
			return &Refusal{Status: http.StatusBadRequest, Code: CodeTemplateRefused, Detail: "the template request was refused"}
		}
		return &Refusal{Status: http.StatusBadRequest, Code: templateCode(tr.Reason), Detail: tr.Error()}
	}
	// Only a name on Proven's list reaches the log.
	e.Template = strings.TrimSpace(req.Template)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	res, err := g.cfg.Zetaris.Query(ctx, sql)
	if err != nil {
		var down *service.Down
		if errors.As(err, &down) {
			e.Reason = down.Reason
		}
		return &Refusal{Status: http.StatusBadGateway, Code: CodeUpstreamDown, Detail: "the evidence store did not answer; tell the owner"}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"template": e.Template,
		"snapshot": res.Snapshot,
		"columns":  res.Columns,
		"rows":     res.Rows,
	})
	return nil
}

func templateCode(reason string) string {
	switch reason {
	case querytpl.ReasonUnknownTemplate:
		return CodeUnknownTemplate
	case querytpl.ReasonMissingParam:
		return CodeMissingParameter
	case querytpl.ReasonUnknownParam:
		return CodeUnknownParameter
	case querytpl.ReasonBadValue:
		return CodeBadParameter
	}
	return CodeTemplateRefused
}

// chatFields are the top-level fields a chat request may carry: the ones Proven's generator sends (internal/generator
// chatRequest) and three common sampling bounds. Nightshift reads the body as it is sent, so the gateway checks and
// forwards one and the same body: any other field (such as a fallback "models" list or a "provider" route that could
// choose another model), a repeated key, or a key in another case is refused, never passed on.
var chatFields = map[string]bool{
	"model": true, "messages": true, "stream": true, "thinking_budget": true,
	"max_tokens": true, "temperature": true, "top_p": true,
}

// chatBody reads a chat request strictly. It returns the model, the body to forward (the checked fields, encoded
// again), or a problem in words.
func chatBody(body []byte) (model string, clean []byte, problem string) {
	const shape = `send an OpenAI-style chat request with a "model"`
	dec := json.NewDecoder(bytes.NewReader(body))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return "", nil, shape
	}
	fields := map[string]json.RawMessage{}
	seen := map[string]bool{}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return "", nil, shape
		}
		key, _ := t.(string)
		if !chatFields[key] {
			return "", nil, fmt.Sprintf("the field %q is not served by this gateway", key)
		}
		if seen[strings.ToLower(key)] {
			return "", nil, fmt.Sprintf("the field %q appears twice", key)
		}
		seen[strings.ToLower(key)] = true
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return "", nil, shape
		}
		fields[key] = v
	}
	if t, err := dec.Token(); err != nil || t != json.Delim('}') {
		return "", nil, shape
	}
	if _, err := dec.Token(); err != io.EOF {
		return "", nil, shape
	}
	if err := json.Unmarshal(fields["model"], &model); err != nil || model == "" {
		return "", nil, shape
	}
	clean, err := json.Marshal(fields)
	if err != nil {
		return "", nil, shape
	}
	return model, clean, ""
}

// model forwards a chat completion to Nightshift with our token. The judge's
// Authorization header is dropped and never copied; the upstream request is
// built fresh from the body and two plain headers. An upstream error is not
// passed on: its body could echo what was sent.
func (g *Gateway) model(w http.ResponseWriter, r *http.Request, e *entry) *Refusal {
	body, ref := readBody(w, r, g.cfg.MaxModelBody)
	if ref != nil {
		return ref
	}
	model, clean, problem := chatBody(body)
	if problem != "" {
		return &Refusal{Status: http.StatusBadRequest, Code: CodeBadRequest, Detail: problem}
	}
	if !g.modelAllowed(model) {
		return &Refusal{Status: http.StatusBadRequest, Code: CodeModelNotAllowed, Detail: "this gateway serves only: " + strings.Join(g.cfg.Models, ", ")}
	}
	e.Model = model

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, g.cfg.NightshiftURL+"/v1/chat/completions", bytes.NewReader(clean))
	if err != nil {
		return &Refusal{Status: http.StatusBadGateway, Code: CodeUpstreamDown, Detail: "the model service is not configured; tell the owner"}
	}
	req.Header.Set("Content-Type", "application/json")
	if a := r.Header.Get("Accept"); a != "" {
		req.Header.Set("Accept", a)
	}
	req.Header.Set("Authorization", "Bearer "+g.cfg.NightshiftToken)
	req.Header.Set("meterless-app", g.cfg.NightshiftApp)
	req.Header.Set("meterless-scope", "user")
	resp, err := g.cfg.Upstream.Do(req)
	if err != nil {
		e.Reason = "unreachable"
		if r.Context().Err() != nil {
			e.Reason = "caller gone"
		}
		return &Refusal{Status: http.StatusBadGateway, Code: CodeUpstreamDown, Detail: "the model service did not answer; tell the owner"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		e.Reason = "http " + http.StatusText(resp.StatusCode)
		return &Refusal{Status: http.StatusBadGateway, Code: CodeUpstreamError, Detail: "the model service refused the call; tell the owner"}
	}
	limit := g.cfg.MaxModelReply
	if resp.ContentLength > limit {
		// Declared too large: refuse before one byte is relayed.
		e.Reason = "declared " + strconv.FormatInt(resp.ContentLength, 10) + " bytes, limit " + strconv.FormatInt(limit, 10)
		return &Refusal{Status: http.StatusBadGateway, Code: CodeResponseTooLarge, Detail: "the model answer is larger than the gateway relays; tell the owner"}
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	ctl := http.NewResponseController(w)
	buf := make([]byte, 8192)
	var sent int64
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			over := sent+int64(n) > limit
			if over {
				chunk = chunk[:limit-sent]
			}
			if len(chunk) > 0 {
				if _, werr := w.Write(chunk); werr != nil {
					return nil
				}
				sent += int64(len(chunk))
				_ = ctl.Flush()
			}
			if over {
				// Cut the stream. The status is already sent, so the end is
				// an aborted connection (the judge's read fails; it never
				// sees a clean end) and a log line that names the reason.
				e.Refused = CodeResponseTooLarge
				e.Reason = "relayed " + strconv.FormatInt(sent, 10) + " bytes, limit " + strconv.FormatInt(limit, 10)
				panic(http.ErrAbortHandler)
			}
		}
		if rerr != nil {
			return nil
		}
	}
}

func (g *Gateway) modelAllowed(m string) bool {
	for _, a := range g.cfg.Models {
		if a == m {
			return true
		}
	}
	return false
}
