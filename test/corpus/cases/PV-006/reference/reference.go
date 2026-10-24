// Package reference is the oracle used only by the check. It is written by
// hand and never derived from the controls: for each input it holds the exact
// notice a visitor must see, and the duration test is exact equality.
package reference

// Notice is the exact text a visitor must see for a named input. The texts
// are written out in full on purpose.
var Notice = map[string]string{
	"well-formed":    "Too many requests. Please wait 30 seconds before trying again.",
	"listed-token":   "Too many requests. Please wait 30 seconds before trying again. Limit: 100.",
	"unlisted-token": "Too many requests. Please wait 30 seconds before trying again. Window: 60.",
	"dollar-token":   "Too many requests. Please wait 30 seconds before trying again. 30 seconds",
}

// CheckText passes only when the rendered notice is exactly the expected text.
func CheckText(input, rendered string) bool {
	want, ok := Notice[input]
	return ok && rendered == want
}

// CheckDuration passes only when the duration shown equals the retry-after value.
func CheckDuration(shown, retryAfter int) bool {
	return shown == retryAfter
}
