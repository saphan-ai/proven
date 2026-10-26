# REVIEW — org-policy · round 4

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T07:58:22Z

## Summary

The visible encoding condition is closed at 58ac96b. The focused test passes and fails when the prior formatter is overlaid. No weakening appears in b966161..58ac96b.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — org-policy

### Ground

Reviewed HEAD 58ac96b against b966161. The diff changes only `internal/cli/format.go` and adds `internal/cli/policy_visible_test.go`. Go module declares 1.26.0; local `go version` reports go1.26.5. `go test -count=1 ./...` and `go vet ./...` both exited 0. `git diff b966161..HEAD --check` exited 0. No dependency file changed.

### Differential

`internal/cli/format.go:68` now encodes the policy statement. Lines 110, 344, 347, 350 and 354 encode every string field in the new expired and accepted text lines. The two finding line numbers are integers. The diff makes no policy evaluation or Result field change. The prior formatter printed these fields raw; this formatter calls `visible`, whose control handling is at `internal/cli/format.go:443`.

### Unplug

`go test -count=1 -run '^TestPolicyExceptionFromFileCannotClearReviewText$' -v ./internal/cli` passed both accepted and expired subtests. A Go overlay replacing only `internal/cli/format.go` with its b966161 version made both subtests fail at `policy_visible_test.go:21` with `policy text can clear the review`. The real worktree file was not changed.

### Closure

ADDRESSED: the prior BLOCKING text encoding condition. Each dynamic string in the policy statement, accepted finding and expired exception lines reaches `visible`. The test detects loss of encoding through the policy exception reason in both paths. The replacement test distinguishes the repaired and prior formatters.

### Remainder

No finding in the fix diff. The test exercises the reason field from a policy file; the remaining dynamic string fields were checked by reading each formatter argument, not by separate input cases. The reviewed property outside this fix diff retains the earlier review's scope.

### Profile

Security boundary: policy file content and finding fields enter human text output. The fix encodes them before terminal output. No authorization, dependencies, destructive path, or secret handling changed in this diff. The control byte in the test and formatter was inspected as escaped source text.

### Focus

Nothing touched by b966161..58ac96b weakens the effective policy floor, exception requirements, expiry, active count, or Result fields: only text encoding and its regression test changed. The master's stated focused test and full-suite results were reproduced. `master-was-wrong: none`.

### Conditions

none

### Unrun

none

## The review status as filed

## REVIEW — org-policy
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: The visible encoding condition is closed at 58ac96b. The focused test passes and fails when the prior formatter is overlaid. No weakening appears in b966161..58ac96b.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T07:56:28Z

## REDACTED

none
