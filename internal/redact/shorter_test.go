package redact

import (
	"reflect"
	"testing"
)

// Shorter names the configured values under a floor, by variable, sorted, and
// counts a value by its trimmed length. A nil Redactor has none.
func TestShorterNamesTheValuesUnderTheFloor(t *testing.T) {
	r := New(map[string]string{"B_SHORT": "abc", "A_SHORT": " abcd ", "C_LONG": "abcdefgh", "D_EMPTY": "  "})
	if got, want := r.Shorter(8), []string{"A_SHORT", "B_SHORT"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Shorter(8) = %v, want %v", got, want)
	}
	if got := r.Shorter(4); !reflect.DeepEqual(got, []string{"B_SHORT"}) {
		t.Fatalf("Shorter(4) = %v, want only B_SHORT: a value of exactly the floor is not under it", got)
	}
	var none *Redactor
	if got := none.Shorter(8); got != nil {
		t.Fatalf("a nil Redactor has no short value, got %v", got)
	}
}

// Two variables that hold one value are both counted, though one needle serves them.
func TestShorterCountsEveryVariableOfASharedValue(t *testing.T) {
	r := New(map[string]string{"A": "short", "B": "short"})
	if got := r.Shorter(8); !reflect.DeepEqual(got, []string{"A", "B"}) {
		t.Fatalf("Shorter = %v", got)
	}
}

// The spelling a JSON encoder with HTML escaping off gives is a needle, and a
// value that ends in a quote keeps its escaped quote in every spelling.
func TestTheEscapeOffJSONSpellingIsRedacted(t *testing.T) {
	value := "tok\"en<12­34\""
	r := New(map[string]string{"K": value})
	for name, text := range map[string]string{
		"raw":     value,
		"escaped": `tok\"en<12` + "­" + `34\"`,
	} {
		if got := r.Text("x " + text + " y"); got != "x [redacted: K] y" {
			t.Errorf("%s spelling: %q", name, got)
		}
	}
}
