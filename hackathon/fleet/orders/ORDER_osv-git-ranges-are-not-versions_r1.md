# ORDER — osv-git-ranges-are-not-versions · round 1

- change: osv-git-ranges-are-not-versions
- round: 1
- base: main @ 2d3126b
- executor: cursor on runner-cursor
- written: 2026-10-23T05:29:19Z
- task: the OSV converter keeps GIT ranges out of the version table and keeps the versions list, and semver refuses an overflowing number

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — osv-git-ranges-are-not-versions
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W23
- review: master-verify-only
- repo: proven
- base: main @ 2d3126b
- branch: fix/osv-git-ranges-are-not-versions
- lands-on: main
- worktree: ███████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 90m
- task: the OSV converter keeps GIT ranges out of the version table and keeps the versions list, and semver refuses an overflowing number

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[osv-git-ranges-are-not-versions] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE DEFECTS (the master's helper's reading, 2026-10-23 — CLAIMS for you to test, law #44)
- **F22:** `internal/osv/convert.go` (used by `cmd/osv-to-csv`) turns every range into spans with no look at the OSV
  range `type`, so a `GIT` range's commit hashes become `introduced`/`fixed` versions; those rows never match and the
  trace says only "not a semantic version". It also drops the `versions` list whenever any range exists, so an advisory
  with only `GIT` ranges plus a `versions` list loses its versions.
- **F23:** `internal/semver` `number` has no overflow guard; a very long numeric field wraps. `internal/pep440` `readInt`
  has one.
**Step 1, before any change:** add the failing tests below and paste their red output. If one is green at the base,
write it in `- master-was-wrong:`.

### WHAT TO BUILD
1. Only `SEMVER` and `ECOSYSTEM` ranges become rows; a `GIT` range is skipped. When no usable range is left, the
   `versions` list gives the rows (as today when there are no ranges).
2. `semver` refuses a numeric field that does not fit, as `pep440` does.
3. `docs/OSV_TABLE.md`: one sentence on GIT ranges.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. A test: an advisory with a GIT range and a SEMVER range gives rows for the SEMVER range only.
3. A test: an advisory with only a GIT range and a `versions` list gives one row per version.
4. A test: `99999999999999999999.0.0` is refused, not wrapped.
5. **Negative control (law #27):** stop skipping GIT ranges → test 2 turns red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Fetching OSV data; any network; the Zetaris table layout.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `█████████████████████████████████████████████` in the worktree, except the
files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
