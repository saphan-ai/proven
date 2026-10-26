# REVIEW — the-query-seam-carries-the-template · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T03:48:17Z

## Summary

Named review of W52 at 692da97 on claude-code / claude-opus-5-5: all four properties hold on every path read; no query is built from model text, each store builds its own form.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-query-seam-carries-the-template

### Ground
- Reviewer: claude-code, model claude-opus-5-5, 2026-10-25, headless review leg, sandboxed (no loopback listeners, no module proxy).
- HEAD measured `692da97` (`git rev-parse --short=7 HEAD`) = reviewed-sha. Diff judged: `git diff main...692da97` (three-dot), 23 files, +536/−252.
- The master's self-measurement (full suite green on macm5; open.go:162 mutant red) was set aside as a claim (R1); every reading below is my own.
- `go build ./...` — exit 0.
- `go vet ./...` — exit 1, ONLY because `internal/cli/sarifschema` test dependency `github.com/santhosh-tekuri/jsonschema/█████████` could not be downloaded (`tls: failed to verify certificate: x509: OSStatus -26276`), sandbox network; not this diff. `go vet` on every touched package (`./internal/query/... ./internal/store/... ./internal/review/... ./internal/pipeline/... ./internal/osv/... ./internal/epss/... ./internal/agent/... ./internal/settingspage/... ./internal/querytpl/... ./cmd/...`) — exit 0, no output.
- `go test -count=1 ./...` — exit 1. Every non-ok package fails ONLY with `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` (cmd/livecheck, generator, guardrail, pipeline, query, review, store) or the sarifschema setup failure above. All other packages ok, including demo, demo/live/review, cmd/proven, cmd/settings-doc, internal/cli, epss, osv, agent/investigator, querytpl, settingspage, source.
- Because a listener panic aborts its package, the touched packages were re-run with only listener tests skipped:
  - `go test -count=1 ./internal/store/ ./internal/query/ -skip "CouchDB|StoresAgree|LiveSelect|Unauthorized|ClosedEndpoint|NonSelect|LocalRunDoesNotDial"` — ok, ok.
  - `go test -count=1 ./internal/pipeline/ -skip "Unsafe|RecordedFixturePrompt|ClosedGuardrail|SafeGuardrail|LivePipelineWithoutAGuardrail|GuardrailOff"` — ok (68.8s).
  - `go test -count=1 ./internal/review/ -skip "TestReviewEvidenceNamesTheFormTheStoreRan/couchdb|LiveReview|Unauthorized|ClosedEndpoint|Echoed|LiveModeWithoutRuntime"` — ok (271.7s).

### Differential
Property by property, every path read (callers and guards named):

1. **No path builds a query from model text — HOLDS.** `grep -rn "\.Query(ctx\|query\.Adapter\|querytpl\.Build(\|source\.Build(" internal cmd` after the diff: the only SQL-text calls left are inside stores (`internal/store/zetaris.go:70`, `internal/store/local.go:64`, `internal/query/sample.go:50`) each fed `forms.SQL` from `querytpl.BuildForms`, plus `cmd/livecheck/main.go:189` (operator view setting, pre-existing, outside diff). Investigator (`investigator.go:78,116`), EPSS (`epss.go:155`), OSV (`check.go:210`) and the OSV→KEV alias lookup (`check.go:307`) now pass `querytpl.Request` + `querytpl.Relations` through `query.Runner`. Every store builds from validated values: Zetaris `zetaris.go:66`, Local `local.go:56`, CouchDB `couchdb.go:133` (Mango built from `forms.Find`, never parsed from SQL; a source-registered template with no Mango form is named down, `couchdb.go:142`). `querytpl.Build` (`querytpl.go:250–292`) holds every value to `source.CheckValue` before any builder runs, so the dropped `source.ValueError` branch in `osv/check.go` loses nothing: the same check now arrives as `*Refusal{ReasonBadValue}` and lands in the same "osv query refused" note.
2. **Evidence names the form that ran, the text sent and its rows — HOLDS on every store by code reading.** `storeRunner.Run` (`pipeline.go:1107–1120`) copies `ev.Query/Form/Engine/Snapshot/Columns/Rows`; `queryEvidence` (`osv/check.go:390`), EPSS (`epss.go:200`), Investigator (`investigator.go:140,151`) copy `res.SQL/Form/Engine`. CouchDB reports `FormMango` + `forms.Mango` (`couchdb.go:190`), Zetaris/Local `FormSQL` + `forms.SQL` (`zetaris.go:78`). Review-level test exists for couchdb + local only (see Remainder).
3. **Sample mode and fixtures unchanged — HOLDS.** Sample path: `pipeline.go:1142` `query.OpenSample` → `Sample.Run` builds the same `forms.SQL` the callers used to build with `querytpl.Build`/`source.Build`. No fixture, scenario or prompt file is in the diffstat. demo, cmd/proven, internal/cli (sample goldens) and cmd/settings-doc (generated-doc freshness) green. The two doc edits (docs/ADAPTERS.md, docs/CONFIGURATION.md + `store/settings.go`, `settingspage/catalog.go`) replace only the now-false sentence "couchdb is down on every query until the seam passes templates"; generator and page agree (settings-doc green). Judged as required, not as a behaviour change outside the order.
4. **Local file opened only after the inside-the-target check; wrong CouchDB value stops Open — HOLDS.** `lazyLocal` opens only inside `Run`/`Register` via `sync.Once` (`open.go:258–281`); nothing in the pipeline calls `Register` on the live store (`grep liveStore internal/pipeline` → only `storeRunner{p.liveStore}`), and `runner.Review` runs `checkLocalFile` first (`open.go:375`). `store.LoadCouch` non-Down error still returns `storeConfig(cerr)` (`open.go:159`).

### Unplug
Mutations run here, each a temporary edit restored with `git checkout --` (worktree confirmed clean after):
- **A** `pipeline.go:1116` `Form: ev.Form` → `Form: ""` → `TestReviewEvidenceNamesTheFormTheStoreRan/local` RED (`evidence form "" engine "local", want "sql" "local"`). Killed.
- **B** `open.go:375` inside-the-target check removed → `TestLocalFileInsideTheTargetIsRefused` RED (`plain: want a refusal naming PROVEN_EVIDENCE_FILE, got <nil>`). Killed.
- **C** `open.go:159` `return nil, storeConfig(cerr)` → install a downStore → `TestCouchDBWrongValueIsRefusedByName` RED (`███████████████████████: <nil>`). Killed.
- **D** `open.go:296–299` `downStore.Run` refusal-first `BuildForms` check removed → `go test ./internal/review/ -run "CouchDB|Zetaris|Missing|Empty|Down|Refus" -skip <listener tests>` ok (136.5s). **SURVIVED** — no offline test pins "a bad template is the same answer whether or not the store is up" for the down store.
- **Master's mutant** (`open.go:162` NewCouchDB → down store; claimed red on `TestReviewEvidenceNamesTheFormTheStoreRan`): UNRUN here — the couchdb subtest needs an httptest listener (see Unrun).

### Closure
Inputs that reach the seam from outside the code: the model's template name and params (Investigator), the skeptic's refutation request (`Execute`), manifest package names (OSV), CVE aliases from OSV rows (KEV), and settings (relation names, CouchDB URL/credentials). All flow as `querytpl.Request` values; the template name is resolved through the registry and an unknown name is a `*Refusal` that does not echo the name (`querytpl.go:256`). Values are checked per parameter format before any builder; relation names checked by `ValidRelation` in both SQL and Mango builders. The CouchDB request body is marshalled from structured `Find` data, not from text. No path concatenates model text into a query.

### Remainder
Owed, non-blocking (R6), each with location and remedy:
1. **Missing test (mutant D survived)** — `internal/review/open.go:296` `downStore.Run`. Remedy: a unit test that a `downStore` given a refused request (e.g. unknown template, bad CVE value) returns `*querytpl.Refusal`, not `*service.Down`; mutated field: the `BuildForms` call.
2. **Refusal ordering differs between the two "no store" shapes** — `internal/pipeline/pipeline.go:1108` `storeRunner.Run` with a nil store returns `service.Down{zetaris}` without validating first, whereas `downStore.Run` and `investigator.Execute` refuse a bad request first. Reachable only through `pipeline.Build` with no `LiveStore` in live mode (Review's `open` always sets one). Remedy: validate with `querytpl.Validate`/`BuildForms` before naming the store down, or document why not.
3. **Review-level Zetaris form not pinned** — `internal/review/form_seam_test.go:30` covers couchdb and local; Zetaris is covered only at store level (`TestStoresAgreeOnOneScenario`, listener). Remedy: a `zetaris` subtest with a recorded Zetaris server asserting `form sql engine <zetaris>` and the advisory row, so replacing `open.go:172` `store.NewZetaris` turns red.
4. **"SQL:" label carries Mango text on CouchDB** — `internal/osv/check.go:237` and `missNote` `check.go:365` print `SQL: <res.SQL>`, which on CouchDB is the Mango JSON. Remedy: label with `res.Form` (e.g. `query (mango): …`). Sample notes unaffected (form is sql), so goldens would not move.
5. **Dead file** — `████████████████████████████████████` is now a package clause plus a comment. Remedy: delete it; the pointer belongs in the commit message.
6. **Unreachable mapping** — `epss.go:161`, `check.go:220`, `check.go:313` map `ReasonUnknownTemplate` to "template is not registered" errors; for built-in templates this cannot fire and no test exercises it. Acceptable as a guard; noted for completeness.

Declined to judge (R11): licences (order: out of scope); ADR-0047 text itself; `cmd/livecheck` SQL composition (pre-existing, outside diff).

### Profile
security-adjacent data-access profile (R9: query construction from model output is a trust boundary — ADR-0032). S1 boundary map in Closure; S3-equivalent controls mutated (B, C) and killed; S4: no dependency manifest change in the diff (`go.mod`/`go.sum` untouched), no secrets added (test credential `couch-test-password` is a declared test value against an in-process recorded server), no invisible/bidi characters found in the diff by reading. backend B1: CouchDB `_find` carries `limit: RowLimit` and the reader caps rows (`couchdb.go:150,174`) — bounded.

### Focus
The four properties of the order. Concentrated on: every former `q.Query(sql)` call site and its replacement; whether `source.Build`'s value check survived the move to `querytpl.Build` (it did); whether CouchDB could ever receive SQL (it cannot — sql-only templates are named down); the lazy local open ordering; the couchdb configuration error path.

### Conditions
Verdict `accept-with-conditions`. None of the conditions is blocking; all four properties hold on every path read. Conditions, owed before or alongside the next wagon:
1. Add the downStore refusal-first test (Remainder 1) — closes surviving mutant D.
2. Add the review-level Zetaris form subtest (Remainder 3).
3. The master runs the listener-dependent tests listed under Unrun outside this sandbox and confirms them green at 692da97, including the master's open.go:162 mutant on the couchdb subtest.
Items Remainder 2, 4, 5 are owed quality work, not conditions.

### Unrun
- All tests that start an `httptest` server — refused by the sandbox: `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. In the touched packages: `internal/review` `TestReviewEvidenceNamesTheFormTheStoreRan/couchdb`, `TestLiveReviewCarriesTheRoutedModel`, `TestLiveReviewRefusesANonNVIDIAModelByName`, `TestUnauthorizedReviewDegrades`, `TestClosedEndpointDegradesTheReview`, `TestZetarisUnauthorizedReviewCompletes`, `TestZetarisClosedEndpointDegradesTheReview`, `TestRuntimeTokenEchoedInTheReceiptIsNotWritten`, `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten`, `TestOrgIDEchoedInTheReceiptIsNotWritten`, `TestLiveModeWithoutRuntimeNamesProverOff`; `internal/store` CouchDB*, `TestStoresAgreeOnOneScenario`, `TestStoresAgreeOnAMissingTable`, `TestLocalRunDoesNotDial`; `internal/query` live_test listener tests; `internal/pipeline` guardrail_test listener tests; `cmd/livecheck`, `internal/generator`, `internal/guardrail` (abort on first listener panic).
- The master's mutant `open.go:162` → down store (same refusal).
- `internal/cli/sarifschema` tests and its `go vet` — module download refused: `tls: failed to verify certificate: x509: OSStatus -26276`.

## The review status as filed

## REVIEW — the-query-seam-carries-the-template
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Named review of W52 at 692da97 on claude-code / claude-opus-5-5: all four properties hold on every path read; no query is built from model text, each store builds its own form.
  Mutants killed: storeRunner form drop, inside-target check removal, couchdb wrong-value acceptance. Survivor: downStore refusal-first check has no test.
  Conditions: add downStore refusal test, add review-level zetaris form subtest, master re-runs httptest-dependent tests (sandbox refused listeners) incl. open.go:162 mutant.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T03:40:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 1 |
| private credential or variable name | 1 |
