# STATUS — couchdb-evidence-viewer · round 2

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: df945ca [couchdb-evidence-viewer] merge: absorb main
- updated: 2026-10-25T12:04:21Z

## Summary

Round 2 landed in five commits (F1 to F6). The evidence design document now carries a validate_doc_update that admits only server and database admins, and the fake models membership-with-write. JSON findings carry evidence_record_url, omitted when empty. Viewer rendering, the sample-mode read set, a wrong records database at Open, and an unchecked viewer link each have a test. UNRUN here: every package that imports telemetry (review, cli, cmd/proven, settingspage, pipeline), because the otel module download is refused; the master runs those.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — couchdb-evidence-viewer
- state: stop-2-ready
- abstract: Round 2 landed in five commits (F1 to F6). The evidence design document now carries a validate_doc_update that admits only server and database admins, and the fake models membership-with-write. JSON findings carry evidence_record_url, omitted when empty. Viewer rendering, the sample-mode read set, a wrong records database at Open, and an unchecked viewer link each have a test. UNRUN here: every package that imports telemetry (review, cli, cmd/proven, settingspage, pipeline), because the otel module download is refused; the master runs those.
- last commit: df945ca [couchdb-evidence-viewer] merge: absorb main
- note: master-was-wrong: nothing refuted on F1, F2, F3 and F6. Two corrections to the order's wording. F4: the remedy "no `<b` survives" cannot hold, since the page writes its own `<b>` for levels and yes/no; the test wraps every record field in `<i>`, a tag the page never writes, and asserts none survives plus the title's `<b onmouseover` element. F1: a member also cannot create the database or write a design document, so the fake refuses those for a member, as CouchDB does.
- done: F1, F2, F3, F4, F5, F6 addressed in code, docs and tests
  - F1 ADDRESSED. Property: only a server or database admin writes to the records database. Test TestReadOnlyAccountCannotWriteARecord (member refused, admin admitted) and TestValidateDocUpdateAdmitsOnlyAdmins (8 cases). Reproduced red at 8b985ba before the fix: `the read-only account's write was admitted: map[id:overwrite-1 ok:true]`, exit 1. Mutated line `"validate_doc_update": designVDU,` in internal/store/records.go deleted: exit 1, same message; fixed: exit 0. Second mutation, the `names.indexOf(userCtx.name)` check replaced by false: TestValidateDocUpdateAdmitsOnlyAdmins exit 1 (admin by name refused). The fake's comment is corrected and the fake runs the stored function in Node. docs/ADAPTERS.md tells the administrator to add the reader as a `_security` member and the writer as a database admin. The design version digest now covers the function, so an existing design document is rewritten.
  - F2 ADDRESSED (master's ruling a). internal/cli/format_json.go carries `evidence_record_url` with omitempty; docs/OUTPUT_FORMATS.md and the contract test TestJSONDocumentIsTheContract updated (present on F1, omitted on F2, exact comparison). The "renderer reads that field" sentence in ADAPTERS.md is replaced. Text and SARIF unchanged. UNRUN: the reproduction (JSON without the field) and the revert (drop the EvidenceRecordURL line in findingFrom, test goes red by the exact comparison) could not be executed, since internal/cli does not build here.
  - F3 ADDRESSED. One sentence of the read set (records and link variables, then, with records set, URL USER PASSWORD PREFIX ROW_LIMIT TIMEOUT RECORDS_USER ████████████████, and GITHUB_REF) in ADR-0047 point 7, docs/ADAPTERS.md, docs/CONFIGURATION.md and settingspage/catalog.go. Test TestSampleReviewWithRecordsReadsExactlyTheDocumentedSet pins the 11 names. UNRUN (internal/review).
  - F4 ADDRESSED. TestViewerShowsEveryRecordValueAsText runs showRecord and showReview in Node with a stub document, location and fetch. Mutation `esc(d.title)` to `d.title`: exit 1. Mutation `esc(v.path)` to `v.path`: exit 1. Restored: exit 0.
  - F5 ADDRESSED. TestOpenRefusesAWrongRecordsDatabaseByName: Open with PROVEN_COUCHDB_RECORDS=Not_Valid returns a ConfigError on that field. Mutated line `return nil, rcerr` in open.go. UNRUN (internal/review).
  - F6 ADDRESSED. publish names a failed service couchdb, `viewer: link not checked, the read-only account is not configured`, when records and a link base are set and the reader is not. TestUncheckedViewerIsNamed; mutated line the `rc.reader == nil` branch in internal/review/records.go. UNRUN (internal/review).
- evidence:
  - `go build ./...` and `go vet ./...`: refused for every package importing telemetry: `go.opentelemetry.io/████████████: Get "https://proxy.golang.org/go.opentelemetry.io/otel/@v/v1.47.0.zip": tls: failed to verify certificate: x509: OSStatus -26276`.
  - `go test -count=1 ./...`: `[setup failed]` with that refusal on cmd/guardcount, cmd/proven, cmd/settings-doc, demo/live/review, examples/formatter, internal/cli, internal/cli/sarifschema, internal/pipeline, internal/review, internal/settingspage, internal/telemetry, internal/tool. These are UNRUN.
  - Listener refusals, not from this diff (`httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`): internal/store (TestStoresAgreeOnOneScenario, TestLocalRunDoesNotDial, TestCouchDBNamesItselfDownForATemplateWithNoMangoForm and others, each a panic that aborts the package run), internal/guardrail, internal/query, internal/generator, cmd/livecheck. I did not run a clean baseline of these at 8b985ba beyond the review's own record of the store one.
  - Passing: all 14 records tests of internal/store run by name (`-run 'TestViews|TestPublish|TestViewer|TestReadOnly|TestValidate|TestCheckViewer|TestLoadRecords|TestEvidenceURL|TestQueryKeys'`, ok), with Node installed so none skipped; internal/redact, internal/change, demo, agents and the other buildable packages ok.
  - Substitute for the unbuildable packages: a go/types check of internal/review, its external test package and internal/cli with unresolvable imports tolerated reported no error in the changed files (cli shows only the Linux-only syscall names of connect_linux_test.go). It cannot see a type that depends on telemetry, so it is weaker than a build.
- next / blocked-on: the master runs internal/review, internal/cli, cmd/proven and internal/settingspage where the module cache is reachable and the listener is allowed.
- artifacts:
  - internal/store/records.go
  - internal/store/records_test.go
  - internal/cli/format_json.go
  - internal/cli/output_contract_test.go
  - internal/review/records.go
  - internal/review/records_test.go
  - internal/review/open.go
  - internal/review/open_test.go
  - internal/settingspage/catalog.go
  - docs/ADAPTERS.md
  - docs/CONFIGURATION.md
  - docs/OUTPUT_FORMATS.md
  - docs/adr/ADR-0047-the-evidence-store-is-an-adapter.md
- updated: 2026-10-25T12:30:00Z

CLOSE-OK · couchdb-evidence-viewer · commit df945ca · branch fix/couchdb-evidence-viewer · worktree ███████████████████████████████████████████████████████████████████████████████

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 3d8c8c71d892546a57d109b7689e2aaeaa728199

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| private credential or variable name | 1 |
| seat identity | 1 |
