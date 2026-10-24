package validator

import (
	"testing"

	"pv018/data"
)

func TestShippedDataPasses(t *testing.T) {
	if got := Validate(data.ShippedPages()); len(got) != 0 {
		t.Fatalf("shipped data refused: %+v", got)
	}
}

func TestCitingRowOverNarrowBudgetIsRefused(t *testing.T) {
	pages := []data.Page{{Name: "install", Rows: []string{data.Row(data.NarrowBudget+1, true)}}}
	got := Validate(pages)
	want := data.Violation{Page: "install", Row: 0, Size: data.NarrowBudget + 1, Budget: data.NarrowBudget}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want [%+v]", got, want)
	}
}

func TestCitingRowAtNarrowBudgetPasses(t *testing.T) {
	pages := []data.Page{{Name: "limits", Rows: []string{data.Row(data.NarrowBudget, true)}}}
	if got := Validate(pages); len(got) != 0 {
		t.Fatalf("refused at the budget: %+v", got)
	}
}

func TestPlainRowKeepsWideBudget(t *testing.T) {
	between := []data.Page{{Name: "glossary", Rows: []string{data.Row(data.NarrowBudget+30, false)}}}
	if got := Validate(between); len(got) != 0 {
		t.Fatalf("plain row between the budgets refused: %+v", got)
	}
	over := []data.Page{{Name: "glossary", Rows: []string{data.Row(data.WideBudget+1, false)}}}
	got := Validate(over)
	want := data.Violation{Page: "glossary", Row: 0, Size: data.WideBudget + 1, Budget: data.WideBudget}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want [%+v]", got, want)
	}
}
