// Package guardrail is a worked example of a guardrail adapter.
//
// List answers Check from a fixed word list: a barred word in the prompt or
// the answer makes that side unsafe, with the word as its category, and
// everything else is safe. A reader copies this package and replaces the list
// with their own rule. What to add is in docs/ADAPTERS.md under
// "Adding a guardrail".
package guardrail

import (
	"context"
	"errors"
	"strings"

	ig "github.com/saphan-ai/proven/internal/guardrail"
)

// The assignment fails the build when internal/guardrail.Adapter changes,
// so a copy of this example stays bound to the real interface the agents call.
var _ ig.Adapter = List{}

// The assignment also binds List to Identity, so Wrap can name its backend.
var _ ig.Identity = List{}

// ErrDuplicateWord is the refusal: the list names one word twice.
var ErrDuplicateWord = errors.New("guardrail: two entries for one word")

// Row is one barred word and the category the verdict names when it appears.
type Row struct {
	Word     string
	Category string
}

// List answers Check from a fixed word list.
type List struct {
	words map[string]string // word -> category
}

// NewList builds the list from rows. A duplicate word is refused.
func NewList(rows []Row) (*List, error) {
	words := make(map[string]string, len(rows))
	for _, row := range rows {
		if _, ok := words[row.Word]; ok {
			return nil, errors.New("guardrail: two entries for word " + row.Word)
		}
		words[row.Word] = row.Category
	}
	return &List{words: words}, nil
}

// Check answers from the list. A barred word in the prompt makes the user
// side unsafe; a barred word in the answer makes the response side unsafe.
// Everything else is safe. A canceled context is refused and nothing is judged.
func (l List) Check(ctx context.Context, prompt, answer string) (ig.Verdict, error) {
	if err := ctx.Err(); err != nil {
		return ig.Verdict{}, err
	}
	v := ig.Verdict{User: "safe", Response: "safe"}
	if category, ok := l.barred(prompt); ok {
		v.User = "unsafe"
		v.Categories = append(v.Categories, category)
	}
	if category, ok := l.barred(answer); ok {
		v.Response = "unsafe"
		v.Categories = append(v.Categories, category)
	}
	return v, nil
}

// barred reports the category of the first barred word in the text, or "" when the text is clean.
// The match is a case-insensitive substring, so "Forbidden" and "forbidden" are the same word.
func (l List) barred(text string) (string, bool) {
	low := strings.ToLower(text)
	for word, category := range l.words {
		if strings.Contains(low, word) {
			return category, true
		}
	}
	return "", false
}

// Identity names this adapter "word-list". It calls no model.
func (List) Identity() (backend, model string) { return "word-list", "" }
