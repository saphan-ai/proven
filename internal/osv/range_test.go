package osv

import "testing"

func TestInRange(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		introduced string
		fixed      string
		last       string
		want       bool
		sentence   string
	}{
		{name: "introduced only", version: "1.2.3", introduced: "1.0.0", want: true, sentence: "introduced 1.0.0 <= 1.2.3"},
		{name: "introduced only below", version: "0.9.0", introduced: "1.0.0", want: false, sentence: "0.9.0 < introduced 1.0.0"},
		{name: "fixed", version: "1.2.3", introduced: "1.0.0", fixed: "1.2.4", want: true, sentence: "introduced 1.0.0 <= 1.2.3 < fixed 1.2.4"},
		{name: "equal to fixed", version: "1.2.4", introduced: "1.0.0", fixed: "1.2.4", want: false, sentence: "fixed 1.2.4 <= 1.2.4"},
		{name: "last_affected", version: "1.2.3", introduced: "1.0.0", last: "1.2.3", want: true, sentence: "introduced 1.0.0 <= 1.2.3 <= last_affected 1.2.3"},
		{name: "above last_affected", version: "1.2.4", introduced: "1.0.0", last: "1.2.3", want: false, sentence: "last_affected 1.2.3 < 1.2.4"},
		{name: "pre-release", version: "1.2.3-rc.1", introduced: "1.2.0", fixed: "1.2.3", want: true, sentence: "introduced 1.2.0 <= 1.2.3-rc.1 < fixed 1.2.3"},
		{name: "pre-release before introduced", version: "1.2.3-rc.1", introduced: "1.2.3", want: false, sentence: "1.2.3-rc.1 < introduced 1.2.3"},
		{name: "outside every range", version: "3.0.0", introduced: "1.0.0", fixed: "2.0.0", want: false, sentence: "fixed 2.0.0 <= 3.0.0"},
		{name: "introduced zero", version: "v1.2.3", introduced: "0", fixed: "1.2.4", want: true, sentence: "v1.2.3 < fixed 1.2.4"},
		{name: "v prefix", version: "v1.2.3", introduced: "1.0.0", fixed: "1.2.4", want: true, sentence: "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4"},
		{name: "fixed wins over last_affected", version: "1.3.0", introduced: "1.0.0", fixed: "1.2.0", last: "1.5.0", want: false, sentence: "fixed 1.2.0 <= 1.3.0"},
		{name: "not semantic", version: "latest", introduced: "1.0.0", want: false, sentence: "latest is not a semantic version"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, sentence := InRange(tc.version, tc.introduced, tc.fixed, tc.last)
			if got != tc.want || sentence != tc.sentence {
				t.Fatalf("got %t %q", got, sentence)
			}
		})
	}
}
