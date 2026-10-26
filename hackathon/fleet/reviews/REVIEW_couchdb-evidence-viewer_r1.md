# REVIEW — couchdb-evidence-viewer · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T11:08:42Z

## Summary

claude-opus-5-5 review of main...8b985ba: reject on two blocking findings. F1: CouchDB members can write non-design docs and the design doc has no validate_doc_update, so the read-only account can write; only the fake refuses. F2: EvidenceRecordURL reaches no formatter, so no reader sees the link. Owed: F3 (sample-mode docs understate the set-path reads), F4/F5 (esc call sites and Open-path tests), F6. Full suite UNRUN: otel module download refused (x509); the 11 new store tests pass.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — couchdb-evidence-viewer

Reviewer: claude-code, model claude-opus-5-5. Reviewed `main...8b985ba` (three-dot), HEAD verified `8b985ba`. Date 2026-10-25.
Profile: security (R9: credentials, a read/write boundary, a page served on the CouchDB origin) plus backend.

### Ground

What I ran, at 8b985ba, in this leg's sandbox:

- `go build ./... && go vet ./... && go test -count=1 ./...`: **did not complete.** The module download for
  `go.opentelemetry.io/otel*@v1.47.0` is refused: `tls: failed to verify certificate: x509: OSStatus -26276`.
  Every package that imports `internal/telemetry` fails at setup. See Unrun.
- `go test -count=1 -v ./internal/store/ -run 'TestViews|TestPublish|TestViewer|TestCheckViewer|TestLoadRecords|TestEvidenceURL|TestQueryKeys'`:
  **PASS, all 11 new store tests.** Node is installed here, so `TestViewsReturnTheSampleScenarioRecords` (the four
  map functions run in Node) and `TestViewerEscapesRecordValuesAsText` really ran and were not skipped.
- `go test ./internal/redact/`: PASS.
- The full `./internal/store/` run fails in the pre-existing `TestStoresAgreeOnOneScenario` with
  `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. That is the sandbox, not this diff.

I read every changed file in full: `internal/store/records.go`, `recordsettings.go`, `records_viewer.html`,
`records_test.go`, `internal/review/records.go`, `records_test.go`, the `open.go`, `review.go`, `couchsettings.go`,
`redact.go` hunks, the five sample-environment guards, the ADR-0047 point 7 hunk, and the `docs/ADAPTERS.md` and
`docs/CONFIGURATION.md` hunks. Callers I read as well: `open.go:45-135` (the sample-mode branch and where the redactor
is built), `couchsettings.go` `LoadCouch` (the keys it reads), and `internal/cli/format.go`, `format_json.go` and
`format_sarif.go` (the output path a finding takes to a reader).
Dependency delta (R4): there are none. `go.mod`/`go.sum` are not in the diff stat.

### Differential

Property by property, with FILE·LINE·CONTROL:

1. **Read-only viewer: partly holds.**
   - Served to and checked with the read-only account: holds. `internal/review/records.go:65-69` runs `CheckViewer` with
     the `store.CouchDB` built from `LoadCouch`, and `internal/store/records.go:322-337` sends only a GET. Control:
     `TestViewerIsServedToTheReadOnlyAccountAndNeedsNoWrite` asserts the single reader GET.
   - The writer is a different user from the reader: holds at load time. `recordsettings.go:100-104`, control:
     `TestLoadRecordsKeepsTheWriterApart`.
   - Never links `/_utils` and loads no external resource: holds. `records_viewer.html` has no `src=`, `<link>` or
     absolute URL. Control: `TestViewerPageIsSelfContainedAndReadOnly` uses a banned-string list and counts
     `method:` == 1 GET.
   - **"A write with that account is refused" holds ONLY IN THE TEST FAKE (finding F1, blocking).** The fake at
     `internal/store/records_test.go:89-92` refuses every non-GET from the reader, and its comment (lines 22-25) says
     "as a CouchDB database's security document makes it". That is not CouchDB's model. In CouchDB, a database
     `_security` **member** can read every document **and create, update and delete every non-design document**. Only
     design documents need admin rights. The design document written at `records.go:391-406` carries **no
     `validate_doc_update`**, and nothing else in the diff restricts writes. `docs/ADAPTERS.md:186` tells the
     administrator to give "the read-only account read access to the records database". In CouchDB the only way to
     do that is membership, and membership includes write. So on a real server the account the viewer is opened with
     can overwrite or delete any evidence record, which also breaks the doc's "the records are a history" claim. The
     page itself only GETs, so exploiting this needs the account's credential or an XSS on the CouchDB origin. But the
     property as ordered does not hold, and the test that claims it measures the fake. Law #52 false-world answer:
     the test passes identically whether or not real CouchDB refuses the write.

2. **Every value shown as text: holds by reading. The control is weak (F4, owed).** Every record-derived value in
   `records_viewer.html:54-120` goes through `esc()`, including both `href` values, which are built from
   `location.pathname` plus `encodeURIComponent` and then escaped. The `innerHTML` sink at line 44 receives only
   escaped strings. No `javascript:` URL can come from a record. `esc` escapes `& < > " '`, which is correct for both
   element and double-quoted attribute contexts. Control: `TestViewerEscapesRecordValuesAsText` runs only the `esc`
   function cut out between its markers. **Dropping `esc(...)` at any single call site (for example
   `esc(d.title)` → `d.title` at line 96) turns no test red.**

3. **The writing credential is never written anywhere: holds on every path I traced.** `Recorder.do`
   (`records.go:427-457`) puts the credential only in BasicAuth. Every error is a fixed `couchFailed`/`couchDown`
   reason. `LoadRecords` refusals print neither the user nor the password (`recordsettings.go:96-108`). The record
   struct has no credential field. `NewRecorder` disables redirects (`records.go:236`), so the credential goes only to
   the configured host. `redact.Names` now includes `EnvRecordsPassword` (`redact.go:32`). Controls:
   `TestPublishOnADownServerIsANamedFailure` and `TestLoadRecordsKeepsTheWriterApart`, which checks the refusal text
   for both passwords and the writer name.
   Caveat (Remainder): in **sample** mode `popts.Redactor` is never built (`open.go:85-120` runs only when
   `!opts.Sample`), so in a sample review with records set the writing password is in no redactor. No output path
   carries it today, so this is not a leak.

4. **Nothing is written and no link is made unless set; a wrong value stops Open: holds in code, but the Open path
   is untested (F5, owed).**
   - `openRecorder` (`open.go:276-304`) returns a nil recorder when `PROVEN_COUCHDB_RECORDS` is empty, and
     `(*recorder).publish` is nil-safe (`review/records.go:47`).
   - A bad value returns `storeConfig(err)` before telemetry or any review runs (`open.go:253-256`).
   - Controls: `TestNoRecorderWritesNothing` and `TestLoadRecordsKeepsTheWriterApart`, both below Open. **No test
     drives `Open` with a bad records value.** Replacing `return nil, rcerr` with a no-op turns nothing red.
   - Silent fallback (F6): `open.go:300-302` swallows a `LoadCouch` error, so with records on and the read-only
     account misconfigured, the link is shown **unchecked**, with no failed service or warning.

5. **Four views and the URL: holds.**
   - The views were run in Node over the documents `Publish` stored, keys asserted per view
     (`TestViewsReturnTheSampleScenarioRecords`). The test also checks that the design document is not emitted.
   - `EvidenceURL` (`recordsettings.go:139-144`) is base + database + the fixed `/_design/evidence/viewer.html` +
     `?doc=<QueryEscape(id)>`. Nothing else goes in: no credential, no user, no query text.
   - `checkBaseURL` refuses user info, query and fragment. Controls: `TestEvidenceURLExistsOnlyWhenTheBaseIsSet` and
     `TestEvidenceRecordLinkExistsOnlyWhenTheBaseURLIsSet` (the latter is in `internal/review`, UNRUN here).

6. **ADR-0047 point 7 and the five guards: the guards are exact; the ADR and docs are not (F3, owed).**
   - All five guards add exactly `k == store.EnvRecordsDB || k == store.EnvEvidenceLink`: equality on two constants,
     with no prefix and no pattern. Locations: `cmd/proven/fix_scenario_test.go:26`, `cmd/proven/main_test.go:304`,
     `cmd/proven/scenario_test.go:243`, `internal/cli/run_test.go:715`, `internal/review/open_test.go:111`. They admit
     those two names and nothing broader.
   - With both unset, `openRecorder` reads exactly those two (`LoadRecords` returns at `recordsettings.go:81-83`,
     `LoadLinkBase` reads one key). So the "with both unset it reads nothing more" half holds.
   - **But all five guards return `""` for both names, so they only ever measure the unset world.** With
     `PROVEN_COUCHDB_RECORDS` set, sample mode also reads:
     - `PROVEN_COUCHDB_URL`, `PROVEN_COUCHDB_RECORDS_USER`, `PROVEN_COUCHDB_USER`, `███████████████████████████████`
       and `PROVEN_COUCHDB_TIMEOUT` (`LoadRecords`);
     - `GITHUB_REF` (`open.go:294`);
     - all six `PROVEN_COUCHDB_*` including `███████████████████████` (`LoadCouch`, `open.go:300`).
   - The three documents disagree with this and with each other:
     - ADR-0047 point 7 names **two** variables;
     - `docs/CONFIGURATION.md` / `settingspage/catalog.go:48` say "the four records and link variables";
     - `docs/ADAPTERS.md:191` says "Sample mode reads these four variables and **no other CouchDB one**". That is
       false: it reads `PROVEN_COUCHDB_URL`, `_USER`, `_PASSWORD`, `_TIMEOUT`, `_PREFIX` and `_ROW_LIMIT`.
   - Reading secrets is needed for the owner's decision (records written in sample mode need a credential), so the
     reads are not the defect. The defect is that the record states a smaller set than the code reads, and no test
     covers the set path.

**F2 — conformance (R2), blocking unless the master rules it out of scope:** the order's "an evidence-record link per
finding" reaches **no reader**. `Finding.EvidenceRecordURL` (`review.go:293-298`) is set (`review/records.go:73`) but
no formatter reads it. `grep -rn EvidenceRecordURL internal cmd docs` finds only the field, its setter, its tests and
`docs/ADAPTERS.md`. `internal/cli/format_json.go`'s `jsonFinding` (lines 77-88) has no such field, and text and SARIF
do not print it. `docs/ADAPTERS.md:191` says "The pull request comment's renderer reads that field". No renderer in
this tree does, and since JSON does not carry it, no out-of-tree renderer fed by Proven's output can either.

### Unplug

Which test turns red when each mechanism is removed. Each answer comes from reading the test against the guard; I did
not run mutants, because the worktree must stay unchanged and the review package does not build here.

| Mechanism removed | Turns red |
|---|---|
| same-user refusal `recordsettings.go:100` | `TestLoadRecordsKeepsTheWriterApart` |
| empty-DB off `recordsettings.go:81` | `TestLoadRecordsKeepsTheWriterApart` (unset case) |
| `linkBase == ""` return `review/records.go:62` | `TestEvidenceRecordLinkExistsOnlyWhenTheBaseURLIsSet/unset` (UNRUN here) |
| refused-record exclusion `records.go:299-310` | `TestPublishLeavesOutARefusedRecordAndNamesIt` |
| `CheckViewer` byte compare `records.go:333` | **nothing**: the fakes serve exactly the page |
| `CheckViewer` call `review/records.go:65-69` | **nothing**: no test serves a missing page through `publish` (`TestCheckViewerNamesAMissingPage` tests `CheckViewer` directly) |
| any one `esc(...)` call site in the page | **nothing** (F4) |
| `return nil, rcerr` in `open.go:254` | **nothing** (F5) |
| design-version skip `records.go:383` | `TestPublishWritesAsTheWritingAccountAndOnce` (designPuts == 1) |
| read-only write refusal on a real CouchDB | **no mechanism exists** (F1); the test asserts the fake's own rule |

### Closure

Attacker-written inputs reach the page from finding titles, messages, query text, rows and tool output. Those come from
the reviewed repository and from model output, so the PR author controls them. All of them reach the DOM only through
`esc`. The `doc`, `pr` and `commit` URL parameters go into `fetch` paths via `encodeURIComponent` or `JSON.stringify`,
and into the DOM only escaped. The page sits on the CouchDB origin, next to `/_utils` and the write API, so an XSS
there would act with whatever credential the browser holds, possibly an admin session. That is why F4's missing
call-site control matters more than its size suggests. I found no unescaped sink. I did not reproduce an exploit,
and the failing-test proof is the remedy named under Conditions.

### Remainder

Declined to judge, or below the bar (R11). Each is for the master to rule on:
- R1: the browser, not the page, picks the account. The page sends whatever credential the browser holds for the
  origin. A reader logged into Fauxton as admin opens it as admin. It still sends only GETs. The docs could say this.
- R2: sample mode builds no redactor (`open.go:85-120`), so the writing password is in none. No output carries it today.
- R3: `RecordID` (`records.go:200-209`) folds distinct finding ids to one document id ("F.1" and "f-1" both become
  `f-1`). In one bulk the second conflicts, and `refused` is keyed by id, so **both** findings lose their link. Low.
- R4: `Recorder.Publish` makes one PUT database plus one GET design document on every review with findings. That is
  bounded and fine.
- R5: the database `_security` is never set by Proven, by design ("the administrator's act"). On CouchDB 3 a new
  database is admin-only, so until an admin acts, `CheckViewer` fails as 401/403 and no link is shown. That is the
  safe failure.

### Profile

Security (S1–S4) and backend (B1, B4).
- S1 trust boundaries:
  - writer → CouchDB: credential in BasicAuth only, redirects off;
  - reader browser → CouchDB origin: the viewer, GET only;
  - PR author → record content: escaped (Closure).
- S3 authorization removal: the reader's write refusal does not exist in code (F1).
- S4: no dependency delta; destructive paths: none (no delete).
- B1: views are key-ranged, and the viewer's `by_review` uses a start/end key range, but the request has no `limit`
  (unbounded rows for one PR+commit; acceptable).
- B4: failures surface as failed service `couchdb` with a `records:` or `viewer:` reason, and carry no secret.

### Focus

The order's six properties, the read/write boundary against CouchDB's real permission model (R13 fake versus real),
the escaping of every sink in the viewer, the set and unset worlds of the sample-mode environment reads, and whether
the link reaches a reader.

### Conditions

Verdict **reject**: two blocking findings. Each finding below carries a location and a named remedy.

- **F1 (blocking, security/correctness)**
  - Location: `internal/store/records.go:391-406` and `records_test.go:22-25,89-92`; docs at `docs/ADAPTERS.md:186`.
  - Remedy: add a `validate_doc_update` to the `evidence` design document. It refuses any write unless the user is a
    server admin (`_admin` role) or is named in, or holds a role of, `secObj.admins`.
  - Test it in Node the way the views are tested: a member is refused, a database admin is admitted.
  - Correct the fake's comment, or make the fake model membership-with-write plus the VDU, so the test can fail when
    the VDU is removed.
  - Tell the administrator to add the reader as a `_security` member and the writer as a database admin.

- **F2 (blocking, conformance)**
  - Location: `internal/cli/format_json.go:77-88`, `format.go`, `format_sarif.go`; docs at `docs/ADAPTERS.md:191`.
  - Remedy, either:
    - (a) emit `evidence_record_url` in the JSON finding (and text/SARIF as the contract decides), with
      `OUTPUT_FORMATS.md` and the output-contract tests updated; or
    - (b) if the master rules output wiring out of W58's scope, delete the "renderer reads that field" sentence and
      name the follow-up stream.

- **F3 (owed)**
  - Location: ADR-0047 point 7, `docs/ADAPTERS.md:191`, `docs/CONFIGURATION.md`/`catalog.go:48`.
  - Remedy: state the set-path reads (the records and read-only `PROVEN_COUCHDB_*` keys and `GITHUB_REF`) in one
    consistent sentence across all three.
  - Add one sample-mode test with `PROVEN_COUCHDB_RECORDS` set that pins the exact set read.

- **F4 (owed)**
  - Location: `records_viewer.html` call sites.
  - Remedy: a test that renders `showRecord`/`showReview` in Node with a stub `document`/`fetch`, fed the sample
    record whose title carries markup. It asserts that no `<b` survives in `out.innerHTML`.

- **F5 (owed)**
  - Location: `internal/review/open.go:253-256`.
  - Remedy: an `Open` test with `PROVEN_COUCHDB_RECORDS=Not_Valid`, asserting a named config error.

- **F6 (owed)**
  - Location: `internal/review/open.go:300-302`.
  - Remedy: when records are on and `LoadCouch` fails, name a failed service or a warning instead of showing the link
    unchecked without a word, or refuse at Open.

### Unrun

- `go build ./...`, `go vet ./...`, `go test -count=1 ./...` (full suite): module download refused:
  `go.opentelemetry.io/████████████: Get "https://proxy.golang.org/go.opentelemetry.io/otel/@v/v1.47.0.zip": tls: failed to verify certificate: x509: OSStatus -26276`.
- `./internal/review` (including `TestEvidenceRecordLinkExistsOnlyWhenTheBaseURLIsSet`, `TestNoRecorderWritesNothing`,
  `TestFailedRecordsAreNamedAndLeaveNoLink`, `TestRecordOfCopiesTheEvidenceAndNamesItsKeys`,
  `TestSampleReviewDoesNotReadTheEnvironment`), `./internal/cli`, `./cmd/proven`, `./internal/pipeline` and
  `./internal/settingspage` all fail with `[setup failed]` for the same refusal.
- `./internal/store` `TestStoresAgreeOnOneScenario` (pre-existing): `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`.
- The invisible/bidi character scan of the diff (S4): refused by the hook, `the Bash command line contains ANSI-C or locale quoting ($' or $") ... [tool-command-line-construct]`.
- check-links: not run (not in an admitted form here).
- Mutation runs (R3): not executed. The Unplug table is derived by reading, because the worktree must be left
  unchanged and the review package does not build here.

## The review status as filed

## REVIEW — couchdb-evidence-viewer
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: claude-opus-5-5 review of main...8b985ba: reject on two blocking findings. F1: CouchDB members can write non-design docs and the design doc has no validate_doc_update, so the read-only account can write; only the fake refuses. F2: EvidenceRecordURL reaches no formatter, so no reader sees the link. Owed: F3 (sample-mode docs understate the set-path reads), F4/F5 (esc call sites and Open-path tests), F6. Full suite UNRUN: otel module download refused (x509); the 11 new store tests pass.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T11:30:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
| private credential or variable name | 2 |
