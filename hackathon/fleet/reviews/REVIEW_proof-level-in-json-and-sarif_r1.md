# REVIEW — proof-level-in-json-and-sarif · round 1

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T03:44:44Z

## Summary

Accepted the JSON and SARIF proof-level projection at 540616b; focused and full Go tests, SARIF schema, vet, and field-removal unplug checks passed as expected.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — proof-level-in-json-and-sarif

### Ground

Reviewed HEAD `540616b` (measured by `git rev-parse --short=7 HEAD`) against `d4785e2...540616b`. The diff changes five files: two writers, two test files, and `docs/OUTPUT_FORMATS.md`. The core contract in `internal/review/review.go:169-209,269-270` defines the four proof levels and says a Result finding has a nonempty level. `internal/review/open.go:427-466` assigns the level and counts it. The order's earlier test claim was set aside and remeasured locally.

### Differential

`internal/cli/format_json.go:164,181-193` copies `Result.ProofCounts` by the core's typed keys and copies `Finding.ProofLevel` as its string. `internal/cli/format_sarif.go:119,135` uses the same count conversion and the same finding field. Neither writer derives a proof level from evidence. `internal/cli/output_contract_test.go:54-85,134-149,214-243` uses deliberately different counts and finding evidence so a recount or rededuction cannot satisfy the expected output. The sample test checks the same finding IDs and four counts in both formats. The old output contract expectations change only by the new fields; no other production output path changes in the diff. No dependency, threshold, skip, suppression, or fixture golden file changes were found.

### Unplug

Using `go test -overlay` with replacement writer files under `████████████████████████████████████████████████████████`, I removed only the JSON finding-level assignment. `TestSampleJSONAndSARIFCarryProofLevel` and `TestJSONDocumentIsTheContract` failed, naming empty JSON levels while SARIF retained the core values. I separately removed only SARIF `properties.proof_level`; `TestSampleJSONAndSARIFCarryProofLevel` and `TestSARIFPropertiesAreTheContract` failed, naming the missing field. Both mutations changed exactly the field guarded and left the checkout's source intact. SARIF schema validation passed on the unmodified tip as part of the focused test run.

### Closure

Accept. The five requested properties hold on the reviewed commit. No BLOCKING or OWED finding.

### Remainder

No code finding. The review did not judge licensing, live generator behavior, or any change outside `d4785e2...540616b`, as ordered. No declared numeric quality bar or baseline was found in the supplied order; the suite and vet are project checks, and the field-removal overlay gives a negative control of the assertion.

### Profile

Docs and output-contract review. `go.mod` pins Go 1.26; `go version` reported go1.26.5 linux/amd64. `go test -count=1 ./internal/cli ./internal/cli/sarifschema ./internal/review` passed all three packages. `go test -count=1 ./...` passed every test-bearing package; packages without tests were reported by Go as such. `go vet ./...` and `git diff --check d4785e2...540616b` exited 0. No coverage or mutation score baseline was declared; check ranking is suite assertions plus SARIF schema validation, with no external metric claimed.

### Focus

The JSON table at `docs/OUTPUT_FORMATS.md:28,31` names `proof_counts` and `findings[].proof_level` with all four exact core strings. The SARIF sentence at line 53 places the same values in result and run `properties`. JSON and SARIF are checked against each other and against core values in the tests. SARIF schema validation remains in `internal/cli/sarifschema/schema_test.go:63` and passed.

### Conditions

none

### Unrun

Base-commit test suite: not run. The order confines this leg to this checkout and requires leaving its source unchanged; the diff and unmodified tip were reviewed directly. Live endpoint and container tests: not run because the task is the sample and output writers, with no live credentials or runtime evidence requested. Race detector: not run because the diff changes no shared state or goroutines.

## The review status as filed

## REVIEW — proof-level-in-json-and-sarif
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: Accepted the JSON and SARIF proof-level projection at 540616b; focused and full Go tests, SARIF schema, vet, and field-removal unplug checks passed as expected.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T03:36:21Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
