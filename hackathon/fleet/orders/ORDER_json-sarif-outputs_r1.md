# ORDER — json-sarif-outputs · round 1

- change: json-sarif-outputs
- round: 1
- base: main @ 9303b2d
- executor: claude-code on macm5
- written: 2026-10-23T18:07:14Z
- task: W2 round 3 — pay the two reviews of 7281552: refuse colliding output paths, carry every text-only line into SARIF and JSON, table-driven schema and field tests with their unplugs; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — json-sarif-outputs
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W2
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/json-sarif-outputs
- lands-on: main
- backend: claude-code
- quote-usd: 11.00
- cap-usd: 22.00
- wall: 120m
- task: W2 round 3 — pay the two reviews of 7281552: refuse colliding output paths, carry every text-only line into SARIF and JSON, table-driven schema and field tests with their unplugs; nothing else

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
Your base `7281552` is W2 (JSON and SARIF 2.1.0 writers, `internal/cli/format_json.go`, `format_sarif.go`) as ONE commit on
proven main `9303b2d`. This stream continues wagon W2 under a new name because main's history was rewritten. The master
measured on 7281552: vet clean, `go test -count=1 ./...` 34 packages ok, 0 FAIL; `TestSampleSARIFMatchesTheSchema` PASS,
and FAIL with `Version` emptied (the reviewer's C0, paid).

### WHY THIS ROUND (the two reviews of 7281552, carried verbatim, law #44)
codex (run on █████, verdict **reject**), its one BLOCKING:
> BLOCKING · internal/cli/format.go:26-32 newly enables both writers, while internal/cli/run.go:198-240 writes primary and
> extras sequentially to caller-selected paths without a collision check. With --json -o report.json --sarif-output
> report.json, the SARIF body replaces the JSON body and both writes succeed, returning exit 0.
> Repair and test path collisions as a blocking condition. Add full-field and edge-schema tests as owed work.

claude-code (verdict accept-with-conditions), blocking before merge:
> - **C0 (UNRUN here; the master runs it)**: run `go test -count=1 -run 'TestSampleSARIFMatchesTheSchema' -v ./internal/cli/` on a seat whose module cache holds `jsonschema/█████████`. Paste PASS, plus a run with `Version: "2.1.0"` removed from `format_sarif.go:104` showing FAIL.
> - **C1 (item 2)**: put the text-only lines into the SARIF run properties: `settings_statement`, `author_model_statement`, `reviewer_statements`, `trace_path`, and the same-family sentence. Carry the family sentence into the JSON as well, and document all of them in `docs/OUTPUT_FORMATS.md`. Add a test that fails when any one is dropped.
> - **C2 (item 1)**: make the schema test table-driven over the shapes: no evidence, unresolved, failed service, empty path, line 0, `endLine < startLine`, unknown severity, and the empty result. Each must validate.
> - **C3 (item 5)**: `TestSARIFLevelFollowsSeverity` must assert each result's `level` by its id. The unplug is SM2 (swap WARNING and ERROR), which must turn the test red.
> - **C4 (items 4 and 5)**: add a decode-and-compare test, golden or struct, over every name in `docs/OUTPUT_FORMATS.md`. It must also check `[]` (not `null`) for empty lists and the SARIF `properties.evidence`, `properties.unresolved` and run `failed_services`. The unplugs are JM1 and SM1, which must turn it red.
> - **C5 (item 4)**: name `findings[].title`, `message` and `models`, and every top-level key, in the JSON table of `docs/OUTPUT_FORMATS.md`.

### WHAT TO DO
1. **Reproduce first:** in a test, `cli.Run` with `--json -o X --sarif-output X` (same path) and show it returns 0 today.
2. **Collisions (codex BLOCKING):** two outputs that resolve to the same file (`-o` and any `--*-output`, or two
   `--*-output`, compared after `filepath.Clean` and `filepath.Abs`) are refused at load with exit 7, naming both flags and
   the path (ADR-0013 form). Test each pair; red with the check removed.
3. **C1–C5** of the claude review, as written above. C3's unplug is swapping WARNING and ERROR; C4's are renaming a JSON
   field and dropping SARIF `properties.evidence`; each must turn its test red — paste the reds.
4. **Owed, take them if they are small:** C6 (a test that an output write failure is exit 2), C7 (the two extra
   `// Package cli is …` comments become plain file comments), C8 (`artifactURI` escapes `:` in a first path segment).
5. Nothing else; `internal/review` and `AGENTS.md` unchanged; never run `proven scan` inside the checkout (it writes
   `logs/` where it starts); run it from `████████████████████████████████`.

### CODE IS WRITTEN TO BE READ (ADR-0027)
Doc comments on exported names; a WHY comment at each guard and refusal; the decision named (`ADR-0020 §1`, `ADR-0013`,
`SARIF 2.1.0`); every test says which property it proves. English, plain; no internal host names or paths.

### ACCEPTANCE
1. `git log --oneline 7281552..HEAD` shows only this round's commits; paste it.
2. `git diff --stat 7281552..HEAD -- AGENTS.md internal/review` is EMPTY.
3. `go vet ./...` clean; `go test -count=1 ./...` green, counts per package; `gofmt -l` empty.
4. The reds of steps 2 and 3 with each check reverted; paste them.
5. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

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
| how an order is bound to the laws | 1 |
| machine name | 1 |
| private workspace path | 4 |
