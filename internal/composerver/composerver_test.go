package composerver

import "testing"

// TestComposerOrder proves the subset: padding and a leading v do not
// change order, stability ranks dev below alpha below beta below RC below
// stable below patch, a bare suffix is below the same suffix numbered 0,
// and build metadata is ignored.
func TestComposerOrder(t *testing.T) {
	equal := [][2]string{
		{"1.2", "1.2.0"},
		{"1.2.0", "1.2.0.0"},
		{"v1.2.3", "1.2.3"},
		{"1.2.3-stable", "1.2.3"},
		{"1.2.3-RC1", "1.2.3-rc.1"},
		{"1.2.3-RC-1", "1.2.3-rc1"},
		{"1.02", "1.2"},
		{"1.2.3+build", "1.2.3"},
		{"1.0.0-alpha", "1.0.0-a"},
		{"1.0.0-beta", "1.0.0-b"},
		{"1.0.0-p1", "1.0.0-patch1"},
		{"1.0.0-pl1", "1.0.0-p1"},
	}
	for _, pair := range equal {
		got, ok := Compare(pair[0], pair[1])
		if !ok || got != 0 {
			t.Fatalf("%s vs %s = %d, %t", pair[0], pair[1], got, ok)
		}
	}
	less := [][2]string{
		{"1.0.0-dev", "1.0.0-alpha"},
		{"1.0.0-alpha", "1.0.0-beta"},
		{"1.0.0-beta", "1.0.0-RC"},
		{"1.0.0-RC1", "1.0.0"},
		{"1.0.0", "1.0.0-p1"},
		{"1.0.0-RC", "1.0.0-RC0"},
		{"1.0.0-alpha", "1.0.0-alpha1"},
		{"1.9.9", "2.0.0-alpha"},
		{"1.2.3", "1.2.4"},
	}
	for _, pair := range less {
		got, ok := Compare(pair[0], pair[1])
		if !ok || got >= 0 {
			t.Fatalf("%s vs %s = %d, %t, want less", pair[0], pair[1], got, ok)
		}
	}
}

// TestComposerRefusals proves a constraint, a branch name and a string
// that is neither are each named, and are not ordered.
func TestComposerRefusals(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"^1.2", "^1.2 is a constraint; this check compares one Composer version"},
		{">=1.0", ">=1.0 is a constraint; this check compares one Composer version"},
		{"1.0.*", "1.0.* is a constraint; this check compares one Composer version"},
		{"1.0 as 1.0.0", "1.0 as 1.0.0 is a constraint; this check compares one Composer version"},
		{"dev-main", "dev-main is a branch name; this check compares numbered Composer versions only"},
		{"1.0.x-dev", "1.0.x-dev is a branch name; this check compares numbered Composer versions only"},
		{"latest", "latest is not a Composer version this check compares"},
		{"", "empty is not a Composer version this check compares"},
		{"1.2.3.4.5", "1.2.3.4.5 is not a Composer version this check compares"},
	}
	for _, tc := range cases {
		if _, ok := Compare(tc.in, "1.0.0"); ok {
			t.Fatalf("%q was ordered", tc.in)
		}
		if got := Reject(tc.in); got != tc.want {
			t.Fatalf("%q: got %q", tc.in, got)
		}
		why, ok := Explain("introduced", tc.in)
		if ok || why != "introduced "+tc.want {
			t.Fatalf("Explain %q: %q %t", tc.in, why, ok)
		}
	}
}
