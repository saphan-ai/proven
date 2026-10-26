# REVIEW — corpus-pv-004 · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:45:38Z

## Summary

The path correction is sound by inspection; npm is absent here, so the check remains unrun on this seat and needs a capable-seat reproduction.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-004

### Ground
`git rev-parse --short=7 HEAD` printed `04c3024`. `git diff 1f794ed..04c3024` changes only `test/corpus/cases/PV-004/CASE.md:90`: `tests/corpus/cases/PV-004/check.sh` becomes `test/corpus/cases/PV-004/check.sh`. `git diff --name-only 9303b2d...04c3024` lists only the PV-004 case and `test/corpus/manifest.json`. `git diff --check 9303b2d...04c3024` exited zero. The command's new path exists (`test/corpus/cases/PV-004/check.sh:1`). The case has no matches for `tests/`, `fleet/`, or `conformance/`. No apparent private paths or credentials appeared in the case's tracked files inspected.

### Differential
`test/corpus/cases/PV-004/change/suite.ts:20` lists eight cases with one opener and only missing or varied closers; `change/extract.ts:16` trims each line and `change/extract.ts:21` accepts an opener without checking its info string. The two extra inputs in `check/run.ts:16` and `check/run.ts:20` target this uncovered class. `check/expected.stdout:10` and `check/expected.stdout:11` record that change returns a command where the reference returns, respectively, an empty command from a later fence and `refuse:no-fenced-block`; `check/expected.stdout:9` records well-formed agreement. `expected/EXPECTED.md` was read last: it requires these contrasts and the eight-case suite result. These are static observations of recorded output, not a fresh runtime reproduction.

### Unplug
`check.sh:7` installs pinned dependencies, `check.sh:8` compiles, `check.sh:9` runs the suite and probes, and `check.sh:10` compares stdout byte-for-byte with `check/expected.stdout`. If the output differs, `diff -u` exits nonzero; if the suite fails, `run.ts:25` exits nonzero before the compared lines. Those are the instrument's false-world answers, assuming the compiled check executes. The observed `npm: command not found` happens earlier and cannot distinguish a defective extractor from a sound one. Matching checked-in golden bytes alone could fit an opposite runtime outcome if the check were not executed or the reference were wrong; source inspection supports the intended contrast but does not replace execution.

### Closure
No blocking defect found in the path correction or inspected case. The prior condition that the documented command name an existing path is ADDRESSED at `CASE.md:90`. The master-provided output has the 11 expected stdout lines and `shipped-suite ok 8`, apart from npm install notices and timing; that is a claim, not my measurement. Since `Unrun` is nonempty, the verdict is `accept-with-conditions`, not `accept`.

### Remainder
OWED: reproduce `bash test/corpus/cases/PV-004/check.sh` on a seat with npm and Node; compare actual stdout and exit with `CASE.md:93-120` before treating the runtime claim as independently checked. No code fix requested. Licence decisions and behaviours beyond this case diff are declined as outside this review's scope.

### Profile
Docs and test corpus. The check is a project-owned golden-output comparison, which is more circular than the pinned CommonMark parser it invokes; its result was not independently executed here. No declared quality metric or baseline is supplied in this order to remeasure; this path-only revision does not lower a test threshold.

### Focus
The master's path fix, case-vs-expected contrast, path existence, lane confinement, and private-content scan. `expected/` was read last as directed. A static pattern scan and git diff cannot establish absence of every possible private datum, only absence of inspected path patterns and apparent secrets.

### Conditions
OWED (`test/corpus/cases/PV-004/check.sh:7`): rerun the check where npm and Node can execute and record stdout, stderr, and exit code. The master must resolve the unrun runtime check before the gate; the path correction itself is not blocked by an observed defect.

### Unrun
`npm_config_offline=true bash test/corpus/cases/PV-004/check.sh` exited 127 here with `test/corpus/cases/PV-004/check.sh: line 7: npm: command not found`. Offline mode kept the attempted check within the no-network boundary. No Node/TypeScript execution, npm install, compiled reference comparison, or independent runtime reproduction occurred here. The master's run is cited as a claim only.

## The review status as filed

## REVIEW — corpus-pv-004
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The path correction is sound by inspection; npm is absent here, so the check remains unrun on this seat and needs a capable-seat reproduction.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T21:42:11Z

## REDACTED

none
