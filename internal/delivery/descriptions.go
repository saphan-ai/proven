package delivery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// DescriptionCheck records a mechanical completeness check, never a judgment
// about meaning, correctness or whether the described work was delivered.
type DescriptionCheck struct {
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Words   int    `json:"words"`
	Letters int    `json:"letters"`
}

var templateComments = regexp.MustCompile(`(?s)<!--.*?-->`)
var webReferences = regexp.MustCompile(`https?://\S+|#[0-9]+`)

func CheckDescription(text string) DescriptionCheck {
	text = templateComments.ReplaceAllString(text, " ")
	var content []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "* [ ]") {
			continue
		}
		content = append(content, line)
	}
	text = strings.TrimSpace(strings.Join(content, " "))
	if text == "" {
		return DescriptionCheck{Status: "missing", Reason: "No description beyond template headings, unchecked boxes or HTML comments."}
	}
	placeholder := strings.Trim(strings.ToLower(text), " .!?:-_`*[]()")
	switch placeholder {
	case "todo", "tbd", "wip", "n/a", "na", "none", "test", "fix", "update", "no description", "no description provided":
		return DescriptionCheck{Status: "placeholder", Reason: "Description consists only of a recognized placeholder."}
	}
	text = webReferences.ReplaceAllString(text, " ")
	words, letters, inWord := 0, 0, false
	for _, r := range text {
		if unicode.IsLetter(r) {
			letters++
			if !inWord {
				words++
				inWord = true
			}
		} else {
			inWord = false
		}
	}
	if words < 3 || letters < 12 {
		return DescriptionCheck{Status: "too_short", Words: words, Letters: letters, Reason: "Description has fewer than 3 words or 12 letters after removing links and issue references."}
	}
	return DescriptionCheck{Status: "present", Words: words, Letters: letters, Reason: "Description passes the mechanical 3-word/12-letter rule; its meaning and delivery were not judged by this check."}
}

func MetadataDigest(m Metadata) string {
	b, _ := json.Marshal(m)
	digest := sha256.Sum256(b)
	return hex.EncodeToString(digest[:])
}

// Descriptions needs no model, database or network. Every source keeps its own
// result, even when several messages say exactly the same thing.
func Descriptions(m Metadata, snapshot string) Result {
	r := Result{Mode: "deterministic", Snapshot: snapshot, MetadataSHA256: MetadataDigest(m), URL: m.URL, Claims: []Claim{}, Assessments: []Assessment{}}
	for _, s := range Sources(m) {
		r.Assessments = append(r.Assessments, Assessment{Source: s, Description: CheckDescription(s.Text), SemanticStatus: "not_requested", Complete: true, Claims: []Claim{}})
	}
	r.Complete = len(m.Issues) == 0
	if !r.Complete {
		r.Issue = fmt.Sprintf("source collection incomplete: %s", strings.Join(m.Issues, "; "))
	}
	return r
}
