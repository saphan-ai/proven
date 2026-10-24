package main

import (
	"strings"
	"testing"

	"pv006/product"
)

func TestRawTokenEvidenceRejectsRemovedToken(t *testing.T) {
	notice, _ := product.Render(entry+" $wait", 30)
	if _, err := rawTokenEvidence(notice); err != nil {
		t.Fatalf("actual notice lost raw token: %v", err)
	}

	withoutToken := strings.Replace(notice, "$wait", "30 sec", 1)
	if _, err := rawTokenEvidence(withoutToken); err == nil {
		t.Fatal("removing the raw token did not raise an alarm")
	} else {
		t.Logf("negative control alarm: %v", err)
	}
}
