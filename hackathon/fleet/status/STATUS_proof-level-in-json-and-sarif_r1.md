# STATUS — proof-level-in-json-and-sarif · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 19 min
- last commit: 540616b9988e8c03b360d2cb6a5733d485397802 [proof-level-in-json-and-sarif] cli: add proof level to JSON and SARIF
- updated: 2026-10-25T02:08:58Z

## Summary

JSON and SARIF now carry each finding's proof level and the counts by level, under the names the reviewer core defines.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — proof-level-in-json-and-sarif
- state: stop-2-ready
- abstract: JSON and SARIF now carry each finding's proof level and the counts by level, under the names the reviewer core defines.
- last commit: 540616b9988e8c03b360d2cb6a5733d485397802 [proof-level-in-json-and-sarif] cli: add proof level to JSON and SARIF
- done: JSON and SARIF carry each finding's level of proof and the totals by level.
  - TestSampleJSONAndSARIFCarryProofLevel passes: F1 is proved-by-test and the dependency finding osv:evil.example/leftpad is proved-by-data in both formats, and proof_counts is 1, 0, 3, 0.
  - go test -count=1 ./... exit 0 on the tip. go vet ./... exit 0. gofmt -l . printed nothing and exited 0.
- next / blocked-on: none
- artifacts:
  - docs/OUTPUT_FORMATS.md
  - internal/cli/format_json.go
  - internal/cli/format_sarif.go
  - internal/cli/format_prooflevel_test.go
  - internal/cli/output_contract_test.go
- evidence:
  - git rev-parse HEAD: 540616b9988e8c03b360d2cb6a5733d485397802
  - git log --oneline d4785e2..HEAD: 540616b [proof-level-in-json-and-sarif] cli: add proof level to JSON and SARIF
  - Baseline on d4785e24d9cc9689a470bdf7913df84214c4e530 before any edit: go test -count=1 ./... exit 0. go version go1.26.10; go.mod says go 1.26 and has no toolchain line.
  - Red on that base, go test -count=1 -v ./internal/cli -run TestSampleJSONAndSARIFCarryProofLevel exit 1. JSON findings F1, F2, F4 and osv:evil.example/leftpad have no proof_level. JSON F1: no proof_level; want proved-by-test. JSON osv:evil.example/leftpad: no proof_level; want proved-by-data. SARIF findings F1, F2, F4 and osv:evil.example/leftpad have no proof_level. SARIF F1: no proof_level; want proved-by-test. SARIF osv:evil.example/leftpad: no proof_level; want proved-by-data. JSON proof_counts is nil. SARIF proof_counts is nil.
  - After the writers copied the fields, the same test exit 0. Removing only those copies made it exit 1 again (JSON proof_level empty, counts 0, 0, 0, 0; SARIF keys absent). Restoring the copies made it exit 0.
  - go test -count=1 ./internal/cli ./internal/cli/sarifschema exit 0. go vet ./... exit 0. gofmt -l . exit 0, no files. go test -count=1 ./... exit 0 on the committed tree.
  - go test -race UNRUN: this change adds no goroutine and no shared mutable state.
  - Testing bar UNRUN: no metric, command or baseline is declared for this repository.
- note: master-was-wrong: the order named a known red in internal/pipeline's mode test. On d4785e2, go test -count=1 ./... exited 0, and internal/pipeline printed ok, so TestMaterializedTreesAreNotGroupWritableUnderAPermissiveUmask passed. ruling: the dependency finding is the one whose id starts with osv: — F2's title also says Dependency, and it is a planner finding; the test pins osv:evil.example/leftpad. Cost if wrong: F2 is not pinned to a level by name. Both are proved-by-data in the core's sample test, and F2's level is still written and compared between JSON and SARIF. ruling: proof_counts is a struct in review.ProofLevels order, and a missing map reads as zeros. The contract golden sets levels that disagree with the evidence and counts 4, 1, 0, 9, which are not a recount, so the writer has to copy the fields. The golden diff adds proof_level and proof_counts only; gofmt re-aligned spaces on two existing composite literals and did not change their values.
- updated: 2026-10-25T02:03:08Z

CLOSE-OK · proof-level-in-json-and-sarif · commit 540616b9988e8c03b360d2cb6a5733d485397802 · branch fix/proof-level-in-json-and-sarif · worktree █████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
