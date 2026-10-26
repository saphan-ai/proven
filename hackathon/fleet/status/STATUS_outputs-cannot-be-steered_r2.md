# STATUS — outputs-cannot-be-steered · round 2

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: unverified (no pool reading on this seat)
- cost: not recorded in this leg
- duration: not recorded in this leg
- last commit: f44da3dd9454ac70a7e649e01bdb89f70e59a364 "[outputs-cannot-be-steered] cli: refuse directory links in the workdir"
- updated: 2026-10-24T23:57:39Z

## Summary

A symbolic link on a directory of an output path inside the working directory is refused at load with exit 7. The refusal names the flag, the path and the linked component, and nothing is written outside. A final link that appears at the write is still refused. On unix that open uses O_NOFOLLOW. The write refusal is exit 2, and the docs say so.

RED: `go test -count=1 ./...` exits 1. `internal/pipeline` `TestMaterializedTreesAreReadableByOthers` reports `go.mod` mode 664. The same failure was present before this round. This round does not change how those trees are written.

## Acceptance, item by item

1. `git log --oneline dbd98ce..HEAD` at the code commit is one commit: `f44da3d [outputs-cannot-be-steered] cli: refuse directory links in the workdir`. Parent is `dbd98ce`.
2. The reproduction, run before any production edit, exits 1.
   `go test -count=1 -v ./internal/cli -run TestIntermediateDirectoryLinkInsideTheWorkingDirectoryIsRefusedAtLoad`
   Every subtest (`-o`, `--text-output`, `--json-output`, `--sarif-output`, nested directory, absolute path under the working directory): exit 0, want 7; stderr empty; the review ran; the outside file no longer held `kept`.
3. `go vet ./...` exit 0. `gofmt -l .` exit 0, no output. `go test -count=1 ./...` exit 1. 44 packages `ok`, 10 `[no test files]`, 1 fail (`internal/pipeline`, the test named above). `internal/cli` ok. `internal/cli/sarifschema` ok (1.667s). Go is `go1.26.10`. `go.mod` says `go 1.26` and has no toolchain line. `GOOS=windows go vet ./internal/cli/` was not run: the permission class refused the command, reading `GOOS=windows` as the program. Allowed form: none. It was not rebuilt.
4. Unplug, each guard removed and restored:
   - `TestIntermediateDirectoryLinkInsideTheWorkingDirectoryIsRefusedAtLoad`: the walk returned no component. Exit 1. Exit 0, want 7. The review ran. The outside file was replaced.
   - `TestOpenRefusesAFinalLinkSwappedInAfterTheCheck`: `O_NOFOLLOW` removed from the open. Exit 1. The open followed the link and the target was replaced.
5. `- master-was-wrong:` —

## Rulings

- The write-time refusal stays exit 2. The order says to state the exit the write has now. It was not changed to 7. Cost if wrong: a caller that wanted every symlink refusal to be exit 7 still sees 2 when the link appears after load.
- Any symbolic link that is a component inside the working directory is refused, including one whose target is also inside it. Cost if wrong: a link that only renames a directory inside the checkout cannot be used in an output path.
- An intermediate link on a path outside the working directory is not refused. A final link is still refused on every path. Cost if wrong: an absolute path that starts outside the checkout can still follow a directory link.
- `syscall` is on the import allow-list for `internal/cli/output_open_unix.go` only. Go 1.26.10 records `O_NOFOLLOW` as `0x20000` on linux/amd64 and `0x8000` on linux/arm64, so a copied constant would be wrong on the other architecture. Cost if wrong: the allow-list grew by one file.
- The same-file checks still run first, so two outputs that are one file are still refused with both flags named. Cost if wrong: a directory link inside the working directory that makes two outputs one file would name only one flag. The alias tests passed.

## Not run

- `GOOS=windows go vet ./internal/cli/`: permission denial, see item 3.
- `go test -race`: the change adds no goroutine and no shared mutable state.
- A testing-bar metric: none is declared for this repository.
- `proven scan`: the order says not to run it in the checkout.

## NOTICED

- `outputIdentity` still says a dangling final link is followed because `os.WriteFile` would create the target. The write no longer follows that link. The identity is still used by the same-file check. Not changed in this round.
- A directory link swapped in after the load walk is still followed. `O_NOFOLLOW` applies to the final component only. `docs/OUTPUT_FORMATS.md` says that.

## REDACTED

- absolute path, 2: the two `go.mod` lines of the pipeline failure. The mode was 664.
