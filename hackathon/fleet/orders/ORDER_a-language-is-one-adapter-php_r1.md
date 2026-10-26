# ORDER — a-language-is-one-adapter-php · round 1

- change: a-language-is-one-adapter-php
- round: 1
- base: main @ 5a9f2fe
- executor: cursor on runner-cursor
- written: 2026-10-24T02:36:54Z
- task: one public adapter interface per ecosystem; the Go, npm and PyPI readers move behind it; PHP (Composer) is the first new language (ADR-0038)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-language-is-one-adapter-php
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W38
- review: master-verify-only
- repo: proven
- base: main @ 5a9f2fe
- branch: fix/a-language-is-one-adapter-php
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 150m
- task: one public adapter interface per ecosystem; the Go, npm and PyPI readers move behind it; PHP (Composer) is the first new language (ADR-0038)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-language-is-one-adapter-php] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0038, docs/adr/ADR-0038-a-language-is-one-small-adapter.md) — read it whole:
one public interface per ecosystem, small enough for one pull request; every adapter ships with a sample scenario and
its tests; new languages in order PHP, Rust, then the others; docs/ADAPTERS.md with PHP as the worked example.

### THE MASTER'S READING (2026-10-24 — CLAIMS to re-measure, law #44)
`internal/manifest/gomod.go` ~65: `Reader{Ecosystem, Match, Diff, Notes, Lock, NoteErr}`; constructors per reader
(Go, npm, PyPI, go.sum, package-lock.json, poetry.lock). The OSV check compares versions per ecosystem
(`internal/semver`, `internal/pep440`).

### WHAT TO BUILD
1. A public, documented adapter interface (an exported type in its own package or in `internal/manifest`, your call,
   say why) covering: which files, how a base and a change differ, the OSV ecosystem name, and how versions compare.
2. The existing readers move behind it with no change in behaviour (the existing tests and scenarios unchanged).
3. PHP: `composer.json` and `composer.lock`, ecosystem `Packagist`, version comparison per Composer's rules (or a
   documented subset), with a sample scenario in testdata/ and a section in docs/SAMPLE_MODE.md.
4. docs/ADAPTERS.md: how to add a language, with PHP as the worked example.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Every existing manifest and scenario test passes unchanged; the PHP scenario reports its dependency.
3. **Negative control (law #27):** drop the PHP adapter from the registry → the PHP scenario test turns red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Rust and the others (W39 onward); any network.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████` in the worktree, except the
files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
