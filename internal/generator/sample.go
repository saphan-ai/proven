package generator

import (
	"context"
	"fmt"

	"github.com/saphan-ai/proven/internal/limit"
)

// Response is one recorded answer, keyed by agent and prompt.
type Response struct {
	Agent    string
	Prompt   string
	Response string
}

// Sample answers from recorded responses. It opens no network.
type Sample struct {
	answers map[string]string
}

// NewSample indexes responses by agent and prompt. A duplicate key is refused.
func NewSample(rows []Response) (*Sample, error) {
	s := &Sample{answers: make(map[string]string, len(rows))}
	for _, row := range rows {
		k := key(row.Agent, row.Prompt)
		if _, ok := s.answers[k]; ok {
			return nil, fmt.Errorf("generator: duplicate recorded response for agent %s", row.Agent)
		}
		s.answers[k] = row.Response
	}
	return s, nil
}

// Complete returns the recorded response for this agent and prompt.
func (s *Sample) Complete(ctx context.Context, agent, prompt string) (string, error) {
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	text, ok := s.answers[key(agent, prompt)]
	if !ok {
		return "", &Missing{Agent: agent, Prompt: prompt}
	}
	return text, nil
}

func key(agent, prompt string) string {
	return agent + "\x00" + prompt
}

// Missing means the fixture has no answer for this agent and prompt.
type Missing struct {
	Agent  string
	Prompt string
}

func (m *Missing) Error() string {
	return fmt.Sprintf("generator: no recorded response for agent %q", m.Agent)
}
