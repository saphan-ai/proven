# REVIEW — json-sarif-outputs · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T20:19:31Z

## Summary

BLOCKING: dangling output symlink bypasses load-time collision refusal; exit 0 replaces JSON with SARIF. Repair and test before gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground
Reviewed HEAD `04c5815` (confirmed by `git rev-parse --short=7 HEAD`) against the W2 files introduced after `ce61e44`; the local `main` ref is absent, so `origin/main...HEAD` includes unrelated work and was not used to attribute W2 findings. Examined the writers, caller, output-path guard, documentation, and schema/contract tests. The order's later-dated 2026-10-24 verification was treated as a claim, not as today's measurement (2026-10-23). No product source was changed.

### Differential
BLOCKING — `internal/cli/config.go:496-505`, `internal/cli/config.go:470-473`, `internal/cli/run.go:233-238`: a symlink to a not-yet-created output file is not resolved. `EvalSymlinks(alias.json)` fails because its target does not exist; the loop then resolves only the parent, leaving `alias.json` as the identity. `os.Stat(alias.json)` also fails, so the identity and same-file checks miss `actual.json`. A normal `os.WriteFile(alias.json, ...)` follows that link and creates `actual.json`; the subsequent SARIF write overwrites the JSON. This is the same collision property as the earlier BLOCKING finding, not an unrelated new review topic. Remedy: resolve a dangling final symlink's target when deriving output identity (including relative link targets), check identity before review, and add a test that asserts exit 7, both flags in the refusal, no review, and no output. Test each pair of output flags with a dangling alias. Negative control: remove that resolution and observe the test return exit 0 with the later format replacing the first.

### Unplug
Measured a genuine false-world answer: a small temporary program under the work directory passed a fake reviewer returning an empty result to `cli.Run`, with `-o alias.json`, `--sarif-output actual.json`, and `alias.json -> actual.json` while the target did not exist. It printed `exit=0` and the final file began `{"$schema":...`, i.e. SARIF replaced JSON. The binary's same-path control refused with exit 7; its dangling-link variant reached reviewer setup and exited 2 (the sample reviewer's local-state restriction), independently confirming no load-time refusal. Existing `TestOutputsThroughASymlinkedDirectoryAreRefused` and `TestSymlinkedFileAndHardLinkAreRefused` do not use a dangling target. Schema tests remove `version` and expect a named rejection. The SARIF shape test includes no evidence, unresolved, failed service, bad/empty locations and empty result. Contract tests decode documented JSON and SARIF fields and compare values; level is asserted by result id. These tests check the existing branches, not the missing dangling-link branch.

### Closure
`go test -count=1 ./internal/cli/...` passed both packages; `go vet ./internal/cli/...` passed. The writers preserve query/test evidence and unresolved in JSON and SARIF, and the JSON contract names the documented fields (`internal/cli/format_json.go:43`, `internal/cli/format_sarif.go:140`, `docs/OUTPUT_FORMATS.md:9`). The schema test validates the edge shapes against the vendored schema and removes `version` as a negative control (`internal/cli/sarifschema/schema_test.go:27`). Ordinary output pairs, symlinked existing directories, existing file symlinks and hard links are covered (`internal/cli/outputs_test.go:32`, `internal/cli/outputs_test.go:69`, `internal/cli/outputs_test.go:100`). The dangling-target pair is not; its positive result is materially wrong and trusted, so reject.

### Remainder
An existing documented limitation is case-only spelling of an as-yet-uncreated file on a case-insensitive volume (`docs/OUTPUT_FORMATS.md:5`); it was not reproduced on this host. The artifact confinement prevents using an external state directory for a real sample scan; the fake reviewer exercises the same `cli.Run` load/write path without a live review. No accepted exception, baseline metric, or declared tolerance was supplied in this review order. No bar-weakening moves were found in the W2 test files inspected. Dependency change adds the schema validator for tests; no install script delta observed. Declined to judge live adapter behavior, other trains' `internal/review`, licences, and formats on case-insensitive volumes.

### Profile
Backend and docs. Check ranking: the vendored external SARIF schema is least circular for structural validity; project documentation is a human-authored contract; same-package unit and integration tests are most circular. No performance or coverage baseline was declared for this review. The repro uses local files and a fake reviewer, no network.

### Focus
The mandated six properties: schema and negative schema control checked; both evidence kinds and unresolved checked by decode tests; multi-output writes checked but collision claim falsified by the dangling-link case; documented JSON keys checked; unplug absent for that case; ADR-0027 commentary and property-named tests observed for the changed writer/guard. The failed collision property determines the verdict regardless of the passing checks. `git diff ce61e44...HEAD -- AGENTS.md internal/review` is empty.

### Conditions
Reject. The owner should not gate this revision as accepted. Repair the dangling-output-symlink collision and supply a regression that fails on `04c5815`, passes after the repair, and runs before review setup. Re-run `go test -count=1 ./internal/cli/...` and `go vet ./internal/cli/...` after the fix. No review-condition reclassification is requested.

### Unrun
Full repository test suite, live generator, live review, and a base-commit test run were not performed: this leg is confined to W2, the live adapters need a runtime/service, and the local `main` ref is absent. The test-author's proposed dangling-symlink regression does not exist yet; the independent fake-reviewer reproduction above ran instead. SARIF behavior on a case-insensitive volume was not run because the host does not provide that measured environment.

Cleanup limitation: the workspace's command interceptor refused artifact deletion (`rm -rf`), so the temporary, untracked `█████████████████████████████████████████████` remains. No equivalent deletion command was attempted. It is not a review return artifact; remove it through the authorized cleanup path before requiring a pristine worktree.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: dangling output symlink bypasses load-time collision refusal; exit 0 replaces JSON with SARIF. Repair and test before gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T20:13:32Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
