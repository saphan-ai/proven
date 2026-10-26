# ORDER — a-live-review-reads-any-repository · round 1

- change: a-live-review-reads-any-repository
- round: 1
- base: main @ 2d3126b
- executor: cursor on runner-cursor
- written: 2026-10-23T05:29:10Z
- task: a live review reads a plain git repository (two commits, or a commit and the working tree) with no fixture.json and no base/ and change/ directories prepared by hand

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — a-live-review-reads-any-repository
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W21
- review: master-verify-only
- repo: proven
- base: main @ 2d3126b
- branch: fix/a-live-review-reads-any-repository
- lands-on: main
- worktree: ██████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 150m
- task: a live review reads a plain git repository (two commits, or a commit and the working tree) with no fixture.json and no base/ and change/ directories prepared by hand

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[a-live-review-reads-any-repository] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE DEFECT (the master's reading, 2026-10-23 — a CLAIM for you to test, law #44)
`internal/pipeline/pipeline.go` `review()` calls `fixture.Load(req.Target)` in sample AND live mode (line ~258).
`fixture.Load` (`internal/fixture/fixture.go`) needs `<target>/fixture.json` whose planner-prompt hash matches, and reads
the trees `<target>/base` and `<target>/change` (`change.Read`). `planner.Plan(ctx, req.Target, gen)` reads the same two
trees. So a live review of an ordinary repository fails before any agent runs; only the prepared `demo/live/` target
works. In live mode the fixture's fields used are only `fx.Root` (the trees the Prover mounts, lines ~386 and ~749) and
`fx.Snapshot` (the evidence snapshot name, lines ~366, ~502, ~511); `Answers`, `Tables`, `Tests` are sample-only.
**Step 1, before any change: reproduce it.** In a scratch git repo with two commits and no fixture.json, run a live-mode
review through `review.Open` + `Review` (a test with a stub generator and query is enough; no network) and paste the
error. If it does NOT fail, write that in `- master-was-wrong:` and stop with `- state: blocked`.

### WHAT TO BUILD
1. Live mode needs no `fixture.json`. Sample mode is unchanged (it still loads and checks the fixture).
2. A live target is one of: (a) a directory that already has `base/` and `change/` (the demo, unchanged); (b) a git
   repository with `BaselineCommit` set: the reviewer writes the two trees into a temporary directory under the run's
   working area — `base/` from `BaselineCommit`, `change/` from `HeadCommit`, or from the working tree when `HeadCommit`
   is empty — and passes that directory to the Planner and the Prover. The temporary directory is removed at the end of
   the review. Only tracked files go in (`git archive` or `git ls-files`), never `.git`.
3. A git target with no `BaselineCommit` and no `base/`/`change/`: refused with a sentence that says what to pass.
4. The live snapshot name (what `fx.Snapshot` gave): a named value, not empty, written in the trace; say in a comment
   where it comes from.
5. `docs/ARCHITECTURE.md` (Modes) and `docs/INSTALL.md`: one sentence each — a live review needs only a git repository.
6. ADR-0027 comments on everything you add (see below).

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. A test: a live review of a scratch git repo (two commits, no fixture.json; stub generator, query and test runner, no
   network) reaches the Planner and returns a result; the planner prompt contains a file changed between the commits.
3. A test: working-tree mode (`HeadCommit` empty) sees an uncommitted edit.
4. A test: the temporary trees are gone after the review.
5. A test: sample mode without fixture.json still fails as before.
6. **Negative control (law #27):** put the `fixture.Load` call back on the live path → test 2 turns red; restore.
7. `proven scan --sample` on `demo/` still prints the same review (paste it).

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Any network call in tests; the CLI (`cmd/proven`, `internal/cli` — another train); the Prover's container settings.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████████████████` in the worktree, except the
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
