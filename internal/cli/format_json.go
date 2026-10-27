// This file is the JSON writer of ADR-0020 §1. The field names are the
// contract in docs/OUTPUT_FORMATS.md. A later step reads the same document
// from json-file.

package cli

import (
	"io"

	"github.com/saphan-ai/proven/internal/review"
)

// jsonFormatter writes the review result as one JSON document (ADR-0020 §1).
type jsonFormatter struct{}

// Write encodes one review as one JSON document.
// DEL and C1 are escaped after encoding; see writeEncoded.
func (jsonFormatter) Write(w io.Writer, result review.Result) error {
	return writeEncoded(w, documentFrom(result))
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
	// ProofCounts is the totals by level (ADR-0044). Every level is a key,
	// with 0 when none, in the order review.ProofLevels lists them.
	ProofCounts jsonProofCounts `json:"proof_counts"`
	// Guardrail is what NVIDIA's guardrail checked in this review, always written.
	Guardrail jsonGuardrail `json:"guardrail"`
	Findings  []jsonFinding `json:"findings"`
}

// jsonGuardrail is the review's guardrail, counted per review. State is
// "checked", "off" or "no_call"; Line is the one line the text and Markdown
// outputs print, so a consumer can show the same words without composing them.
type jsonGuardrail struct {
	State      string `json:"state"`
	Line       string `json:"line"`
	Prompts    int    `json:"prompts"`
	Answers    int    `json:"answers"`
	Blocked    int    `json:"blocked"`
	Unanswered int    `json:"unanswered"`
	Backend    string `json:"backend"`
	Model      string `json:"model"`
}

// guardrailFrom is the guardrail object the JSON document and the SARIF run share.
func guardrailFrom(result review.Result) jsonGuardrail {
	c := result.Guardrail
	return jsonGuardrail{
		State: result.GuardrailState(), Line: result.GuardrailLine(),
		Prompts: c.Prompts, Answers: c.Answers, Blocked: c.Blocked, Unanswered: c.Unanswered,
		Backend: c.Backend, Model: c.Model,
	}
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
	ID string `json:"id"`
	// ProofLevel is this finding's one level of proof (ADR-0044). The string
	// is the core's, and it is always written, including when it is empty.
	ProofLevel string         `json:"proof_level"`
	RuleID     string         `json:"rule_id"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	Unresolved bool           `json:"unresolved"`
	Models     []string       `json:"models"`
	Location   jsonLocation   `json:"location"`
	Evidence   []jsonEvidence `json:"evidence"`
	// EvidenceRecordURL is the address of the finding's stored evidence record. It is omitted when
	// the finding has none: records are off, the record was not stored, or no link base is set.
	EvidenceRecordURL string `json:"evidence_record_url,omitempty"`
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
		Guardrail:            guardrailFrom(result),
		ProofCounts:          proofCountsFrom(result.ProofCounts),
		Findings:             findings,
	}
}

// jsonProofCounts is the totals by level. The names are review.ProofLevel.
// A level with no findings is 0, so a reader can add the four numbers and
// get the number of findings. Opinions stay in the agreed key at 0.
type jsonProofCounts struct {
	ProvedByTest    int `json:"proved-by-test"`
	ProvedByTool    int `json:"proved-by-tool"`
	ProvedByData    int `json:"proved-by-data"`
	AgreedNotProven int `json:"agreed-by-two-models-not-proven"`
}

// proofCountsFrom copies the core's totals. A missing map reads as zeros,
// which is the same document the core writes when it has counted nothing.
func proofCountsFrom(counts map[review.ProofLevel]int) jsonProofCounts {
	return jsonProofCounts{
		ProvedByTest:    counts[review.ProvedByTest],
		ProvedByTool:    counts[review.ProvedByTool],
		ProvedByData:    counts[review.ProvedByData],
		AgreedNotProven: counts[review.AgreedNotProven],
	}
}

func findingFrom(finding review.Finding) jsonFinding {
	return jsonFinding{
		ID:         finding.ID,
		ProofLevel: string(finding.ProofLevel),
		RuleID:     finding.RuleID,
		Title:      finding.Title,
		Message:    finding.Message,
		Severity:   string(finding.Severity),
		Unresolved: finding.Unresolved,
		Models:     stringSlice(finding.Models),
		Location: jsonLocation{
			Path: finding.Location.Path, StartLine: finding.Location.StartLine, EndLine: finding.Location.EndLine,
		},
		Evidence:          evidenceFrom(finding.Evidence),
		EvidenceRecordURL: finding.EvidenceRecordURL,
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
