package redact

import (
	"strings"
	"testing"
)

// TestAValueIsReplacedByAMarkerNamingItsVariable proves that a configured
// value is replaced whole by a marker naming the variable, never in part.
func TestAValueIsReplacedByAMarkerNamingItsVariable(t *testing.T) {
	r := New(map[string]string{"A_KEY": "secret-value-1"})
	got := r.Text("model secret-value-1 and again secret-value-1.")
	want := "model [redacted: A_KEY] and again [redacted: A_KEY]."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestALongerValueContainingAShorterOneIsReplacedWhole proves that when one
// secret contains another, no tail of the longer one is left in the text.
func TestALongerValueContainingAShorterOneIsReplacedWhole(t *testing.T) {
	r := New(map[string]string{"SHORT": "abcdef", "LONG": "abcdef-and-more"})
	got := r.Text("x abcdef-and-more y")
	if strings.Contains(got, "and-more") || strings.Contains(got, "abcdef") {
		t.Fatalf("part of the longer value survived: %q", got)
	}
}

// TestShortAndEmptyValuesAreNotRedacted proves the MinLength rule: an empty
// value or one shorter than MinLength leaves the text unchanged.
func TestShortAndEmptyValuesAreNotRedacted(t *testing.T) {
	for name, v := range map[string]string{"EMPTY": "", "BLANK": "   ", "SHORT": "12345"} {
		r := New(map[string]string{name: v})
		in := "line 12345 of 123456 files, ok"
		if got := r.Text(in); got != in {
			t.Errorf("%s: the text changed: %q", name, got)
		}
	}
}

// TestValueAtTheMinimumLengthIsRedacted proves the boundary of the rule: a
// value of exactly MinLength bytes is redacted.
func TestValueAtTheMinimumLengthIsRedacted(t *testing.T) {
	r := New(map[string]string{"K": strings.Repeat("x", MinLength)})
	if got := r.Text("a " + strings.Repeat("x", MinLength)); got != "a [redacted: K]" {
		t.Fatalf("got %q", got)
	}
}

// TestSurroundingWhiteSpaceOfAValueIsNotPartOfIt proves that a value loaded
// with a trailing newline is still found in text that lacks the newline.
func TestSurroundingWhiteSpaceOfAValueIsNotPartOfIt(t *testing.T) {
	r := New(map[string]string{"K": "value-123456\n"})
	if got := r.Text("echo value-123456."); got != "echo [redacted: K]." {
		t.Fatalf("got %q", got)
	}
}

// TestTheMarkerDoesNotDependOnMapOrder proves that two variables holding one
// value always give the marker of the name that sorts first.
func TestTheMarkerDoesNotDependOnMapOrder(t *testing.T) {
	for range 20 {
		r := New(map[string]string{"B": "same-value-9", "A": "same-value-9"})
		if got := r.Text("same-value-9"); got != "[redacted: A]" {
			t.Fatalf("got %q", got)
		}
	}
}

// TestFromEnvReadsEveryNamedVariable proves that FromEnv covers each name in
// Names, and that a nil environment redacts nothing.
func TestFromEnvReadsEveryNamedVariable(t *testing.T) {
	values := map[string]string{}
	for _, name := range Names {
		values[name] = "value-of-" + name
	}
	r := FromEnv(func(k string) string { return values[k] })
	for _, name := range Names {
		if got := r.Text(values[name]); got != "[redacted: "+name+"]" {
			t.Errorf("%s: got %q", name, got)
		}
	}
	if got := FromEnv(nil).Text("value-of-" + Names[0]); got != "value-of-"+Names[0] {
		t.Errorf("a nil environment changed the text: %q", got)
	}
}

// TestANilRedactorLeavesTextUnchanged proves that code holding no Redactor
// needs no guard before calling it.
func TestANilRedactorLeavesTextUnchanged(t *testing.T) {
	var r *Redactor
	if r.Text("abc") != "abc" || r.Texts(nil) != nil {
		t.Fatal("a nil Redactor changed its input")
	}
}
