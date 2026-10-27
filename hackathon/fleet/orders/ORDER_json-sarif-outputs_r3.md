# ORDER — json-sarif-outputs · round 3

- change: json-sarif-outputs
- round: 3
- base: main @ 9303b2d
- executor: claude-code on macm5
- written: 2026-10-23T21:00:34Z
- task: W2 round 5 (stream round 3) — output identity follows a dangling final symlink, and the write itself refuses to overwrite a file another output already wrote; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — json-sarif-outputs
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- round: 3
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/json-sarif-outputs
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W2 round 5 (stream round 3) — output identity follows a dangling final symlink, and the write itself refuses to overwrite a file another output already wrote; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[json-sarif-outputs] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit or a file.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHERE YOU START (master, 2026-10-24; claims to test, law #44)
Your branch tip \`04c5815\` (rounds 1-2 of this stream). The master measured it merged with main 0b954a9: vet clean, 38 packages ok,
0 FAIL; a symlinked DIRECTORY collision and a relative-vs-absolute collision both exit 7.

### WHY THIS ROUND (codex review of 04c5815, verdict **reject**; carried verbatim, law #44)
> BLOCKING — `internal/cli/config.go:496-505`, `internal/cli/config.go:470-473`, `internal/cli/run.go:233-238`: a symlink to a not-yet-created output file is not resolved. `EvalSymlinks(alias.json)` fails because its target does not exist; the loop then resolves only the parent, leaving `alias.json` as the identity. `os.Stat(alias.json)` also fails, so the identity and same-file checks miss `actual.json`. A normal `os.WriteFile(alias.json, ...)` follows that link and creates `actual.json`; the subsequent SARIF write overwrites the JSON. This is the same collision property as the earlier BLOCKING finding, not an unrelated new review topic. Remedy: resolve a dangling final symlink's target when deriving output identity (including relative link targets), check identity before review, and add a test that asserts exit 7, both flags in the refusal, no review, and no output. Test each pair of output flags with a dangling alias. Negative control: remove that resolution and observe the test return exit 0 with the later format replacing the first.

### WHAT TO DO
1. **Reproduce first:** in \`t.TempDir()\`, \`ln -s actual.json alias.json\` (target absent), then \`-o actual.json --sarif-output
   alias.json\` (and the reverse); paste exit 0 and the overwritten file today.
2. **Identity follows links all the way:** when the final component is a symlink, resolve it with \`os.Readlink\` (relative
   targets against the link's directory), repeatedly, with a bound on the number of hops, whether or not the target exists;
   then resolve the existing prefix as today. A link loop or too many hops is refused at load (exit 7, naming the flag).
   Test every pair of output flags with a dangling alias; each must exit 7, name both flags, run no review, write no file;
   red with the \`Readlink\` resolution removed — paste it.
3. **A second line of defence at write time:** before writing each output, refuse (exit 2, naming both outputs) when the
   path is the same file as one already written in this run (\`os.SameFile\` against the \`FileInfo\` of each written file). A
   test drives an alias the load check cannot see (e.g. created between load and write through a test hook) and asserts
   exit 2 and that the first output is intact.
4. Nothing else; \`internal/review\` and \`AGENTS.md\` unchanged; new commits on top of 04c5815.

### CODE IS WRITTEN TO BE READ (ADR-0027)
A WHY comment at the link resolution and at the write-time guard (two outputs must never share one file, and why both
checks exist); each test names the property it proves. English, plain.

### ACCEPTANCE
1. \`git log --oneline 04c5815..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat 04c5815..HEAD -- AGENTS.md internal/review\` is EMPTY.
3. \`go vet ./...\` clean; \`go test -count=1 ./...\` green, counts per package; \`gofmt -l\` empty.
4. The reds of steps 2 and 3; paste them.
5. A \`- master-was-wrong:\` line: \`—\`, or what you measured that contradicts this order.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████` in the worktree. Nothing outside it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their exit codes, the reds. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
