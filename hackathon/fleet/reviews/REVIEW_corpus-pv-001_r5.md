# REVIEW — corpus-pv-001 · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T05:19:29Z

## Summary

Round-three source addresses the named conditions; independent execution is unrun because Node and npm are absent on this seat. Re-run the case on a Node seat before treating its recorded output as independently reproduced.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-001

### Ground
Reviewed `9303b2d...f23a1a4` and the focused round-three diff `f23a1a4^..f23a1a4`; `git rev-parse --short=7 HEAD` printed `f23a1a4`. The latter changes five files, all under `test/corpus/cases/PV-001/`, and no `expected/` file. The larger diff also changes the corpus manifest line. `git diff --check` returned zero for both ranges. The actual round-two review documents were not supplied; the kickoff's account of their conditions is a claim, not independent evidence of the full condition set.

### Differential
At `test/corpus/cases/PV-001/change/guard.ts:144`, the tag attribute map still overwrites a first value with a later duplicate. The caller at `:147` reads that map, and `judge` checks for required `home`. `test/corpus/cases/PV-001/reference/read.ts:13` now collects every attribute supplied by the pinned parser instead of stopping at the first. The `duplicate-attribute` input at `test/corpus/cases/PV-001/check.ts:31` sets `other` first and `home` second. From source, the guard selects `home`; the reference is expected to select `other`. The check also includes six named positions, each with and without a real frame, at `check.ts:33-47`. The author-recorded thirteen output lines in `CASE.md` §8 and in the kickoff agree with that reading, but they are not this reviewer's execution.

### Unplug
False-world answer: if the parser retained the last duplicate, `reference_ids` would read `["home"]` and the duplicate line would no longer separate the implementations; if the real-frame case failed, its paired `change=pass reference=pass` line would not hold. The output contract can distinguish those worlds, but no executable result was obtained on this seat. Mutating the duplicate-map assignment or removing the reference reading was not run; no mutation score or negative-control success is claimed. A static reading alone cannot prove the pinned dependency actually produces these rows.

### Closure
`base/guard.ts:1` and `change/guard.ts:1,141` remove answer-revealing commentary without changing executable statements. `check.ts:6-8,58-68` prints all three frame-ID arrays; `reference/read.ts:11-15` removes the early break; `CASE.md:79-126` records a new dated install and check transcript. These address the three changes the kickoff attributes to round two. No dependency manifest, lockfile, or expected answer changes in round three. No path in the focused diff names an unavailable private location; the case's own relative paths exist. The `CASE.md` cache path is a relative artifact path, not a secret. No source code was changed by this review.

### Remainder
OWED — independently run the pinned install and check on a seat with Node/npm, compare all thirteen output rows byte-for-byte with `CASE.md` §8, and reconcile any difference before treating the reproduction as measured by review. OWED — if the original round-two conditions include more than the three described in the kickoff, check those against the focused diff. Neither item demonstrates that the case is wrong; neither is reported BLOCKING. Declined to judge: the historical independent reviewers' underlying records, the original browser implementation, and licences (excluded by the order). The answer-key method name `checkFrames` at `expected/EXPECTED.md:4` differs from the fixture's exported `judge`; the described duplicate-attribute property still matches the source. This naming drift is an OWED editorial correction, not a new round-three defect.

### Profile
Review profile: test fixture and documentation; no security, runtime service, or UI surface changed in this focused diff. Check ranking: `check.sh` is a project script wrapping the fixture's own suite (most circular); TypeScript's pinned compiler and the independent parse5 dependency would strengthen a successful run, but remain unrun here. No numeric testing-bar baseline, command, or tolerance is declared for this case; no base/tip metric delta is asserted. No bar-weakening move appears in the focused diff; no exception was added or used.

### Focus
FILE · LINE · CONTROL: `check.ts:55-69` is the diagnostic control; it exposes IDs but does not assert them. `change/guard.ts:144` retains the deliberate last-wins property, and `reference/read.ts:11-15` preserves the separate parser's attribute traversal. `CASE.md:104-125` is a recorded output, not this reviewer's fresh proof. The result supports `accept-with-conditions`, rather than `accept`, because execution is unrun.

### Conditions
OWED: independently reproduce §8 with the pinned dependencies on a Node seat. OWED: compare against the original round-two review records if those are available to the master. OWED editorial: align the method name in the answer key with the fixture. No BLOCKING condition established by this review.

### Unrun
`command -v node` and `command -v npm` found neither executable; `node_modules` and `dist` are absent. `bash test/corpus/cases/PV-001/check.sh` exited 1 and printed `pv-001: dependencies are not installed. Run: npm ci --prefix test/corpus/cases/PV-001`. Consequently `npm ci`, the TypeScript compile, the thirteen-row executable check, base/tip test metrics, and a live mutation/unplug run were not performed. The master's reported successful run is explicitly second-hand evidence, not a substitute for these runs.

## The review status as filed

## REVIEW — corpus-pv-001
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Round-three source addresses the named conditions; independent execution is unrun because Node and npm are absent on this seat. Re-run the case on a Node seat before treating its recorded output as independently reproduced.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T05:13:37Z

## REDACTED

none
