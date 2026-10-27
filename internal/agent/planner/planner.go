// Package planner reads a change and names the risks.
//
// The Planner is the first agent of the review. It reads each changed file's
// base and change, asks the model for candidate risks, and hands them on to
// the Investigator and the Prover. A risk it names is not a finding: only
// evidence makes one (ADR-0002). A change above the bounded-steps threshold is
// read one file or hunk per step instead, with only a small carried state
// between steps (PlanBounded, ADR-0048).
package planner

import (
	"context"
	"encoding/base64"
	"fmt"
	"unicode/utf8"

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

// fileRecord is one changed file as the prompt carries it: its path, its base
// and its change, each a JSON string. A part that is not valid UTF-8 is carried
// as base64 under its own key, so two different parts never give one text.
type fileRecord struct {
	Path         *string `json:"path,omitempty"`
	PathBase64   string  `json:"path_base64,omitempty"`
	Base         *string `json:"base,omitempty"`
	BaseBase64   string  `json:"base_base64,omitempty"`
	Change       *string `json:"change,omitempty"`
	ChangeBase64 string  `json:"change_base64,omitempty"`
}

// part returns b as a JSON string when it is text, and as base64 when it is not.
func part(b []byte) (text *string, b64 string) {
	if utf8.Valid(b) {
		s := string(b)
		return &s, ""
	}
	return nil, base64.StdEncoding.EncodeToString(b)
}

// filesText writes the files as one JSON array, one record per file. The
// change's bytes are data: every line break and quote inside them is escaped,
// so no byte of a file can create a second file record or a new field.
func filesText(files []change.File) string {
	records := make([]fileRecord, 0, len(files))
	for _, f := range files {
		var r fileRecord
		r.Path, r.PathBase64 = part([]byte(f.Path))
		r.Base, r.BaseBase64 = part(f.Base)
		r.Change, r.ChangeBase64 = part(f.Change)
		records = append(records, r)
	}
	return answer.Encode(records)
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
	files, err := change.PolicyFrom(ctx).Read(root)
	if err != nil {
		return nil, err
	}
	// A live review exports a real repository whole; its Planner is asked about the files that changed.
	files = change.ChangedFor(ctx, files)
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
	for i, r := range rows {
		// Later steps and the trace key every finding on its id, so a missing or
		// repeated id is refused rather than merged. The error names the position and
		// not the id, because the id is model text and an error must carry none
		// (ADR-0005, ADR-0013 point 4).
		if r.ID == "" {
			return nil, fmt.Errorf("planner: risk %d has no id", i+1)
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("planner: risk %d repeats an earlier id", i+1)
		}
		seen[r.ID] = true
		out = append(out, candidate.Risk{
			ID: r.ID, RuleID: r.RuleID, Title: r.Title, Message: r.Message,
			Severity: r.Severity, Path: r.Path, StartLine: r.StartLine, EndLine: r.EndLine,
		})
	}
	return out, nil
}
