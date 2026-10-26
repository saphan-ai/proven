# STATUS — org-policy · round 21

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 10 min
- last commit: 2abb5cc [org-policy] merge: absorb main, keep the policy and evidence pack settings
- updated: 2026-10-25T10:33:39Z

## Summary

Round 4. B1 fixed: a policy that fixes the classes but names no doubled_classes now compares the effective doubled set under a second reviewer, and a narrower local list is refused by name. B3 fixed: after exceptions the proof totals are recomputed from active findings only. Both reviewer tests fail at 1199ee2 and pass at 18197ef. Whole module: 47 packages ok, 10 without tests, go vet clean, gofmt clean. JSON/SARIF (B2) not touched: proven-cli-and-action W2b.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — org-policy
- state: stop-2-ready
- abstract: Round 4. B1 fixed: a policy that fixes the classes but names no doubled_classes now compares the effective doubled set under a second reviewer, and a narrower local list is refused by name. B3 fixed: after exceptions the proof totals are recomputed from active findings only. Both reviewer tests fail at 1199ee2 and pass at 18197ef. Whole module: 47 packages ok, 10 without tests, go vet clean, gofmt clean. JSON/SARIF (B2) not touched: proven-cli-and-action W2b.
- last commit: 2abb5cc [org-policy] merge: absorb main, keep the policy and evidence pack settings
- done: B1 and B3 repaired, both tests committed as regression tests, POLICY.md updated with one sentence each
  - B1 ADDRESSED: TestLocalDoubledListCannotNarrowEffectivePolicyWithSecondReviewer, policy.go weaker() case doubled_classes
  - B3 ADDRESSED: TestAcceptedExceptionDoesNotCountInProofTotals, policy.go applyExceptions recomputes ProofCounts
  - go test -count=1 ./... : 47 ok, 10 no test files, 0 failed; go vet ./... clean
- master-commit: b966161 absorbs main 77a72fc (additive conflict in internal/review/open.go, both sides kept; 51 ok, 0 FAIL); 58ac96b shows the policy text fields through visible() with the codex reviewer's test (passes; red with the old formatter; 51 ok, 0 FAIL); the codex re-read of 58ac96b returned accept. 2abb5cc absorbs main 29534ef after the evidence pack landed (additive conflicts in docs/CONFIGURATION.md and internal/settingspage/catalog.go, both sections kept; build, vet, go test ./... 52 ok, 0 FAIL).
- next: none
- artifacts:
  - internal/review/policy.go
  - internal/review/policy_second_probe_test.go
  - internal/review/policy_count_probe_test.go
  - docs/POLICY.md
- evidence:
  - red at 1199ee2, before any change: "policy_count_probe_test.go:18: accepted finding still counts in proof totals: 1"; "policy_second_probe_test.go:20: main.go is doubled by organisation policy as backend but reviewed once under local backend; weaker setting was not refused". The master was not wrong: both failed.
  - negative control B1 (the effective-doubled branch disabled, nothing else): TestLocalDoubledListCannotNarrowEffectivePolicyWithSecondReviewer FAIL, same message as above. Restored.
  - negative control B3 (the totals recompute disabled, nothing else): TestAcceptedExceptionDoesNotCountInProofTotals FAIL, "accepted finding still counts in proof totals: 1". Restored.
  - first B3 attempt (decrement by f.ProofLevel) did not pass the reviewer's test, whose finding has no ProofLevel; the totals are recomputed from the active findings instead.
  - existing policy-inside, owner/expiry and expiry tests stayed green in the whole-module run.
  - not run: go test -race, repeat runs; B2 (JSON, SARIF) is out of scope by the owner's ruling.
  - note: _EXEC_ORDER.md in the worktree is the round 2 order; I worked from the round 4 brief.
- updated: 2026-10-25T08:00:00Z

CLOSE-OK · org-policy · commit 18197ef · branch fix/org-policy · worktree ██████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
