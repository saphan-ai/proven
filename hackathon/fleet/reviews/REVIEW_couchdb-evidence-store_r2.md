# REVIEW — couchdb-evidence-store · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T23:29:37Z

## Summary

claude-opus-5-5 re-review of W49 round 2 at ab6b44b: C1, C2 and C3 are all ADDRESSED. C1 and C3 each turned red under their negative control (the restored clip(raw); a Query that returns nil), and both were restored afterwards.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — couchdb-evidence-store

### Ground
- Reviewer: claude-opus-5-5, headless REVIEW leg, 2026-10-25, worktree `wt-couchdb-evidence-store-review`.
- `git rev-parse --short=7 HEAD` → `ab6b44b` (the reviewed SHA named by the order). `ae8167d` is an ancestor of it.
- Scope (R14): the fix diff `ae8167d..ab6b44b`. The stream's own commit is `6ac1f30` and touches 6 files: `internal/store/couchsettings.go`, `internal/store/settings.go`, `internal/store/settings_test.go`, `internal/settingspage/catalog.go`, `docs/CONFIGURATION.md` and the new `████████████████████████████████████`. The other commits in the range come from main (the formatter docs/example, SECURITY.md, and the docs/README conflict resolution in `177bba6`) and arrive through the absorb `ab6b44b`.
- `go build ./...` passed. `go vet ./...` and the full `go test ./...` could not complete in this sandbox (see Unrun).
- Targeted tests at `ab6b44b`, all PASS: `TestLoadCouchReadsTheSettings`, `TestLoadCouchEmptyIsADownStoreAndWrongIsRefused`, `TestEnvironmentListsEveryCouchSetting` (internal/store); `TestCouchDBLiveReviewNamesCouchDBDown`, `TestCouchDBSQLAdapterNamesCouchDBDown`, `TestCouchDBIsAccepted`, `TestCouchDBWrongValueIsRefusedByName`, `TestCouchDBEmptyAddressIsADownStore` (internal/review); `internal/settingspage` package ok (the generated-page check).
- 40 packages report `ok` in this sandbox (`go test -count=1 ./... | grep -cE '^ok'` → 40), including `cmd/settings-doc`, `internal/querytpl`, `internal/redact`, `internal/cli` and `examples/formatter`.

### Differential
- C1: `internal/store/couchsettings.go:172-175` sets `shown := clip(raw)` and blanks it when `raw` contains `@`. `:180-182` also blanks it when `url.Parse` fails. Both refusals (`:177`, `:184`) now pass `shown` instead of `clip(raw)`. In net/url, user info can only appear before an `@`, so every address that can carry user info is refused with an empty Value. This covers the length/whitespace branch as well as the parse branch.
- C1 test: `internal/store/settings_test.go:111-115` asserts that the refusal of `http://u:p@c` contains neither `u:p` nor `p@` in `err.Error()`, and that `cfg.Value` does not contain `p@`.
- C2: `internal/store/settings.go:34` (the PROVEN_EVIDENCE_STORE description), `internal/settingspage/catalog.go:44` (`evidenceAfter`) and the generated `docs/CONFIGURATION.md:54,72` now say that with `couchdb` a live review names couchdb down on every query until the query seam passes templates. They point to docs/ADAPTERS.md, which says the same at `docs/ADAPTERS.md:142`. The sentence "its evidence is a Mango query" now reads "the store runs a Mango query, but a live review does not reach it yet". The generated page and its source agree, because `internal/settingspage` and `cmd/settings-doc` pass.
- C3: `██████████████████████████████████████████` calls Open with `PROVEN_EVIDENCE_STORE=couchdb` and valid URL, user and password, then runs a live review over the `twoCommits` fixture. It asserts a `FailedServices` entry with `Service=="couchdb"` and `Reason==couchdbSQLReason`. The reason check is what makes this the query-seam path. The settings-down path (`open.go:153-155`) carries the LoadCouch reason, never `couchdbSQLReason`, so only the adapter installed at `open.go:159` can satisfy the test. A second test, `:43-48`, pins the adapter itself.
- No behaviour outside the conditions changed. The zetaris and local branches of `open.go` are not touched by `6ac1f30` (R13: the paired paths are unchanged).

### Unplug
- C1, mutated field `PROVEN_COUCHDB_URL`. I neutralised the guard at `couchsettings.go:173` (`if false && …`), which restores `clip(raw)` as the Value. `TestLoadCouchEmptyIsADownStoreAndWrongIsRefused` went RED: `settings_test.go:114: the refusal printed the user info: configuration refused: PROVEN_COUCHDB_URL="http://u:p@c": not a base URL; …`. Restored.
- C3, mutated field: the adapter's return value. I changed `open.go:278` to `return query.Result{}, nil`. `TestCouchDBLiveReviewNamesCouchDBDown` went RED (`couchdb_seam_test.go:37: want couchdb down with the seam reason, got []review.ServiceFailure{}`), and so did `TestCouchDBSQLAdapterNamesCouchDBDown` (`:46: want a down couchdb, got <nil>`). Restored.
- After both restorations, `git status --porcelain` shows no tracked change.
- The master's claim (law #44: C1 negative control red at `ab6b44b`) is confirmed by an independent mutation. The master's "full suite green" claim is NOT confirmed here; see Unrun.

### Closure
- C1 — ADDRESSED. The guard turns red when removed (Unplug).
- C2 — ADDRESSED. All three surfaces named by the condition carry the sentence, and the generated page matches its source (the settingspage check passes).
- C3 — ADDRESSED. The test asserts the named service and the seam reason, and it turns red under the order's own negative control.

### Remainder
- Owed, non-blocking (R3/R6): the parse-error blanking at `couchsettings.go:180-182` survives mutation. No test fails when it is removed, because an address without `@` cannot carry user info, so it protects nothing the `@` guard does not already cover. Its cost is that a malformed address with no `@` is refused with an empty Value, so the operator loses the echo of an address that held no secret. Remedy: drop it, or keep it and add a one-line comment that it is defensive only. Either is acceptable; this does not open a round.
- Observation, not a finding: the C1 test asserts on `u:p` and `p@` only. An address like `http://user@host` (user name, no password) is refused with an empty Value by the same `@` guard. That case is not pinned by a test, but the same mutation would turn the existing test red.

### Profile
- security (R9: configuration, credential echo). S1 boundary: the operator's environment → the refusal text printed to the operator and the trace. The disclosure that was blocking is closed by C1. S3: the guard was removed once and a test failed (Unplug). S4: the fix diff adds no dependency (the `go.mod` of the stream commit is untouched) and no secret appears in the diff. The test password `couch-test-password` is a fixture value. No destructive path is involved.
- docs (C2): D4. The command quoted in the docs is not re-run, because none is added; the settings-page generation check passes.

### Focus
- Property 4: a refused `PROVEN_COUCHDB_URL` never carries user info. It HOLDS at `ab6b44b`.
- The live review with couchdb names couchdb down with the seam reason and never reads as "no such row". It HOLDS, and it is pinned by a test with a negative control.
- Declined to judge (R11): the Mango store itself (`internal/store/couchdb.go`), the querytpl mango forms and the main-side commits brought in by the absorb are outside the R14 fix-diff scope. The docs/README conflict resolution (`177bba6`) was checked only for leftover conflict markers: none were found by `grep -rnE '^(<<<<<<<|>>>>>>>) ' docs internal examples SECURITY.md README.md`.

### Conditions
- RC1 — owed, verification (master, before the gate): run `go vet ./...` and `go test -count=1 ./...` on `ab6b44b` outside a sandbox that refuses loopback listeners, and record the result. This review could not run them; see Unrun. The master states this was green on macm5 on 2026-10-24, but that was not re-measured here. Under law #53 a non-empty Unrun forbids a plain `accept`, so this is the condition that keeps the verdict at accept-with-conditions.
- RC2 — owed, non-blocking: the parse-error blanking at `internal/store/couchsettings.go:180-182` (Remainder). Drop it or comment it.

### Unrun
- `go vet ./...` and the build of `internal/cli/sarifschema`. Both failed because the module could not be downloaded: `internal/cli/sarifschema/schema_test.go:12:2: github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip": tls: failed to verify certificate: x509: OSStatus -26276` (the sandbox has no network and the module is not in the cache).
- Seven packages abort on the sandbox's listener refusal, which means every test after the panicking one in each package is also unrun: `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review` and `internal/store`. Quoted: `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. Panicking tests seen include `TestCouchDBAgreesWithTheLocalStore`, `TestStoresAgreeOnOneScenario`, `TestLiveReviewCarriesTheRoutedModel`, `TestAllServicesReady`, `TestLiveReadsTheProvisionedStream`, `TestLiveSelectReadsRowsFromTheSQLEditor`, `TestServerPostsTheMeasuredRequest` and `TestUnsafePromptStopsTheStepAndNamesCategories`. In these packages no assertion failure was observed, only the bind panic. The condition-bearing tests in `internal/store` and `internal/review` were run by name and pass (Ground).

## The review status as filed

## REVIEW — couchdb-evidence-store
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: claude-opus-5-5 re-review of W49 round 2 at ab6b44b: C1, C2 and C3 are all ADDRESSED. C1 and C3 each turned red under their negative control (the restored clip(raw); a Query that returns nil), and both were restored afterwards.
  The condition tests pass and the build is green. The full suite and vet are Unrun here (sandbox: loopback bind refused, module fetch refused), so RC1 asks the master to run them before the gate (law #53).
  Owed, non-blocking RC2: the parse-error blanking at couchsettings.go:180-182 survives mutation.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T00:00:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 2 |
