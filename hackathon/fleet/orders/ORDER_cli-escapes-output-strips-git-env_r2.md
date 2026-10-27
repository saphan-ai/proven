# ORDER — cli-escapes-output-strips-git-env · round 2

- change: cli-escapes-output-strips-git-env
- round: 2
- base: main @ 0b954a9
- executor: claude-code on macm5
- written: 2026-10-24T02:30:46Z
- task: W1c round 2 — a continuation line of a multi-line field can never read as Proven's own line; tab and malformed UTF-8 escaped too; every printed field covered by the test; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — cli-escapes-output-strips-git-env
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W1c
- round: 2
- review: named
- repo: proven
- base: main @ 0b954a9
- branch: fix/cli-escapes-output-strips-git-env
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W1c round 2 — a continuation line of a multi-line field can never read as Proven's own line; tab and malformed UTF-8 escaped too; every printed field covered by the test; nothing else

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
Your round 1 tip \`1e77598\` (025fb8c git child env, 1e77598 text escaping). Both reviews accept the git-environment half (P2-F3)
and the field coverage of \`visible\`; both reject the text half on the points below.

### WHY THIS ROUND (both reviews of 1e77598 returned **reject**; carried verbatim, law #44)
claude-code:
> C1 (blocking) — internal/cli/format.go:203 / visibleLines: a continuation line of a multi-line field must not be able to read
> as a line of Proven's own output. Prefix every kept newline with a constant marker no Proven line starts with, and add a test
> whose fixture Output holds "\nfailed services: none" (and a finding … ERROR line) and which asserts that no output line other
> than Proven's own starts at column 0 with those keywords; unplug: remove the prefix → red.
> C2 (owed) — terminal_escape_test.go:20-37: give every printed field a value carrying raw (Class, ClassRule, the three
> statement kinds, Service, TracePath, an EPSS phrase), so that removing visible from any one call site turns the test red.
codex:
> BLOCKING — format.go:203-205, 239-264: visibleLines leaves an attacker-controlled test-output newline intact and does not
> indent subsequent lines. A test-output value containing first\nfailed services: none\nlast prints an unprefixed failed
> services: line.
> BLOCKING — format.go:229-264: isControl explicitly preserves tab (C0 U+0009), contrary to the requested C0 range, and Go's
> rune iteration replaces malformed UTF-8 bytes with U+FFFD rather than rendering their original \xNN values.

### WHAT TO DO
1. **Reproduce first:** paste today's output for a test output \`first\nfailed services: none\nlast\`, and for a title holding a
   tab and the byte 0xff.
2. **No forged lines:** every continuation line of a multi-line field is printed behind a constant prefix that no line Proven
   writes itself starts with (name it and say why it cannot collide). Test: no line outside Proven's own starts at column 0
   with \`failed services:\`, \`finding \`, \`review class:\`, \`trace:\`; red with the prefix removed — paste it.
3. **The whole byte range:** tab is escaped like the rest of C0; an invalid UTF-8 byte is shown as \`\xNN\` of the original
   byte (iterate bytes where the input is not valid UTF-8). Test: tab and 0xff in a title; red with each change reverted.
4. **Coverage (claude C2):** the escape test fills every printed field with a raw control byte, so removing \`visible\` from any
   one call site turns it red; name the call sites and paste one red.
5. Nothing else; the git-environment half stays as it is; \`internal/review\` and \`AGENTS.md\` unchanged; new commits on top of
   1e77598.

### CODE IS WRITTEN TO BE READ (ADR-0027)
The byte contract written once, in the escaping function's doc comment (which bytes, how shown, why continuation lines carry a
prefix); each test names the property it proves. English, plain; private until merged — name properties, not weaknesses.

### ACCEPTANCE
1. \`git log --oneline 1e77598..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat 1e77598..HEAD -- AGENTS.md internal/review\` is EMPTY.
3. \`go vet ./...\` clean; \`go test -count=1 ./...\` green, counts per package; \`gofmt -l\` empty.
4. The reds of steps 2, 3 and 4; paste them.
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
