# REVIEW — corpus-pv-003 · round 3

- verdict: reject
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 3 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T16:45:53Z

## Summary

REJECT at cbab00f. The case logic matches §2, and the output it records would establish EXPECTED, with a discriminating false-world answer.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-003

🔴 **RED — CROSS-CUTTING (law #6): a rename-only carry across ADR-0030 breaks every path written INSIDE a moved file.**
The case was carried from `tests/corpus` to `test/corpus` "renames only, byte-identical". That is exactly the defect:
`check.sh` and `check/main.go` still name `tests/corpus/...`, a directory that does not exist at the tip. At cbab00f the
check cannot run, and `go build ./...` for the whole `github.com/saphan-ai/proven` module fails. A byte-identity check
on a move cannot see a path string inside the moved file. Every other case carried the same way needs the same grep.

- reviewer: claude-code, model claude-opus-5-5 (Anthropic family), headless, permission class `██████████████████████████`
- reviewed range: `9303b2d...cbab00f` (2 commits: abf0c0f round 1, cbab00f round 2)
- verdict: **reject**

### Ground

`git rev-parse --short=7 HEAD` printed, verbatim: `cbab00f`. Role lock holds.

Read in full: the diff `9303b2d...cbab00f` (7 files, +277 −5), `CASE.md`, `expected/EXPECTED.md`, `base/gate.go`,
`change/gate.go`, `check/main.go`, `check/reference.go`, `check.sh`, the manifest entry for PV-003 (lines 50–66).

Measurements at cbab00f (my own, 2026-10-23):

- `ls -d tests test` → both exist. `git ls-files tests` → only `tests/fixtures/pull_request.json`. No `tests/corpus`.
  At the base 9303b2d, `git ls-tree` shows the corpus already under `test/corpus`; `tests/corpus` existed at neither end.
- `grep -rn "tests/corpus" test/corpus/cases/PV-003` →
  - `check.sh:3` (comment) and `check.sh:5` `exec go run ./tests/corpus/cases/PV-003/check`
  - `check/main.go:9` and `:10`, imports `github.com/saphan-ai/proven/tests/corpus/cases/PV-003/{base,change}`
  - `CASE.md:94`, `:95`, `:117`: the recorded commands
- `go list ./tests/corpus/cases/PV-003/check` → exit 1:
  `stat …/tests/corpus/cases/PV-003/check: directory not found` (worktree prefix elided)
- `go vet ./test/corpus/cases/PV-003/...` → exit 1:
  ```
  test/corpus/cases/PV-003/check/main.go:9:2: no required module provides package github.com/saphan-ai/proven/tests/corpus/cases/PV-003/base
  test/corpus/cases/PV-003/check/main.go:10:2: no required module provides package github.com/saphan-ai/proven/tests/corpus/cases/PV-003/change
  ```
- `go build ./...` → exit 1, same two lines. `go build ./cmd/... ./internal/... ./test/corpus/cases/PV-003/base ./test/corpus/cases/PV-003/change`
  → exit 0, no output. The new `check` package is the only failing package. This diff added it.

**Question 1: run `check.sh` and paste its output. UNRUN.** The hook refused `bash`:
`the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████ … allowed form: none`.
I did not rebuild the act from other commands (R15, law #57). The measurements above show what the script would meet:
line 5 names a directory that does not exist, so `go run` fails before it prints any line.

The master's claim (macm5, 2026-10-22), "`check.sh` exit 0", is **refuted at cbab00f** by the `go list` and `go vet`
readings above. It held on the pre-move layout. The carry onto `test/corpus` broke it.

### Differential

What the diff does, judged against `CASE.md` §2. I traced the logic by reading, not by running.

| input | base | change | reference | trace |
|---|---|---|---|---|
| own-case-cooled-slot | allowed | refused(A) | refused(A) | base: B's clean run on n1 deletes `cool[n1]` (`base/gate.go:39`). This is the defect the change fixes. |
| own-case-other-slot | allowed | allowed | allowed | nothing in cooldown on (n1,2) or for B |
| move-moved-tenant | refused(X) | allowed | refused(X) | change looks up (n1,2), and X's entry is under (n1,1) (`change/gate.go:48`) |
| vacated-slot | refused(X) | refused(X) | allowed | change: (n1,1) still holds X's entry; reference: Y is on (n1,1) and has no cooldown |
| control-no-cooldown | allowed | allowed | allowed | no exhausted run |

My trace matches, line for line, the output `CASE.md` §8 records for both runs. On the pre-move layout, the code
computes what §8 says.

- **§2 conformance.** Base keys on the node (`base/gate.go:21`). Change keys on (node, slot) and never reads the tenant
  for gating (`change/gate.go:22`, `:36–43`). The reference keys on the tenant (`check/reference.go:16`, `:29–34`) and
  resolves the slot's current tenant (`:38–45`). All four required inputs are present (`check/main.go:25–60`). Each line
  carries input, base, change and reference. **Met.**
- **Where base and change differ:** own-case-cooled-slot and move. They agree on vacated-slot and on both controls.
  **Where change and the reference differ (§2's disagreement):** exactly move and vacated-slot, and nowhere else.
  **Holds by reading.**
- **Question 2: expected outcome.** `EXPECTED.md` asks for: change = allowed / refused(X) on move / vacated, the
  reference = refused / allowed, and agreement on the own cases and the control. The recorded lines carry exactly that.
  **False-world answer (law #52):** if the change were correct (keyed on the tenant), the move line would read
  `change=refused(X)` and the vacated line `change=allowed`. If the change were the base again, own-case-cooled-slot
  would read `change=allowed`. Each world prints a different line, so the output discriminates. **Holds, conditional on
  the check running.**
- **Question 3: was a check bent?** No. The open build choices (`CASE.md:105–107`: a clean run clears its own key, a
  clock that never moves, an unassigned slot asked with no tenant) apply alike to all three gates. None of them feeds the
  move or vacated lines: in those inputs, no clean run happens. **Holds by reading.**
- **Question 5: lane.** `git diff --name-status 9303b2d...cbab00f`: 5 files under `test/corpus/cases/PV-003/` (none under
  `expected/`) and `M test/corpus/manifest.json`. The manifest hunk is lines 58–64, inside the PV-003 entry (lines 50–66).
  The status object is intact. **Holds.** The master's claim "never `expected/`" is confirmed.
- **Question 4: private strings.** `grep -rniE "macm|███████|wato|cyberasis|████████████████████████████████████████████████████)"`
  over the case and the manifest finds only two hits. Both are titles of other cases (manifest lines 247 and 298), not
  in this diff. **Holds.** Note: the commit author `██████████████████████` is the leg identity the engine sets. It is
  outside the files under review, and the order put no question on it.
- **Question 6: portability.** `check.sh` uses `#!/usr/bin/env bash`, `set -euo pipefail` (present in macOS's Bash 3.2)
  and `exec go run`. No GNU-only flag, no `sed -i`, no `readlink -f`, no `date`. Nothing non-portable by reading. But
  the script runs on **neither** Linux nor macOS at the tip, because of the path defect. `go.mod` says `go 1.26`: a judge
  with an older Go needs a toolchain download, which needs the network. That applies to the whole repo, not only this
  case. The file mode is `100644`, which is fine because the script is called as `bash check.sh`.

### Unplug

- **The R3 mutants were reasoned, not run.** (a) Change keyed on the tenant: move → refused(X), vacated → allowed; the
  defect lines turn. (b) Change keyed on the node (base again): own-case-cooled-slot → allowed. (c) The reference's
  `Allow` does not resolve the slot's tenant: vacated would still read allowed, but move would read allowed. The check
  can see each mutant. I could not run any of them (see Unrun).
- **The case's own alarm for the path defect is absent.** Nothing in the case or the carry procedure reddens when an
  internal path goes stale. The byte-identity check on the move is an instrument whose answer is the same whether the
  paths inside are right or wrong (law #52). That is why the defect passed the carry.

### Closure

Round-1 conditions. **Neither round-1 review was carried into this order** (law #44 §3: carry, do not point), and none
is in this repository (`grep -rln PV-003` finds only the corpus, `_REVIEW_ORDER.md` and the train order, and the train
order holds no condition text). What I can judge from the artifact:

- the status OBJECT in `manifest.json` is restored (lines 58–64, four keys) → **ADDRESSED**;
- native macOS verification: `CASE.md:120–122` says, in its own words, that "the review's cross-platform verification
  condition remains open" → **NOT ADDRESSED** (declared, honestly, under "Not run");
- every other round-1 condition: **cannot be judged.** The text did not reach this leg. The master must name them and
  map each to its fix before the gate.

New in round 2 and blocking: the paths are stale (see Conditions, C1).

### Remainder

- **NOTE.** `check/reference.go:38–45` walks a Go map. If two tenants were ever assigned to the same slot, the refusal's
  name would be chosen at random. No input reaches this: `Assign` never puts two tenants on one slot in the five inputs.
  The output is deterministic today.
- **NOTE.** `CASE.md` §5 lists "the change's tests" as part of the review target. The case ships no tests for
  `change/`. §2 does not require them, so this is not a §2 failure. It is an open question for the master: what does
  the evaluated reviewer receive as "the change's tests"?
- **NOTE.** `base/gate.go` keeps `slots` and the tenant and slot parameters of `Record` but never reads them, and so
  does `change/gate.go` for the tenant. This is deliberate under §2 ("ignores the tenant identity"). `go vet` would not
  flag it. Fine as written.
- Declined to judge (R11): licences (out of scope by order); the wording of `EXPECTED.md` (not in the diff); the
  provenance claims in §3 and §6 (not in the diff).

### Profile

docs + test fixture. No security profile: the diff touches no authentication, secret, network or destructive path.
It is in-memory Go and a two-line shell wrapper.

### Focus

Order questions 1–6, the round-1 condition closure, and the master's re-measurable claim (`check.sh` exit 0; the diff
stays in its lane). The weight went on question 1, because the carry onto `test/corpus` is the one change after the
last measured run.

### Conditions

- **C1 · BLOCKING · the check cannot run at the reviewed commit, and `go build ./...` for the module fails.**
  - Where: `check.sh:5`, `check/main.go:9`, `check/main.go:10`.
  - Trigger: `bash test/corpus/cases/PV-003/check.sh` from the root at cbab00f, or `go build ./...` / `go vet ./...` /
    `go test ./...`.
  - Evidence: the `go list`, `go vet` and `go build` output in Ground.
  - Remedy: replace `tests/corpus` with `test/corpus` in those three lines and in the comment at `check.sh:3`. Then
    re-run `bash test/corpus/cases/PV-003/check.sh` and `go build ./...` on the tip, and record the exit codes and the
    full output.
- **C2 · BLOCKING · `CASE.md` §8 records commands that fail at this commit as if they ran.**
  - Where: `CASE.md:94`, `:95`, `:117`.
  - Problem: they name `tests/corpus/...` and an "exit code: 0" that cannot be reproduced at cbab00f (AGENTS.md: "No
    document claims more than was run").
  - Remedy: update the commands to `test/corpus/...`. Record a third run made at the new path, and say that the earlier
    two ran on the pre-move layout.
- **C3 · OWED · the native macOS run remains open** (`CASE.md:120–122`), which is a round-1 condition by the case's own
  account. Remedy: one run of the fixed `check.sh` on macOS with its output recorded, or an explicit ruling by the owner
  that the declared "Not run" is enough.
- **C4 · OWED (master) · name the round-1 conditions.** List each round-1 condition and the commit that pays it. This
  leg could not read them, so it could not close them.
- **C5 · OWED (train) · sweep the other carried cases** for `tests/corpus` inside files (the RED header). This is
  outside this diff: a shout, not a condition on PV-003.

### Unrun

- `bash test/corpus/cases/PV-003/check.sh`: the hook denied `bash` under `██████████████████████████` (refusal quoted
  in Ground). Not rebuilt from other commands. The static readings show it would fail at line 5.
- The R3 mutants: they need edits to the code under review and a run of the check. This leg does not modify the code
  under review, and the check does not compile at the tip.
- macOS: no macOS host on this seat.
- `gofmt -l`: `gofmt` is not in the permitted command list.
- `go build ./...` at the base 9303b2d: not run, because a second checkout would write outside this worktree. The base
  holds no Go code under `test/corpus` (`git ls-tree`), so the new `check` package is the only candidate cause.

## The review status as filed

## REVIEW — corpus-pv-003
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: REJECT at cbab00f. The case logic matches §2, and the output it records would establish EXPECTED, with a discriminating false-world answer.
  But the rename-only carry onto test/corpus left check.sh:5 and check/main.go:9-10 on tests/corpus, which does not exist.
  At the tip, the check cannot run and `go build ./...` for the module fails (C1). CASE.md §8 records exit 0 for commands that now fail (C2).
  macOS stays open (C3); the round-1 conditions were not carried to this leg (C4). check.sh was UNRUN: bash denied by the class.
  RED: sweep every case carried across ADR-0030 for stale internal paths.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 2 |
| policy binding name | 4 |
