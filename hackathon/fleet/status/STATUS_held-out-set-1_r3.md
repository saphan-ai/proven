# STATUS — held-out-set-1 · round 3

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 6 min
- last commit: 65e4797c689fb9278786ba58ba9fe9bb34343fec [held-out-set-1] docs: name all freeze commits of held-out set 1
- updated: 2026-10-25T00:13:51Z

## Summary

Round 2 built all six held-out cases from CASE.md section 2 alone, one commit each (4b9a518 HO-001 to 324a63c HO-006), all after 8636864. Each check ran from a fresh clone and printed exactly the lines now in that case's section 8. `expected/` was never opened and its diff is empty. NOT RUN: `make check-links` (make is denied to this role), `go vet ./...`, `go test ./...` (only test/corpus/held-out changed).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — held-out-set-1
- state: stop-2-ready
- abstract: Round 2 built all six held-out cases from CASE.md section 2 alone, one commit each (4b9a518 HO-001 to 324a63c HO-006), all after 8636864. Each check ran from a fresh clone and printed exactly the lines now in that case's section 8. `expected/` was never opened and its diff is empty. NOT RUN: `make check-links` (make is denied to this role), `go vet ./...`, `go test ./...` (only test/corpus/held-out changed).
- last commit: 65e4797c689fb9278786ba58ba9fe9bb34343fec [held-out-set-1] docs: name all freeze commits of held-out set 1
- done: Six cases built and recorded, one commit each; every check reproduces its section 8 from a fresh clone.
  - master, 2026-10-25: after the codex and claude reviews of 589541d: 31ab87c amends HO-002's answer key before any run (advisor-approved, reason in held-out/README.md); 0773f83 makes the HO-005 and HO-006 check labels follow §2, replaces person names with roles and the work path with a temp dir; 65e4797 names all three freeze commits; every check exits 0 (4,4,10,3,4,3 lines); check-links green
  - master, 2026-10-24: on macOS bash 3.2 HO-005's check failed to parse (a case pattern inside $( ) without its opening parenthesis); fixed in 589541d, output unchanged; all six checks run on macm5 and every output line is in its CASE.md; expected/ untouched since 8636864; check-links green
  - HO-001 to HO-006 each have base/, change/, check.sh and a filled section 8; git log 8636864..HEAD shows six commits (instrument: git log).
  - fresh clone at HEAD, all six check.sh outputs equal section 8 (instrument: the six runs; HO-003 under go1.26.10, the rest sh or bash).
  - HO-003 shows the defect under umask 002 (change 0775 and 0664 against reference 0755 and 0644); its own test passes under 022 and fails under 002.
- next / blocked-on: none
- artifacts:
- note: master-was-wrong: none found. Helpers beyond base, change and check.sh: DESCRIPTION.md in each change/, api.sh (HO-004), service.sh and wrapper.sh (HO-005), cmd/snapcopy (HO-003). Scratch lives under █████████████ as in PV-020. Rulings: HO-006 reference for the dirty change tree is should fail (the change is defective in any tree); HO-004 and HO-005 print a version or step column the contract did not name. git diff 8636864..HEAD -- '*/expected/*' is empty; no merge commit, no Co-authored-by; author is the identity the spawn set.
- updated: 2026-10-24T16:16:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
