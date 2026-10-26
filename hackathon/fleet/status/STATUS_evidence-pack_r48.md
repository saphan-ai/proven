# STATUS — evidence-pack · round 48

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: bb1c3cc [evidence-pack] evidence: test the whole-file bound of VerifyFile
- updated: 2026-10-25T07:03:10Z

## Summary

Round 5. The BLOCKING is real: the reviewer's test was red at 5856357 and is green now. `--evidence-pack` joins the load-time collision check and the write-time SameFile check, so every output flag, a link, and a link to its directory are refused. Whole module green: vet clean, 49 ok, 0 FAIL. Owed round-3 items were already in the tree (see evidence).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — evidence-pack
- state: stop-2-ready
- abstract: Round 5. The BLOCKING is real: the reviewer's test was red at 5856357 and is green now. `--evidence-pack` joins the load-time collision check and the write-time SameFile check, so every output flag, a link, and a link to its directory are refused. Whole module green: vet clean, 49 ok, 0 FAIL. Owed round-3 items were already in the tree (see evidence).
- last commit: bb1c3cc [evidence-pack] evidence: test the whole-file bound of VerifyFile
- done: ADDRESSED — the pack is part of the output collision check
  - TestEvidencePackOutputCannotReplacePack (reviewer's test, verbatim): red at 5856357 with `colliding output was accepted: ""`, green now; `go test -count=1 -run 'TestEvidencePack|TestWriteOutputsLeaves' -v ./internal/cli`
  - TestEvidencePackCollidesWithEveryOutputFlag (-o, --output, --text-output, --json-output, --sarif-output): red then green; asserts the refusal names --evidence-pack and no pack is written
  - TestEvidencePackAliasedOutputCannotReplacePack (link to existing pack path, link to a not-yet-made pack path, link to the pack's directory): red then green
  - TestWriteOutputsLeavesThePackWhenALinkAppearsAfterTheLoadCheck (second line, write time): red then green; the pack bytes are unchanged
  - negative control (mutated field: the pack's entry in outputFiles and its seeding in writeOutputs): all four tests red by name, including the reviewer's; restored; capture in ███████████████████████████████████████████████
- next: none
- blocked-on: none
- evidence:
  - red capture at 5856357 (product code unchanged): ██████████████████████████████████████████
  - whole module: `go vet ./...` no output; `go test -count=1 ./...` 49 package rows ok, 0 FAIL, gofmt -l empty; capture ███████████████████████████████████
  - master-was-wrong: none. The reviewer's finding reproduced.
  - ruling: refuseStdoutOnAnOutput also lists the pack (shared outputFiles list) — a stdout redirect onto the pack path is refused at load; cost if wrong: one extra refusal for an odd redirect
  - ruling: the load check names the second flag in the pair, so the sentence reads "same file as --evidence-pack"; cost if wrong: wording only
  - round-3 owed items, already present at 5856357 and not changed: duplicate subtest asserts Detail (pack_test.go wantDetail); FileRef.Size verified in verify.go and tested (round4_test.go "size is"); VerifyFile reads through LimitReader(maxPack+1); TestResolveCommit covers option-like names, the --end-of-options case
- master-commit: bb1c3cc, by the master after the round, adds the verifyFileWithin seam and the claude reviewer's TestAPackOverTheWholeFileBoundIsRefused (C1); verified on █████████████ (test passes, red with the bound check disabled; go test ./... 49 ok, 0 FAIL); the claude re-read of bb1c3cc returned accept.
- note: NOTICED, not done (C5): VerifyFile's oversize refusal has no test of its own. AGENTS.md section 1 asks for a public copy of this status under fleet/status in the same change; the order says to commit nothing it does not name, so none was made — the master owns it. Commit landed despite packed-refs.lock Permission denied (git log -1 read).
- artifacts:
  - internal/cli/config.go
  - internal/cli/run.go
  - internal/cli/pack_round5_test.go
  - docs/EVIDENCE_PACK.md
- updated: 2026-10-25T07:10:00Z

CLOSE-OK · evidence-pack · commit 0b71da0 · branch fix/evidence-pack · worktree █████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 4 |
| machine name | 1 |
