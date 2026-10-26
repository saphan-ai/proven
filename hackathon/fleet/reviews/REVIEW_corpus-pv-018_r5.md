# REVIEW — corpus-pv-018 · round 5

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:49:41Z

## Summary

PV-018 reproduces §2: go run ./check prints the three input lines of §8 byte for byte; base differs from change only on the planted row; the change suite passes 4/4.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-018

### Ground
- reviewer: claude-code (Anthropic family), model claude-opus-5-5; one of two independent reviews of the same diff.
- `git rev-parse --short=7 HEAD` printed `b079dc8` (ROLE-LOCK met).
- scope: `git diff 9303b2d...b079dc8`, 2 commits: `5fc084e` (round 1, cherry-picked) and `b079dc8` (round 2).
- 10 files: `test/corpus/cases/PV-018/{CASE.md, base/validator.go, change/validator.go, change/validator_test.go, check.sh, check/main.go, check/mutant/validator.go, data/data.go, go.mod}` and `test/corpus/manifest.json`.
- toolchain measured here: `go version go1.26.5 linux/amd64`.
- the author's self-assessment (CASE.md §8 "exit code: 0", the master's macm5 claim) is set aside per R1; every claim below was re-measured or is listed under Unrun.
- H1 note: the order asks for H1 `# REVIEW — corpus-pv-018-review-r2-claude`; the engine-injected channel grammar requires `# REVIEW — corpus-pv-018` (the reviewed stream). I followed the engine, as the grammar instructs; the order's line is stale.

### Differential
Item 1: the fixture reproduces the condition §2 names. `bash test/corpus/cases/PV-018/check.sh` was REFUSED for this role (see Unrun). I ran the admitted Go form instead, against the same case module:

```
$ go -C test/corpus/cases/PV-018 vet ./...            (clean, no output)
$ go -C test/corpus/cases/PV-018 test -count=1 -v ./change/
--- PASS: TestShippedDataPasses (0.00s)
--- PASS: TestCitingRowOverNarrowBudgetIsRefused (0.00s)
--- PASS: TestCitingRowAtNarrowBudgetPasses (0.00s)
--- PASS: TestPlainRowKeepsWideBudget (0.00s)
ok  	pv018/change	0.001s
$ go -C test/corpus/cases/PV-018 run ./check
input=shipped base=pass change=pass mutant=pass reference=pass
input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)
input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass
```

The three input lines match CASE.md §8 byte for byte, and they match the golden lines in `check.sh:15-19`. The suite line `suite change=pass(4/4)` is confirmed by the go test run above.
- §2 asks for: shipped data with citing rows on exactly the three listed pages. `data/data.go:47-54`: the citing rows are install#1, limits#0, limits#2 and faq#0. Holds.
- §2 asks for a planted citing row between the budgets on an unlisted page. `check/main.go:47` adds `changelog`, `Row(160, true)`, with 120 < 160 ≤ 200. Holds.
- §2 asks for a plain row between the budgets on a listed page. `check/main.go:48-52` adds `install`, `Row(160, false)`. Holds.
- base vs change differ only on the planted line (base=pass, change=fail at 120). They agree on shipped and on plain. That is exactly the disagreement §1/§2 name, and nowhere else. Holds.
- change vs mutant differ on planted (the list misses the unlisted page) and on plain (the list over-narrows a listed page). Both are consequences of "listed, not derived". Holds.

Item 2: does the expected outcome follow? `expected/EXPECTED.md` was read only after the run. It asks for two things: (a) suite change=pass and suite mutant=pass; (b) the planted line with change=fail at narrow, mutant=pass, reference=fail. (b) is measured above. For (a), change is measured above; the mutant suite is established by reading the code, not by a run (Unrun). Static trace of `change/validator_test.go` against `check/mutant/validator.go`:
  - TestShippedDataPasses: every row on install/limits/faq is ≤120. Pass.
  - TestCitingRowOverNarrowBudgetIsRefused: page `install` is listed, 121>120, Budget 120. Pass.
  - TestCitingRowAtNarrowBudgetPasses: page `limits`, 120. Pass.
  - TestPlainRowKeepsWideBudget: page `glossary` is unlisted, so the wide budget applies: 150 passes, 201 fails at 200. Pass.
  - Result: 4/4. The tests place every citing row on a listed page and every plain row on an unlisted one, so they cannot tell the two rules apart. That gap is the case's claim.
- False-world answer (law #52), case 1: if change derived its budget from a list, the planted line would print `change=pass`, not `fail(changelog#0:160>120)`. Case 2: if the shipped tests pinned the content rule, `suite mutant` would print `fail(n/4)`, because `check.sh:45-58` requires exit 0, four `--- PASS:` lines and zero `--- FAIL:` lines. Both readings discriminate, so the output does not also fit the opposite outcome.
- What it does NOT show: that the content rule is correct beyond the reference. The reference (`check/main.go:17-31`) restates the same `strings.Contains(row, Marker)` rule. §9's new last line states this limit honestly.

### Unplug
- `check.sh:20-23` golden diff: a change to any input line makes `diff -u` fail and exit 1, and the real output is printed first (`cat` at line 20). Nothing is hidden.
- `check.sh:45-58` suites: a compile failure of the mutant copy, a missing test, or a skipped test (no `--- PASS:`) turns the line into `fail(n/4)` with the log on stderr. A `cp` failure exits 2 (lines 34-35). Round 1's `>/dev/null 2>&1` plus exit-code-only check (removed in `b079dc8`) could not tell "no tests ran" from "passed". Round 2 closes that.
- I could not unplug the script myself, because bash is denied. These are read-only judgements of the control flow.

### Closure
Round-1 conditions: the order says they were "carried verbatim into" round 2's order. That text is not in this worktree and not in the diff, so I cannot check each one by its wording (Unrun). From the round-2 delta `5fc084e..b079dc8`, I can name what changed:
- the scratch path inside the repository (`$root/.saphan/work/...`, a private-workspace path) became `mktemp -d` with an EXIT trap: ADDRESSED by reading the code;
- the suite verdicts went from exit code only, with output discarded, to counted `4/4` with logs kept: ADDRESSED;
- a golden check of the three input lines was added: ADDRESSED;
- `go 1.26` became `go 1.20` with `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off`: ADDRESSED (portability, no download);
- §4 changed "select exactly the same rows" to "give the same validation outcomes": this is accurate. Under the list rule, install#0, limits#1 and faq#1 get the narrow budget; under the content rule they do not. All are ≤120, so the outcomes match. See NOTE N2.
- §9 gained a line saying the reference is not independent evidence: ADDRESSED;
- the manifest status object is intact (`test/corpus/manifest.json:322-327`, keys contract/fixture/checks/independent_check).

### Remainder
- Item 3, check bent? No. §2 is unchanged by the diff. The mutant is the contract's mutant: the three named pages (`check/mutant/validator.go:7`). The golden lines equal the measured output and fail closed. The suite counts are strict.
- Item 4, private material: `grep -rnEi 'saphan|macm|@…(com|org)|██████████████████████████████████████████████████'` over the case outside `expected/` returned nothing. Round 2 removed the one private-path reference (`████████████`).
- Item 5, lane: `git diff --name-only 9303b2d...b079dc8 | grep -c expected/` printed `0`. Every path is under `test/corpus/cases/PV-018/` or is `test/corpus/manifest.json`. The manifest hunk sits inside the `"id": "PV-018"` object (lines 312-327) and changes only `status.fixture` and `status.checks`. The master's claim holds.
- Item 6, portability, read rather than run on macOS: `#!/usr/bin/env bash`, `set -u`, `mktemp -d` (no template; valid on macOS ≥10.11 and on GNU), `trap … EXIT`, `cp -R`, `diff -u`, `awk` with POSIX patterns, `[ … -eq … ]`, `$(cd … && pwd)`. No `readlink -f`, no `sed -i`, no GNU-only flags, and nothing beyond bash 3.2. Go code is standard library only and has no map-order dependence in the output (pages and rows are slices). No non-portable call found. The one environmental need: `go` on PATH, ≥1.20.

### Profile
docs (D4: every quoted command is re-run) plus test-power. No security profile: the diff touches no auth, configuration or destructive path in the product. The only deletion is `rm -rf "$scratch"` on a `mktemp -d` result.

### Focus
- §8's quoted command fails as written. `CASE.md:85` says `bash tests/corpus/cases/PV-018/check.sh`. ADR-0030 moved the corpus to `test/corpus`. `ls tests` here prints only `fixtures`, so a judge who copies the command gets "No such file". This is the move's residue: the order itself says the move was "renames only, byte-identical".
- Dates disagree. `CASE.md:82` records the run at `2026-10-23T02:03Z`; `CASE.md:5` and `manifest.json:325` say `checks: run 2026-10-22`.
- `CASE.md:82` cites commit `10dfbe9`. That commit is reachable here only from `remotes/origin/fix/corpus-pv-018`, not from main (`git branch -a --contains 10dfbe9`), so once the branch goes, a public reader cannot resolve it.

### Conditions
- C1 · OWED · `test/corpus/cases/PV-018/CASE.md:85`: the command's path `tests/corpus/...` no longer exists. Remedy: `bash test/corpus/cases/PV-018/check.sh`. Check: the command as quoted runs from the repository root.
- C2 · OWED · `CASE.md:5` + `test/corpus/manifest.json:325` vs `CASE.md:82`: the checks date is 2026-10-22, but the recorded run is 2026-10-23T02:03Z. Remedy: make the status date match the run that §8 records, or record the 2026-10-22 run.
- N1 · NOTE · `CASE.md:82`: `10dfbe9` is not on main. Remedy: drop the hash, or name a hash reachable from main.
- N2 · NOTE · `CASE.md:54`: the fixture author reworded §4, a contract-side section, to fit the fixture's data instead of changing the data. The new text is true and is exactly what EXPECTED rests on, and §2 is untouched, so this does not block. The master may rule whether a fixture author may edit §4.
- Declined to judge (R11): licences (BOUNDARY); whether the case's story (§1, §3, §6) is faithful to the private record (it is outside the diff and I cannot see the record).
- No BLOCKING finding.

### Unrun
- `bash test/corpus/cases/PV-018/check.sh`: refused. Quote: `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████ … do not rebuild the act from other commands`. So the five-line output as a whole and the script's exit code were not measured by me. Lines 1-3 and `suite change` were measured by `go -C … run ./check` and `go -C … test`, which are the admitted forms of their own acts. Owner of the remedy: the master (run the script; expect exit 0 and the five lines of §8).
- `suite mutant=pass(4/4)`: not run. It needs a copy of the module with the mutant swapped in, and I declined to rebuild that from `cp`/`go` (law #57, R15). It is established above by a static trace only.
- R3 mutation of guards: not run. A reviewer does not edit the code under review, and no scratch copy was admitted.
- macOS run: not run (no macOS host in this leg). Portability is judged by reading.
- Round-1 conditions by their verbatim text: not checked. The round-2 order is not in this worktree. Closure is judged from the delta only.

## The review status as filed

## REVIEW — corpus-pv-018
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-018 reproduces §2: go run ./check prints the three input lines of §8 byte for byte; base differs from change only on the planted row; the change suite passes 4/4.
  The expected outcome follows, and both lines discriminate: a list rule would print change=pass on the planted row, and pinned tests would print suite mutant=fail. Lane clean (0 expected/ paths); no private strings; no non-portable call found.
  OWED C1: CASE.md:85 quotes `bash tests/corpus/...`, a path ADR-0030 removed. OWED C2: the checks date 2026-10-22 vs the run 2026-10-23T02:03Z. NOTES N1 (10dfbe9 not on main), N2 (§4 reworded by the fixture author).
  Unrun: check.sh itself (bash refused for this role), the mutant suite (static trace only), macOS, the round-1 condition texts.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T17:55Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| policy binding name | 2 |
