// Command check runs the case's inputs against base, change, the mutant and the
// reference, and prints one line per input.
package main

import (
	"fmt"
	"strings"

	base "pv018/base"
	change "pv018/change"
	mutant "pv018/check/mutant"
	"pv018/data"
)

// reference applies the stated rule to each row itself: a citing row gets the
// narrow budget, wherever it is. It is not a validator.
func reference(pages []data.Page) []data.Violation {
	var out []data.Violation
	for _, p := range pages {
		for i, r := range p.Rows {
			b := data.WideBudget
			if strings.Contains(r, data.Marker) {
				b = data.NarrowBudget
			}
			if len(r) > b {
				out = append(out, data.Violation{Page: p.Name, Row: i, Size: len(r), Budget: b})
			}
		}
	}
	return out
}

func show(v []data.Violation) string {
	if len(v) == 0 {
		return "pass"
	}
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprintf("%s#%d:%d>%d", x.Page, x.Row, x.Size, x.Budget)
	}
	return "fail(" + strings.Join(parts, ",") + ")"
}

func main() {
	planted := data.ShippedPages()
	planted = append(planted, data.Page{Name: "changelog", Rows: []string{data.Row(160, true)}})
	plain := data.ShippedPages()
	for i := range plain {
		if plain[i].Name == "install" {
			plain[i].Rows = append(plain[i].Rows, data.Row(160, false))
		}
	}
	inputs := []struct {
		name  string
		pages []data.Page
	}{
		{"shipped", data.ShippedPages()},
		{"planted-citing-row-outside-the-three-pages", planted},
		{"plain-row-between-budgets-on-a-listed-page", plain},
	}
	for _, in := range inputs {
		fmt.Printf("input=%s base=%s change=%s mutant=%s reference=%s\n", in.name,
			show(base.Validate(in.pages)), show(change.Validate(in.pages)),
			show(mutant.Validate(in.pages)), show(reference(in.pages)))
	}
}
