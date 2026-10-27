# ORDER — couchdb-evidence-store · round 1

- change: couchdb-evidence-store
- round: 1
- base: main @ 5faabc9
- executor: claude-code
- written: 2026-10-24T14:11:09Z
- task: PROVEN_EVIDENCE_STORE=couchdb — each template's Mango-selector form, the store behind the W48 interface, the conformance test covers it (ADR-0047 point 2)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — couchdb-evidence-store
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W49
- review: named
- repo: proven
- base: main @ 5faabc9
- branch: fix/couchdb-evidence-store
- lands-on: main
- backend: claude-code
- quote-usd: 6.00
- cap-usd: 12.00
- wall: 180m
- task: PROVEN_EVIDENCE_STORE=couchdb — each template's Mango-selector form, the store behind the W48 interface, the conformance test covers it (ADR-0047 point 2)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[couchdb-evidence-store] <area>: <what>` included. No AI attribution anywhere.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0047, docs/adr/ADR-0047-the-evidence-store-is-an-adapter.md) — read it whole:
2. Three stores: Zetaris, a local one-file store, and CouchDB. 3. One setting chooses the store
(`PROVEN_EVIDENCE_STORE=zetaris|local|couchdb`); the default stays Zetaris. 4. A template has one form per query
language: SQL for Zetaris and the local store, a Mango selector for CouchDB. The evidence in a review is the query in the
form that actually ran, with the rows it returned. A conformance test runs one scenario on every store and compares the
evidence. And ADR-0032: no model-written query reaches a store; every query is built from a template and strictly
validated values.

### MASTER'S READINGS (claims, measured by the master 2026-10-24 at 5faabc9 — re-measure, you may refute them)
- W48 landed the interface: `internal/store/store.go:26` (`store.Store`), `zetaris.go`, `local.go`, `settings.go`
  (`PROVEN_EVIDENCE_STORE`, `couchdb` refused by name as not built), `conform_test.go` (the conformance test),
  `missing_test.go` (a missing table is a named `service.Down`).
- `internal/querytpl/forms.go`: `Forms{SQL, Mango}`; `BuildForms` leaves `Mango` empty — this wagon fills it.
- Four templates: OSVByPackage, OSVByAdvisory, KEVByCVE, EPSSByCVE (`querytpl.Catalogue()`).
- A parallel wagon (W47) moves these templates behind a source-adapter interface; build the Mango form where the SQL
  form is built today, and the master absorbs the two at landing.
If any of this is false at your base, say so in `- master-was-wrong:` and build on what is there.

### WHAT TO BUILD
1. A Mango selector for every template, built from the same validated values as the SQL form (never from text a model
   wrote); `BuildForms` fills `Mango`.
2. A CouchDB store behind `store.Store`: one database per registered table, settings per ADR-0013 (URL, database
   prefix, credential from the environment, redacted), a missing database is a named down store, rows returned in the
   same column order as the SQL form.
3. `PROVEN_EVIDENCE_STORE=couchdb` accepted; the evidence names the Mango form that ran.
4. The conformance test covers CouchDB through a recorded HTTP server (`httptest`), so it runs with no CouchDB present;
   a live test against a real CouchDB runs only when an env var names its URL, and skips by name otherwise.
5. docs: CONFIGURATION (generated) and the store section of docs/ADAPTERS.md.

### ACCEPTANCE
1. Reproduce first: before changing anything, run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).
3. **Negative controls (law #27):** (a) make the CouchDB store drop one row → the conformance test red; (b) make a
   missing database return an empty result instead of a down store → its test red; restore each. Name the mutated lines.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it
(e.g. `ADR-0047: …`); (5) every test says which property it proves; (6) English, plain; no AI attribution, no host
names, internal paths or credential names, no commented-out code. Commit subjects `[couchdb-evidence-store]
<area>: <what>`, the area ONE token, at most 72 characters (the merge refuses anything else).

### WHAT IS NOT IN SCOPE
Source adapters (W47); loading data into CouchDB beyond what a test needs.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████` in the worktree, except the files this
order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you
end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
