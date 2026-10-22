# STATUS — a-review-end-to-end-in-sample-mode · round 1

- state: ready-for-review
- executor: cursor executor
- model: grok-4.7-high
- pool: not shown to this leg
- cost: not shown to this leg
- duration: not measured as one figure; the work commit time is 2026-10-22T01:32:04Z
- last commit: 8e5147d6574ac2344050a43d9d22303bc3369a2b "[a-review-end-to-end-in-sample-mode] review: sample-mode agents"
- updated: 2026-10-22T01:32:04Z

## Summary

Sample mode runs the six agents behind the review seam. A fixture change yields F1 with a recorded test and F2 with a query and its row. F3 is dropped. The trace follows F3 to the verdict `dropped: no evidence`. No network is opened. The live test-run adapter names the Prover as off and does not prove F1.

## Acceptance, item by item

1. `go build ./...` exit 0. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0. The same suite with `-race` exit 0. Before this change, `go test -count=1 ./...` exit 0 with `internal/review` and no test files. The three commands were run separately: the leg hook admits one program per call.
2. The printed review is below. It has F1 and F2, each with evidence, and no F3. It was produced by `go test -count=1 -v ./cmd/proven -run TestScanSamplePrintsF1AndF2NotF3`, which calls the same `run` function as `proven scan --sample <fixture>`. Executing the built binary was refused by the executor permission class. The allowed form was none, so the binary was not run. The trace line from that test named a temporary `logs/trace.jsonl`; that path is omitted here.
3. F3's four trace events are below. The verdict text is `dropped: no evidence`.
4. `TestNegativeControl_F3RecordedQueryRow` mutates `tables.sql` (an INSERT for package `billing/note`, the row F3's query selects). F3 appears with that row. `TestNegativeControl_RemoveF1RecordedTest` mutates `recorded_tests` (the entry whose `finding_id` is F1). F1 disappears and F2 stays.
5. `TestUnplugLiveTestRunNamesProverOff` runs the review with `testrun.Live`. The review returns no error, names service `prover` with reason `Prover off: no container runner yet`, and does not report F1. F2 stays.
6. `TestSampleModeOpensNoNetwork` runs the sample review and reads 0 dials. `TestGuardDialIsVisible` performs one HTTP round trip on the guard and the dial is counted and refused. `TestSamplePackagesDoNotImportNetOrExec` refuses imports of `net`, `net/http` and `os/exec` outside the guard package.

## The printed review

```
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test ./billing -run TestDiscounted
  output: base: PASS
change: FAIL got 60 want 80
finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
  go.mod:5-5
  The change requires evil.example/leftpad v1.2.3, and the sample advisory table has a known-exploited row for that version.
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'
  columns: package, version, cve, kev
  row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true
failed services: none
```

## F3 trace events

```
{"finding_id":"F3","kind":"claim","agent":"planner","text":"note may include a payment token"}
{"finding_id":"F3","kind":"evidence","agent":"investigator","text":"none"}
{"finding_id":"F3","kind":"refutation","agent":"skeptic","text":"no refutation"}
{"finding_id":"F3","kind":"verdict","agent":"judge","text":"dropped: no evidence"}
```

## Evidence

- Toolchain: `go1.26.8`, module line `go 1.26`. No Makefile and no CI configuration found in the repository.
- Baseline: `go test -count=1 ./...` exit 0, package `internal/review`, no test files.
- RED, before `Open` was wired: `TestOpenSampleReviewNotImplementedYet` failed with `review: not implemented`.
- Tip: `go test -count=1 -v ./...` — 22 passed, 0 failed, 0 skipped. `go test -count=1 ./...` after the log lines were removed: exit 0. `go test -race -count=1 ./...` exit 0.
- Testing bar: no metric is declared. UNRUN.
- Contract check against a schema: no contract file and no compatibility tool in the repository. UNRUN. The seam types were not given new fields. `Open` now returns a reviewer. Sample evidence is copied into the existing `Evidence` and `ServiceFailure` types.
- Session recording: none. This leg was not given a recording file to publish.

## master-was-wrong

none.

- M1 held. `Result` already has test evidence, query evidence and `FailedServices`. The sample review fills those fields. Instrument: `internal/review/review.go`, `TestOpenSampleReviewKeepsF1AndF2`, `TestUnplugLiveTestRunNamesProverOff`.
- M2 held. A SELECT on the local engine returns the sample row. Instrument: `TestLocalEngineReturnsTheSampleRow`. No cgo and no new module.
- M3 held. Six agent packages. `go build ./...` exit 0. No import cycle.

## Rulings

- `AGENTS.md` says the owner owns that file. This order names one section, so this wagon adds that section and changes nothing else in the file.
- The acceptance list names the tests, and this dispatch has no separate test-author leg, so the tests are in this change.
- `cmd/proven` implements `scan --sample` only, so the acceptance command has a function to run. Flags, formats and the exit-code table stay with the shell train.
- The judge and the writer do not call the generator. They apply the evidence rule in code.
- `MinSeverity`, include and exclude are not applied here. The seam says the shell filters.
- `ConfigPath` is accepted and not read.
- The local engine runs CREATE TABLE, INSERT, and SELECT with equality and AND.
- `internal/review/fixture` is not in this tree. This wagon does not add the shell's fixture reviewer.
- The order names base `58f653f`. The worktree HEAD at the start was `3a1a9f0`, which contains that commit. The work is on that HEAD.
- The dial counter sees dials on the review's guarded client. A new client inside an agent would not increment it. The import walk is the second instrument.

## Blocked on / next

none. Ready for the master's verification. Merge is the owner's act.
