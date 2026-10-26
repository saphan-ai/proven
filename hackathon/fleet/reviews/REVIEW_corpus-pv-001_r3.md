# REVIEW — corpus-pv-001 · round 3

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T17:51:50Z

## Summary

PV-001 reproduces the duplicate-attribute disagreement at cec11bd. Condition: show the reference frame id in the duplicate check line. Generated files remain because cleanup was refused; macOS was not run.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-001

### Ground

Reviewed SHA `cec11bdc1e3a7401dbbb1af704456f19879c0718` (`git rev-parse HEAD`); the required short form printed `cec11bd`. Scope: `git diff fe146e0...cec11bd`. I read `tests/corpus/cases/PV-001/expected/EXPECTED.md` only after running the check. Verdict: accept-with-conditions.

The first `bash tests/corpus/cases/PV-001/check.sh` exited 1 because dependencies were absent. `npm ci --prefix tests/corpus/cases/PV-001` then exited 243 because the default npm cache was not writable. Repeating with a cache under `████████████████████████████████████████` exited 0. The check then exited 0 on Linux with Node v22.23.2 and npm 10.9.8. The case pins TypeScript 7.0.2 and parse5 8.0.1. The author's recorded run used different Node and npm versions; I did not rely on it.

### Differential

`bash tests/corpus/cases/PV-001/check.sh` printed exactly:

```text
comment-without-frame base=pass change=fail missing=home reference=fail missing=home
comment-with-frame base=pass change=pass reference=pass
script-without-frame base=pass change=fail missing=home reference=fail missing=home
script-with-frame base=pass change=pass reference=pass
style-without-frame base=pass change=fail missing=home reference=fail missing=home
style-with-frame base=pass change=pass reference=pass
attribute-value-without-frame base=pass change=fail missing=home reference=fail missing=home
attribute-value-with-frame base=pass change=pass reference=pass
cdata-without-frame base=pass change=fail missing=home reference=fail missing=home
cdata-with-frame base=pass change=pass reference=pass
split-line-without-frame base=pass change=fail missing=home reference=fail missing=home
split-line-with-frame base=pass change=pass reference=pass
duplicate-attribute base=pass change=pass reference=fail missing=home
```

The six named positions differ between base and change only when no real frame exists. Each paired real-frame input passes in all three readings. The duplicate case is the intended further disagreement: both guards pass, while the reference fails. In `change/guard.ts:149`, `Map.set` overwrites the first attribute; `check.ts:26` puts `other` first and required `home` second. A separate direct call to the compiled `listFrameIds` operations printed `change=["home"] reference=["other"]`. The reference reader uses parse5 rather than either guard (`reference/read.ts:3-29`).

### Unplug

The false-world answer for a duplicate element whose browser reading contains required `home` would be `reference=pass`; the observed `reference=fail missing=home` distinguishes that world. It does not by itself distinguish first-value retention from discarding both duplicate values. The direct frame-id probe above distinguishes those: change returns `home`, reference returns `other`. A real frame paired with every named position is the negative control for a guard that simply rejects everything. The check has no assertion and would still exit 0 if its printed results changed; its observable is the printed line, not its exit status. This is permitted by §2's print-only contract, but limits use as an automated gate.

### Closure

OWED — `tests/corpus/cases/PV-001/check.ts:49-63` and `expected/EXPECTED.md:5`: the answer key requires the duplicate check line to show the reference's first value, but the line prints only `reference=fail missing=home`. Expose the reference frame id in that line, or align the answer key's evidence requirement with the separate frame-id probe. The observed pass/fail disagreement and the direct probe support the established-defect outcome. This is an evidence-format gap, not a contrary result.

### Remainder

`git diff --name-status fe146e0...cec11bd` lists only nine files in `tests/corpus/cases/PV-001/` outside `expected/`, plus `tests/corpus/manifest.json`. The manifest diff changes only PV-001's fixture/check status. `git diff --check fe146e0...cec11bd` exited 0. A targeted scan of changed text found no private host, user path, token or email. Lockfile URLs point to the public npm registry. No code was changed in this review. Cleanup of generated `dist/`, `node_modules/`, and the review's npm cache was refused by the command tool: `rm -f style commands are not permitted. Use a safer approach`. I did not retry the denied act in another spelling. The worktree therefore is not back to its initial untracked-file set; these three artifacts need authorized cleanup.

### Profile

The case is a TypeScript fixture and documentation review. `check.sh` uses Bash, `dirname`, `pwd`, `node`, and a locally installed TypeScript compiler; it contains no Linux-specific option. The lockfile includes macOS TypeScript packages for arm64 and x64. This supports portability by inspection; no macOS execution was available. The dependency delta adds pinned parse5 8.0.1, entities 8.1.0 and TypeScript 7.0.2, with platform-specific optional compiler packages. No changed bar, skip, suppression, or test exclusion was found. No numeric coverage, mutation, lint or type-warning bar was declared for this case, so no numeric bar was measured; the check is a suite-level instrument and its output can be changed by the case author.

### Focus

The fixture supplies both required guard operations (`base/guard.ts:8-24`, `change/guard.ts:34-167`). `check.ts:28-42` has all six positions with and without a real frame, plus the duplicate. The duplicate's first value is not required and its second is required. The check exercises base, change and independent reference on each input. The new check was not bent to make the intended duplicate line appear: it prints a disagreement even though it exits 0. I found no other changed behavior or extra files in the diff. I declined to judge browser runtime behavior, original private case history, and licensing, which are outside this order.

### Conditions

Resolve the OWED evidence-format gap in Closure before treating the printed check line as proof of the *first* duplicate value. The master can use the measured direct probe as evidence if the answer key is aligned. No condition changes the established-defect finding.

### Unrun

macOS execution: not run; this seat is Linux. A live browser: not run; §2 specifies a parser reference and no browser. Full Go suite and project-wide quality metrics: not run; the reviewed diff is isolated TypeScript corpus material, and no case metric command or baseline is declared.

## The review status as filed

## REVIEW — corpus-pv-001
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-001 reproduces the duplicate-attribute disagreement at cec11bd. Condition: show the reference frame id in the duplicate check line. Generated files remain because cleanup was refused; macOS was not run.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T17:24:36Z

<!-- ENGINE-EMITTED REVIEW SKELETON, rendered from ██████████████████████████████.
     Everything below this line contains WORKED FORMS to copy as applicable.
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy the applicable form OUT.

     Write-legal `- state:` tokens: spawned · recommendation-ready.
     Last act before done: set `- state:` above to recommendation-ready and add every
     return-time key shown below. Their names, value classes, and vocabularies
     are projected from the review family's rows, never copied here.

     `verdict`: exactly one write-legal verdict token, alone on the line — mandatory for `recommendation-ready`; no parenthetical and no conditions after it; the conditions belong in the review body the `- artifacts:` manifest names
     - verdict: <accept | accept-with-conditions | reject>

     `ref`: the review's returned artifact — mandatory for `recommendation-ready`; one of this document's own `- artifacts:` entries, VERBATIM (a basename that is not a declared entry is refused at delivery); never a git ref
     - ref: _REVIEW_NOTES.md

     `artifacts`: NESTED DASH LIST of paths the review returns; mandatory for `recommendation-ready`, including an explicit empty list when nothing is returned
     - artifacts:
       - _REVIEW_NOTES.md
       - _REVIEW_STATUS.md

     An entry names a FILE, or a DIRECTORY: a directory is delivered as ONE
     deterministic `.tar` archive (sorted entries, fixed mtimes — the same
     bytes, and so the same digest, every time) recorded under the entry's
     own name plus that suffix, which is the name the projected manifest and
     the teardown gate then read. A path that EXISTS and is neither is
     refused by what it IS, never as `missing`.

     Explicit empty manifest form when the review returns no files:
     - artifacts:

     🔴 THE ANALYSIS DOES NOT GO IN THIS FILE. It goes in _REVIEW_NOTES.md, whose
     mandatory members are H2 HEADINGS — never top-level dash bullets in
     _REVIEW_STATUS.md. A section heading written as `- ground:` here is an unknown
     key: it is now reported as a WARNING rather than refused, so your
     verdict still lands — but the analysis lands nowhere at all.

     The complete _REVIEW_NOTES.md form, projected from the same one section source
     this engine renders it from (██████████████████████████████):

     # REVIEW NOTES — corpus-pv-001

     ## Ground
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Differential
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unplug
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Closure
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Remainder
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Profile
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Focus
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Conditions
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unrun
     <replace with measured review content; write `none` only when none is the measured answer>
-->

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 2 |
