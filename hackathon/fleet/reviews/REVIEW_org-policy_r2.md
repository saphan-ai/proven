# REVIEW — org-policy · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 6 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T00:31:13Z

## Summary

Three blocking regressions reproduced at 1199ee2: local settings narrow doubled review with a second reviewer; JSON and SARIF hide exceptions; accepted findings remain in proof totals. The whole Go suite passes at base and tip, but the new regression probes fail. Repair and rerun the named tests before the owner's gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — org-policy

### Ground
**RED — The policy floor and the finding visibility contract are not enforced across all paths.** Verdict: reject. At `1199ee2`, an explicitly configured second reviewer can review a policy-doubled path once after a local `doubled_classes: off`; JSON and SARIF omit accepted and expired exception metadata; accepted findings remain in proof totals.

Scope: `origin/main...1199ee2` (merge base `fafce65447838bf2d2d1700b97d606c9fcdc72e6`; this checkout has `origin/main` but no local `main`). Search instruments: `git diff --name-only origin/main...HEAD`; `rg -n 'PolicyPath|EnvPolicy|parseLocal|refuseWeaker|withSecond|applyExceptions|Accepted|Expired|ProofCounts|FormatterFor' internal/review internal/cli`; `rg -n 'doubled_classes|accepted|expired|json|sarif' docs/POLICY.md internal/review internal/cli`; `rg -n 'loadSettings|TopLevel|refuseInside' internal/change internal/review`; `git diff origin/main...HEAD -- go.mod go.sum` (no dependency change). Inspected all 16 changed paths, their policy/config/CLI callers, and the untouched JSON and SARIF writers. The diff changes `internal/change/between.go`, `internal/review/{config,open,policy,review}.go`, `internal/cli/{config,format,run}.go`, CLI and review tests, configuration/catalog, and policy docs. No product code changed by this leg.

Measured: Go module `github.com/saphan-ai/proven`, pinned Go 1.26, local `go1.26.5 linux/amd64`. `go test ./internal/review ./internal/cli ./internal/change ./cmd/proven` passed four packages; `go test ./...` passed 47 packages and identified 10 without tests; `go test -C ██████████████████████████████████████ ./...` on `fafce65` passed 47 packages and identified 10 without tests. These suite readings do not measure the absent cases below. `go vet ./internal/review ./internal/cli ./internal/change ./cmd/proven` passed; `git diff --check origin/main...HEAD` passed. Counts are lines prefixed `ok ` and `? ` in the corresponding saved outputs. No base failures were observed.

### Differential
1. **BLOCKING — An effective doubled class can be narrowed.** `internal/review/policy.go:294-302,346-361` checks `doubled_classes` only when the policy explicitly names that key, while `internal/review/open.go:290-306` doubles every class by default with a connected second reviewer unless the repository explicitly sets the list. A policy with `path_rules: go.mod=security` fixes classification, doubles `main.go` as backend under the second reviewer, and accepts a local `doubled_classes: off`, which reviews `main.go` once. `internal/review/policy_doubled_test.go:120-153` skips the sweep exactly when the policy has not explicitly set `doubled_classes`. Measured failing test below: `main.go is doubled by organisation policy as backend but reviewed once under local backend`. Remedy: compare the effective doubled set after `withSecond` for a class-fixing policy even if it did not explicitly set `doubled_classes`, and refuse the local key by name. Add a negative control that mutates only that key.
2. **BLOCKING — Machine-readable reviews hide accepted and expired exceptions.** `internal/review/policy.go:563-576` moves a live exception's finding out of Findings and attaches accepted/expired metadata to Result; `internal/cli/format_json.go:43-62,117-132` serializes only Findings; `internal/cli/format_sarif.go:85-116` serializes only Findings and old run properties. The text writer (`internal/cli/format.go:103-113`) is the only one that prints exceptions. A JSON- or SARIF-only user cannot see the accepted risk or the fact that an exception expired; a live accepted risk disappears completely. Both format subtests below fail. Remedy: represent accepted and expired records, with reason, owner, expiry and associated finding, in JSON and SARIF and test each writer.
3. **BLOCKING — Accepted findings still count in the proof totals.** `internal/review/open.go:461` increments ProofCounts for each finding; `internal/review/policy.go:563-576` moves accepted findings but never removes them from those totals. The review API returns `len(Findings)==0` with `ProofCounts[proved-by-test]==1`, contradicting the stated not-counted result. The isolated test below fails. Remedy: recompute or adjust totals after exception application, maintaining the invariant that totals count active findings.

Full text of independent Go regression tests follows. They were overlaid into their respective packages without editing tracked source. A fix must make all three pass. Exact command: `GOFLAGS='-overlay=███████████████████████████████████████████████████████████████████████████████████████████████████████' go test ./internal/cli ./internal/review -run 'TestPolicyExceptionsVisibleInEveryOutputFormat|TestAcceptedExceptionDoesNotCountInProofTotals|TestLocalDoubledListCannotNarrowEffectivePolicyWithSecondReviewer' -count=1`. At the reviewed SHA, all four subcases fail for the expected behaviour, not a compile error. The test names and logs were observed before filing. The existing `writeFile` helper used by the third test is in `internal/review/policy_test.go`.

`internal/review/policy_second_probe_test.go`:
```go
package review

import "testing"

func TestLocalDoubledListCannotNarrowEffectivePolicyWithSecondReviewer(t *testing.T) {
 p, err := loadPolicy(writeFile(t, "policy.yaml", "path_rules: go.mod=security\n"))
 if err != nil { t.Fatal(err) }
 local, err := p.parseLocal([]byte("doubled_classes: off\n"))
 if err != nil { return }
 policySettings := withSecond(p.settings, "second-model")
 localSettings := withSecond(local, "second-model")
 path := "main.go"
 policyClass, _ := chooseClass([]string{path}, policySettings.PathRules, policySettings.ClassOrder, policySettings.Fallback)
 localClass, _ := chooseClass([]string{path}, localSettings.PathRules, localSettings.ClassOrder, localSettings.Fallback)
 if policySettings.Doubled[policyClass] && !localSettings.Doubled[localClass] {
  t.Fatalf("%s is doubled by organisation policy as %s but reviewed once under local %s; weaker setting was not refused", path, policyClass, localClass)
 }
}
```

`████████████████████████████████████████████`:
```go
package cli

import (
 "bytes"
 "encoding/json"
 "testing"
 "github.com/saphan-ai/proven/internal/review"
)

func TestPolicyExceptionsVisibleInEveryOutputFormat(t *testing.T) {
 finding := review.Finding{ID: "F1", RuleID: "rule-one", Location: review.Location{Path: "a.go", StartLine: 2}}
 result := review.Result{
  PolicyStatement: "organisation policy: policy.yaml",
  Accepted: []review.AcceptedFinding{{Finding: finding, ExceptionID: "approved", Reason: "triaged", Owner: "team", Expires: "2999-12-31"}},
  Expired: []review.ExpiredException{{Finding: "F2", ExceptionID: "old", Rule: "rule-two", Location: "b.go:3", Reason: "triaged", Owner: "team", Expires: "2000-01-01"}},
 }
 for _, format := range []string{"json", "sarif"} {
  t.Run(format, func(t *testing.T) {
   formatter, err := FormatterFor(format)
   if err != nil { t.Fatal(err) }
   var output bytes.Buffer
   if err := formatter.Write(&output, result); err != nil { t.Fatal(err) }
   var document any
   if err := json.Unmarshal(output.Bytes(), &document); err != nil { t.Fatal(err) }
   for _, want := range []string{"approved", "old", "triaged", "team", "2999-12-31", "2000-01-01"} {
    if !bytes.Contains(output.Bytes(), []byte(want)) { t.Errorf("%s output hides policy exception detail %q: %s", format, want, output.String()) }
   }
  })
 }
}
```

`internal/review/policy_count_probe_test.go`:
```go
package review

import (
 "testing"
 "time"
)

func TestAcceptedExceptionDoesNotCountInProofTotals(t *testing.T) {
 finding := Finding{ID: "F1", RuleID: "rule-one", Location: Location{Path: "a.go", StartLine: 2}}
 level := ProvedByTest
 result := Result{Findings: []Finding{finding}, ProofCounts: map[ProofLevel]int{level: 1}}
 policy := &policy{exceptions: []exception{{ID: "approved", Rule: "rule-one", Path: "a.go", Reason: "triaged", Owner: "team", Expires: "2999-12-31"}}}
 policy.applyExceptions(&result, time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC))
 if len(result.Accepted) != 1 || len(result.Findings) != 0 { t.Fatalf("accepted=%d active=%d", len(result.Accepted), len(result.Findings)) }
 if result.ProofCounts[level] != 0 { t.Fatalf("accepted finding still counts in proof totals: %d", result.ProofCounts[level]) }
}
```

### Unplug
Mutation overlays lived entirely under `██████████████████████████████████` and never changed tracked source. Each mutant flips one named branch and each test below turns red: `refuseWeaker` bypass -> `TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne` (weaker values accepted); `applyExceptions` disabled -> `TestAnAcceptedFindingIsShownWithItsOwnerAndDoesNotCount` (F1 remains active); `refuseInside` disabled -> `TestAPolicyInTheRepositoryOfASubdirectoryTargetIsRefused`; expiry arm disabled -> `TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo`; owner presence check disabled -> `TestAnExceptionWithoutAnOwnerOrAnExpiryIsRefused/owner`; text accepted rendering disabled -> `TestAnAcceptedFindingIsPrintedWithItsOwnerAndDoesNotFailTheRun`. The *missing* alarms are the three new regression tests above, which fail against the unmutated tip. This tests major added mechanisms rather than every source line.

### Closure
All three findings are independently reproduced at `1199ee2`; none is addressed at this SHA. Ruling: JSON and SARIF are included in “shown, never hidden” because both are supported review outputs and a consumer can request only one; if wrong, an owner would be accepting a format-specific loss of visibility. Ruling: ProofCounts is a published Result field and an accepted finding “does not count”; if wrong, API clients may count an accepted risk as active while the CLI does not. Neither ruling weakens a failing check.

### Remainder
No additional finding admitted. The policy file is rejected inside the target repository, including a subdirectory and a link; missing/invalid owner and expiry are refused. A file outside the review target can still be replaced between `Open` and `Review`; this review did not test concurrent external modification. No `go.mod` or `go.sum` changes and no dependency added. No threshold, skip, suppression or deleted test appears in the changed test files. Quality note: the new `internal/review/policy.go` is 585 lines; this is maintainability, not an independent blocker. Declined to judge: licences, networked policy repositories, external policy authorization, and live service availability are outside the order.

### Profile
Security trust boundary: organisation policy input is trusted runner-controlled data; a repository's `proven.yaml` is lower trust; output format is the observation boundary; exceptions are owner-controlled risk decisions. Spoof/tamper: local configuration can narrow the second-review default (finding 1); hide/disclose: machine output drops exception identity (finding 2); aggregate integrity: proof count is stale (finding 3). Other boundaries: policy file inside target refused with resolved roots, exception owner/expiry required. No manifest or lockfile delta; no new install script. Scan of the changed diff found no new `nolint`, test skip or secret literal. Checks ranked: repository policy and docs contract are project checks; test suite and mutation overlays are suite checks; no independent external check declared. No testing-bar baseline, tolerance or mutation-score target is recorded for this change, so no numeric bar was inferred.

### Focus
Can the diff break the requested property? **Yes**, through the three reproduced cases. `master-was-wrong:` the master's 2026-10-24 whole-suite-green reading is true again on October 25 (and main is green), but its implied coverage of every policy path is false; the fresh negative controls fail at the reviewed tip. This is a coverage gap, not a newly failing existing test.

### Conditions
- BLOCKING: local settings must not narrow the effective doubled paths under a connected second reviewer; add the precise regression above and check the effective configuration rather than explicit-key presence.
- BLOCKING: JSON and SARIF must display accepted and expired exceptions, including reason, owner and expiry; add output-specific regression tests.
- BLOCKING: accepted findings must be excluded from ProofCounts; assert that the totals equal active findings after exceptions.
- OWED: retain proof from the existing policy-inside, owner/expiry, and expiration unplug tests on the repaired tip. No class is downgraded here.

### Unrun
`go test -race ./...` and repeat/flaky runs were not run: the change does not introduce concurrency and the ordered property was tested with targeted mutation and ordinary whole-module runs. Live generator, second-review service, and production containers were not run because the review boundary is local/no network; the second-review setting was measured through the deterministic `withSecond` function. No dedicated mutation-score tool or recorded baseline exists; targeted unplug runs above are the measured mutation checks. No customer licence path was evaluated (outside scope).

## The review status as filed

## REVIEW — org-policy
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Three blocking regressions reproduced at 1199ee2: local settings narrow doubled review with a second reviewer; JSON and SARIF hide exceptions; accepted findings remain in proof totals. The whole Go suite passes at base and tip, but the new regression probes fail. Repair and rerun the named tests before the owner's gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T00:25:42Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| file and line in Saphan Studio's private code | 1 |
