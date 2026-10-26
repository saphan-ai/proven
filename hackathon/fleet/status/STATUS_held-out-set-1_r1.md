# STATUS — held-out-set-1 · round 1

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 86368647504aae1abfcd9a0d9ababea62952e47f [held-out-set-1] docs: name both freeze commits of held-out set 1
- updated: 2026-10-24T15:22:59Z

## Summary

Freeze commit 0fabc628934f790fa92b0a3d3d39362a2ff29bfe holds only test/corpus/held-out/** (6 cases, 6 answer keys, manifest, README); README commit ece0efb names it.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — held-out-set-1
- state: stop-2-ready
- abstract: Freeze commit 0fabc628934f790fa92b0a3d3d39362a2ff29bfe holds only test/corpus/held-out/** (6 cases, 6 answer keys, manifest, README); README commit ece0efb names it.
  Run: git show --stat (only held-out paths), grep for product/vendor/host/person names (no hits), git log 492f03b..HEAD (no merge, no Co-authored-by, author Marcin Marzec).
  Not run: make check-links (make is denied for this role), go vet and go test (no code added, not run).
  master-was-wrong: base is not 9b11366; the branch tip was 492f03b, three commits later. Manifest records 492f03b.
- last commit: 86368647504aae1abfcd9a0d9ababea62952e47f [held-out-set-1] docs: name both freeze commits of held-out set 1
- done: six held-out cases frozen in one commit, then the README commit
  - master, 2026-10-24: the codex review found HO-003's reference ambiguous; second freeze commit b4bab25 fixes it as literals (0644/0755) before any run, with the reason in held-out/README.md; 8636864 names both freeze commits in the README; check-links green
  - git show --stat 0fabc62 lists 14 files, all under test/corpus/held-out/
  - grep -rniE for product, vendor, host and person names over test/corpus/held-out printed nothing
- next: owner review; fixture round by test authors who do not read expected/; check-links to be run by a role that holds make
- blocked-on: none
- artifacts:
  - test/corpus/held-out/manifest.json
  - test/corpus/held-out/README.md
- note: master-was-wrong — the order names base main @ 9b11366, but the branch tip at start was 492f03b (three security-piece-8 merges later); no local main ref exists. The manifest records 492f03b.
- updated: 2026-10-24

## REDACTED

none
