# REVIEW — json-sarif-outputs · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T04:07:07Z

## Summary

Writers and collision controls pass; mutations remain unrun, and temporary base cleanup was refused.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground

`git rev-parse HEAD` returned `f4819cf715c09c3f52f82b5d746947d2ba7ee403`. The checkout has no `main` ref; the available commit `ce61e444ead43c21a846f2123853d03881316a13` is an ancestor of this tip, so I reviewed `ce61e44...HEAD`. I did not use the author's claimed verification as evidence. No `internal/review` or agent code is in this diff; I read their result and evidence types and callers to check the formatter seam.

### Differential

- `internal/cli/format.go:24`, `internal/cli/format_json.go:119`, `internal/cli/format_sarif.go:69`: the base rejected JSON and SARIF. The tip writes both from `review.Result`; decoded contract tests compare the documented JSON keys and SARIF run/result properties to literal expected values, including tests, queries, unresolved marks, failed services, empty arrays and text-only statements (`internal/cli/output_contract_test.go:105`, `internal/cli/output_contract_test.go:188`). Per-result severity is asserted by id (`internal/cli/format_sarif_test.go:82`).
- `internal/cli/config.go:448`, `internal/cli/config.go:514`: the new load guard resolves each component in filesystem order, including links before `..`, dangling links and relative names; existing hard links are compared by `os.SameFile`. The base `writeOutputs` wrote one path after another without a collision check. The tip additionally checks file identity after review, before writing the next path (`internal/cli/run.go:245`). The same result is rendered once per format (`internal/cli/run.go:198`), and a write error returns exit 2 (`internal/cli/run.go:100`).
- `go.mod` adds `github.com/santhosh-tekuri/jsonschema/v5 v5.3.1` for the SARIF schema tests; `go.sum` has its two hashes. No install scripts changed. `git diff --check ce61e44...HEAD` passed.
- Measured base: `go -C █████████████████████████████████████████████████ test ./internal/cli/...` passed (one package); base `go -C █████████████████████████████████████████████████ vet ./internal/cli/...` clean. Measured tip: `go test ./internal/cli/...` passed (two packages); `go test ./...` passed across reported packages; `go vet ./...` clean. These readings prove compilation and asserted cases, not coverage or all production environments. No numeric baseline, tolerance or performance target is declared. Ranking: OASIS schema is external-standard content exercised by suite code; vet is a tool check; writer and path tests are suite checks, which are easiest for an author to make circular.

### Unplug

`internal/cli/sarifschema/schema_test.go:27` removes only SARIF `version` and requires rejection; this negative control was executed by `go test ./internal/cli/...`. `TestSARIFShapesMatchTheSchema` includes missing evidence, unresolved, failed service, empty path, zero start, reversed end, unknown severity and empty result. `TestAliasTableIsCaughtByTheLoadCheckNotTheWriteGuard` (`internal/cli/outputs_alias_test.go:129`) would turn red if link resolution before `..` were removed; it checks the load guard directly, not just the write guard. `TestAliasCreatedAfterLoadIsRefusedAtWrite` (`internal/cli/outputs_test.go:324`) would turn red without the late `os.SameFile` check. `TestJSONDocumentIsTheContract` would turn red on a renamed JSON tag, `TestSARIFPropertiesAreTheContract` on dropped SARIF evidence or unresolved, and `TestSARIFLevelFollowsSeverity` on interchanged warning/error branches. The last mutations are predictions from the assertions, not executed mutation runs.

### Closure

The original identical-path collision and the later symlinked-directory collision are ADDRESSED: `TestCollisionReproduction`, `TestOutputsThroughASymlinkedDirectoryAreRefused` and the alias-table tests now require exit 7 before review and no write. Earlier text-only-line, schema-shape, per-id-level and complete JSON/evidence conditions have direct assertions. I found no reproduced blocking product defect. The condition on this recommendation is independent mutation verification, not another production repair.

### Remainder

NOTE: `docs/OUTPUT_FORMATS.md:5` says paths are made absolute and cleaned and then describes link resolution; that wording can imply cleaning before a link is followed. The implementation rightly resolves links before `..` (`internal/cli/config.go:492`). Clarify the ordering when next editing the documentation. Declined to judge live service responses, GitHub's SARIF presentation, Windows/case-insensitive volumes and concurrent hostile path replacement between checks and writes. No exception, skipped test, weakened assertion or lowered threshold appears in the diff.

### Profile

CLI output/filesystem and documentation. User-supplied output paths may refer to hard links or symbolic links; read the load guard, write guard, and their caller. The writer adds no network call; the schema test uses the vendored copy. No promise of atomic publication across all output files or a restricted destination root is in this order. ADR-0020 is named beside the formatter decision; changed exported names have name-leading comments, choice points have WHY comments, and tests name their property, as ADR-0027 requires.

### Focus

Conditions 1–4 pass for the tested result shapes and output aliases. Condition 5 has targeted assertions and one executed schema unplug; source mutations for the other assertions were not executed here. Condition 6 is met on the changed code. R14 disposition: previous product findings ADDRESSED; the new documentation wording is non-blocking remainder. Any proposed reclassification belongs to the master, not this reviewer.

### Conditions

Before an unconditional accept, independently invert the load-time symlink-before-`..` guard, the write-time same-file guard, a JSON field mapping and the SARIF severity branch in disposable copies. Run the named focused tests and record that each turns red. Do not equate a green suite with a measured unplug. This is owed verification, not a production-code change.

### Unrun

Standalone mutation campaign for output guards and serializer branches: not executed. Full-suite base comparison and numeric coverage/mutation baseline: no metric, command, threshold or tolerance was declared; only the scoped base package test and vet ran. Live services and GitHub SARIF upload: outside this local review. `main...HEAD`: `main` is absent; the available `ce61e44...HEAD` comparison was used instead.

Cleanup limitation: the attempt to remove the disposable base snapshot was refused by the command environment with `rm -f style commands are not permitted. Use a safer approach`. Under the denied-act rule I did not retry that deletion in another spelling. The untracked `██████████████████████████████████████████████████` therefore remains. This review does not claim the checkout was restored to its starting state; removal must be performed by an authorized actor before teardown.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Writers and collision controls pass; mutations remain unrun, and temporary base cleanup was refused.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T04:03:19Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
