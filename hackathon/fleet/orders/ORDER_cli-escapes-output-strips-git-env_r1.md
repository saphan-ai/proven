# ORDER — cli-escapes-output-strips-git-env · round 1

- change: cli-escapes-output-strips-git-env
- round: 1
- base: main @ 0b954a9
- executor: claude-code on macm5
- written: 2026-10-23T20:13:06Z
- task: two findings of the security review on proven's command line: text written to a terminal carries no control byte taken from the change, and the git subprocess gets a minimal environment; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-escapes-output-strips-git-env
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1c
- review: named
- repo: proven
- base: main @ 0b954a9
- branch: fix/cli-escapes-output-strips-git-env
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: two findings of the security review on proven's command line: text written to a terminal carries no control byte taken from the change, and the git subprocess gets a minimal environment; nothing else

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

### THE TWO FINDINGS (proven-security-review; carried verbatim, law #44)
P4-F5, terminal side:
> control bytes other than CR, LF and NUL (for example ESC) in a package name from the change pass through to the terminal
> output (reviewer cites cmd/proven/main.go:133; not re-verified by the master). Property: output to a terminal carries no
> control byte taken from the change. Proof expected: a test that prints a finding whose package name carries ESC and asserts
> the byte is escaped or removed.
P2-F3:
> every git and container-runtime subprocess gets a minimal environment (childenv.Minimal) except one: internal/cli/run.go:328
> runs `git -C repo merge-base base head` with Proven's whole environment, keys included. Property: no key reaches a
> subprocess. Proof expected: a test that sets a key to a marker and asserts the git child's environment does not carry it
> (the shape of internal/testrun's environment tests).

### WHAT THE MASTER READ (2026-10-24, at main 0b954a9, without building; claims to test, law #44)
- The text writer is `internal/cli/format.go` (`textFormatter.Write`, `writeFinding`, `writeEvidence`); main.go:133 is the
  pre-W1b location. It prints change-derived fields (title, path, message, rule id, evidence test name, command, output, SQL,
  rows, failed-service reason) with `%s`.
- `gitMergeBase` in `internal/cli/run.go` builds `exec.Command("git", …)` with no `Env`. `internal/childenv.Minimal(extra ...string)`
  exists and `internal/change/between.go:21` already uses it for git.

### WHAT TO DO
1. **Reproduce first, in tests, before any change:** (a) a finding whose title, path and package name carry ESC (0x1b), CR
   and a C1 byte (0x9b) through `cli.Run` text output — paste the raw bytes it writes today; (b) `gitMergeBase` with
   `████████████████████=marker-…` set — show the child sees it (use a fake `git` on `PATH` that prints its environment).
2. **Terminal text:** every field the text writer prints that can come from the change is passed through one escaping
   function: control bytes (C0 except LF inside multi-line output fields, DEL, and C1) are shown as a visible escape
   (e.g. `\x1b`), never written raw. JSON and SARIF are NOT changed (their encoders already escape). Test (a) asserts no raw
   control byte remains; red with the escaping removed — paste it.
3. **Git environment:** `gitMergeBase` sets `cmd.Env = childenv.Minimal(…)` with exactly what git needs. Test (b) asserts the
   marker is absent from the child's environment; red with the line removed — paste it. `TestCIMergeBaseIsTheForkPoint` stays
   green.
4. Nothing else; `internal/review` and `AGENTS.md` unchanged.

### CODE IS WRITTEN TO BE READ (ADR-0027)
A WHY comment at the escaping function and at the env line (what a raw byte or an inherited key would allow, in one plain
sentence, named as the property); each test names the property it proves. English, plain.

### ACCEPTANCE
1. `git log --oneline 0b954a9..HEAD` shows only this round's commits; paste it.
2. `git diff --stat 0b954a9..HEAD -- AGENTS.md internal/review` is EMPTY.
3. `go vet ./...` clean; `go test -count=1 ./...` green, counts per package; `gofmt -l` empty.
4. The two reds with each fix removed; paste them.
5. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

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
| private credential or variable name | 1 |
| private workspace path | 4 |
