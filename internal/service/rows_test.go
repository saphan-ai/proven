package service

import (
	"encoding/json"
	"testing"
)

// TestRowCopiesKeepTheirNilPolicies proves CopyRows leaves nil as nil and
// ListRows turns nil into an empty list. JSON is null for the first and []
// for the second. Swapping the two policies makes the matching case fail.
func TestRowCopiesKeepTheirNilPolicies(t *testing.T) {
	t.Run("nil stays nil", func(t *testing.T) {
		got := CopyRows(nil)
		if got != nil {
			t.Fatalf("CopyRows(nil) = %#v, want nil", got)
		}
		body, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "null" {
			t.Fatalf("json %s want null", body)
		}
	})
	t.Run("nil becomes an empty list", func(t *testing.T) {
		got := ListRows(nil)
		if got == nil || len(got) != 0 {
			t.Fatalf("ListRows(nil) = %#v, want an empty list", got)
		}
		body, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "[]" {
			t.Fatalf("json %s want []", body)
		}
	})
	t.Run("empty stays an empty list", func(t *testing.T) {
		for _, fn := range []func([][]string) [][]string{CopyRows, ListRows} {
			got := fn([][]string{})
			if got == nil || len(got) != 0 {
				t.Fatalf("empty = %#v", got)
			}
			body, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != "[]" {
				t.Fatalf("json %s want []", body)
			}
		}
	})
}
