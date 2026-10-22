package change

import "testing"

func TestStale(t *testing.T) {
	tests := []struct {
		name string
		note string
		want bool
	}{
		{"current", "LANGUAGES 41 -> 42", false},
		{"stale", "LANGUAGES 41 -> 43", true},
		{"stale after current", "LANGUAGES 41 -> 42; 42 -> 43", true},
		{"current after stale", "LANGUAGES 41 -> 43; 43 -> 42", true},
		{"no arrow", "LANGUAGES are available", false},
		{"largest convertible", "LANGUAGES 41 -> 9223372036854775807", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Stale([]string{test.note}, 42)
			if (len(got) == 1) != test.want || len(got) > 1 || (len(got) == 1 && got[0] != test.note) {
				t.Fatalf("Stale(%q, 42) = %q, want stale=%t", test.note, got, test.want)
			}
		})
	}
}
