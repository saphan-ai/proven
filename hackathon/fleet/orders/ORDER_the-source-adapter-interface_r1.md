# ORDER — the-source-adapter-interface · round 1

- change: the-source-adapter-interface
- round: 1
- base: main @ 5faabc9
- executor: cursor on runner-cursor
- written: 2026-10-24T14:10:08Z
- task: a public data-source adapter interface; OSV, KEV and EPSS move behind it; one worked example per adapter kind in docs/ADAPTERS.md (ADR-0046)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-source-adapter-interface
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W47
- review: named
- repo: proven
- base: main @ 5faabc9
- branch: fix/the-source-adapter-interface
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 6.00
- cap-usd: 12.00
- wall: 180m
- task: a public data-source adapter interface; OSV, KEV and EPSS move behind it; one worked example per adapter kind in docs/ADAPTERS.md (ADR-0046)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-source-adapter-interface] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-24 (translated from Polish): *"what if we add an adapter for adding more data sources, feeding the database and
plugging into the decision engine?"* 2026-10-21: *"Working first, pretty later."* 2026-10-23: the judges read this code
(ADR-0027).

### THE OWNER'S DECISION (ADR-0046, docs/adr/ADR-0046-a-data-source-is-one-adapter.md) — read it whole:
1. A third public adapter interface, for data sources, beside languages and tools. A source adapter declares: how it
   loads (origin, refresh, the table it is registered as in the configured evidence store, the snapshot date); its
   query templates, which the Investigator may pick (a model never writes the query); its rule — what a returned row
   means for a change, which becomes a finding whose evidence is the query and its rows, decided by the Judge; a sample
   table and its tests.
2. The sources of ADR-0039 are written as adapters (that is W43's round 2, NOT this wagon).
3. An organisation can add its own sources the same way.
4. docs/ADAPTERS.md covers all three kinds, with one worked example each.
And ADR-0047 (landed in W48): source adapters name no store; the store (`PROVEN_EVIDENCE_STORE`) runs whatever form the
template carries. ADR-0032: no model-written query reaches a store.

### MASTER'S READINGS (claims, measured by the master 2026-10-24 at 5faabc9 — re-measure, you may refute them)
- Templates live in `internal/querytpl/querytpl.go`: `Catalogue()` (line ~135) lists four templates (OSVByPackage,
  OSVByAdvisory, KEVByCVE, EPSSByCVE); `Build`/`Validate`/`lookup` dispatch on the name; `OSVPackageSQL` (~354),
  `KEVSQL` (~384), `EPSSSQL` (~396) build the fixed queries. `forms.go` holds `Forms{SQL, Mango}` and `BuildForms`.
- The three callers that build a source query by hand: `internal/osv/check.go:207` (`querytpl.OSVPackageSQL`),
  `internal/osv/check.go:297` (`querytpl.KEVSQL`), `internal/epss/epss.go:152` (`querytpl.EPSSSQL`).
- The other adapter interfaces to sit beside: `internal/tool/tool.go:23` (`tool.Adapter`), `internal/manifest/adapter.go`
  (language), `internal/store/store.go:26` (`store.Store`).
If any of this is false at your base, say so in `- master-was-wrong:` and build on what is there.

### WHAT TO BUILD
1. A `source` package with the adapter interface: ID, load description (origin, refresh, table name, snapshot date),
   its templates (each a `querytpl.Template` plus its builder), and its rule (rows → finding evidence or nothing).
   A registry lists the built-in adapters in a fixed order.
2. OSV, KEV and EPSS become the first three adapters. `querytpl.Catalogue()` and `Build` are written FROM the registry,
   so a template exists in exactly one place. The three call sites above go through the adapter.
3. Behaviour unchanged: every existing scenario, fixture prompt and generated doc (docs/QUERIES.md) byte-identical unless
   a change is required and named in the status.
4. A test that adds one source ONLY through the interface (a test-only adapter with a sample table) and shows its
   template in the catalogue, its query run on the local store, and its rule produce evidence.
5. docs/ADAPTERS.md: one worked example each for a language, a tool and a source.

### ACCEPTANCE
1. Reproduce first: before changing anything, run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).
3. **Negative control (law #27):** remove the test-only adapter from the registry → its catalogue/evidence test red;
   restore. Name the mutated line.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it
(e.g. `ADR-0046: …`); (5) every test says which property it proves; (6) English, plain; no AI attribution, no host
names, internal paths or credential names, no commented-out code. Commit subjects `[the-source-adapter-interface]
<area>: <what>`, the area ONE token, at most 72 characters (the merge refuses anything else).

### WHAT IS NOT IN SCOPE
New sources (W43 round 2); CouchDB (W49); loading data into Zetaris (the Zetaris track).

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
