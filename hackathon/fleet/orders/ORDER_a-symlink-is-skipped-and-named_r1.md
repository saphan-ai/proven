# ORDER — a-symlink-is-skipped-and-named · round 1

- change: a-symlink-is-skipped-and-named
- round: 1
- base: main @ 90d6951
- executor: cursor on runner-cursor
- written: 2026-10-23T08:39:00Z
- task: a live review of a git repository that tracks a symlink skips the link and names it, instead of refusing the whole repository

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-symlink-is-skipped-and-named
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W26
- review: master-verify-only
- repo: proven
- base: main @ 90d6951
- branch: fix/a-symlink-is-skipped-and-named
- lands-on: main
- worktree: ██████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 90m
- task: a live review of a git repository that tracks a symlink skips the link and names it, instead of refusing the whole repository

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-symlink-is-skipped-and-named] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23, deciding finding F31: skip the
symlink and say so in the trace (option "Pomijaj + ślad").

### THE DEFECT (the master's reading, 2026-10-23 — a CLAIM for you to test, law #44)
W21 (on main) builds the live trees from git: `internal/change/archive.go` `untar` returns an error on a
`tar.TypeSymlink`, and `copyTracked` returns an error on a symlink. So a live review of a repository that tracks any
symbolic link fails as a whole. Many real repositories track one.
**Step 1, before any change:** a failing test — a scratch git repo with two commits, one of which adds a tracked symlink
next to a changed regular file; a live review (stub generator, query and test runner, as in
`internal/review/git_live_test.go`) returns an error at the base. Paste it. If it does not fail, write that in
`- master-was-wrong:` and stop with `- state: blocked`.

### WHAT TO BUILD
1. A tracked symlink is not written into the trees the Prover mounts (base or change), in commit mode and in
   working-tree mode. The link is never followed.
2. Every skipped link is named: one trace event per path (`Kind: "skipped"`, the path, the reason "symlink"), and the
   review result names the skipped paths so a reader sees what was not reviewed (pick the existing field that fits, or
   add one to `review.Result` with a doc comment; say which in the status).
3. `change.Read` (prepared base/ and change/ trees, the demo) keeps refusing a symlink: a hand-made tree has no reason to
   carry one.
4. docs/ARCHITECTURE.md (Modes): one sentence.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. The step-1 test passes: the review completes, the changed regular file reaches the Planner, the link is named.
3. A test: the link's target outside the repository is never read (its content does not appear anywhere in the result
   or the trace).
4. **Negative control (law #27):** restore the refusal in `untar` → test 2 turns red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Submodules; any network; sample mode.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████████████` in the worktree, except the
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
