# STATUS — outputs-cannot-be-steered · round 3

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: unverified (no pool reading on this seat)
- cost: not recorded in this leg
- duration: not recorded in this leg
- last commit: 84a22301d1fa871d9592de143408a52abc86f31b "[outputs-cannot-be-steered] cli: contain writes inside the workdir"
- updated: 2026-10-25T04:00:53Z

## Summary

An output path inside the working directory is created through that directory opened as a root. A directory of the path that becomes a link to a place outside the directory after the load check does not receive the report. The run exits 2 and names the path. The load check resolves the working directory and the prefix already walked, so another spelling of the checkout is the same directory. A path outside the working directory is still the operator's.

RED: `go test -count=1 ./...` exits 1. `internal/pipeline` `TestMaterializedTreesAreReadableByOthers` reports `go.mod` mode 664. The same failure was present before this round. This round does not change how those trees are written.

## Acceptance, item by item

1. `git log --oneline a6b27d4..HEAD` at the code commit is one commit: `84a2230 [outputs-cannot-be-steered] cli: contain writes inside the workdir`. Parent is `a6b27d4`. This status commit follows it.
2. The reproduction, run on `a6b27d4` before any production edit, exits 1.
   `go test -count=1 -v ./internal/cli -run TestDirectoryLinkSwappedAfterLoadIsNotWrittenThrough`
   Both subtests then present (relative path, absolute path under the working directory): exit 0, stderr empty, the outside file no longer held `kept`. The load check had passed, because the swap ran during the review.
   The same command after the fix exits 0. A third subtest, a relative link target, was added with the fix and passes.
   `TestAliasedWorkingDirectorySpellingRefusesTheInsideLink` and `TestDirectoryLinkSwappedAfterLoadThroughAnAliasedSpellingIsNotWrittenThrough`, run before the production edit, exit 1: exit 0, the review ran for the load case, and the outside file no longer held `kept`. Both pass after the fix.
3. `go vet ./...` exit 0. `gofmt -l .` exit 0, no output. `go test -count=1 ./...` exit 1. 44 packages `ok`, 10 `[no test files]`, 1 fail (`internal/pipeline`, the test named above). `internal/cli` ok. `internal/cli/sarifschema` ok (1.233s). Go is `go1.26.10`. `go.mod` says `go 1.26` and has no toolchain line. No Makefile and no `.github` workflow were found; the suite above is the order's acceptance command. `GOOS=windows go vet ./internal/cli/` was not run: the permission class refused the command, reading `GOOS=windows` as the program. Allowed form: none. It was not rebuilt.
4. Unplug, the root open replaced by `os.WriteFile` and then restored:
   - `TestDirectoryLinkSwappedAfterLoadIsNotWrittenThrough`: exit 1. All three subtests exit 0, stderr empty, and the outside file was replaced. The root open was restored. The same command then exited 0.
   - `TestAliasedWorkingDirectorySpellingRefusesTheInsideLink` and `TestDirectoryLinkSwappedAfterLoadThroughAnAliasedSpellingIsNotWrittenThrough`: `resolvedPath` returned the cleaned spelling and did not resolve links. Exit 1. Exit 0, the outside file was replaced. Resolution was restored.
5. `- master-was-wrong:` —

## Rulings

- Containment at the write uses `os.OpenRoot` on the working directory and `Root.OpenFile` of the relative path, with mode `0o644`. `Root.Create` uses mode `0666`, so it was not the call. Cost if wrong: a caller that wanted the create mode sees `0o644`, which is the mode the outside open already used.
- A path outside the working directory is still opened as before, including `O_NOFOLLOW` on the final component where the system has it. No new `syscall` import. The existing allow-list entry for `internal/cli/output_open_unix.go` stays. Cost if wrong: an outside path can still follow a directory link, which is the operator's path the order leaves in place.
- The load comparison resolves the working directory and the prefix already walked, and does not resolve the component under test. The docs keep the broader sentence and say the spelling is resolved. They were not narrowed to "spelled under the working directory". Cost if wrong: a prefix that cannot be resolved falls back to its cleaned spelling, and a directory that does not exist yet is compared as written.
- Review condition C2 asked paragraph 3 to say that a directory link swapped in after the load check is followed. This order closes that window for a path inside the working directory and says the docs must state what is contained and what is not. Paragraph 3 says an inside path cannot name a file outside, and one sentence says a path outside the working directory is not opened through the root, so a directory link on it is still followed. The condition's sentence would be false for an inside path after this fix.
- A final link that the write's `Lstat` sees is still refused on every path. `os.Root` follows a link that stays inside the root, so a final link to a relative name inside the working directory can be followed if it appears after that `Lstat`. A final link that leaves the working directory is refused by the root. Cost if wrong: that race replaces a different file inside the working directory.
- The same-file checks still run first. Cost if wrong: two outputs that are one file would be named as a root error instead of as the same file. The existing alias tests passed.

## Not run

- `GOOS=windows go vet ./internal/cli/`: permission denial, see item 3.
- `go test -race`: the change adds no goroutine and no shared mutable state.
- A testing-bar metric: none is declared for this repository. No baseline file was found. No bar-weakening move was made (no skip, no loosened assertion, no suppression).
- `proven scan`: the order says not to run it in the checkout.

## NOTICED

- `os.Root` does not stop a traversal that stays inside the root by a bind mount or a device file. The Go 1.26.10 `os.Root` comment says so. Not changed in this round.

## REDACTED

- absolute path, 2: the two `go.mod` lines of the pipeline failure. The mode was 664.
