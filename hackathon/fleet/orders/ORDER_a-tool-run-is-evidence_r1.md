# ORDER — a-tool-run-is-evidence · round 1

- change: a-tool-run-is-evidence
- round: 1
- base: main @ 5a9f2fe
- executor: cursor on runner-cursor
- written: 2026-10-24T02:36:50Z
- task: a deterministic tool run on base and change is a third kind of evidence, through a public tool adapter; the Go race detector first (ADR-0040)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-tool-run-is-evidence
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W40
- review: master-verify-only
- repo: proven
- base: main @ 5a9f2fe
- branch: fix/a-tool-run-is-evidence
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 5.00
- cap-usd: 10.00
- wall: 180m
- task: a deterministic tool run on base and change is a third kind of evidence, through a public tool adapter; the Go race detector first (ADR-0040)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-tool-run-is-evidence] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0040, docs/adr/ADR-0040-a-tool-run-is-evidence.md) — read it whole:
a third kind of evidence, a deterministic tool's run compared between base and change (tool, version, exact command,
the output that differs); tools plug in through a public adapter (files, command, reading output into findings, base
vs change); first tools the race detector, fuzzing of the changed function, gosec, coverage of changed lines; every
tool runs inside the Prover's container with no network and no secret, with a time limit from the settings; each tool
adapter ships with a sample scenario and its tests, in docs/ADAPTERS.md.

### WHAT TO BUILD (this wagon: the interface and the first tool)
1. A public tool adapter interface and the evidence kind `tool` on `review.Finding` evidence (W34, in flight, adds the proof
   level "proved by a tool"; if it is not on your base, add the evidence kind only and say so; the master wires the
   level when both land).
2. The Go race detector as the first tool: `go test -race` on base and change in the Prover's container (W28's
   locked-down container, W29's argv builder, no shell); a race on the change and none on the base is evidence.
3. A recorded sample scenario (sample mode runs no container: recorded outcomes) and its tests; docs/ADAPTERS.md.
4. Time limit from a named setting (ADR-0013).

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. The sample race scenario reports a finding proved by a tool; a tool that cannot run is named, never silent.
3. A real run on a host with Docker is UNRUN on your seat if you cannot reach one; the master runs it.
4. **Negative control (law #27):** report a race present on the base too → its test red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Fuzzing, gosec and coverage (later wagons); govulncheck (W41).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████` in the worktree, except the
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
