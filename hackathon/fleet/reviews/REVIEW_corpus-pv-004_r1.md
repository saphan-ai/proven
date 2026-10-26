# REVIEW — corpus-pv-004 · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-21T18:46:23Z

## Summary

PV-004 fixture reproduces both opener defects §2 names (info-string backtick, four-space indent: change returns a command, reference does not), well-formed agrees, suite green; the outcome in expected/ follows. Hand-traced all 11 lines against §8 and they match; check.sh NOT run here (bash outside this seat's class) — C4.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-004

Reviewer: claude-code (Anthropic family), opus. Reviewed: `b0578227d...6fbc99639` (1 commit, 13 files), HEAD measured `6fbc99639` by `git rev-parse --short=9 HEAD`. The answer key `expected/EXPECTED.md` was read only after questions 1 and 3 had been traced.

### Ground

- Contract read: `tests/corpus/cases/PV-004/CASE.md` §2 (two extractors, base never refuses, change refuses three named classes, strips each line, never looks at the info string, an eight-case suite that varies only the closer, a CommonMark reference, eleven inputs, no network, no model).
- Code read in full: `base/extract.ts`, `change/extract.ts`, `change/suite.ts`, `check/run.ts`, `check/reference.ts`, `types/ambient.d.ts`, `check.sh`, `package.json`, `tsconfig.json`, and the lock file's `resolved` lines.
- **The check was NOT run by this reviewer.** `bash tests/corpus/cases/PV-004/check.sh` was refused by this seat's permission class (`bash` not in `[saphan doc check go git ls cat grep find wc head tail diff pwd]`); tried once, stopped, and did not rebuild it from `node`/`npm` (law #57). No output is pasted because none was produced here. The output quoted below is the AUTHOR'S, from CASE.md §8 lines 110–120.
- In place of a run: a line-by-line hand trace of all eleven inputs through base, change and the CommonMark 0.31 §4.5 rules. **Every one of the eleven recorded lines matches the trace**, including the two non-obvious ones: `info-string-backtick reference=command:` (the opener is not a fence, so the paragraph is interrupted by the closing ```` ``` ````, which opens an empty fenced block that runs to the end of the document) and `indent-four base=command:` (base's `^ {0,3}` matches no line, so base finds no block and returns the empty command it uses for "no block").
- remark: a hand trace cannot detect a fault in the pinned `commonmark` 0.31.2 build, in `_isFenced` being the field it reads, or in the TypeScript 7.0.2 emit; only a run can.
- remark: the trace agreeing with §8 is what an honest record AND a carefully hand-written record would both show; it raises confidence in the code's logic, not in the claim that the command was executed.

### Differential

Question 1 — does the fixture reproduce the condition §2 names, and only there?

- Shipped eight (`change/suite.ts:20-61`): base returns `echo hello` for all eight (`base/extract.ts:29` closes on any fence-shaped line, the earlier round's defect); change refuses each with the class §2 names (`change/extract.ts:41-49`). Base and change differ here, as §2 requires.
- `well-formed` (`check/run.ts:13-15`): all three return `echo hello`. Agreement, as §2 requires.
- `info-string-backtick` (`check/run.ts:17-19`): change returns `echo hello` (`change/extract.ts:5` captures the rest of the opener and never inspects it); reference has no block containing the command. This is §2's first named opener defect.
- `indent-four` (`check/run.ts:21-23`): change returns `echo hello` because `change/extract.ts:16` trims every line before classifying it; reference refuses (indented code block). This is §2's second named opener defect.
- **A divergence §2 does not name:** `closer-too-long` and `closer-longer-still` are refused by change (`change/extract.ts:44`, `fence.length !== openLen`) while the reference accepts both (`reference=command:echo hello`, §8 lines 112–113). CommonMark §4.5 requires a closer **at least** as long as the opener, so two of the eight shipped "malformed" cases are well-formed CommonMark, and the change described in §1 as "strict, by the CommonMark rules" over-refuses them. §2 says only "a closing fence of the wrong length"; the builder chose the `!==` reading. The check prints the divergence honestly — it is not hidden — but neither CASE.md nor the answer key names it. See Conditions C2.
- `unclosed`: change refuses, reference accepts (CommonMark runs an unclosed fence to the end of the document). This one IS mandated by §2 ("refuses … an unclosed fence"), so it is a declared choice, not a fixture defect.

### Unplug

Questions 2 and 3 — false-world answers, reasoned from the code (not executed):

- If change rejected a backtick inside a backtick info string, the `info-string-backtick` line would read `change=refuse:…`, not `change=command:echo hello`. Distinguishable.
- If change did not strip leading white space, `indent-four` would read `change=command:` (no opener found). Distinguishable — the line is sensitive to exactly the `trim()` at `change/extract.ts:16`.
- If CommonMark read the info-string opener as a fence, the reference would read `command:echo hello`, not `command:`. Distinguishable.
- If any shipped refusal stopped firing, `check/run.ts:44-50` writes the failing case to stderr and exits 1 before any per-input line is printed — the run turns red rather than printing a quietly changed line.
- `well-formed` agreeing rules out "the change is simply broken".
- remark: the answer key says the shipped suite green "shows that the suite does not cover openers". `shipped-suite ok 8` reads identical in a world where the suite does cover openers; what shows non-coverage is the fixed opener at `change/suite.ts:3` (`OPEN = "````"`, used by every case through `fenced()`), together with the two opener lines. The evidence set supports the outcome; that one sentence of the key over-attributes it to the green line. The key is outside this diff's lane, so this is a remark, not a condition.
- No control was bent to green (question 3): no input in §2's list is missing, no expected line is asserted by the check (it prints, it does not compare), and the only assertion — the shipped suite — asserts what §2 says the suite asserts.

### Closure

Question 2 — does the recorded output establish `expected/EXPECTED.md`'s outcome ("established defect in change/: two malformed openers returned as commands")? **Yes**, on the author's recorded output and my trace of it: both opener lines show change returning a command where the reference does not hold that command in a fenced block; the well-formed line agrees; the suite is green. The opposite outcome ("change is correct") would require at least one of those lines to read differently, so the output does not also fit the opposite outcome.

One part of the key is weakened by the fixture: "what must NOT be claimed: that the shipped refusal classes do not work — they refuse every closing-fence case." They do fire on every case, but by the check's own reference two of those firings refuse valid CommonMark. An evaluated reviewer who reports "change over-refuses longer closers, contrary to CommonMark" would be correct and unscored by the key (C2).

### Remainder

- **No network (§2) is not held by `check.sh:5`.** `npm ci --ignore-scripts` fetches six tarballs from `registry.npmjs.org` and the recorded stderr (`audited 7 packages`, `found 0 vulnerabilities`) shows an audit round-trip to the registry. §2 opens with "The case must provide, with no network and no model". See C1.
- **Manifest shape change (`tests/corpus/manifest.json:76`).** PV-004's `status` changed from the object every one of the other 20 entries carries (`contract · fixture · checks · independent_check`, measured by `grep -n '"status"'`: 20 objects, 1 string) to the bare string `"built"`. It drops `independent_check: not done`, which CASE.md line 5 still records, and it makes the field two types across one array. No in-repo reader of the manifest was found (`grep -rln manifest.json tests/corpus` names only CASE.md prose pointers), so nothing breaks today. See C3.
- Fleet vocabulary in a public file: CASE.md line 79, "The `date` program is not in this seat's command class." Not a host, seat name, path or token — but it describes the author's private tooling inside a public corpus case. Owed tidy-up, not blocking (R6).
- `types/ambient.d.ts` hand-declares `process` instead of pinning `@types/node`; acceptable for a fixture and it keeps the dependency set small, noted for the record.

### Profile

Test-fixture code (a public corpus case): TypeScript, no I/O beyond stdout/stderr, one runtime dependency (`commonmark` 0.31.2, pinned exact, its three transitive deps locked) and one dev dependency (`typescript` 7.0.2, exact). All 25 `resolved` lines in the lock file point at `registry.npmjs.org` (measured: 0 lines not matching). No auth, persistence or network code in the fixture itself; the only network use is the install step in `check.sh`.

### Focus

1. Fixture reproduces §2: **holds** for both named opener defects and the well-formed agreement; one extra, undeclared change/reference divergence (longer closers).
2. Outcome follows from the evidence: **holds**, with the false-world answers above; one key sentence over-credits the green line (remark).
3. Case still proves something: **holds** — nothing bent to green, no input dropped.
4. Nothing private: **holds** — no host, seat name, private path, repository name, token or e-mail in the diff (grep over the diff for `@ /Users saphan cyberasis fleet seat token http ssh` and the owner's names: only public npm URLs and the one "seat's command class" phrase).
5. Lane: **holds** — every changed path is under `tests/corpus/cases/PV-004/` outside `expected/` (`git diff --stat`), plus PV-004's own manifest entry (lines 76, a 7-line hunk touching only that entry).

### Conditions

- **C1 (owed, contract):** `tests/corpus/cases/PV-004/check.sh:5` — make the check runnable with no network, as §2 requires: e.g. `npm ci --ignore-scripts --offline --no-audit` against a vendored/declared cache, or declare in §8/§9 that installation needs the registry once and the check itself does not. As it stands the §2 promise and the script disagree.
- **C2 (owed, case validity):** `tests/corpus/cases/PV-004/change/extract.ts:44` + `change/suite.ts:31-40` — the longer-closer refusals diverge from CommonMark and from the case's own reference. Either change the rule to `fence.length < openLen` and replace the two longer-closer cases with other closer-only malformations, or declare the divergence in CASE.md §9 (and, by its owner, in the answer key) so an evaluated reviewer who finds it is not scored against.
- **C3 (owed, record):** `tests/corpus/manifest.json:76` — keep the `status` object shape the other 20 entries use (e.g. `fixture: built 2026-10-21`, `checks: run 2026-10-21`, `independent_check: not done`), unless the master's order to the builder named the bare string, in which case say so in the gate note.
- **C4 (owed, verification):** the check has not been executed by any reviewer of this leg. A role whose class holds `bash`/`node`/`npm` runs `bash tests/corpus/cases/PV-004/check.sh` at `6fbc99639` and diffs its stdout against CASE.md §8 lines 110–120 before the gate.

### Unrun

- `bash tests/corpus/cases/PV-004/check.sh` — refused by this seat's permission class (`bash` not allowed); not reconstructed from other commands (law #57). Covered by C4.
- R3 mutations (remove `trim()` at `change/extract.ts:16`; add an info-string check; flip `!==` to `<` at line 44) — reasoned in Unplug, not executed, same cause.

## The review status as filed

## REVIEW — corpus-pv-004
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-004 fixture reproduces both opener defects §2 names (info-string backtick, four-space indent: change returns a command, reference does not), well-formed agrees, suite green; the outcome in expected/ follows. Hand-traced all 11 lines against §8 and they match; check.sh NOT run here (bash outside this seat's class) — C4.
  Owed: C1 check.sh:5 npm ci uses the network against §2's no-network rule; C2 change/extract.ts:44 refuses longer closers that CommonMark and the case's own reference accept, which the case does not declare; C3 manifest.json:76 status went from object to a bare string.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T00:00:00Z

## REDACTED

none
