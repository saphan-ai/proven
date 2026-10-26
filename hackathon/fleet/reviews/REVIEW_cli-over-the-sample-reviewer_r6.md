# REVIEW — cli-over-the-sample-reviewer · round 6

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 4 OWED · 2 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T07:59:33Z

## Summary

The round-12 missing-baseline refusal and owner-file restoration are addressed. Two existing CLI edges remain owed; mutation controls were not run.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground
Reviewed `26343c2...09a52cd`; HEAD measured with `git rev-parse --short=7 HEAD`: `09a52cd`. The contract is ADR-0020 §§1–2, the kickoff brief, and the generated channel grammar. `git diff --name-only 26343c2...09a52cd -- AGENTS.md internal/review go.mod go.sum` returned empty. No dependency manifest changed. The author's status was read as a claim, not as verification.

### Differential
ADDRESSED: `internal/cli/run.go:256-268` refuses `ci` without a baseline before opening the reviewer, at exit 2; `internal/cli/run_test.go:724-751` checks no event and an event with no base SHA. The `561eb8b` diff added the refusal and its test. ADDRESSED: `b11e583` restores `AGENTS.md`; its base-to-tip diff is empty. The three earlier shell repairs (merge base, flag precedence, panic exit 2) remain represented in `internal/cli/run.go:29-82,263-275` and their passing tests; the original earlier review report was not supplied, so its exact three conditions cannot be independently matched to their original wording. The additional `rule` line in `internal/cli/format.go:124-127` preserves a distinct rule ID where a finding already has a separate ID; this is evidence, not a lost field.

### Unplug
The new no-context test at `internal/cli/run_test.go:724` asserts both exit 2 and that the reviewer never opened; without the guard at `internal/cli/run.go:256` the call would continue with an empty baseline. I did not remove that guard and rerun the test in this review-only checkout. `go test -count=1 ./internal/cli ./cmd/proven` passed both packages; `go test -count=1 ./...` passed all packages and `go vet ./...` exited 0. `git diff --check 26343c2...09a52cd` exited 0. These commands test the positive path; a passing suite alone does not prove that every flag's mutant is detected.

### Closure
Recommendation: accept-with-conditions, 0 BLOCKING, 2 OWED, 1 NOTE. Both round-12 changes are addressed. The OWED items below predate that fix diff and therefore do not open a new repair round under re-review rule R14. No owner gate or merge was performed.

### Remainder
OWED — `internal/cli/run.go:263-265`, `internal/cli/run_test.go:769-773`: when `ci --baseline-commit=from-flag` has no event and no `GITHUB_SHA`, it proceeds with `HeadCommit == ""` and exit 0; `internal/review/review.go:61` defines an empty head as the working tree. A CI invocation can therefore review a different checkout state from the requested PR head without a refusal. Control: require a resolved head in `ci`, including explicit-baseline mode, or explicitly constrain and document that mode; test the empty-head case and a valid event head. Trigger and caller are established by the existing explicit-baseline test, not by an end-to-end CI reproduction.

OWED — `internal/cli/format.go:20-27`, `internal/cli/run.go:111-126`: ADR-0020 §1 offers `--json`, `--sarif` and output files, but both writers refuse with exit 2. Their flag-resolution tests do not prove a usable format. Control: implement writers and output assertions in the owning format change, or narrow the public contract explicitly by owner decision. The help text currently discloses the refusal, so this is not a silent success. This predates round 12.

NOTE — `internal/cli/config.go:321` resolves `SAMPLE_MODE` and other non-secret CLI settings before the sample reviewer opens. The review did not interpret the brief's “reads no environment key” literally as forbidding the documented sample-mode variable itself; the no-secret/no-network sample claim needs its own instrument, not the passing suite alone.

### Profile
Configuration and CI inputs are trust boundaries. A malformed event is refused (`internal/cli/run_test.go:752-767`), and git merge-base failures are mapped to exit 2 (`internal/cli/run.go:269-275`). No dependency changes, deleted tests, skip markers, or altered quality thresholds were found in the CLI diff. No declared numeric coverage or mutation baseline was found in the provided order; do not infer one from the package pass count. Declined to judge: the Action, licensing, `internal/review` internals, actual service/network traffic, and unpublished independent review.

### Focus
The round-12 regression, old shell checks, ADR flag precedence, four documented exit classes, error messages listing allowed values, sample output, and event base/head selection were inspected. `TestCICommitsComeFromTheFixtureEvent`, `TestCIFlagAndSHAOverrideTheEvent`, and `TestCIMissingContextExits2` exercise the CI routes. The no-baseline guard has the false-world answer exit 2 without opening the reviewer; this does not establish that every other CI path has the same safeguard.

### Conditions
1. OWED: require or deliberately specify the CI head for an explicit baseline; `internal/cli/run.go:263-265`.
2. OWED: implement JSON/SARIF output or obtain an owner decision narrowing ADR-0020; `internal/cli/format.go:20-27`.
3. NOTE: add a no-secret/no-network sample-mode instrument, distinguishing documented CLI variables from adapter secrets.

### Unrun
Mutation/revert runs for each flag and variable were not performed: this review leg must not alter the code under review. The prior review's full three-condition report was unavailable in this checkout. An end-to-end GitHub Action, live service traffic, and an actual runner were not exercised; this review concerns the CLI diff only. The master's outside-repository sample invocation and trace event count were not reproduced here.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The round-12 missing-baseline refusal and owner-file restoration are addressed. Two existing CLI edges remain owed; mutation controls were not run.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T07:54:36Z

## REDACTED

none
