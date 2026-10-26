# ORDER — the-engines-keep-memory-and-disk-equal · round 1

- change: the-engines-keep-memory-and-disk-equal
- round: 1
- base: main @ 6b442ac
- executor: cursor on runner-cursor
- written: 2026-10-23T07:46:10Z
- task: H-MEM and the World Model keep memory and disk equal on a failed write, two agents guard a nil adapter, and the test runner has one refusal type

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-engines-keep-memory-and-disk-equal
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W24
- review: master-verify-only
- repo: proven
- base: main @ 6b442ac
- branch: fix/the-engines-keep-memory-and-disk-equal
- lands-on: main
- worktree: ██████████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: H-MEM and the World Model keep memory and disk equal on a failed write, two agents guard a nil adapter, and the test runner has one refusal type

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-engines-keep-memory-and-disk-equal] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE DEFECTS (the master's helpers' reading, 2026-10-23 — CLAIMS for you to test, law #44)
- **F28:** `internal/hmem/hmem.go` `Add` (line ~259) and `RememberRefutation`: when the audit or the persist step fails,
  the record stays in the in-memory map while the call returns an error, so memory and disk disagree. `Clear` (~407)
  writes no audit line and does not reset `seq`.
- **F29:** `MarkSuperseded` (~416) sets only `SupersededBy`, never `Supersedes` on the replacing record, and audits with
  the actor "retrieval".
- **F2/F3:** `internal/agent/investigator` `Investigate` (~47) calls `q.Query` with no nil check; `internal/agent/prover`
  `proveNamed` (~85) and `proveWritten` (~104) call `run.Run` with no nil check; each package's `Execute` checks it.
- **F4:** `proveWritten`: a written test that was not found but reports `PassesOnBase` gets the note "written test passed
  on the change".
- **F30:** `internal/testrun` has two refusal types with the same `Error()` body: `SettingError` (container.go ~71) and
  `ConfigError` (settings.go ~27). `ProverOffReason` (live.go ~12) still reads "no container runner yet".
**Step 1, before any change:** write a failing test for F28, F29, F2 and F4 and paste the red output. If one is green at
the base, write it in `- master-was-wrong:` and leave that item out.

### WHAT TO BUILD
1. F28: a failed audit or persist leaves memory as it was (the record is not added); `Clear` audits and resets `seq`.
2. F29: both sides of a supersession are set; the actor names who superseded.
3. F2/F3: a nil adapter is a named refusal, not a panic.
4. F4: the note says what happened (not found).
5. F30: one refusal type (keep the exported name callers use; `go vet` and every caller compile); the reason text is true.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. The failing tests from step 1 pass.
3. **Negative control (law #27):** undo the F28 rollback → its test turns red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
internal/pipeline (another wagon is changing it); any network; the Meterless conformance suites' contract.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████████████████████` in the worktree, except the
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
