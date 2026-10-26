# REVIEW — cli-over-the-sample-reviewer · round 4

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T14:58:41Z

## Summary

BLOCKING: four new CI tests fail during fixture creation under the active commit-message hook; the full Go suite is red. Use conforming fixture commit subjects, then rerun the full suite and merge-base control.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground
Reviewed `2a1ea87...6f30e14` (seven commits). First `git rev-parse --short=7 HEAD` printed `6f30e14`. Contract: ADR-0020 §§1–2 and the order. Go module pins 1.26; measured seat version is go1.26.5 linux/amd64 (`go version`). No module or lockfile change. No declared coverage, mutation, lint, or type metric/baseline was found in the order.

RED HEADER — BLOCKING: `internal/cli/run_test.go:742-746`, `:755`, `:765` create fixture commits named `root`, `advance base`, `feature work`, and `only`. The inherited commit-message hook refuses these subjects. Measured `go test -count=1 ./...`: `internal/cli` FAIL; four tests (`TestCIBaseRefWhenTheEventHasNoSHA`, `TestCIWorkspaceIsTheTarget`, `TestCIMergeBaseIsTheForkPoint`, `TestCIUnresolvableBaseExits2`) fail before their assertions. The hook reports `git-rule-subject-pattern` for `root` and `only`. The other listed packages pass. This is a runner/fixture compatibility defect, not proof that merge-base behavior itself is wrong. Remedy: give fixture commits subjects conforming to the active contract while preserving the graph, and rerun the full suite without disabling the hook. These tests are added by this diff.

### Differential
`internal/cli/config.go:229-329` resolves flags before variables; targeted precedence tests pass. `internal/cli/run.go:224-269` computes the default CI baseline using the event and merge base; an explicit baseline is kept as supplied. `internal/cli/run.go:130-163` recovers reviewer and opener panics into exit 2. `internal/cli/config.go:468-481` bounds seconds before converting to `time.Duration`. `internal/cli/format.go:46-158` prints finding evidence, failed services, unresolved and same-family lines. The CLI and tests replace the former `cmd/proven` shell; `internal/review` is unchanged.

### Unplug
`TestPrecedence` checks source ordering; `TestTimeoutDurationBound` checks the value immediately past the safe bound for both flag and variable; `TestOpenerPanicExits2` and `TestReviewerPanicExits2` check error recovery. `TestCIMergeBaseIsTheForkPoint` is intended to fail if the base tip replaces the fork point, but cannot reach its assertion under the measured hook. Actual source reversion/mutation was UNRUN: a reviewer must not modify product code. The suite's red outcome demonstrates a fixture-path alarm only.

### Closure
Earlier shell findings appear addressed: default CI merge base (`run.go:243-269`), winning-source-before-validation (`config.go:354-494`), and opener panic recovery (`run.go:130-137`). Later timeout overflow control is present (`config.go:468-481`). Measured: `go vet ./...` PASS; `git diff --check 2a1ea87...6f30e14` PASS; `go test -count=1 -run 'TestPrecedence|TestTimeout|TestSample|TestRun|TestConfig|TestUnknown|TestCICommits|TestCIFlag|TestJSONFormat' ./internal/cli` PASS; `go test -count=1 ./...` FAIL on the four named fixture commits. A passing subset is not a passing suite.

### Remainder
No other blocking product defect verified. I did not modify owner-owned `AGENTS.md`, although prose changes to it are part of the reviewed diff. Declined to judge the future Action, JSON/SARIF writers, licenses and live services as outside this shell scope. The check credibility ranking is inherited external hook, project `go vet`/diff check, and self-owned Go suite. No declared numeric quality bar has a base/tip value or delta; no exception is proposed.

### Profile
Configuration and CLI trust-boundary profile. Sample passes `Sample: true` to `review.Open` (`run.go:139-153`); no new external dependency, network call within the sample CLI, or credential logging appears in changed files. The CI merge-base subprocess passes the selected target and refs as separate arguments (`run.go:272-290`) and maps its failure to exit 2. `TestFailedServiceIsNamedAndDoesNotExit2`, opener-panic, reviewer-panic and explicit-CI-override tests passed in a separate targeted run; no actual live endpoint was contacted.

### Focus
`TestPrecedence` enumerates ADR-0020 flags, variables and defaults; unknown flags enumerate known flags (`config.go:229-239`), and invalid severity enumerates INFO, WARNING, ERROR (`config.go:449-456`). Sample exit 0/1 and evidence tests pass in `cmd/proven`; exit 7 timeout and config refusal checks pass in targeted `internal/cli` tests. Fixture-based CI behavior remains unverified because fixture setup fails. JSON/SARIF writers explicitly refuse until the formats wagon, as scoped.

### Conditions
Reject until four Git fixture tests use accepted subjects and `go test -count=1 ./...` passes with the active hook. Keep assertions contrasting fork point, base tip and head; rerun `TestCIMergeBaseIsTheForkPoint` as the negative control. Do not suppress the hook.

### Unrun
The four blocked CI tests never reached their assertions; executable source mutations, base-versus-tip numeric metrics, and live network checks were not run. Reasons: the hook refuses fixture commits; reviewer cannot change code; no declared numerical baseline or live-service setup. A nonempty Unrun heading precludes an accept verdict.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: four new CI tests fail during fixture creation under the active commit-message hook; the full Go suite is red. Use conforming fixture commit subjects, then rerun the full suite and merge-base control.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T14:56:00Z

## REDACTED

none
