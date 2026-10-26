# REVIEW — held-out-set-1 · round 6

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T01:19:14Z

## Summary

Six cases match their frozen contracts and answer keys; all six checks reproduce the intended defects. No blocking findings.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — held-out-set-1

### Ground
Reviewed `9b11366...65e4797` at `65e4797` (`git rev-parse --short=7 HEAD` printed `65e4797`). The master's outputs were hypotheses, not evidence. This Linux seat has Go 1.26.5, Bash 5.3.9 and Git 2.53.0 (`go version`, `bash --version`, `git --version`). `Makefile:60` defines `make check-links`; each case has its own §8 command. These are deliberately defective candidate changes, not fixes to make.

### Differential
`git diff b4bab25..HEAD -- 'test/corpus/held-out/*/expected/*'` shows exactly HO-002's amended key sentence: its mixed state contains a valid current bot marker and must pass. `git diff 8636864..HEAD -- 'test/corpus/held-out/*/expected/*'` is empty. For each of six CASE.md files, `diff -u` between the `b4bab25` and tip versions with §8 excluded, and separately for §9, returned equal: §1–§7 and §9 stayed frozen. §8 records the case implementation and output. No project dependency, build or CI file changed (`git diff 9b11366...HEAD -- go.mod Makefile .github` empty); HO-003's two new Go modules use only the standard library.

### Unplug
The false-world answer differs in each check. HO-001: old success/new failure gives base `fail`, change `pass`; a single success gives `pass/pass` (`change/gate.sh:11`). HO-002: a person-only current marker gives change `pass` versus `should fail`; a bot current marker gives `pass` versus `should pass` (`change/gate.sh:14`). HO-003: switching umask `022` to `002` changes four entries from reference modes to group-writable modes, and its new test from `ok` to `FAIL` (`change/snapshot/snapshot.go:38`). HO-004: refused email gives base `create`, change `team`, whereas existing user gives change `ok` (`change/setup.sh:15`). HO-005: unknown model gives exit 0, logged `model-z`, served `model-a`, whereas the listed model matches (`change/tool.sh:13`). HO-006: the clean clone without the old directory gives `fail` against the dirty copy's `ok` (`change/verify.sh:20`). If these defects were absent, the contrasting outputs would not occur. References are literal or fixed by the independently defined fixture; none is derived by the candidate under test.

### Closure
`bash test/corpus/held-out/HO-001/check.sh` exit 0: `one-run-success base pass change pass reference should pass`; `one-run-failure base fail change fail reference should fail`; `success-then-failure base fail change pass reference should fail`; `failure-then-success base pass change pass reference should pass`.

`bash test/corpus/held-out/HO-002/check.sh` exit 0: `bot-marker-current base pass change pass reference should pass`; `no-marker base fail change fail reference should fail`; `person-marker-current base pass change pass reference should fail`; `person-marker-older base pass change pass reference should pass`.

`bash test/corpus/held-out/HO-003/check.sh` exit 0: under `022`, root/sub `0755` and both files `0644` for base/change/reference; under `002`, base/reference stay `0755/0644`, change gives root/sub `0775` and files `0664`. Change test `022 ok`, `002 FAIL (4 entries with a wrong mode)`.

`bash test/corpus/held-out/HO-004/check.sh` exit 0: `new-user-good-email base ok change ok last step team reference should succeed`; `user-exists base create change ok last step team reference should succeed`; `new-user-refused-email base create change team last step team reference should fail at create`.

`bash test/corpus/held-out/HO-005/check.sh` exit 0: base listed `model-b` logs/serves `model-b`; base unknown `model-z` exits 2 and logs none; change listed `model-b` logs/serves `model-b`; change unknown `model-z` exits 0, logs `model-z`, serves/reference `model-a`.

`bash test/corpus/held-out/HO-006/check.sh` exit 0: `base base ok change - reference should pass`; `change-dirty base - change ok reference should pass`; `change-clean base - change fail reference should pass`. `make check-links` exit 0: `check-links: every relative link and metadata.json path resolves`. Only after reading the contracts, source and check outputs did I read the six `expected/EXPECTED.md` files; each key agrees with both defect and control.

### Remainder
No blocking or owed finding. NOTE: HO-004's check invokes the change script twice to retrieve the word and last trace (`check.sh:42`–`check.sh:44`). The second invocation sees the same newly created user, but the first result and final traced step still produce the contracted three outcomes; there is no false verdict. No threshold, skip, suppression or exception changed. Declined to judge: actual detection by the product (this is a pre-run freeze review), macOS/Bash 3.2 execution (not available here), and production integrations (not fixture scope).

### Profile
Test-corpus review. Check ranking: §2 contracts and reference labels are project controls; executable fixture checks and HO-003's Go test are suite checks, not independent correctness oracles. Freeze diffs, code-to-contract inspection and answer-key-last inspection complement the circular suite. HO-003's test runs under both umasks with `-count=1`. No metric baseline or tolerance is declared; coverage, mutation and lint deltas are UNRUN. Generic shell, Go, Linux and Git names occur in the case material, but no specific vendor, host or person name. The README's project name describes the corpus, not a case subject.

### Focus
Accept. The six distinct causal defects are stale success, unchecked comment author, inherited creation mode, conflated status code, silent model substitution and dirty-tree path dependency. Each pair implements its §2, each check shows its defect and counterexample, and each answer key follows from the check. No product code was edited by this reviewer.

### Conditions
none

### Unrun
macOS/Bash 3.2: no such host/shell available; Linux/Bash 5.3.9 results are not a macOS measurement. Actual product review run: outside this frozen-set review. Full-repository Go tests, coverage and mutation score: not relevant to fixture contract review, not claimed. Manual mutations to the source: not performed, since the worktree must remain unchanged; the contrast inputs, umasks and clean clone are measured negative controls. Model choice, pool share and final cost: not supplied by Saphan Studio; no actual expense is inferred from the order's quote.

## The review status as filed

## REVIEW — held-out-set-1
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: Six cases match their frozen contracts and answer keys; all six checks reproduce the intended defects. No blocking findings.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T01:15:36Z

## REDACTED

none
