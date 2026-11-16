package generator

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
)

const (
	EnvMaxTokens        = "PROVEN_GENERATOR_MAX_TOKENS"
	EnvMaxTokensField   = "PROVEN_GENERATOR_MAX_TOKENS_FIELD"
	EnvAgentMaxTokens   = "PROVEN_GENERATOR_AGENT_MAX_TOKENS"
	EnvHeaderTimeout    = "PROVEN_GENERATOR_HEADER_TIMEOUT"
	EnvIdleTimeout      = "PROVEN_GENERATOR_IDLE_TIMEOUT"
	EnvIncludeUsage     = "PROVEN_GENERATOR_INCLUDE_USAGE"
	EnvExtraBody        = "PROVEN_GENERATOR_EXTRA_BODY"
	EnvAgentExtraBody   = "PROVEN_GENERATOR_AGENT_EXTRA_BODY"
	EnvRecoveryRetries  = "PROVEN_GENERATOR_RECOVERY_RETRIES"
	EnvRepairContext    = "PROVEN_GENERATOR_REPAIR_CONTEXT"
	EnvRepairMaxBytes   = "PROVEN_GENERATOR_REPAIR_MAX_BYTES"
	EnvDeadlineReserve  = "PROVEN_GENERATOR_DEADLINE_RESERVE"
	EnvRateLimitRetries = "PROVEN_GENERATOR_RATE_LIMIT_RETRIES"
	EnvTransportRetries = "PROVEN_GENERATOR_TRANSPORT_RETRIES"
	EnvRetryBackoff     = "PROVEN_GENERATOR_RETRY_BACKOFF"
	EnvRetryMaxBackoff  = "PROVEN_GENERATOR_RETRY_MAX_BACKOFF"
	EnvTraceReasonBytes = "PROVEN_GENERATOR_TRACE_REASONING_BYTES"
)

// CallOptions are provider-neutral controls. Optional request fields are omitted
// unless configured, so an endpoint never has to support another vendor's options.
type CallOptions struct {
	MaxTokens, RecoveryRetries, RepairMaxBytes              int
	RateLimitRetries, TransportRetries, TraceReasoningBytes int
	MaxTokensField                                          string
	HeaderTimeout, IdleTimeout, DeadlineReserve             time.Duration
	RetryBackoff, RetryMaxBackoff                           time.Duration
	IncludeUsage, RepairContext                             bool
	AgentMaxTokens                                          map[string]int
	ExtraBody                                               map[string]json.RawMessage
	AgentExtraBody                                          map[string]map[string]json.RawMessage
}

func DefaultCallOptions() CallOptions {
	return CallOptions{RecoveryRetries: 1, RepairMaxBytes: 64 << 10, MaxTokensField: "max_tokens",
		RetryBackoff: 250 * time.Millisecond, RetryMaxBackoff: 5 * time.Second}
}

func loadCallOptions(env func(string) string) (*CallOptions, error) {
	o := DefaultCallOptions()
	bad := func(name, allowed string) error { return refuse(name, "", "invalid model call option", allowed) }
	for _, f := range []struct {
		name string
		dst  *int
		max  int
	}{
		{EnvMaxTokens, &o.MaxTokens, 1 << 20}, {EnvRecoveryRetries, &o.RecoveryRetries, 8},
		{EnvRepairMaxBytes, &o.RepairMaxBytes, maxStream}, {EnvRateLimitRetries, &o.RateLimitRetries, 10},
		{EnvTransportRetries, &o.TransportRetries, 10}, {EnvTraceReasonBytes, &o.TraceReasoningBytes, maxStream},
	} {
		if raw := env(f.name); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 || v > f.max {
				return nil, bad(f.name, "a non-negative integer within the documented limit")
			}
			*f.dst = v
		}
	}
	if o.RepairMaxBytes == 0 {
		return nil, bad(EnvRepairMaxBytes, "a positive byte limit")
	}
	for _, f := range []struct {
		name string
		dst  *time.Duration
	}{
		{EnvHeaderTimeout, &o.HeaderTimeout}, {EnvIdleTimeout, &o.IdleTimeout}, {EnvDeadlineReserve, &o.DeadlineReserve},
		{EnvRetryBackoff, &o.RetryBackoff}, {EnvRetryMaxBackoff, &o.RetryMaxBackoff},
	} {
		if raw := env(f.name); raw != "" {
			d, err := time.ParseDuration(raw)
			if err != nil || d < 0 {
				return nil, bad(f.name, "a non-negative Go duration; zero disables this control")
			}
			*f.dst = d
		}
	}
	if o.RetryBackoff > o.RetryMaxBackoff {
		return nil, bad(EnvRetryBackoff, "no greater than the maximum retry backoff")
	}
	for _, f := range []struct {
		name string
		dst  *bool
	}{{EnvIncludeUsage, &o.IncludeUsage}, {EnvRepairContext, &o.RepairContext}} {
		switch env(f.name) {
		case "":
		case "true":
			*f.dst = true
		case "false":
			*f.dst = false
		default:
			return nil, bad(f.name, "true or false")
		}
	}
	if raw := env(EnvMaxTokensField); raw != "" {
		o.MaxTokensField = raw
	}
	if o.MaxTokensField != "max_tokens" && o.MaxTokensField != "max_completion_tokens" {
		return nil, bad(EnvMaxTokensField, "max_tokens or max_completion_tokens")
	}
	if raw := env(EnvAgentMaxTokens); raw != "" {
		if len(raw) > 64<<10 || json.Unmarshal([]byte(raw), &o.AgentMaxTokens) != nil || o.AgentMaxTokens == nil {
			return nil, bad(EnvAgentMaxTokens, "a JSON object mapping agent names to token limits")
		}
		for a, n := range o.AgentMaxTokens {
			if !knownAgent(a) || n < 0 || n > 1<<20 {
				return nil, bad(EnvAgentMaxTokens, "known agent names and token limits from 0 to 1048576")
			}
		}
	}
	var err error
	if raw := env(EnvExtraBody); raw != "" {
		if o.ExtraBody, err = extraObject(raw); err != nil {
			return nil, bad(EnvExtraBody, err.Error())
		}
	}
	if raw := env(EnvAgentExtraBody); raw != "" {
		var objects map[string]json.RawMessage
		if len(raw) > 64<<10 || json.Unmarshal([]byte(raw), &objects) != nil || objects == nil {
			return nil, bad(EnvAgentExtraBody, "a JSON object mapping agent names to extra-body objects")
		}
		o.AgentExtraBody = make(map[string]map[string]json.RawMessage)
		for a, raw := range objects {
			if !knownAgent(a) {
				return nil, bad(EnvAgentExtraBody, "known agent names")
			}
			if o.AgentExtraBody[a], err = extraObject(string(raw)); err != nil {
				return nil, bad(EnvAgentExtraBody, err.Error())
			}
		}
	}
	return &o, nil
}

func knownAgent(name string) bool {
	for _, a := range agents {
		if a == name {
			return true
		}
	}
	return false
}

func extraObject(raw string) (map[string]json.RawMessage, error) {
	var obj map[string]json.RawMessage
	if len(raw) > 64<<10 || json.Unmarshal([]byte(raw), &obj) != nil || obj == nil {
		return nil, refuse(EnvExtraBody, "", "not an object", "a JSON object no larger than 64 KiB")
	}
	for k := range obj {
		switch strings.ToLower(k) {
		case "model", "messages", "stream", "stream_options", "max_tokens", "max_completion_tokens", "thinking_budget", "n":
			return nil, refuse(EnvExtraBody, "", "client-owned field", "extra fields that do not override model, messages, streaming or token controls")
		}
	}
	return obj, nil
}

func CallEnvironment() []envdoc.Setting {
	return []envdoc.Setting{
		{Name: EnvMaxTokens, Default: "0 (omitted)", Description: "response token limit, 0 to 1048576; includes reasoning when the provider counts it there"},
		{Name: EnvMaxTokensField, Default: "max_tokens", Description: "token-limit request field: max_tokens or max_completion_tokens"},
		{Name: EnvAgentMaxTokens, Default: "none", Description: "JSON object of per-agent token limits; zero omits the field for that agent"},
		{Name: EnvHeaderTimeout, Default: "0s (off)", Description: "time allowed to receive response headers, within the total attempt timeout"},
		{Name: EnvIdleTimeout, Default: "0s (off)", Description: "stream inactivity timeout, reset by every received byte, including reasoning and heartbeats"},
		{Name: EnvIncludeUsage, Default: "false", Description: "send stream_options.include_usage; enable only for endpoints that support it"},
		{Name: EnvExtraBody, Default: "none", Description: "provider-specific request fields as a JSON object up to 64 KiB; client-owned fields are refused; never printed"},
		{Name: EnvAgentExtraBody, Default: "none", Description: "JSON objects of extra-body fields per agent; override matching top-level fields of the common extra body"},
		{Name: EnvRecoveryRetries, Default: "1", Description: "shared additional completion budget per agent step, for empty answers and JSON repair; 0 to 8"},
		{Name: EnvRepairContext, Default: "false", Description: "include the original prompt in JSON repair; otherwise only the schema and previous answer are sent"},
		{Name: EnvRepairMaxBytes, Default: "65536", Description: "previous-answer byte limit in a repair prompt, 1 to 8388608"},
		{Name: EnvDeadlineReserve, Default: "0s (off)", Description: "review finalization reserve, capped at one tenth of the original remaining review time; new attempts need a full attempt window before this reserve"},
		{Name: EnvRateLimitRetries, Default: "0", Description: "pre-stream HTTP 429 retries, 0 to 10; Retry-After is honored or the retry is skipped if it exceeds the backoff cap"},
		{Name: EnvTransportRetries, Default: "0", Description: "pre-stream connection and HTTP 500/502/503 retries, 0 to 10; timeouts, cancellations, redirects and other 4xx are never retried"},
		{Name: EnvRetryBackoff, Default: "250ms", Description: "initial backoff; linear for transport errors and exponential for rate limiting"},
		{Name: EnvRetryMaxBackoff, Default: "5s", Description: "maximum backoff; retries remain within the current attempt deadline"},
		{Name: EnvTraceReasonBytes, Default: "0", Description: "maximum reasoning text bytes retained per attempt; 0 records counts only, up to 8388608"},
	}
}
