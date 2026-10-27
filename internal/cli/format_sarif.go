// This file writes one review result as a SARIF 2.1.0 log (ADR-0020 §1).
// Where each field sits is the contract in docs/OUTPUT_FORMATS.md.

package cli

import (
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

// sarifSchemaURI names the OASIS SARIF 2.1.0 schema so a consumer can see
// which document this log claims. version is the field the schema requires;
// this URL is not a substitute for it. The test checks a vendored copy of
// that schema and does not fetch this URL.
const sarifSchemaURI = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"

// sarifFormatter writes one SARIF 2.1.0 log: one run, driver name proven.
type sarifFormatter struct{}

// Write encodes one review as one SARIF 2.1.0 log.
// DEL and C1 are escaped after encoding, the same way as the JSON writer.
func (sarifFormatter) Write(w io.Writer, result review.Result) error {
	return writeEncoded(w, sarifFrom(result))
}

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool       sarifTool      `json:"tool"`
	Results    []sarifResult  `json:"results"`
	Properties map[string]any `json:"properties,omitempty"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name string `json:"name"`
}

type sarifResult struct {
	RuleID     string          `json:"ruleId,omitempty"`
	Level      string          `json:"level"`
	Message    sarifMessage    `json:"message"`
	Locations  []sarifLocation `json:"locations,omitempty"`
	Properties map[string]any  `json:"properties,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           *sarifRegion  `json:"region,omitempty"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine,omitempty"`
}

func sarifFrom(result review.Result) sarifLog {
	results := make([]sarifResult, 0, len(result.Findings))
	for _, finding := range result.Findings {
		results = append(results, sarifResultFrom(finding))
	}
	services := make([]jsonService, 0, len(result.FailedServices))
	for _, failed := range result.FailedServices {
		services = append(services, jsonService{Service: failed.Service, Reason: failed.Reason})
	}
	// One review is one run (docs/OUTPUT_FORMATS.md). Failed services belong
	// to the run: they are the review's own gaps, not a finding.
	// results and failed_services are empty arrays when nothing is there.
	// A nil slice would marshal as null, and the contract says [].
	// The lines the text output prints beside the findings ride in the run's
	// properties too, so a SARIF-only consumer loses none of them.
	return sarifLog{
		Schema:  sarifSchemaURI,
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool:    sarifTool{Driver: sarifDriver{Name: "proven"}},
			Results: results,
			Properties: map[string]any{
				"class":           result.Class,
				"class_rule":      result.ClassRule,
				"partial":         result.Partial,
				"failed_services": services,

				"settings_statement":     result.SettingsStatement,
				"author_model_statement": result.AuthorModelStatement,
				"reviewer_statements":    stringSlice(result.ReviewerStatements),
				"trace_path":             result.TracePath,
				"family_sentence":        familySentence(result.TracePath),
				// Same object the JSON document writes under proof_counts.
				"proof_counts": proofCountsFrom(result.ProofCounts),
				// Same object the JSON document writes under guardrail.
				"guardrail": guardrailFrom(result),
			},
		}},
	}
}

func sarifResultFrom(finding review.Finding) sarifResult {
	res := sarifResult{
		RuleID:  sarifRuleID(finding),
		Level:   sarifLevel(finding.Severity),
		Message: sarifMessage{Text: sarifMessageText(finding)},
		// Evidence uses the same objects as the JSON document, under properties.
		// codeFlows would need a path through code. A query has none, so the
		// proof would be dropped. docs/OUTPUT_FORMATS.md puts evidence here.
		Properties: map[string]any{
			"id":          finding.ID,
			"proof_level": string(finding.ProofLevel),
			"rule_id":     finding.RuleID,
			"title":       finding.Title,
			"severity":    string(finding.Severity),
			"unresolved":  finding.Unresolved,
			"evidence":    evidenceFrom(finding.Evidence),
		},
	}
	if uri := artifactURI(finding.Location.Path); uri != "" {
		loc := sarifLocation{
			PhysicalLocation: sarifPhysical{ArtifactLocation: sarifArtifact{URI: uri}},
		}
		// The vendored schema sets startLine's minimum at 1. A zero line is not a region.
		// endLine is written only when it is at or after startLine, so the
		// region does not end before it starts. That schema does not check the order.
		if finding.Location.StartLine >= 1 {
			region := &sarifRegion{StartLine: finding.Location.StartLine}
			if finding.Location.EndLine >= finding.Location.StartLine {
				region.EndLine = finding.Location.EndLine
			}
			loc.PhysicalLocation.Region = region
		}
		res.Locations = []sarifLocation{loc}
	}
	return res
}

// sarifRuleID is rule_id, or the finding id when the finding has no rule
// (docs/OUTPUT_FORMATS.md). An empty string is omitted: SARIF ruleId is optional.
func sarifRuleID(finding review.Finding) string {
	if finding.RuleID != "" {
		return finding.RuleID
	}
	return finding.ID
}

// sarifLevel maps the review's severity onto SARIF (ADR-0020 §1):
// INFO is note, WARNING is warning, ERROR is error.
// An unnamed severity is a note, the same rank the text filter gives INFO,
// so a finding is never dropped for a name this writer does not know.
func sarifLevel(severity review.Severity) string {
	switch severity {
	case review.SeverityWarning:
		return "warning"
	case review.SeverityError:
		return "error"
	default:
		return "note"
	}
}

// sarifMessageText fills result.message.text.
// The vendored schema requires message on a result, and a message requires
// text or an id. This writer always sends text. The finding's message wins,
// then its title, then its id (docs/OUTPUT_FORMATS.md). The word "finding"
// is the last resort, so the text is never empty when all three are blank.
func sarifMessageText(finding review.Finding) string {
	if finding.Message != "" {
		return finding.Message
	}
	if finding.Title != "" {
		return finding.Title
	}
	if finding.ID != "" {
		return finding.ID
	}
	return "finding"
}

// artifactURI makes a SARIF artifact URI from a path.
// The result is a relative reference under the repository root: no leading
// "/", no authority, and no ".." segment. Each segment is escaped on its own
// so "/" stays the separator. A ":" in the first segment is escaped too:
// url.PathEscape leaves it, and "a:b/c" would read as a URI with scheme "a"
// (RFC 3986 §4.2).
//
// WHY a path that cannot be made so has no location: this log does not set
// uriBaseId. An absolute path, a path with an authority, or a ".." that
// survives cleaning would name a file outside the checkout. The finding is
// kept; sarifResultFrom omits the location when this returns empty. Cleaning
// comes first so "dir/../a.go" is still "a.go", inside the repository.
// An empty path returns empty: there is no artifact to point at.
func artifactURI(p string) string {
	p = filepath.ToSlash(strings.TrimSpace(p))
	if p == "" {
		return ""
	}
	cleaned := path.Clean(p)
	if !staysInRepository(cleaned) {
		return ""
	}
	parts := strings.Split(cleaned, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	parts[0] = strings.ReplaceAll(parts[0], ":", "%3A")
	return strings.Join(parts, "/")
}

// staysInRepository reports whether cleaned is a relative path that does not
// climb out of the repository. path.Clean has already collapsed inner ".."
// and duplicate slashes, so a leading "/" is a root or an authority, and a
// ".." segment is a path that still leaves the checkout.
func staysInRepository(cleaned string) bool {
	if cleaned == "" || strings.HasPrefix(cleaned, "/") || cleaned == ".." {
		return false
	}
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == "" || seg == ".." {
			return false
		}
	}
	return true
}
