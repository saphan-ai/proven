# STATUS — corpus-pv-001 · round 34

- state: blocked
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: not recorded
- updated: 2026-10-23T08:32:05Z

## Summary

See the report as filed below.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-001
- state: blocked
- next / blocked-on: A role that may repair the host npm cache must restore access, then rerun the installation and case check in a fresh round.
- note: npm ci --prefix tests/corpus/cases/PV-001 exited 243 with EACCES creating the npm cache directory. The order requires stopping after one permission denial. No case files changed; no check ran.
- artifacts:
  - SHOUT_npm-cache-permission.md

<!-- ENGINE-EMITTED SKELETON, rendered from ██████████████████████████████.
     Everything below this line is a WORKED EXAMPLE for you to copy. It is
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy it OUT of the comment.

     Last act before done: set the state bullet above to a returned token
     (stop-2-ready · close-ok · closed-without-merge), add an `- abstract:` — mandatory for
     those states, at most 5 lines — plus `- last commit:` and `- done:`
     naming the real commit and a one-line summary, then emit the close line:

     CLOSE-OK · corpus-pv-001 · commit <sha> · branch <branch> · worktree █████████████████████████████████████████████████

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
