# STATUS — security-piece-8-fixes · round 3

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 2f36e31 [security-piece-8-fixes] merge: absorb main
- updated: 2026-10-24T12:39:43Z

## Summary

Round 2 closes the four items. A second-reviewer key under 8 bytes is refused at load. The inside-the-target check now decides by file identity as well as spelling. A working-tree export test pins the explicit file modes. The redactor comments and ADR-0013 now say a short organisation id is written.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — security-piece-8-fixes
- state: stop-2-ready
- abstract: Round 2 closes the four items. A second-reviewer key under 8 bytes is refused at load. The inside-the-target check now decides by file identity as well as spelling. A working-tree export test pins the explicit file modes. The redactor comments and ADR-0013 now say a short organisation id is written.
- last commit: 2f36e31 [security-piece-8-fixes] merge: absorb main
- done: all four items committed, each with a test that goes red when its fix is reverted
  - second-reviewer key floor: TestLoadSecondRefusesAKeyTooShortToBeRedacted (red at base, exit 1; green after, exit 0; fix reverted at second.go `|| service.TooShortSecret(key)` → red exit 1)
  - identity not spelling: TestInsideDecidesByFileIdentityNotSpelling in internal/pathguard (red at base, exit 1; green after; `return sameFileAbove(rd, rr), nil` replaced by `return false, nil` → red exit 1; skips by name on a case-sensitive volume, ran on darwin)
  - export modes: TestExportWorktreeSetsFileModesWhateverTheUmask in internal/change (green; Chmod in copyTracked disabled → red, modes 600/700 vs 644/755)
  - redact Names and FromEnv comments corrected; ADR-0013 point 4a states the organisation-id exception
- next / blocked-on: none; the master runs the listener-refused packages
- artifacts:
- evidence:
  - go build ./... and go vet ./... clean; gofmt -l internal clean
  - go test -count=1 ./... green except six packages, UNRUN, each refused at its listener: cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review. Quoted refusal: `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`
  - internal/generator run with -run 'TestLoadSecond|TestLoadRefuses|TestLoadAccepts': ok
  - git commit printed `packed-refs.lock: Operation not permitted`; git log shows both commits landed
- note: master-was-wrong — the order's O1 asks for a twin under umask 002, but 0644 survives umask 002, so no test under 002 could go red; the test uses umask 077, which does (measured: mode 600 without the Chmod).
- updated: 2026-10-24
CLOSE-OK · security-piece-8-fixes · commit 2f36e31 · branch fix/security-piece-8-fixes · worktree ██████████████████████████████████████████████████████████████████████████████

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 3e737d82f77c911ba08a4927b3f923da12340d1d

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| seat identity | 1 |
