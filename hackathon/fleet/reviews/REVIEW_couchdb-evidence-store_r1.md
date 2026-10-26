# REVIEW — couchdb-evidence-store · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T15:34:51Z

## Summary

Named review of W49 at ae8167d on claude-opus-5-5: properties 1, 2, 3 and 5 hold by reading and by the pure-logic tests (PASS); property 4 fails on one path.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — couchdb-evidence-store

Reviewer: REVIEW leg `couchdb-evidence-store-review`, model claude-opus-5-5, 2026-10-24.
Diff judged: `main...ae8167d` (three-dot), two commits (5d41f71, ae8167d). HEAD verified `ae8167d`.
Profile: security (credentials, a new network client, a new query language) plus backend (a new store).

### Ground

- `git rev-parse --short=7 HEAD` → `ae8167d` (ROLE-LOCK met).
- `go build ./...` → exit 0, no output.
- `go vet ./...` → FAILED AT SETUP, not on a finding: `internal/cli/sarifschema/schema_test.go:12:2: github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/...": Forbidden` (the sandbox has no module proxy; the package is outside this diff). `go vet` over every changed package (`store querytpl review redact query pipeline settingspage`) → exit 0, no output.
- `go test -count=1 ./...` → FAIL in 8 packages, every one of them the sandbox and not the code: 7 by `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` (cmd/livecheck, generator, guardrail, pipeline, query, review, store) and `internal/cli/sarifschema [setup failed]` (the module fetch above). All other packages `ok`, including querytpl, redact and settingspage.
- Pure-logic tests of the diff, run by name, all PASS: `TestCouchDBDatabaseNames`, `TestCouchCellKeepsNumbersAndRefusesNesting`, `TestLoadDefaultsToZetaris`, `TestLoadCouchReadsTheSettings`, `TestLoadCouchEmptyIsADownStoreAndWrongIsRefused`, `TestEnvironmentListsEveryCouchSetting`, `TestBuildFormsFillsBothForms`, `TestMangoSelectorHoldsTheValuesTheSQLFormMatches`, `TestBuildFormsRefusesWhatBuildRefuses`, `TestCouchDBIsAccepted`, `TestCouchDBWrongValueIsRefusedByName`, `TestCouchDBEmptyAddressIsADownStore`, `TestInternalAndCmdDoNotImportBannedPackages`.
- The master's claim "full suite green at ae8167d" is NOT refuted and NOT confirmed here: every listener-bound test is UNRUN (see Unrun). The claim "the 404 branch at internal/store/couchdb.go:156 returns the down error" is CONFIRMED by reading: `couchdb.go:156-159` returns `couchFailed("no table "+table)`, a `*service.Down{Service:"couchdb"}`. The claim "the password name is in redact.Names" is CONFIRMED: `internal/redact/redact.go:32` appends `store.EnvCouchPassword`.

### Differential

Per property, FILE·LINE·CONTROL, and the test that turns red when the control is removed.

1. **Validated values reach the selector and the database name — HOLDS.**
   - Values: `querytpl/querytpl.go:247` `checked()` is now the one gate both languages call; `literal()` (`:261`) wraps it for SQL, `buildFind` (`querytpl/mango.go:39,52,69`) calls `osvKey`/`checked` itself, so it does not lean on `Build` having run. The selector is `map[string]string` with fixed keys and marshalled by `encoding/json` (`store/couchdb.go:142`), so a value is always a JSON string under implicit `$eq`: an operator in a value is inert even if a format were loosened. Red test: `TestBuildFormsRefusesWhatBuildRefuses` calls `buildFind` directly with `CVE-2021-1"} {"$or":[` (forms_test.go) and fails if the Mango builder's own check is removed; `TestCouchDBSendsOnlyTheBuiltSelector` asserts zero requests reach the server for a hostile value and that the exact body sent is `{selector,fields,limit}` (listener-bound, UNRUN here).
   - Database name: relation re-checked by `ValidRelation` in `buildFind` (`mango.go:36,49,66`), table checked by `checkTable` on Register (`couchdb.go:79` → `local.go:85`), prefix by `prefixFormat` (`couchsettings.go:41,120`). `ValidRelation` admits only `[A-Za-z_][A-Za-z0-9_]*` with at most one `.`, so `/`, `%`, `?`, `..` cannot reach the URL path (`couchdb.go:150`). Red tests: `TestBuildFormsRefusesWhatBuildRefuses` (`a b`, `a.b.c`), `TestCouchDBRegisterRefusesASecondLoad` (`x;drop`), `TestLoadCouchEmptyIsADownStoreAndWrongIsRefused` (`Bad-Prefix`).
2. **Same columns, same order as the SQL form — HOLDS.** The field lists live once (`mango.go:13-15`, `OSVPackageColumns`); `Run` builds each row by iterating `forms.Find.Fields` and indexing the document (`couchdb.go:175-182`), never trusting the answer's key order. Red tests: `TestBuildFormsFillsBothForms` compares `Find.Fields` with the parsed SELECT list of the SQL text for all four templates (a reordered list fails it — PASS here); `TestCouchDBAgreesWithTheLocalStore` and `TestCouchDBAppearsInTheConformanceOfStores` compare columns and rows across stores against a fake that answers fields in REVERSE order (listener-bound, UNRUN).
3. **A missing database is a named down store — HOLDS** on Run (`couchdb.go:156-159`). Register's non-2xx/412 also maps to `couchDown` (`:91-92,113-114`). Red test: `TestCouchDBMissingDatabaseIsADownStore` (UNRUN, listener-bound; its docstring names the exact negative control).
4. **The password never reaches a log, evidence or error; the URL in an error carries no credential — DOES NOT HOLD on one path.** See Condition C1.
   - Holds: `classify` drops the transport error text (`couchdb.go:236-247`); `couchDown` never reads the body (`:251-256`) — the fake server deliberately echoes the Authorization header in its 401 body and `TestCouchDBFailuresAreNamedAndKeepTheCredentialOut` asserts it does not surface; redirects are not followed (`:45`); the BaseURL refuses user info (`couchsettings.go:173`); user and password refusals print no value (`:126,132`), tested; the password is in `redact.Names` (`redact.go:32`) and `EnvCouchPassword` is `Secret: true` on the page (`couchsettings.go:70`, tested); a short password that the redactor could not safely hide is refused (`:131`).
   - Fails: `couchsettings.go:173-174` — a `PROVEN_COUCHDB_URL` WITH user info (`http://admin:s3cret@couch:5984`) is refused, correctly, but the refusal carries `clip(raw)` as its Value, and `store.ConfigError.Error()` (`store/settings.go:60`) and `review.ConfigError.Error()` (`review/review.go:52`) print it with `%q`. `review.Open` returns it before any redactor runs (`review/open.go:157` → `storeConfig`, `open.go:591-597`), and the redactor would not know a URL-embedded password anyway (it holds only the value of `███████████████████████`). The leg's own test pins the refusal (`settings_test.go`, `http://u:p@c` → Field `PROVEN_COUCHDB_URL`) but never asserts the value is withheld. This is exactly the path property 4 names.
5. **The net/http exemption is one file with its reason — HOLDS.** `pipeline/imports_test.go:35-37` keys the exemption on `internal/store/couchdb.go` alone with a one-line reason; the census refuses any other file importing `net/http` and refuses an unused exemption (`:125-131`). `couchsettings.go` imports `net/url` only, which is not banned. `TestInternalAndCmdDoNotImportBannedPackages` → PASS.

**Ruling 1 (the live path names couchdb down on every query).** Honest at RUNTIME, not honest on every READER SURFACE.
- Runtime: `review/open.go:148-160` installs `couchdbSQL`, whose `Query` (`open.go:277-279`) returns `*service.Down{Service:"couchdb", Reason: couchdbSQLReason}`; the pipeline's `soften`/`note` (`pipeline/pipeline.go:1340-1362`) records it as a failed service with the reason, so the review degrades and says why. No other store is read; no Zetaris client is built (`TestCouchDBIsAccepted` checks `query.LiveBuilt()`). The OSV and EPSS checks run through the same adapter and degrade the same way. Nothing else on the live path is wrong because of it.
- `docs/ADAPTERS.md` states the limit plainly ("it is not yet on the review's query path").
- But the settings surfaces do not: `store/settings.go` (PROVEN_EVIDENCE_STORE row: "which evidence store a live review uses: `zetaris`, `local` or `couchdb`"), its generated twin in `docs/CONFIGURATION.md`, and `settingspage/catalog.go:44` (`evidenceAfter`: "its evidence is a Mango query") tell an operator that a live review uses CouchDB and produces Mango evidence. On the live path no review ever produces Mango evidence. See C2.
- And no test turns red if `couchdbSQL.Query` stops returning the down error: `grep -rln couchdbSQL internal cmd` names only `review/open.go`; the three new open tests stop at `Open`. Returning `query.Result{}, nil` there would turn every live CouchDB review into "the source holds no such row" (ADR-0002's hazard) with the suite green. See C3.

### Unplug

| Mechanism | Unplugged how | What turns red |
|---|---|---|
| `checked()` in `buildFind` | drop the format check | `TestBuildFormsRefusesWhatBuildRefuses` (direct `buildFind` call) — PASS here, so it is live |
| Field order | reorder `kevColumns` | `TestBuildFormsFillsBothForms` |
| Row built by field name | iterate the doc map instead | `TestCouchDBAgreesWithTheLocalStore` (fake reverses order) — UNRUN here |
| 404 → down | return empty Evidence | `TestCouchDBMissingDatabaseIsADownStore` — UNRUN here |
| Body kept out of errors | put the body in the reason | `TestCouchDBFailuresAreNamedAndKeepTheCredentialOut` (fake echoes Authorization) — UNRUN here |
| Exemption scope | import net/http in another store file | `TestInternalAndCmdDoNotImportBannedPackages` — PASS |
| URL value withheld on refusal | (does not exist — C1) | **nothing** |
| `couchdbSQL` down on query | return empty result | **nothing** — C3 |

Mutations were derived by reading, not executed: the scratch clone that would have hosted them was refused (see Unrun).

### Closure

The attacker-written member is the request's parameter values (model-proposed) and, secondarily, the relation settings. Both are closed by allow-list formats before either form is built, and the Mango body is built from structure, not from text, so the evidence's `Query` and the bytes sent are one value (`couchdb.go:140-146`, `mango.go:79-83`). The CouchDB answer is the other untrusted input: bounded at 8 MiB (`couchdb.go:24,226`), row-capped (`:169`), nested values refused (`:280-281`), body never echoed. The operator's own environment is trusted for content but not for display — that is where C1 sits.

### Remainder

Owed, non-blocking (R6/R8), each with a remedy:
- R-a `couchdb.go:147-148`: a `json.Marshal` failure is reported as `couchFailed("not configured")`; it cannot happen on a `map[string]string`, but the label is wrong if it ever did. Remedy: name it `"bad query"` or drop the branch with a comment.
- R-b `couchdb.go:97-103`: a registered column named `_id` or `_rev` (both pass `ValidRelation`) overwrites the document id or is refused by CouchDB, so Register would diverge from the local store. Not reachable from the shipped templates. Remedy: refuse a leading-underscore column in CouchDB Register, or prefix the doc keys.
- R-c `review/open.go:159`: with valid CouchDB settings the store is NOT pre-registered in `popts.Down`, so a review that issues no query never mentions that CouchDB was configured and unused. Remedy: also append the down entry at Open, as the empty-settings branch does.
- R-d `couchdb.go:156`: a 404 from a mis-pathed base URL (a proxy subpath) reads as "no table X". Still a down store, so honest in class; the reason could say "no database or wrong address".

**FIX-OR-SHOUT (law #7), outside this diff:** the same echo-the-URL-on-refusal pattern exists at `internal/query/settings.go:191-195` (`ZETARIS_REST_URL`) and `internal/generator/settings.go:317-321`. The CouchDB code copied it. I do not fix code under review; the master should register the two siblings for the same remedy as C1.

### Profile

security + backend.
- S1 trust boundaries: model → template params (closed, property 1); settings → URL/user/password/prefix (closed except C1's display); Proven → CouchDB (Basic auth, no redirects, timeout, bounded read); CouchDB → Proven (bounded, typed, body not echoed).
- S3 auth removal: the fake server refuses a missing or wrong credential (`couchdb_test.go:55-58`), so a store that dropped `SetBasicAuth` fails every listener-bound test.
- S4: no manifest or lockfile change in the diff (`git diff --stat` lists no go.mod/go.sum); no new dependency. Secrets searched: only test fixtures (`couch-test-password`, `a-long-password`). No destructive path on derived file paths. No invisible or bidirectional characters observed in the diff text.
- B1 bounded result: `limit` sent and re-enforced client-side. B3: one `http.Client` per store, no shared mutable state beyond the clock. B4: failures are named down services with short reasons, findable from the review's failed-services list.

### Focus

The order's five properties and ruling 1. Declined to judge (R11), for the master to rule on: licences (out of scope per order); CouchDB index creation and `_find` performance without an index (no index is created; CouchDB warns but answers); the live test `TestCouchDBLive` (needs a real CouchDB, boundary forbids it); the content of ADR-0047 itself.

### Conditions

- **C1 — BLOCKING (property 4).** `internal/store/couchsettings.go:173-174`: a refused `PROVEN_COUCHDB_URL` must not carry user info in its Value. Remedy: when `u.User != nil` (or the parse fails), refuse with Value `""` or the URL with user info replaced (`u.Redacted()` / scheme+host only), and add to `TestLoadCouchEmptyIsADownStoreAndWrongIsRefused` an assertion that `http://u:p@c`'s refusal text contains neither `u:p` nor `p@`. Negative control: restore `clip(raw)` → the assertion fails. Mutated field: `PROVEN_COUCHDB_URL`.
- **C2 — owed, user-facing.** `internal/store/settings.go` (PROVEN_EVIDENCE_STORE description), `internal/settingspage/catalog.go:44` (`evidenceAfter`) and the generated `docs/CONFIGURATION.md`: say that with `couchdb` a live review names couchdb down on every query until the query seam passes templates, as `docs/ADAPTERS.md` already does; drop "its evidence is a Mango query" for the live review or qualify it.
- **C3 — owed, test.** Add a review-level test: Open with `PROVEN_EVIDENCE_STORE=couchdb` and valid settings, run a query through the installed adapter (or a live review over a fixture that asks one), and assert a `service.Down{Service:"couchdb"}` with the seam reason. Negative control: `couchdbSQL.Query` returning `query.Result{}, nil` → the test fails.

### Unrun

- Every listener-bound test, quoted refusal: `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. Packages: `internal/store` (all `TestCouchDB*` with a fake server, `TestStoresAgreeOnOneScenario`, `TestCouchDBAppearsInTheConformanceOfStores`), `internal/review`, `internal/pipeline`, `internal/query`, `internal/generator`, `internal/guardrail`, `cmd/livecheck`.
- `internal/cli/sarifschema` and therefore a whole-tree `go vet ./...`: `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/...": Forbidden`.
- Executed mutations (R3): a scratch clone into the run's artifact directory was refused — `git clone -q --no-hardlinks <worktree> <artifact>/mut` → `This command requires approval`. Mutations in Unplug are derived by reading.
- `TestCouchDBLive`: requires a real CouchDB; forbidden by the order's boundary.

## The review status as filed

## REVIEW — couchdb-evidence-store
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Named review of W49 at ae8167d on claude-opus-5-5: properties 1, 2, 3 and 5 hold by reading and by the pure-logic tests (PASS); property 4 fails on one path.
  C1 BLOCKING: couchsettings.go:173 echoes a PROVEN_COUCHDB_URL carrying user:password in the refusal text. C2/C3 owed: the settings page claims couchdb serves live reviews, and no test guards couchdbSQL's down answer.
  Build green; vet green on changed packages; every httptest-bound test UNRUN (sandbox bind refused), sarifschema setup failed (module proxy Forbidden).
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T15:40:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 2 |
| private credential or variable name | 1 |
