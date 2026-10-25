package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestOSVIDViewSettingReachesTheAdvisoryQuery proves the loaded name is the one
// the osv-by-advisory template reads, and that an empty setting (sample mode)
// still reads the default view.
func TestOSVIDViewSettingReachesTheAdvisoryQuery(t *testing.T) {
	req := querytpl.Request{Template: querytpl.OSVByAdvisory, Params: map[string]string{"advisory_id": "GHSA-px-sample-0001"}}
	for _, c := range []struct{ set, from string }{
		{"", " FROM " + querytpl.DefaultOSVIDView + " "},
		{"px_usl.px_advisory", " FROM px_usl.px_advisory "},
	} {
		p := &Pipeline{osvIDView: c.set}
		sql, err := querytpl.Build(req, p.relations(context.Background()))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(sql, c.from) {
			t.Errorf("setting %q: %s", c.set, sql)
		}
	}
}
