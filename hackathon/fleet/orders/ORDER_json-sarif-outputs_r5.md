# ORDER — json-sarif-outputs · round 5

- change: json-sarif-outputs
- round: 5
- base: main @ 9303b2d
- executor: claude-code on macm5
- written: 2026-10-24T04:56:23Z
- task: W2 round 7 (stream round 5) — standard output redirected onto an output file is refused at load like any other pair; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — json-sarif-outputs
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- round: 5
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/json-sarif-outputs
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W2 round 7 (stream round 5) — standard output redirected onto an output file is refused at load like any other pair; nothing else

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
Your tip \`f4819cf\`. Both reviews of it: codex accept-with-conditions (four mutations; the master ran them, all red), claude-code
accept-with-conditions (10 unplugs killed; the schema tests and a line-0 unplug run by the master, PASS and red). One condition
remains, carried verbatim:
> - **C1 (blocking-class, item 3; `run.go:198-266` and `config.go:391`).** Standard output, when it is a regular file, is
>   not compared with the `--*-output` files. P1 reproduces it: exit 0, and one body silently replaces the other. This is
>   the same defect class as the round-2 BLOCKING. Remedy: in `Run`, before the review runs, if `stdout` is an `*os.File`
>   whose `Stat` is a regular file, `os.Stat` each `--*-output` path and compare it with `os.SameFile`. The shell has
>   already created stdout's file, so an output that is the same file exists under some spelling, and `os.Stat` follows
>   links. Also add stdout's `FileInfo` to the `written` list of the write-time guard. Refuse with exit 7, naming "standard output" and the flag. Test:
>   `TestStdoutRedirectedOntoAnOutputIsRefused`. Mutated field: stdout is an `*os.File` on the `--json-output` path.
>   Unplug: drop the stdout entry, and the test sees exit 0. If the owner rules this out of scope instead, the doc must
>   say in one sentence that standard output redirected onto an output file is not detected.
>   on APFS) as the same gap (P3). For both, say what happens: the write-time check stops the run with exit 2 and the
>   earlier output stays as written. Not blocking.
> - **C3 (owed, tests).** Add schema shapes for a finding whose message, title and id are all blank (`"finding"`, no
>   `ruleId`), for an absolute path and a path with `:` in its first segment (`artifactURI`), and for a non-sample finding
>   carrying both test and query evidence. The external `shapes` table and the internal `sarifShapes` table are kept in
>   sync by hand. A test that compares the two tables' names would stop them drifting. Not blocking.
> 
> 
> - `go test ./internal/cli/sarifschema/` (`TestSampleSARIFMatchesTheSchema`, `TestSARIFShapesMatchTheSchema`): refused
>   at setup. `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/...": tls: failed to verify
>   certificate: x509: OSStatus -26276`. The module is not in the local cache. → C0.
> - The schema test's own unplugs: same cause. → C0.
> - `go vet ./...` across `sarifschema`: same cause. `go vet ./internal/cli/` is clean.
> - The master's "38 packages ok, 0 FAIL": not reproduced. Here it is 30 ok and 7 FAIL: six are sandbox listen denials
>   outside the diff, and one is the setup failure above.
> - The built-binary check "refuses `-o` and `--sarif-output` on one file with exit 7": the binary built
>   (`go build ./cmd/proven`), but the role's command class does not admit running it. The same path was exercised
>   in-process through `cli.Run` (`TestCollisionReproduction` and the alias table, exit 7).
> - Residue: `git clean`, `rm` and `find -delete` are refused for this role. Two untracked files of mine remain, and the
>   master should remove them: `████████████████████████████████████` (cut down to the single line `package cli`, so it
>   changes no behaviour) and `██████████████████████████████████████████████` (the built binary and a sha256 helper,
>   inside the order's confinement directory). No tracked file is modified (`git diff --stat` is empty).

### WHAT TO DO
1. **Reproduce first:** \`cli.Run\` with stdout an \`*os.File\` opened on \`X\` and \`--sarif-output X\`; paste exit 0 and the replaced body.
2. **Refuse at load:** when stdout is a regular file (an \`*os.File\` whose \`Stat\` is a regular file), it takes part in the same
   identity check as \`-o\` and every \`--*-output\`: an output whose resolved path exists and is \`os.SameFile\` with stdout is refused
   with exit 7, naming the flag and "standard output", before the review runs and before anything is written. Tests: stdout vs
   each \`--*-output\`, through a symlink too; red with the check removed — paste it. A terminal or a pipe on stdout is unaffected.
3. Nothing else; new commits on top of f4819cf; \`internal/review\` and \`AGENTS.md\` unchanged.

### ACCEPTANCE
1. \`git log --oneline f4819cf..HEAD\` shows only this round's commits; paste it.
2. \`git diff --stat f4819cf..HEAD -- AGENTS.md internal/review\` is EMPTY.
3. \`go vet ./...\` clean; \`go test -count=1 ./...\` green, counts per package; \`gofmt -l\` empty.
4. The red of step 2; paste it.
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
| absolute or private path | 2 |
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
