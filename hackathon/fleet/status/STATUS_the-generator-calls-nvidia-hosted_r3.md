# STATUS — the-generator-calls-nvidia-hosted · round 3

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 8 min
- last commit: c981d1c2c4600bfbf19e03ac936db35b0982403a [the-generator-calls-nvidia-hosted] merge: absorb main
- updated: 2026-10-24T04:49:49Z

## Summary

C1 closed: nvidia mode reads PROVEN_NVIDIA_BASE_URL (https only, refused by name), never PROVEN_GENERATOR_BASE_URL; the host is the model event's route and is on livecheck's line. C2 closed: the key's redaction is tested in the trace, the model list and livecheck. O1-O6 tests added. One red, internal/pipeline TestMaterializedTreesAreReadableByOthers (files 664 on this host); not placed against the base, see note.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-generator-calls-nvidia-hosted
- state: stop-2-ready
- abstract: C1 closed: nvidia mode reads PROVEN_NVIDIA_BASE_URL (https only, refused by name), never PROVEN_GENERATOR_BASE_URL; the host is the model event's route and is on livecheck's line. C2 closed: the key's redaction is tested in the trace, the model list and livecheck. O1-O6 tests added. One red, internal/pipeline TestMaterializedTreesAreReadableByOthers (files 664 on this host); not placed against the base, see note.
- last commit: c981d1c2c4600bfbf19e03ac936db35b0982403a [the-generator-calls-nvidia-hosted] merge: absorb main
- done: C1, C2, O1-O6 closed in two commits (d57f1bd, ab00024)
  - C1 reproduced first at 1a45a80: `go test -run TestNVIDIAModeIgnoresTheNightshiftBase ./internal/generator` failed with `request went to "127.0.0.1:8088", want the hosted address`; green after the fix
  - C1 negative control: nvidia mode reads PROVEN_GENERATOR_BASE_URL again (mutated field: the base variable in Load's nvidia branch) turned TestNVIDIAModeIgnoresTheNightshiftBase, TestNVIDIABaseMustBeHTTPSAndIsRefusedByName and TestEachModeReadsItsOwnBase red; restored
  - C2 negative control: generator.EnvKey removed from redact.Names turned TestNVIDIAKeyEchoedInTheModelFieldIsNotWritten (internal/review) and TestNVIDIAKeyEchoedInTheModelFieldIsNotPrinted (cmd/livecheck) red; restored
  - O4 control: Absent without Service in internal/review/open.go turned TestEmptyNVIDIAKeyNamesNVIDIAOnEveryRefusedCall red (nightshift named); restored
- next / blocked-on: none
- artifacts:
  - internal/generator/settings.go
  - internal/generator/live.go
  - internal/generator/hosted_trust_test.go
  - internal/generator/nvidia_test.go
  - cmd/livecheck/main.go
  - cmd/livecheck/nvidia_line_test.go
  - internal/review/nvidia_mode_test.go
  - internal/settingspage/catalog.go
  - docs/LAB.md
  - docs/CONFIGURATION.md
- evidence:
  - `go build ./...` exit 0; `go vet ./...` exit 0; `gofmt -l .` empty
  - `go test -count=1 ./...`: every package ok, including the listener-based ones (this host allowed loopback listeners), except internal/pipeline
  - internal/pipeline: TestMaterializedTreesAreReadableByOthers fails, "go.mod mode 664 is writable by group or others". I did not change that code; I could not run it at the base (git worktree creation was denied) and could not try umask 022 (umask is outside the permission class). Not placed against a clean base, so the master runs it.
  - One test failure of my own was fixed on the way: cmd/livecheck/main.go may not import net/http (TestInternalAndCmdDoNotImportBannedPackages), so the test seam is a function variable, newLive
  - No Co-authored-by trailer on either commit (checked with git log)
  - O6: the deadline-before-bad-stream ordering is kept, pinned by TestABadEventAfterTheDeadlineIsATimeout and explained in a comment in live.go
  - Judgement: livecheck line in nvidia mode now ends with `, host <host>`; the order's example `mode nvidia, … receipt no receipt` is a prefix of it
  - Session recording: none; this one-shot headless leg has no recorder (ADR-0024)
- note: master-was-wrong: nothing; the C1 reproduction failed at the base as the order said. The review-level key test uses a stub generator that reports the echoed model, since Open builds its own live client with no transport seam; the redactor path it proves is the same one a live answer takes. Laws echoed: see last message.
- updated: 2026-10-24T04:10:00Z

CLOSE-OK · the-generator-calls-nvidia-hosted · commit ab00024 · branch fix/the-generator-calls-nvidia-hosted · worktree █████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
