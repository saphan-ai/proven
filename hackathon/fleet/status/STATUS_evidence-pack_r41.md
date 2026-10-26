# STATUS — evidence-pack · round 41

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 8 min
- last commit: 5856357 [evidence-pack] docs: required files, listed sizes, short secrets
- updated: 2026-10-24T18:23:39Z

## Summary

Round 4 done on a rebase onto origin/main (local `main` ref absent; `origin/main` d9083c0). B1: Verify requires every payload file by name. B2: a pack is refused, by variable name, when a configured secret is under 8 bytes. F1: the pack's own JSON spelling (HTML escaping off) is redacted; a quote-ended value is no longer cut. Five missing controls (U2, U14, U18, U22, U23) pinned; owed items done (listed size verified, VerifyFile read bounded, ResolveCommit tested incl. option-like revisions, duplicate detail asserted). go vet clean; whole module go test -count=1: 49 packages ok, 0 FAIL.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — evidence-pack
- state: stop-2-ready
- abstract: Round 4 done on a rebase onto origin/main (local `main` ref absent; `origin/main` d9083c0). B1: Verify requires every payload file by name. B2: a pack is refused, by variable name, when a configured secret is under 8 bytes. F1: the pack's own JSON spelling (HTML escaping off) is redacted; a quote-ended value is no longer cut. Five missing controls (U2, U14, U18, U22, U23) pinned; owed items done (listed size verified, VerifyFile read bounded, ResolveCommit tested incl. option-like revisions, duplicate detail asserted). go vet clean; whole module go test -count=1: 49 packages ok, 0 FAIL.
- last commit: 5856357 [evidence-pack] docs: required files, listed sizes, short secrets
- done: all of the order's build items are committed on fix/evidence-pack, nothing pushed
  - rebase: `git rebase --onto origin/main c1e4ba8`; one conflict (internal/settingspage/catalog.go) kept both main's store rows and the pack section; the pack section's `evidence.Environment()` was shadowed by a local variable, renamed in 63e4a16 (build broke without it)
  - B1 TestRequiredPackFilesCannotBeOmitted (all 7 required files); negative control: `if !listed[name]` made false, 7 subtests red (████████████████████████████████████)
  - B2 TestAConfiguredSecretShorterThanTheFloorRefusesThePack (abcde and floor-1 refused naming ███████████, value never in the error; floor accepted); negative control: refusal disabled, 2 subtests red (nc-b2.txt)
  - F1 TestASecretInThePacksOwnJSONSpellingIsRedacted, TestASecretEndingInAQuoteIsRedactedWhole, redact.TestTheEscapeOffJSONSpellingIsRedacted; negative control: escape-off needle disabled, both the evidence and the redact test red (nc-f1.txt)
  - U2 TestANonCanonicalSignatureSpellingIsRefused, U14 TestAModelCountThatDiffersFromTheReceiptsIsRefused, U18 TestASecretInTheTargetOrHeadReachesNoFile, U22 cli TestALiveRunRecordsNoSecretSettingValue, U23 cli TestEveryReportFieldIsShownEscaped; these are controls over existing behaviour, so they pass on 3b6bac5's logic by design and were not run red against it
  - owed: TestAListedSizeThatIsNotTheFilesIsRefused, TestVerifyFileReadsAPackAndRefusesAMissingPath, change TestResolveCommit, duplicate/path/symlink detail asserted, redact.Shorter tests
  - whole module: go vet ./... clean; go test -count=1 ./... 49 ok, 0 FAIL (███████████████████████████████████)
- next / blocked-on: none
- artifacts:
- evidence:
  - refutation of the order (law #44): the order says the redactor ignores values under six bytes; on origin/main it does not (New redacts any length, the floor is service.MinSecretBytes = 8 in the loaders). The B2 refusal therefore uses 8 and a new Redactor.Shorter accessor, and the reviewer's codex short-secret test, which expects a pack to build, is replaced by a test that expects the refusal
  - claude's F1 fix needed one more change than the order named: strings.Trim(..., quote) in redact.spellings cut every trailing quote of a quote-ended value; replaced by inner(), tested
  - not run: B1, B2 and F1 tests against 3b6bac5 itself; the negative controls above stand for them. Local `main` ref does not exist in this worktree, so the rebase target was origin/main
  - the engine's review-verdict block said the delivered evidence belongs to an earlier round; the order carries the round 3 conditions in full (law #44, carry), so the work was done from the order, and the master should confirm that the review record matches round 3
- note: the evidence-pack docs gained three sentences (required files, listed sizes, short secrets); settingspage rename is a rebase repair, not a behaviour change
- updated: 2026-10-24T12:00:00Z

CLOSE-OK · evidence-pack · commit 5856357 · branch fix/evidence-pack · worktree █████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| private credential or variable name | 1 |
