# STATUS — the-evidence-store-interface · round 4

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 630c671 [the-evidence-store-interface] merge: absorb main
- updated: 2026-10-24T14:00:16Z

## Summary

Round 2 closes C1 and C2. A missing table in the configured local store is a `service.Down{local}`, so the review names it and goes on. PROVEN_EVIDENCE_FILE is required for the local store, a path inside the review target (links resolved) is refused by name, and the file is opened lazily so nothing is written into the tree. Seven listener packages are UNRUN here (sandbox refuses bind); all else is green.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-evidence-store-interface
- state: stop-2-ready
- abstract: Round 2 closes C1 and C2. A missing table in the configured local store is a `service.Down{local}`, so the review names it and goes on. PROVEN_EVIDENCE_FILE is required for the local store, a path inside the review target (links resolved) is refused by name, and the file is opened lazily so nothing is written into the tree. Seven listener packages are UNRUN here (sandbox refuses bind); all else is green.
- last commit: 630c671 [the-evidence-store-interface] merge: absorb main
- done: C1 and C2 addressed, each with a test and a measured negative control
  - C1 reproduced first: at 5403de4 `TestLocalMissingTableIsADownSource` failed with `Run: want a down source, got query: no table px_evidence.px_kev_vulns` (go test -run, internal/store)
  - C1 fixed in 45e478c: typed `query.NoTableError`; `File.Query` converts it to `&service.Down{Service: "local", Reason: "no table <name>"}` when engine is `local`; `local-sample` unchanged
  - C1 end to end: `TestEmptyLocalStoreNamesLocalAsDown` completes a review on a fresh local file and finds `local` in FailedServices
  - C1 conformance: `TestStoresAgreeOnAMissingTable` holds Local and Zetaris (404) to one answer; it needs a listener, so it compiles (go vet) but is UNRUN
  - C1 control: engine test changed to a name that never matches, so the conversion is off: `TestLocalMissingTableIsADownSource`, `TestEmptyLocalStoreNamesLocalAsDown` red; restored, green
  - C2 in 17ba318: empty or blank PROVEN_EVIDENCE_FILE with store=local is refused by name at `store.Load` (`TestLoadLocalReadsTheFileAndRefusesABrokenPath`, `TestLocalStoreWithoutAFileIsRefusedByName`)
  - C2: `store.Within` resolves links; `runner.Review` refuses a file inside `Request.Target` with a ConfigError naming PROVEN_EVIDENCE_FILE, and nothing is created in the tree (`TestWithinResolvesLinks`, `TestLocalFileInsideTheTargetIsRefused` plain and link cases)
  - C2 control: the `checkLocalFile` call removed, `TestLocalFileInsideTheTargetIsRefused` red; restored, green
  - Settings page regenerated with `go run ./cmd/settings-doc` (docs/CONFIGURATION.md); docs/ADAPTERS.md sentence updated
- note: master-was-wrong: nothing refuted (the order asks for this field, the grammar has no such key, so it lives here). One refinement: the target is unknown when `review.Open` runs, and `OpenLocal` creates the file, so the inside-target check could not sit at Open alone. The local store is now opened on the first query (`lazyLocal`), after the check in `Review`. Also an existing test changed: `TestLoadLocalReadsTheFileAndRefusesABrokenPath` expected the default file; it now expects the refusal.
- next / blocked-on: master runs the listener-refused packages
- artifacts:
  - internal/store/missing_test.go
  - internal/store/within.go
  - internal/review/local_store_test.go
- evidence:
  - `go build ./...` exit 0; `go vet ./...` exit 0; `gofmt -l .` empty
  - `go test -count=1 ./...`: UNRUN (listener refused, quoted `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`): cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review, internal/store (TestStoresAgreeOnOneScenario, TestStoresAgreeOnAMissingTable, TestLocalRunDoesNotDial). Every other package ok.
  - Non-listener subsets run alone: internal/store with the three listener tests skipped, ok; internal/query local and sample tests, ok; internal/review TestLocalStoreWithout/TestLocalFileInside/TestEmptyLocalStore/TestLocalStoreNeedsNoZetaris/TestCouchDB, ok
  - No Co-authored-by line on either commit. `packed-refs.lock: Operation not permitted` printed after each commit; `git log -1` shows both landed.
- updated: 2026-10-24T13:10:00Z

## REDACTED

none
