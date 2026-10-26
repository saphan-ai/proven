# REVIEW — the-evidence-store-interface · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T07:59:19Z

## Summary

Named review of W48 (60f7088...5403de4) on claude-code / claude-opus-5-5: build+vet green; no model text reaches either store; couchdb/unknown refused (mutant killed); Zetaris unchanged.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-evidence-store-interface

Reviewer: claude-code, model claude-opus-5-5, review leg `the-evidence-store-interface-review`, 2026-10-24.
Diff judged: `60f7088...5403de4` (three-dot), HEAD verified `5403de4` (`git rev-parse --short=7 HEAD`).
Commits: `0942f07` (leg: store behind one interface) and `5403de4` (master's absorb of main).

### Ground

- `go build ./...`: clean (exit 0, no output).
- `go vet ./...`: clean (exit 0, no output).
- `go test -count=1 ./...`: every package that opens no local listener passes. The packages that start an
  `httptest` server panic in this sandbox before a single assertion runs (see Unrun). The new tests that need
  no listener were run on their own and pass:
  `go test -count=1 -v -run "TestLocalFileKeepsARegisteredTable|TestOpenSampleIsTheLocalFile|TestLoadDefaultsToZetaris|TestLoadLocalReadsTheFileAndRefusesABrokenPath|TestBuildForms" ./internal/query ./internal/store ./internal/querytpl`
  → 5 PASS, 3 packages `ok`.
- Read: `internal/store/{store,local,zetaris,settings}.go`, `internal/store/{conform,settings}_test.go`,
  `internal/query/{file,file_test,sample,live,query,engine}.go`, `internal/querytpl/forms{,_test}.go`,
  `internal/review/open{,_test}.go`, and the callers of the query adapter: `internal/agent/investigator/investigator.go:67-139`,
  `internal/osv/check.go:214-220,297-307`, `internal/epss/epss.go:156-159`, `internal/pipeline/pipeline.go:602-619,780-816,1225-1237`,
  `internal/cli/run.go:171-189`.

### Differential

Order question by order question.

1. **Can a model-chosen string reach the local store's SQL other than as a validated value?** No.
   `investigator.go:57-61` refuses any answer that carries SQL (`ReasonModelSQL`); `investigator.go:100` and `:70`
   build the text with `querytpl.Build`, which refuses (never escapes) a value outside its format. `Local.Run`
   (`local.go:60`) goes through `querytpl.BuildForms`, which is `Build` (`forms.go:24`). The pipeline's local path
   is `Local.Query` (`local.go:74`) fed by the same built SQL. The local engine accepts exactly one SELECT
   (`engine.go:85-97`) on an in-memory copy, so even a foreign statement could not change the file. Table and
   column names written by `Register` pass `ValidRelation`/`plainIdent` (`file.go:159,171`), and cells are quoted
   with the single quote doubled (`file.go:218`); that path is stored data, not model input, and has no production
   caller today (see Remainder). Note: the Zetaris path keeps its second gate `oneSelect` (`live.go:213`); the local
   path has none. It does not need one for safety (the engine refuses anything but one SELECT), but the
   defence-in-depth sentence in `engine.go:9-11` ("Live checks the text once more") is now true of one store only.
   **Is the file opened read-only for queries?** Yes for queries: `openFile` reads the file once into memory
   (`file.go:54-62`) and `File.Query` (`file.go:124-152`) never writes. Opening is not read-only: `CreateFile`
   writes an empty file when none exists (`file.go:39-44`), and `review/open.go:127` calls it on every live local
   review. **Where does it live?** See Condition C2: the default is the relative `proven-evidence.sql`
   (`settings.go:28`), resolved against the process's working directory, which is where `cli/run.go:172` also
   roots the run; nothing keeps it out of the reviewed tree (`Request.Target`, `review.go:79`).
2. **Does the conformance test compare the evidence in full and run on both stores?** Yes, by reading:
   `conform_test.go:87-104` compares engine (expected to differ), form, query text, columns, rows and snapshot,
   and runs the one scenario on `Local` (real file) and `Zetaris` (recorded `httptest` server). The snapshot
   comparison is weaker than it looks: both stores take the instant from the same injected clock
   (`conform_test.go:42,64`), so the comparison proves the field is plumbed, not that the stores agree on a
   source's snapshot. `Table.Snapshot` is accepted by `Register` on both stores and then discarded by the local
   one (`local.go:51` never passes it; the file stores no snapshot) — see Remainder R2. **Not executed here**:
   the test starts an `httptest` listener (see Unrun).
3. **Is Zetaris behaviour unchanged?** Yes in substance. `internal/query/live_test.go` and every other existing
   Zetaris test are untouched by the diff (`git diff --stat` lists no `live_test.go`). `live.go` gains only
   `Form: querytpl.FormSQL` on the result (`live.go:189`) and a test-only `SetClock` (`live.go:60-65`).
   `review/open.go:136-142` is the old Zetaris block moved under an `else`, byte-equivalent. Observable additions
   that touch a Zetaris user: (a) the text line `evidence query zetaris snapshot=… form=sql` (`cli/format.go:308`)
   gains a suffix; (b) the trace JSON gains `"form"` (`trace.go:110`, `omitempty`); (c) an unknown
   `PROVEN_EVIDENCE_STORE` value now refuses a review that used to run on Zetaris. All three are what the order
   asked for.
4. **Build, vet, test**: see Ground and Unrun.

### Unplug

- Mutant run here: `internal/store/settings.go` — the `couchdb` case folded into the `zetaris` case (so naming
  couchdb would silently run Zetaris). `go test -count=1 -run TestLoad ./internal/store` →
  `--- FAIL: TestLoadDefaultsToZetaris … settings_test.go:24: <nil>`. File restored with
  `git checkout -- internal/store/settings.go`; `git status --porcelain` shows no source edit.
- Read, not run (each guard inverted on paper against the test that names it):
  `file.go:91` duplicate-register guard → `TestLocalFileKeepsARegisteredTable:39-48` fails;
  `file.go:218` quote doubling → same test fails to reload at `:59`; `sample.go` engine name → 
  `TestOpenSampleIsTheLocalFile:90` fails; `settings.go:102-105` newline refusal → `settings_test.go:77` fails.
- **`TestLocalRunDoesNotDial` cannot fail when its property is false.** It closes an `httptest` server that the
  local store was never given (`conform_test.go:110-113`), so a local store that dialled any other address would
  still pass. The master's control (dropping a row turns it red) reddens it through `len(got.Rows) != 1`
  (`:133`), i.e. it measures row count, not the absence of a dial. The air-gap holds today by construction
  (`store/local.go` and `query/file.go` import no `net`), and no test guards that construction — see Remainder R1.
- The master's measured negative control (`TestStoresAgreeOnOneScenario` red on a dropped row) could not be
  reproduced here: the test needs a listener (Unrun).

### Closure

Trust boundaries the diff touches:
- **Model → query text.** Closed by `querytpl` (refusal, never escape) and the SQL-carrying-answer refusal; no new
  path from model text to either store. Verified by reading the three callers (investigator, osv, epss).
- **Operator → `PROVEN_EVIDENCE_STORE` / `PROVEN_EVIDENCE_FILE`.** Validated at load (ADR-0013): closed set of
  store names, couchdb refused by name, file path a single line of ≤4096 bytes with no control byte. Any path is
  accepted (absolute or relative) — the operator names the file. Refusal messages echo the value clipped to 128
  bytes (`settings.go:137-142`); the clip is by byte and can split a UTF-8 rune in the printed refusal (cosmetic).
- **Local store → network.** No dial by construction; not guarded by a test (Unplug, R1).
- **Local store → filesystem.** Creates `0600` (`file.go:42`), writes via temp + rename (`file.go:109-116`).
  The temp name is fixed (`<path>.tmp`), so two processes registering into one file race; no production caller
  registers today.

### Remainder

Owed, non-blocking (R6):
- **R1** — `TestLocalRunDoesNotDial` (`internal/store/conform_test.go:107-136`) measures row count, not the absence
  of a dial. Remedy: replace it with an import guard (for example `go list -deps` / `go/packages` over
  `internal/store` minus `zetaris.go` and over `query/file.go`, asserting no `net` or `net/http`), or route the local
  store through a dialer that fails the test when called.
- **R2** — `Table.Snapshot` is accepted and dropped by the local store (`internal/store/local.go:51`), and the
  evidence's snapshot on both stores is the wall-clock time of the read, not the source's snapshot date the
  `Table` doc (`store.go:31`) promises to carry. Remedy: either persist the snapshot with the table and return it,
  or say in `store.go:31-45` that the read time is what `Snapshot` means.
- **R3** — The `Store` interface (`Register`/`Run`) and `Zetaris` have no production caller; the pipeline uses
  `Local` only through `Local.Query` (`review/open.go:132`). The conformance test therefore proves `Run`, not the
  path a live review takes. Remedy: route the pipeline through `Store.Run` in the next wagon, or extend the
  conformance test to `Query` with built SQL.
- **R4** — `engine.go:9-11` and `store.go` doc: "Live checks the text once more" is true of Zetaris only; the local
  path has no `oneSelect`. Remedy: one sentence, or run `oneSelect` in `File.Query` for symmetry.
- **R5** — fixed `<path>.tmp` in `file.go:109`; use `os.CreateTemp` in the same directory.

Declined to judge (R11): licences (out of scope by order); performance of loading a large local file into memory
on every open (no size claim made); the CLI text-line suffix `form=…` as a compatibility break for line-exact
readers (the order asked for the form to be named).

### Profile

Security (R9: data access and configuration). S1 boundaries in Closure. S2/S3: no authorization check is added
or removed by this diff. S4: no dependency change (`go.mod`/`go.sum` untouched by the diff); no secret introduced;
no destructive operation on a derived path (the only write is create/rename of the operator-named file); diff
scanned by reading — no invisible or bidirectional characters noticed in the new files. Backend B1: the local
engine returns every matching row with no row limit of its own (Zetaris caps at `RowLimit`, `live.go:169`) —
noted, not blocking, since the file is operator-authored. B4: a failed local query is not observable as a named
service (Condition C1).

### Focus

The property: every query built from a template with validated values — **holds**. Evidence names the form and
carries rows and snapshot — **holds** (snapshot is read time, R2). Local store air-gapped — **holds by
construction, unguarded** (R1). `PROVEN_EVIDENCE_STORE` defaults to Zetaris, refuses an unknown value and
`couchdb` by name — **holds**, mutant-killed. **A missing table is a named down source, never an empty answer —
does NOT hold for the local store** (C1). Two stores give the same evidence for the same scenario — holds for the
happy path by reading; diverges on a missing table (C1). Zetaris behaviour unchanged — **holds**.

### Conditions

- **C1 (blocking) — a missing table on the local store aborts the review instead of naming a down source.**
  Location: `internal/query/engine.go:149-151` and `:456-458` return `fmt.Errorf("query: no table %s")`;
  `internal/query/file.go:136-139` passes it through unchanged; `internal/store/local.go:74-79` likewise.
  Trigger: `PROVEN_EVIDENCE_STORE=local` with a file that lacks the table a template names — which is every live
  local review today, because `review/open.go:127` → `CreateFile` creates an EMPTY file when none exists
  (`file.go:39-44`) and nothing in production calls `Register`. Effect, read through the callers:
  `pipeline.go:1225-1237` (`soften`) treats only a `*service.Down` as degraded; any other error is fatal, so
  `osv.Check` (`check.go:219`), `investigator.Investigate` (`:115-116` → `pipeline.go:610-614`) and `epss.Attach`
  (`epss.go:157-158`) end the review with `query: no table …`. The same missing view on Zetaris comes back as a
  non-200 → `down("http NNN")` (`live.go:146-148`) and the review goes on naming `zetaris`. So the stores diverge
  (R13), and the order's property "a missing table is a named down source" fails for `local`. No test covers a
  missing table on either store. Remedy: in the configured local store (engine `local`), return
  `&service.Down{Service: "local", Reason: "no table <name>"}` (or a typed not-found error that `File.Query` maps to
  `service.Down`), keep `local-sample` behaviour as decided by the master, and add a test that opens an empty
  local store, runs a template, and asserts `service.As(err)` succeeds — negative control: the down conversion
  removed, the test red. Add the same scenario to `TestStoresAgreeOnOneScenario`'s family so both stores are held
  to one answer.
- **C2 (blocking) — the default local-store file is created in the working directory, which is the reviewed tree
  in the ordinary invocation.** Location: `internal/store/settings.go:25-28` (`defaultFile = "proven-evidence.sql"`,
  relative), `internal/review/open.go:127`, `internal/query/file.go:39-44`. Trigger: `proven` run at a repository
  root (the same directory `cli/run.go:172-179` roots `logs/` in) with `PROVEN_EVIDENCE_STORE=local` and no
  `PROVEN_EVIDENCE_FILE`. Effect: an untracked `proven-evidence.sql` is written into the checkout under review; the
  order states "inside the reviewed tree is not allowed". Remedy: either require `PROVEN_EVIDENCE_FILE` when the
  store is `local` (refuse by name when empty), or default it to a path outside the target (e.g. a user data
  directory), and refuse a path that resolves inside `Request.Target`; test with the mutated field
  `PROVEN_EVIDENCE_FILE` empty / pointing inside the target.

### Unrun

The sandbox refuses local TCP listeners. Quoted refusal, from `go test -count=1 ./...`:
`panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`.
Packages whose run aborted on that panic (every later test in the package is unrun with it):
- `cmd/livecheck` (first panic in `TestAllServicesReady`)
- `internal/generator` (`TestLiveReadsTheProvisionedStream`)
- `internal/guardrail` (`TestServerPostsTheMeasuredRequest`)
- `internal/pipeline` (`TestUnsafePromptStopsTheStepAndNamesCategories`)
- `internal/query` (`TestLiveSelectReadsRowsFromTheSQLEditor`) — the two new `file_test.go` tests were run alone and pass
- `internal/review` (`TestLiveReviewCarriesTheRoutedModel`) — the two new tests `TestCouchDBIsRefusedByName` and
  `TestLocalStoreNeedsNoZetaris` were run alone (`go test -count=1 -v -run … ./internal/review`) and pass
- `internal/store` (`TestStoresAgreeOnOneScenario`, panicked at 0.03s) — the two settings tests were run alone and pass
Every other package in `go test -count=1 ./...` reported `ok` (or `[no test files]`).
Specifically unrun among the diff's own controls: `TestStoresAgreeOnOneScenario` and `TestLocalRunDoesNotDial`
(both start `httptest`), so the master's 2026-10-24 measurement of them (green, and red on a dropped row) is not
re-measured here, neither confirmed nor refuted.

## The review status as filed

## REVIEW — the-evidence-store-interface
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Named review of W48 (60f7088...5403de4) on claude-code / claude-opus-5-5: build+vet green; no model text reaches either store; couchdb/unknown refused (mutant killed); Zetaris unchanged.
  C1 blocking: a missing table on the local store is a plain error, so soften() aborts the review instead of naming a down source (Zetaris names it) — and a fresh local file has no tables.
  C2 blocking: default proven-evidence.sql is relative to cwd, i.e. written inside the reviewed tree.
  Unrun: httptest packages (sandbox bind refused), incl. TestStoresAgreeOnOneScenario/TestLocalRunDoesNotDial; NVIDIA second review still owed.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T09:00:00Z

## REDACTED

none
