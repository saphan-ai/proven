package skeptic

import "strings"

// Family is the model family a review compares.
// An NVIDIA Nemotron model, whatever its vendor prefix, is the family "nemotron".
// Any other "vendor/name" model is the vendor. An empty model and "unknown" are "unknown".
func Family(model string) string {
	model = strings.TrimSpace(model)
	if model == "" || strings.EqualFold(model, "unknown") {
		return "unknown"
	}
	low := strings.ToLower(model)
	if strings.Contains(low, "nemotron") {
		return "nemotron"
	}
	if i := strings.IndexByte(low, '/'); i > 0 {
		return low[:i]
	}
	return low
}

// SameFamilySentence is the review's statement when the skeptic's family is
// the finder's family. It names that family and both models. The review still runs.
// The second result is false when the families differ or neither named a family.
func SameFamilySentence(skepticModel, finderModel string) (string, bool) {
	sf := Family(skepticModel)
	ff := Family(finderModel)
	if sf == "unknown" || sf != ff {
		return "", false
	}
	return "skeptic family " + sf + " is the same as the finder family " + ff + " (" + skepticModel + " and " + finderModel + ")", true
}
