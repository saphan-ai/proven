# STATUS — json-sarif-outputs · round 2

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 7 min
- last commit: 04c5815 [json-sarif-outputs] cli: move the SARIF schema tests to a package
- updated: 2026-10-23T19:48:28Z

## Summary

Outputs that reach one file through a symlinked directory, a link or hard link to the file, or a relative and an absolute spelling are now refused at load with exit 7. The SARIF schema tests moved to internal/cli/sarifschema, so internal/cli builds and passes without the validator module. UNRUN on this seat: the two schema tests (module not cached, network refused), so the version-removed control was not re-run here.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — json-sarif-outputs
- state: stop-2-ready
- abstract: Outputs that reach one file through a symlinked directory, a link or hard link to the file, or a relative and an absolute spelling are now refused at load with exit 7. The SARIF schema tests moved to internal/cli/sarifschema, so internal/cli builds and passes without the validator module. UNRUN on this seat: the two schema tests (module not cached, network refused), so the version-removed control was not re-run here.
- last commit: 04c5815 [json-sarif-outputs] cli: move the SARIF schema tests to a package
- done: C1, C2 and C3 of the review, in two commits on top of f6b6b1f
  - C1: outputIdentity (EvalSymlinks on the longest existing prefix) plus os.SameFile in internal/cli/config.go; TestOutputsThroughASymlinkedDirectoryAreRefused and TestSymlinkedFileAndHardLinkAreRefused pass
  - C1 red: with the EvalSymlinks branch disabled, TestOutputsThroughASymlinkedDirectoryAreRefused fails: "outputs_test.go:82: exit 0, want 7; stderr " (the reproduction of the reviewed defect)
  - C2: TestRelativeAndAbsoluteSpellingsAreOneFile passes; red with filepath.Abs replaced by filepath.Clean: "outputs_test.go:138: exit 0, want 7; stderr "
  - C3: schema tests moved by git mv to internal/cli/sarifschema/schema_test.go through cli.Run and cli.FormatterFor; go test ./internal/cli prints ok (30.450s) with no validator module
- next / blocked-on: none
- artifacts:
  - internal/cli/config.go
  - internal/cli/outputs_test.go
  - internal/cli/sarifschema/doc.go
  - internal/cli/sarifschema/schema_test.go
  - docs/OUTPUT_FORMATS.md
- note: master-was-wrong: — (nothing measured contradicts the order). git log --oneline f6b6b1f..HEAD printed 04c5815 and 5fa7d24 only. git diff --stat f6b6b1f..HEAD -- AGENTS.md internal/review printed nothing. gofmt -l internal cmd printed nothing. go vet ./... exits non-zero only on internal/cli/sarifschema (module fetch refused, x509); go vet ./internal/cli is clean. go test -count=1 ./...: 30 packages ok, internal/cli ok; FAIL: sarifschema [setup failed], and cmd/livecheck, generator, guardrail, pipeline, query, review panic with listen tcp6 bind operation not permitted (seat limits, outside W2, same six as the review). The disabled-resolution red was run before the commits and the mutants were reverted; the committed config.go carries neither. The packed-refs.lock Permission denied lines appeared and both commits landed (git log). Case-only spellings of a not-yet-existing file stay uncaught and are said so in config.go and docs/OUTPUT_FORMATS.md. The new package keeps a copy of the shape inputs, because an external test package cannot read cli's test table.
- evidence:
  - go test -count=1 -run 'Outputs|Symlink|Relative|Collid' ./internal/cli/ printed ok before the mutants
  - the two mutant runs printed FAIL on the named rows only
- updated: 2026-10-24T12:00:00Z

CLOSE-OK · json-sarif-outputs · commit 04c581558c6a429bf1f5f33bc9e20377d2cc53e7 · branch fix/json-sarif-outputs · worktree ██████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
