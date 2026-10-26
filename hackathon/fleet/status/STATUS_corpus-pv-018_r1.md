# STATUS — corpus-pv-018 · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: c172f0cea566dcf31d627acc8ecd43b1659d48fa [corpus-pv-018] tests: build the PV-018 case code and record its run
- updated: 2026-10-22T16:06:26Z

## Summary

Model claude-sonnet-5-5 (Sonnet 5.5). Built PV-018 base, change, mutant, reference and check.sh in Go 1.26.8 (stdlib only); check exit 0.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-018
- state: stop-2-ready
- abstract: Model claude-sonnet-5-5 (Sonnet 5.5). Built PV-018 base, change, mutant, reference and check.sh in Go 1.26.8 (stdlib only); check exit 0.
  For the input the case is about (planted citing row, 160 B, on page outside the three): base=pass, change=fail(changelog#0:160>120), mutant=pass, reference=fail.
  Shipped tests: change suite=pass, mutant suite=pass (all four tests green on the hard-coded list). Plain 160 B row on a listed page: mutant=fail, others pass.
- last commit: c172f0cea566dcf31d627acc8ecd43b1659d48fa [corpus-pv-018] tests: build the PV-018 case code and record its run
- done: PV-018 case code built and run; CASE.md §8 and header and the manifest line updated
  - bash tests/corpus/cases/PV-018/check.sh → exit 0, output pasted in CASE.md §8
  - gofmt -l tests/corpus/cases/PV-018 → empty
  - git diff --stat main...HEAD → only tests/corpus/cases/PV-018 (outside expected/) and tests/corpus/manifest.json
- next: none
- artifacts:
  - tests/corpus/cases/PV-018/check.sh
  - tests/corpus/cases/PV-018/CASE.md
- evidence:
  - commit hit packed-refs.lock "Operation not permitted" after git commit; git log -1 shows the commit landed
  - first commit subject was 74 chars and refused by the hook; re-committed shorter
  - date is denied to the executor class; run time taken from python3 clock, 2026-10-22T16:03:58Z (CASE.md §8 says 16:04Z)
  - manifest: order said status "to built"; the status is an object, so I set fixture and checks (the fields the order's header lines name), left contract and independent_check untouched
  - refutation field: contract §2 was buildable as written; no contradiction found
- updated: 2026-10-22T16:10:00Z

CLOSE-OK · corpus-pv-018 · commit c172f0cea566dcf31d627acc8ecd43b1659d48fa · branch fix/corpus-pv-018 · worktree █████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
