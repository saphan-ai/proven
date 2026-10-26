# REVIEW — cli-over-the-sample-reviewer · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T04:19:59Z

## Summary

Three shell repairs are addressed; CI silently misses the workspace's default config and independent mutation remains unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground
RED HEADER — `proven ci` can silently ignore the checkout's default config. `internal/cli/config.go:284` calls `resolveConfig` while the target is still `.`; `internal/cli/run.go:42` calls `applyCI`, which only then changes target to `GITHUB_WORKSPACE` (`internal/cli/run.go:220`). Trigger: invoke ci outside the checkout with `GITHUB_WORKSPACE` pointing to a checkout containing `proven.yaml`, without explicit config. The opener receives an empty ConfigPath and the review silently uses defaults. Shell-owned remedy: resolve the workspace before looking for default config; add a test that checks the opener receives its config path. `TestCIWorkspaceIsTheTarget` checks the request target, not ConfigPath. Under R14 this newly noticed shell issue is OWED outside the three repairs; it must not silently be treated as a correct default.

### Differential
Compared `cca7e5c..ae34b74` and full `0ed349b...ae34b74` against ADR-0020 §1–2. Prior finding 1 ADDRESSED: event baseline's merge base computed in `internal/cli/run.go:237-266`; `TestCIMergeBaseIsTheForkPoint` asserts fork rather than base tip. Prior finding 2 ADDRESSED: `internal/cli/config.go:344-459` chooses a flag before parsing its variable; `TestValidFlagBeatsInvalidVariable` and `TestInvalidVariableAloneExits7` cover both states. Prior finding 3 ADDRESSED: `internal/cli/run.go:125-132` recovers opening panic; `TestOpenerPanicExits2` checks exit 2. No dependency manifest or lockfile changed.

### Unplug
Independent fix reversions and mutations of every branch were not run. The executor/master's reported red tests on reverting fixes are not this review's measurement. The new workspace/config mismatch has no test that would redden if configuration again resolved against `.`. Negative control owed: from outside a checkout, create `proven.yaml` inside `GITHUB_WORKSPACE`, assert its path reaches the opener, revert target ordering and show that assertion fail.

### Closure
Measured on `ae34b74`: `go test ./internal/cli ./cmd/proven`, `go vet ./...`, `go test ./...`, `go test -count=1 ./...`, `go test -count=1 -race ./internal/cli ./cmd/proven` each exited 0. The first full suite included cached results; the fresh suite did not. `internal/cli/run.go:25-98` implements 0/1/2/7 and preserves named failed services (`TestFailedServiceIsNamedAndDoesNotExit2`). `internal/cli/format.go` prints ID, severity, title, location, message, test/query evidence and rows, and failed services. `TestSampleFixtureExitCodes` exercises the sample through `review.Open`; refusals for unknown flags and invalid severity list allowed values. JSON/SARIF intentionally refuse until the formats wagon. Verdict: reject because the shell's default CI config selection is wrong and required independent unplug evidence is unrun; the three older repairs themselves are addressed.

### Remainder
Known reviewer-train seam: `internal/review/open.go` forwards only Target to pipeline and ignores baseline, head, include, exclude; it does not interpret ConfigPath. Shell forwards these through `internal/cli/run.go:66-73` and `internal/cli/run.go:139-146`. This existing reviewer gap is not assigned to shell. Further OWED shell concern: `--timeout` context starts after `applyCI` and `openReviewer` (`internal/cli/run.go:60-65`), so slow Git resolution or opening exceeds the promised whole-review deadline. No other behaviour beyond the named boundary was judged.

### Profile
Go module declares `go 1.26`; tests and vet ran as stated. Config is a trust boundary, and silent default fallback is observable in the code path. No numeric quality metric, baseline, tolerance or metric-producing command is declared: numeric bar is UNMEASURED. Check ranking: ADR external to implementation; vet/toolchain project check; author's tests suite check. No test skips, removed assertions, new suppressions, or metric exception found in the diff.

### Focus
`TestPrecedence` exercises formats, output files/path, error mode, baseline variable priority, config, include/exclude, severity, timeout and sample; help/version and quiet/verbose also have tests. `ci` reads event base/head and computes merge base when no baseline is supplied. Sample adapter isolation has tests in reviewer/pipeline; CLI reads its own settings by design. CI workspace/default-config ordering is the exception to default precedence.

### Conditions
Repair shell CI target/config ordering; add and unplug its negative-control test. Run independent negative controls for the repaired conditions before unconditional acceptance. Leave the known reviewer seam gap with the reviewer train. No owner command is requested here.

### Unrun
Independent mutation of all new branches and reproduction of the three fix-reversion reds: not run; a passing fresh suite does not prove a negative control. Base/tip numeric testing-bar delta: no declared metric/baseline/command to measure. Live adapters, Action and formats wagons: out of review scope.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Three shell repairs are addressed; CI silently misses the workspace's default config and independent mutation remains unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T04:16:13Z

## REDACTED

none
