package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/saphan-ai/proven/internal/modelcall"
	"github.com/saphan-ai/proven/internal/service"
)

var errHeaderTimeout = errors.New("response headers timeout")
var errIdleTimeout = errors.New("stream idle timeout")

func (l *Live) requestBody(agent, prompt, model string) ([]byte, error) {
	o := l.options()
	base, err := json.Marshal(chatRequest{Model: model, Stream: true, Messages: []chatMessage{{Role: "user", Content: prompt}}, ThinkingBudget: l.cfg.ThinkingBudget})
	if err != nil {
		return nil, err
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(base, &body); err != nil {
		return nil, err
	}
	put := func(k string, v any) { body[k], _ = json.Marshal(v) }
	n := o.MaxTokens
	if v, ok := o.AgentMaxTokens[agent]; ok {
		n = v
	}
	if n > 0 {
		field := o.MaxTokensField
		if field == "" {
			field = "max_tokens"
		}
		put(field, n)
	}
	if o.IncludeUsage {
		put("stream_options", map[string]bool{"include_usage": true})
	}
	for k, v := range o.ExtraBody {
		body[k] = v
	}
	for k, v := range o.AgentExtraBody[agent] {
		body[k] = v
	}
	return json.Marshal(body)
}

// attempt's total clock includes pre-stream retries and backoff. Header and
// inactivity clocks can only shorten it; retries never restart its deadline.
func (l *Live) attempt(parent context.Context, agent, prompt, model string) (string, streamMeta, error) {
	if parent.Err() != nil {
		return "", streamMeta{}, parent.Err()
	}
	if !timeForAttempt(parent, l.cfg.Timeout, false) {
		return "", streamMeta{}, ErrReviewBudget
	}
	deadline := time.Now().Add(l.cfg.Timeout)
	cause := error(context.DeadlineExceeded)
	if d, ok := parent.Deadline(); ok && d.Add(-reserveFor(parent)).Before(deadline) {
		deadline = d.Add(-reserveFor(parent))
		if reserveFor(parent) > 0 {
			cause = ErrReviewBudget
		}
	}
	total, cancel := context.WithDeadlineCause(parent, deadline, cause)
	defer cancel()
	ctx, stop := context.WithCancelCause(total)
	defer stop(nil)
	payload, err := l.requestBody(agent, prompt, model)
	if err != nil {
		return "", streamMeta{}, l.down("invalid request options")
	}
	o := l.options()
	rateRetries, transportRetries := 0, 0
	for {
		meta, status, retryAfter, err := l.send(parent, ctx, stop, payload, model)
		if err == nil {
			switch meta.finishReason {
			case "length", "max_tokens":
				if meta.text == "" && meta.reasoningBytes > 0 {
					return "", meta, l.down("response token limit reached before answer (reasoning)")
				}
				return "", meta, l.down("response token limit reached")
			case "content_filter", "tool_calls", "function_call":
				return "", meta, l.down("completion did not produce a text answer")
			}
			return meta.text, meta, nil
		}
		if ctx.Err() != nil || parent.Err() != nil {
			return "", meta, l.classify(parent, ctx, err)
		}
		var down *service.Down
		if errors.As(err, &down) && (down.Reason == "timeout" || down.Reason == errHeaderTimeout.Error() || down.Reason == errIdleTimeout.Error()) {
			return "", meta, err
		}
		var delay time.Duration
		switch {
		case status == http.StatusTooManyRequests && rateRetries < o.RateLimitRetries:
			rateRetries++
			var allowed bool
			delay, allowed = rateDelay(retryAfter, rateRetries, o)
			if !allowed {
				return "", meta, err
			}
		case (status == 0 || status == 500 || status == 502 || status == 503) && transportRetries < o.TransportRetries:
			transportRetries++
			delay = cappedBackoff(o.RetryBackoff, o.RetryMaxBackoff, transportRetries)
		default:
			return "", meta, err
		}
		if time.Until(deadline) <= delay {
			return "", meta, err
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", meta, l.classify(parent, ctx, ctx.Err())
		case <-timer.C:
		}
	}
}

func rateDelay(header string, retry int, o CallOptions) (time.Duration, bool) {
	var d time.Duration
	if seconds, err := strconv.ParseInt(header, 10, 32); err == nil && seconds >= 0 {
		d = time.Duration(seconds) * time.Second
	} else if when, err := http.ParseTime(header); err == nil {
		d = max(time.Until(when), 0)
	} else {
		d = cappedBackoff(o.RetryBackoff, o.RetryMaxBackoff, 1<<uint(retry-1))
	}
	// Never retry sooner than a valid server hint to fit our own cap.
	return d, d <= o.RetryMaxBackoff
}

func cappedBackoff(initial, cap time.Duration, multiplier int) time.Duration {
	if initial > cap/time.Duration(multiplier) {
		return cap
	}
	return initial * time.Duration(multiplier)
}

func (l *Live) send(parent, ctx context.Context, cancel context.CancelCauseFunc, payload []byte, model string) (meta streamMeta, status int, retryAfter string, resultErr error) {
	o := l.options()
	start := time.Now()
	stat := modelcall.Attempt{StartedAt: start.UTC(), Requested: model, Route: l.cfg.BaseHost()}
	defer func() {
		stat.DurationMS = time.Since(start).Milliseconds()
		stat.HTTPStatus, stat.Model, stat.ResponseID = status, meta.chunkModel, meta.responseID
		stat.FinishReason, stat.AnswerBytes, stat.ReasoningBytes = meta.finishReason, len(meta.text), meta.reasoningBytes
		stat.Reasoning, stat.ReasoningClipped, stat.Usage = meta.reasoning, meta.reasoningClipped, meta.usage
		if resultErr != nil {
			if d, ok := service.As(resultErr); ok {
				stat.Error = d.Reason
			} else if errors.Is(resultErr, ErrReviewBudget) {
				stat.Error = "review budget exhausted"
			} else if errors.Is(resultErr, context.Canceled) {
				stat.Error = "canceled"
			} else {
				stat.Error = "deadline exceeded"
			}
		} else if meta.finishReason == "length" || meta.finishReason == "max_tokens" {
			stat.Error = "response token limit reached"
		} else if meta.finishReason == "content_filter" || meta.finishReason == "tool_calls" || meta.finishReason == "function_call" {
			stat.Error = "completion did not produce a text answer"
		} else if strings.TrimSpace(meta.text) == "" {
			stat.Error = "empty answer"
		}
		l.attempts = append(l.attempts, stat)
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint(l.cfg.BaseURL, "/chat/completions"), bytes.NewReader(payload))
	if err != nil {
		return meta, 0, "", l.down("not configured")
	}
	l.authorize(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	var headers atomic.Bool
	var timer *time.Timer
	if o.HeaderTimeout > 0 {
		timer = time.AfterFunc(o.HeaderTimeout, func() {
			if !headers.Load() {
				cancel(errHeaderTimeout)
			}
		})
	}
	resp, err := l.client.Do(req)
	headers.Store(true)
	if timer != nil {
		timer.Stop()
	}
	if err != nil {
		return meta, 0, "", l.classify(parent, ctx, err)
	}
	hms := time.Since(start).Milliseconds()
	stat.HeadersMS = &hms
	status, retryAfter = resp.StatusCode, resp.Header.Get("Retry-After")
	body := newProgressBody(resp.Body, o.IdleTimeout, func() { cancel(errIdleTimeout) }, func() {
		if stat.FirstByteMS == nil {
			ms := time.Since(start).Milliseconds()
			stat.FirstByteMS = &ms
		}
	})
	defer body.Close()
	if l.rejectsKey(status) {
		discard(body)
		return meta, status, retryAfter, l.down("unauthorized")
	}
	if status != http.StatusOK {
		discard(body)
		return meta, status, retryAfter, l.down(httpReason(status))
	}
	meta, err = parseStreamOptions(body, o.TraceReasoningBytes)
	if err != nil {
		if ctx.Err() != nil || parent.Err() != nil || timedOut(err) {
			return meta, status, "", l.classify(parent, ctx, err)
		}
		if errors.Is(err, errProviderStream) {
			return meta, status, "", l.down("provider error in stream")
		}
		return meta, status, "", l.down("bad stream")
	}
	if ctx.Err() != nil {
		return meta, status, "", l.classify(parent, ctx, ctx.Err())
	}
	return meta, status, "", nil
}

// progressBody's timer observes wire activity, not just answer tokens.
type progressBody struct {
	io.ReadCloser
	mu     sync.Mutex
	timer  *time.Timer
	idle   time.Duration
	closed bool
	first  func()
	once   sync.Once
}

func newProgressBody(body io.ReadCloser, idle time.Duration, expire, first func()) *progressBody {
	b := &progressBody{ReadCloser: body, idle: idle, first: first}
	if idle > 0 {
		b.timer = time.AfterFunc(idle, func() {
			b.mu.Lock()
			if b.closed {
				b.mu.Unlock()
				return
			}
			b.closed = true
			b.mu.Unlock()
			expire()
			body.Close()
		})
	}
	return b
}

func (b *progressBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.once.Do(b.first)
		b.mu.Lock()
		if !b.closed && b.timer != nil {
			b.timer.Reset(b.idle)
		}
		b.mu.Unlock()
	}
	return n, err
}

func (b *progressBody) Close() error {
	b.mu.Lock()
	b.closed = true
	if b.timer != nil {
		b.timer.Stop()
	}
	b.mu.Unlock()
	return b.ReadCloser.Close()
}
