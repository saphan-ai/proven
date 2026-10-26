# ORDER — corpus-pv-013 · round 3

- change: corpus-pv-013
- round: 3
- base: main @ fe146e0
- executor: codex
- written: 2026-10-23T02:17:24Z
- task: round 2 — pay the two reviews' conditions on PV-013's case code (tip bf6722d); nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — corpus-pv-013
track: proven-test-corpus
- train: proven-test-corpus
- wagon: W2
- round: 2
- review: master-verify-only
- repo: proven
- base: main @ fe146e0
- branch: fix/corpus-pv-013
- lands-on: main
- quote-usd: 11
- cap-usd: 22
- wall: 45m
- task: round 2 — pay the two reviews' conditions on PV-013's case code (tip bf6722d); nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as TEST AUTHOR of one corpus case. One-shot headless: foreground commands only; your first act
writes `_EXEC_STATUS.md` (`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[corpus-pv-013] tests: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-19 (translated from Polish): *"agreed, we build now; the test code is "prior art", but it must exist to show how
Proven works."* 2026-10-21 (translated): *"since tomorrow they write the competition work, we have to warm them up a bit
and find the bugs ... I'd rather catch them today than tomorrow."* A failure you report today is a result, not a setback.

### ROUND 2 — THE REVIEWS' CONDITIONS (carried verbatim, law #44)
Your round 1 is commit bf6722d on `fix/corpus-pv-013` (carried unchanged onto proven's rewritten main). Two independent
reviews judged it; their Conditions sections, verbatim:

#### codex review (corpus-pv-013-review-r2, reject)
Repair the combined leak flag and add a failing negative control as above; rerun the seven contract inputs and verify the explicit change-side refusal still prints the synthetic value. No accepted exception or downgrade of this blocking condition is recommended.

Also, for every case (master-was-wrong): round 1's order told you to set the manifest status to the string `built`. That
broke `tests/corpus/manifest.json`: every other case keeps the status OBJECT (`contract`, `fixture`, `checks`,
`independent_check`) and `tests/README.md`'s `jq` count reads it. Restore PV-013's status to that object, with your
real dates, e.g. `{"contract": "written 2026-10-19", "fixture": "built 2026-10-22", "checks": "run 2026-10-22",
"independent_check": "not done"}`.

WHAT TO DO in this round: pay each condition above (or, where you measure a condition to be wrong, say so with the
measurement in `- master-was-wrong:` and leave it); keep `expected/` unread; rerun `check.sh` and record its full output
and exit code in CASE.md §8; make the check portable to Linux and macOS (no GNU-only flags). Same branch, new commits on
top of bf6722d. ACCEPTANCE: `git diff --stat fe146e0..HEAD` lists only `tests/corpus/cases/PV-013/` (outside `expected/`) and
`tests/corpus/manifest.json`; `bash tests/corpus/cases/PV-013/check.sh` output and exit pasted.

### WHAT YOU WORK FROM, AND WHAT YOU DO NOT READ
You work from `tests/corpus/cases/PV-013/CASE.md` at your base: its §2 Contract is your specification. **Do not open
`tests/corpus/cases/PV-013/expected/` or any other case's `expected/`**: the expected outcome is withheld from the author on
purpose, and a fixture written towards it proves nothing. Do not read other branches, worktrees or this case's orders.

### WHAT TO BUILD
1. Under `tests/corpus/cases/PV-013/` only: `base/` and `change/` — the two versions §2 describes — and one check script
   (`check.sh`) that runs §2's inputs against both and prints one line per input exactly as §2 asks. Language: Go
   (§2 leaves the toolchain to you); use the Go standard library only; Go as installed on your host (record `go version`). No network beyond localhost servers the check itself starts; no model,
   no keys, no access to anything outside the worktree.
2. Run `check.sh` against base and change. **Report what it prints; do not change base or change to make a line come
   out the way you expect.** If §2 says change ships its own control, build that control as §2 describes it, even if
   you believe it is weak — the case exists because it is weak.
3. In `CASE.md` §8 (Reproduction plan): replace "not run" with the exact commands, the toolchain versions, the date and
   time (UTC), and the full output of `check.sh` with its exit code. In the header line, set `case code:` to built
   (date, declared prior art) and `checks:` to run (date). Change nothing else in CASE.md.
4. In `tests/corpus/manifest.json`: this case's status only, to `built`; nothing else.

### ACCEPTANCE
1. `bash tests/corpus/cases/PV-013/check.sh` runs from the repository root and prints one line per §2 input; its output
   and exit code are pasted in §8.
2. `git diff --stat main...HEAD` lists only files under `tests/corpus/cases/PV-013/` (outside `expected/`) and
   `tests/corpus/manifest.json`.
3. Your `- abstract:` names the seat's model as the run reports it, and states in one line what base and change printed
   for the input the case is about.

### THIS ORDER CAN BE WRONG (law #44)
The master wrote this order from the contract without building it. If §2 cannot be built as written — a contradiction,
an input that cannot exist, a property that no check can observe — write `SHOUT_contract-pv-013.md` saying exactly
which sentence and why, set `- state: blocked`, and stop. That is a correct result, not a failure.

### WHAT IS NOT IN SCOPE
- Any file outside `tests/corpus/cases/PV-013/` and the one manifest line; any `expected/` directory; Proven's own code.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run
and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code
is the verdict. Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
