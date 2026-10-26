# STATUS — couchdb-evidence-viewer · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 8b985ba [couchdb-evidence-viewer] tests: records settings in sample mode
- updated: 2026-10-25T10:21:32Z

## Summary

Reviews now write one CouchDB evidence record per finding (PROVEN_COUCHDB_RECORDS, own writing account kept apart from the read-only one), with a design document of four views and the viewer page as its attachment.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — couchdb-evidence-viewer
- state: stop-2-ready
- abstract: Reviews now write one CouchDB evidence record per finding (PROVEN_COUCHDB_RECORDS, own writing account kept apart from the read-only one), with a design document of four views and the viewer page as its attachment.
  Finding.EvidenceRecordURL carries the link when PROVEN_EVIDENCE_LINK_BASE_URL is set too; the comment renderer must read it (no per-finding link place existed at the base).
  Store tests (recorded CouchDB, no listener) and both negative controls ran green/red; the review package could not build on this host, so its wiring and tests are UNRUN there (module fetch refused).
- last commit: 8b985ba [couchdb-evidence-viewer] tests: records settings in sample mode
- done: the owner's order is built and committed in three commits (21e4608 store, 44e2614 review, 8fd5ad1 docs)
  - views return the sample scenario's records: TestViewsReturnTheSampleScenarioRecords, map functions run in node over the stored documents
  - viewer served to the read-only account and a write with it refused: TestViewerIsServedToTheReadOnlyAccountAndNeedsNoWrite
  - link only when the setting is set: TestEvidenceURLExistsOnlyWhenTheBaseIsSet (store) and TestEvidenceRecordLinkExistsOnlyWhenTheBaseURLIsSet (review, UNRUN here)
  - HTML in a value is shown as text: TestViewerEscapesRecordValuesAsText runs the page's esc function in node
- next / blocked-on: the master runs the review package, settingspage and the listener-bound store tests, coordinates with the CLI track so the comment renderer prints Finding.EvidenceRecordURL, and removes any Co-authored-by line if present
- artifacts:
  - internal/store/records.go
  - internal/store/recordsettings.go
  - internal/store/records_viewer.html
  - internal/store/records_test.go
  - internal/review/records.go
  - internal/review/records_test.go
- evidence:
  - REPRODUCE (acceptance 1), by reading, not by running a review: at base 74d46e7 `git grep -n "_bulk_docs|MethodPut|MethodPost" -- internal` finds CouchDB writes only in internal/store/couchdb.go:83 and :109, inside CouchDB.Register, which writes SOURCE tables; nothing in internal/review writes a Result. A sample review could not be run here (see the build line below), so the claim is read from code only.
  - `go build ./...` exit 1: go.opentelemetry.io/████████████ and siblings cannot be fetched (tls: failed to verify certificate); internal/telemetry, so internal/review, internal/settingspage and everything above them cannot build here. NOT caused by this leg.
  - `go vet ./internal/store ./internal/redact` exit 0; `gofmt -l internal cmd` empty.
  - `go test -count=1 ./internal/store` exit 1 on the base's own listener tests: `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` (TestStoresAgreeOnOneScenario first; the panic ends the package run).
  - `go test -count=1 -skip 'TestStoresAgreeOnOneScenario|TestLocalRunDoesNotDial|TestCouchDB|TestStoresAgreeOnAMissingTable' ./internal/store` exit 0 (ok). UNRUN for the listener refusal: those four names, which include every old test built on newFakeCouch(t) in couchdb_test.go, couchdb_nomango_test.go and missing_test.go.
  - `go test -count=1 ./internal/redact ./internal/envdoc ./internal/querytpl` exit 0.
  - The new store tests need no listener: the recorded CouchDB is an http.RoundTripper over httptest.NewRecorder, a departure from "httptest server" so the tests run where a socket cannot be bound. Two of them skip by name when node is missing (views, esc); node v26.7.0 was present, so both ran.
  - internal/review/records.go and records_test.go could not be built in the review package; the same files with a stub of the types they read were built and tested in ████████████████████████████████████████ (exit 0, ok). open.go and review.go edits are NOT type-checked anywhere; the master must build them.
  - NEGATIVE CONTROL (a): in internal/store/recordsettings.go EvidenceURL, `if base == "" || database == "" || id == ""` changed to `if database == "" || id == ""`; `go test -run EvidenceURL ./internal/store` FAILED with `unset base gave a link: "/proven_records/_design/evidence/viewer.html?doc=42-abc-run-f1"`; restored (grep shows the original line at recordsettings.go:140). The review-level twin mutates `if rc.linkBase == "" { return }` in internal/review/records.go publish; not run (package cannot build here).
  - NEGATIVE CONTROL (b): in internal/store/recordsettings.go LoadRecords, `if user == env(EnvCouchUser) {` changed to `if false && user == env(EnvCouchUser) {`; `go test -run LoadRecords ./internal/store` FAILED with `the_writing_user_is_the_read-only_user: accepted: <nil>`; restored (recordsettings.go:100). It is the line that keeps the writing account from ever being the viewer's account.
  - Not run: the settingspage drift test (CONFIGURATION rows were generated by printing store.RecordsEnvironment through a scratch program under ████████████, then pasted; the catalog's evidenceAfter and the page's sentence were edited identically); no real CouchDB.
- note: master-was-wrong: (1) The claim "if the comment renderer already has a per-finding place for a link, use it" had no place to use: no comment renderer exists at this base (grep for a PR comment finds none), so the field Finding.EvidenceRecordURL is exposed and the CLI track must print it. (2) The order's "httptest" recorded CouchDB could not bind a listener on this host; the tests use an in-process RoundTripper instead. Nothing else in the readings was refuted. Judgement calls: sample mode reads the four new settings (the lab review is a sample review; cost if wrong: one rule in the sample-mode sentence); an enabled records setting with a missing writer is a refusal, not a down store; a link is dropped when the read-only account cannot fetch the viewer page; a record id carries the run stamp so re-reviews add records; the writing account needs database admin rights to publish the design document (documented). The commit tool printed `packed-refs.lock: Operation not permitted` after each commit; git log shows all three landed. No Co-authored-by line was added by this leg.
- updated: 2026-10-25T08:00:00Z

CLOSE-OK · couchdb-evidence-viewer · commit 8b985ba · branch fix/couchdb-evidence-viewer · worktree ███████████████████████████████████████████████████████████████████████████████

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 8fd5ad1111bd4e0f36bfec3d7b5b8fed55218f90

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| e-mail or user name | 1 |
| seat identity | 1 |
