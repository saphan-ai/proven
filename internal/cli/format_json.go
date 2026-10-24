// This file is the JSON writer of ADR-0020 §1. The field names are the
// contract in docs/OUTPUT_FORMATS.md. A later step reads the same document
// from json-file.

package cli

import (
	"encoding/json"
	"io"

	"github.com/saphan-ai/proven/internal/review"
)

// jsonFormatter writes the review result as one JSON document (ADR-0020 §1).
type jsonFormatter struct{}

// Write encodes one review as one JSON document.
// The indent is two spaces so a person can read the file. HTML escaping stays
// off: evidence and messages are text, and the default encoder would turn "<"
// into an escape a person would have to undo.
func (jsonFormatter) Write(w io.Writer, result review.Result) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(documentFrom(result))
}

// jsonDocument is one review result.
// The names are the contract in docs/OUTPUT_FORMATS.md (ADR-0020 §1).
// A bool is always written: false means the review says so, and a missing
// key would look like a writer that forgot the field.
// A slice is never nil. documentFrom substitutes an empty slice, so the file
// has [] and a later step can range it. A nil slice would marshal as null.
type jsonDocument struct {
	Class                string            `json:"class"`
	ClassRule            string            `json:"class_rule"`
	Doubled              bool              `json:"doubled"`
	Partial              bool              `json:"partial"`
	SettingsStatement    string            `json:"settings_statement"`
	AuthorModelStatement string            `json:"author_model_statement"`
	ReviewerStatements   []string          `json:"reviewer_statements"`
	AuthorModels         []jsonAuthorModel `json:"author_models"`
	DefaultReviewer      string            `json:"default_reviewer"`
	SpecialReviewer      string            `json:"special_reviewer"`
	ReviewerModels       []string          `json:"reviewer_models"`
	TracePath            string            `json:"trace_path"`
	FamilySentence       string            `json:"family_sentence"`
	Models               []jsonModelUse    `json:"models"`
	FailedServices       []jsonService     `json:"failed_services"`
	Findings             []jsonFinding     `json:"findings"`
}

type jsonAuthorModel struct {
	Agent  string `json:"agent"`
	Model  string `json:"model"`
	Family string `json:"family"`
	Source string `json:"source"`
}

type jsonModelUse struct {
	Agent       string `json:"agent"`
	Route       string `json:"route"`
	Model       string `json:"model"`
	ReceiptID   string `json:"receipt_id"`
	FromReceipt bool   `json:"from_receipt"`
}

type jsonService struct {
	Service string `json:"service"`
	Reason  string `json:"reason"`
}

type jsonFinding struct {
	ID         string         `json:"id"`
	RuleID     string         `json:"rule_id"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	Unresolved bool           `json:"unresolved"`
	Models     []string       `json:"models"`
	Location   jsonLocation   `json:"location"`
	Evidence   []jsonEvidence `json:"evidence"`
}

type jsonLocation struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// jsonEvidence is one proof.
// test and query are omitted when absent: one proof is a test or a query,
// and a null object for the other is not the contract in docs/OUTPUT_FORMATS.md.
type jsonEvidence struct {
	Kind  string             `json:"kind"`
	Test  *jsonTestEvidence  `json:"test,omitempty"`
	Query *jsonQueryEvidence `json:"query,omitempty"`
}

type jsonTestEvidence struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	FailsOnChange bool   `json:"fails_on_change"`
	PassesOnBase  bool   `json:"passes_on_base"`
	Output        string `json:"output"`
}

type jsonQueryEvidence struct {
	SQL      string     `json:"sql"`
	Engine   string     `json:"engine"`
	Snapshot string     `json:"snapshot"`
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
}

func documentFrom(result review.Result) jsonDocument {
	findings := make([]jsonFinding, 0, len(result.Findings))
	for _, finding := range result.Findings {
		findings = append(findings, findingFrom(finding))
	}
	services := make([]jsonService, 0, len(result.FailedServices))
	for _, failed := range result.FailedServices {
		services = append(services, jsonService{Service: failed.Service, Reason: failed.Reason})
	}
	authors := make([]jsonAuthorModel, 0, len(result.AuthorModels))
	for _, author := range result.AuthorModels {
		authors = append(authors, jsonAuthorModel{
			Agent: author.Agent, Model: author.Model, Family: author.Family, Source: author.Source,
		})
	}
	uses := make([]jsonModelUse, 0, len(result.Models))
	for _, use := range result.Models {
		uses = append(uses, jsonModelUse{
			Agent: use.Agent, Route: use.Route, Model: use.Model,
			ReceiptID: use.ReceiptID, FromReceipt: use.FromReceipt,
		})
	}
	// Empty lists are [], not null (docs/OUTPUT_FORMATS.md). The slices above
	// are made with length zero, so a review with nothing to report still
	// writes an array. A nil slice would marshal as null.
	return jsonDocument{
		Class:                result.Class,
		ClassRule:            result.ClassRule,
		Doubled:              result.Doubled,
		Partial:              result.Partial,
		SettingsStatement:    result.SettingsStatement,
		AuthorModelStatement: result.AuthorModelStatement,
		ReviewerStatements:   stringSlice(result.ReviewerStatements),
		AuthorModels:         authors,
		DefaultReviewer:      result.DefaultReviewer,
		SpecialReviewer:      result.SpecialReviewer,
		ReviewerModels:       stringSlice(result.ReviewerModels),
		TracePath:            result.TracePath,
		FamilySentence:       familySentence(result.TracePath),
		Models:               uses,
		FailedServices:       services,
		Findings:             findings,
	}
}

func findingFrom(finding review.Finding) jsonFinding {
	return jsonFinding{
		ID:         finding.ID,
		RuleID:     finding.RuleID,
		Title:      finding.Title,
		Message:    finding.Message,
		Severity:   string(finding.Severity),
		Unresolved: finding.Unresolved,
		Models:     stringSlice(finding.Models),
		Location: jsonLocation{
			Path: finding.Location.Path, StartLine: finding.Location.StartLine, EndLine: finding.Location.EndLine,
		},
		Evidence: evidenceFrom(finding.Evidence),
	}
}

func evidenceFrom(in []review.Evidence) []jsonEvidence {
	out := make([]jsonEvidence, 0, len(in))
	for _, evidence := range in {
		one := jsonEvidence{Kind: evidence.Kind}
		if evidence.Test != nil {
			test := evidence.Test
			one.Test = &jsonTestEvidence{
				Path: test.Path, Name: test.Name, Command: test.Command,
				FailsOnChange: test.FailsOnChange, PassesOnBase: test.PassesOnBase, Output: test.Output,
			}
		}
		if evidence.Query != nil {
			query := evidence.Query
			rows := make([][]string, 0, len(query.Rows))
			for _, row := range query.Rows {
				rows = append(rows, stringSlice(row))
			}
			one.Query = &jsonQueryEvidence{
				SQL: query.SQL, Engine: query.Engine, Snapshot: query.Snapshot,
				Columns: stringSlice(query.Columns), Rows: rows,
			}
		}
		out = append(out, one)
	}
	return out
}

// stringSlice returns an empty slice when in is empty, never nil.
// encoding/json writes a nil slice as null, and the contract says [].
// The copy also keeps the document from sharing the result's backing array.
func stringSlice(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
