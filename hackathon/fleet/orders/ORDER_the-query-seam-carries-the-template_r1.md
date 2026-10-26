# ORDER — the-query-seam-carries-the-template · round 1

- change: the-query-seam-carries-the-template
- round: 1
- base: main @ fa21bc8
- executor: cursor on runner-cursor
- written: 2026-10-25T01:25:16Z
- task: the review's query seam carries the template request, not SQL text, so every store — CouchDB included — runs the form of its own language in a live review (ADR-0047, ADR-0032)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-query-seam-carries-the-template
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W52
- review: named
- repo: proven
- base: main @ fa21bc8
- branch: fix/the-query-seam-carries-the-template
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 5.00
- cap-usd: 10.00
- wall: 180m
- task: the review's query seam carries the template request, not SQL text, so every store — CouchDB included — runs the form of its own language in a live review (ADR-0047, ADR-0032)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-query-seam-carries-the-template] <area>: <what>` included, the area ONE token. No AI attribution anywhere. The tool
re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0047 points 3–4; ADR-0032)
One setting chooses the store (zetaris | local | couchdb). A template has one form per query language; the evidence in a
review is the query in the form that actually ran, with its rows. No model-written query reaches a store.

### MASTER'S READINGS (claims, measured by the master 2026-10-25 on main — re-measure, you may refute them)
- The live review path is SQL-only: `query.Adapter` (internal/query/query.go:30) is `Query(ctx, sql string)`. Callers that
  build SQL and then call it: internal/agent/investigator/investigator.go:68-78 (`Execute`) and :115;
  internal/osv/check.go:227 and :319; internal/epss/epss.go:163.
- `store.Store` (internal/store/store.go:26, `Register` / `Run(ctx, querytpl.Request, querytpl.Relations)`) has Zetaris,
  local and CouchDB implementations, but review.Open (internal/review/open.go:138-160) installs SQL adapters instead:
  `lazyLocal` (:246-261) for local and `couchdbSQL` (:155-159, :264) which names CouchDB down on every query.
- A template a source adapter registered has the sql form only; CouchDB names itself down for it ("no mango form for
  <template>", internal/store/couchdb.go).
If any of this is false at your base, say so in `- master-was-wrong:` and build on what is there.

### WHAT TO BUILD
1. The pipeline's evidence seam takes `querytpl.Request` + `querytpl.Relations` and returns the evidence of the form that
   ran (form name, query text, columns, rows, snapshot). The callers above stop building SQL themselves.
2. review.Open installs the configured `store.Store` on that seam for every store; `couchdbSQL` and the SQL-only
   `lazyLocal` path go away (the local file is still opened lazily, after the inside-the-target check).
3. Sample mode unchanged: every sample scenario, fixture prompt and generated doc byte-identical unless a change is
   required and named in the status.
4. A review-level test: with `PROVEN_EVIDENCE_STORE=couchdb` against a recorded CouchDB, a review's evidence names the
   `mango` form and its rows; the same scenario on the local store names `sql`.

### ACCEPTANCE
1. Reproduce first: at your base, run the review-level couchdb test you will write and paste its failure (couchdb down).
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).
3. **Negative control (law #27):** make the seam send the sql form to CouchDB → the review-level test red; restore.
   Name the mutated line.

### WHAT IS NOT IN SCOPE
New sources (W43); Mango forms for registered templates.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it;
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `█████████████████████████████████████████████████` in the worktree, except the files this order names.

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
