# REVIEW — json-sarif-outputs · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T19:03:52Z

## Summary

Blocking symlinked-directory output collision returns exit 0 and overwrites JSON with SARIF; direct collisions and format/schema checks pass. Repair path identity guard and add a failing alias test before gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground
RED — BLOCKING: Output aliases through a symlinked directory overwrite one another, rather than being refused at load. `internal/cli/config.go:434-463` compares `filepath.Abs` paths without resolving symlinks; `internal/cli/run.go:198-244` then writes both bodies in order. This violates condition 3 of the review order, including its explicit symlinked-directory case. Remedy: resolve existing path components, including the parent of a not-yet-created output, before comparing all four output destinations; refuse the same destination with exit 7 and add pairwise symlink and relative/absolute tests. Protect against output-path changes between validation and writing if concurrent path mutation is in scope.

Fresh HEAD: `git rev-parse HEAD` = `f6b6b1f4b36631aaaf2dae3399d411e07f172c06` (short form `f6b6b1f`). `main` has no local ref; `origin/main` = `a14ecd0914048a2913a0aeed392eb3d206bde123`, the merge base used for `git diff origin/main...HEAD`. Reviewed W2's `internal/cli` output writers, configuration, tests and `docs/OUTPUT_FORMATS.md`; changes to the reviewer and agents belong to the other train. The author's report of 36 passing packages and a direct spelling collision was not treated as a measurement.

### Differential
Measured reproduction: using a stub reviewer and `cli.Run` from the checkout, with `-o=█████████████████████████████████████████████████████████████` and `--sarif-output=██████████████████████████████████████████████████████████████`, where `alias` is a symlink to `real`, returned `exit=0 same_file=true`; the file began with SARIF's `$schema` rather than JSON's `class`. This discriminates from a correctly refused collision (exit 7, no output). The temporary stub and files were confined to the worktree; cleanup was denied by the command guard, so they remain under `█████████████████████████████████████████████`. The ordinary `proven scan` binary instead returned `hmem: state directory ... is inside the repository` from this confined location; the stub exercised the same CLI load and output path without running the review pipeline.

`internal/cli/outputs_test.go:24-80` covers the six flag pairs under `..` and exact same-path spelling; it does not exercise symlinked directories. `docs/OUTPUT_FORMATS.md:5-7` explicitly says symlinks are not followed, in conflict with the review-order requirement. `internal/cli/config.go:390-392` calls the guard before review execution; `internal/cli/run.go:198-244` confirms the later output replaces the first.

### Unplug
`go test ./internal/cli -count=1` PASS; `go test ./... -count=1` PASS; `go vet ./internal/cli` PASS; `git diff --check origin/main...HEAD -- internal/cli docs/OUTPUT_FORMATS.md` clean. These checks do not falsify the symlink case: removing symlink resolution changes nothing because no such check exists. Required negative control: a test creating a real directory and symlink alias must assert exit 7, no reviewer call and no output for every flag pair; current implementation gives exit 0 and a SARIF overwrite for the reproduced pair. Mutated field: second output path only.

Schema check: `█████████████████████████████████████████` validates the sample and table-driven empty-evidence, unresolved, failed-service, no-path and invalid-line shapes against the vendored SARIF schema; removing `version` makes the schema reject the log. `internal/cli/format_sarif_test.go:82-105` asserts level by result id, and `internal/cli/output_contract_test.go:104-230` decodes and compares documented fields, `[]`, unresolved and evidence. These are suite tests (circular); the vendored schema is an external rule set, subject to its pinned version. No baseline metric, tolerance, mutation command or declared quality floor was found in this review order; do not infer a mutation score.

### Closure
Earlier direct `-o`/`--sarif-output` collision: ADDRESSED for identical and cleaned paths (`internal/cli/config.go:390-392,440-463`, `internal/cli/outputs_test.go:64-80`). Earlier text-only fields: ADDRESSED in `internal/cli/format_json.go:30-52,146-175` and `internal/cli/format_sarif.go:107-118`; the tests compare their values. Edge-schema rows, result-id severity, exact documented JSON keys and SARIF evidence: ADDRESSED by the tests cited under Unplug. Full collision property: NOT ADDRESSED for symlinked directories, with a confirmed overwritten report. Verdict: reject until the collision guard and an unplugging test cover that path.

### Remainder
No additional blocking finding established. Documentation and implementation agree with one another on the symlink exception, but the explicit review contract does not allow that exception. Declined to judge: concurrent replacement of directory entries after load, cross-filesystem hard-link aliases, and any changes outside W2's output seam. The owner must rule on scope if these are expected. No project-declared quality baseline was found in the order; first measure a baseline rather than inventing a threshold.

### Profile
CLI-facing output and local filesystem writes. Inputs cross from user-supplied paths into file writes; the tested output alias causes an apparently successful but wrong file. External conformance check (vendored SARIF schema) is stronger than project checks (vet/diff) and suite checks (in-process Go tests); none tests the symlink-path guarantee. No dependency delta attributable to W2's CLI output seam was found in the scoped files. No threshold, test suppression or exception changed in the scoped diff.

### Focus
1. SARIF shapes and negative schema control: covered by named tests; not an exhaustive proof over arbitrary invalid future result values.
2. Finding evidence, unresolved, text-only review statements and failed services: preserved by the JSON and SARIF writers with decode tests.
3. Multi-output: distinct paths and direct/cleaned collisions tested; symlinked-directory collision fails (BLOCKING). Write-failure exit 2 is asserted by `internal/cli/outputs_test.go:99-120`.
4. Documented JSON names and arrays: exact decode comparison in `internal/cli/output_contract_test.go:104-180`.
5. Unplug: version deletion, per-id severity comparison and exact fields; missing symlink negative control identified above.
6. ADR-0027: writers and exported formatter names have doc comments, decisions are cited, tests name properties; no further blocking documentation omission found in the scoped paths.

### Conditions
Reject. Reconcile two paths after directory symlink resolution before any review or write, including every pair of `-o` and `--*-output`; make an aliasing-path test fail on this commit and pass after the repair. Update `docs/OUTPUT_FORMATS.md` to match the required behavior. Do not mark this review accepted while the symlink control is absent.

### Unrun
A full real-review `proven scan` from the worktree's confined scratch directory could not run: H-MEM refused a state directory inside the repository. The stub-reviewer CLI reproduction establishes the path guard and overwrite, not the real reviewer's end-to-end result. Scratch cleanup was attempted but refused by the command guard; no alternate deletion spelling was used. No base-commit test run, runtime path mutation test or mutation score was performed; the base is not checked out in this review worktree and no such metric was declared.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Blocking symlinked-directory output collision returns exit 0 and overwrites JSON with SARIF; direct collisions and format/schema checks pass. Repair path identity guard and add a failing alias test before gate.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T18:59:32Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| file and line in Saphan Studio's private code | 1 |
