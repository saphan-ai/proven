# ORDER — corpus-pv-012 · round 3

- change: corpus-pv-012
- round: 3
- base: main @ 9303b2d
- executor: codex
- written: 2026-10-24T06:58:52Z
- task: round 3 — attest the master-carried tip 479cfb8: confirm it, run the check, refile the status; NO code change

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — corpus-pv-012
track: proven-test-corpus
- train: proven-test-corpus
- wagon: W7
- round: 3
- review: master-verify-only
- repo: proven
- base: main @ 9303b2d
- branch: fix/corpus-pv-012
- lands-on: main
- backend: codex
- quote-usd: 3
- cap-usd: 6
- wall: 30m
- task: round 3 — attest the master-carried tip 479cfb8: confirm it, run the check, refile the status; NO code change

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as TEST AUTHOR of one corpus case. One-shot headless: foreground commands only; your first act
writes `_EXEC_STATUS.md` (`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[corpus-pv-012] tests: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-19 (translated from Polish): *"agreed, we build now; the test code is "prior art", but it must exist to show how
Proven works."* 2026-10-22 (translated): *"have we written all the test cases? They do not have to be written with Cursor."* What
lands after 2026-10-22 00:00 UTC is dated as such and declared, never back-dated. A failure you report is a result, not a setback.

### ROUND 3 — ATTEST THE CARRIED TIP (master, 2026-10-24; NO code change)
Your round 2 was reviewed by two families. Since then the master carried your branch onto proven's rewritten main (no AI
trailer, no seat author; your case files identical) and replaced the retired `tests/corpus` path with `test/corpus` in
your case (479cfb8). One codex review of 479cfb8 accepted it. Your far worktree held the pre-rewrite history and was cleared; this
worktree is fresh at 479cfb8. What you owe is the status that describes this tip:
1. `git rev-parse --short=7 HEAD` must print `479cfb8`; otherwise `- state: blocked` and stop.
2. Run `bash test/corpus/cases/PV-012/check.sh` from the repository root and compare it with CASE.md §8 (test timings
   aside). Paste the output into your status. If it differs, say how, and do NOT edit anything: `- state: blocked`.
3. Write `_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` the full SHA of `479cfb8`, `- done:` naming the master's
   two acts (carry, path fix) and your check run. Make NO commit; the branch stays at 479cfb8.
**master-was-wrong:** if HEAD, the check or the path fix is not as stated, show the measurement.

### ROUND 2 — THE REVIEWS' CONDITIONS (carried verbatim, law #44)
Your round 1 is commit 657ce27 on `fix/corpus-pv-012` (carried unchanged onto proven's rewritten main). Two independent
reviews judged it; their Conditions sections, verbatim:

#### codex review (corpus-pv-012-review, reject)

No acceptance conditions: reject this revision. Repair both blocking findings and repeat the same check with an ownership negative control. The master must resolve the §2 request wording before a new verdict. Ruling: I treat the literal request-team requirement as operative because the fixture has a separate `request.team` field; if the intended contract used “request” to mean page construction, that wording needs correction, and the cost of this ruling is one unnecessary fixture revision.


#### claude review (corpus-pv-012-review-r2, accept-with-conditions)

- C1 — the master (or a seat that admits `bash`) runs `bash tests/corpus/cases/PV-012/check.sh` at `657ce27` and confirms
  its seven lines equal the Differential table above and CASE.md §8; this reviewer could not (Unrun). Class: blocking
  until done.
- C2 — F4: separate the input driver and mutation switch from the review target, or state in §5 what the target is,
  before PV-012 is used in an evaluation. Class: owed.
- C3 — F2 and F3 in `check.sh`. Class: owed.
- C4 — F1 and F6, one sentence each in CASE.md §8. Class: owed.

Declined to judge (R11): licences (out of scope by order); whether `.saphan` as a name is "private" (the repo is
`github.com/saphan-ai/proven` with a root `SAPHAN.md`, so I judged it public and raised only the convention, F3).

Also, for every case (master-was-wrong): round 1's order told you to set the manifest status to the string `built`. That
broke `tests/corpus/manifest.json`: every other case keeps the status OBJECT (`contract`, `fixture`, `checks`,
`independent_check`) and `tests/README.md`'s `jq` count reads it. Restore PV-012's status to that object, with your
real dates, e.g. `{"contract": "written 2026-10-19", "fixture": "built 2026-10-22", "checks": "run 2026-10-22",
"independent_check": "not done"}`.

WHAT TO DO in this round: pay each condition above (or, where you measure a condition to be wrong, say so with the
measurement in `- master-was-wrong:` and leave it); keep `expected/` unread; rerun `check.sh` and record its full output
and exit code in CASE.md §8; make the check portable to Linux and macOS (no GNU-only flags). Same branch, new commits on
top of 657ce27. ACCEPTANCE: `git diff --stat fe146e0..HEAD` lists only `tests/corpus/cases/PV-012/` (outside `expected/`) and
`tests/corpus/manifest.json`; `bash tests/corpus/cases/PV-012/check.sh` output and exit pasted.

### WHAT YOU WORK FROM, AND WHAT YOU DO NOT READ
You work from `tests/corpus/cases/PV-012/CASE.md` at your base: its §2 Contract is your specification. **Do not open
`tests/corpus/cases/PV-012/expected/` or any other case's `expected/`**: the expected outcome is withheld from the author on
purpose, and a fixture written towards it proves nothing. Do not read other branches, worktrees or this case's orders.

### WHAT TO BUILD
1. Under `tests/corpus/cases/PV-012/` only: `base/` and `change/` — the two versions §2 describes — and one check script
   (`check.sh`) that runs §2's inputs against both and prints one line per input exactly as §2 asks. Language: the one
   CASE.md's header names as the intended fixture language (a hint; §2 decides). Go: the standard library only, Go as installed on your host. TypeScript/JavaScript: Node's built-in modules only, no `npm install`, Node as installed on your host. Record the toolchain version. If the named toolchain is absent on your host, say so in `_EXEC_STATUS.md` and use Go. No network beyond localhost servers the check itself starts; no model,
   no keys, no access to anything outside the worktree.
2. Run `check.sh` against base and change. **Report what it prints; do not change base or change to make a line come
   out the way you expect.** If §2 says change ships its own control, build that control as §2 describes it, even if
   you believe it is weak — the case exists because it is weak.
3. In `CASE.md` §8 (Reproduction plan): replace "not run" with the exact commands, the toolchain versions, the date and
   time (UTC), and the full output of `check.sh` with its exit code. In the header line, set `case code:` to built
   (date, declared: built inside the build window) and `checks:` to run (date). Change nothing else in CASE.md.
4. In `tests/corpus/manifest.json`: this case's status only, to `built`; nothing else.

### ACCEPTANCE
1. `bash tests/corpus/cases/PV-012/check.sh` runs from the repository root and prints one line per §2 input; its output
   and exit code are pasted in §8.
2. `git diff --stat main...HEAD` lists only files under `tests/corpus/cases/PV-012/` (outside `expected/`) and
   `tests/corpus/manifest.json`.
3. Your `- abstract:` names the seat's model as the run reports it, and states in one line what base and change printed
   for the input the case is about.

### THIS ORDER CAN BE WRONG (law #44)
The master wrote this order from the contract without building it. If §2 cannot be built as written — a contradiction,
an input that cannot exist, a property that no check can observe — write `SHOUT_contract-pv-012.md` saying exactly
which sentence and why, set `- state: blocked`, and stop. That is a correct result, not a failure.

### WHAT IS NOT IN SCOPE
- Any file outside `tests/corpus/cases/PV-012/` and the one manifest line; any `expected/` directory; Proven's own code.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run
and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code
is the verdict. Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md` before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim; never restate HEAD from memory.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
