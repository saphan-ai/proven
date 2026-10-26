# STATUS — formatter-docs-and-example · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 18 min
- last commit: d39e0928c919406504929d7c4ee82b0bc95a0b15 [formatter-docs-and-example] merge: absorb main, keep both docs sections
- updated: 2026-10-24T15:23:48Z

## Summary

Documented how to add an output format, with examples/formatter proving Formatter and a golden line. go test ./... still fails one pre-existing pipeline test (file mode 664). Next is master verification.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — formatter-docs-and-example
- state: stop-2-ready
- abstract: Documented how to add an output format, with examples/formatter proving Formatter and a golden line. go test ./... still fails one pre-existing pipeline test (file mode 664). Next is master verification.
- last commit: d39e0928c919406504929d7c4ee82b0bc95a0b15 [formatter-docs-and-example] merge: absorb main, keep both docs sections
- done: Adding an output format is documented, and the line formatter is built by go test.
  - go doc ./internal/cli Formatter exit 0: Write(w io.Writer, result review.Result) error. go doc ./internal/cli FormatterFor exit 0: closed set text, json, sarif.
  - go test -count=1 -v ./examples/formatter exit 0: TestLinesMatchGolden, TestEmptyResultNamesItself, TestWriteReturnsTheWriterError passed, 0 failed, 0 skipped.
  - Negative control mutated one byte of examples/formatter/testdata/lines.golden (12 to 13). go test -count=1 -v ./examples/formatter -run TestLinesMatchGolden exit 1. The byte was restored and the same package test exited 0.
  - gofmt -l . exit 0, no files listed. go vet ./... exit 0.
  - git log --oneline 9b11366..HEAD is only b30da44 and db08839. git diff --name-only 9b11366..HEAD lists docs/ADAPTERS.md, examples/README.md, and examples/formatter (lines.go, lines_test.go, testdata/lines.golden).
- next / blocked-on: master verification. The pre-existing pipeline failure is outside this wagon.
- artifacts:
  - docs/ADAPTERS.md
  - examples/README.md
  - examples/formatter/lines.go
  - examples/formatter/lines_test.go
  - examples/formatter/testdata/lines.golden
  - ███████████████████████████████████████
- evidence:
  - go version exit 0: go1.26.10 linux/amd64. go.mod says go 1.26 and names no toolchain.
  - Baseline before edits, go test -count=1 ./... exit 1. The only FAIL was internal/pipeline TestMaterializedTreesAreReadableByOthers: materialized go.mod mode 664 is writable by group or others. examples/formatter did not exist yet.
  - Tip, same command, exit 1. examples/formatter ok. The same pipeline test failed the same way (mode 664). Every other tested package printed ok. Ten packages printed no test files. Per-test pass, fail and skip counts for the whole suite were not printed without -v. UNRUN at that grain.
  - Negative control red, exit 1: lines differ; got billing/discount.go:12; want billing/discount.go:13. Full lines are in █████████████████████████████████████████████████████████████
  - testing-bar: the order declares no metric. UNRUN. No bar was moved.
  - go test -race UNRUN. The example adds no goroutine and no shared mutable state.
  - ruling: the example writes path:line: title, which the order allowed beside CSV. Cost if wrong: a reader wanted quoted fields. The page says a program must not split the line on the first colon.
  - ruling: the example is not added to FormatterFor. The order forbids editing internal/cli. Cost if wrong: a reader expects the proven command to emit this format. The page says it is not registered.
- note: master-was-wrong: there is no --format flag. Formatter and FormatterFor match the order (Write takes an io.Writer and a review.Result; FormatterFor is a closed switch over text, json and sarif, with no registry). The standard-output flags are --text, --json and --sarif, else PROVEN_OUTPUT_FORMAT, else text. RED: docs/README.md at 9b11366 still contains conflict markers between main and fix/a-language-is-one-adapter-php. This wagon did not edit that file. internal/pipeline TestMaterializedTreesAreReadableByOthers failed on the base and on the tip because the materialized go.mod is mode 664. internal/, cmd/ and AGENTS.md were not edited. The umask command is outside this role's command class, so the mode was not changed.
- note: master-absorb: main moved and conflicted on docs/ADAPTERS.md (main added the evidence-store section where this branch added the output-format section); the master resolved it by keeping both, commit d39e092 (merge of db08839 and main fafce65); on the merged tree build, vet, examples/formatter, internal/cli and internal/cli/sarifschema are ok.
- updated: 2026-10-24T15:20:00Z

CLOSE-OK · formatter-docs-and-example · commit d39e0928c919406504929d7c4ee82b0bc95a0b15 · branch fix/formatter-docs-and-example · worktree ██████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
