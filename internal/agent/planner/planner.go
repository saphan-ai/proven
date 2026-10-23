// Package planner reads a change and names the risks.
//
// The Planner is the first agent of the review. It reads each changed file's
// base and change, asks the model for candidate risks, and hands them on to
// the Investigator and the Prover. A risk it names is not a finding: only
// evidence makes one (ADR-0002).
package planner

import (
	"context"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/generator"
)

// riskSpec is the JSON the planner asks a model to return.
var riskSpec = answer.Spec{
	Task:    "Name the risks in this change.",
	Shape:   "a JSON array of objects. Each object has id (string), rule_id (string), title (string), message (string), severity (string), path (string), start_line (number) and end_line (number).",
	Example: `[{"id":"F1","rule_id":"example-rule","title":"Example risk","message":"What is wrong.","severity":"ERROR","path":"file.go","start_line":1,"end_line":1}]`,
}

// Prompt is the planner's prompt for these files. The sample generator keys on it.
func Prompt(files []change.File) string {
	return answer.Instruct(riskSpec, filesText(files))
}

// filesText lays out each file as its path, its base and its change. Each part
// ends with a line break, so the next marker always starts its own line.
func filesText(files []change.File) string {
	var b strings.Builder
	for _, f := range files {
		b.WriteString("FILE ")
		b.WriteString(f.Path)
		b.WriteString("\nBASE\n")
		b.Write(f.Base)
		if len(f.Base) == 0 || f.Base[len(f.Base)-1] != '\n' {
			b.WriteByte('\n')
		}
		b.WriteString("CHANGE\n")
		b.Write(f.Change)
		if len(f.Change) == 0 || f.Change[len(f.Change)-1] != '\n' {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// row is one risk as the model returns it.
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
	// An empty change has nothing to review; it is refused rather than reported clean.
	if len(files) == 0 {
		return nil, fmt.Errorf("planner: the change has no files")
	}
	var rows []row
	if err := answer.Read(ctx, gen, "planner", Prompt(files), riskSpec, &rows); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]candidate.Risk, 0, len(rows))
	for _, r := range rows {
		// Later steps and the trace key every finding on its id, so a missing or
		// repeated id is refused rather than merged.
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
