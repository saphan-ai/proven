package review

import (
	"context"
	"errors"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// TestDownStoreRefusesBeforeItNamesItselfDown proves a store that is down
// returns *querytpl.Refusal for a request that cannot be built, and not
// *service.Down. An unknown template and a CVE id outside its format are
// the two requests. A request that does build is the down store, so the
// refusal is not a blanket error. The mutated field is the BuildForms call
// in downStore.Run. Deleting it turns both refusals into *service.Down and
// this test goes red.
func TestDownStoreRefusesBeforeItNamesItselfDown(t *testing.T) {
	d := downStore{service: "zetaris", reason: "not configured"}
	cases := []querytpl.Request{
		{Template: "not-a-template"},
		{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "not-a-cve"}},
	}
	for _, req := range cases {
		_, err := d.Run(context.Background(), req, querytpl.Relations{})
		var ref *querytpl.Refusal
		if !errors.As(err, &ref) {
			t.Fatalf("%s: %v, want *querytpl.Refusal", req.Template, err)
		}
		if _, ok := service.As(err); ok {
			t.Fatalf("%s was a down store: %v", req.Template, err)
		}
	}
	_, err := d.Run(context.Background(), querytpl.Request{
		Template: querytpl.KEVByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"},
	}, querytpl.Relations{})
	if down, ok := service.As(err); !ok || down.Service != "zetaris" {
		t.Fatalf("a request that builds: %v, want zetaris down", err)
	}
}
