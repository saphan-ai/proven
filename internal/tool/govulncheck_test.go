package tool

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/writer"
	"github.com/saphan-ai/proven/internal/candidate"
)

// TestGovulncheckMixedLabelsOneReachedAndOneUnreached proves the recorded
// scenario: one vulnerable function is reached, with the call path, and one
// is imported but not reached. A Packagist finding is not measured. Nothing
// is dropped. Reached ranks first (ADR-0043).
func TestGovulncheckMixedLabelsOneReachedAndOneUnreached(t *testing.T) {
	out, err := NewSample(Govulncheck{DB: "file:///checkout/vulndb"}, GovulncheckMixed()).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !out.Answered || !out.PresentOnChange || out.PresentOnBase || out.Version != "v1.1.4" {
		t.Fatalf("%#v", out)
	}
	if out.Command != govulnCommand {
		t.Fatalf("command %q", out.Command)
	}
	syms := Symbols(out.ChangeOutput)
	if len(syms) != 2 {
		t.Fatalf("%#v", syms)
	}
	in := []candidate.Candidate{
		depFinding("example.com/quiet", "the quiet module is in range"),
		depFinding("example.com/reached", "the reached module is in range"),
		{Risk: candidate.Risk{ID: "osv:monolog/monolog", Severity: "ERROR", Message: "php"}, Ecosystem: "Packagist"},
	}
	// The quiet finding is listed first. Ranking has to move the reached one ahead.
	labeled := LabelFindings(in, true, syms)
	if len(labeled) != 3 {
		t.Fatalf("dropped: %#v", labeled)
	}
	byID := map[string]candidate.Candidate{}
	for _, c := range labeled {
		byID[c.ID] = c
	}
	reached := byID["osv:example.com/reached"]
	if reached.ReachLabel != candidate.ReachReached || reached.CallPath == "" {
		t.Fatalf("%#v", reached)
	}
	if !strings.Contains(reached.CallPath, "example.com/app.main") || !strings.Contains(reached.CallPath, "example.com/reached.Bad") {
		t.Fatalf("call path %q", reached.CallPath)
	}
	if !strings.Contains(reached.Message, "Reachability: reached") || !strings.Contains(reached.Message, reached.CallPath) {
		t.Fatalf("message %s", reached.Message)
	}
	quiet := byID["osv:example.com/quiet"]
	if quiet.ReachLabel != candidate.ReachImported || quiet.CallPath != "" {
		t.Fatalf("%#v", quiet)
	}
	if !strings.Contains(quiet.Message, "the quiet module is in range") || !strings.Contains(quiet.Message, "Reachability: imported but not reached") {
		t.Fatalf("message %s", quiet.Message)
	}
	php := byID["osv:monolog/monolog"]
	if php.ReachLabel != candidate.ReachNotMeasured || strings.Contains(php.Message, "imported but not reached") {
		t.Fatalf("%#v", php)
	}
	got := idsOf(writer.Write(writer.Input{Candidates: labeled}).Findings)
	if got != "osv:example.com/reached,osv:example.com/quiet,osv:monolog/monolog" {
		t.Fatalf("order %s", got)
	}
}

// TestUnreachedFunctionIsNotLabeledReached proves a package-only trace is
// not a call. The field this guards is Symbol.Label.
func TestUnreachedFunctionIsNotLabeledReached(t *testing.T) {
	syms := Symbols(govulnChangeOutput)
	var quiet Symbol
	found := false
	for _, sym := range syms {
		if sym.Module == "example.com/quiet" {
			quiet = sym
			found = true
		}
	}
	if !found {
		t.Fatal("the unreached function is missing")
	}
	if quiet.Label != candidate.ReachImported || quiet.Label == candidate.ReachReached || quiet.CallPath != "" {
		t.Fatalf("Symbol.Label = %q path %q", quiet.Label, quiet.CallPath)
	}
}

// TestSilenceIsNotMeasured proves a tool that answered and did not mention
// the module does not say the function is unreached. Not measured is the
// honest label (ADR-0043).
func TestSilenceIsNotMeasured(t *testing.T) {
	in := []candidate.Candidate{depFinding("example.com/other", "in range")}
	got := LabelFindings(in, true, Symbols(govulnChangeOutput))
	if got[0].ReachLabel != candidate.ReachNotMeasured || strings.Contains(got[0].Message, "imported but not reached") {
		t.Fatalf("%#v", got[0])
	}
}

// TestAMissedRunIsNotMeasured proves a tool that did not answer labels every
// Go finding not measured, never imported but not reached.
func TestAMissedRunIsNotMeasured(t *testing.T) {
	in := []candidate.Candidate{depFinding("example.com/reached", "in range")}
	got := LabelFindings(in, false, Symbols(govulnChangeOutput))
	if got[0].ReachLabel != candidate.ReachNotMeasured || got[0].CallPath != "" {
		t.Fatalf("%#v", got[0])
	}
	if strings.Contains(got[0].Message, "imported but not reached") || strings.Contains(got[0].Message, "Reachability: reached") {
		t.Fatalf("message %s", got[0].Message)
	}
}

// TestGovulncheckArgvIsJSONAndAFileDatabase proves the command is govulncheck
// -json with the local database and the whole module, and that a missing
// database or a network address starts nothing.
func TestGovulncheckArgvIsJSONAndAFileDatabase(t *testing.T) {
	argv, reason := (Govulncheck{DB: "file:///checkout/vulndb"}).Argv([]string{"go.mod"})
	if reason != "" || strings.Join(argv, " ") != govulnCommand {
		t.Fatalf("%q %q", argv, reason)
	}
	if _, reason := (Govulncheck{}).Argv([]string{"app.go"}); reason == "" || !strings.Contains(reason, "no local vulnerability database") {
		t.Fatalf("%q", reason)
	}
	if _, reason := (Govulncheck{DB: "https://vuln.go.dev"}).Argv([]string{"app.go"}); reason == "" {
		t.Fatal("a network database was accepted")
	}
	if (Govulncheck{}).Applies([]string{"composer.lock", "README.md"}) {
		t.Fatal("a non-Go change applies")
	}
	if !(Govulncheck{}).Applies([]string{"app/go.mod"}) || !(Govulncheck{}).Applies([]string{"a.go"}) {
		t.Fatal("a Go change does not apply")
	}
	if (Govulncheck{}).Applies([]string{"/tmp/a.go"}) {
		t.Fatal("an absolute path applies")
	}
}

// TestAFailedStartIsNotAScan proves an error line is not a measurement.
// The field this guards is the config object's scanner_name.
func TestAFailedStartIsNotAScan(t *testing.T) {
	if !ScanStarted(govulnChangeOutput) {
		t.Fatal("the recorded scan did not start")
	}
	if ScanStarted("govulncheck: no such file or directory\n") {
		t.Fatal("a missing binary was a scan")
	}
	if ScanStarted("{\"config\":{\"scanner_name\":\"other\"}}\n") {
		t.Fatal("another scanner was govulncheck")
	}
}

// TestGovulncheckDoesNotReadABannerThatIsNotJSON proves a race banner, or a
// line that is not a JSON object, is not a reached function.
func TestGovulncheckDoesNotReadABannerThatIsNotJSON(t *testing.T) {
	if (Govulncheck{}).Read("WARNING: DATA RACE\n").Present {
		t.Fatal("a race banner was a reached function")
	}
	if len(Symbols("progress: scanning\n{not json}\n")) != 0 {
		t.Fatal("a broken line became a symbol")
	}
}

// depFinding is one Go dependency finding, the shape the OSV check builds.
func depFinding(module, message string) candidate.Candidate {
	return candidate.Candidate{
		Risk:       candidate.Risk{ID: "osv:" + module, Severity: "ERROR", Message: message},
		Ecosystem:  "Go",
		ReachLabel: "",
	}
}

// idsOf lists finding ids in order.
func idsOf(in []candidate.Candidate) string {
	parts := make([]string, len(in))
	for i, c := range in {
		parts[i] = c.ID
	}
	return strings.Join(parts, ",")
}
