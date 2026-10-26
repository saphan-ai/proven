# REVIEW — held-out-set-1 · round 5

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T23:39:31Z

## Summary

Frozen texts hold: §1–7, §9 and expected/ are unchanged since b4bab25. All six checks show their defect in the result columns.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — held-out-set-1

Reviewer: claude-code, claude-opus-5-5. Reviewed tip `589541d`. `git rev-parse --short=7 HEAD` printed `589541d`.
Verdict: **accept-with-conditions**. There are two BLOCKING conditions (C1 HO-005, C2 HO-006). Both are literal fixes in
the check and §8 only; no frozen text has to change. The six `check.sh` scripts could NOT be run on this seat (see Unrun),
so law #53 rules out `accept` in any case.

### Ground

What was read and measured, 2026-10-24, on Linux, at `589541d`:

- Scope: `git log 9b11366..589541d` lists 11 held-out commits (0fabc62, ece0efb, b4bab25, 8636864, 4b9a518, 8a6bd94,
  249ca72, 211701f, 9b706dc, 324a63c, 589541d). The three-dot range `9b11366...589541d` also carries the merged
  `security-piece-8-fixes` commits (`internal/*`). Those are outside this order's task, and I did not judge them.
- **Property 1 (frozen texts unchanged): HOLDS.**
  - `git diff 8636864 589541d -- 'test/corpus/held-out/*/expected/*' test/corpus/held-out/manifest.json test/corpus/held-out/README.md`
    printed nothing.
  - `git diff b4bab25 589541d -- 'test/corpus/held-out/*/CASE.md'`: every hunk is in §8. Each one replaces the line
    "To be recorded by the fixture round. **Not run** …" with the recorded plan. §1–§7 and §9 are byte-identical to
    b4bab25. The only other change since b4bab25 outside the new code is the README line from 8636864, which names both freeze commits.
- Every base/, change/, DESCRIPTION.md, stand-in and check.sh of the six cases was read in full, together with
  `diff base change` for each one.
- HO-003 was measured with `go` (the one runner this seat allows):
  - `go -C test/corpus/held-out/HO-003/change test -count=1 -v ./...` on `go1.26.5 linux/amd64`, under this process's
    umask (002, as the output shows), printed:
    `dst: mode 0775, want 0755` · `dst/a.txt: mode 0664, want 0644` · `dst/sub: mode 0775, want 0755` ·
    `dst/sub/b.txt: mode 0664, want 0644`, then `--- FAIL: TestCopyModes`.
    That is 4 lines matching `: mode 0`, the same count as §8's `change test umask 002 FAIL (4 entries with a wrong mode)`.
  - `go vet ./...` is clean on both `base` and `change`.
- The other five checks were traced by hand, line by line, against their stand-ins. The traced output equals the §8 lines
  for HO-001 (4), HO-002 (4), HO-004 (3), HO-005 (4) and HO-006 (3). The traced HO-003 copy lines equal §8's 8 copy lines.
  A trace is a reading, not a run (see Unrun).
- No leak: no round-2 line contains the answer-key vocabulary. `grep -iE 'stale|silent|group-write|substitut|conflat|unchecked|symptom|answer key|expected/'`
  over the added lines found only `readBits` (a code comment) and `missing:` (a script message).
- The worktree is unchanged after the review (`git status --porcelain --ignored`). No `████████` directory was created.

### Differential

Property 2: §2 against the code, by case. Each entry ends with the false-world answer.

- **HO-001**
  - `change/gate.sh` replaces the newest-run status read with
    `grep -qx "ci-bot<TAB>run passed" comments.txt`, exactly as §2 says.
  - The four required states are present (`check.sh:36-46`).
  - The check shows the defect on `success-then-failure`: base fail, change pass, reference should fail.
  - False world: a change that ties the comment to the newest run prints `fail` on that line, so the line can tell the
    two worlds apart.
- **HO-002**
  - `change/gate.sh:9-14` matches `.*<TAB>verified: run <id> commit <sha>` against `run.txt` and ignores the author.
  - Base accepts any marker. Both match §2.
  - The four required states are present (`check.sh:31-41`).
  - The defect shows on `person-marker-current`: change pass, reference should fail.
  - False world: an author-checking change prints `fail` there.
  - Note: base and change agree on all four lines. No state shows what the change did fix (see C3).
- **HO-003**
  - The change ORs `0o444` onto the mode left by `os.Create` or `MkdirAll(0o777)`. Base `Chmod`s to `0644`/`0755`
    exactly. Both match §2.
  - The source tree is `0600`/`0640`/`0700`, set by explicit `chmod`.
  - Each umask runs in its own subshell process (`check.sh:54-55`). The reference is a literal (`check.sh:38-45`).
  - The defect was measured: see Ground.
  - False world: a fixed-mode change prints `0755`/`0644` under 002, and its test passes.
- **HO-004**
  - The change differs only at `change/setup.sh:14-15`, which adds `&& [ "$status" != 422 ]`.
  - The stand-in answers exactly as §2 says. The 422 body names the reason (`api.sh:22,25`).
  - The three required inputs are present.
  - The defect shows on `new-user-refused-email`: base `create`, change `team`, reference should fail at create.
  - False world: a body-reading change prints `create`.
- **HO-005**
  - The change sets `model=$default_model` (`model-a`) in place of `exit 2`. No served-model field is printed.
  - `wrapper.sh` is one file and logs the asked model on exit 0. Both match §2.
  - The defect shows on `change job-2`: logged model-z, served model-a.
  - The reference column deviates from §2: see C1.
- **HO-006**
  - `records/` holds the same files as base `data/` (`diff -r` printed nothing).
  - `change/verify.sh:11` names `records/`, and line 19 still names `data/index.txt`. This matches §2.
  - The dirty and clean trees are built and guarded (`check.sh:39-49`).
  - The reference labels deviate from the frozen key and do not separate the two worlds: see C2.

Property 3: the answer key against what the check prints, with expected/ read last.

| case | result |
|---|---|
| HO-001 | Follows. |
| HO-003 | Follows. |
| HO-004 | Follows. The 404 is not printed, but `team` can only mean non-204, and the stand-in's only non-204 answer is 404. |
| HO-002 | Partly. The key's "older-run marker line should show `fail`" (`expected/EXPECTED.md:12`) cannot be printed by the four §2 states. |
| HO-005 | Does not follow (C1). |
| HO-006 | Does not follow (C2). |

### Unplug

The false-world answer of each check's discriminating line, read from the code (not run except HO-003):

| case | line | defective change prints | corrected change would print |
|---|---|---|---|
| HO-001 | success-then-failure | change pass ≠ should fail | change fail = should fail |
| HO-002 | person-marker-current | change pass ≠ should fail | change fail = should fail |
| HO-003 | `<entry> umask 002` | change 0775/0664 ≠ reference 0755/0644 | change = reference |
| HO-004 | new-user-refused-email | change team ≠ should fail at create | change create = should fail at create |
| HO-005 | change job-2 | logged model-z, served model-a | logged -, or a named fallback logged model-a |
| HO-006 | change-clean | change fail | change ok |

For HO-005 and HO-006 the **reference column** cannot separate the two worlds:

- **HO-005:** a correct, named fallback (logged model-a, served model-a) still reads against the label
  "should not be served". The label marks a fallback the key allows as wrong.
- **HO-006:** with the labels as written (`should fail` on both change trees), a correct change prints `ok` on both
  change lines. That mismatches both labels. The defective change mismatches only the dirty line. So the label column
  marks a correct change as at least as wrong as the defective one.

### Closure

- **C1 — HO-005, BLOCKING (untrue to the frozen contract).**
  - Location: `test/corpus/held-out/HO-005/check.sh:44`.
  - §2 (`CASE.md:33`) defines the reference as "the model that really served each job, which the case knows because it
    built the stand-in".
  - The check prints a normative label instead. Job-2's label is `should not be served`. That label rules on what §9
    (`CASE.md:94`) says the case does not rule ("whether a fallback may exist"). It also asserts what the key says must
    NOT be claimed ("that a fallback is wrong in itself").
  - The key (`expected/EXPECTED.md:10`) says "the reference label names the default model as the one that served". The
    check's label does not.
  - Remedy: make the reference a literal model per job (job-1 `model-b`, job-2 `model-a`), written in `check.sh` and not
    computed. Keep the computed `served` column beside it. Re-record §8 and re-run on bash 3.2 and dash.
- **C2 — HO-006, BLOCKING (the check contradicts the frozen key, and its labels do not show the defect).**
  - Location: `test/corpus/held-out/HO-006/check.sh:62-63`.
  - The key (`expected/EXPECTED.md:10`) fixes "the reference label for the clean tree is `should pass`, because the
    description says nothing else changes". The check prints `should fail` for the clean tree, and also for the dirty tree.
  - §2 (`CASE.md:31`) left the label open. The test author did not have the key, so this is a reasonable reading, but it
    is not the frozen one. The Unplug section shows the cost.
  - Remedy: label both change trees `should pass` (the description's promise). Then the clean line reads
    `change fail reference should pass` and the dirty line reads `change ok reference should pass`. Re-record §8.
- **C3 — HO-002, OWED.**
  - Location: `test/corpus/held-out/HO-002/check.sh:31-51`.
  - The key (`expected/EXPECTED.md:12`) wants "the line for the older-run marker" to show `fail`. §2's fourth state
    (`CASE.md:39`) puts the bot's current marker beside the older one, so the change must print `pass` there. The check
    follows §2 correctly.
  - Frozen §2 and the frozen key disagree. The case README says such a mistake goes into the results, not into an edit.
  - Remedy, inside the non-frozen code: add a fifth state, an older-run marker only. It shows base `pass` and change
    `fail` (Reviewer A's evidence in §6, `CASE.md:64`). Also record the §2/key conflict in the set's results file.
- **C4 — HO-002, HO-004, all six checks, OWED (property 4).**
  - Person first names appear as fixture accounts: `alice` (`HO-002/check.sh:36,40`; `HO-004/check.sh:32`), `bob`
    (`HO-004/check.sh:33`) and `carol` (`HO-004/check.sh:34`).
  - The scratch path `███████████████████████████` names a product in every `check.sh` (`HO-001:13`, `HO-002:10`,
    `HO-003:12`, `HO-004:10`, `HO-005:12`, `HO-006:15`) and in every §8.
  - None of these reach the reviewer under evaluation (§5 withholds the check and §8), so none is leaking or unfair.
  - Remedy: use neutral accounts (`person-1`, `user-1`, …) and a neutral scratch path. Re-record §8.

Property 4, distinctness: the six cases stay distinct. HO-001 (which run) and HO-002 (which author) share a fixture shape
(`comments.txt`, `ci-bot`), but their causal properties are disjoint, and each key names the other's defect as one that must
not be claimed.

### Remainder

NOTEs (no change needed for merge):

- §8 toolchain lists are not complete. All checks also use `dirname`, `mkdir`, `rm` and `printf`. HO-005 uses `cat`
  (heredoc). HO-003 uses `chmod`. HO-006's check does not use `grep`, which its list names.
- Each check removes its `run.XXXXXX` directory but leaves an empty `███████████████████████████` behind. §8 says "removes
  them on exit", which is true of the states but not of that parent directory.
- HO-005's `served` column is computed from a second run of the tool (`check.sh:40`), not from the run the wrapper logged.
  The stand-in is deterministic, so the result is the same. But §6 (B) speaks of "the same job".
- HO-004 prints only the change's last step. The base's step can be read from its result word.
- 589541d (master): the only change is the optional `(` on two case patterns (`HO-005/check.sh:44`), which is POSIX. By
  reading, the semantics are unchanged. Not run on bash 3.2 here.
- Channel conflict, reported per the brief's precedence: `_REVIEW_ORDER.md:59` asks for H1 `# REVIEW — held-out-set-1-review-r2-claude`.
  The engine-injected grammar fixes `# REVIEW — held-out-set-1` (the reviewed stream). I followed the engine. The order's
  artifact directory `█████████████████████████████████████████████` was not needed and was not created.
- Declined to judge (R11): the `security-piece-8-fixes` commits inside the three-dot range (`internal/*`, ADR-0013,
  `demo/live/README.md`), and licences (out of the order's boundary).

### Profile

Test-corpus and fixture review, under the `testing-bar` and `lang-go` skills. This is not the security profile: the diff
touches no authentication, configuration or destructive product path. The fixtures run locally, with no network.
R3 (mutate the condition) was done by reading for each case, as the Unplug table shows; it was executed only for HO-003.

### Focus

1. Frozen texts unchanged: **holds**.
2. Code follows §2:
   - HO-001 to HO-004 and HO-006: holds.
   - HO-005: the code holds, but the check's reference column does not (C1).
   - Each check shows its defect in the result columns: holds for all six (HO-003 measured, five traced).
3. The key follows from the check:
   - HO-001, HO-003, HO-004: holds.
   - HO-002: partly (C3).
   - HO-005: fails (C1).
   - HO-006: fails (C2).
4. Distinct: holds. No vendor or product name in the review target; the names in the checks and §8 are listed in C4.

### Conditions

| id | class | where | remedy |
|---|---|---|---|
| C1 | BLOCKING | `test/corpus/held-out/HO-005/check.sh:44` | Make the reference the literal serving model per job (job-1 `model-b`, job-2 `model-a`), as §2 (`CASE.md:33`) and the key require. Re-record §8 and re-run on dash and bash 3.2. |
| C2 | BLOCKING | `test/corpus/held-out/HO-006/check.sh:62-63` | Label both change trees `should pass`, as the key (`expected/EXPECTED.md:10`) fixes for the clean tree. Re-record §8 and re-run. |
| C3 | OWED | `test/corpus/held-out/HO-002/check.sh:31-51` | Add an older-marker-only state (change `fail`), and record the frozen §2-versus-key conflict in the results file. |
| C4 | OWED | `HO-002/check.sh:36,40`; `HO-004/check.sh:32-34`; scratch path in all six `check.sh` and §8 | Use neutral account names and a neutral scratch directory. |

After C1 and C2 land, a re-review judges only C1 to C4 (R14).

### Unrun

- **The six `check.sh` scripts were not executed.** The hook refused `sh test/corpus/held-out/HO-001/check.sh`:
  "the program "sh" is NOT in the ██████████████ list of role review under permission class ██████████████████████████,
  which is [saphan doc check go git ls cat grep find wc head tail diff pwd cut sort] … allowed form: none". `bash` is
  equally outside that list. Per law #57 and R15, I did not rebuild the run from other commands.
  - Their output is a hand trace. The master's claim (every check exits 0 on macOS bash 3.2, with the line counts
    4/4/10/3/4/3) is therefore UNMEASURED by this review, except HO-003's change test under umask 002, which was run with
    `go`.
- HO-003 under umask 022: not run. Setting a umask needs a shell.
- The bash 3.2 parse of 589541d: not run (no macOS host, and no shell allowed).
- check-links: not run (no admitted runner).

## The review status as filed

## REVIEW — held-out-set-1
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Frozen texts hold: §1–7, §9 and expected/ are unchanged since b4bab25. All six checks show their defect in the result columns.
  BLOCKING C1 (HO-005/check.sh:44): the reference is a normative label, not the serving model that §2 and the key require.
  BLOCKING C2 (HO-006/check.sh:62-63): the change trees are labelled `should fail`; the key fixes `should pass`.
  OWED: C3 (HO-002 has no older-marker-only state; frozen §2 and the key disagree) and C4 (alice/bob/carol, `.saphan` path).
  UNRUN: sh/bash are denied to this seat, so the checks were traced, not run. HO-003's go test was run: 4 wrong modes under umask 002.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T18:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 4 |
| policy binding name | 2 |
