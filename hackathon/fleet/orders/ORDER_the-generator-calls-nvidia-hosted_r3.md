# ORDER — the-generator-calls-nvidia-hosted · round 3

- change: the-generator-calls-nvidia-hosted
- round: 3
- base: fix/the-generator-calls-nvidia-hosted @ 1a45a80
- executor: claude-code
- written: 2026-10-24T03:46:51Z
- task: round 2 — close review conditions C1, C2 and owed tests O1-O5 of the named review of 1a45a80

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-generator-calls-nvidia-hosted
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W35
- review: named
- repo: proven
- base: fix/the-generator-calls-nvidia-hosted @ 1a45a80
- branch: fix/the-generator-calls-nvidia-hosted
- lands-on: main
- backend: claude-code
- model: claude-sonnet-5-5
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: round 2 — close review conditions C1, C2 and owed tests O1-O5 of the named review of 1a45a80

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-generator-calls-nvidia-hosted] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY THIS ROUND
The named review of 1a45a80 (claude-opus-5-5, 2026-10-24) returned accept-with-conditions. The conditions are pasted
below verbatim in substance; they are the work. Keep every commit on the same branch.

### C1 · blocking — a set base address follows the NVIDIA key (`internal/generator/settings.go:209-216`, `live.go:192-219`)
The reviewer measured: in nvidia mode a set `PROVEN_GENERATOR_BASE_URL` is taken over unchanged, so the NVIDIA key goes
to that address (plain http allowed) and the chunk `model` field it returns is trusted as the answering model.
`docs/LAB.md` tells Nightshift users to set exactly that variable.
**The master's ruling (the remedy you build):** nvidia mode does NOT read `PROVEN_GENERATOR_BASE_URL`. It has its own
setting, `PROVEN_NVIDIA_BASE_URL`, defaulting to NVIDIA's hosted address; a value that is not `https://` is refused by
name (ADR-0013: a wrong non-empty value refuses). The base host is named on livecheck's generator line and in the
trace's model event. `docs/LAB.md` and the generated settings page say which variable each mode reads.
Proof expected: a test with `PROVEN_GENERATOR_BASE_URL` set to another address in nvidia mode that fails before the
remedy (the request goes there) and passes after it (it does not); a test that `http://` in `PROVEN_NVIDIA_BASE_URL`
is refused by name.

### C2 · blocking — the key's redactor entry has no test (`internal/redact/redact.go:26`)
Remedy: a test in which a hosted answer echoes `████████████████████` (in its `model` field) and the trace, the
review's model list and livecheck's line carry `[redacted: ████████████████████]`. DoD names the mutated field (the
key's entry in `redact.Names`) and shows the test red with it removed.

### OWED, non-blocking — add these tests (each survives a named mutant today)
- O1 `live.go:202-205`: a `meterless` object in a hosted stream is ignored.
- O2 `stream.go:78-81`: a later chunk's `model` cannot replace the first (`nvidia/a` then `openai/b` → `nvidia/a`).
- O3 `cmd/livecheck/main.go:135`: the ready line in nvidia mode (`mode nvidia, … receipt no receipt`).
- O4 `internal/review/open.go:79`: an empty NVIDIA key names `nvidia` both in failed services and on each refused call.
- O5 `live.go:295`: in nightshift mode a 403 stays `http 403`.
- O6 ruled by the master: KEEP the new ordering (a non-JSON event after the deadline is `timeout`); add one test that
  pins it and a comment saying why.

### ACCEPTANCE
1. Reproduce first: run the C1 test before your fix and paste its failure — if it passes at 1a45a80, the reviewer
   (and the master) were wrong: say so in `- master-was-wrong:` with the output, and stop C1 there.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named as UNRUN
   with the quoted refusal (the master runs those).
3. **Negative controls (law #27):** C1 — restore the read of `PROVEN_GENERATOR_BASE_URL` in nvidia mode → its test red;
   C2 — remove the key from `redact.Names` → its test red. Restore both; name each mutated field.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it
(e.g. `ADR-0031: …`); (5) every test says which property it proves; (6) English, plain; no AI attribution, no host
names, internal paths or credential names, no commented-out code. Review findings stay private: no condition text in
commit messages, test names or comments — describe the property, never the review.

### WHAT IS NOT IN SCOPE
A live NVIDIA call; any change outside the generator, redact, livecheck, review's failed-services naming and the docs
named above.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████████` in the worktree, except the
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
| private credential or variable name | 2 |
| private workspace path | 4 |
