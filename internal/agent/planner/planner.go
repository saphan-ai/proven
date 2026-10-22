// Package planner reads a change and names the risks.
package planner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/generator"
)

// Prompt is the planner's prompt for these files. The sample generator keys on it.
func Prompt(files []change.File) string {
	b := "name the risks\n"
	for _, f := range files {
		b += "FILE " + f.Path + "\nBASE\n"
		b += string(f.Base)
		if len(f.Base) == 0 || f.Base[len(f.Base)-1] != '\n' {
			b += "\n"
		}
		b += "CHANGE\n"
		b += string(f.Change)
		if len(f.Change) == 0 || f.Change[len(f.Change)-1] != '\n' {
			b += "\n"
		}
	}
	return b
}

type row struct {
	ID        string `json:"id"`
	RuleID    string `json:"rule_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// Plan asks the generator for the risks in root's base and change trees.
func Plan(ctx context.Context, root string, gen generator.Generator) ([]candidate.Risk, error) {
	files, err := change.Read(root)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("planner: the change has no files")
	}
	raw, err := gen.Complete(ctx, "planner", Prompt(files))
	if err != nil {
		return nil, err
	}
	var rows []row
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, fmt.Errorf("planner: response is not a risk list: %w", err)
	}
	seen := map[string]bool{}
	out := make([]candidate.Risk, 0, len(rows))
	for _, r := range rows {
		if r.ID == "" {
			return nil, fmt.Errorf("planner: a risk has no id")
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("planner: duplicate id %s", r.ID)
		}
		seen[r.ID] = true
		out = append(out, candidate.Risk{
			ID: r.ID, RuleID: r.RuleID, Title: r.Title, Message: r.Message,
			Severity: r.Severity, Path: r.Path, StartLine: r.StartLine, EndLine: r.EndLine,
		})
	}
	return out, nil
}
