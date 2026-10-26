# ORDER — a-scenario-per-kind-of-evidence · round 1

- change: a-scenario-per-kind-of-evidence
- round: 1
- base: fix/epss-scores-each-cve-finding @ e354e85
- executor: cursor on runner-cursor
- written: 2026-10-23T01:18:32Z
- task: sample mode gains one small scenario per kind of evidence and refusal, so a judge sees each rule of Proven work with no network and no keys

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-scenario-per-kind-of-evidence
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W19
- review: master-verify-only
- repo: proven
- base: fix/epss-scores-each-cve-finding @ e354e85
- branch: fix/a-scenario-per-kind-of-evidence
- lands-on: main
- worktree: ███████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: sample mode gains one small scenario per kind of evidence and refusal, so a judge sees each rule of Proven work with no network and no keys

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-scenario-per-kind-of-evidence] <area>: <what>` included — use short subjects.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### STACKED (read this)
Your base is W16's branch (`fix/epss-scores-each-cve-finding` at e354e85), which has every feature on main plus EPSS
scores. W16 lands first. The existing scenario is `testdata/fixture/discount` (read docs/SAMPLE_MODE.md: its shape —
base/, change/, fixture.json with the recorded answers, tables.sql — and how to add a scenario).

### WHAT TO BUILD
Four new scenarios under `testdata/fixture/`, each small, each with a Go test that runs it in sample mode and asserts its
outcome:
1. `refuted`: the Skeptic's recorded refutation carries a test that passes on the change, so the finding is DROPPED, and
   the trace says why (W6's rule).
2. `unsafe-prompt`: the sample guardrail answers `unsafe` for one recorded prompt, so that step stops and the trace names
   the categories (W3's rule).
3. `osv-go-module`: a go.mod change raises a module to a version inside a sample OSV range, so the finding's evidence is
   the OSV query, its rows and the version comparison (W13's rule), with its EPSS score when its alias has one (W16).
4. `unknown-author`: the change's commit or description names no author model, so the review says "author model not
   stated" and which default it used (W7's rule). If sample mode has no commit history to read, say so and use the
   description field the sample request carries; if neither exists, skip this one and say why.
Then `docs/SAMPLE_MODE.md`: one short section per scenario with the exact command and its real output (run them).

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Each scenario's command from a scratch directory: paste each output.
3. **Negative control (law #27):** make the `refuted` scenario's recorded refutation carry no evidence → its test turns red
   (the finding stays); restore.

### WHAT IS NOT IN SCOPE
New features; the live demo (demo/live); `cmd/proven`'s flags.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `█████████████████████████████████████████████` in the worktree, except the files this order names.

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
