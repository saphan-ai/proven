package hmem_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// TestDreamDomainProposalNeedsStaleDomain proves the property the Dream doc
// comment states: records added through the public API are stored with the
// domain the classifier returns, so Dream proposes no domain change for them.
// A proposal needs a stale stored domain; TestApproveDomainProposalRecoversForward
// covers that case.
func TestDreamDomainProposalNeedsStaleDomain(t *testing.T) {
	s := openApprove(t, t.TempDir(), t.TempDir())
	inputs := []hmem.AddInput{
		{Content: "the weather is mild", Source: "t"},
		{Content: "deploy the api server", Source: "t"},
		{Content: "client meeting before the deadline", Source: "t"},
		{Content: "I like quiet mornings", Type: hmem.TypePreference, Source: "t"},
		{Content: "plain note", Tags: []string{"project:atlas"}, Source: "t"},
		{Content: "deploy the service", Tags: []string{"tech:go"}, Source: "t"},
	}
	for _, in := range inputs {
		if _, err := s.Add(in); err != nil {
			t.Fatal(err)
		}
	}
	ps, err := s.Dream()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		if p.Type == "domain" {
			t.Fatalf("a domain proposal arose from ordinary Add: %+v", p)
		}
	}
}
