# STATUS — corpus-pv-004 · round 24

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: 04c302474609b99772b60667c45b736a7ba245bf test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T06:15:11Z

## Summary

Seat model unreported; at 04c3024 the case check exits 0, with base=command:echo hello and change=command:echo hello for the well-formed input.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-004
- state: stop-2-ready
- abstract: Seat model unreported; at 04c3024 the case check exits 0, with base=command:echo hello and change=command:echo hello for the well-formed input.
- last commit: 04c302474609b99772b60667c45b736a7ba245bf test/corpus: name the corpus by its path after the ADR-0030 move
- done: Attested the master's carried tip and path fix, then ran the corpus check.
  - `git rev-parse --short=7 HEAD` printed `04c3024`; `git rev-parse HEAD` printed `04c302474609b99772b60667c45b736a7ba245bf`.
  - `test/corpus/cases/PV-004/CASE.md` §8 names `bash test/corpus/cases/PV-004/check.sh`; the path exists and the check ran from the repository root.
  - `bash test/corpus/cases/PV-004/check.sh` exited 0. Its 11 standard-output lines match CASE.md §8 byte for byte; the script's `diff -u` also passed.
- next / blocked-on: none
- artifacts:
- evidence:
  - The check's stderr differs from CASE.md §8 only in install timing and an npm update notice; the suite count is `shipped-suite ok 8`.
  - The master-was-wrong control found no mismatch in HEAD, command path, or output.
- note: No code or tracked document was changed and no commit was made in this round. The reviewer condition to reproduce the check is addressed by this run.
- updated: 2026-10-24T06:10:13Z

Command: `bash test/corpus/cases/PV-004/check.sh` from the repository root
Exit code: 0

Combined output (stderr precedes stdout in this capture):

`````

added 6 packages, and audited 7 packages in 939ms

2 packages are looking for funding
  run `npm fund` for details

found 0 vulnerabilities
npm notice
npm notice New major version of npm available! 10.9.8 -> 12.2.1
npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.2.1
npm notice To update run: npm install -g ██████████
npm notice
shipped-suite ok 8
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

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
