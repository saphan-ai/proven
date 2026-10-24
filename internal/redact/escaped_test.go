package redact

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestAnEscapedFormOfAValueIsRedactedToo proves that a value written the way %q, %+q and a JSON
// encoder write it is replaced as well as the raw value (ADR-0005, ADR-0013): an error that quotes
// a key does not carry it past the redactor. Mutated field: only the spelling of the value.
func TestAnEscapedFormOfAValueIsRedactedToo(t *testing.T) {
	const secret = "sk-\"quoted\"\\slash\tünï-key"
	r := New(map[string]string{"A_KEY": secret})
	asJSON, err := json.Marshal(secret)
	if err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{
		"raw":         "token " + secret + " end",
		"%q":          fmt.Sprintf("token %q end", secret),
		"%+q":         fmt.Sprintf("token %+q end", secret),
		"json string": "token " + string(asJSON) + " end",
	} {
		got := r.Text(text)
		if !strings.Contains(got, "[redacted: A_KEY]") {
			t.Errorf("%s: not redacted: %q", name, got)
		}
		for _, leak := range []string{"quoted", "slash", "key"} {
			if strings.Contains(got, leak) {
				t.Errorf("%s: part of the value survived (%s): %q", name, leak, got)
			}
		}
	}
}

// TestAPlainValueHasNoExtraNeedles proves a value that is written the same escaped adds no second
// needle, so the marker count and the text of an ordinary run do not change.
func TestAPlainValueHasNoExtraNeedles(t *testing.T) {
	r := New(map[string]string{"A_KEY": "secret-value-1"})
	if len(r.needles) != 1 {
		t.Fatalf("%d needles for a plain value", len(r.needles))
	}
}
