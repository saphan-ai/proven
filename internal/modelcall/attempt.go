// Package modelcall holds transport observations without model clients or I/O.
package modelcall

import "time"

// Usage contains only counters the provider reported; nil means unknown.
type Usage struct {
	PromptTokens     *int64 `json:"prompt_tokens,omitempty"`
	CompletionTokens *int64 `json:"completion_tokens,omitempty"`
	TotalTokens      *int64 `json:"total_tokens,omitempty"`
	ReasoningTokens  *int64 `json:"reasoning_tokens,omitempty"`
}

// Attempt describes one HTTP request, including requests that failed or were retried.
type Attempt struct {
	StartedAt        time.Time `json:"started_at"`
	DurationMS       int64     `json:"duration_ms"`
	HeadersMS        *int64    `json:"headers_ms,omitempty"`
	FirstByteMS      *int64    `json:"first_byte_ms,omitempty"`
	HTTPStatus       int       `json:"http_status,omitempty"`
	Requested        string    `json:"requested,omitempty"`
	Model            string    `json:"model,omitempty"`
	Route            string    `json:"route,omitempty"`
	ResponseID       string    `json:"response_id,omitempty"`
	FinishReason     string    `json:"finish_reason,omitempty"`
	AnswerBytes      int       `json:"answer_bytes"`
	ReasoningBytes   int       `json:"reasoning_bytes"`
	Reasoning        string    `json:"reasoning,omitempty"`
	ReasoningClipped bool      `json:"reasoning_clipped,omitempty"`
	Usage            *Usage    `json:"usage,omitempty"`
	Error            string    `json:"error,omitempty"`
}
