// Package product is the renderer both versions of the controls are checked
// against. It is the same in base and change.
//
// It fills only a token in braces: an opening brace, one or more characters
// that are not braces, a closing brace. It knows one name, "wait". Any other
// token in braces, and any other text, is copied as it is.
package product

import (
	"regexp"
	"strconv"
)

// Token is the pattern the product uses to find a token to fill.
var Token = regexp.MustCompile(`\{[^{}]+\}`)

// Render returns the notice text and the retry-after value, in seconds, for a
// catalog entry and a wait time in seconds.
func Render(entry string, waitSeconds int) (text string, retryAfter int) {
	text = Token.ReplaceAllStringFunc(entry, func(tok string) string {
		if tok == "{wait}" {
			return strconv.Itoa(waitSeconds) + " seconds"
		}
		return tok
	})
	return text, waitSeconds
}
