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
	"time"
)

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
	case ModeOff:
		return Off{}, nil
	case ModeServer, ModeNVIDIA:
		if cfg.Key == "" || cfg.BaseURL == "" || cfg.Model == "" || cfg.Timeout <= 0 {
			return nil, Refuse("not configured")
		}
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

// Off is the guardrail switched off. Check refuses, so a caller that forgets
// the switch cannot treat the prompt as safe.
type Off struct{}

// Check refuses. The review writes "no guardrail" and does not call Check.
func (Off) Check(context.Context, string, string) (Verdict, error) {
	return Verdict{}, Refuse("off")
}

type client struct {
	cfg  Settings
	http *http.Client
}

func newClient(cfg Settings, httpClient *http.Client) *client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return &client{cfg: cfg, http: httpClient}
}

type chatRequest struct {
	Model       string         `json:"model"`
	Temperature float64        `json:"temperature"`
	Stream      bool           `json:"stream"`
	Messages    []chatMessage  `json:"messages"`
	Kwargs      templateKwargs `json:"chat_template_kwargs"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type templateKwargs struct {
	EnableThinking    bool   `json:"enable_thinking"`
	RequestCategories string `json:"request_categories"`
}

type completion struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

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
	if perr != nil {
		return v, Refuse("bad verdict")
	}
	return v, nil
}

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

func messages(prompt, answer string) []chatMessage {
	out := []chatMessage{{Role: "user", Content: prompt}}
	if answer != "" {
		out = append(out, chatMessage{Role: "assistant", Content: answer})
	}
	return out
}

func classify(parent, ctx context.Context, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || timedOut(err) {
		return Refuse("timeout")
	}
	return Refuse("unreachable")
}

func timedOut(err error) bool {
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

func httpReason(code int) string {
	return "http " + strconv.Itoa(code)
}

func endpoint(base, method string) string {
	b := strings.TrimRight(base, "/")
	if strings.HasSuffix(b, "/v1") {
		return b + method
	}
	return b + "/v1" + method
}
