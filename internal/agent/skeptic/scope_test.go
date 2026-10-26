package skeptic

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestPromptListsOnlyTheScopesTemplates proves the Skeptic is told of the
// templates its scope allows and of no other, and that the nil scope tells it
// of every template. Mutated field: the scope.
func TestPromptListsOnlyTheScopesTemplates(t *testing.T) {
	narrow := Prompt("F1", "title", querytpl.Scope{querytpl.KEVByCVE})
	if !strings.Contains(narrow, querytpl.KEVByCVE+" (") {
		t.Fatal("the allowed template is not in the prompt")
	}
	if strings.Contains(narrow, querytpl.EPSSByCVE) || strings.Contains(narrow, querytpl.OSVByPackage) {
		t.Fatalf("the prompt names a template outside the scope:\n%s", narrow)
	}
	whole := Prompt("F1", "title", nil)
	for _, tpl := range querytpl.Catalogue() {
		if !strings.Contains(whole, tpl.Name+" (") {
			t.Errorf("the unscoped prompt lacks %s", tpl.Name)
		}
	}
}
