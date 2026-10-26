# REVIEW — corpus-pv-004 · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 3 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:19:35Z

## Summary

Static evidence supports the opener-defect outcome; independent execution stopped at missing npm. Rerun with the pinned toolchain before calling the case verified.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-004

### Ground

Reviewed `9303b2d...1f794ed`; `git rev-parse --short=7 HEAD` printed `1f794ed`. I read `test/corpus/cases/PV-004/CASE.md` §2 before `expected/EXPECTED.md`. The two commits are `8d9eddd` and `1f794ed`. The master's earlier exit-zero claim is not an independent measurement.

`bash test/corpus/cases/PV-004/check.sh` exited **127** with precisely this output:

```
test/corpus/cases/PV-004/check.sh: line 7: npm: command not found
```

There was no check stdout. The script stopped before installation, compilation, the shipped suite, the reference parser and golden comparison. The empty cache directories it created were removed. The checked-in 11 lines below are **expected data, not output from this run**.

### Differential

`test/corpus/cases/PV-004/check/run.ts:15` supplies a well-formed document, an opener whose info string contains a backtick, and a four-space-indented opener. `test/corpus/cases/PV-004/change/extract.ts:16` trims indentation; lines 21–26 accept an opener without checking its info string. Both malformed openers thus reach the command branch at line 47. The base at `test/corpus/cases/PV-004/base/extract.ts:5` matches only zero to three spaces, never refuses, and returns an empty command when it finds no opener. `test/corpus/cases/PV-004/check/reference.ts:14` takes fenced blocks from the pinned CommonMark parser. The checked-in expectations are:

```
well-formed base=command:echo hello change=command:echo hello reference=command:echo hello
info-string-backtick base=command:echo hello change=command:echo hello reference=command:
indent-four base=command: change=command:echo hello reference=refuse:no-fenced-block
```

For the info-string case, the final three-backtick line is itself a *new empty fenced block* in the reference: `reference=command:` does not mean the invalid first opener was accepted. The false-world answer if that first opener were accepted is `reference=command:echo hello`; if change refused it, change's field would be `refuse:…`. If the four-space opener were accepted as a fenced block, the reference would instead return `reference=command:echo hello`. These predicted answers are based on source and checked-in expectations, not an execution of this parser on this seat.

The eight cases in `test/corpus/cases/PV-004/change/suite.ts:17` vary only the closing fence or omit it; every opening fence is valid. A green suite cannot establish correct handling of malformed *openers*. On closing-fence cases base returns a command while change refuses; on the indented opener base returns empty while change returns `echo hello`, consistent with §2's specified base behaviour. The well-formed control has all three returning the command.

### Unplug

The round-two guard `test/corpus/cases/PV-004/check.sh:10` compares all output to `check/expected.stdout` with `diff -u`. I ran a harmless negative control: `printf 'wrong\n' | diff -u test/corpus/cases/PV-004/check/expected.stdout - >/dev/null`; exit **1**. This establishes that `diff` can reject wrong data, **not** that the script reaches it, or that the golden file is independent of the author. Altering runner and golden together could pass. A stronger opener-mutation run is unrun because `npm` is missing.

### Closure

Verdict **accept-with-conditions**, not `accept`: independent execution did not complete. The source supports the answer key's established-defect outcome, but this review cannot claim the expected lines were observed. `git diff --name-only 9303b2d...1f794ed` contained only `test/corpus/cases/PV-004/` outside `expected/` and the PV-004 line of `test/corpus/manifest.json`; `git diff --check 9303b2d...1f794ed` exited **0**. Round two added golden lines and a byte comparison, clarified why the reference result can be empty, and restored the manifest status object. The original round-one review conditions were not provided verbatim, so their complete discharge cannot be certified.

### Remainder

- **OWED · independent check:** `test/corpus/cases/PV-004/check.sh:7` fails at missing `npm` here (exit 127). Run on an equipped seat; capture both streams and exit status; verify all 11 lines, especially the two malformed openers and well-formed control. Do not mark an independent check complete from the golden file.
- **OWED · portability/offline:** Test on macOS and Linux. `test/corpus/cases/PV-004/check.sh:3` uses `dirname --`, an option outside the portable POSIX `dirname` form; verify the macOS implementation or remove that option. A cold `npm ci` can access the npm registry, contrary to §2's no-network expectation unless dependencies are preseeded. Neither offline installation nor macOS execution was measured here.
- **OWED · round-one conditions:** Obtain both earlier review texts and check each condition against the round-two diff. The order asserts their satisfaction but supplies no exact conditions.
- No private host, personal path, credential or email was found in the changed case text outside public npm registry URLs. Licences and unrelated cases were outside scope. Ruling: missing npm is an unrun measurement, not a demonstrated product defect; calling it a failed case would be unjustified.

### Profile

Test-corpus fixture. Direct versions `commonmark` 0.31.2 and `typescript` 7.0.2; locked transitives `entities` 3.0.1, `mdurl` 1.0.1 and `minimist` 1.2.8, plus platform-optional TypeScript binaries. `npm ci --ignore-scripts` disallows install scripts; no project install script is declared. Independence ranking: CommonMark rules and parser (external), golden byte comparison and compiler (project), eight-case shipped suite (suite, circular for openers). No numerical baseline/metric was declared; none invented or reported as met. Go skill read; this diff is TypeScript-only, and no Go tests are claimed.

### Focus

The defect class is an opener accepted where CommonMark does not find the intended fenced command. After consulting `test/corpus/cases/PV-004/expected/EXPECTED.md`, its outcome agrees with the static trace. The well-formed control checks that the changed extractor is not universally broken; the shipped suite's closers do not prove opener correctness. No product code or answer key was changed.

### Conditions

1. Execute `bash test/corpus/cases/PV-004/check.sh` independently with the pinned toolchain, retain exit status and both streams, and compare all 11 measured lines to §2 rather than changing the golden file to fit the result.
2. Retrieve the exact round-one verdict conditions and report each as addressed or not addressed from the round-two diff.
3. Verify the check on macOS and ensure a judge with a cold checkout can run it without network, or explicitly narrow the no-network claim.

### Unrun

The script's install, compilation, suite, reference and golden comparison did not run: `npm: command not found` (exit 127). macOS was not available on this Linux seat. Opener-guard mutation and base/tip metrics were not run. Exact round-one conditions were not supplied; their individual closure remains unverified.

## The review status as filed

## REVIEW — corpus-pv-004
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Static evidence supports the opener-defect outcome; independent execution stopped at missing npm. Rerun with the pinned toolchain before calling the case verified.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T17:13:20Z

## REDACTED

none
