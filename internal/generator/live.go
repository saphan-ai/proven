package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

// liveBuilt counts live clients NewLive has constructed in this process.
// Sample mode does not call NewLive, so the count does not move during a
// sample review.
var liveBuilt atomic.Int64

// LiveBuilt is how many live clients NewLive has constructed.
// A sample review leaves it unchanged.
func LiveBuilt() int64 { return liveBuilt.Load() }

// Use is what one completion reported about the model that answered.
// Model is "unknown" when the stream did not name a routed model.
type Use struct {
	Model       string
	Route       string
	ReceiptID   string
	FromReceipt bool
}

// Reported names the model of the completion that just finished.
// A call that failed before a model was known reports false.
type Reported interface {
	LastUse() (Use, bool)
}

// Live calls an OpenAI-compatible chat endpoint and reads the streamed answer.
// Complete is sequential: LastUse reports the call that just returned.
type Live struct {
	cfg    Settings
	client *http.Client
	last   Use
	has    bool
}

// WithModel returns a generator that sends model on the request.
// The two generators do not share the model of the call that just finished.
func (l *Live) WithModel(model string) *Live {
	if l == nil {
		return nil
	}
	cfg := l.cfg
	cfg.Model = model
	return &Live{cfg: cfg, client: l.client}
}

// NewLive builds a live generator. A nil client uses one that does not follow
// redirects, so the runtime token is not sent to another host.
func NewLive(cfg Settings, client *http.Client) *Live {
	liveBuilt.Add(1)
	if client == nil {
		client = &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return &Live{cfg: cfg, client: client}
}

// LastUse reports the model of the call that just finished.
func (l *Live) LastUse() (Use, bool) {
	if l == nil || !l.has {
		return Use{}, false
	}
	return l.last, true
}

// chatRequest is the body of one OpenAI-compatible chat completion request.
type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

// chatMessage is one message of the request. The review sends the whole prompt
// as a single user message.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Complete sends one streamed chat completion.
// An empty answer is tried once more. A wrong token, a timeout, an
// unreachable endpoint, or a second empty answer is a named failure of the
// provider in use (nightshift or nvidia).
// A model outside nvidia/* is refused when this agent requires NVIDIA.
func (l *Live) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if l == nil || l.client == nil || l.cfg.Token == "" || l.cfg.BaseURL == "" || l.cfg.Timeout <= 0 {
		return "", l.down("not configured")
	}
	// Clear the last use first, so a failed call never reports the model of the call before it.
	l.has = false
	l.last = Use{}
	model := l.cfg.ModelFor(agent)
	text, meta, err := l.attempt(ctx, prompt, model)
	if err != nil {
		return "", err
	}
	// An empty answer is retried once with the same prompt and model; a second
	// empty answer is a failure, not an empty review.
	if strings.TrimSpace(text) == "" {
		text, meta, err = l.attempt(ctx, prompt, model)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(text) == "" {
			return "", l.down("empty answer")
		}
	}
	return l.finish(agent, text, meta)
}

// Models is GET /v1/models with the same contract headers as a completion.
// A review does not call it. The completion itself is the call a review makes.
func (l *Live) Models(ctx context.Context) error {
	if l == nil || l.client == nil || l.cfg.Token == "" || l.cfg.BaseURL == "" || l.cfg.Timeout <= 0 {
		return l.down("not configured")
	}
	// parent is kept so a cancellation by the caller is told apart from this call's own timeout.
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, l.cfg.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint(l.cfg.BaseURL, "/models"), nil)
	if err != nil {
		return l.down("not configured")
	}
	l.authorize(req)
	resp, err := l.client.Do(req)
	if err != nil {
		return l.classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	discard(resp.Body)
	if l.rejectsKey(resp.StatusCode) {
		return l.down("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return l.down(httpReason(resp.StatusCode))
	}
	return nil
}

// NoReceipt is the receipt id a call carries in nvidia mode. The hosted API
// signs nothing, so the trace says so instead of inventing an id (ADR-0031
// point 3). FromReceipt is false beside it.
const NoReceipt = "no receipt"

// finish records the model that answered and applies the NVIDIA check.
// ADR-0019 point 4: in nightshift mode the check reads the receipt, because
// Nightshift does not obey the model a request names. With no receipt the model
// is "unknown" and the answer is kept here unless NVIDIA is required.
func (l *Live) finish(agent, text string, meta streamMeta) (string, error) {
	if l.cfg.Mode == ModeNVIDIA {
		return l.finishHosted(agent, text, meta)
	}
	use := Use{Model: "unknown", Route: meta.route, ReceiptID: meta.receipt}
	if meta.model != "" {
		use.Model = meta.model
		use.FromReceipt = true
	}
	// The use is recorded before the check, so a refused answer still names its model in the trace.
	l.last = use
	l.has = true
	if !l.cfg.requiresNVIDIA(agent) {
		return text, nil
	}
	// ADR-0019 point 4: an answer must name its model from the signed receipt.
	// With no receipt the model cannot be shown to be NVIDIA's, so it is refused
	// the same way as a non-NVIDIA model.
	if !use.FromReceipt {
		return "", l.down("answer carries no receipt, refused: the model cannot be shown to be NVIDIA's")
	}
	if !nvidiaModel(use.Model) {
		return "", l.down("answered by " + use.Model + ", refused: not an NVIDIA model")
	}
	return text, nil
}

// finishHosted is finish for nvidia mode (ADR-0031, the master's ruling of
// 2026-10-24). The two modes differ on purpose. Nightshift is a router: it may
// answer with a model other than the one asked for, so ADR-0019 point 4 trusts
// only its signed receipt and refuses an answer without one. Here Proven calls
// NVIDIA directly and no router stands between, so the `model` field of the
// hosted API's own response is the only statement of who answered. It is kept
// as the answering model, recorded with NoReceipt, and the nvidia/* requirement
// is checked against it. A response that names no model cannot be shown to be
// NVIDIA's and is refused where NVIDIA is required.
func (l *Live) finishHosted(agent, text string, meta streamMeta) (string, error) {
	// Route carries the host the answer came from, so the trace's model event
	// names where the trusted model field was read.
	use := Use{Model: "unknown", Route: l.cfg.BaseHost(), ReceiptID: NoReceipt}
	if meta.chunkModel != "" {
		use.Model = meta.chunkModel
	}
	// The use is recorded before the check, so a refused answer still names its model in the trace.
	l.last = use
	l.has = true
	if !l.cfg.requiresNVIDIA(agent) {
		return text, nil
	}
	if use.Model == "unknown" {
		return "", l.down("answer names no model, refused: the model cannot be shown to be NVIDIA's")
	}
	if !nvidiaModel(use.Model) {
		return "", l.down("answered by " + use.Model + ", refused: not an NVIDIA model")
	}
	return text, nil
}

// nvidiaModel reports whether model is "nvidia/" followed by a name.
func nvidiaModel(model string) bool {
	rest, ok := strings.CutPrefix(model, "nvidia/")
	return ok && rest != ""
}

// attempt makes one POST and reads the stream. Each attempt gets its own
// timeout from PROVEN_GENERATOR_TIMEOUT.
func (l *Live) attempt(parent context.Context, prompt, model string) (string, streamMeta, error) {
	if err := parent.Err(); err != nil {
		return "", streamMeta{}, err
	}
	ctx, cancel := context.WithTimeout(parent, l.cfg.Timeout)
	defer cancel()
	payload, err := json.Marshal(chatRequest{
		Model:    model,
		Stream:   true,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", streamMeta{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint(l.cfg.BaseURL, "/chat/completions"), bytes.NewReader(payload))
	if err != nil {
		return "", streamMeta{}, l.down("not configured")
	}
	l.authorize(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := l.client.Do(req)
	if err != nil {
		return "", streamMeta{}, l.classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	if l.rejectsKey(resp.StatusCode) {
		discard(resp.Body)
		// A wrong token is not retried: the next call would be refused the same way.
		return "", streamMeta{}, l.down("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		discard(resp.Body)
		return "", streamMeta{}, l.down(httpReason(resp.StatusCode))
	}
	meta, err := parseStream(resp.Body)
	if err != nil {
		// A stream cut by the caller's cancellation returns that cancellation,
		// not a provider failure.
		if parent.Err() != nil {
			return "", streamMeta{}, parent.Err()
		}
		// The deadline is checked before the stream's own fault on purpose: a body
		// cut by this call's deadline can end on a half event, and the cause worth
		// naming is the timeout, not the fragment it left.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", streamMeta{}, l.down("timeout")
		}
		return "", streamMeta{}, l.down("bad stream")
	}
	return meta.text, meta, nil
}

// authorize sets the request's headers. Nightshift's app contract adds the app
// name that selects the operator's routing mode, and the user scope; the hosted
// API takes the key as a bearer token and nothing else, so those headers are
// not sent to it.
func (l *Live) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+l.cfg.Token)
	if l.cfg.Mode == ModeNVIDIA {
		return
	}
	req.Header.Set("meterless-app", l.cfg.App)
	req.Header.Set("meterless-scope", "user")
}

// rejectsKey reports whether status says the credential was refused. The hosted
// API answers a wrong or unauthorised key with 403 as well as 401.
func (l *Live) rejectsKey(status int) bool {
	return status == http.StatusUnauthorized || (l.cfg.Mode == ModeNVIDIA && status == http.StatusForbidden)
}

// classify names a failed request: the caller's cancellation, a timeout, or an
// unreachable endpoint.
func (l *Live) classify(parent, ctx context.Context, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || timedOut(err) {
		return l.down("timeout")
	}
	return l.down("unreachable")
}

// timedOut reports whether err says it is a timeout, as net errors do.
func timedOut(err error) bool {
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

// httpReason names an unexpected HTTP status without the body, which may be large or echo the request.
func httpReason(code int) string {
	// ADR-0031: a free NVIDIA key is rate limited, and the review must name that rather than a bare status.
	if code == http.StatusTooManyRequests {
		return "http 429, rate limited"
	}
	return "http " + strconv.Itoa(code)
}
