package cli

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/settingspage"
)

// A live (non-sample) run records every documented setting. Each secret
// setting must be recorded as "set" or "unset", and no value of one may appear,
// even a value too short for the redactor. The mutated fields are the values
// of the secret settings.
func TestALiveRunRecordsNoSecretSettingValue(t *testing.T) {
	env := map[string]string{}
	for _, sec := range settingspage.Catalog() {
		for _, row := range sec.Rows {
			if row.Secret {
				env[row.Name] = ""
			}
		}
	}
	for _, name := range redact.Names {
		env[name] = ""
	}
	i := 0
	for name := range env {
		env[name] = "k3y" + string(rune('A'+i)) // short: under the redactor's minimum
		i++
	}
	getenv := func(k string) string { return env[k] }
	got := packSettings(config{}, getenv, redact.FromEnv(getenv))
	for name, v := range env {
		if got[name] != "set" {
			t.Errorf("%s is recorded as %q, want set", name, got[name])
		}
		for k, rec := range got {
			if strings.Contains(rec, v) {
				t.Errorf("setting %s holds the value of %s", k, name)
			}
		}
	}
}

// Every string of a verified report comes from the signer's pack, or from the
// pack's path, and reaches the terminal through visible. The mutated fields are
// each string field of the report and the path, one hostile value in all.
func TestEveryReportFieldIsShownEscaped(t *testing.T) {
	h := "x\x1b[2J" + string(rune(0x202e)) + "\x9by"
	rep := &evidence.Report{
		Created: h, KeyID: h,
		Commits:     evidence.Commits{Target: h, Baseline: h, BaselineResolved: h, Head: h, HeadResolved: h},
		Models:      []evidence.Model{{Agent: h, Route: h, Model: h, ReceiptID: h}},
		ProofCounts: map[string]int{h: 1},
		FailedSvcs:  []string{h},
	}
	out := renderReport(h, rep)
	assertNoControl(t, out)
	if !strings.Contains(out, `x\x1b[2J`) {
		t.Fatalf("the hostile value is not shown escaped: %q", out)
	}
}
