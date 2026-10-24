// Package validator is the mutant of change, used only by the check: the content
// rule is replaced by "the page is one of these three names".
package validator

import "pv018/data"

var narrowPages = map[string]bool{"install": true, "limits": true, "faq": true}

// Validate returns every row over budget; an empty result means pass.
func Validate(pages []data.Page) []data.Violation {
	var out []data.Violation
	for _, p := range pages {
		for i, r := range p.Rows {
			budget := data.WideBudget
			if narrowPages[p.Name] {
				budget = data.NarrowBudget
			}
			if len(r) > budget {
				out = append(out, data.Violation{Page: p.Name, Row: i, Size: len(r), Budget: budget})
			}
		}
	}
	return out
}
