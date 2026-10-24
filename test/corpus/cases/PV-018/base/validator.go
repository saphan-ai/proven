// Package validator is the base version: every row is held to one wide budget.
package validator

import "pv018/data"

// Validate returns every row over budget; an empty result means pass.
func Validate(pages []data.Page) []data.Violation {
	var out []data.Violation
	for _, p := range pages {
		for i, r := range p.Rows {
			if len(r) > data.WideBudget {
				out = append(out, data.Violation{Page: p.Name, Row: i, Size: len(r), Budget: data.WideBudget})
			}
		}
	}
	return out
}
