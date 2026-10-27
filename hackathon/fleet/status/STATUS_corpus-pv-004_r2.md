# STATUS — corpus-pv-004 · round 2

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 18 min
- last commit: 6fbc99639ef8fb06700539ec81ce75624711024f [corpus-pv-004] tests: build the fenced-block case and record the check
- updated: 2026-10-21T17:29:53Z

## Summary

Seat model Grok 4.7 (agent name cursor; the environment has no model slug). info-string-backtick: base and change both printed command:echo hello. indent-four: base printed an empty command and change printed command:echo hello. Shipped suite 8 ok; check.sh exit 0.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-004
- state: stop-2-ready
- abstract: Seat model Grok 4.7 (agent name cursor; the environment has no model slug). info-string-backtick: base and change both printed command:echo hello. indent-four: base printed an empty command and change printed command:echo hello. Shipped suite 8 ok; check.sh exit 0.
- last commit: 6fbc99639ef8fb06700539ec81ce75624711024f [corpus-pv-004] tests: build the fenced-block case and record the check
- done: PV-004 case code is built and check.sh has been run
  - bash tests/corpus/cases/PV-004/check.sh from the repository root exited 0 and wrote 11 stdout lines, one per input
  - standard error included shipped-suite ok 8
  - well-formed printed command:echo hello for base, change, and reference
- next / blocked-on: none
- artifacts:
  - tests/corpus/cases/PV-004/CASE.md
  - tests/corpus/cases/PV-004/check.sh
  - tests/corpus/cases/PV-004/package.json
  - tests/corpus/cases/PV-004/package-lock.json
  - tests/corpus/cases/PV-004/tsconfig.json
  - tests/corpus/cases/PV-004/.gitignore
  - tests/corpus/cases/PV-004/base/extract.ts
  - tests/corpus/cases/PV-004/change/extract.ts
  - tests/corpus/cases/PV-004/change/suite.ts
  - tests/corpus/cases/PV-004/check/reference.ts
  - tests/corpus/cases/PV-004/check/run.ts
  - tests/corpus/cases/PV-004/types/ambient.d.ts
  - tests/corpus/manifest.json
- evidence:
  - node --version printed v24.5.0
  - npm --version printed 11.19.1
  - typescript 7.0.2 and commonmark 0.31.2 are the exact pins in package.json
  - bash tests/corpus/cases/PV-004/check.sh exit 0
  - git diff --stat main...HEAD lists 13 files under tests/corpus/cases/PV-004/ and tests/corpus/manifest.json
  - git log -1 is 6fbc99639ef8fb06700539ec81ce75624711024f
- note: A closing fence is the wrong length when its length is not exactly the opener length, so closer-too-long is refused by change while commonmark 0.31.2 accepts it. No testing-bar metric is declared, so that bar is UNRUN. The product suite was not run; this order confines the leg to the case check. The date program is outside the executor command class; the run clock 2026-10-21T17:17:29Z is the mtime of dist/check/run.js.
- updated: 2026-10-21T17:19:49Z
remark: info-string-backtick reference=command: is the empty fenced block commonmark 0.31.2 opened from the closing line of three backticks, not a missing result.
remark: the environment names the agent cursor and does not name a model slug; this leg is Grok 4.7. A missing slug is not evidence that no model ran.
remark: closer-too-long and closer-longer-still are refused by change and accepted by commonmark 0.31.2, because change requires the closing fence length to equal the opener.

CLOSE-OK · corpus-pv-004 · commit 6fbc99639ef8fb06700539ec81ce75624711024f · branch fix/corpus-pv-004 · worktree █████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
