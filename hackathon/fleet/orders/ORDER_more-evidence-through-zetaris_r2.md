# ORDER — more-evidence-through-zetaris · round 2

- change: more-evidence-through-zetaris
- round: 2
- base: main @ fa21bc8
- executor: cursor on runner-cursor
- written: 2026-10-25T01:27:01Z
- task: round 2: the four ADR-0039 sources (malicious packages, project health, CVSS, SPDX licences) as source adapters on the W47 interface

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — more-evidence-through-zetaris
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W43
- review: named
- repo: proven
- base: main @ fa21bc8
- branch: fix/more-evidence-through-zetaris
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 6.00
- cap-usd: 12.00
- wall: 180m
- task: round 2: the four ADR-0039 sources (malicious packages, project health, CVSS, SPDX licences) as source adapters on the W47 interface

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[more-evidence-through-zetaris] <area>: <what>` included, the area ONE token. No AI attribution anywhere. The tool
re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### WHY THIS ROUND
Round 1 (commit 964b683, on an old base c5164f9) built the four ADR-0039 sources as hand-written templates. Since then W47
(the source-adapter interface, ADR-0046) and W48/W49 (the evidence-store interface, ADR-0047) landed. The master has moved
your branch to current main; round 1 stays readable as `git show 964b683` — PORT it, do not merge it.

### THE OWNER'S DECISIONS (ADR-0039, ADR-0046, ADR-0047, ADR-0032)
Malicious packages (OSV MAL-), project health (Scorecard through deps.dev), CVSS beside EPSS and SPDX licences, as tables
read with templates; each a source adapter: how it loads (origin, refresh, table, snapshot date), its templates, its
rule, a sample table and tests. A missing table is a named down source. Source adapters name no store.

### MASTER'S READINGS (claims, measured by the master 2026-10-25 — re-measure, you may refute them)
- The interface is internal/source (`Adapter`: ID, Load, Templates, Rule; `Register`); the built-ins OSV, KEV, EPSS are
  registered from internal/querytpl/adapters.go; every parameter has a `Match` format checked by Proven before any builder.
- `git show 964b683 --stat` lists round 1's 41 files, including testdata/fixture/* sample scenarios and tables.sql files.

### WHAT TO BUILD
1. The four sources as built-in source adapters (beside OSV, KEV, EPSS), each with its templates, its rule, its sample
   table and a sample scenario; the catalogue, docs/QUERIES.md and the prompts follow from the registry.
2. A missing table is a named down source on every store.
3. Every existing scenario unchanged.

### ACCEPTANCE
1. Reproduce first: run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).
3. **Negative control (law #27):** one source's rule made to accept an empty row list → its scenario test red; restore.
   Name the mutated line.

### WHAT IS NOT IN SCOPE
Loading the tables into Zetaris (the Zetaris track); the query seam (W52).

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it;
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████` in the worktree, except the files this order names.

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
