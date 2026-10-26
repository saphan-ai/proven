# REVIEW — corpus-pv-009 · round 4

- verdict: reject
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 2 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T18:25:21Z

## Summary

REJECT at 627907d: PV-009's check does not compile, because check.go:9-10 and check.sh:4 still name tests/corpus after the move to test/corpus. go build ./... and go vet ./... exit 1, and only this package fails.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-009

> **RED — the reviewed tip does not build.** At `627907d` the module's `go build ./...` exits 1. The one failing
> package is `test/corpus/cases/PV-009`, and this diff adds it. `check.go` imports
> `github.com/saphan-ai/proven/tests/corpus/cases/PV-009/{base,change}`, and `check.sh` runs
> `go run ./tests/corpus/cases/PV-009/check.go`. After the ADR-0030 move, neither path exists. So the case's check
> cannot run, and every `go build ./...`, `go vet ./...` and `go test ./...` of proven fails on this branch. The
> master's claim (macm5, 2026-10-22), "`check.sh` exit 0", does not hold at this commit.

### Ground

- `git rev-parse --short=7 HEAD` printed `627907d`. This equals the order's `reviewed-sha` and its `base`.
- Scope: `git log --oneline 9303b2d..627907d` lists two commits: `771342c` (round 1, build) and `627907d` (round 2,
  "record review condition evidence").
- `git diff --stat 9303b2d...627907d` shows 7 files, +166/−6. All of them are under `test/corpus/cases/PV-009/`, plus
  `test/corpus/manifest.json`.
- Toolchain found (C1): `go.mod` names `module github.com/saphan-ai/proven` and `go 1.26`. The case uses only the
  standard library.
- What I ran, with outcomes:
  - `go vet ./test/corpus/cases/PV-009/...` → exit 1:
    `check.go:9:2: no required module provides package github.com/saphan-ai/proven/tests/corpus/cases/PV-009/base`
    (and the same error for `/change` at `check.go:10:2`).
  - `go build ./...` → exit 1, with the same two errors.
  - `go list -e -f '{{.ImportPath}} {{if .Error}}ERR{{end}}{{if .DepsErrors}}DEPERR{{end}}' ./...` → 42 packages. Only
    `test/corpus/cases/PV-009` is flagged (`DEPERR`). So the module build fails only because of this diff.
  - `go test ./test/corpus/cases/PV-009/change/ ./test/corpus/cases/PV-009/base/` → `ok …/PV-009/change`;
    `base` has no test files.
  - `ls tests test/corpus` → `tests/` holds only `fixtures`; `find tests -path '*PV-009*'` → nothing.
  - `grep -rn "tests/corpus" test/corpus/cases/PV-009` → 4 stale references: `check.go:9`, `check.go:10`,
    `check.sh:4` and `CASE.md:96`.
- `bash test/corpus/cases/PV-009/check.sh`: **refused by the leg hook**, quoted verbatim:
  `the program "bash" is NOT in the ██████████████ list of role review … allowed form: none`. The check is UNRUN by me
  (R15), and I did not rebuild it as a bare `go run` (law #57). I therefore have no output of my own to paste. The
  `go vet` and `go build` failures above show statically that the script's `go run` target cannot compile:
  `check.go` is the file it runs, and it does not resolve its imports.
- Author's self-assessment set aside (R1): the CASE.md §8 line "Exit code: 0" and the master's check were not taken as
  evidence.

### Differential

Property 1: the fixture reproduces the condition that §2 names. I judged this by reading the code, because the check
cannot run.
- `base/guard.go:18` uses `arrow.FindStringSubmatch` (first pair only). `change/guard.go:18` uses
  `FindAllStringSubmatch(after, -1)` (every pair). This is the one difference §2 names.
- Both read the closing number the same way: `strconv.ParseInt(pair[1], 10, 64)`, then `continue` on `err != nil ||
  closing == pinned` (`base/guard.go:22-24`, `change/guard.go:19-21`). This is §2's shared conversion rule, including
  "skip when the conversion fails".
- The reference (`check.go:13-30`) compares the captured digits as text with `strconv.FormatInt(pinned, 10)`. It does
  not use the guard's pattern or conversion. This matches §2.
- The inputs (`check.go:44-54`) are exactly the nine §2 requires, with pinned count 42.
- I traced each input through the code by hand. The traced output equals the nine lines recorded in CASE.md §8. Base
  and change differ on exactly two lines, `stale-after-current` and `three-pairs`, which are the "first pair only"
  shapes. They agree on the other seven, including both overflow lines, where both print `pass` and the reference
  prints `fail`. So base and change disagree exactly where §2 puts the disagreement, and nowhere else. **This holds
  for the code as written. It is not a measurement of the check at this tip, which does not compile.**

Property 2: the expected outcome follows from the evidence. `expected/EXPECTED.md` asks for four kinds of line:
- the twenty-digit and 2^63 lines with change=pass and reference=stale;
- the 2^63−1 line with change=fail;
- the stale-after-current line with base=pass and change=fail.

Every one of these appears in the recorded §8 output and in my trace. False-world answer (law #52): if the guard
refused a number it cannot convert, the `overflow-int64` and `twenty-digits` lines would read `change=fail`. If the
reader repair were absent, `stale-after-current` would read `change=pass`. Both false worlds print different lines, so
the output tells the outcome apart from its opposite.
- remark: the check prints only pass or fail. It does not say why a note was skipped. A pass caused by "regex did not
  match" and a pass caused by "conversion failed" look the same in the output. Only the code reading settles it: the
  pattern `[0-9]+` has no length bound (`change/guard.go:9`), so the skip comes from `ParseInt`.
- remark: the recorded §8 output was produced before the layout move (§8 run time 2026-10-23 09:16:00 UTC, with the
  `tests/corpus` command). It is not reproducible at this tip.

### Unplug

Mutations reasoned, not executed. I may not edit the code under review, and I must leave the worktree clean.
- Make the change refuse a number it cannot convert (`err != nil` → refuse): `overflow-int64` and `twenty-digits` flip
  to `change=fail`. The check detects this mutation.
- Revert the change to `FindStringSubmatch`: `stale-after-current` and `three-pairs` flip to `change=pass`. Detected.
- Remove `break` at `change/guard.go:24`: no input has two non-pinned pairs, so neither the check output nor
  `change/guard_test.go` changes. The `len(got) > 1` guard in the test at line 21 is never exercised. This mutant
  survives (NOTE). It does not touch the property the case exists to show.
- Unplug the check as a whole: it is already unplugged at this tip. Nothing turns red inside the case, and only the
  module build reports it. That is the blocking finding below.

### Closure

- Round-1 conditions: my order says round 2 paid both round-1 reviews' conditions, "carried verbatim into its order".
  That text is **not carried into this order** (law #44(3)), so I cannot check each condition against its own wording.
  From the round-2 diff (`git diff 771342c 627907d`, CASE.md only), round 2 did three things:
  - pointed §6's "newly reconstructed evidence" at §8;
  - re-stamped the run time and wrote "on Linux";
  - added the Go ≥ 1.26 and `GOTOOLCHAIN=auto` network note, and stated that macOS is unverified.

  These look like a portability or disclosure condition and an evidence-pointer condition. I mark them
  inferred-ADDRESSED, not verified.
- Manifest status object: `test/corpus/manifest.json:168-173` (entry `"id": "PV-009"`, line 159) is an object with
  `contract`, `fixture`, `checks` and `independent_check`. The diff changes only `fixture` and `checks`. The status is
  present as an object.

### Remainder

- **BLOCKING B1 — the check does not compile at the reviewed tip.**
  - Where: `test/corpus/cases/PV-009/check.go:9-10` (imports under `…/tests/corpus/…`) and
    `test/corpus/cases/PV-009/check.sh:4` (`go run ./tests/corpus/cases/PV-009/check.go`).
  - Trigger: proven's move from `tests/corpus` to `test/corpus` (ADR-0030) renamed the files but left their contents
    byte-identical. The order states that as a virtue, but here it is the defect: the Go import paths and the script
    path name the directory, so a pure rename breaks them.
  - Evidence: `go vet` and `go build ./...` both exit 1, as quoted in Ground. `go list` flags only this package.
  - Effect: the case cannot reproduce anything as committed. Proven's whole-module build, vet and test are red on
    this branch. The §8 "Exit code: 0" is false for the command a judge would run today.
  - Remedy: the executor changes `tests/corpus` to `test/corpus` in `check.go:9`, `check.go:10`, `check.sh:4` and
    `CASE.md:96`. Then they re-run `bash test/corpus/cases/PV-009/check.sh`, paste the new output and run time into
    §8, and confirm `go build ./...` and `go vet ./...` exit 0.
- **OWED O1 — §8 records a command that does not exist** (`CASE.md:96`). This is part of B1's remedy, listed so it
  is not dropped.
- **NOTE N1 — the `break` mutant survives** (`change/guard.go:24`). No input or test has two stale pairs in one note.
  An optional input such as `LANGUAGES 41 -> 43; 43 -> 44` would pin that the note is named only once.
- **NOTE N2 — the check output cannot tell "no match" from "cannot convert"** (see the remark in Differential). The
  code reading settles it, so this does not block.

### Profile

None of the named profiles applies (R9). This is a public test-corpus fixture: no authentication, data access, config
or destructive path. I judged it against the craft's general rules and the testing-bar skill.

Testing-bar evidence:
- metric: the project declares none for this case. No numeric bar is set, and none is invented.
- bar-weakening moves found: none. No test was skipped, loosened or deleted, and `expected/` was not edited.
- check ranking: `check.go`'s reference is a suite check, written by the case author. Its independence rests on a
  different reading (a text comparison), which I confirmed in code.

### Focus

1. Fixture reproduces the §2 condition: **holds in code, not executable at this tip.** See Differential and B1.
2. The expected outcome follows from the evidence: **yes**, and the false-world lines differ. See Differential.
3. Was a check bent to come out "as expected"? **No sign of it.**
   - The reference uses a different reading (text, not int64). It agrees with the guards on 7 of 9 lines, so it does
     not mark everything stale.
   - The inputs are exactly §2's list.
   - `change/guard_test.go` keeps to §2: no number the conversion refuses. Its largest case, 2^63−1, converts.
4. Nothing private is left: `grep -rniE "macm5|███████|██████████████████████████████████████]+\.(com|org)|token|bridge"
   test/corpus/cases/PV-009` → no output. The commits' author identity is the owner's public git identity, not a seat
   name. CASE.md §3 names model families and `PROVENANCE.md` only.
5. The diff stays in its lane: **yes.** `git diff --name-only 9303b2d...627907d -- test/corpus/cases/PV-009/expected`
   → no output. The manifest hunk is inside the `PV-009` entry only (lines 159-173). The master's lane claim holds.
6. Runs on Linux and macOS:
   - `check.sh` uses `#!/usr/bin/env bash`, `set -euo pipefail`, `cd "$(dirname "$0")/../../../.."` and `go run`.
     All of these are portable to macOS bash 3.2 and BSD `dirname`.
   - The four `..` steps still reach the repository root after the move, because the depth is the same.
   - No non-portable call found. The only blocker on both systems is B1.
   - macOS was not run (no macOS seat). CASE.md §8 already discloses this.

Rulings (C6):
- ruling: I wrote the H1 as `# REVIEW — corpus-pv-009`, not the order's `corpus-pv-009-review-r2-claude` — the
  engine's channel grammar binds the verdict row to the reviewed stream's slug and says to trust it over the brief —
  cost if wrong: the verdict row lands under another slug and the master refiles it.
- ruling: verdict `reject`, not `accept-with-conditions` — B1 is a wrong result recorded as fact (§8 "Exit code: 0")
  and a red module build, not a judgement finding (R6) — cost if wrong: one extra round for a four-line path fix.
- ruling: I did not run `go run ./test/corpus/cases/PV-009/check.go` as a stand-in for the refused `bash` — law #57 and
  R15 — cost if wrong: none for the verdict, because the target does not compile in any spelling.

### Conditions

- C-1 (BLOCKING, B1): fix the four stale `tests/corpus` references (`check.go:9`, `check.go:10`, `check.sh:4`,
  `CASE.md:96`). Re-run `bash test/corpus/cases/PV-009/check.sh` at the new tip and record its output and run time in
  §8. Show `go build ./...` and `go vet ./...` exit 0. The re-review checks only this fix (R14).
- C-2 (NOTE, N1): optional. Add an input or test with two stale pairs in one note, so the `break` is pinned.
- C-3 (NOTE, N2): optional. Add a `remark:` in §8 saying the output does not tell "no match" from "cannot convert",
  and that the unbounded pattern settles it.

### Unrun

- `bash test/corpus/cases/PV-009/check.sh` — refused by the leg hook (`"bash" is NOT in the ██████████████ list …
  allowed form: none`). Not rebuilt in another spelling.
- macOS run — no macOS seat on this leg.
- Mutations of base and change — reasoned only. The order forbids editing the code under review, and the worktree must
  stay clean.
- Round-1 conditions checked one by one against their own text — the text is not carried into this order.
  Inferred from the round-2 diff only.
- Comparing the tree before and after the move byte for byte (the master's "byte-identical" claim) — not measured.
  The stale paths are consistent with the claim.

## The review status as filed

## REVIEW — corpus-pv-009
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: REJECT at 627907d: PV-009's check does not compile, because check.go:9-10 and check.sh:4 still name tests/corpus after the move to test/corpus. go build ./... and go vet ./... exit 1, and only this package fails.
  The case logic is sound in code: base and change differ only on stale-after-current and three-pairs, both overflow lines read change=pass with reference=stale, and expected/ follows from those lines with distinct false-world answers.
  Lane clean, nothing private found. bash check.sh was refused by the hook (UNRUN). Remedy: four path edits, a re-run, and the new output recorded in §8.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T18:40:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| policy binding name | 2 |
