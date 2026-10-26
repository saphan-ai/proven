# STATUS — outputs-cannot-be-steered · round 1

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: unverified (no pool reading on this seat)
- cost: not recorded in this leg
- duration: not recorded in this leg
- last commit: bc5dab3ac385b0124aa5b4a666a7a8429e2e8c06 "[outputs-cannot-be-steered] cli: keep reports inside the checkout"
- updated: 2026-10-24T11:46:05Z

## Summary

RED: `go test -count=1 ./...` exits 1. `internal/pipeline` `TestMaterializedTreesAreReadableByOthers` reports `go.mod` mode 664. The same failure was present before any edit on this branch. This wagon does not change that package.

A final symbolic link is refused before the review (exit 7) and is not followed at the write. JSON and SARIF write DEL and C1 as `\u007f` and `\u0080` to `\u009f`. A SARIF artifact URI stays a relative path in the repository. A path that cannot is omitted, and the finding is kept. This log sets no `uriBaseId`.

## Acceptance, item by item

1. `git log --oneline 386bbf0..HEAD` is not only this wagon. This wagon's commits are `652f972` and `bc5dab3`. Already on the branch before this leg: `00d21f7`, `c1e4ba8`, `881623d`, `045391c`, `0b9c079`. The worktree tip at the start was `00d21f7`, not `386bbf0`. `git diff 386bbf0 HEAD` of `internal/cli/run.go`, `format_json.go` and `format_sarif.go` was empty at the start.
2. The three tests, run before any production edit, exit 1.
   `go test -count=1 -v ./internal/cli -run 'TestJSONAndSARIFCarryNoRawDELOrC1|TestSARIFArtifactURIStaysARelativePathInTheRepository|TestOutputFileIsNotWrittenThroughALinkInTheCheckout'`
   - JSON: raw control byte at offset 434. SARIF: raw control byte at offset 361.
   - `artifactURI("//example.invalid/a.go")` = `"//example.invalid/a.go"` starts with a slash.
   - `artifactURI("/etc/a.go")` = `"/etc/a.go"` starts with a slash.
   - `artifactURI("../outside/a.go")` = `"../outside/a.go"` has a `..` segment.
   - `artifactURI("a/../../b.go")` = `"a/../../b.go"` has a `..` segment (reported twice).
   - The report was written through the link. The file no longer held `kept\n`.
3. `go vet ./...` exit 0. `gofmt -l .` exit 0, no output. `go test -count=1 ./...` exit 1. 44 packages `ok`, 10 `[no test files]`, 1 fail (`internal/pipeline`, the test named above). `internal/cli` ok. `internal/cli/sarifschema` ok. Go is `go1.26.10`. `go.mod` says `go 1.26` and has no toolchain line. Per-test skip counts are UNMEASURED: a package line of `ok` does not say whether a test was skipped.
4. Unplug, each fix removed and restored:
   - `TestJSONAndSARIFCarryNoRawDELOrC1`: `escapeDELAndC1` returned the bytes unchanged. Exit 1. JSON offset 434, SARIF offset 361.
   - `TestSARIFArtifactURIStaysARelativePathInTheRepository`: the in-repository check was skipped. Exit 1. The four paths again had a leading slash or a `..` segment. `TestSARIFKeepsAFindingWhenThePathLeavesTheRepository` also went red (1 location).
   - `TestOutputFileIsNotWrittenThroughALinkInTheCheckout`: `writeOutputFile` called `os.WriteFile` only. Exit 1. The file the link names was replaced.
   - `TestFinalSymlinkOutputIsRefusedAtLoad`: the load call was skipped. Exit 1. Each flag returned exit 2, want 7. The write still refused the link.
5. `- master-was-wrong:` the three reds matched the order. The branch tip was not `386bbf0`. See item 1.

## Rulings

- Refuse a final symlink (exit 7, flag and path named, before the review) rather than replace the link. The write also refuses, because the reviewer's test calls `writeOutputs` directly. Cost if wrong: a lone dangling symlink that used to create its target now exits 7.
- The symlink refusal runs after the same-file checks and the standard-output check, so those refusals still name both sides. Cost if wrong: an existing collision test would fail. They passed.
- A path that cleans to a relative path inside the repository is kept (`dir/../a.go` is `a.go`). A path that stays absolute or still has `..` is omitted. No `uriBaseId`, because the writer does not set one. Cost if wrong: a consumer that wanted the absolute path as a URI does not get a location. The finding stays.

## Not run

- `go test -race`: the change adds no goroutine and no shared mutable state.
- A testing-bar metric: none is declared for this repository.
- `proven scan`: the order says not to run it in the checkout.

## REDACTED

- absolute path, 2: the two `go.mod` lines of the pipeline failure. The mode was 664.
