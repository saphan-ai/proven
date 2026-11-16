package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/modelcall"
	"github.com/saphan-ai/proven/internal/redact"
)

// runModelCheck measures exactly one completion request, without recovery or
// HTTP retries. A successful microprobe establishes compatibility, not quality.
func runModelCheck(args []string, env func(string) string, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "check" {
		fmt.Fprintln(stderr, "usage: proven model check (uses PROVEN_GENERATOR_*; one request, no retries)")
		return exitFatal
	}
	r := redact.FromEnv(env)
	cfg, err := generator.Load(env)
	if err != nil {
		fmt.Fprintln(stderr, visible(r.Text(err.Error())))
		return exitConfig
	}
	if cfg.Mode != generator.ModeOpenAI && cfg.Mode != generator.ModeNVIDIA && cfg.Mode != generator.ModeNightshift {
		fmt.Fprintln(stderr, "model check requires a configured HTTP model transport")
		return exitConfig
	}
	options := *cfg.Options
	options.RecoveryRetries, options.RateLimitRetries, options.TransportRetries = 0, 0, 0
	cfg.Options = &options
	g := generator.NewLive(cfg, nil)
	text, err := g.Complete(context.Background(), "investigator", `Reply with JSON only: {"ok":true}`)
	var decoded struct {
		OK bool `json:"ok"`
	}
	valid := answer.Decode(text, &decoded) == nil && decoded.OK
	result := struct {
		OK        bool                `json:"ok"`
		Nonempty  bool                `json:"nonempty"`
		ValidJSON bool                `json:"valid_json"`
		Error     string              `json:"error,omitempty"`
		Attempts  []modelcall.Attempt `json:"attempts"`
	}{OK: err == nil && valid, Nonempty: text != "", ValidJSON: valid, Attempts: g.LastAttempts()}
	if err != nil {
		result.Error = r.Text(err.Error())
	} else if !valid {
		result.Error = "answer did not match the probe schema"
	}
	b, err := json.Marshal(result)
	if err != nil {
		return exitFatal
	}
	if _, err := fmt.Fprintln(stdout, r.Text(string(b))); err != nil {
		return exitFatal
	}
	if !result.OK {
		return exitIncomplete
	}
	return exitOK
}
