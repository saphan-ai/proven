# REVIEW — corpus-pv-001 · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-21T19:03:34Z

## Summary

The duplicate fixture statically suggests a guard/reference mismatch, but the check omits the reference's actual first value required by the answer key. Independent execution is unrun because Node and npm are unavailable on this seat.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-001

### Ground
Reviewed `b0578227d...5b4f1dcbf` at `5b4f1dcbf`. The diff has ten changed files: nine under `tests/corpus/cases/PV-001/` outside `expected/` and one existing manifest entry. No other changed paths or changes to `expected/` were found. The fixture pins TypeScript 7.0.2 and parse5 8.0.1 in its npm lockfile. It uses the TypeScript compiler and no declared test runner or metric baseline. No private names, paths, credentials, or email addresses were found on static inspection. Licenses were not judged.

My attempted `bash tests/corpus/cases/PV-001/check.sh` exited 1, printing exactly:

```
pv-001: dependencies are not installed. Run: npm ci --prefix tests/corpus/cases/PV-001
```

`node --version` and `npm --version` both failed with `command not found`; the case's `node_modules/typescript/bin/tsc` is absent. The contract disallows network access; no install was attempted. The thirteen lines in `CASE.md` §8 are the author's recorded output, not this review's reproduced output.

### Differential
`tests/corpus/cases/PV-001/check.ts:20-43` statically lists all six named inputs with and without a real frame, plus the duplicate-attribute input. The base pattern at `base/guard.ts:5` matches textual `home`; the change at `change/guard.ts:148-153` overwrites the earlier value `other` with `home`. The independent parser at `reference/read.ts:8-24` reads the parsed tree, not either guard. The author's recorded lines say all six without-frame cases differ between base and change, all six with-frame cases agree, and the duplicate differs between change and reference. These are consistent with the source, but their runtime outcomes are UNRUN here. The base also passes the duplicate; that is not an additional hidden disagreement.

### Unplug
Runtime negative controls were not run: no Node/npm/toolchain exists on this seat. Removing later-wins `attrs.set` at `change/guard.ts:149` should change the duplicate result; removing the real frame at `check.ts:16-17` should turn a with-frame control red. Neither alarm has been observed. Static inspection is not a measurement of these controls.

### Closure
**BLOCKING — `tests/corpus/cases/PV-001/check.ts:48-53`: the check discards the reference's actual frame id before printing it.** Trigger: `<div data-screen="other" data-screen="home"></div>` at `check.ts:26`; the answer key requires a line showing `reference=first value`, but the code prints only `reference=fail missing=home`. The false-world answer is identical if the parser returns no frame or a different non-`home` value: both print `reference=fail missing=home`. Thus the output shows the required frame absent under the reference, not that the first value was kept. Remedy: print the reference frame ids or assert the exact `other` id independently, then rerun and record the full output. This defect is order-mandated insofar as §2 requested four results but not raw ids; the answer key requires the stronger proof.

**BLOCKING — `tests/corpus/cases/PV-001/CASE.md:77-113`: independent verification unavailable on this seat.** The attempted check exited before compilation; the author's thirteen lines cannot be endorsed as reproduced. Remedy: rerun the case on a seat with the pinned dependencies available without network access, compare every line with §8, and run the negative controls. This is an evidence limitation, not a claim that the author's check failed on its own seat.

### Remainder
The case statically retains its named inputs; the diff does not weaken an existing test or baseline. The check is project-owned and can agree with a wrong fixture; the separate parse5 reader is a stronger comparison, but its one-bit result is insufficient for the exact first-value claim. No external browser was run. No declared coverage, mutation, lint, or type-warning baseline/command was found for this case; no numeric metric was manufactured. No dependency install script was observed in the changed lockfile. Declined to judge licenses and browser behavior beyond the pinned parser.

### Profile
TypeScript/Node case; npm lockfile, compiler, parse5 reference reader. Static review only; runtime toolchain absent. No performance claims. No author-provided result was treated as independent evidence. Ruling: reject until the first-value observable and independent run are supplied; risk if wrong is delaying a valid corpus case rather than publishing an unproved first-value claim.

### Focus
The contract requires thirteen named inputs and base/change/reference verdicts on each. The source constructs them. The expected defect is plausible from `Map.set` overwrite and the reference parser, but the displayed reference verdict loses the distinguishing value. The author's output would also fit a parser that drops the attribute, even though it would not fit a reference that accepts `home`.

### Conditions
Expose or explicitly assert the reference's first duplicate value; independently rerun every input and a single-real-frame control on an equipped seat. Retain all six attack positions and both variants. Do not report acceptance while Unrun is non-empty.

### Unrun
`bash tests/corpus/cases/PV-001/check.sh` attempted, exited 1 before running checks; Node, npm and installed dependencies absent. Compiler, case outputs, actual parse5 behavior, runtime negative controls, and metrics on base and tip were not run. No network operation was attempted.

## The review status as filed

## REVIEW — corpus-pv-001
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The duplicate fixture statically suggests a guard/reference mismatch, but the check omits the reference's actual first value required by the answer key. Independent execution is unrun because Node and npm are unavailable on this seat.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-21T19:00:21Z

## REDACTED

none
