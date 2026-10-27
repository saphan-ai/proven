# STATUS — meterless-conformance-shim · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: a53aac43d2c8103bead5cc4bebdd3a30dd4cf6ca [meterless-conformance-shim] conformance: shim the meterless runners
- updated: 2026-10-22T13:21:00Z

## Summary

RED: go test ./... exit 1. Six tests in internal/change and internal/review fail because the seat commit hook refuses subjects first and base. The same six failed before this edit. H-MEM 9/13: FAIL sleep-guardrails, conflict-gate, ledger-audit, dream-boundary (ADR-0018 deferred). World 7/8: FAIL bitemp-001 (no bitemporal graph, outside the minimum). Minimum checks passed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — meterless-conformance-shim
- state: stop-2-ready
- abstract: RED: go test ./... exit 1. Six tests in internal/change and internal/review fail because the seat commit hook refuses subjects first and base. The same six failed before this edit. H-MEM 9/13: FAIL sleep-guardrails, conflict-gate, ledger-audit, dream-boundary (ADR-0018 deferred). World 7/8: FAIL bitemp-001 (no bitemporal graph, outside the minimum). Minimum checks passed.
- last commit: a53aac43d2c8103bead5cc4bebdd3a30dd4cf6ca [meterless-conformance-shim] conformance: shim the meterless runners
- done: Meterless runners load the shims and the minimum checks pass
  - bash conformance/meterless/run.sh exit 1, H-MEM 9/13, World 7/8, neither runner exit 2
  - go test -count=1 ./conformance/meterless ./internal/hmem ./internal/world exit 0
  - go build ./... exit 0 and go vet ./... exit 0
  - go test -race -count=1 ./internal/hmem ./internal/world exit 0
  - negative control restored: add drops its argument, rv-002 and every later H-MEM check FAIL, rv-001 stays PASS
- next / blocked-on: none
- artifacts:
  - cmd/meterless-conform/main.go
  - conformance/meterless/bridge.ts
  - conformance/meterless/hmem/index.ts
  - conformance/meterless/world-model/index.ts
  - conformance/meterless/run.sh
  - conformance/meterless/run_test.go
  - conformance/meterless/SCORECARD_hmem.txt
  - conformance/meterless/SCORECARD_world-model.txt
  - internal/hmem/hmem.go
  - internal/hmem/hmem_test.go
  - internal/world/world.go
  - internal/world/world_test.go
- evidence:
  - go version go1.26.8, go.mod says go 1.26. No Makefile and no CI workflow in the tree. The order's go test is the gate.
  - Baseline before edits: go test -count=1 ./... exit 1. FAIL internal/change TestBetweenTwoCommits, TestBetweenWorkingTreeIncludesEditsAndUntracked, TestBetweenMissingBaselineNamesTheCommit, TestBetweenMissingHeadNamesTheCommit, TestBetweenWorkingTreeInSubdirectoryKeepsUntracked. FAIL internal/review TestBaselineKeepsOnlyTheChangedPath. Each output quotes the commit hook refusing subject first or base.
  - Tip: go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./... exit 1 with the same six failures. conformance/meterless, internal/hmem and internal/world are ok.
  - remark: the hook refusal is what the test printed. The same tests were not run with the hook removed, so a second cause is unmeasured.
  - node v24.5.0. node on the reference H-MEM runner exit 2: TypeScript parameter property is not supported in strip-only mode. A strip-only class then loaded and printed probe-ok.
  - bash conformance/meterless/run.sh from /tmp exit 1. H-MEM 9/13. World 7/8.
  - Negative control, shim add called with an empty input, then restored. H-MEM 1/13: PASS rv-001, FAIL rv-002 expected accept got reject, FAIL fd-001 through fd-005 and ro-001 ro-002 and the four deferred checks with hmem: a record with no source is refused.
  - git log --format=%B a9aced6..HEAD | grep -ci co-authored printed 0. Commits d94076e, 8196c5a, feca6a0.
  - testing-bar: no metric is declared. UNRUN.
  - Session recording: none. ADR-0024 points at a recorder this leg was not given, and no recording file was in the worktree.
- note: master-was-wrong: M1 is partial. node v24.5.0 runs runner.ts. The reference module does not load under strip-only mode (parameter property in trustLedger.ts), so it is not a working control. A strip-only shim does load. No package was fetched. The shim was not stopped. M2 holds: each call is one request and one answer. ruling: runner threshold 0 is sent as -1 because the Go store treats 0 as the default 0.35. Cost if wrong: a query that wanted the default would keep every score. ruling: assigning supersededBy on the record add returned is forwarded as MarkSuperseded. Cost if wrong: ro-001 would fail. ruling: world ingest writes each observation as a trace line and idempotency is the applied-id set. The suite does not inspect entity shape. Cost if wrong: a later check that reads entities would fail. ruling: one process per call with the store on disk, because Node 24 pipe handles have no file descriptor. NOTICED: the README does not name this shim. Not in the order's file list.
- updated: 2026-10-22T12:44:22Z

## REDACTED

none
