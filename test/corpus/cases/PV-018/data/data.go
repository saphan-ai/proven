// Package data holds what the validators and the check share: the page and
// violation types, the two budgets, the citation marker and the shipped rows.
package data

import "strings"

const (
	// WideBudget is the byte budget of every row in base.
	WideBudget = 200
	// NarrowBudget is the byte budget of a row that cites a measurement, in change.
	NarrowBudget = 120
	// Marker is the fixed citation marker; a row containing it cites a measurement.
	Marker = "{{measure:"
)

// Page is a named page and its rows of text.
type Page struct {
	Name string
	Rows []string
}

// Violation names one row over its budget and the budget it was held to.
type Violation struct {
	Page   string
	Row    int // zero-based index of the row on its page
	Size   int // bytes
	Budget int
}

// Cites reports whether a row cites a measurement.
func Cites(row string) bool { return strings.Contains(row, Marker) }

// Row returns a row of exactly size bytes, citing a measurement or not.
func Row(size int, cites bool) string {
	s := "Plain help text."
	if cites {
		s = "Takes " + Marker + "latency}} on a warm start."
	}
	if len(s) > size {
		panic("row too short for its content")
	}
	return s + strings.Repeat("x", size-len(s))
}

// ShippedPages is today's data: the citing rows sit on exactly three pages,
// install, limits and faq.
func ShippedPages() []Page {
	return []Page{
		{Name: "intro", Rows: []string{Row(60, false), Row(110, false), Row(190, false)}},
		{Name: "install", Rows: []string{Row(80, false), Row(100, true)}},
		{Name: "limits", Rows: []string{Row(120, true), Row(90, false), Row(115, true)}},
		{Name: "faq", Rows: []string{Row(70, true), Row(110, false)}},
		{Name: "glossary", Rows: []string{Row(40, false), Row(150, false)}},
	}
}
