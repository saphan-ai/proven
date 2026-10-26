# STATUS — evidence-pack · round 9

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 10 min
- last commit: 3b6bac5 [evidence-pack] evidence: close the four gaps of the round 2 reviews
- updated: 2026-10-24T10:24:27Z

## Summary

Round 3 closes C1-C5 in one commit after the rebase onto origin/main: manifest models come from the redacted receipts, a key file with a group or other bit is refused (exit 7, chmod 600), pack text is escaped (bidi too), Verify requires the exact archive bytes. go vet clean; whole module 45 packages ok and 1 red (internal/pipeline TestMaterializedTreesAreReadableByOthers, red at the rebased base before my change, sandbox umask). Each of C1-C4 turns its test red when its guard is removed; the six named mutants are killed; all restored.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — evidence-pack
- state: stop-2-ready
- abstract: Round 3 closes C1-C5 in one commit after the rebase onto origin/main: manifest models come from the redacted receipts, a key file with a group or other bit is refused (exit 7, chmod 600), pack text is escaped (bidi too), Verify requires the exact archive bytes. go vet clean; whole module 45 packages ok and 1 red (internal/pipeline TestMaterializedTreesAreReadableByOthers, red at the rebased base before my change, sandbox umask). Each of C1-C4 turns its test red when its guard is removed; the six named mutants are killed; all restored.
- last commit: 3b6bac5 [evidence-pack] evidence: close the four gaps of the round 2 reviews
- done: C1-C5 addressed, each with a test that fails without the fix
  - C1 ADDRESSED: Build reads Models from the redacted receipts.json; Build also verifies its own pack. Test TestASecretInAModelUseReachesNoFileAndThePackVerifies (secret in Route and Model; absent from every file; pack verifies). Control: Models back to models(in.Result) turns it red.
  - C2 ADDRESSED: LoadPrivateKey refuses any bit of 0077, naming setting, path, mode and chmod 600; exit 7 before the review. Tests TestAPrivateKeyThatGroupOrOthersCanReachIsRefused (0644 0640 0604 0660 0666 0610 refused; 0600 0400 load) and TestAReadableKeyIsRefusedBeforeTheReview. Control: mask check disabled turns both red. Docs: docs/EVIDENCE_PACK.md, docs/CONFIGURATION.md, the generated row.
  - C3 ADDRESSED: evidence.Visible in Refusal.Error; cli visible gained bidi marks (evidence.IsBidi) and renderReport, describeCommits and the verify refusal use it. Tests TestARefusalShowsNoControlOrBidirectionalCharacter and TestVerifyShowsNoControlByteFromThePack (ESC in a model name). Controls: Visible as identity reds the first; cli visible as identity reds the second and three older cli tests.
  - C4 ADDRESSED: Verify rebuilds archive(files) and requires equality; the signature file must be base64 plus one newline. Tests TestThePackIsOneByteStringAndAnyOtherIsRefused (appended byte, header byte, mtime, padding, signature space) and TestEveryChangedByteOfThePackIsRefused (every position, one bit each). Control: equality disabled turns both red.
  - C5 ADDRESSED: TestEachManifestGuardRefusesOnItsOwn kills version, key id, duplicate file, unknown field, receipts-vs-manifest; TestAnEntryOverTheBoundIsRefusedBeforeItIsRead kills the maxEntry bound; TestThePublicKeyFlagBeatsTheSetting pins flag over setting. Each mutant was applied alone and read red by test name.
- next / blocked-on: none
- note: Rebase was onto origin/main (no local main exists); one conflict in internal/settingspage/catalog.go, both constants kept. Build now verifies its pack, so the receipt test uses an unexported build for packs that must fail Verify. Reviewer premise refuted in part, as the order said: a visible helper exists on main. It had no bidi case, so I added one. Not run: -race, live mode. The one red test fails on a mode 664 file under the sandbox umask, outside the write-set. Scratch file ██████████████████████████████████████ could not be removed (rm denied); it is uncommitted and empty of effect.
- artifacts:
  - internal/evidence/key.go
  - internal/evidence/pack.go
  - internal/evidence/verify.go
  - internal/evidence/hardening_test.go
  - internal/cli/pack.go
  - internal/cli/format.go
  - internal/cli/pack_hardening_test.go
  - docs/EVIDENCE_PACK.md
- updated: 2026-10-24

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
