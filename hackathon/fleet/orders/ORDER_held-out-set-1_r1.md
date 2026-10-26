# ORDER — held-out-set-1 · round 1

- change: held-out-set-1
- round: 1
- base: main @ 9b11366
- executor: claude-code
- written: 2026-10-24T13:33:47Z
- task: round 1 — write six held-out case contracts and their answer keys, and commit them as ONE freeze commit before Proven is ever run on them; no fixture code yet

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — held-out-set-1
track: proven-test-corpus
- train: proven-test-corpus
- wagon: W9
- round: 1
- review: named
- repo: proven
- base: main @ 9b11366
- branch: fix/held-out-set-1
- lands-on: main
- backend: claude-code
- quote-usd: 11
- cap-usd: 22
- wall: 120m
- task: round 1 — write six held-out case contracts and their answer keys, and commit them as ONE freeze commit before Proven is ever run on them; no fixture code yet

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as CASE AUTHOR. One-shot headless: foreground commands only; your first act writes
`_EXEC_STATUS.md` (`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT · ZERO push/merge/PR · commit subject `[held-out-set-1] <area>: <what>` matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`,
one-token area, at most 72 characters, English. No AI attribution anywhere.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-24, via the advisor (cce910fa00a2720c, 3ac42c5711359aaf): a HELD-OUT SET — fresh cases Proven has never been
run on, written and committed (frozen) BEFORE the first run, the freeze commit named; then one run, results published as
they come, misses included. A fix found by a miss is a general rule, checked on ANOTHER fresh set, never on these cases.
The set tests spread, not depth.

### WHAT THE MASTER MEASURED (2026-10-24; claims to test, law #44)
- The 21 PV cases use 23 of the population's source pairs; the 7 unused pairs are not fair tests (two are not
  disagreements, two are folded into PV-007/008, three are environment artefacts with no fixture). So this set comes from
  reviewer disagreements born in the build window, described below in the master's words, without names.
- A case's shape on main: `test/corpus/cases/<id>/CASE.md` with sections 1 Story · 2 Contract (written before any fixture)
  · 3 Provenance · 4 Synthetic transformation · 5 Review target · 6 The two claims and their evidence · 7 Expected outcome
  · 8 Reproduction plan · 9 Limits, and the answer key in `expected/EXPECTED.md`. Read PV-021's two files as the format.

### THE SIX CASES (ids HO-001 … HO-006; the advisor picked them)
Each describes a defect a code reviewer should catch in a CHANGE; the base is the code before it. Fixtures will be
synthetic and minimal; no product, vendor, host or person names anywhere (describe "a CLI", "a forge", "a test runner").
- **HO-001 stale-evidence** (sh): a check passes when ANY old success comment exists on a pull request, even though the
  current run of the workflow failed. The defect: evidence from an earlier run accepted as proof of the current one.
- **HO-002 forged-author** (sh): the check accepts a success marker that names the run and the commit, but from ANY
  account; a person can post the exact marker after a run that posted nothing. The defect: the evidence's source unchecked.
- **HO-003 umask-dependent-mode** (Go): code that copies a tree adds read bits but keeps group-write; its test passes under
  umask 022 and fails under 002. The defect: a result that depends on the environment's umask.
- **HO-004 conflated-status** (sh): a client treats an API's 422 as "already exists", but the API also answers 422 for an
  input it refuses (an email it will not take); the object is never created and a later step fails with 404.
- **HO-005 silent-fallback** (Go or sh): a CLI asked for a model it does not list silently serves another model and
  reports success; the caller records the model it asked for. The defect: a silent substitution reported as success.
- **HO-006 retired-path** (sh or Go): after a directory move, a check script still names the old path; on a machine with a
  leftover copy it passes, on a fresh clone it fails. The defect: a stale reference that only a clean checkout reveals.

### WHAT TO BUILD (this round: TEXT ONLY, no base/change/check code)
1. For each case, `test/corpus/held-out/HO-00N/CASE.md` (§1–§9 as on main; §2 the contract a test author will build
   from, WITHOUT the answer; §8 left as "to be recorded by the fixture round"; §3 provenance: "build-window reviewer
   disagreement, 2026-10-23/24, described by the master; no source text is quoted") and `expected/EXPECTED.md` (the answer
   key: the defect, where it is, and what a correct review says).
2. `test/corpus/held-out/manifest.json` listing the six, with `"frozen": true` and the base commit, and
   `test/corpus/held-out/README.md`: what a held-out set is, the rule (never run before the freeze; a fix is checked on
   another set), and where results will be published.
3. ONE commit holding all of it: subject `[held-out-set-1] corpus: freeze held-out set 1 (six cases, no run yet)`.
   Nothing else in that commit. Then a second commit may update the main README's corpus section to name the freeze
   commit (by its hash) — never the same commit.

### ACCEPTANCE
- `git show --stat <freeze commit>` lists only `test/corpus/held-out/**`; paste it.
- `make check-links` exit 0; `go vet ./...` and `go test ./...` unaffected (no code added).
- No name of a product, vendor, host or person in the six cases (grep your own text and paste the grep).
- `git log main..HEAD` has no merge commit and no `Co-authored-by`.

### THIS ORDER CAN BE WRONG (law #44)
Write a `- master-was-wrong:` line: `—`, or what contradicts this order (for example, a case the corpus already covers).

### WHAT IS NOT IN SCOPE
Fixture code (next round, by test authors who never read `expected/`); running Proven on these cases (only after the freeze
lands, with the W8 command).

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse HEAD` output
verbatim; never restate HEAD from memory.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████` in the worktree, except the committed files named
above. Nothing outside the worktree.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, `- abstract:` at most 5 lines naming the freeze
commit, the acceptance lines run and not run. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`.

### ATTRIBUTION (owner, 2026-10-23)
No commit of yours may carry a `Co-authored-by:` trailer or a seat identity as its author. Check
`git log --format="%an <%ae>%n%B" main..HEAD` before you end and say so in your status.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
