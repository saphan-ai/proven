# REVIEW — cli-over-the-sample-reviewer · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 4 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T03:07:55Z

## Summary

Reject: CI and scan silently ignore the baseline and path filters; overridden environment values can cause refusals. Repair and test before the gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground
Reviewed `0ed349b...cca7e5c` at `cca7e5c` against ADR-0020 §§1–2 and the supplied six conditions. `go test ./internal/cli ./cmd/proven ./internal/review ./internal/pipeline`, `go test ./...`, and `go vet ./...` passed on 2026-10-22. No dependencies or lockfiles changed. No product code was changed in this review.

### Differential
**BLOCKING — selection flags and CI commits have no effect on the real reviewer.** `internal/cli/run.go:64-71` populates `BaselineCommit`, `HeadCommit`, `Include`, and `Exclude`; `internal/review/open.go:35-40` discards all four and passes only `Target` to the pipeline. `internal/cli/run_test.go:100-119` checks a capture reviewer rather than the production opener. In a confined work directory, the built binary `proven scan --sample --exclude='*' ../../../testdata/fixture/discount` exited 0 and printed F1 and F2, as did `--baseline-commit=does-not-exist`. This silently defeats changed-only CI and explicit path selection. Remedy: make the change and path selection operative at the shell/reviewer boundary, and test it with the actual opener and a divergent Git fixture. The separate reviewer train's seam cannot be assumed to do this simply because the shell populated the request.

**BLOCKING — no CI merge base.** `internal/cli/run.go:227-237` copies the event's base SHA or `GITHUB_BASE_REF` verbatim, rather than computing its merge base with the selected head. `internal/cli/run_test.go:100-119` asserts the same copied SHA. An advanced base branch changes the desired comparison. Remedy: compute and assert the merge base with an advanced-base fixture, after agreeing on ownership of Git selection with the reviewer train.

**BLOCKING — invalid lower-priority variables override valid flags by refusal.** `internal/cli/config.go:342-359`, `:372-387`, `:437-453`, `:465-481`, and `:492-504` parse an environment setting before selecting its overriding flag. Measured: `PROVEN_SEVERITY=invalid proven scan --sample --severity=INFO ../../../testdata/fixture/discount` exited 7 instead of reviewing; the error named `PROVEN_SEVERITY`. The same ordering affects output format, error, timeout, and sample mode. Remedy: choose the winning source before parsing it; test each flag against an invalid lower-priority variable. Existing precedence cases use only valid lower-priority values.

**BLOCKING — a present configuration file is never loaded.** `internal/cli/config.go:405-427` checks only existence; `internal/cli/run.go:128-135` sets `review.Options.ConfigPath`, but `internal/review/open.go:10-27` never reads it. A present invalid file cannot produce the required load-time refusal, and a valid file cannot set the review's default timeout. Remedy: connect actual validation/configuration to the reviewer and test invalid and valid real files through the production opener. This crosses the reviewer train seam; coordinate rather than silently claiming its behavior.

**OWED — opening panic is not translated to exit 2.** `internal/cli/run.go:48` opens the reviewer before the recovery in `reviewCall` at `:140-147`; the panic test exercises `Review` only. Remedy: catch a panic while opening and test both stages.

### Unplug
An exclude-all scan and a nonexistent baseline were behavioral negative controls; neither reddened the scan, confirming the selection gap. Invalid `PROVEN_SEVERITY` under a valid overriding flag reddened with code 7, confirming the precedence gap. Source-mutant tests were not run; the executor's reported three-test unplug is not independently verified here.

### Closure
Reject. The real CLI cannot substantiate change-only review, path filters, flag priority, or loaded configuration. The master must route repairs and verify at the real opener before a gate; this review does not fix product code.

### Remainder
JSON and SARIF writers explicitly refuse until the formats wagon (`internal/cli/format.go:19-27`), so that announced limitation is not counted here. Text output does show F1/F2 evidence and failed services. Declined to judge the separately owned agent implementations, Action packaging, real outside-service credentials and network behavior beyond existing sample tests. Ruling: changed-only CI remains binding despite a downstream seam that ignores selection; cost if wrong is a false claim of CI coverage.

### Profile
Configuration and file output were inspected as input boundaries. No dependency manifest changed. The suite passes, but capture-reviewer tests cannot prove the production selection path; no mutation or coverage threshold was supplied. No numeric score is asserted. No additional destructive path operation was exercised.

### Focus
Condition 1 fails on malformed overridden environment values and flags that have no downstream effect. Condition 2 partly passes via unit tests (0/1, review panic 2, config 7, named service failure); opener panic remains. Condition 3 passes in tests for unknown flags and invalid severity. Condition 4 has existing sample no-environment/no-network tests and inspected `internal/pipeline/pipeline.go:41-54`, but independent network tracing was not done. Condition 5 fails: the fixture tests copied base/head values but not their effect. Condition 6 fails for selection and configuration: no unplug-sensitive production tests exist.

### Conditions
Fix and verify baseline/head plus glob selection, compute the CI merge base, apply flag priority before parsing variables, and actually load/refuse Proven configuration. Add real-reviewer differential tests and negative controls. The master must not reclassify these blocking conditions downward.

### Unrun
Source mutation of each branch; coverage and mutation score (no thresholds supplied); GitHub-hosted CI; independent network tracing; model/cost receipt (not available in this worktree). Existing suite success does not substitute for these measurements.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject: CI and scan silently ignore the baseline and path filters; overridden environment values can cause refusals. Repair and test before the gate.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T03:03:43Z

## REDACTED

none
