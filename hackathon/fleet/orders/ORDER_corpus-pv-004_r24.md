# ORDER — corpus-pv-004 · round 24

- change: corpus-pv-004
- round: 24
- base: main @ 9303b2d
- executor: codex
- written: 2026-10-24T06:09:45Z
- task: round 3 — attest the master-carried tip 04c3024: confirm it, run the check, refile the status; NO code change

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — corpus-pv-004
track: proven-test-corpus
- train: proven-test-corpus
- wagon: W2
- round: 3
- review: master-verify-only
- repo: proven
- base: main @ 9303b2d
- branch: fix/corpus-pv-004
- lands-on: main
- backend: codex
- quote-usd: 3
- cap-usd: 6
- wall: 30m
- task: round 3 — attest the master-carried tip 04c3024: confirm it, run the check, refile the status; NO code change

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as TEST AUTHOR of one corpus case. One-shot headless: foreground commands only; your first act
writes `_EXEC_STATUS.md` (`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[corpus-pv-004] tests: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ███████████████████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-19 (translated from Polish): *"agreed, we build now; the test code is "prior art", but it must exist to show how
Proven works."* 2026-10-21 (translated): *"since tomorrow they write the competition work, we have to warm them up a bit
and find the bugs ... I'd rather catch them today than tomorrow."* A failure you report today is a result, not a setback.

### ROUND 3 — ATTEST THE CARRIED TIP (master, 2026-10-24; NO code change)
Your round 2 was reviewed by two families. Since then the master carried your branch onto proven's rewritten main (no AI
trailer, no seat author; your case files identical) and replaced the retired `tests/corpus` path with `test/corpus` in
your case (04c3024). One codex review of 04c3024 accepted it. Your far worktree held the pre-rewrite history and was cleared; this
worktree is fresh at 04c3024. What you owe is the status that describes this tip:
1. `git rev-parse --short=7 HEAD` must print `04c3024`; otherwise `- state: blocked` and stop.
2. Run `bash test/corpus/cases/PV-004/check.sh` from the repository root and compare it with CASE.md §8 (test timings
   aside). Paste the output into your status. If it differs, say how, and do NOT edit anything: `- state: blocked`.
3. Write `_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` the full SHA of `04c3024`, `- done:` naming the master's
   two acts (carry, path fix) and your check run. Make NO commit; the branch stays at 04c3024.
**master-was-wrong:** if HEAD, the check or the path fix is not as stated, show the measurement.

### ROUND 2 — THE REVIEWS' CONDITIONS (carried verbatim, law #44)
Your round 1 is commit d786c35 on `fix/corpus-pv-004` (carried unchanged onto proven's rewritten main). Two independent
reviews judged it; their Conditions sections, verbatim:

#### codex review (corpus-pv-004-review-r4, reject)

Reject until both blocking items are resolved. The output assertion is owed work. No condition was reclassified. Ruling: apply the explicit 21 October provenance and no-network clause even though the manual differential is sound; if wrong, the cost is an unnecessary rebuild and offline packaging step. Declined to judge: the private original reviews and licenses, which are outside this review's evidence and scope.


#### claude review (corpus-pv-004-review-r5, accept-with-conditions)
- **C-1 · `tests/corpus/manifest.json:76`. BLOCKING CLASS (it breaks a published instrument).**
  - The diff replaces PV-004's `status` OBJECT (`contract`, `fixture`, `checks`, `independent_check`) with the STRING
    `"built"`.
  - All other 20 cases keep the object (`grep -n '"status"' tests/corpus/manifest.json`: line 76 is the only string).
  - `tests/README.md:66-71` publishes the count command `jq '… select(.status.contract|startswith("written")) …
    .status.fixture …'`, and `tests/README.md:76-77` documents "four statuses (contract, case code, checks, independent
    check)".
  - On a string, `.status.contract` is a jq error ("Cannot index string with \"contract\""). The README's count command
    stops working, and it loses the contract date and the independent-check state. This is inferred from jq
    semantics; `jq` is not runnable on this seat.
  - Remedy: restore the object, for example `{"contract": "written 2026-10-19", "fixture": "built 2026-10-21 (prior
    art)", "checks": "run 2026-10-21", "independent_check": "not done"}`, matching `CASE.md:5`.
- **C-2 · Unrun execution.** A role whose class admits `bash`, `npm` and `node` runs `bash tests/corpus/cases/PV-004/check.sh`
  at `d786c35` on Linux AND on macOS, and confirms that stdout equals `CASE.md:110-120` byte for byte and that it exits 0.
  Until then, properties 1 and 6 rest on a source trace.

Also, for every case (master-was-wrong): round 1's order told you to set the manifest status to the string `built`. That
broke `tests/corpus/manifest.json`: every other case keeps the status OBJECT (`contract`, `fixture`, `checks`,
`independent_check`) and `tests/README.md`'s `jq` count reads it. Restore PV-004's status to that object, with your
real dates, e.g. `{"contract": "written 2026-10-19", "fixture": "built 2026-10-22", "checks": "run 2026-10-22",
"independent_check": "not done"}`.

WHAT TO DO in this round: pay each condition above (or, where you measure a condition to be wrong, say so with the
measurement in `- master-was-wrong:` and leave it); keep `expected/` unread; rerun `check.sh` and record its full output
and exit code in CASE.md §8; make the check portable to Linux and macOS (no GNU-only flags). Same branch, new commits on
top of d786c35. ACCEPTANCE: `git diff --stat fe146e0..HEAD` lists only `tests/corpus/cases/PV-004/` (outside `expected/`) and
`tests/corpus/manifest.json`; `bash tests/corpus/cases/PV-004/check.sh` output and exit pasted.

### WHAT YOU WORK FROM, AND WHAT YOU DO NOT READ
You work from `tests/corpus/cases/PV-004/CASE.md` at your base: its §2 Contract is your specification. **Do not open
`tests/corpus/cases/PV-004/expected/` or any other case's `expected/`**: the expected outcome is withheld from the author on
purpose, and a fixture written towards it proves nothing. Do not read other branches, worktrees or this case's orders.

### WHAT TO BUILD
1. Under `tests/corpus/cases/PV-004/` only: `base/` and `change/` — the two versions §2 describes — and one check script
   (`check.sh`) that runs §2's inputs against both and prints one line per input exactly as §2 asks. Language: TypeScript
   (§2 leaves the toolchain to you); use a minimal TypeScript toolchain pinned by exact version in the case's own `package.json` and lock file, installed from the npm registry; no other dependency unless the contract needs it (record `node --version` and `npm --version`). No network beyond localhost servers the check itself starts; no model,
   no keys, no access to anything outside the worktree.
2. Run `check.sh` against base and change. **Report what it prints; do not change base or change to make a line come
   out the way you expect.** If §2 says change ships its own control, build that control as §2 describes it, even if
   you believe it is weak — the case exists because it is weak.
3. In `CASE.md` §8 (Reproduction plan): replace "not run" with the exact commands, the toolchain versions, the date and
   time (UTC), and the full output of `check.sh` with its exit code. In the header line, set `case code:` to built
   (date, declared prior art) and `checks:` to run (date). Change nothing else in CASE.md.
4. In `tests/corpus/manifest.json`: this case's status only, to `built`; nothing else.

### ACCEPTANCE
1. `bash tests/corpus/cases/PV-004/check.sh` runs from the repository root and prints one line per §2 input; its output
   and exit code are pasted in §8.
2. `git diff --stat main...HEAD` lists only files under `tests/corpus/cases/PV-004/` (outside `expected/`) and
   `tests/corpus/manifest.json`.
3. Your `- abstract:` names the seat's model as the run reports it, and states in one line what base and change printed
   for the input the case is about.

### THIS ORDER CAN BE WRONG (law #44)
The master wrote this order from the contract without building it. If §2 cannot be built as written — a contradiction,
an input that cannot exist, a property that no check can observe — write `SHOUT_contract-pv-004.md` saying exactly
which sentence and why, set `- state: blocked`, and stop. That is a correct result, not a failure.

### WHAT IS NOT IN SCOPE
- Any file outside `tests/corpus/cases/PV-004/` and the one manifest line; any `expected/` directory; Proven's own code.

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
