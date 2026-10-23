package guardrail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// liveBuilt counts live clients New has constructed in this process.
// Mode off does not construct one.
var liveBuilt atomic.Int64

// LiveBuilt is how many live guardrail clients New has constructed.
// A sample review leaves it unchanged.
func LiveBuilt() int64 { return liveBuilt.Load() }

// maxBody is the most of one guardrail completion this adapter will read.
const maxBody = 1 << 20

// Server calls our model server.
type Server struct {
	c *client
}

// NVIDIA calls NVIDIA's hosted API. The request and the answer use the same shape as Server.
type NVIDIA struct {
	c *client
}

// New builds the live adapter for cfg. Mode off returns an adapter that refuses
// if it is asked to check: a review with the guardrail off does not call it.
func New(cfg Settings, httpClient *http.Client) (Adapter, error) {
	switch cfg.Mode {
	// ADR-0005 point 3: the switch that turns the guardrail off.
	case ModeOff:
		return Off{}, nil
	// Both backends speak the same request; only the address, model and key differ.
	case ModeServer, ModeNVIDIA:
		if cfg.Key == "" || cfg.BaseURL == "" || cfg.Model == "" || cfg.Timeout <= 0 {
			return nil, Refuse("not configured")
		}
		liveBuilt.Add(1)
		c := newClient(cfg, httpClient)
		if cfg.Mode == ModeNVIDIA {
			return &NVIDIA{c: c}, nil
		}
		return &Server{c: c}, nil
	default:
		return nil, Refuse("not configured")
	}
}

// Check judges prompt, and answer when it is not empty.
func (s *Server) Check(ctx context.Context, prompt, answer string) (Verdict, error) {
	if s == nil || s.c == nil {
		return Verdict{}, Refuse("not configured")
	}
	return s.c.check(ctx, prompt, answer)
}

// Check judges prompt, and answer when it is not empty.
func (n *NVIDIA) Check(ctx context.Context, prompt, answer string) (Verdict, error) {
	if n == nil || n.c == nil {
		return Verdict{}, Refuse("not configured")
	}
	return n.c.check(ctx, prompt, answer)
}

// Identity names this backend by its mode and the model id the settings chose.
func (s *Server) Identity() (backend, model string) {
	return ModeServer, s.model()
}

// Identity names this backend by its mode and the model id the settings chose.
func (n *NVIDIA) Identity() (backend, model string) {
	if n == nil {
		return ModeNVIDIA, ""
	}
	return ModeNVIDIA, n.c.model()
}

// model is the configured model id. A nil adapter has none: its Check
// refuses as not configured, and the trace must not invent a model for it.
func (s *Server) model() string {
	if s == nil {
		return ""
	}
	return s.c.model()
}

// Off is the guardrail switched off. Check refuses, so a caller that forgets
// the switch cannot treat the prompt as safe.
type Off struct{}

// Check refuses. The review writes "no guardrail" and does not call Check.
func (Off) Check(context.Context, string, string) (Verdict, error) {
	return Verdict{}, Refuse("off")
}

// Identity names the switched-off guardrail "off". It calls no model.
func (Off) Identity() (backend, model string) { return ModeOff, "" }

// client is the HTTP side both backends share.
type client struct {
	cfg  Settings
	http *http.Client
}

// model is the model id the request sends; empty for a nil client.
func (c *client) model() string {
	if c == nil {
		return ""
	}
	return c.cfg.Model
}

// newClient uses, when httpClient is nil, a client that does not follow
// redirects, so the key is not sent to another host.
func newClient(cfg Settings, httpClient *http.Client) *client {
	if httpClient == nil {
		// ADR-0034: in the air-gapped mode this client also refuses, at dial time, any address off the network.
		httpClient = cfg.Network.NewClient(cfg.Timeout)
	}
	return &client{cfg: cfg, http: httpClient}
}

// chatRequest is the measured request of the content-safety model: no
// sampling, no stream, no thinking, and categories asked for.
type chatRequest struct {
	Model       string         `json:"model"`
	Temperature float64        `json:"temperature"`
	Stream      bool           `json:"stream"`
	Messages    []chatMessage  `json:"messages"`
	Kwargs      templateKwargs `json:"chat_template_kwargs"`
}

// chatMessage is one turn of the conversation being judged.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// templateKwargs are the model's chat-template switches.
type templateKwargs struct {
	EnableThinking    bool   `json:"enable_thinking"`
	RequestCategories string `json:"request_categories"`
}

// completion is the part of the response that carries the verdict text.
type completion struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// check posts one conversation and parses its verdict. The latency is kept
// even on failure, because ADR-0005 has the trace show what each check cost.
func (c *client) check(ctx context.Context, prompt, answer string) (Verdict, error) {
	if c == nil || c.http == nil || c.cfg.Key == "" || c.cfg.BaseURL == "" || c.cfg.Timeout <= 0 {
		return Verdict{}, Refuse("not configured")
	}
	start := time.Now()
	text, err := c.post(ctx, prompt, answer)
	lat := time.Since(start)
	if err != nil {
		return Verdict{Latency: lat}, err
	}
	v, perr := Parse(text, answer != "")
	v.Latency = lat
	// A verdict that cannot be read is refused, never treated as safe.
	if perr != nil {
		return v, Refuse("bad verdict")
	}
	return v, nil
}

// post sends one request with its own timeout from PROVEN_GUARDRAIL_TIMEOUT and
// returns the completion text.
func (c *client) post(parent context.Context, prompt, answer string) (string, error) {
	if err := parent.Err(); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(parent, c.cfg.Timeout)
	defer cancel()
	payload, err := json.Marshal(chatRequest{
		Model:       c.cfg.Model,
		Temperature: 0,
		Stream:      false,
		Messages:    messages(prompt, answer),
		Kwargs:      templateKwargs{EnableThinking: false, RequestCategories: "/categories"},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint(c.cfg.BaseURL, "/chat/completions"), bytes.NewReader(payload))
	if err != nil {
		return "", Refuse("not configured")
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	// The body is read before the status is checked, so the connection can be
	// reused; it is not put in any error, because it may echo the request.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		if parent.Err() != nil {
			return "", parent.Err()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", Refuse("timeout")
		}
		return "", Refuse("bad verdict")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return "", Refuse("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return "", Refuse(httpReason(resp.StatusCode))
	}
	var parsed completion
	if err := json.Unmarshal(body, &parsed); err != nil || len(parsed.Choices) == 0 {
		return "", Refuse("bad verdict")
	}
	text := parsed.Choices[0].Message.Content
	if strings.TrimSpace(text) == "" {
		return "", Refuse("empty answer")
	}
	return text, nil
}

// messages is the conversation to judge: the prompt as the user turn, and the
// answer as the assistant turn when there is one.
func messages(prompt, answer string) []chatMessage {
	out := []chatMessage{{Role: "user", Content: prompt}}
	if answer != "" {
		out = append(out, chatMessage{Role: "assistant", Content: answer})
	}
	return out
}

// classify names a failed request: the caller's cancellation, a timeout, or an
// unreachable guardrail.
func classify(parent, ctx context.Context, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || timedOut(err) {
		return Refuse("timeout")
	}
	return Refuse("unreachable")
}

// timedOut reports whether err says it is a timeout, as net errors do.
func timedOut(err error) bool {
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

// httpReason names an unexpected HTTP status without the body.
func httpReason(code int) string {
	return "http " + strconv.Itoa(code)
}

// endpoint joins the API root with a method path. A base whose path is /v1 is already the root.
func endpoint(base, method string) string {
	b := strings.TrimRight(base, "/")
	if strings.HasSuffix(b, "/v1") {
		return b + method
	}
	return b + "/v1" + method
}
