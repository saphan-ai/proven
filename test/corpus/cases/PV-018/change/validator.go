// Package validator is the change: a row that cites a measurement is held to a
// narrow budget, because a citation expands when it is rendered; every other
// row keeps the wide budget. The narrow budget is derived from the row's
// content, never from a list of page names.
package validator

import "pv018/data"

// Validate returns every row over budget; an empty result means pass.
func Validate(pages []data.Page) []data.Violation {
	var out []data.Violation
	for _, p := range pages {
		for i, r := range p.Rows {
			budget := data.WideBudget
			if data.Cites(r) {
				budget = data.NarrowBudget
			}
			if len(r) > budget {
				out = append(out, data.Violation{Page: p.Name, Row: i, Size: len(r), Budget: budget})
			}
		}
	}
	return out
}
