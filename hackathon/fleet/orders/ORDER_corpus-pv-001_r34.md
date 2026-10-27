# ORDER — corpus-pv-001 · round 34

- change: corpus-pv-001
- round: 34
- base: main @ fe146e0
- executor: codex
- written: 2026-10-23T08:28:25Z
- task: round 2 — pay the two reviews' conditions on PV-001's case code (tip cec11bd); nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — corpus-pv-001
track: proven-test-corpus
- train: proven-test-corpus
- wagon: W2
- round: 2
- review: master-verify-only
- repo: proven
- base: main @ fe146e0
- branch: fix/corpus-pv-001
- lands-on: main
- quote-usd: 11
- cap-usd: 22
- wall: 45m
- task: round 2 — pay the two reviews' conditions on PV-001's case code (tip cec11bd); nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as TEST AUTHOR of one corpus case. One-shot headless: foreground commands only; your first act
writes `_EXEC_STATUS.md` (`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[corpus-pv-001] tests: <what>`.
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

### ROUND 2 — THE REVIEWS' CONDITIONS (carried verbatim, law #44)
Your round 1 is commit cec11bd on `fix/corpus-pv-001` (carried unchanged onto proven's rewritten main). Two independent
reviews judged it (both accepted it, with conditions); their Conditions sections, verbatim:

#### codex review (corpus-pv-001-review-r5, accept-with-conditions)

Resolve the OWED evidence-format gap in Closure before treating the printed check line as proof of the *first* duplicate value. The master can use the measured direct probe as evidence if the answer key is aligned. No condition changes the established-defect finding.


#### claude review (corpus-pv-001-review-r6, accept-with-conditions)

1. **C1 · BLOCKING · F1** — strip the answer-revealing comments from `tests/corpus/cases/PV-001/change/guard.ts:1-3,150` and `base/guard.ts:1-3`. Control: `grep -nE 'standard|case exists|keeps the first|also matches' tests/corpus/cases/PV-001/base/guard.ts tests/corpus/cases/PV-001/change/guard.ts` prints nothing.
2. **C2 · BLOCKING · Unrun** — a role that may run node runs `npm ci --prefix tests/corpus/cases/PV-001 && bash tests/corpus/cases/PV-001/check.sh` at the tip and pastes the output. It must match `CASE.md:98-112` line for line, or §8 is re-recorded from that run. This review's Q1 rests on a static trace only.
3. **C3 · OWED · F2** — print the frame ids each side read, and re-record §8.
4. **C4 · OWED · F3** — remove the `break` in `reference/read.ts:14`, and record the reference ids for the duplicate input.
5. **C5 · OWED** — ignore `tests/corpus/cases/*/node_modules/` and `tests/corpus/cases/*/dist/`. This touches `.gitignore`, which is outside this diff's lane, so the master routes it.
6. **C6 · OWNER** — rule on prior-art case code against `AGENTS.md` §3's window rule (Remainder).
7. NOTE — `manifest.json:26` loses the prior-art marker. EXPECTED.md:5 `checkFrames` ≠ code's `judge`. No `engines` floor in `package.json`. `CASE.md:79` uses GNU-only `ls --full-time`.

Also, for every case (master-was-wrong): round 1's order told you to set the manifest status to the string `built`. That
broke `tests/corpus/manifest.json`: every other case keeps the status OBJECT (`contract`, `fixture`, `checks`,
`independent_check`) and `tests/README.md`'s `jq` count reads it. Restore PV-001's status to that object, with your
real dates, e.g. `{"contract": "written 2026-10-19", "fixture": "built 2026-10-22", "checks": "run 2026-10-22",
"independent_check": "not done"}`.

WHAT TO DO in this round: pay each condition above (or, where you measure a condition to be wrong, say so with the
measurement in `- master-was-wrong:` and leave it); keep `expected/` unread; rerun `check.sh` and record its full output
and exit code in CASE.md §8; make the check portable to Linux and macOS (no GNU-only flags). Same branch, new commits on
top of cec11bd. ACCEPTANCE: `git diff --stat fe146e0..HEAD` lists only `tests/corpus/cases/PV-001/` (outside `expected/`) and
`tests/corpus/manifest.json`; `bash tests/corpus/cases/PV-001/check.sh` output and exit pasted.

### WHAT YOU WORK FROM, AND WHAT YOU DO NOT READ
You work from `tests/corpus/cases/PV-001/CASE.md` at your base: its §2 Contract is your specification. **Do not open
`tests/corpus/cases/PV-001/expected/` or any other case's `expected/`**: the expected outcome is withheld from the author on
purpose, and a fixture written towards it proves nothing. Do not read other branches, worktrees or this case's orders.

### WHAT TO BUILD
1. Under `tests/corpus/cases/PV-001/` only: `base/` and `change/` — the two versions §2 describes — and one check script
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
1. `bash tests/corpus/cases/PV-001/check.sh` runs from the repository root and prints one line per §2 input; its output
   and exit code are pasted in §8.
2. `git diff --stat main...HEAD` lists only files under `tests/corpus/cases/PV-001/` (outside `expected/`) and
   `tests/corpus/manifest.json`.
3. Your `- abstract:` names the seat's model as the run reports it, and states in one line what base and change printed
   for the input the case is about.

### THIS ORDER CAN BE WRONG (law #44)
The master wrote this order from the contract without building it. If §2 cannot be built as written — a contradiction,
an input that cannot exist, a property that no check can observe — write `SHOUT_contract-pv-001.md` saying exactly
which sentence and why, set `- state: blocked`, and stop. That is a correct result, not a failure.

### WHAT IS NOT IN SCOPE
- Any file outside `tests/corpus/cases/PV-001/` and the one manifest line; any `expected/` directory; Proven's own code.

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
