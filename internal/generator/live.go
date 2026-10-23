package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/service"
)

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
// unreachable endpoint, or a second empty answer is a named nightshift failure.
// A routed model outside nvidia/* is refused when this agent requires NVIDIA.
func (l *Live) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if l == nil || l.client == nil || l.cfg.Token == "" || l.cfg.BaseURL == "" || l.cfg.Timeout <= 0 {
		return "", down("not configured")
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
			return "", down("empty answer")
		}
	}
	return l.finish(agent, text, meta)
}

// Models is GET /v1/models with the same contract headers as a completion.
// A review does not call it. The completion itself is the call a review makes.
func (l *Live) Models(ctx context.Context) error {
	if l == nil || l.client == nil || l.cfg.Token == "" || l.cfg.BaseURL == "" || l.cfg.Timeout <= 0 {
		return down("not configured")
	}
	// parent is kept so a cancellation by the caller is told apart from this call's own timeout.
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, l.cfg.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint(l.cfg.BaseURL, "/models"), nil)
	if err != nil {
		return down("not configured")
	}
	l.authorize(req)
	resp, err := l.client.Do(req)
	if err != nil {
		return classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	discard(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return down("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return down(httpReason(resp.StatusCode))
	}
	return nil
}

// finish records the model that answered and applies the NVIDIA check.
// ADR-0019 point 4: the check reads the receipt, because Nightshift does not
// obey the model a request names. With no receipt the model is "unknown" and
// the answer is kept here.
func (l *Live) finish(agent, text string, meta streamMeta) (string, error) {
	use := Use{Model: "unknown", Route: meta.route, ReceiptID: meta.receipt}
	if meta.model != "" {
		use.Model = meta.model
		use.FromReceipt = true
	}
	// The use is recorded before the check, so a refused answer still names its model in the trace.
	l.last = use
	l.has = true
	if use.FromReceipt && l.cfg.requiresNVIDIA(agent) && !nvidiaModel(use.Model) {
		return "", down("answered by " + use.Model + ", refused: not an NVIDIA model")
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
		return "", streamMeta{}, down("not configured")
	}
	l.authorize(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := l.client.Do(req)
	if err != nil {
		return "", streamMeta{}, classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		discard(resp.Body)
		// A wrong token is not retried: the next call would be refused the same way.
		return "", streamMeta{}, down("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		discard(resp.Body)
		return "", streamMeta{}, down(httpReason(resp.StatusCode))
	}
	meta, err := parseStream(resp.Body)
	if err != nil {
		// A stream cut by the caller's cancellation returns that cancellation,
		// not a Nightshift failure.
		if parent.Err() != nil {
			return "", streamMeta{}, parent.Err()
		}
		var d *service.Down
		if errors.As(err, &d) {
			return "", streamMeta{}, err
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", streamMeta{}, down("timeout")
		}
		return "", streamMeta{}, down("bad stream")
	}
	return meta.text, meta, nil
}

// authorize sets the headers of Nightshift's app contract: the token, the app
// name that selects the operator's routing mode, and the user scope.
func (l *Live) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+l.cfg.Token)
	req.Header.Set("meterless-app", l.cfg.App)
	req.Header.Set("meterless-scope", "user")
}

// classify names a failed request: the caller's cancellation, a timeout, or an
// unreachable endpoint.
func classify(parent, ctx context.Context, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || timedOut(err) {
		return down("timeout")
	}
	return down("unreachable")
}

// timedOut reports whether err says it is a timeout, as net errors do.
func timedOut(err error) bool {
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

// httpReason names an unexpected HTTP status without the body, which may be large or echo the request.
func httpReason(code int) string {
	return "http " + strconv.Itoa(code)
}
