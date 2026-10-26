# REVIEW — corpus-pv-004 · round 3

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:23:41Z

## Summary

Traced by hand, PV-004's fixture reproduces §2's condition: change accepts both malformed openers, the reference reads neither as a fence, and the well-formed line and the green suite are the controls. The lane holds and nothing private is in it.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-004

### Ground
- `git rev-parse --short=7 HEAD` printed `d786c35`, as the order requires. Scope `fe146e0...d786c35` is one commit:
  `d786c35 [corpus-pv-004] tests: build the fenced-block case and record the check` (`git log --oneline fe146e0..d786c35`).
- I read every file the diff adds or changes: `.gitignore`, `CASE.md`, `base/extract.ts`, `change/extract.ts`,
  `change/suite.ts`, `check.sh`, `check/reference.ts`, `check/run.ts`, `package.json`, `package-lock.json`,
  `tsconfig.json`, `types/ambient.d.ts`, `tests/corpus/manifest.json`. After that I read `expected/EXPECTED.md` (item 2),
  plus `tests/README.md` §4–5, because it is the manifest's documented reader.
- I did NOT run the check. This seat's permission class (`██████████████████████████`) allows only
  `saphan doc check go git ls cat grep find wc head tail diff pwd cut sort`. `bash`, `node`, `npm` and `jq` are refused
  (verbatim refusal under Unrun). Everything below about runtime output comes from tracing the source by hand against
  CommonMark 0.31 §4.5. It is not a measurement.
- The master's claims, re-measured:
  - "touches only its own case directory and manifest line, never `expected/`": CONFIRMED. `git diff --name-only`
    lists 12 paths under `tests/corpus/cases/PV-004/` (none under `expected/`) plus `tests/corpus/manifest.json`, and that
    diff hunk touches only the PV-004 object.
  - "`check.sh` exit 0": UNVERIFIED (I could not run it).
- Set aside under R1: the author's own §8 claims are treated as claims, and I re-derived each line below.

### Differential
I traced the 11 inputs through `base/extract.ts`, `change/extract.ts` and CommonMark 0.31 §4.5 (the reference). Every
line I derived matches the stdout recorded in `CASE.md:110-120` byte for byte:
- **Eight shipped cases** (opener ```` ```` ````, 4 backticks, `change/suite.ts:4`):
  - base: `command:echo hello` on all eight. Its `FENCE` (`base/extract.ts:5`) closes on any fence line, and EOF returns
    the content.
  - change: the classes in `change/extract.ts:41-49`:
    - `unclosed` → `unclosed-fence`
    - `"```"` / `"`````"` / `"``````"` / `"~~~~"` / `"~~~"` → `closing-fence-wrong-length-or-character` (`fence.length !== openLen`, line 44)
    - `"```` trailing"` / `"````note"` → `text-after-closing-fence` (`rest.length > 0`, line 41)
  - reference:
    - `"`````"` and `"``````"` close the block, because a closer may be longer than the opener → `echo hello`.
    - The other lines are content, so the block runs to the end of the document → `echo hello` + that line.
    - `unclosed` → `echo hello`.
- **well-formed**: all three return `echo hello`.
- **info-string-backtick** (`"``` foo ` bar"`):
  - change trims the line and matches `^(`{3,}|~{3,})(.*)$`. It never reads `match[2]` for the opener (lines 21-26),
    so it opens, closes at line 3 and returns `echo hello`.
  - base does the same.
  - reference: line 1 is not an opener (§4.5: a backtick info string may not contain a backtick). Line 1 is paragraph
    text, then line 3 opens a fence that is never closed → empty literal → `command:` (matches the §8 remark).
- **indent-four**:
  - base: `^ {0,3}` refuses 4 spaces, finds no fence and returns `command:` (empty).
  - change: `.trim()` (`change/extract.ts:16`) removes the indent, then opens, closes and returns `echo hello`.
  - reference: an indented code block, `_isFenced` false → `refuse:no-fenced-block`.
- **Item 1 (does base differ from change exactly where §2 puts the disagreement?):** yes.
  - They differ on the eight shipped cases, as §2 requires: base never refuses, change refuses each class.
  - They differ on indent-four. Base keeps the 0–3 indent rule that change loses by stripping first, which is §4's
    named causal property.
  - They agree on well-formed and info-string-backtick (both lenient on the info string).
  - The disagreement that carries the case is change vs reference on the two openers, and it appears on exactly those
    two lines.
- **Contract conformance (§2), item by item:**
  - one operation per version: met
  - base lenient and never refusing: met (`base/extract.ts` has no refuse branch)
  - change has the three named classes, strips each line, ignores the info string: met
  - the suite has 8 cases, each varying only the closer or leaving it out: met (`suite.ts:7-13`, only `closer` varies)
  - the reference is a pinned CommonMark parser (commonmark 0.31.2): met
  - the check runs the suite first, then prints one line per input: met (`check/run.ts:45-60`)
  - all required inputs are present: met (`run.ts:11-24` plus the 8 shipped cases)

### Unplug
Mutations, traced by hand (not executed):
- **Remove `.trim()` (`change/extract.ts:16`).** indent-four change → `command:` (no opener matches at column 0). The
  line then differs from what §8 records, so the check output would show the fix.
- **Make change refuse a backtick in a backtick fence's info string.** info-string-backtick change →
  `refuse:unclosed-fence`: line 1 is no longer an opener, line 3 opens and nothing closes it. Again visible.
- **Unplug the reference's discriminator (`node._isFenced` reads undefined).** Every line's reference would read
  `refuse:no-fenced-block`, including well-formed. The well-formed line (`reference=command:echo hello`) is the control
  that would go red. indent-four (`refuse`) controls the opposite direction. So the private field is guarded in both
  directions by the case's own output.
- **Break the shipped suite.** `run.ts:45-51` exits 1 before any stdout, and `set -euo pipefail` in `check.sh`
  propagates it.
- **What does not go red:** `check.sh` exits 0 whatever the extractors return. It prints and never asserts. A future
  edit that "fixed" change would still exit 0, and only a reader comparing lines with §8 would notice. §2 asks for
  printing only, so this is not a defect of this diff (see Remainder R-3).

### Closure
The order's six properties:
1. **The fixture reproduces §2's condition.** By source trace, yes (Differential). Execution is UNRUN, so this rests on
   a trace, not a measurement. Condition C-2.
2. **The expected outcome follows from the evidence.**
   - `expected/EXPECTED.md` asks for:
     - change=command on both openers
     - the reference not reading the command as fenced
     - the well-formed line agreeing on all three
     - the suite green
   - The §8 output carries all four (`CASE.md:104,118-120`).
   - **False-world answer (law #52):** if change refused malformed openers, the two lines would read `change=refuse:*`
     (or, for info-string-backtick, `refuse:unclosed-fence`), not `command:echo hello`.
   - If the extractor were just broken, the well-formed line would disagree.
   - So the output discriminates, and the opposite outcome would print different lines.
   - One reach limit: the token `command:` (empty) means "no fence found" in the base column (indent-four) but "an
     empty fenced block" in the reference column (info-string-backtick). The §8 remark explains only the second. See
     Remainder R-2.
3. **The case still proves something; no check was bent.**
   - `run.ts` computes every column from the code; nothing is hard-coded to an expected value.
   - The suite's expectations are the change's own policy, as §2 orders.
   - Not bent.
4. **Nothing private.**
   - I scanned the diff for host and seat names, `███████`, `██████`, private repo or org names, e-mail addresses and
     token prefixes. One hit: `CASE.md:79`, "The `date` program is not in this seat's command class". That is fleet
     jargon, not a secret (Remainder R-4).
   - The lockfile resolves only to `registry.npmjs.org`.
   - The only non-ASCII bytes in the added lines are `·` at `CASE.md:5`, which matches the house style.
5. **The diff stays in its lane.**
   - All paths are under `tests/corpus/cases/PV-004/` outside `expected/`, plus the PV-004 object in `manifest.json`.
   - **But the manifest edit changes the SHAPE of that object** (Condition C-1).
6. **The check runs on Linux and macOS.**
   - `check.sh` uses only `#!/usr/bin/env bash`, `set -euo pipefail` and `CDPATH= cd -- "$(dirname -- "$0")" && pwd`,
     which is portable bash on both.
   - The lockfile pins all 20 `@typescript/typescript-<os>-<cpu>` 7.0.2 optional binaries, including linux-x64,
     linux-arm64, darwin-x64 and darwin-arm64.
   - Non-portable or environment-bound calls:
     - (a) `npm ci` needs the npm registry (network) unless the cache is warm. That is in tension with §2's
       "with no network" (R-1).
     - (b) `CASE.md:79` documents `ls -l --time-style=full-iso`, a GNU `ls` option that BSD/macOS `ls` refuses. It is
       prose, not a step a judge runs.
     - (c) `check.sh` is mode 100644 and must be run as `bash check.sh`, which §8 documents.
     - (d) §8 does not say which OS the recorded run was made on.
- **R4 dependency delta:**
  - commonmark 0.31.2 (BSD-2-Clause), its deps entities 3.0.1, mdurl 1.0.1 and minimist 1.2.8, typescript 7.0.2
    (dev) and its 20 optional platform binaries.
  - The lock carries no `hasInstallScript`, and `check.sh` passes `--ignore-scripts`.
  - Licences are out of scope per the order.

### Remainder
Owed findings. None of them blocks:
- **R-1 · `check.sh:5`.** `npm ci --ignore-scripts` fetches from the registry. §2 says the case provides its parts "with
  no network". Remedy: either state in §8 that installation needs the registry once, and that only the check proper is
  offline, or vendor or cache the packages. The master rules.
- **R-2 · `check/run.ts:33-38`.** The empty `command:` is one token with two meanings: base "no fence" vs reference
  "empty fenced block". Remedy: add one `remark:` line in §8 saying that base's `command:` on indent-four means base
  found no fence.
- **R-3 · `check.sh`.** It exits 0 whatever the extractors print. That matches §2, so this is not a defect, but the
  check cannot fail on the property the case exists to show. Remedy, if the master wants one: a committed copy of the
  §8 stdout and a `diff` against it.
- **R-4 · `CASE.md:79`.** "The `date` program is not in this seat's command class" leaks orchestration jargon into a
  public file. Remedy: drop the sentence, or write "the clock is the mtime of …". Also record the OS of the run (Closure
  6(d)).
- **R-5 · ANSWER-KEY TENSION. ORDER-MANDATED (R12); outside this diff's lane.**
  - The case's own reference prints a command for ALL EIGHT shipped inputs.
  - By CommonMark §4.5, none of them is malformed:
    - a longer closer closes the block (`closer-too-long`, `closer-longer-still`)
    - a shorter, differently charactered or trailing-text line is content
    - an unclosed fence runs to the end of the document
  - So the §8 output shows change refusing eight CommonMark-valid documents while it is described as "strict, by the
    CommonMark rules" (§1).
  - `expected/EXPECTED.md` says what "must NOT be claimed" is "that the shipped refusal classes do not work". A
    reviewer under evaluation who reports "change refuses a CommonMark-valid longer closer" makes a true claim that the
    check supports. The key neither credits nor excludes it.
  - Remedy, for the master or owner (not this diff): rule in `expected/` whether that observation is a permitted
    additional finding.
- **Declined to judge (R11):** licences (the order puts them out of scope); `types/ambient.d.ts` hand-declaring
  `process` and the `commonmark` types instead of using `@types/*` (a structural choice, not a defect).

### Profile
None of the craft profiles applies. This is a test-fixture and documentation diff: no authentication, payments, data
access, configuration or destructive path. I reviewed it against the order's six properties, with R1–R4 and R10–R13.

### Focus
- The order's six properties.
- The paired paths (R13): base vs change, and extractor vs reference.
- The manifest's documented reader, `tests/README.md:66-71`, because the diff changes the shape that reader indexes.

### Conditions
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

### Unrun
- `bash tests/corpus/cases/PV-004/check.sh`. Refused verbatim: `the program "bash" is NOT in the ██████████████ list of
  role review under permission class ██████████████████████████, which is [saphan doc check go git ls cat grep find wc
  head tail diff pwd cut sort] — deny by default. … allowed form: none`. Not rebuilt from other commands (R15, law #57).
- The Linux/macOS portability run of the same script. Same refusal class.
- The README `jq` count command over the changed manifest (`jq` is not in the class). The C-1 failure is inferred from
  jq semantics, not observed.
- Reading `commonmark` 0.31.2's source to confirm `_isFenced` and its CJS entry point. `node_modules/` is absent and
  installing it needs `npm`. The output-level control is described under Unplug.

## The review status as filed

## REVIEW — corpus-pv-004
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Traced by hand, PV-004's fixture reproduces §2's condition: change accepts both malformed openers, the reference reads neither as a fence, and the well-formed line and the green suite are the controls. The lane holds and nothing private is in it.
  C-1: manifest.json:76 turns PV-004's four-field status object into the string "built", which breaks the count command published at tests/README.md:66-71. Restore the object.
  C-2: check.sh was NOT run (bash, node and npm are denied on this seat). A permitted role must run it on Linux and macOS and match CASE.md:110-120.
  Five owed remarks (R-1..R-5), none blocking; R-5 is an answer-key tension in expected/ for the master to rule on.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| policy binding name | 3 |
