# ORDER — cli-escapes-output-strips-git-env · round 3

- change: cli-escapes-output-strips-git-env
- round: 3
- base: main @ 0b954a9
- executor: claude-code on macm5
- written: 2026-10-24T03:57:21Z
- task: W1c round 3 — what the command line writes to stderr carries no raw control byte from the change; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-escapes-output-strips-git-env
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1c
- round: 3
- review: named
- repo: proven
- base: main @ 0b954a9
- branch: fix/cli-escapes-output-strips-git-env
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W1c round 3 — what the command line writes to stderr carries no raw control byte from the change; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[cli-escapes-output-strips-git-env] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit or a file.
These two findings are PRIVATE until merged: name them by their property in commits and comments, never as a weakness.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHERE YOU START (master, 2026-10-24; claims to test, law #44)
Your round 2 tip \`40a9570\`. The claude-code review of it returned accept-with-conditions: text output, continuation lines and
the git child environment hold (12 escaping mutants and the env unplug each turn a named test red). One condition remains.

### WHY THIS ROUND (claude-code review of 40a9570; carried verbatim, law #44)
> - **C1 (condition; the class of property P4-F5): stderr prints errors that carry the change's bytes raw.**
>   - Location: `internal/cli/run.go:92` (and `:38`, `:50`, `:56`, `:65`, `:68`, `:89`, `:102`, the same pattern) prints `err.Error()` unescaped.
>   - Trigger: the errors carry file names from the change through `%s`. `internal/change/read.go:78` (`change: symlink %s`), `internal/change/archive.go:145` and `:224` (`change: unsupported file %s`), and `OverBound("the file "+clean, …)` (`archive.go:121`, `:228` → `limits.go:168`).
>   - Evidence (`review/stderr_probe_test.go`, run with `go test -tags reviewprobe ./review/`):
>     - (a) `cli.Run(["scan","--sample",<fixture>])` on a fixture whose `change/` holds a symlink named `x\x1b[2Jy` exits 2 with stderr `"change: symlink …/change/x\x1b[2Jy\n"`, where the raw ESC is in the buffer.
>     - (b) On the git path, `change.ExportCommit` of a commit holding `big\x1b[2J.txt` over a 4-byte file bound returns `change: the file big\x1b[2J.txt is larger than the 4-byte bound …` with the raw ESC. That is the error the pipeline returns (`pipeline/trees.go:106`) and `run.go:92` prints.
>   - Remedy: pass the stderr sentences through `visible` at the print sites in `run.go`, which keeps one encoder, or quote paths with `%q` at the error sites. Add a test that drives `Run` into one of these errors and asserts no raw control byte on stderr.
>   - The master rules whether this lands in this wagon or as a named follow-on. Either way, P4-F5 ("output to a terminal") is not closed until it lands.

### WHAT TO DO
1. **Reproduce first:** a test that drives \`cli.Run\` into the symlink error with a change file named \`x\x1b[2Jy\` (the
   reviewer's case a); paste the raw byte on stderr today.
2. **Every stderr print site in \`internal/cli/run.go\`** passes its sentence through the same \`visible\` encoder the text output
   uses (one encoder, one byte contract). Test: case (a) and the over-bound file name of case (b) show no raw control byte on
   stderr; red with \`visible\` removed from the print site — paste it.
3. Nothing else; the text writer and the git environment stay as they are; \`internal/review\`, \`internal/change\` and
   \`AGENTS.md\` unchanged; new commits on top of 40a9570.

### CODE IS WRITTEN TO BE READ (ADR-0027)
One line at the stderr print helper: why errors pass through the same encoder as the output. Tests name the property.

### ACCEPTANCE
1. \`git log --oneline 40a9570..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat 40a9570..HEAD -- AGENTS.md internal/review internal/change\` is EMPTY.
3. \`go vet ./...\` clean; \`go test -count=1 ./...\` green, counts per package; \`gofmt -l\` empty.
4. The red of step 2; paste it.
5. A \`- master-was-wrong:\` line: \`—\`, or what you measured that contradicts this order.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████████` in the worktree. Nothing outside it.
Never run `proven scan` inside the checkout: it writes `logs/` where it starts.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their exit codes, the reds. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
