# ORDER — poetry-locks-and-pypi-names-are-read · round 1

- change: poetry-locks-and-pypi-names-are-read
- round: 1
- base: main @ 2d3126b
- executor: cursor on runner-cursor
- written: 2026-10-23T05:29:15Z
- task: a real poetry.lock is read, and Python package names are compared after PEP 503 normalisation

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — poetry-locks-and-pypi-names-are-read
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W22
- review: master-verify-only
- repo: proven
- base: main @ 2d3126b
- branch: fix/poetry-locks-and-pypi-names-are-read
- lands-on: main
- worktree: ████████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: a real poetry.lock is read, and Python package names are compared after PEP 503 normalisation

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[poetry-locks-and-pypi-names-are-read] <area>: <what>` included — the slug is long, so keep `<what>` short.
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
- **F21:** `internal/manifest/poetry.go` `parsePoetry` refuses any line outside a `[[package]]` table that has no `=`. Real
  lock files have multi-line arrays in `[package.dependencies]` (`numpy = [`, `{version = ">=1.22", python = "<3.12"},`,
  `]`) and older ones a `[metadata.files]` table with multi-line arrays. Such a file fails with "line N: not a package
  table", so through `NoteErr` a whole poetry.lock gives no finding, only a trace note.
- **F25:** `internal/manifest/diff.go` `diffNamed` compares requirements.txt names with exact case, so a case-only rename
  counts as "added"; PEP 503 normalisation (lower case; runs of `-`, `_`, `.` become one `-`) is done nowhere. Lock
  pairing only lower-cases.
**Step 1, before any change:** add the two failing tests below and paste their red output. If either is green at the
base, write it in `- master-was-wrong:`.

### WHAT TO BUILD
1. `parsePoetry` reads a real poetry.lock: lines inside sub-tables and multi-line arrays are skipped, a file with no
   `[[package]]` at all is still refused (it is not a lock file).
2. One PEP 503 `normalise` for PyPI names, used by requirements.txt, poetry.lock and the manifest/lock pairing; the OSV
   query uses the normalised name.
3. `docs/OSV_TABLE.md`: one sentence on name normalisation.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. A test with a realistic poetry.lock (multi-line dependency arrays and a `[metadata.files]` table) that finds a raised
   package.
3. A test: `Requests` → `requests` and `zope_interface` → `zope.interface` are the same package, not added.
4. **Negative control (law #27):** restore the old "no `=` outside a package" refusal → test 2 turns red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Pipfile.lock, yarn.lock, pnpm-lock.yaml; any network.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `██████████████████████████████████████████████████` in the worktree, except the
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
