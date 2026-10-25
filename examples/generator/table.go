// Package generator is a worked example of a model-provider adapter.
//
// Table answers Complete from a fixed table keyed by agent name. It
// implements internal/generator.Generator with no network, no secret and no
// environment variable. A reader copies this package and replaces the table
// with their own model. What to add is in docs/ADAPTERS.md under
// "Adding a model provider".
package generator

import (
	"context"
	"errors"

	"github.com/saphan-ai/proven/internal/generator"
)

// The assignment fails the build when internal/generator.Generator changes,
// so a copy of this example stays bound to the real interface the agents call.
var _ generator.Generator = Table{}

// ErrUnknownAgent is the refusal: the table has no answer for this agent.
var ErrUnknownAgent = errors.New("generator: no recorded answer for this agent")

// Row is one agent and the answer the table gives it.
type Row struct {
	Agent  string
	Answer string
}

// Table answers one prompt for one agent from a fixed table.
// The prompt chooses nothing: one agent has one answer.
type Table struct {
	answers map[string]string
}

// NewTable builds a table from rows. A duplicate agent is refused.
func NewTable(rows []Row) (*Table, error) {
	answers := make(map[string]string, len(rows))
	for _, row := range rows {
		if _, ok := answers[row.Agent]; ok {
			return nil, errors.New("generator: two answers for agent " + row.Agent)
		}
		answers[row.Agent] = row.Answer
	}
	return &Table{answers: answers}, nil
}

// Complete returns the table's answer for the agent. An unknown agent
// returns ErrUnknownAgent and no answer. The context cancels the call; a
// table lookup cannot be slow, so it never blocks.
func (t Table) Complete(_ context.Context, agent, _ string) (string, error) {
	answer, ok := t.answers[agent]
	if !ok {
		return "", ErrUnknownAgent
	}
	return answer, nil
}
