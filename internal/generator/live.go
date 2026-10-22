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

type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

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
	l.has = false
	l.last = Use{}
	text, meta, err := l.attempt(ctx, prompt)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" {
		text, meta, err = l.attempt(ctx, prompt)
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

func (l *Live) finish(agent, text string, meta streamMeta) (string, error) {
	use := Use{Model: "unknown", Route: meta.route, ReceiptID: meta.receipt}
	if meta.model != "" {
		use.Model = meta.model
		use.FromReceipt = true
	}
	l.last = use
	l.has = true
	if use.FromReceipt && l.cfg.requiresNVIDIA(agent) && !nvidiaModel(use.Model) {
		return "", down("answered by " + use.Model + ", refused: not an NVIDIA model")
	}
	return text, nil
}

func nvidiaModel(model string) bool {
	rest, ok := strings.CutPrefix(model, "nvidia/")
	return ok && rest != ""
}

func (l *Live) attempt(parent context.Context, prompt string) (string, streamMeta, error) {
	if err := parent.Err(); err != nil {
		return "", streamMeta{}, err
	}
	ctx, cancel := context.WithTimeout(parent, l.cfg.Timeout)
	defer cancel()
	payload, err := json.Marshal(chatRequest{
		Model:    l.cfg.Model,
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
		return "", streamMeta{}, down("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		discard(resp.Body)
		return "", streamMeta{}, down(httpReason(resp.StatusCode))
	}
	meta, err := parseStream(resp.Body)
	if err != nil {
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

func (l *Live) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+l.cfg.Token)
	req.Header.Set("meterless-app", l.cfg.App)
	req.Header.Set("meterless-scope", "user")
}

func classify(parent, ctx context.Context, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || timedOut(err) {
		return down("timeout")
	}
	return down("unreachable")
}

func timedOut(err error) bool {
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

func httpReason(code int) string {
	return "http " + strconv.Itoa(code)
}
