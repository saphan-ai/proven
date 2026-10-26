# REVIEW — evidence-pack · round 7

- verdict: accept
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 2 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T06:44:38Z

## Summary

C1 is closed at bb1c3cc. With the bound check unplugged, TestAPackOverTheWholeFileBoundIsRefused goes red, and the test also kills the >= and missing-+1 mutants.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground
- Reviewed at `bb1c3cc` (`git rev-parse --short=7 HEAD` printed `bb1c3cc`). Scope of the re-read: `0b71da0..bb1c3cc`, two files: `internal/evidence/verify.go` (+9 −3) and the new `internal/evidence/bound_test.go` (+25).
- Measured on 2026-10-25 in this worktree, go1.26.10 linux/amd64:
  - `go vet ./...`: clean.
  - `go test -count=1 ./...`: 49 packages `ok`, 0 `FAIL`, 0 `panic`.
  - `go test -count=1 -run 'TestAPackOverTheWholeFileBoundIsRefused|TestVerifyFile' -v ./internal/evidence/`: both PASS.

### Differential
- `verify.go:116-118`: `VerifyFile` now only calls `verifyFileWithin(packPath, pub, maxPack)`. The public signature has not changed. Its one production caller, `internal/cli/pack.go:177`, has not changed.
- `verify.go:122-136`: the body moved into `verifyFileWithin` with `limit int64` in place of the constant `maxPack`. `io.LimitReader(f, limit+1)` and `int64(len(body)) > limit` behave the same as before for `limit = maxPack`. The refusal still has check `format` and detail `larger than the <n> byte bound`. The `<n>` is a constant, not pack text, and `Refusal.Error` escapes the detail anyway (`verify.go:48-54`).
- `Verify` (`verify.go:141` and after) has not changed. The signature, every file hash, the file set, the canonical-archive byte check, secret redaction, the private-key check and the terminal escaping are not in this diff.
- `bound_test.go`: holds the built pack to `len(pack)-1` and expects a `format` refusal with `byte bound`. It then holds the same pack to `len(pack)` and expects it to verify. Helpers `builtPack` and `wantCheck` come from `hardening_test.go:56-75` and have not changed.

### Unplug
Each mutation was made on its own and reverted with `git checkout -- internal/evidence/verify.go`. Tests run: `go test -count=1 ./internal/evidence/` (and `./internal/cli/` for M1 and M5).
- M1, the C1 unplug: `if false && int64(len(body)) > limit` (check removed). RED: `TestAPackOverTheWholeFileBoundIsRefused`, `bound_test.go:21: want a format refusal, got <nil>`. No other test went red; `internal/cli` stayed ok.
- M2: `>` changed to `>=`. RED: same test, `bound_test.go:23: a pack at the bound must verify: ... larger than the 12800 byte bound`.
- M3: `LimitReader(f, limit)` without the `+1`. RED: same test, `bound_test.go:21: refusal is format "not a readable tar archive: unexpected EOF", want format with "byte bound"`.
- M4: `io.ReadAll(f)` with no `LimitReader`. SURVIVED. The decision is still right, but the read is no longer bounded in memory before the check.
- M5: `VerifyFile` passes `64*maxPack` instead of `maxPack`. SURVIVED in `internal/evidence` and `internal/cli`.

### Closure
- C1 of the claude review of 0b71da0 (`verify.go:126`, the whole-file bound of `VerifyFile` untested): CLOSED at `bb1c3cc`. The `verifyFileWithin` seam exists (`verify.go:122`). `TestAPackOverTheWholeFileBoundIsRefused` passes, and it goes red when the `len(body) > limit` check is removed (M1 above). It also pins the comparison's boundary (M2) and the error's name (M3).

### Remainder
- NOTE (not blocking, not introduced by this diff): two mutants survive. M5: no test pins that `VerifyFile` passes `maxPack` to the seam. The wiring is one line at `verify.go:117`, and a reader can check it. The test's comment says "VerifyFile calls the seam with maxPack", but the test does not check this. M4: no test pins that the read is bounded before the whole file is held in memory. Both are about the resource bound, not about which packs verify. Removing them would not let a changed pack verify.
- NOTE: the switch to `int64(len(body)) > limit` may also change how this package builds for 32-bit targets. I did not measure this: the review role's command allow-list refused an inline `GOARCH=386` prefix. It is outside the property and outside what the order asked.

### Profile
security. The property is that a pack verifies only if its Ed25519 signature, every file hash, the file set and the exact archive bytes hold. Any change is refused by name. No configured secret is in the pack. A readable private key is refused. Text from a pack reaches the terminal escaped. This diff touches only the whole-file read bound in front of `Verify`. It does not change the signature, hash, file-set, canonical-bytes, secret, key-permission or escaping code paths.

### Focus
- Can anything the diff touches break the property? No. Every byte `verifyFileWithin` passes to `Verify` still goes through the unchanged `Verify`. A pack that is cut off at the bound is refused by name. A pack over the bound is refused as `format ... byte bound`. M1 to M3 are killed. The only new parameter, `limit`, is not exported, and its one production value is the constant `maxPack`.
- master-was-wrong: none. I re-measured every claim the master made. `go vet ./...` is clean. `TestAPackOverTheWholeFileBoundIsRefused` passes and fails with the check disabled. `go test -count=1 ./...` gives 49 ok and 0 FAIL. All three hold.

### Conditions
none

### Unrun
none

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: C1 is closed at bb1c3cc. With the bound check unplugged, TestAPackOverTheWholeFileBoundIsRefused goes red, and the test also kills the >= and missing-+1 mutants.
  Nothing in 0b71da0..bb1c3cc weakens the pack's property. Verify did not change; the seam only makes the whole-file bound a parameter.
  Re-measured: go vet clean, go test ./... 49 ok and 0 FAIL. The master's claims hold.
  NOTE: two resource-bound mutants survive (no LimitReader; VerifyFile passing a larger bound). Neither lets a changed pack verify.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T06:45:00Z

## REDACTED

none
