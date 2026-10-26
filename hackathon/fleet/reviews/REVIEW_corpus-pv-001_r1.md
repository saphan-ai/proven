# REVIEW — corpus-pv-001 · round 1

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 2 NOTE
- reviewed: see the notes below
- updated: 2026-10-21T18:52:33Z

## Summary

PV-001 fixture matches §2 by static trace (check.sh refused to this seat: bash/node/npm not in the review class; UNRUN). Only the duplicate line splits change from the reference.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-001

Reviewer: claude-code, model claude-opus-5-5. Scope: `b0578227d...5b4f1dcbf` (1 commit, 10 files, +792 −5).
`git rev-parse --short=9 HEAD` printed `5b4f1dcbf`. Verdict: **accept-with-conditions**.

### Ground

- HEAD measured at `5b4f1dcbf`; three-dot diff read in full, except the lockfile body, which was read by grep (R4 below).
- Read: `CASE.md` (whole file, at tip), `check.sh`, `check.ts`, `base/guard.ts`, `change/guard.ts`, `reference/read.ts`,
  `package.json`, `tsconfig.json`, the `manifest.json` hunk, and `expected/EXPECTED.md` (only after judging items 1 and 3
  from the code, as the order asks).
- **The check was NOT run by this reviewer.** `bash tests/corpus/cases/PV-001/check.sh` was refused by the hook:
  `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████ ... allowed form: none`.
  `node` and `npm` are not in that list either, and `node_modules/` is not present in the worktree. Per R15 and law #57 I
  did not rebuild the run from other commands. Item 1 below is therefore a **static trace** of each input through the
  three readers, not a measurement. The author's recorded output in CASE.md §8 is set aside as evidence (R1); I compare
  my trace with it only as a cross-check.

Static trace, one line per input (my reading of the code, not program output):

| input | base | change | reference (parse5 8.0.1, WHATWG rules) | matches §8 line |
|---|---|---|---|---|
| comment ± frame | pass / pass | fail / pass (comment skipped, `change/guard.ts:44-47`) | fail / pass | yes |
| script ± frame | pass / pass | fail / pass (raw text skipped, `:155-157`) | fail / pass | yes |
| style ± frame | pass / pass | fail / pass (same) | fail / pass | yes |
| attribute-value ± frame | pass / pass | fail / pass (`alt` value is quoted, `:131-136`) | fail / pass | yes |
| cdata ± frame | pass / pass | fail / pass (`<!` branch `:49-52` skips to the first `>`, which ends `<div data-screen="home">`) | fail / pass (CDATA in HTML content is a bogus comment ending at the same `>`) | yes |
| split-line ± frame | pass / pass | fail / pass (text inside `<p>`) | fail / pass | yes |
| duplicate-attribute | pass (regex sees both values) | pass (`attrs.set` at `:149`, later wins → `home`) | fail (tokenizer drops the second attribute → `other`) | yes |

My trace agrees with all 13 lines recorded in CASE.md §8.

### Differential

**Item 1 — the fixture reproduces the condition §2 names (static, not run).** Base differs from change on the 6 named
positions without a frame (base pass, change fail) and on no line with a real frame. Change differs from the reference on
exactly one line, `duplicate-attribute` (change pass, reference fail), which is where §2 puts the disagreement
(§2 bullets "change ... later duplicate overwrites", and the last required input). Change and reference agree on all
12 other lines. Every input §2 requires is present (`check.ts:29-43`): 6 positions × with/without frame, plus the
duplicate with a non-required first value and a required second value (`check.ts:27`). The reference does not call
either guard (`reference/read.ts` imports only parse5). §2 is met.

**Item 2 — the expected outcome follows from the evidence, with one gap.** `expected/EXPECTED.md` says: established
defect in `change/`; evidence = the duplicate line with change = pass and reference = first value, and a single-frame
document where all three agree.

- The outcome class is discriminated. False-world answer: if change kept the first duplicate (no defect), the
  duplicate line would read `change=fail missing=home`, matching the reference. The recorded output is not that, so
  the line separates "defect" from "no defect".
- The "all three agree on a real frame" control is present: the 6 `*-with-frame` lines.
- **Gap (OWED, O1):** EXPECTED asks for "reference = first value". The check prints only pass/fail against `home`
  (`check.ts:50-54`), never the ids the reference reads. `reference=fail missing=home` is the same sentence in a world
  where parse5 kept `other` (first value), and in a world where it kept no `data-screen` at all on that element. The
  same holds on the change side: `change=pass` is printed both by a last-wins map and by a guard that counts every
  duplicate value. So the line proves "change passes where the spec reader has no `home` frame" — enough for the
  outcome class — but it does not show the mechanism EXPECTED names ("reads the first value" / "later wins").
  Remedy: print operation 1 (`listFrameIds`) for base, change and reference on each line, or add the reverse-order
  input `data-screen="home" data-screen="other"` (last-wins change → fail, reference → pass; a count-all guard → pass).
- NOTE: EXPECTED names the function `checkFrames`; the fixture exports `judge` (`base/guard.ts:20`,
  `change/guard.ts:163`). Not a defect of this diff (`expected/` is outside its lane); the master decides which side
  changes.

**Item 5 — lane.** `git diff --stat` lists only `tests/corpus/cases/PV-001/**` outside `expected/`, and the PV-001
entry of `tests/corpus/manifest.json` (2 lines, lines 26-27 of the tip file). `expected/` is untouched. Met.

**Item 4 — nothing private.** Grep of the full diff for home paths, user and host names, the private store and data
host names, e-mail domains, IPv4 addresses, `token`, `secret`: no hit. The lockfile resolves only to
`https://registry.npmjs.org/`. §8 records tool versions and one timestamp, no path outside the repository. Met.

### Unplug

Static, because the check could not be run.

- Mutate `change/guard.ts:149` to keep the first duplicate (`if (!attrs.has(attrName)) attrs.set(...)`): the
  duplicate line turns to `change=fail missing=home`; the case loses its finding. The check sees this mutant.
- Remove `REAL_FRAME` from `withRealFrame` (`check.ts:16-18`): all 6 `*-with-frame` lines turn change/reference to
  fail. The check sees it.
- Make base element-aware: the 12 positional lines change. The check sees it.
- **No mutant turns the check red by itself.** `check.ts` only prints; `check.sh` exits 0 whatever the lines say
  (`check.sh:15-16`). A drift of any line is caught only by a human comparing against CASE.md §8 (law #41: nothing
  turns red when the case stops proving its point). OWED, O2.
- The reference walker's `break` (`reference/read.ts:15`) cannot be unplugged meaningfully: parse5 already removes a
  duplicate attribute in the tokenizer, so the element carries one `data-screen`. Harmless, but it means the reference
  side's "first value" rests on parse5, not on this file — which CASE.md §9 states.

### Closure

Inputs are fixed strings in `check.ts`; no network, no model, no file read at run time. The reference parser is
pinned exactly (`parse5` `8.0.1`, `entities` `8.1.0`, `typescript` `7.0.2`, no ranges). Lockfile delta (R4): 3 direct
or transitive packages plus 20 platform-specific `@typescript/typescript-*` optional packages; no `hasInstallScript`
and no `scripts` entry in the diff. No runtime pin (`engines`, `.nvmrc`) is declared; §8 records `node v24.5.0` as
measured, not as required. NOTE.

### Remainder

**BLOCKING-class, B1 — the review target carries the answer key.** CASE.md §5: the reviewer under evaluation receives
"the two versions of the guard and a readable diff between them ... never the check, the reference or `expected/`".
But `change/guard.ts:1-3` says *"A later duplicate attribute overwrites an earlier one. The HTML standard keeps the
first duplicate; this map does not."* and `change/guard.ts:148` says *"Later duplicate wins. This is the property the
case exists to show."* That is the content of `expected/EXPECTED.md`, inside the file §5 hands over. A reviewer
evaluated on this case is told the finding. The fixture still reproduces the condition, but any evaluation run on it
measures reading, not reviewing — a silently trusted result (R6). `base/guard.ts:1-3` likewise lists the positions
base fails on; that matches §1's description and is less harmful, but it is the same kind of guidance. Remedy: remove
the answer-bearing comments from both guards (keep neutral descriptions such as "tokenize tags and fold attributes
into a map"); the explanation already lives in CASE.md §2/§4 and EXPECTED.md.

OWED:
- **O1** — print the ids each reader keeps (operation 1), or add the reverse-order duplicate input (Differential).
- **O2** — make the check fail on drift: `check.sh` compares the printed lines with a committed golden file of the
  §8 output (inside the case directory, not in `expected/`) and exits non-zero on a difference.
- **O3** — `tests/corpus/manifest.json:26` drops the prior-art declaration: before `"not built (prior art, built before
  the window)"`, after `"built"`. CASE.md:5 still says "declared prior art", and the code is dated 2026-10-21, before
  the 22 October window in AGENTS.md §3. Restore the declaration in the manifest so the two agree.
- **O4** — running the documented commands creates `tests/corpus/cases/PV-001/node_modules/` and `dist/`; the root
  `.gitignore` covers neither and the case has none. A later `git add` picks up build output. Remedy: a case-local
  `.gitignore` with both.

NOTE:
- **N1** — CASE.md:69 "Newly reconstructed evidence: none yet; it is built with the case code." is stale now that §8
  holds the output.
- **N2** — `checkFrames` vs `judge` naming (Differential).
- **N3** — split-line is read as a text split (`<p>data-screen=\n"home"</p>`). §2 says only "split across lines"; a
  split inside a real tag would be a real frame. The author's reading is defensible; stated so the master can rule.

Item 3 — the case still proves something: no check was bent. No input is dropped, no line forced; the duplicate
line shows change and reference disagreeing, which is the failure the contract asks for, not turned green. B1 is a
contamination of the review target, not a bent check.

### Profile

No security profile: the diff touches no authentication, payment, data access or destructive path (R9). Instruments:
`lang-typescript-node` (toolchain, lockfile, ESM/`Node16` config read — `type: module` and `module: Node16` agree;
type check not run) and `testing-bar` (no metric is declared for the corpus; NOTE per its "Do not use when"). Check
ranking: the reference is an *external* check (a published parser of a published standard); the guards and the
printer are *suite*.

### Focus

1. Fixture reproduces §2 — met, by static trace (not run).
2. Expected outcome follows — outcome class met; mechanism ("first value") not shown by the output (O1).
3. Case still proves something — no bent check; answer leaked into the review target (B1).
4. Nothing private — met.
5. Lane — met.

Declined to judge (R11): the 22 October code window itself (owner's rule; only the record's consistency, O3, is
judged); licences (order boundary); whether TypeScript 7.0.2 accepts this `tsconfig.json` (not run).

### Conditions

1. **B1 (blocking before the case is used for evaluation):** strip the answer-bearing comments from
   `change/guard.ts:1-3` and `:148`, and the position list from `base/guard.ts:1-3`.
2. **Before the gate:** the master runs `npm ci --prefix tests/corpus/cases/PV-001` and
   `bash tests/corpus/cases/PV-001/check.sh` at the tip and confirms the 13 lines equal CASE.md §8 (this reviewer
   could not run it).
3. **O1–O4:** owed, may land in this round or the next; they do not reopen this review (R6).

### Unrun

- `bash tests/corpus/cases/PV-001/check.sh` — refused by hook: `the program "bash" is NOT in the ██████████████ list
  of role review under permission class ██████████████████████████ ... allowed form: none`.
- `npm ci --prefix tests/corpus/cases/PV-001` — `npm` is not in the same allow list; not attempted after the refusal
  above; `node_modules/` is absent.
- `tsc --version`, type check of the case — `node` is not in the allow list.
- Mutations in Unplug — traced by reading, not executed.

## The review status as filed

## REVIEW — corpus-pv-001
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-001 fixture matches §2 by static trace (check.sh refused to this seat: bash/node/npm not in the review class; UNRUN). Only the duplicate line splits change from the reference.
  BLOCKING B1: change/guard.ts:1-3,148 comments state the answer key inside the review target §5 hands over; strip them before any evaluation.
  OWED: O1 output cannot show "first value" (print ids or add reverse-order input); O2 check never exits red; O3 manifest dropped the prior-art note; O4 no ignore for node_modules/dist.
  Master must run check.sh at 5b4f1dcbf and confirm the 13 §8 lines before the gate.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-21T19:20:00Z

## REDACTED

| what | times |
|---|---|
| policy binding name | 4 |
