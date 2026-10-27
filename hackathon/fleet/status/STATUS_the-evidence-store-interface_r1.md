# STATUS — the-evidence-store-interface · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 27 min
- last commit: 5403de44342eb5a7c534527082743875c55ab37b [the-evidence-store-interface] merge: absorb main
- updated: 2026-10-24T07:45:21Z

## Summary

One evidence-store interface: Zetaris by default, a local one-file store, the sql form named on the evidence, and couchdb refused by name. Next is the named review.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-evidence-store-interface
- state: stop-2-ready
- abstract: One evidence-store interface: Zetaris by default, a local one-file store, the sql form named on the evidence, and couchdb refused by name. Next is the named review.
- last commit: 5403de44342eb5a7c534527082743875c55ab37b [the-evidence-store-interface] merge: absorb main
- done: The evidence store is one interface, with Zetaris and a local file behind it.
  - PROVEN_EVIDENCE_STORE defaults to zetaris and refuses couchdb by name, proved by TestLoadDefaultsToZetaris and TestCouchDBIsRefusedByName
  - the local file round-trips a quoted cell and sample mode keeps the engine name local-sample, proved by TestLocalFileKeepsARegisteredTable and TestOpenSampleIsTheLocalFile
  - one KEV scenario matches on a recorded Zetaris HTTP server and the real local file, proved by TestStoresAgreeOnOneScenario
  - negative control mutated res.Rows in Local.Run, TestStoresAgreeOnOneScenario went red, and the field was restored
  - go build ./... exited 0 and go vet ./... exited 0
- next / blocked-on: none
- artifacts:
  - internal/store/store.go
  - internal/store/settings.go
  - internal/store/settings_test.go
  - internal/store/local.go
  - internal/store/zetaris.go
  - internal/store/conform_test.go
  - internal/query/file.go
  - internal/query/file_test.go
  - internal/querytpl/forms.go
  - internal/querytpl/forms_test.go
  - docs/ADAPTERS.md
  - docs/CONFIGURATION.md
- evidence:
  - base go test -count=1 ./... at b3a0303af55a87bd481381947810f6ee1ced0798 exited 1. Already red: TestMaterializedTreesAreReadableByOthers, go.mod mode 664
  - tip go build ./... exited 0
  - tip go vet ./... exited 0
  - tip go test -count=1 ./... exited 1. The only failure is the same TestMaterializedTreesAreReadableByOthers. Every other package exited 0, including internal/store and internal/review
  - negative control: res.Rows in (*Local).Run shortened by one row. TestStoresAgreeOnOneScenario failed because the local rows were empty and the Zetaris rows held CVE-2021-44228. Restored. The test then exited 0
- note: master-was-wrong: none. Measured at b3a0303af55a87bd481381947810f6ee1ced0798, which contains the order base c1805b0: the live engine is internal/query (Load, NewLive, OpenSample) and templates are internal/querytpl. Ruling: the local file is that sample engine's SQL, not DuckDB or SQLite; if that choice is wrong, a later wagon swaps the file engine and the interface stays. NOTICED: TestMaterializedTreesAreReadableByOthers was already red on the base (mode 664). It is outside this wagon and was not changed. NOTICED: AGENTS.md does not list internal/store; that file is the owner's and was not edited. The commit message has no Co-authored-by trailer.
- updated: 2026-10-24T07:39:39Z

CLOSE-OK · the-evidence-store-interface · commit 5403de4 · branch fix/the-evidence-store-interface · worktree ████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
