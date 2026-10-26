# ORDER — the-prover-runs-only-its-test · round 1

- change: the-prover-runs-only-its-test
- round: 1
- base: main @ 21af174
- executor: claude-code on macm5
- written: 2026-10-23T11:20:48Z
- task: the command the Prover's container runs is built by Proven from the written test's name and language, never taken from model text

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-prover-runs-only-its-test
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W29
- review: named
- repo: proven
- base: main @ 21af174
- branch: fix/the-prover-runs-only-its-test
- lands-on: main
- backend: claude-code
- model: claude-sonnet-5-5
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 150m
- task: the command the Prover's container runs is built by Proven from the written test's name and language, never taken from model text

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-prover-runs-only-its-test] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### PRIVATE UNTIL MERGED (read first)
These are security findings of the proven-security-review train, PRIVATE until each fix is merged. Proven is a public
repository: **no finding text in a commit message, a code comment, a test name or a doc**. Commit subjects describe the
behaviour plainly (for example `testrun: run the container as an unprivileged user`), never as a defect, a weakness or
an attack. Comments say what the code guarantees and why (ADR-0025, ADR-0005, ADR-0013), not what was wrong.
The reviewers' notes with their failing tests are provisioned at `████████████████████████` — read the sections named
below; copy a test only after renaming it to the property it proves.

### THE PROPERTY THAT MUST HOLD (SECURITY_REVIEW.md, piece 1)
Nothing in the change, and nothing a model answers, can shape the container's arguments or the command it runs.

### WHAT THE REVIEWS MEASURED (at main 90d6951 — CLAIMS for you to re-measure, law #44)
- `internal/agent/prover/prover.go` line ~112 and `internal/pipeline/pipeline.go` lines ~651-652: the command string
  from the Prover's model answer becomes the `sh -c` command run in the container.
  Notes: `█████████████████████████████████████████████████` (`TestReviewModelTextDoesNotBecomeTheContainerCommand`),
  `████████████████████████████████████████████████` (`TestReviewRunnerRefusesUnboundedCommand`).
**Step 1, before any change:** add the tests (renamed to the property) and paste their red output at the base.

### WHAT TO BUILD
1. The model names a test (its file and name) and nothing else that is executed. Proven builds the argv itself from a
   fixed per-language template (Go: `go test -run '^<Name>$' ./<pkg>` with the name checked against a strict
   identifier pattern and the package path checked to stay inside the checkout); no shell (`sh -c`) anywhere.
2. A model answer that carries a command, or a name or path that fails the checks, is refused with a named reason in
   the trace; the finding then has no test evidence (the Judge decides as today).
3. The sample fixtures and docs/SAMPLE_MODE.md still produce the same review; regenerate anything they record if the
   answer shape changes, and say what changed.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. A test: a model answer whose command field carries extra shell text → refused, the container is not started.
3. A test: a valid test name → the argv is exactly the template, no shell.
4. **Negative control (law #27):** let the answer's command through again → test 2 turns red; restore.
5. `proven scan --sample` on `testdata/fixture/discount` from a scratch directory: same findings as main.
### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
The container flags (W28); traces and error messages (W30); the CLI.

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
| private workspace path | 7 |
