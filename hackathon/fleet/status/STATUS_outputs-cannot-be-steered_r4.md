# STATUS — outputs-cannot-be-steered · round 4

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: unverified (no pool reading on this seat)
- cost: not recorded in this leg
- duration: not recorded in this leg
- last commit: 6435e0881c84d714ff6f282c64a996dc4347ed66 "[outputs-cannot-be-steered] cli: name which root calls renameat"
- updated: 2026-10-25T07:26:59Z

## Summary

An output path inside the working directory is written as a new file in the same directory, through that directory opened as a root, and that file is renamed over the final name. The final name is replaced, never written through. A symbolic link swapped in as that name after the final-component check does not receive the report in the file it names. A final link the check still sees is refused, as before. A directory link that leaves the working directory is still refused by the root.

RED: `go test -count=1 ./...` exits 1. `internal/pipeline` `TestMaterializedTreesAreReadableByOthers` reports `go.mod` mode 664. The same failure was present before this round. This round does not change how those trees are written.

## Acceptance, item by item

1. `git log --oneline b2d5f51..HEAD` at the code tip is two commits: `1053dc3 [outputs-cannot-be-steered] cli: replace the final output name` and `6435e08 [outputs-cannot-be-steered] cli: name which root calls renameat`. Parent of `1053dc3` is `b2d5f51`. This status commit follows them.
2. The review seam, run on `b2d5f51` before any production edit, does not fail.
   `go test -count=1 -v ./internal/cli -run TestFinalComponentSwappedForALinkInsideTheCheckoutDoesNotChangeTheOtherFile`
   Exit 0. The log line: `exit 2; stderr --sarif-output proven.sarif: final component is a symbolic link; the file it names is not written`. The other file still held `kept`. The swap ran during the review, which is before `writeOutputFile`'s `Lstat`, so the check refused the link and the rooted open was not reached. That test name is not in the tree; the regression below replaces it.
   The after-check regression, with the rooted open still `OpenFile` on the final name, exits 1.
   `go test -count=1 -v ./internal/cli -run TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile`
   `--- FAIL: TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile (0.00s)`. The other file held the SARIF log, including `"class": "c"`.
   The same command after the rename exits 0. `--- PASS: TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile (0.00s)`.
3. `go vet ./...` exit 0. `gofmt -l .` exit 0, no output. `go test -count=1 ./...` exit 1. 44 packages `ok`, 10 `[no test files]`, 1 fail (`internal/pipeline`, the test named above). `internal/cli` ok (39.111s). `internal/cli/sarifschema` ok (3.512s). Go is `go1.26.10`. `go.mod` says `go 1.26` and has no toolchain line. No Makefile and no `.github` workflow were found; the suite above is the order's acceptance command. `GOOS=windows go vet ./internal/cli/` was not run: the permission class refused the command, reading `GOOS=windows` as the program. Allowed form: none. It was not rebuilt.
4. Unplug, the rename replaced by `root.OpenFile` of the final name and then restored:
   `TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile` exit 1. The other file held the SARIF log. The rename was restored. The same command then exited 0. The `internal/cli` package then exited 0 (20.737s) before the full suite.
5. `- master-was-wrong:` the review seam does not show the follow. On `b2d5f51` a final component swapped for a relative link during the review is refused by `Lstat` with exit 2, and the other file is unchanged. The follow the review measured is the window after that `Lstat`: `Root.OpenFile` of a relative link that stays inside the checkout writes the report onto the other file. The committed regression swaps in that window. The review condition asked for that after-check control.

## Rulings

- The inside write creates `.proven-output.` plus 16 hex digits in the same directory, with `O_WRONLY|O_CREATE|O_EXCL` and mode `0o644`, then `Root.Rename` over the final name. Eight attempts. A failed write or rename removes the new file. Cost if wrong: eight existing names of that shape fail the write; a removal that fails is included in the error and the name can remain.
- `beforeRootedOutputCreate` is nil except in the regression, which replaces the final component after the check. `t.Cleanup` clears it. The package has no `t.Parallel`. Cost if wrong: a test that leaves the hook set can swap a later test's output.
- The regression also requires exit 0, a regular file at the final name, and the report in that file. A refusal leaves the other file unchanged and would not go red when the final name is opened. The `Lstat` refusal of a link seen before the create stays, including a swap during the review. Cost if wrong: a refusal that leaves the link in place fails this test.
- The load refusal and the outside open, including `O_NOFOLLOW` where the system has it, are unchanged. No new `syscall` import and no new dependency. Cost if wrong: an outside path can still follow a directory link, which is the operator's path the order leaves in place.

## Findings

- Review condition, final in-checkout link at the rooted open: ADDRESSED. The rename does not open the final name. `TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile` fails when that name is opened with `OpenFile` and passes after the rename.
- Order step 1, swap during the review: the red does not appear. Recorded under master-was-wrong. The check already refuses that link.

## Not run

- `GOOS=windows go vet ./internal/cli/`: permission denial, see item 3.
- `go test -race`: the change adds no goroutine and no shared mutable state.
- A testing-bar metric: none is declared for this repository. No baseline file was found. No bar-weakening move was made (no skip, no loosened assertion, no suppression).
- `proven scan`: the order says not to run it in the checkout.
- The full suite was run at `1053dc3`, before the comment-only commit `6435e08`. The single regression was run again at `6435e08` and exited 0. `go vet ./...` was run at `1053dc3` and exited 0.

## NOTICED

- none

## REDACTED

- absolute path, 4: two `go.mod` lines of the pipeline failure, and the `kept.txt` path in the reproduction fatal and in the unplug fatal. The mode was 664. The other file held the SARIF log.
