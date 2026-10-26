# REVIEW — cli-over-the-sample-reviewer · round 5

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 3 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T02:51:49Z

## Summary

CI can review the whole checkout when pull-request context is absent. The branch also edits owner-owned AGENTS.md and has no required public copies. Full Go tests and vet passed; the named defects need repair before a gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground

RED — CI silently widens an intended pull-request review to the whole checkout when its baseline is unavailable. Reject.

HEAD measured with `git rev-parse HEAD`: `d488a0f0c5edf247f76464aa342d32d389013a46`. Reviewed `26343c2...d488a0f`, ADR-0020 §§1–2, `AGENTS.md`, and the paired `internal/review` request contract. `go.mod` pins Go 1.26. `go test -count=1 ./...` exited 0 on this seat; `go vet ./...` and `git diff --check 26343c2...d488a0f` exited 0. These results are local measurements, not the master's claims.

### Differential

1. BLOCKING — `internal/cli/run.go:243–257`. Trigger: `proven ci` with no `GITHUB_EVENT_PATH`, `GITHUB_BASE_REF`, or explicit baseline, or with an event lacking a base. `applyCI` returns nil with an empty baseline. `internal/review/review.go:60` defines that value as the whole target, and `internal/review/decide.go:64` only takes the change path when the baseline is nonempty. The command can report success for a review other than the pull-request diff promised by ADR-0020 §1. Remedy: require a usable base and head for `ci`, return exit 2 when either is missing, and test the missing and malformed-event cases. The existing `TestCIUnresolvableBaseExits2` only covers a nonempty, unresolvable base.
2. BLOCKING — `AGENTS.md:67–82` changes an owner-owned file. `AGENTS.md` §3 says never change that file, and this diff adds a command-line chapter to it. Remedy: the owner decides and makes any change to this file; remove this edit from the executor's branch. This is a scope and ownership finding, regardless of the added text's accuracy.
3. BLOCKING — the diff contains no `fleet/status/STATUS_cli-over-the-sample-reviewer_*.md` and no public order copy. `AGENTS.md` §1 requires bridge documents and public copies in the same change. Remedy: the master supplies the public copies from registered bridge originals, with actual model and cost data, before the change lands. Do not invent those values.

The round-six repairs remain present: `TestCIMergeBaseIsTheForkPoint` checks the fork point; `TestValidFlagBeatsInvalidVariable` checks priority; `TestOpenerPanicExits2` checks the panic path. The sample adapter path, text evidence and failed-service lines are covered by `TestSampleFixtureExitCodes`, `TestFailedServiceIsNamedAndDoesNotExit2`, and the `cmd/proven` sample tests. `internal/cli/format.go:144–147` adds a rule line when a finding's rule differs from its ID; this makes the evidence more specific and does not contradict the text contract.

### Unplug

Removing merge-base computation would fail `TestCIMergeBaseIsTheForkPoint`; making a lower-priority malformed variable win would fail `TestValidFlagBeatsInvalidVariable`; dropping opener recovery would fail `TestOpenerPanicExits2`; removing sample evidence would fail `TestSampleFixtureExitCodes`. These are test-to-branch inspections; I did not edit or execute mutants. No test turns red when `ci` receives neither a baseline nor event context, so the CI scope guard has no negative control.

### Closure

Reject until all three blocking findings are addressed. Re-run the full suite and vet on the repaired tip, and demonstrate `ci` exits 2 on absent pull-request context while preserving the fixture merge-base path. The changed `AGENTS.md` and missing public copies require their named remedies before gate.

### Remainder

No additional owed finding. The branch has no declared coverage, mutation, lint-count, or type-warning bar. The suite is an author-controlled check; vet is a project check. No threshold or skip was changed. I declined to judge the separate formats, Action, live reviewer, licensing, and the other independent review.

### Profile

Security/configuration profile: command arguments, environment precedence, CI event input, configuration refusal and output paths were inspected. `gitMergeBase` uses argument-form execution; no shell interpolation is present. The change adds no dependency. The reviewer and live/sample adapters are outside this diff; their request contract was read for the baseline consequence.

### Focus

ADR-0020 flags and variables are represented in `internal/cli/config.go`; `TestPrecedence` covers the principal default and override cases. Codes 0, 1, 2, and 7 and named degraded services have tests in `internal/cli/run_test.go`. Unknown-flag and severity refusals have tests in `internal/cli/config_test.go`. The sample output includes finding identity, severity, title, location, message, evidence and failed services. CI fixture and real temporary-Git tests cover the normal merge-base path. The missing-context case above prevents acceptance of the full CI claim.

### Conditions

Three BLOCKING findings, zero OWED, zero NOTE. No reclassification recommended.

### Unrun

Actual source mutations were not run because this review leg may not edit product code. A standalone `docs/SAMPLE_MODE.md` run outside the repository was not repeated; the full test suite includes an outside-directory sample path. Live network and container review were not run because the order limits this review to the CLI diff. No numeric quality-bar command exists to run at base and tip.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: CI can review the whole checkout when pull-request context is absent. The branch also edits owner-owned AGENTS.md and has no required public copies. Full Go tests and vet passed; the named defects need repair before a gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T02:47:34Z

## REDACTED

none
