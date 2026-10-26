# ORDER — json-sarif-outputs · round 2

- change: json-sarif-outputs
- round: 2
- base: main @ 9303b2d
- executor: claude-code on macm5
- written: 2026-10-23T19:37:56Z
- task: W2 round 4 (stream round 2) — refuse outputs that reach one file through a symlinked directory or a relative and an absolute spelling; keep internal/cli's tests runnable without the schema module; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — json-sarif-outputs
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- round: 2
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/json-sarif-outputs
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W2 round 4 (stream round 2) — refuse outputs that reach one file through a symlinked directory or a relative and an absolute spelling; keep internal/cli's tests runnable without the schema module; nothing else

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
Your branch tip \`f6b6b1f\` is your round 1 (250301b, 3aabc4b) plus the master's absorb of main ce61e44. The master measured on
it: both SARIF schema tests PASS, vet clean, 36 packages ok, 0 FAIL; the binary refuses \`-o X --sarif-output X\` with exit 7.

### WHY THIS ROUND (both reviews of f6b6b1f returned **reject** on the same finding; carried verbatim, law #44)
codex:
> RED — BLOCKING: Output aliases through a symlinked directory overwrite one another, rather than being refused at load.
> \`internal/cli/config.go:434-463\` compares \`filepath.Abs\` paths without resolving symlinks; \`internal/cli/run.go:198-244\`
> then writes both bodies in order.
claude-code:
> C1 · BLOCKING · internal/cli/config.go:441-463. Two outputs that reach one file through a symlinked directory are not
> refused. Measured: exit 0, and the SARIF body replaced the JSON body in \`real/r.json\`. Remedy: resolve each path's existing
> parent directory with \`filepath.EvalSymlinks\` before comparing.
> C2 · OWED · Add a relative-vs-absolute row (\`t.Chdir(dir)\`, \`-o r.json\`, \`--sarif-output <dir>/r.json\` must exit 7).
> Unplug: m9 (\`filepath.Abs\` replaced by \`filepath.Clean\`) must turn it red; today m9 survives.
> C3 · OWED · The test-only module makes the whole \`internal/cli\` package \`[setup failed]\` on a seat without it cached.
> Remedy: move the schema test behind a build tag or into its own package.

### WHAT TO DO
1. **Reproduce first:** a test with a real symlinked directory (\`os.Symlink\` in \`t.TempDir()\`): \`-o real/r.json\` and
   \`--sarif-output link/r.json\` where \`link\` → \`real\`; paste what it returns today (exit 0).
2. **C1:** two outputs are the same file when their paths are equal after resolving symlinks in the longest existing prefix
   of each path (\`filepath.EvalSymlinks\` on the existing parent, then join the rest; and \`os.SameFile\` when both files
   exist). Refused at load with exit 7, naming both flags and the path. Test: the symlinked-directory row; red with the
   resolution removed (paste it).
3. **C2:** the relative-vs-absolute row; red with \`filepath.Abs\` replaced by \`filepath.Clean\` (paste it).
4. **C3:** move the SARIF schema tests into their own package (e.g. \`internal/cli/sarifschema\`, testing the writer through
   the exported seam), so \`go test ./internal/cli\` builds without the validator module; the schema tests still run in
   \`go test ./...\` and still fail with \`version\` removed.
5. Nothing else; \`internal/review\` and \`AGENTS.md\` unchanged; new commits on top of f6b6b1f.

### CODE IS WRITTEN TO BE READ (ADR-0027)
A WHY comment at the path-identity check (what counts as one file, and why symlinks are resolved); each test names the
property it proves. English, plain.

### ACCEPTANCE
1. \`git log --oneline f6b6b1f..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat f6b6b1f..HEAD -- AGENTS.md internal/review\` is EMPTY.
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
