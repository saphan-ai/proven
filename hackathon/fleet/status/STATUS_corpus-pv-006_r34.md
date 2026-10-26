# STATUS — corpus-pv-006 · round 34

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 1 min
- last commit: bb594c8bbcd131cd5822848e2c9140ea970ee285 test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T06:38:12Z

## Summary

Attested bb594c8 on the Codex seat (model not reported to this leg); the dollar-token input printed base=pass and change=pass with raw=$wait. The six check lines match CASE.md §8; review provenance corrections remain owed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-006
- state: stop-2-ready
- abstract: Attested bb594c8 on the Codex seat (model not reported to this leg); the dollar-token input printed base=pass and change=pass with raw=$wait. The six check lines match CASE.md §8; review provenance corrections remain owed.
- last commit: bb594c8bbcd131cd5822848e2c9140ea970ee285 test/corpus: name the corpus by its path after the ADR-0030 move
- done: Attested the master's carried tip and path correction without a code change.
  - `git rev-parse --short=7 HEAD` printed `bb594c8`; `git rev-parse HEAD` printed `bb594c8bbcd131cd5822848e2c9140ea970ee285`.
  - `git show -1 --format=fuller --stat HEAD` shows the master's path correction in CASE.md; its two documented commands name `test/corpus`.
  - `bash test/corpus/cases/PV-006/check.sh` exited 0; all six lines match CASE.md §8, aside from timing which this command does not print.
- next / blocked-on: Master to reconcile the review's manifest date and macOS provenance conditions in a later authorized round; round 3 orders no code change.
- artifacts:
- evidence:
  - `git rev-parse --short=7 HEAD` exited 0: bb594c8
  - `git rev-parse HEAD` exited 0: bb594c8bbcd131cd5822848e2c9140ea970ee285
  - `bash test/corpus/cases/PV-006/check.sh` exited 0; full output follows.
  - well-formed base=pass change=pass reference=pass
  - listed-token base=fail({limit}) change=fail({limit}) reference=fail
  - unlisted-token base=pass change=fail({window}) reference=fail
  - dollar-token base=pass change=pass reference=fail raw=$wait
  - duration-0-vs-1 base=pass change=pass reference=fail shown=0 retry-after=1
  - duration-equal base=pass change=pass reference=pass shown=30 retry-after=30
- note: The provisioned review asks for a manifest check-date correction and a cited macOS run. The current order explicitly forbids edits, so neither condition was changed. The manifest says checks ran 2026-10-22; CASE.md §8 records 2026-10-23. A macOS run was not independently verified here. Whole-repository tests were not run because this round attests the existing tip only.

<!-- ENGINE-EMITTED SKELETON, rendered from ██████████████████████████████.
     Everything below this line is a WORKED EXAMPLE for you to copy. It is
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy it OUT of the comment.

     Last act before done: set the state bullet above to a returned token
     (stop-2-ready · close-ok · closed-without-merge), add an `- abstract:` — mandatory for
     those states, at most 5 lines — plus `- last commit:` and `- done:`
     naming the real commit and a one-line summary, then emit the close line:

     CLOSE-OK · corpus-pv-006 · commit <sha> · branch <branch> · worktree █████████████████████████████████████████████████

     The `- artifacts:` manifest (and `- return:` and `- evidence:`, which read the SAME grammar) is a
     NESTED DASH LIST. The key line carries NOTHING; the items sit beneath it:

     - artifacts:
       - path/to/first.go
       - path/to/second.go

     `- artifacts: a, b, c` parses as ONE path literally named "a, b, c", not
     three paths — a comma NAMES a file, it never separates list members.

     An entry names a FILE, or a DIRECTORY: a directory is delivered as ONE
     deterministic `.tar` archive (sorted entries, fixed mtimes — the same
     bytes, and so the same digest, every time) recorded under the entry's
     own name plus that suffix, which is the name the projected manifest and
     the teardown gate then read. A path that EXISTS and is neither is
     refused by what it IS, never as `missing`.

     `- done:` is PROSE: its value is the KEY LINE. A nested list under it is
     a superset of that line, never a substitute — an empty `- done:` reads as
     an empty value, which is a headline missing from the record. `saphan doc
     check` refuses that shape by name (doc-prose-key-empty).
-->

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
