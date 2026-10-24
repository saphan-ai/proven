package generator

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/service"
)

// Environment variables of the second reviewer (ADR-0037 point 2). The set is closed. The key is
// the switch: with no key there is no second reviewer and the review is as it was.
const (
	EnvSecondBaseURL = "PROVEN_SECOND_REVIEWER_BASE_URL"
	EnvSecondKey     = "PROVEN_SECOND_REVIEWER_KEY"
	EnvSecondModel   = "PROVEN_SECOND_REVIEWER_MODEL"
)

// SecondService is the name a second-reviewer failure carries in the review's list of failed services.
const SecondService = "second-reviewer"

// SecondEnvironment lists the second reviewer's settings in the order the page shows.
// Each Name is the Env constant beside this function.
func SecondEnvironment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvSecondKey,
			Default:     "none (empty: no second reviewer)",
			Secret:      true,
			Description: "the API key of the second reviewer. Secret: never printed. Empty: not connected, and the review has one reviewer",
		},
		{
			Name:        EnvSecondBaseURL,
			Default:     "none (required with the key)",
			Description: "the address of any OpenAI-compatible chat API, for example `https://api.openai.com/v1` or a local server",
		},
		{
			Name:        EnvSecondModel,
			Default:     "none (required with the key)",
			Description: "the model asked for; the review names it as the second reviewer",
		},
	}
}

// SecondSettings is one validated second-reviewer configuration. The key is held so the call can
// send it. It is not a value to print.
type SecondSettings struct {
	BaseURL string
	Key     string
	Model   string
	Timeout time.Duration
}

// LoadSecond reads the second reviewer's settings. The bool is false when no key is set: the
// reviewer is not connected, which is the default and not a refusal. A wrong value is refused even
// then, so a typo in the address is found before the key is added. A key without an address or
// without a model is refused, because there is nothing to call. The timeout is the generator's own
// PROVEN_GENERATOR_TIMEOUT, so one setting bounds every model call.
func LoadSecond(env func(string) string) (SecondSettings, bool, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	key := env(EnvSecondKey)
	if strings.ContainsAny(key, "\r\n") || len(key) > 4096 || service.TooShortSecret(key) {
		return SecondSettings{}, false, refuse(EnvSecondKey, "", "not a usable key", "a non-empty secret in the environment, at least 8 and at most 4096 characters, with no line break")
	}
	connected := strings.TrimSpace(key) != ""
	base := env(EnvSecondBaseURL)
	if base == "" {
		if connected {
			return SecondSettings{}, false, refuse(EnvSecondBaseURL, "", "a second reviewer needs an address", "an absolute http or https URL")
		}
	} else {
		var err error
		if base, err = checkBaseFor(EnvSecondBaseURL, base); err != nil {
			return SecondSettings{}, false, err
		}
	}
	model := env(EnvSecondModel)
	if model == "" {
		if connected {
			return SecondSettings{}, false, refuse(EnvSecondModel, "", "a second reviewer needs a model", "a non-empty model name without spaces, at most 128 characters")
		}
	} else if err := checkToken(EnvSecondModel, model, 128, "a non-empty model name without spaces, at most 128 characters", modelName); err != nil {
		return SecondSettings{}, false, err
	}
	timeout := defaultTimeout
	if raw := env(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return SecondSettings{}, false, refuse(EnvTimeout, raw, "not a Go duration greater than zero", "a Go duration greater than zero, for example 60s")
		}
		timeout = d
	}
	if !connected {
		return SecondSettings{}, false, nil
	}
	return SecondSettings{BaseURL: base, Key: key, Model: model, Timeout: timeout}, true, nil
}

// Second is the generator of the second reviewer: any OpenAI-compatible chat API, behind the same
// Generator interface as Nightshift. It sends only the key and the prompt, asks for the model it
// was configured with, and names that model as the reviewer, because no receipt of this API is read.
type Second struct {
	live  *Live
	model string
}

// Complete sends one prompt to the second reviewer. A failure is renamed to the second reviewer,
// so the review never blames Nightshift for another server's answer.
func (s *Second) Complete(ctx context.Context, agent, prompt string) (string, error) {
	text, err := s.live.Complete(ctx, agent, prompt)
	var d *service.Down
	if errors.As(err, &d) {
		return "", &service.Down{Service: SecondService, Reason: d.Reason}
	}
	return text, err
}

// LastUse names the configured model for the call that just finished. FromReceipt is false: this
// name is the setting, not a receipt's word.
func (s *Second) LastUse() (Use, bool) {
	use, ok := s.live.LastUse()
	if !ok {
		return Use{}, false
	}
	use.Model = s.model
	use.FromReceipt = false
	return use, true
}
