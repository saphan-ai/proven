# REVIEW — cli-over-the-sample-reviewer · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T13:01:57Z

## Summary

Reject: an overflowing --timeout produces an expired deadline and incomplete review with exit 0. Full-suite checks are obstructed by the inherited fixture-commit hook.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground
HEAD: a4eae19, measured with git rev-parse --short=7 HEAD. Reviewed 6db4272...a4eae19 against ADR-0020 §§1–2. The diff has ten paths; go.mod, go.sum, and internal/review are unchanged. Local Go: go1.26.5 linux/amd64; the order reports go1.26.8 on the master. git diff --check 6db4272...a4eae19 passed.

### Differential
BLOCKING — internal/cli/config.go:462 and internal/cli/run.go:63. parseTimeout accepts any nonnegative int; multiplication by time.Second overflows above 9,223,372,036 seconds. Measured: go run ./cmd/proven scan --sample --timeout=9223372037 testdata/fixture/discount returned exit 0, printed failed services: none and review partial: the deadline was reached, and printed no findings. This input creates an already-expired context, so an incomplete scan looks successful. The false-world answer for a normal in-range sample scan is two findings, checked in the passing TestSampleFixtureExitCodes. Remedy: refuse out-of-range timeout with exit 7 and supported bound, or safely represent the timeout; test flag and PROVEN_TIMEOUT boundaries. The existing tests cover only small positive values (internal/cli/config_test.go:450; internal/cli/run_test.go:382).
Earlier shell repairs remain in code: opener panic recovery (internal/cli/run.go:130), precedence before lower-priority parsing (internal/cli/config.go:348), and implicit CI merge base (internal/cli/run.go:257). Focused tests passed for the first two; CI merge-base fixture commits were refused by the seat hook before assertions.

### Unplug
The missing negative control is a boundary test that substitutes 9223372037 for an in-range --timeout and PROVEN_TIMEOUT and expects exit 7 instead of partial exit 0; the current code fails. Removing opener recovery should fail TestOpenerPanicExits2. Parsing an overridden invalid variable should fail TestValidFlagBeatsInvalidVariable. Removing merge-base calculation should fail TestCIMergeBaseIsTheForkPoint once fixture commits can run. These mutations were specified, not performed.

### Closure
Reject until overflow is refused or bounded and both input sources have boundary tests. Repeat the full suite and CI merge-base test in an environment that permits temporary fixture commits without bypassing the governing hook. No product code changed and no commit was made by this review leg.

### Remainder
No numeric coverage, mutation, or warning baseline was declared in the order or touched in the diff. go vet ./internal/cli ./cmd/proven passed; vet is a toolchain/project check, while local Go tests are suite-owned and more circular. Coverage and mutation deltas are UNRUN without a declared baseline. No test skip, suppression, dependency delta, or threshold edit was found. Declined to judge the Action, future JSON/SARIF writers, licences, live agents, and internal/review implementation outside this diff.

### Profile
Configuration and user-facing CLI output. Input is accepted from flags, environment and event JSON; the shell calls the reviewer seam. The overflow yields an incomplete yet superficially successful review. Text output contains finding id, severity, title, location, message, test/query evidence and failed services. This is correctness, not a claim of a security exploit.

### Focus
ADR-0020 §1: internal/cli/config_test.go has table cases for the declared flags and variables, defaults, overrides and precedence. §2: focused tests passed for sample exits 0/1, configuration exit 7, reviewer panic exit 2, named degraded service, unknown-flag and severity refusals. The sample injected live opener is not called; no independent network instrument was used. The fixture event names base and head; TestCICommitsComeFromTheFixtureEvent checks an unresolved base and TestCIFlagAndSHAOverrideTheEvent checks overrides. The merge-base behavior remains unverified on this seat.

### Conditions
Validate timeout seconds before conversion to time.Duration; test the largest valid second and the next second for both --timeout and PROVEN_TIMEOUT. Repeat affected fixture-based checks with their commits permitted. No exception proposed.

### Unrun
go test -count=1 ./... ran and failed: ten tests across internal/change, internal/cli and internal/review were stopped before their assertions when the inherited Saphan commit-message hook refused temporary fixture commit subjects (first, root, only, base). This is an environment refusal, not evidence the assertions failed; full-suite green and the CI merge-base test remain UNRUN as behavioral proof. Focused go test -count=1 ./internal/cli -run for precedence, sample, CI explicit overrides, errors, panic, and refusals passed. The hook was not bypassed. Base test counts, race checks, mutation, and instrumented network checks were UNRUN; no additional checkout was made inside the confined worktree.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject: an overflowing --timeout produces an expired deadline and incomplete review with exit 0. Full-suite checks are obstructed by the inherited fixture-commit hook.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T12:58:00Z

## REDACTED

none
