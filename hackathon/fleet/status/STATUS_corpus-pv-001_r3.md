# STATUS — corpus-pv-001 · round 3

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 5b4f1dcbf74c4f73d33ce5ee376a9f5d4d6ab21d [corpus-pv-001] tests: add the PV-001 frame guards and record what the check printed
- updated: 2026-10-21T17:27:18Z

## Summary

Grok 4.7: on duplicate-attribute, base printed pass and change printed pass (reference printed fail missing=home).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-001
- state: stop-2-ready
- abstract: Grok 4.7: on duplicate-attribute, base printed pass and change printed pass (reference printed fail missing=home).
- last commit: 5b4f1dcbf74c4f73d33ce5ee376a9f5d4d6ab21d [corpus-pv-001] tests: add the PV-001 frame guards and record what the check printed
- done: PV-001 case code is built and check.sh printed one line per contract input
  - bash tests/corpus/cases/PV-001/check.sh exited 0 and printed 13 lines, pasted in CASE.md section 8
  - git diff --stat main...HEAD lists 10 files, all under tests/corpus/cases/PV-001 or tests/corpus/manifest.json, and none under expected
- next / blocked-on: none
- artifacts:
  - tests/corpus/cases/PV-001/CASE.md
  - tests/corpus/cases/PV-001/package.json
  - tests/corpus/cases/PV-001/package-lock.json
  - tests/corpus/cases/PV-001/tsconfig.json
  - tests/corpus/cases/PV-001/check.sh
  - tests/corpus/cases/PV-001/check.ts
  - tests/corpus/cases/PV-001/base/guard.ts
  - tests/corpus/cases/PV-001/change/guard.ts
  - tests/corpus/cases/PV-001/reference/read.ts
  - tests/corpus/manifest.json
- evidence:
  - ███████████████████████████████████████████
- note: Result spelling is pass, or fail missing= and the missing ids. Base uses a raw quoted data-screen pattern, so whitespace including a newline around = still matches. The split-line input is that pattern in text, not an element. The reference applies the same required-id test to parse5 8.0.1, which drops a duplicate attribute and keeps the first. Every input requires home. duplicate-attribute is other then home. check.sh exits 0 when every line prints. Manifest PV-001 fixture is built and checks is run; independent_check stays not done; no other case was edited. The program date is denied (allowed form none); the run clock is the mtime of dist/check.js, 2026-10-21 17:15:20 UTC. No environment variable named the model; this session states Grok 4.7. No testing-bar metric is declared, so that bar is UNRUN. No bar-weakening edit. The case had no code to baseline, and the repository has no root test script. Refutation field: duplicate-attribute printed base=pass change=pass reference=fail missing=home, so the contract's disagreement is observable. NOTICED and not changed: tests/README.md counts fixture values that are not the exact string not built, and the previous values were already longer than that string. NOTICED: the order names main at 0623358; this worktree's main, and the parent of this commit, is b0578227d0cd24f1ae192be79d72a3b028cae1bb. Commands and exits: npm ci --prefix tests/corpus/cases/PV-001 exit 0; bash tests/corpus/cases/PV-001/check.sh exit 0; node --version v24.5.0; npm --version 11.19.1; tsc --version 7.0.2.
- updated: 2026-10-21T17:17:49Z

CLOSE-OK · corpus-pv-001 · commit 5b4f1dcbf74c4f73d33ce5ee376a9f5d4d6ab21d · branch fix/corpus-pv-001 · worktree █████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
