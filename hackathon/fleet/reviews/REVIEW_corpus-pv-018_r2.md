# REVIEW — corpus-pv-018 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 4 OWED · 2 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:31:09Z

## Summary

PV-018 at 10dfbe9 reproduces §2 by trace: the planted and plain lines separate change from mutant in both directions, and the four shipped tests pass on both (4/4 on change measured with go test). EXPECTED follows from the recorded output. No check was bent, the lane holds, and nothing private beyond one internal scratch path.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-018

### Ground
- reviewer: claude-code (Anthropic family), headless, review role, permission class `██████████████████████████`.
- `git rev-parse --short=7 HEAD` printed `10dfbe9` (ROLE-LOCK satisfied). Merge-base of the scope: `fe146e0`. Scope `fe146e0...10dfbe9`, one commit: `[corpus-pv-018] tests: build the PV-018 case code and record its run`.
- write-set read in full (`git diff fe146e0...10dfbe9`): `tests/corpus/cases/PV-018/{CASE.md, base/validator.go, change/validator.go, change/validator_test.go, check.sh, check/main.go, check/mutant/validator.go, data/data.go, go.mod}` and `tests/corpus/manifest.json` (2 lines, 324–325, PV-018's `status.fixture` and `status.checks`). Also read: `expected/EXPECTED.md` (after judging item 1), `README.md:9,47,135` (definition of the build window), `.gitignore`.
- master's claim (macm5, 2026-10-22), re-measured: "touches only its own case directory and manifest line, never `expected/`" — CONFIRMED by `git diff --stat` (10 files, none under `expected/`, manifest hunk is PV-018's own entry). "`check.sh` exit 0" — UNRUN here (see Unrun); not re-measured.
- the header/manifest flip from "not built yet — built before the window, declared prior art" to "built 2026-10-22 (declared: built inside the build window)" is consistent with `README.md:47` (contracts are prior art; case code is written inside the window starting 22 October 00:00 UTC). Not a provenance defect.
- toolchain on this seat: `go version go1.26.8 darwin/arm64` — the same version §8 records.

### Differential
- `go -C tests/corpus/cases/PV-018 vet ./...` → no output, exit 0.
- `go -C tests/corpus/cases/PV-018 test -count=1 -v ./...` →
  ```
  ?   	pv018/base	[no test files]
  --- PASS: TestShippedDataPasses (0.00s)
  --- PASS: TestCitingRowOverNarrowBudgetIsRefused (0.00s)
  --- PASS: TestCitingRowAtNarrowBudgetPasses (0.00s)
  --- PASS: TestPlainRowKeepsWideBudget (0.00s)
  ok  	pv018/change	0.441s
  ?   	pv018/check	[no test files]
  ?   	pv018/check/mutant	[no test files]
  ?   	pv018/data	[no test files]
  ```
  4 tests run, 4 pass on change. Base has no tests (it is all new code; no pre-existing reds possible).
- `check.sh` itself was NOT run (bash refused by the seat; see Unrun). In its place, a hand trace of `check/main.go` against `data/data.go` (citing prefix `"Takes {{measure:latency}} on a warm start."` = 42 bytes, so no `Row` call panics):
  - shipped: citing rows are install#1=100, limits#0=120, limits#2=115, faq#0=70 (all ≤120); largest plain row intro#2=190 (≤200). Mutant narrows install/limits/faq: largest row there is 120. ⇒ base=pass change=pass mutant=pass reference=pass.
  - planted: `changelog` + citing row 160 (120<160≤200). base wide → pass; change cites → `changelog#0:160>120`; mutant: page not listed → pass; reference → `changelog#0:160>120`.
  - plain: `install` gains plain row 160 at index 2. base pass; change wide → pass; mutant listed → `install#2:160>120`; reference pass.
  - suites: the four tests put citing rows only on listed pages (install, limits) and plain rows only on an unlisted page (glossary), so the mutant gives identical answers to every test ⇒ change=pass, mutant=pass.
  The trace agrees with every line §8 records. It is a reading, not a run.

### Unplug
- Not executed (a mutation run means writing a module copy and running `go test` on it — the same effect as `check.sh`'s mutant step, which this seat may not run; law #57). Traced instead, from `change/validator_test.go`:
  - remove the `if data.Cites(r)` branch at `change/validator.go:15–17` (change collapses to base): `TestCitingRowOverNarrowBudgetIsRefused` expects one violation at 121>120 and gets none ⇒ RED. So the suite does pin that a narrow budget exists.
  - replace the content rule by the three-name list (the case's mutant): no test reddens ⇒ the suite does NOT pin how the narrow budget is chosen. This is the defect the case exists to show, reproduced as §1/§4 require.
- Unplugging the check's own discrimination: if the planted row were placed on a listed page, or the plain row on an unlisted one, base/change/mutant would agree on that line and the case would show nothing. `check/main.go:47` (`changelog`, not listed) and `:50` (`install`, listed) are what keep both lines discriminating, in both directions.

### Closure
- False-world instrument test (law #52) on the load-bearing line `suite mutant=pass`: `check.sh:23` maps `go test` exit 0 to `pass`, and `go test` exits 0 for a package with no test files — measured in the Differential above (`?  pv018/base [no test files]`, exit 0). ⇒ if `change/validator_test.go` were not copied into the scratch module, or were excluded by a build tag, the line would still read `suite mutant=pass`. The same sentence prints in the world where the four tests survived the mutant and in the world where no test ran. Today the copy at `check.sh:21` does carry the test file (no build tags in it), so the recorded output is true; the instrument just cannot prove that by itself.
- `check.sh` exits 0 whatever the suites print (the exits after line 13 are not tied to the suite results), so the master's "exit 0" establishes only that the check ran to the end, never the outcome.

### Remainder
- declared by the author (§9): both readings are true; one gap in test power; not reconstructed: the other two parts of the original merge.
- found here:
  - the reference (`check/main.go:17–31`) is the change's rule re-typed (`strings.Contains(r, data.Marker)` vs `data.Cites`, same constants from the same package). Its agreement with change is near-tautological: it shows change does what its own code says, not that the rule is right. §2 defines the reference exactly this way, so this is a limit of the contract, not a defect of the build.
  - `data/data.go:1–2` tells the reviewer under evaluation (who receives `data/` per §5) that a "check" exists. §5 withholds the check; the comment names it. Low impact.

### Profile
- no named profile in the order; the diff is test-corpus code with no auth, data access, configuration or destructive path in the product. The one destructive call is `check.sh:20` `rm -rf` on a derived path (S4-style bound read): path is `<case>/../../../../.saphan/work/corpus-pv-018/mutant-suite`, fixed suffix, so the blast radius is bounded to that leaf; it does resolve outside the case directory.

### Focus
1. **Fixture reproduces §2:** YES, by trace (the run itself is UNRUN). base, change, the four shipped tests, the mutant, the reference and all three inputs §2 names are present with the required sizes and placements. The planted line separates change from mutant (change fails at narrow, mutant passes); the plain line separates them in the other direction; the shipped line separates nothing, as §4 requires. Base differs from change only on the planted line.
2. **Expected follows from the evidence:** YES. `expected/EXPECTED.md` asks for suite change=pass + mutant=pass and the planted line change=fail(narrow)/mutant=pass/reference=fail — all present in §8 and in the trace. False-world answer: if the suite pinned the rule, the mutant suite would read `fail`; that is a different line, so the evidence discriminates — with the one hole named in Closure (no-tests-ran also prints `pass`).
3. **Not bent:** no check was bent. The weakness of the four tests (citing rows only on listed pages) is the reproduced defect itself, as §1 states. Reference and mutant are written exactly as §2 defines them.
4. **Nothing private:** no hosts, seats, tokens, e-mail addresses or private repo names in the diff (`grep` over the case directory outside `expected/`). One internal-convention path: `check.sh:9` `███████████████████████████` — the fleet's work-directory convention and the stream slug, in a public script (OWED below). Pre-existing lines 45–46 of CASE.md name model families; not in this diff.
5. **Lane:** YES, confirmed by `--stat` (Ground).
6. **Portability:** the shell is portable (`#!/usr/bin/env bash`, `cd`/`dirname`/`pwd`, `rm -rf`, `mkdir -p`, `cp -R` — same on GNU and BSD; no `readlink -f`, `sed -i`, `mktemp` flags or GNU-only options). Non-portable or environment-sensitive: (a) `go.mod:3` `go 1.26` — on Go < 1.26 the default `GOTOOLCHAIN=auto` downloads a toolchain (network, against §2 "no network"), and `GOTOOLCHAIN=local` refuses; the code needs nothing newer than early Go 1.x; (b) `check.sh:8–9` derives a repo root four levels up and writes there — wrong place if the case directory is copied out on its own (as an evaluation that withholds `expected/` may do); (c) the script file is mode 100644, so it must be invoked as `bash check.sh`, which §8 does.

### Conditions
- OWED (gate-before) · `tests/corpus/cases/PV-018/check.sh` · the check was not run by this review (bash refused on the review seat). Closes when a role holding `bash` runs `bash tests/corpus/cases/PV-018/check.sh` on 10dfbe9 and the five printed lines match CASE.md §8 byte for byte.
- OWED · `check.sh:16,23` · `suite …=pass` also prints when no test ran (`go test` exits 0 on "no test files"). Closes when the check asserts the number of tests run (e.g. count `--- PASS`/`--- FAIL` from `go test -v`, or `go test -json`) and prints it on the suite lines, e.g. `suite mutant=pass(4/4)`.
- OWED · `check.sh:9,20–22` · the scratch module is written into the repository root under `███████████████████████████`, which `.gitignore` does not cover and which is never removed: a judge's checkout is left dirty, the path leaks the fleet's internal convention and stream slug, and it lands outside the tree if the case is copied out. Closes with `scratch="$(mktemp -d)"` plus `trap 'rm -rf "$scratch"' EXIT`.
- OWED · `go.mod:3` · `go 1.26` forces a toolchain download (network) on older Go. Closes by lowering the directive to the oldest version the code needs, or by stating the floor and `GOTOOLCHAIN=local` behaviour in §8.
- NOTE · `data/data.go:1–2` · the package comment names "the check", which §5 withholds from the reviewer under evaluation. Reword to "what the validators share".
- NOTE · `check/main.go:17–31` · the reference repeats the change's rule (§2-mandated); its agreement is not independent evidence.

### Unrun
- `bash tests/corpus/cases/PV-018/check.sh` — refused at the act: `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████, which is [saphan doc check go git ls cat grep find wc head tail diff pwd cut sort] — deny by default … allowed form: none — "bash" is denied under permission class ██████████████████████████; do not rebuild the act from other commands, hand it to the role that holds it`. Not rebuilt from `go run ./check` + a hand-made mutant copy (law #57, craft R15). Its five lines were verified by trace only.
- executed mutation of the shipped tests (R3) — same reason; traced in Unplug.
- Linux run — no Linux seat in this leg; portability judged by reading.
- (ran, for the record: `saphan doc check _REVIEW_STATUS.md --family review-status` → `family=review-status ok (checked=true)`.)

## The review status as filed

## REVIEW — corpus-pv-018
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-018 at 10dfbe9 reproduces §2 by trace: the planted and plain lines separate change from mutant in both directions, and the four shipped tests pass on both (4/4 on change measured with go test). EXPECTED follows from the recorded output. No check was bent, the lane holds, and nothing private beyond one internal scratch path.
  check.sh is UNRUN here: bash is refused on the review seat. Owed before the gate: a bash-holding role re-runs it and matches §8.
  Also owed: the suite lines can't tell "no tests ran" from pass; the scratch dir sits in the repo root, isn't ignored and is never removed; `go 1.26` pulls a toolchain over the network on older Go.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T01:29:00+02:00

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| policy binding name | 4 |
