# REVIEW — couchdb-evidence-viewer · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T11:54:21Z

## Summary

W58 round-2 re-review on claude-opus-5-5 (claude-code) at 3d8c8c7: F1-F6 all ADDRESSED; the VDU admits only server/db admins and its removal or weakening turns TestReadOnlyAccountCannotWriteARecord red (measured).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — couchdb-evidence-viewer

Named re-review of W58 round 2. Reviewer: claude-code, claude-opus-5-5. HEAD `3d8c8c7` (checked with
`git rev-parse --short=7 HEAD`). Scope under R14: the fix diff `8b985ba..3d8c8c7`. Its first-parent stream commits are
47005ee, ca29a5c, b63fdc3, c4213b9 and f9073bd, followed by the absorb of main, 3d8c8c7. The stream's own change is
`git diff 8b985ba..f9073bd`: 13 files, +393/−32. The absorb brings in only other streams' merged work, so this review did
not judge it.

### Ground

Measured on 2026-10-25 by this reviewer, in this leg's sandbox:

- `go build ./... && go vet ./... && go test -count=1 ./...` exits 1. Building `internal/telemetry` needs
  `go.opentelemetry.io/████████████`, which is not in GOMODCACHE (`███████████████████████████████████`). The download
  is refused with: `tls: failed to verify certificate: x509: OSStatus -26276`. `internal/review` depends on it, with 9
  otel packages in its dependency list, and so do `internal/cli` and `internal/settingspage`. All three fail with
  `[setup failed]`.
- In `internal/store`, the whole package aborts on the first httptest listener:
  `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. This happens in
  TestStoresAgreeOnOneScenario, and in TestLocalRunDoesNotDial when that test is skipped.
- `records_test.go` does not need a listener, because its fake is an in-process RoundTripper over `httptest.NewRecorder`.
  I ran all 14 of its tests by name, and all 14 PASS. Node is installed, so no test skipped and the VDU and viewer tests
  really ran.
- `go vet ./internal/store/` is clean.
- The master's measurement (full suite green on macm5, two unplugs red) is a CLAIM. I re-measured the two store-side
  unplugs myself; see Unplug. I could not re-measure the two review-side claims; see Unrun.

### Differential

What the round-2 commits change, judged against the round-1 conditions:

- **F1. `47005ee`**
  - `internal/store/records.go:79-100` adds `designVDU`. It admits a write when `userCtx.roles` holds `_admin`, or the
    user's name is in `secObj.admins.names`, or one of the user's roles is in `secObj.admins.roles`. It throws
    `{forbidden: …}` for every other write.
  - The function is written into the design document at `records.go:423`. It is also hashed into `designVersion` at
    `records.go:384`, so a database that already holds the round-1 design document is rewritten on the next publish.
  - `records_test.go:22-30, 94-104, 153-163, 184-225`: the fake no longer refuses every non-GET from the reader. It now
    lets a member POST `_bulk_docs` for non-design documents. It runs the stored VDU in Node with the real security
    object: the writer is in admins and the reader in members.
  - `docs/ADAPTERS.md:187` tells the administrator to make the reader a member and the writer a database admin.
- **F2. `ca29a5c`**
  - `format_json.go:89-91, 206-207` adds `evidence_record_url,omitempty`.
  - `output_contract_test.go:71-74, 150` sets the field on F1 and leaves it unset on F2. `jsonEqual` compares the whole
    document, so the key must be present on one finding and absent on the other.
  - `OUTPUT_FORMATS.md:9, 45` documents the key as the one exception to "always present", and says text and SARIF do
    not carry it.
- **F3. `f9073bd`**: ADR-0047 point 7, `ADAPTERS.md:191`, `CONFIGURATION.md:76` and `catalog.go:48` now state the same
  set of variables. The catalog constant and CONFIGURATION.md are byte-identical in that sentence. `open_test.go:128-171`
  pins the set with an exact-equality check over the sorted names it reads.
- **F4. `b63fdc3`**: `records_test.go:708-805` runs the page's own `<script>` in Node with stub
  `document`/`location`/`fetch`. It calls `showRecord` and `showReview`, with every text field wrapped in `<i>` and the
  title carrying `<b onmouseover="x">`.
- **F5**: `open_test.go:173-193` calls `Open` with `PROVEN_COUCHDB_RECORDS=Not_Valid` in sample mode and asserts
  `*review.ConfigError` with `Field == PROVEN_COUCHDB_RECORDS`.
- **F6. `c4213b9`**: at `review/records.go:65-69`, when `linkBase` is set and `reader` is nil, the review appends the
  failed service `couchdb` with the reason "viewer: link not checked, the read-only account is not configured".
  `records_test.go:162-181` is TestUncheckedViewerIsNamed.

The diff changes no behaviour the conditions did not ask for. The removed write assertion in
TestViewerIsServedToTheReadOnlyAccountAndNeedsNoWrite was not dropped: it moved into
TestReadOnlyAccountCannotWriteARecord, which is stricter. That makes it a relocation, not a weakening of the bar.

### Unplug

I ran each control below by editing the file, running the test, and restoring the file with `git checkout -- <file>`.
`git status --porcelain` afterwards showed nothing of mine.

| mutation | test | result |
|---|---|---|
| `records.go`: delete the `"validate_doc_update": designVDU,` entry | TestReadOnlyAccountCannotWriteARecord | **RED** — `records_test.go:475: the read-only account's write was admitted: map[id:overwrite-1 ok:true]` |
| `records.go` VDU: `secObj.admins` → `secObj.members` | TestReadOnlyAccountCannotWriteARecord, TestValidateDocUpdateAdmitsOnlyAdmins | **RED, both** — the writer is refused (3 of 3 records refused), and the cases "member" and "member by role" are admitted while "admin by name" and "admin by role" are refused |
| `records_viewer.html:96`: `esc(d.title)` → `d.title` | TestViewerShowsEveryRecordValueAsText | **RED** — "showRecord wrote markup from a record" and "did not show the title as text" |
| `open.go:255` `return nil, rcerr` → no-op | TestOpenRefusesAWrongRecordsDatabaseByName | **UNRUN here** (the package does not build in this sandbox). The master claims it goes red. |
| `review/records.go:65` delete the `reader == nil` branch | TestUncheckedViewerIsNamed | **UNRUN here** (same cause) |
| `format_json.go:207` drop `EvidenceRecordURL:` | TestJSONDocumentIsTheContract | **UNRUN here** (same cause). By reading: `jsonEqual` over the whole map must fail on the missing key. |

### Closure

- **F1: ADDRESSED.**
  - The VDU itself is correct for CouchDB's semantics. `userCtx.roles` carries `_admin` for a server admin. A database
    admin is recognised only through `secObj.admins`, and the VDU checks it by name and by role. CouchDB runs a VDU on
    every write path: PUT, POST, `_bulk_docs`, deletes and replication. So the one function covers paths the fake does not
    model. The anonymous user and a missing security object are refused, and both are tested.
  - The test fails when the VDU is removed or weakened; both controls above are red.
- **F2: ADDRESSED**, under the master's ruling (a). JSON carries the field, and OUTPUT_FORMATS.md and the contract test
  are updated, with presence and absence both pinned. The text and SARIF decision is stated in the docs.
- **F3: ADDRESSED.** The sentences agree in all three places (plus the catalog that renders CONFIGURATION.md). The
  pinning test asserts exact equality, so a variable missing from the list and an extra variable both fail it. Its run is
  UNRUN here.
- **F4: ADDRESSED.** The test renders through the page's real functions and is red when one `esc` is removed. It
  exceeds the remedy: it covers every field, not only the title.
- **F5: ADDRESSED** by reading. The test is exactly the remedy and names the field. Its run and its unplug are UNRUN here.
- **F6: ADDRESSED** by reading, with the remedy's first option: a named failed service. Its run is UNRUN here. See
  Remainder R-2 on the wording of the reason.

### Remainder

Owed and non-blocking (R6, R14): none of these opens another round.

- **R-1 (owed, test fidelity).**
  - Problem: the fake's header comment (`records_test.go:22-30`) says a member "may write any document that is not a
    design document". The fake refuses every member method other than GET and POST at `records_test.go:101`, so a
    member's PUT of a record never reaches the VDU. Only `_bulk_docs` exercises it. The Recorder writes only through
    `_bulk_docs`, so the test does guard the path the product uses.
  - Remedy: narrow the comment to "may POST `_bulk_docs`", or route a member PUT through `f.validate`.
- **R-2 (owed, wording).**
  - Problem: `review/records.go:68` says "the read-only account is not configured" whenever `store.LoadCouch` fails
    (`open.go:300`). That includes a set-but-malformed `PROVEN_COUCHDB_TIMEOUT` or `ROW_LIMIT`, and in that case Open
    swallows the config error silently.
  - Remedy: carry `cerr`'s field into the reason, so it reads "not configured or set wrongly: <field>".
- **R-3 (owed, docs scope).**
  - Problem: `docs/ADAPTERS.md:191` says "the pull request comment is rendered from that output". No renderer in this
    repository reads `evidence_record_url`; I grepped `*.go/*.md/*.yml/*.js/*.sh`. Under the master's ruling that rendering
    belongs to the CLI track, the sentence describes a consumer that does not exist yet.
  - Remedy: name the CLI-track stream that will render it, or say "a renderer may read it".

### Profile

security, because the diff touches authorization: who may write a record. Docs (D1, D4) also applies.

- **S1, the trust boundary.** The read-only account is a database member, and the viewer runs under that account in a
  browser. The asset is the integrity of the records history. The boundary is now held server-side by the VDU, not by
  the client, and not by an assumption about membership.
- **S3.** I removed the authorization check once and weakened it once; a test failed both times (see Unplug).
- **S4.** No dependency delta in the stream's own commits (I read `git diff --stat 8b985ba..f9073bd`). No secrets
  appear in the diff: the test passwords are fixtures. I saw no destructive operations, and no bidirectional or invisible
  characters in the read hunks.

### Focus

The property that must hold is "a record in the records database can be written only by a database admin or a server
admin". It holds in the VDU's code, and in the tests over eight security-object shapes. Second, every record value
reaches the viewer's HTML as text: this holds in a rendered run of the page's own script. Declared not judged (R11): the
absorbed main commits (other streams, already merged), and licences, which the order puts out of scope.

### Conditions

Verdict **accept-with-conditions**. Both former blocking findings are fixed and proven by red controls. All four owed
items are addressed. The single condition comes from what this sandbox could not run:

- **C1 (owed, before the gate).** The master runs `go build ./... && go vet ./... && go test -count=1 ./...` at
  `3d8c8c7` outside the sandbox and records the result, together with three unplugs:
  - `open.go:255` → no-op, with TestOpenRefusesAWrongRecordsDatabaseByName red.
  - `review/records.go:65` branch removed, with TestUncheckedViewerIsNamed red.
  - `format_json.go:207` field dropped, with TestJSONDocumentIsTheContract red.

  The master's note already claims the first of these and the suite. C1 asks that the other two land with it.

R-1 to R-3 above are owed follow-ups and are not conditions of this verdict.

### Unrun

- `go build ./...`, `go vet ./...` and `go test -count=1 ./...` for every package that imports `internal/telemetry`.
  This includes `internal/review`, `internal/cli` and `internal/settingspage`, whose result was `[setup failed]`. The
  quoted refusal: `go.opentelemetry.io/████████████: Get "https://proxy.golang.org/go.opentelemetry.io/otel/@v/v1.47.0.zip": tls: failed to verify certificate: x509: OSStatus -26276`.
- So the following are not run by this reviewer: TestSampleReviewWithRecordsReadsExactlyTheDocumentedSet (F3),
  TestOpenRefusesAWrongRecordsDatabaseByName (F5), TestUncheckedViewerIsNamed (F6), TestJSONDocumentIsTheContract (F2),
  and their unplugs.
- `internal/store` tests that open an httptest listener, among them TestStoresAgreeOnOneScenario and
  TestLocalRunDoesNotDial. Quoted refusal: `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`.
- `gofmt -l` was denied by the role's permission class (`"gofmt" is NOT in the ██████████████ list of role review`).
  Formatting is not checked, and I did not rebuild it from another command (law #57).

## The review status as filed

## REVIEW — couchdb-evidence-viewer
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: W58 round-2 re-review on claude-opus-5-5 (claude-code) at 3d8c8c7: F1-F6 all ADDRESSED; the VDU admits only server/db admins and its removal or weakening turns TestReadOnlyAccountCannotWriteARecord red (measured).
  F4 viewer test red when one esc() is dropped (measured). 14/14 records tests pass with Node.
  Condition C1: review/cli/settingspage do not build here (otel download refused, x509) and httptest listeners are refused, so the master records the full suite plus the F5/F6/F2 unplugs before the gate.
  Owed, non-blocking: R-1 fake comment vs member PUT, R-2 unchecked-link reason wording, R-3 ADAPTERS names a PR-comment renderer that does not exist yet.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T12:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 2 |
| policy binding name | 1 |
