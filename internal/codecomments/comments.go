// Package codecomments checks comment structure through language adapters.
// It performs no I/O, executes no target code and calls no model.
package codecomments

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/change"
)

// Unit is one changed declaration or comment, with exact head source lines.
type Unit struct {
	ID       string   `json:"id"`
	Path     string   `json:"path"`
	Symbol   string   `json:"symbol"`
	Public   bool     `json:"public"`
	Line     int      `json:"line"`
	CodeLine int      `json:"code_line"`
	Text     string   `json:"text"`
	Code     string   `json:"code"`
	Params   []string `json:"params,omitempty"`
}

// Issue is a deterministic documentation observation, not a code-defect proof.
type Issue struct {
	UnitID  string `json:"unit_id"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// Assessment retains one model interpretation and its checked code quote.
type Assessment struct {
	UnitID string `json:"unit_id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	Quote  string `json:"quote,omitempty"`
}

// Result separates structural issues, model assessments and coverage gaps.
type Result struct {
	Mode        string       `json:"mode"`
	Complete    bool         `json:"complete"`
	Units       []Unit       `json:"units"`
	Issues      []Issue      `json:"issues"`
	Assessments []Assessment `json:"assessments"`
	Unsupported []string     `json:"unsupported_paths,omitempty"`
	Issue       string       `json:"issue,omitempty"`
}

// Adapter extracts comment/declaration associations without executing code.
// Language support here is independent of manifest and test-run adapters.
type Adapter interface {
	Name() string
	Matches(string) bool
	Extract(path string, content []byte) ([]Unit, error)
}

// Adapters returns the controlled registry of comment-language parsers.
func Adapters() []Adapter { return []Adapter{goAdapter{}} }

var paramTag = regexp.MustCompile(`@param\s+(?:\{[^}]*\}\s*)?([A-Za-z_][A-Za-z_0-9]*)`)
var placeholder = regexp.MustCompile(`(?i)^(todo|fixme|tbd|wip)(?:[.!:]?\s*(?:document|describe|explain|add docs|write docs)\b.*|[.!:]?)$`)

// Check compares base/head associations. Unchanged legacy comments and symbols
// are outside this change review, even when another part of their file changed.
func Check(files []change.File) Result {
	r := Result{Mode: "deterministic", Complete: true, Units: []Unit{}, Issues: []Issue{}, Assessments: []Assessment{}}
	for _, f := range files {
		if len(f.Change) == 0 {
			continue
		}
		var adapter Adapter
		for _, a := range Adapters() {
			if a.Matches(f.Path) {
				adapter = a
				break
			}
		}
		if adapter == nil {
			switch strings.ToLower(path.Ext(f.Path)) {
			case ".py", ".js", ".jsx", ".ts", ".tsx", ".kt", ".kts", ".rs", ".php", ".swift", ".java":
				r.Unsupported = append(r.Unsupported, f.Path)
			}
			continue
		}
		head, err := adapter.Extract(f.Path, f.Change)
		if err != nil {
			r.Complete = false
			r.Issue = "A selected Go file could not be parsed; comment coverage is incomplete."
			continue
		}
		old := map[string]Unit{}
		if len(f.Base) > 0 {
			base, err := adapter.Extract(f.Path, f.Base)
			if err != nil {
				r.Complete = false
				r.Issue = "A baseline Go file could not be parsed; changed comment coverage is incomplete."
				continue
			}
			for _, u := range base {
				old[u.ID] = u
			}
		}
		for _, u := range head {
			if b, ok := old[u.ID]; ok && b.Text == u.Text && b.Code == u.Code {
				continue
			}
			if len(r.Units) >= 256 || len(u.Text) > 8192 || len(u.Code) > 32768 {
				r.Complete = false
				r.Issue = "Comment review reached its 256-unit, 8 KiB comment or 32 KiB declaration bound; omitted units were not checked."
				continue
			}
			r.Units = append(r.Units, u)
			add := func(rule, message string) { r.Issues = append(r.Issues, Issue{u.ID, rule, message}) }
			txt := strings.TrimSpace(u.Text)
			if u.Public && txt == "" {
				add("missing_api_doc", "Changed exported declaration has no associated documentation comment.")
			}
			if txt != "" && placeholder.MatchString(txt) {
				add("placeholder_comment", "Comment is only a recognized documentation placeholder.")
			}
			if u.Public && txt != "" && !startsDocName(txt, u.Symbol) {
				add("doc_name", "Documentation does not start with the exported name; this is a structural convention check, not a semantic verdict.")
			}
			if u.Params != nil {
				for _, tag := range paramTag.FindAllStringSubmatch(txt, -1) {
					found := false
					for _, p := range u.Params {
						found = found || p == tag[1]
					}
					if !found {
						add("stale_param_tag", fmt.Sprintf("@param %s does not name an input parameter in the current declaration.", tag[1]))
					}
				}
			}
		}
	}
	return r
}

func startsDocName(text, name string) bool {
	if strings.HasPrefix(text, "["+name+"]") {
		return true
	}
	if !strings.HasPrefix(text, name) {
		return false
	}
	if len(text) == len(name) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(text[len(name):])
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
}

// Redacted returns a deep copy with operator-defined secret filtering applied.
func (r Result) Redacted(text func(string) string) *Result {
	r.Issue = text(r.Issue)
	r.Units = append([]Unit{}, r.Units...)
	r.Issues = append([]Issue{}, r.Issues...)
	r.Assessments = append([]Assessment{}, r.Assessments...)
	r.Unsupported = append([]string(nil), r.Unsupported...)
	for i, u := range r.Units {
		u.ID, u.Path, u.Symbol, u.Text, u.Code = text(u.ID), text(u.Path), text(u.Symbol), text(u.Text), text(u.Code)
		u.Params = append([]string(nil), u.Params...)
		for j := range u.Params {
			u.Params[j] = text(u.Params[j])
		}
		r.Units[i] = u
	}
	for i := range r.Issues {
		x := &r.Issues[i]
		x.UnitID, x.Rule, x.Message = text(x.UnitID), text(x.Rule), text(x.Message)
	}
	for i := range r.Assessments {
		x := &r.Assessments[i]
		x.UnitID, x.Status, x.Reason, x.Quote = text(x.UnitID), text(x.Status), text(x.Reason), text(x.Quote)
	}
	for i := range r.Unsupported {
		r.Unsupported[i] = text(r.Unsupported[i])
	}
	return &r
}
