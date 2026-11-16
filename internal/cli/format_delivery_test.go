package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/delivery"
	"github.com/saphan-ai/proven/internal/review"
)

func TestDeliveryIsVisibleAndNeverTurnedIntoAProvedFinding(t *testing.T) {
	r := review.Result{Delivery: &delivery.Result{Complete: true, Claims: []delivery.Claim{{Requirement: "<script>alert(1)</script> | @someone", Status: "contradiction", Reason: "Missing behavior", Citations: []delivery.Citation{{Path: "price.py", StartLine: 2}}}}}}
	var out bytes.Buffer
	if err := (jsonFormatter{}).Write(&out, r); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc["delivery"] == nil || len(doc["findings"].([]any)) != 0 || doc["completion_status"] != "complete" {
		t.Fatal(out.String())
	}
	html := htmlPage(r)
	if strings.Contains(html, "<script>alert(1)") || !strings.Contains(html, "PR delivery") {
		t.Fatal("HTML unsafe or missing delivery")
	}
	var md strings.Builder
	mdDelivery(&md, r)
	if strings.Contains(md.String(), "<script>") || !strings.Contains(md.String(), "| Stated requirement |") || !strings.Contains(md.String(), "Contradicted") {
		t.Fatal(md.String())
	}
	if !strings.Contains(mdVerdict(r), "delivery claims contradicted (advisory)") {
		t.Fatal(mdVerdict(r))
	}
	log := sarifFrom(r)
	if log.Runs[0].Properties["delivery"] == nil || len(log.Runs[0].Results) != 0 {
		t.Fatal("delivery became a SARIF finding")
	}
}
