# ORDER — json-sarif-outputs · round 4

- change: json-sarif-outputs
- round: 4
- base: main @ 9303b2d
- executor: claude-code on macm5
- written: 2026-10-24T03:39:11Z
- task: W2 round 6 (stream round 4) — output identity is computed the way the filesystem resolves a path (links before ..), proved over a table of aliases; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — json-sarif-outputs
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- round: 4
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/json-sarif-outputs
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W2 round 6 (stream round 4) — output identity is computed the way the filesystem resolves a path (links before ..), proved over a table of aliases; nothing else

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
Your branch tip \`3d54504\` (stream rounds 1-3). The master measured it merged with main 84cf11d: 40 packages ok, 0 FAIL; a dangling
final symlink exits 7 with no output.

### WHY THIS ROUND (codex review of 3d54504, verdict **reject**; carried verbatim, law #44)
> BLOCKING — `internal/cli/config.go:505-514`, `internal/cli/run.go:243-255`, `internal/cli/outputs_test.go:33-68`. At load, `filepath.Abs` cleans `link/../report` before resolving `link`. The filesystem instead resolves the directory link before traversing `..`. With `link -> elsewhere/dir`, `-o elsewhere/report` and `--sarif-output link/../report` address one file, but their calculated identities differ. In a confined reproduction with both output paths absent, `cli.Run` returned `exit=2` and `stderr="-o and --sarif-output are the same file (.../link/../report); -o is left as it was written"`; the first file existed and retained JSON. The write-time `os.SameFile` check saves the first file from replacement, but only after the reviewer has run and the first output has been written. The required behavior is a load-time exit 7 with nothing written. Remedy: resolve symlinks in lexical path order before applying subsequent `..` segments (or use an equivalent filesystem-aware identity), and add a test that supplies the raw `link/../report` string, not `filepath.Join`'s pre-cleaned equivalent; assert exit 7, zero reviewer calls and no output file. The current same-file write guard should remain for changes after load.

### THE MASTER'S READING (claims to test)
This is the fourth alias the reviews found one at a time (same path, symlinked directory, dangling final link, link then
\`..\`). Close the class, not the instance: the identity must be computed the way the kernel resolves the path — walk the RAW
path component by component, never \`filepath.Clean\` or \`filepath.Abs\` before resolving, resolve each symlink as you meet it
(relative targets against the link's own directory, a hop bound), apply \`..\` to the resolved directory, and stop at the first
component that does not exist (the rest is appended lexically). The write-time guard stays as the second line.

### WHAT TO DO
1. **Reproduce first:** \`link -> elsewhere/dir\`; \`-o elsewhere/report --sarif-output link/../report\`; paste exit 2 today.
2. **One resolver:** implement the walk above as one function used by the load check; delete the \`Abs\`-then-resolve path.
3. **A table of aliases** in one test, each pair must exit 7 at load, name both flags, run no review, write no file:
   identical spelling; \`./\` and \`a/../\`; relative vs absolute; symlinked directory; symlinked directory then \`..\`; a chain of
   two directory links; a dangling final link; a final link to a link; a hard link to an existing file; trailing slash on the
   directory. Paste the red of the \`link/../\` row with the old \`Abs\` path restored.
4. **Agreement:** a test that, for every alias in the table, the load check refuses BEFORE the write guard could see it (the
   write guard is never the one that fires for a row of the table).
5. Nothing else; \`internal/review\` and \`AGENTS.md\` unchanged; new commits on top of 3d54504.

### CODE IS WRITTEN TO BE READ (ADR-0027)
The resolver's doc comment states the rule (links resolved before \`..\`, as the filesystem does) and why the write guard
remains; each row names the alias it covers. English, plain.

### ACCEPTANCE
1. \`git log --oneline 3d54504..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat 3d54504..HEAD -- AGENTS.md internal/review\` is EMPTY.
3. \`go vet ./...\` clean; \`go test -count=1 ./...\` green, counts per package; \`gofmt -l\` empty.
4. The red of step 3; paste it.
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
