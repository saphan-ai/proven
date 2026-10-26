# REVIEW — org-policy · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T07:47:11Z

## Summary

One blocking text-output defect: accepted and expired exception fields can carry terminal controls, so a review can be hidden. A local Go regression test fails at b966161 for both states; the full base and tip suites pass.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — org-policy

### Ground

Reviewed commit `b96616135f3091478f5453671ba9d60106330cd9` against `77a72fc` (`git merge-base 77a72fc HEAD` returned `77a72fc`). The worktree was at the ordered SHA. The Go module pins `go 1.26.0`; the runner used `go1.26.5 linux/amd64`. The Makefile has no test target and no tracked CI workflow was found, so I ran `go test -count=1 ./...` on a Git archive of `77a72fc` and on the reviewed tip. Both completed with exit 0 and no failed packages. `go vet ./...`, `git diff --check 77a72fc...HEAD`, and `gofmt -l` on the changed production Go files were clean at the tip.

Searches used to establish the review surface: `git diff --name-only 77a72fc...HEAD` and `git diff --stat 77a72fc...HEAD` (40 changed files); `rg -n 'organi[sz]ation|exception|accepted|expires|expiry|Policy|Effective' internal cmd --glob '*.go'`; `rg -n 'withSecond|chooseClass|ProofCounts|len\(.*Findings|\.Findings' internal/review internal/cli`; `rg -n '^func Test' internal/review/policy*_test.go internal/cli/policy_test.go`; `git diff 77a72fc...HEAD -- go.mod go.sum package.json package-lock.json` (no root manifest or lockfile change). I read the policy parser, local settings readers, effective class and doubled checks, exception match/count path, CLI text and exit path, related tests, and `docs/POLICY.md`. The diff also contains main's additive code and fixture changes; no separate policy weakening was found in those paths.

The project declares no numeric coverage, mutation, lint, or type baseline with a command. Test-suite checks are the available project-owned bar; there is no external or project metric to rank above them. This is a NOTE, not a measured metric claim. Record a baseline and its command before setting a numeric floor.

remark: The passing suite establishes its named cases, not terminal-safe rendering of every policy value; the separate red test below supplies a false-world answer.

### Differential

The policy is loaded by `review.Open`, inherited by all three local settings reads, and checked before a review when the policy file is inside the target. `parseLocal` compares each supplied key with the effective policy. Class rules, fallback, order, and doubled classes are checked together. Exceptions are parsed only from the policy file, matched by rule/path/optional line, split into active and expired, and active findings are removed from `Findings` and `ProofCounts`. The CLI text writer adds policy, accepted, and expired lines. The CLI exit path reads the remaining `Findings`.

The one changed output path below skips the text writer's existing `visible` control-byte encoding. That makes the new text lines depend on terminal interpretation of policy data.

### Unplug

Each mutation was made only in a Go overlay under `██████████████████████████████████`; no product file was edited. All commands used `go test -overlay ... -count=1` with the named test. The false-world answers were actual failures:

| Mechanism unplugged | Test that turned red | Observed result |
|---|---|---|
| Local policy floor (`parseLocal` passed an empty `given` map) | `TestAWeakerLocalRuleIsRefusedByName` | Wanted `*ConfigError`, got nil. |
| Doubled path guard (`leavesDoubled` forced false) | `TestADoubledPathIsRefusedByName` | Raised fallback and raised rule were accepted. |
| Required exception fields (empty-field guard forced false) | `TestAnExceptionWithoutAnOwnerOrAnExpiryIsRefused` | Missing rule, reason, and owner were accepted. |
| Expiry transition (expired branch forced false) | `TestAnExpiredExceptionNoLongerAcceptsAndTheReviewSaysSo` | Three active findings and one accepted finding instead of four active. |
| Policy outside target guard (containment branch forced false) | `TestAPolicyInTheRepositoryIsRefusedByAReview` | Expected named policy refusal; execution reached a later missing-base error. |
| Active count reset (zeroing loop skipped) | `TestAcceptedExceptionDoesNotCountInProofTotals` | Accepted finding still counted as one. |
| Accepted text output (accepted loop emptied) | `TestAnAcceptedFindingIsPrintedWithItsOwnerAndDoesNotFailTheRun` | Accepted finding and owner lines absent. |

The new policy text output has no terminal-control negative control. `TestPolicyExceptionFromFileCannotClearReviewText` below fails on the unmodified tip in both active and expired cases.

### Closure

**BLOCKING — `internal/cli/format.go:68`, `:109`, `:344`, `:347`, `:350`, `:353`: policy and finding fields reach text output without `visible`.** The existing `visible` helper at line 443 encodes control bytes, and ordinary finding fields use it. The new policy statement, accepted finding, and expired exception lines print their fields directly. A policy `reason` containing an ANSI clear-screen control is accepted by the parser and appears raw in both active and expired text output. A terminal can erase or rewrite the review the user sees, so the promise that accepted and expired findings are shown is not met. The accepted finding title and location and policy path have the same gap. Encode every dynamic field in these new text lines with `visible`, then make this test pass. Keep the separate output labels and the active/expired counts.

Full regression test (place beside `internal/cli/policy_test.go`; it uses its existing `policyException` and `runWithPolicy` helpers):

```go
package cli

import (
    "strings"
    "testing"
)

func TestPolicyExceptionFromFileCannotClearReviewText(t *testing.T) {
    for _, expires := range []string{"2999-12-31", "2000-01-01"} {
        t.Run(expires, func(t *testing.T) {
            body := policyException("a", "discount-applied-twice", "billing/discount.go", expires)
            body = strings.Replace(body, "reviewed by the platform team", "reviewed\x1b[2Jhidden", 1)
            _, stdout, stderr := runWithPolicy(t, body)
            if stderr != "" {
                t.Fatalf("policy was refused: %s", stderr)
            }
            if strings.Contains(stdout, "\x1b") {
                t.Fatalf("policy text can clear the review")
            }
            if !strings.Contains(stdout, `\x1b`) {
                t.Fatalf("control byte was not shown visibly")
            }
        })
    }
}
```

Command: `go test -overlay ██████████████████████████████████████████████ -count=1 ./internal/cli -run '^TestPolicyExceptionFromFileCannotClearReviewText$' -v`. Result at `b966161`: both `2999-12-31` and `2000-01-01` subtests failed with `policy text can clear the review`. The test executes the real sample CLI path; it did not need a network service.

### Remainder

No other blocking or owed finding was established. The master's reported full-suite pass held at this SHA; the master was wrong only if that pass was read as proof that policy text cannot be hidden by control bytes. No product code was changed. A direct check of the root Go dependency manifests found no change; the new package lock is in a corpus fixture. No performance claim was made.

Declined to judge by the owner's explicit scope: JSON, SARIF, and PR-comment display of accepted/expired exceptions; W2b owns those surfaces. Their result fields were read here, but this verdict does not claim those writers show them. Licences are also out of scope.

### Profile

Security trust boundaries read: organisation policy file to parser; repository `proven.yaml` to local settings; policy path versus target repository; exception fields to finding match and count; `Result` to terminal text. The local configuration cannot supply an exception. The policy containment check resolves links. The terminal is the failing boundary: control bytes from a policy file are trusted as display instructions. No root Go dependency delta or destructive derived-path operation was in this policy change. The changed code and inputs were searched for secret exposure and instruction-like comments; no separate finding was established.

### Focus

Can anything the diff touches break the required property? Yes. The text output can fail to show accepted or expired findings reliably because policy data can control the terminal. The effective-class and exception-count mechanisms resisted the concrete unplug checks above. `master-was-wrong:` the master's 51-package pass (reported in the order) did not cover this terminal-control case; my fresh suite also passed before the added regression test.

### Conditions

Verdict: `reject`. Fix the one blocking text-encoding defect and run the regression test on the fixed tip, then run `go test -count=1 ./...` and `go vet ./...`. A later review should check the same output boundary; no reclassification is requested.

### Unrun

Race detector: not run because this diff adds no shared-state or goroutine behavior. Numeric coverage, mutation, lint, and type metrics: no declared baseline or command exists in this repository. JSON, SARIF, and PR-comment exception display: excluded by the owner's scope ruling. Live generator and external services: not run; the order confined this review to local tests.

## The review status as filed

## REVIEW — org-policy
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: One blocking text-output defect: accepted and expired exception fields can carry terminal controls, so a review can be hidden. A local Go regression test fails at b966161 for both states; the full base and tip suites pass.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T07:44:41Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
