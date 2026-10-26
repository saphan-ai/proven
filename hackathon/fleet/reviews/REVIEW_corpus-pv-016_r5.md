# REVIEW — corpus-pv-016 · round 5

- verdict: reject
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 2 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:58:20Z

## Summary

PV-016 at 3a84be4: fixture content matches §2, expected outcome follows, lane and privacy hold. BLOCKING B1:

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-016

Reviewer: claude-code, model `claude-opus-5-5`. Run of stream corpus-pv-016-review-r6, reviewing corpus-pv-016.
Scope: `9303b2d...3a84be4` (2 commits: 516e3d4 round 1, 3a84be4 round 2). Verdict: **reject** (one BLOCKING, below).

### Ground

- `git rev-parse --short=7 HEAD` printed `3a84be4` (role-lock satisfied).
- `git log 9303b2d..3a84be4`: `3a84be4 [corpus-pv-016] tests: enforce page quotes and clean up binaries`,
  `516e3d4 [corpus-pv-016] tests: build PV-016 base, change and check; record run`.
- `git diff --name-status 9303b2d...3a84be4`: 13 files added under `test/corpus/cases/PV-016/` (base/, change/, check.sh,
  claims.tsv, reference.md), `M CASE.md`, `M test/corpus/manifest.json`. Nothing under `expected/`.
- `ls tests` → only `fixtures`; `find tests -maxdepth 3` → `tests`, `tests/fixtures`, `tests/fixtures/pull_request.json`.
  The directory `tests/corpus` does not exist at 3a84be4. ADR-0030 line 33: `test/ corpus/ (from tests/corpus)`.
- `grep -rn "tests/corpus" test/corpus/cases/PV-016 test/corpus/manifest.json` →
  `check.sh:2` (comment), `check.sh:7` (`case_dir=tests/corpus/cases/PV-016`), `CASE.md:102` (the documented command).
  No other case in `test/corpus/cases` names `tests/corpus`.
- `diff -r test/corpus/cases/PV-016/base test/corpus/cases/PV-016/change` → `Only in .../change: docs`. Code identical.
- `go -C test/corpus/cases/PV-016/base vet .` → clean, no output.
- **`bash test/corpus/cases/PV-016/check.sh` was NOT run.** The leg's permission class refused it, verbatim:
  `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████ ... allowed form: none`.
  Per law #57 I did not rebuild the check from `go run` calls. Everything about the check below is from reading it
  against the tree; see Unrun.
- The order's round-1 review conditions were said to be "carried verbatim into its order" (round 2's), but they were not
  carried into THIS order. I cannot read them. See Unrun.

### Differential

Property by property, as the order asks.

1. **Fixture reproduces §2 — the content does; the check cannot run at this SHA.**
   - §2 items, each read in the tree:
     - one code for base and change: `diff -r` shows only `change/docs` differs. ✓
     - op 1 accept/refuse with reason: `alerts.go:31-43` (`ParseDestination`), `main.go:16-22`. ✓
     - op 2 considered/delivered to an in-memory receiver: `alerts.go:52-54` (`Receiver`), `alerts.go:75-91` (`Emit`). ✓
     - op 3 field list: `alerts.go:94-101` (`FieldsOf`). ✓
     - sender always sends nothing, reason "no transport yet": `alerts.go:63-65` returns `nullSender{reason: "no transport yet"}`, `alerts.go:58` `Send` returns false. ✓
     - four pages (index, what-goes-out, chat-setup, related): present under `change/docs/alerts/`. ✓
     - promise of delivery: `what-goes-out.md:3`. ✓
     - setup step the code refuses: `chat-setup.md:4` `chat:t-backup-ops` vs `alerts.go:36` (`c-` prefix required). ✓
     - fields claimed and not emitted: `retry_count` (backup_failed), `threshold` (storage_low) vs `alerts.go:19-25`. ✓
     - four page-internal findings: "carries no actor" vs related.md routing by actor; no page says whose environment
       holds the credential or which machine sends; `delivery_failed` goes to the same single destination; index says
       ids begin with `c`, chat-setup allows `t`. All four are present in the pages and listed in `reference.md:19-25`. ✓
     - synthetic credential only: `tok_synthetic_0000_not_a_secret`. ✓
   - Base differs from change exactly where §2 puts the disagreement: only the pages. ✓
   - **BLOCKING B1.** `check.sh:7` sets `case_dir=tests/corpus/cases/PV-016`. That path does not exist at 3a84be4
     (Ground). Run from the repository root as `CASE.md:102` and `check.sh:2` say, the first loop
     (`check.sh:15-20`) calls `diff -q tests/.../base/go.mod tests/.../change/go.mod`. `diff` exits 2 on a missing file,
     the `!` branch fires, and the script prints `check: base/go.mod and change/go.mod differ` and exits 1. It never
     builds or prints a claim line. The command in `CASE.md:102` (`bash tests/corpus/cases/PV-016/check.sh`) does not
     find the script at all. So the recorded output in §8 and "Exit code: 0" were true of the pre-move tree and cannot be
     reproduced at the reviewed SHA. Cause: the carry onto ADR-0030's layout was renames only; the path inside the files
     was not re-pointed. This refutes the master's claim "check.sh exit 0" for SHA 3a84be4 (it may have held before the
     carry; I did not measure that tree).
2. **Expected outcome follows from the evidence — yes, given a check that runs.** `expected/EXPECTED.md` says:
   established defect in change; proof = delivered=0 for each listed alert, the setup step refused, the four
   page-internal pairs. The recorded change lines pair each page quote with `delivered=0` and the `t-` refusal.
   False-world answer: if the code delivered, the `emit` line would print `delivered=1`; if the code accepted `t-` ids,
   D-setup would print `accepted:`. Both differ from what was recorded, so the instrument can tell the worlds apart.
   Base also prints delivered=0; that is correct, because the defect is the pages' promise against shared code, and
   only change carries the promise. EXPECTED's "do not claim a defect in other operations" is consistent with D-index
   accepted and three F-lines `claimed-but-absent=none`.
3. **The case still proves something — yes.** The check prints and does not judge; nothing is tuned to come out "as
   expected". Round 2 made it stricter, not looser: a missing change-side quote now sets exit 1 (`check.sh:37-40`).
   Before round 2, a missing quote printed `(no such page)` and exited 0, the same answer as a world where the page
   says nothing (law #52); round 2 closes that, with a negative control recorded in `CASE.md:134-136`.
4. **Nothing private is left — holds.** `grep -rniE "macm5|███████|███████████████████████████████████████████████████)|ghp_|sk-|xox"`
   over the case → no match. Model names in `CASE.md` §3 are public and match `PROVENANCE.md:46` (P29). The go module
   is `example.invalid/...`.
5. **The diff stays in its lane — holds.** Only `test/corpus/cases/PV-016/` and lines 290-291 of the manifest (the
   status object of the PV-016 entry: `fixture`, `checks`). `expected/` untouched. Master's lane claim confirmed.
6. **Runs on Linux and macOS — not as it stands (B1), otherwise portable by reading.** Calls used: `bash` (shebang
   `/usr/bin/env bash`), `mktemp -d <template>` (BSD and GNU both accept a template), `go -C` (Go ≥ 1.20; go.mod says
   1.26), `diff -q`, `grep -qF --`, `rmdir` with several args, `trap ... EXIT`, `${claimed//,/ }` and `IFS=$'\t'`
   (both work in macOS's bash 3.2). No GNU-only flag found. Not executed on either OS by me.

### Unplug

- Quote guard (`check.sh:37-40`): unplugged, a change page that lost its quoted sentence prints `(no such page)` and the
  run exits 0 — indistinguishable from base. The case records this exact negative control (`CASE.md:134-136`). Not
  re-run by me (bash refused).
- Same-code guard (`check.sh:15-20`): it treats a missing file as "differ". So it fires red at this SHA, but with a false
  reason. A judge would read "base and change differ" and look in the wrong place. NOTE N1.
- Unplugging `resolveSender` (making it deliver) would flip every P-line to `delivered=1`; the check would report it. No
  test in the case asserts delivered=0 by itself; the check only prints. That is by design (§2: the check judges nothing).

### Closure

- B1 is not closed at 3a84be4. Remedy: make the case directory independent of the layout — e.g. derive it from the
  script's own location (`case_dir=$(cd "$(dirname "$0")" && pwd)`), or set it to `test/corpus/cases/PV-016`; fix
  `check.sh:2` and `CASE.md:102`; re-run from the repository root; record the new run time, exit code and output in §8;
  keep or refresh the `checks` date in the manifest only after that run.
- Round-1 conditions: the round-2 diff shows quote enforcement with a negative control, binaries moved into a
  `mktemp -d` dir removed by an EXIT trap, a dated manifest status object, and a "Not run: macOS" line. Whether these
  are ALL the round-1 conditions, and each one paid as written, I cannot say: the conditions were not given to me.

### Remainder

Owed / notes (none blocks on its own):
- **N1** `check.sh:16` — `diff -q` exit 2 (missing file) is reported as "differ". Distinguish a missing file from a
  difference so the refusal names what it met.
- **N2** `check.sh:35-40` — on a missing change-side quote the line still prints `pages say: (no such page)`, which is
  false when the page exists. The exit code is right; the label is not.
- **N3** `claims.tsv` column 4 (claimed fields) is a hand copy of the list inside column 6 (the quote). Nothing checks
  they agree. Today they agree on all five F-rows (read). A drift would make `claimed-but-absent` wrong while the quote
  still matches.
- **N4** `check.sh:29` — `IFS=$'\t'` is whitespace IFS, so an empty TSV field collapses and shifts the columns. Today
  every field is filled (`-` for empty). Fragile, not wrong.
- **N5** `check.sh:8` writes under `$PWD/.saphan/work/` and `.gitignore` has no entry for it; the EXIT trap removes it,
  but a SIGKILL leaves untracked files in the judge's checkout. Consider a plain `mktemp -d` outside the tree.
- **N6** `CASE.md:5` and manifest line 291 say "checks: run 2026-10-23". True of a run, not of the tree as reviewed (B1).

Declined to judge (R11): licences (order: out of scope); the evaluation harness that consumes `reference.md`.

### Profile

Corpus / test-fixture review. Not the security profile: the diff touches no auth, secrets handling, configuration or
destructive path of the product. The skills `lang-go` and `testing-bar` were named; the fixture is standard-library Go
and passed `go vet`.

### Focus

- The six properties of the order, in order.
- The ADR-0030 carry, because the order states it as "renames only": renames only is exactly what left `tests/corpus`
  inside the files.
- The master's two claims to re-measure: lane (confirmed) and `check.sh exit 0` (refuted at 3a84be4 by reading; see
  Unrun for why not by running).

### Conditions

1. **BLOCKING B1** — `test/corpus/cases/PV-016/check.sh:7` (and `:2`, `CASE.md:102`): the case path is the pre-ADR-0030
   `tests/corpus/...`, which does not exist at 3a84be4; the check exits 1 before printing a claim line. Remedy: as in
   Closure; re-run and re-record §8. Proof: the documented command, run from the repository root, exits 0 and prints
   the 24 lines.
2. OWED N1–N6 as in Remainder. They do not open another round.

### Unrun

- `bash test/corpus/cases/PV-016/check.sh` — refused by the leg's permission class (`"bash" is NOT in the ██████████████
  list of role review ... allowed form: none`). Not rebuilt from other commands (law #57). B1 rests on reading
  `check.sh:7` and on `ls tests` / `find tests -maxdepth 3`, which show the path absent.
- The round-1 review conditions — not carried into this order, so I cannot mark each ADDRESSED / NOT ADDRESSED.
- macOS execution — no macOS host in this leg.
- The pre-move tree (to confirm the recorded "Exit code: 0" was true before the carry) — out of the reviewed SHA.

## The review status as filed

## REVIEW — corpus-pv-016
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: PV-016 at 3a84be4: fixture content matches §2, expected outcome follows, lane and privacy hold. BLOCKING B1:
  check.sh:7 still uses pre-ADR-0030 path tests/corpus/..., absent at this SHA, so the documented command exits 1
  ("base/go.mod and change/go.mod differ") and §8's exit 0 cannot be reproduced. check.sh itself UNRUN (bash refused by
  permission class); round-1 conditions not carried into this order, so not judged one by one. Six owed notes.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T18:20:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| policy binding name | 3 |
