# ORDER — the-evidence-store-interface · round 1

- change: the-evidence-store-interface
- round: 1
- base: main @ c1805b0
- executor: cursor on runner-cursor
- written: 2026-10-24T07:14:17Z
- task: one evidence-store interface, Zetaris and a local one-file store behind it, one template form per query language, a conformance test (ADR-0047)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-evidence-store-interface
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W48
- review: named
- repo: proven
- base: main @ c1805b0
- branch: fix/the-evidence-store-interface
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 6.00
- cap-usd: 12.00
- wall: 180m
- task: one evidence-store interface, Zetaris and a local one-file store behind it, one template form per query language, a conformance test (ADR-0047)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-evidence-store-interface] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.


### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0047, docs/adr/ADR-0047-*.md) — read it whole:
1. One evidence-store interface: register a source's table, run a query built from a template, return its columns and
   rows with the snapshot date.
2. Three stores: Zetaris (over its REST API, as today); a local store in one file with no server (DuckDB or SQLite,
   chosen when built), grown from the engine sample mode already uses; and CouchDB (a later wagon, W49).
3. One setting chooses the store (`PROVEN_EVIDENCE_STORE=zetaris|local|couchdb`). The default stays Zetaris.
4. A template has one form per query language: SQL for Zetaris and the local store, a Mango selector for CouchDB. The
   evidence in a review is the query in the form that actually ran, with the rows it returned. A conformance test runs
   one scenario on every store and compares the evidence they produce.
5. Source adapters (ADR-0046) load into the configured store, whichever it is.
6. The local store works fully air-gapped (ADR-0034).
And ADR-0032: no model-written SQL reaches a store; Proven builds every query from a template and strictly validated values.

### MASTER'S READINGS (claims, measured by the master 2026-10-24 — re-measure, you may refute them)
- Today the live engine is `internal/query` (Zetaris REST client, `query.Load`, `NewLive`) and the sample engine is
  `query.OpenSample`, which loads each fixture's `tables.sql`; templates live in `internal/querytpl`.
- The local store should grow from the sample engine (same SQL dialect), so the conformance test can run the sample
  scenarios on it with no network.
If any of this is false at your base, say so in `- master-was-wrong:` and build on what is there.

### WHAT TO BUILD (this wagon: the interface, Zetaris and local; CouchDB is W49)
1. The interface (register table, run template query, columns + rows + snapshot date) and `PROVEN_EVIDENCE_STORE`
   (`zetaris` default, `local`; `couchdb` refused by name as not built yet) per ADR-0013.
2. Zetaris behind it, unchanged in behaviour; the local one-file store behind it, air-gapped; sample mode uses it.
3. Templates carry a `sql` form now, with room for a `mango` form (W49) — the evidence names the form that ran.
4. A conformance test: one sample scenario on every built store (Zetaris through a recorded HTTP server, local for
   real) producing the same evidence.
5. docs: the store setting in CONFIGURATION (generated) and a short section in docs/ADAPTERS.md.

### ACCEPTANCE
1. Reproduce first: before changing anything, run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green; every existing scenario unchanged.
3. **Negative control (law #27):** make the local store drop one row → the conformance test red; restore. Name the
   mutated field.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it
(e.g. `ADR-0047: …`); (5) every test says which property it proves; (6) English, plain; no AI attribution, no host
names, internal paths or credential names, no commented-out code. Commit subjects `[the-evidence-store-interface]
<area>: <what>`, the area ONE token, at most 72 characters (the merge refuses anything else).

### WHAT IS NOT IN SCOPE
CouchDB (W49); new sources (W43/W47).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `██████████████████████████████████████████` in the worktree, except the
files this order names.

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
