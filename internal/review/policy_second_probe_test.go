package review

import (
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

func TestLocalDoubledListCannotNarrowEffectivePolicyWithSecondReviewer(t *testing.T) {
	p, err := loadPolicy(change.ProcessPolicy(), writeFile(t, "policy.yaml", "path_rules: go.mod=security\n"))
	if err != nil {
		t.Fatal(err)
	}
	local, err := p.parseLocal([]byte("doubled_classes: off\n"))
	if err != nil {
		return
	}
	policySettings := withSecond(p.settings, "second-model")
	localSettings := withSecond(local, "second-model")
	path := "main.go"
	policyClass, _ := chooseClass([]string{path}, policySettings.PathRules, policySettings.ClassOrder, policySettings.Fallback)
	localClass, _ := chooseClass([]string{path}, localSettings.PathRules, localSettings.ClassOrder, localSettings.Fallback)
	if policySettings.Doubled[policyClass] && !localSettings.Doubled[localClass] {
		t.Fatalf("%s is doubled by organisation policy as %s but reviewed once under local %s; weaker setting was not refused", path, policyClass, localClass)
	}
}
