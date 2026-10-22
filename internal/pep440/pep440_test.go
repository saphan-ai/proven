package pep440

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0", "1.0.0", 0},
		{"1.2", "1.2.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0a1", "1.0b1", -1},
		{"1.0b1", "1.0rc1", -1},
		{"1.0rc1", "1.0", -1},
		{"1.0", "1.0.post1", -1},
		{"1.0a1", "1.0a2", -1},
		{"1.0alpha1", "1.0a1", 0},
		{"1.0.beta.1", "1.0b1", 0},
		{"1.0rc1", "1.0c1", 0},
		{"1.0preview1", "1.0rc1", 0},
		{"1.0pre1", "1.0rc1", 0},
		{"1.0-1", "1.0.post1", 0},
		{"1.0rev2", "1.0.post2", 0},
		{"1.0r1", "1.0.post1", 0},
		{"1.0-post1", "1.0.post1", 0},
		{"v1.2.3", "1.2.3", 0},
		{"1.01", "1.1", 0},
		{"1.0a1.post1", "1.0a1", 1},
		{"1.0A1", "1.0a1", 0},
		{"2.0", "10.0", -1},
		{"1.0.post2", "1.0.post10", -1},
		{"1.0a", "1.0a0", 0},
		{"1.0.a.1", "1.0a1", 0},
		{"1.0-alpha-1", "1.0a1", 0},
		{"1.2.3.4", "1.2.3", 1},
	}
	for _, tc := range cases {
		got, ok := Compare(tc.a, tc.b)
		if !ok || got != tc.want {
			t.Fatalf("Compare(%s, %s) = %d, %t; want %d", tc.a, tc.b, got, ok, tc.want)
		}
	}
}

func TestReject(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"latest", "latest is not a release, pre-release or post-release"},
		{"1.0.dev1", "1.0.dev1 has a dev release; this check does not compare it"},
		{"1.0dev1", "1.0dev1 has a dev release; this check does not compare it"},
		{"1!1.0", "1!1.0 has an epoch; this check compares release, pre-release and post-release only"},
		{"1.0+local", "1.0+local has a local version; this check does not compare it"},
		{"1.2.*", "1.2.* is not a final release, pre-release or post-release"},
		{"", "empty is not a release, pre-release or post-release"},
		{"1.0.", "1.0. is not a release, pre-release or post-release"},
	}
	for _, tc := range cases {
		if _, ok := Parse(tc.in); ok {
			t.Fatalf("Parse(%q) succeeded", tc.in)
		}
		if got := Reject(tc.in); got != tc.want {
			t.Fatalf("Reject(%q) = %q", tc.in, got)
		}
		if _, ok := Compare(tc.in, "1.0"); ok {
			t.Fatalf("Compare(%q, 1.0) succeeded", tc.in)
		}
	}
	if Reject("1.2.3") != "" {
		t.Fatal("a release was rejected")
	}
}
