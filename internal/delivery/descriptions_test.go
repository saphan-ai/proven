package delivery

import (
	"strings"
	"testing"
)

func TestDescriptionsAreMechanicalAndNeedNoServices(t *testing.T) {
	for _, tc := range []struct{ text, want string }{
		{"", "missing"}, {"<!-- template -->\n## Summary\n- [ ] I described the change", "missing"},
		{"TBD", "placeholder"}, {"fix", "placeholder"}, {"fix bug", "too_short"},
		{"https://example.com/long/project/description #123", "too_short"},
		{"Reject all negative amounts", "present"}, {"Napraw walidację ujemnych kwot", "present"},
		// Passing this rule says nothing about whether a sentence makes sense.
		{"arbitrary unrelated gibberish", "present"},
	} {
		if got := CheckDescription(tc.text); got.Status != tc.want {
			t.Errorf("%q: %+v", tc.text, got)
		}
	}
	m := Metadata{Commits: []Commit{{SHA: "a", Message: "fix"}, {SHA: "b", Message: "fix"}}, Comments: []Comment{{Kind: "pr_comment", ID: 7, Body: "Reject all negative amounts"}}}
	r := Descriptions(m, "snapshot")
	if !r.Complete || r.Mode != "deterministic" || len(r.Claims) != 0 || len(r.Assessments) != 4 {
		t.Fatalf("%+v", r)
	}
	if r.Assessments[0].Description.Status != "missing" || r.Assessments[1].Source.ID == r.Assessments[2].Source.ID {
		t.Fatal("lost missing description or merged equal messages")
	}
	m.Context = "git"
	if got := Descriptions(m, "snapshot"); len(got.Assessments) != 3 {
		t.Fatal("local Git context invented a PR description")
	}
	copy := r.Redacted(func(s string) string { return strings.ReplaceAll(s, "negative", "redacted") })
	copy.Assessments[3].Source.Text = "changed"
	if r.Assessments[3].Source.Text == "changed" {
		t.Fatal("redaction aliased the original source")
	}
}
