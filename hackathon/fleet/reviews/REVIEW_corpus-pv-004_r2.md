# REVIEW — corpus-pv-004 · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:08:59Z

## Summary

The case demonstrates both malformed openers and supports the expected outcome, but the case code predates the repository's 22 October start and the check has no offline guarantee.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-004

### Ground

Reviewed `fe146e0...d786c35`. `git rev-parse --short=7 HEAD` printed `d786c35`. `git diff --check fe146e0...d786c35` exited 0. The diff changes only PV-004 files outside `expected/` and PV-004's manifest status. I read CASE.md §2 before running the check and read `expected/EXPECTED.md` afterward. This case uses TypeScript, pinned to 7.0.2, and commonmark 0.31.2. This seat ran Node v22.23.2 and npm 10.9.8. `go.mod` pins Go 1.26, but this diff contains no Go code. No numeric quality bar is declared for this case. The CommonMark parser is an external reference; the shipped suite is author-controlled.

### Differential

The first `bash tests/corpus/cases/PV-004/check.sh` exited 243 because npm could not write its default cache. Repeating it with `npm_config_cache` set to `█████████████████████████████████████████████████` exited 0 and printed `shipped-suite ok 8`, followed by:

`````text
unclosed base=command:echo hello change=refuse:unclosed-fence reference=command:echo hello
closer-too-short base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello\n```
closer-too-long base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello
closer-longer-still base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello
closer-tildes base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello\n~~~~
closer-tildes-short base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello\n~~~
text-after-closer base=command:echo hello change=refuse:text-after-closing-fence reference=command:echo hello\n```` trailing
text-after-closer-word base=command:echo hello change=refuse:text-after-closing-fence reference=command:echo hello\n````note
well-formed base=command:echo hello change=command:echo hello reference=command:echo hello
info-string-backtick base=command:echo hello change=command:echo hello reference=command:
indent-four base=command: change=command:echo hello reference=refuse:no-fenced-block
`````

For all eight shipped cases, base returns a command and change returns the named refusal. All three agree on the well-formed command. On both malformed openers, change returns `echo hello`; the reference does not include that command in a fenced block. In `info-string-backtick`, the final three-backtick line opens a new, empty reference block. The first line is still not a valid opener.

### Unplug

The false-world answer for the central claim would have change refuse either malformed opener or fail to return `echo hello`. The observed output distinguishes that world. The shipped suite cannot: its eight inputs vary only the closing fence, so opener handling can fail while the suite stays green. `check/run.ts:44-59` asserts only the suite and prints the other results; a regression in the malformed-opener output could exit 0. Remedy: assert the two opener outcomes and the well-formed control while retaining the printed differential.

### Closure

**BLOCKING — code predates the permitted window.** `CASE.md:4` and `CASE.md:79` date the case code and its run to 21 October 2026. `git show --format=fuller --no-patch d786c35` gives an AuthorDate of 21 October. AGENTS.md requires all project code to be written between 22 and 27 October and forbids copying older work. The 22 October cherry-pick CommitDate does not change when the code was written. Remedy: have the test author rebuild the case code from §2 within the permitted window and record its new provenance.

**BLOCKING — the check has no offline guarantee.** `CASE.md:21` requires the case to work with no network, but `check.sh:5` runs `npm ci --ignore-scripts`. Its dependencies are pinned in the lockfile, not included locally. A fresh judge with an empty package cache may need the registry. The successful local run had a populated cache and cannot prove offline operation. Remedy: provide a local offline dependency path and run the check with network unavailable and an empty cache.

**OWED — the differential is not an asserted result.** `check/run.ts:44-59` exits nonzero for shipped-suite failures only. The two malformed opener results can cease to demonstrate the defect without failing the command. The manual output still suffices for this review's conclusion. Remedy as under Unplug.

The expected outcome is supported: change returns a command without refusal for both malformed openers while the CommonMark reference does not find that command in a fenced block. The opposite outcome would require a refusal or a matching reference command, neither of which appears. The eight closing-fence refusals work; this review does not claim otherwise.

### Remainder

The diff stays inside the named case and its own manifest entry; `expected/` is unchanged. A content scan found no internal host or seat name, internal absolute path, private repository name, real token, or real email address in the added case. `CASE.md:44` names a private record only in general terms. `CASE.md:79` documents `ls --time-style=full-iso`, a GNU option unavailable on standard macOS `ls`; it is not called by `check.sh`. The check requires Bash, npm, Node and local `tsc`. macOS execution was not measured, so `dirname --` in `check.sh:3` remains a portability risk to verify there. `npm ci --ignore-scripts` does not run install scripts.

### Profile

Docs and test corpus. The description stands alone and its recorded output matches this run. The dependency delta adds exact pins for commonmark 0.31.2 and TypeScript 7.0.2, with entities 3.0.1, mdurl 1.0.1 and minimist 1.2.8 locked transitively. No quality threshold, suppression, skip, or test assertion was weakened in this diff. Check ranking: external CommonMark parser, then the author's suite. No numeric coverage, mutation, lint, or type bar is declared.

### Focus

1. The fixture reproduces the two malformed-opener conditions and the well-formed control.
2. The expected defect follows from those lines; the reference's empty command is explained above.
3. The eight closing checks still prove their named refusals, but the opener output is unasserted.
4. No private material was found in the added case.
5. The diff changes only PV-004 outside `expected/` and its manifest row.
6. The Linux run succeeded after cache redirection. macOS execution is unmeasured; the documented GNU `ls` option is nonportable.

### Conditions

Reject until both blocking items are resolved. The output assertion is owed work. No condition was reclassified. Ruling: apply the explicit 21 October provenance and no-network clause even though the manual differential is sound; if wrong, the cost is an unnecessary rebuild and offline packaging step. Declined to judge: the private original reviews and licenses, which are outside this review's evidence and scope.

### Unrun

macOS execution: no macOS seat was available. Fresh-cache, network-disabled execution: not run; this seat had a populated npm cache and the check has no offline mode. Go CI suite and race detector: not applicable to this TypeScript-only diff. No numerical test bar could be remeasured because none is declared.

## The review status as filed

## REVIEW — corpus-pv-004
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The case demonstrates both malformed openers and supports the expected outcome, but the case code predates the repository's 22 October start and the check has no offline guarantee.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:05:28Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
