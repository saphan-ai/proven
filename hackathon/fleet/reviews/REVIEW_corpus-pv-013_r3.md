# REVIEW — corpus-pv-013 · round 3

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 4 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:10:18Z

## Summary

PV-013 at 4c1d2cb: base and change differ only at validation; the code read and go vet match the CASE.md §8 lines and EXPECTED.md; lane holds, expected/ untouched; round-2 conditions addressed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-013

### Ground
`git rev-parse --short=7 HEAD` printed `4c1d2cb` (role lock met). Scope `9303b2d...4c1d2cb`, two commits:
d4ff8b7 (round 1, author date 2026-10-21T16:53:42Z) and 4c1d2cb (round 2, author date 2026-10-23T02:21:42Z).
Write-set (`git diff --name-status 9303b2d...4c1d2cb`): M `test/corpus/cases/PV-013/CASE.md`, A `base/go.mod`,
A `base/main.go`, A `change/go.mod`, A `change/main.go`, A `check.sh`, M `test/corpus/manifest.json`.
No path under `expected/` is in the write-set. The manifest hunk is lines 236-243 only, inside the PV-013 entry
(`"id": "PV-013"` at line 228). Master's claim "only its own case directory and manifest line, never `expected/`":
re-measured, holds. Master's claim "`check.sh` exit 0": NOT re-measured (see Unrun).
Note on the brief: the order asks for the status H1 `# REVIEW — corpus-pv-013-review-r2-claude`; the engine-injected
channel grammar names `# REVIEW — corpus-pv-013` and says it wins. I used the engine's form.
Round-1 conditions were NOT carried into this order (law #44 (3)); I judged them from round 2's commit message and
diff only (see Closure).

### Differential
I could not run the check (`bash` refused by the hook, quoted in Unrun), so there is no executed base/tip red-green
from this leg. What I ran:
- `go -C test/corpus/cases/PV-013/base vet .` — no output, exit 0.
- `go -C test/corpus/cases/PV-013/change vet .` — no output, exit 0.
- `diff base/main.go change/main.go` — the change adds exactly: imports `net/url`, `regexp`; `syntheticToken`;
  the regexes `envName`, `channelName`; the `control` verb and its usage line; `credentialsPath`; `judge` and
  `validEndpoint`; the `judge` call at the head of `setField` (change/main.go:150-152); `runControl`. Nothing else
  differs except two header comments. So base differs from change exactly where §2 says: base stores any value
  (base/main.go:100-107); change refuses three fields and quotes the value in full (change/main.go:120-136).
Reading the code against CASE.md §8's recorded output: every recorded line is what this code prints.
`token-in-name-then-show` on change = the refusal from change/main.go:124 (exit 0, stdout) plus `show` = `(empty)`,
because `judge` returns before `store`. The `control` line = `pass`, because runControl (change/main.go:184-210)
sets only legal values and never reads the credentials file into any output. This is a reading, not a run.

### Unplug
Not executed (bash refused; editing code under review is outside this role). Reasoned false-world answer (law #52),
from check.sh:181-187 and change/main.go:124:
- If change did not quote the value (the later repair), `c-paste.txt` would hold the refusal without the token:
  `change_token_in_output: no`, `assert_change_refusal_leaks` false, `problems=1`, exit 1. Distinguishable from the
  defect world. The exit code and the flag both turn.
- If change did not refuse at all, the refusal grep fails, exit 1. Distinguishable.
- If only base leaked, round 2 split the flag per side (check.sh:50-56) and asserts on the change file alone; the
  `--negative-control` mode feeds a base-side leak and exits 1. Distinguishable.
- One world it does NOT separate: change stores the pasted token AND prints some refusal that omits it, then `show`
  prints the token. `assert_change_refusal_leaks` (check.sh:60-63) greps `refused: token_env` and the token anywhere
  in the combined set+show file, so it would still pass. The flag column alone cannot tell "leaks through the
  refusal" from "leaks through show". Only the printed text separates them. That world is not the fixture's today
  (`judge` runs before `store`), so the recorded result stands; the instrument is narrower than EXPECTED.md's claim.
  OWED below.
The two negative-control modes (check.sh:65-86) feed hand-written text files, not the binaries. They prove the
predicate can return false (and lines 69-76 re-prove it on every run, raising `problems`). They do not unplug the
product.

### Closure
Round-1 conditions, as round 2's commit message names them (the conditions' own text was not given to me):
- "a base-only leak could satisfy its flag" — ADDRESSED. check.sh:50-56 reports `base_token_in_output` and
  `change_token_in_output` separately; the pass/fail decision (check.sh:184-187) reads the change file only.
- "assert the change refusal and leaked synthetic value independently" — ADDRESSED. check.sh:60-63 needs both
  greps; lines 69-76 prove each half can fail.
- "prove both negative controls" — ADDRESSED as predicate controls (see Unplug for their reach).
- "restore the manifest status object" — ADDRESSED. manifest.json:237-242 is an object with the four keys every
  entry carries (`"independent_check"` appears 21 times in the file).
- Also in round 2, not named in the message: a non-`pass` shipped control and each missing refusal now set
  `problems` (check.sh:93-100, 160-179). Before, the check could exit 0 with a failed control.
I cannot say whether a round-1 condition exists that the commit message does not name. That is in Unrun.
Item 3 (was a check bent?): no. Exit 0 needs the defect to be present (check.sh:184-187). No assertion was loosened
between rounds; round 2 only added assertions. No line was made to come out "as expected" by special-casing.

### Remainder
- Stale command paths. The ADR-0030 move renamed `tests/corpus` to `test/corpus`, and `git ls-files tests/corpus`
  is empty, but the command text still says `tests/`: check.sh:3, CASE.md:88, CASE.md:108, CASE.md:118. A judge who
  pastes §8's command gets "No such file or directory". The move was carried as renames only, so the text was not
  touched.
- check.sh:13-15 makes `███████████████████████████` at the root of the reader's checkout and leaves that empty
  directory behind (the trap removes only `$work`). Round 1 used `mktemp -d`. The path also depends on the case
  sitting exactly four levels down. It is not in `.gitignore`; git does not list empty directories, so
  `git status` stays clean after a run.
- check.sh:15 does not check `mktemp`. If it fails, `work` is empty and the builds write `/basebin` and
  `/changebin`. The build then fails loudly, so this is no false green.
- Portability (item 6), by reading: `#!/usr/bin/env bash`, `local`, `read -r … || [ -n ]`, `${1-}`,
  `mktemp -d TEMPLATE`, `grep -q -F` and `grep -q -F -x` all work on GNU and BSD/macOS userlands, including
  macOS's bash 3.2. No bash-4-only form was found. The one hard requirement is a local Go >= 1.26
  (`GOTOOLCHAIN=local` and go.mod `go 1.26`), and an older Go fails the build loudly. CASE.md §8 records linux/amd64
  only; macOS has not been run by anyone in the record.
- Manifest wording: round 1 had `"fixture": "not built (prior art, built before the window)"`; the tip says
  `"built 2026-10-21"`. CASE.md:5 still says "declared prior art", but the manifest line no longer does. AGENTS.md §3
  says nothing written before 22 October may be copied in. d4ff8b7's author date is 2026-10-21. Declined to judge;
  this is the owner's call.
- Commit messages carry five `(cherry picked from commit …)` hashes. `git cat-file -t 92bcc06…` fails here, so they
  name commits outside this public repository. AGENTS.md §2 keeps only this repository's hashes. These are commit
  metadata, not file content.

### Profile
None declared in the order. The diff is a test fixture. It touches no auth, payment, data access, config or
destructive path of proven itself, so no security profile is triggered (R9). The secret is synthetic and follows no
vendor format.

### Focus
1. The fixture reproduces §2's condition: yes, by reading. Base and change differ exactly at validation; every §2
   input is present (valid name check.sh:149-151; credentials+show 153-155; one bad value per added refusal 157-179;
   pasted token+show 181-187); the shipped control runs (89-100). Output not pasted by me: UNRUN.
2. The expected outcome follows from the evidence: EXPECTED.md needs "change = refused and token in output = yes"
   plus the control passing. §8's `token-in-name-then-show` and `control` lines give exactly that, and the text shows
   the token inside the refusal message with show `(empty)`. False-world: a repaired change prints
   `change_token_in_output: no` and exits 1. That is distinguishable, with the one gap named in Unplug.
3. Not bent (Closure).
4. Nothing private in the case files: a grep of the case directory for internal paths, host or seat names,
   e-mail addresses and real tokens found only the synthetic token. Out-of-repo cherry-pick hashes are in commit
   messages (Remainder).
5. In its lane: yes (Ground).
6. Portable by reading; macOS unrun (Remainder).

### Conditions
- OWED: check.sh:3, CASE.md:88, CASE.md:108, CASE.md:118: `tests/corpus/...` → `test/corpus/...`. Closed when
  every command in §8 runs as pasted from the repository root.
- OWED: check.sh:60-63, the pasted-value assertion matches the token anywhere in set+show output. Closed by asserting
  on the refusal line itself (for example `grep -q -F "rejected value: $token"` on the set step's output alone), so
  the instrument proves what EXPECTED.md claims: the refusal message holds the token.
- OWED: CASE.md:105-124, label the two `--negative-control` modes as controls on the assertion over fixed text, not
  as runs of base/change. The printed `change: refused: token_env rejected` is not change's output.
- NOTE: check.sh:13-15 should go back to `mktemp -d` (or remove `$work_root` on exit) and fail when `mktemp` fails.
- GATE PRECONDITION (master, before gate): run `bash test/corpus/cases/PV-013/check.sh`,
  `… --negative-control` and `… --negative-control=no-refusal` on 4c1d2cb. Paste the output and exit codes, and
  confirm they match CASE.md §8. This leg could not run them, and item 1 asks for them.
- DECLINED TO JUDGE (owner): the case code's 2026-10-21 author date against AGENTS.md §3; the manifest's dropped
  "prior art" wording; out-of-repo cherry-pick hashes in the commit messages.
Verdict reading under law #53: Unrun is not `none`, so this review does not return `accept`. It returns
`accept-with-conditions`. The verdict rests on the code read and `go vet`, and needs the gate precondition above.

### Unrun
- `bash test/corpus/cases/PV-013/check.sh` (and both `--negative-control` modes). Hook refusal, verbatim:
  `the program "bash" is NOT in the ██████████████ list of role review under permission class
  ██████████████████████████ … allowed form: none — "bash" is denied …; do not rebuild the act from other commands`.
  I did not rebuild it from `go run` (law #57, R15).
- Product-level unplug (mutating change/main.go:124 and re-running): needs the code edit and the same bash run.
- macOS run: no macOS seat on this leg.
- Round-1 review conditions in their own words: not carried into this order. I could only check them against
  round 2's commit message.

## The review status as filed

## REVIEW — corpus-pv-013
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-013 at 4c1d2cb: base and change differ only at validation; the code read and go vet match the CASE.md §8 lines and EXPECTED.md; lane holds, expected/ untouched; round-2 conditions addressed.
  OWED: stale tests/corpus paths (check.sh:3, CASE.md:88/108/118); pin the leak assertion to the refusal line; label the predicate-only negative controls.
  UNRUN: check.sh itself (bash refused by the hook). The master runs it and pastes the output before the gate.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T16:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| policy binding name | 2 |
