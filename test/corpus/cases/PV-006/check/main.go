// Command check runs every input against base, change and the reference, and
// prints one line per input: input name, base result, change result,
// reference result. A text result names the raw placeholders the control found.
package main

import (
	"fmt"
	"strings"

	"pv006/base"
	"pv006/change"
	"pv006/product"
	"pv006/reference"
)

const entry = "Too many requests. Please wait {wait} before trying again."

type textInput struct {
	name  string
	entry string
}

type durationInput struct {
	name              string
	shown, retryAfter int
}

func verdict(pass bool, raw []string) string {
	if pass {
		return "pass"
	}
	if len(raw) == 0 {
		return "fail"
	}
	return "fail(" + strings.Join(raw, ",") + ")"
}

// rawTokenEvidence requires the actual rendered notice to carry the raw token.
// A reference mismatch alone cannot establish which text reached the visitor.
func rawTokenEvidence(notice string) (string, error) {
	const token = "$wait"
	if !strings.Contains(notice, token) {
		return "", fmt.Errorf("rendered notice does not contain %q", token)
	}
	return token, nil
}

func main() {
	texts := []textInput{
		{"well-formed", entry},
		{"listed-token", entry + " Limit: {limit}."},
		{"unlisted-token", entry + " Window: {window}."},
		{"dollar-token", entry + " $wait"},
	}
	durations := []durationInput{
		{"duration-0-vs-1", 0, 1},
		{"duration-equal", 30, 30},
	}

	for _, in := range texts {
		notice, _ := product.Render(in.entry, 30)
		evidence := ""
		if in.name == "dollar-token" {
			token, err := rawTokenEvidence(notice)
			if err != nil {
				panic(err)
			}
			evidence = " raw=" + token
		}
		bp, braw := base.CheckText(notice)
		cp, craw := change.CheckText(notice)
		rp := reference.CheckText(in.name, notice)
		fmt.Printf("%s base=%s change=%s reference=%s%s\n",
			in.name, verdict(bp, braw), verdict(cp, craw), verdict(rp, nil), evidence)
	}
	for _, in := range durations {
		fmt.Printf("%s base=%s change=%s reference=%s shown=%d retry-after=%d\n", in.name,
			verdict(base.CheckDuration(in.shown, in.retryAfter), nil),
			verdict(change.CheckDuration(in.shown, in.retryAfter), nil),
			verdict(reference.CheckDuration(in.shown, in.retryAfter), nil),
			in.shown, in.retryAfter)
	}
}
