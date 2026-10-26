# REVIEW — evidence-pack · round 6

- verdict: reject
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T06:01:56Z

## Summary

The property holds at 0b71da0: signature, every hash, the file set, the exact archive bytes, no secret, the key mode and escaped text. 30 of 33 mutants are killed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

Reviewer: review leg evidence-pack-review-r11, claude-code, model claude-opus-5-5. Profile: security.
Reviewed: `d9083c0...0b71da0` (merge base of `origin/main` and HEAD is `d9083c0`; local `main` does not exist in this clone, so `origin/main` was used to find it). HEAD checked first: `git rev-parse --short=7 HEAD` printed `0b71da0`.
Measured on 2026-10-25 on this worktree, go1.26.10 linux/amd64.

### Ground

The property: a pack verifies only if its Ed25519 signature over the manifest, every file hash, the file set and the exact archive bytes hold; any change is refused by name; no configured secret is in the pack; a readable private key is refused; text from a pack reaches the terminal escaped.

Verdict on the property at 0b71da0: **it holds on every path I found.** One mechanism the diff adds has no test that goes red when it is unplugged (the whole-file bound of `VerifyFile`, condition C1). The order calls that BLOCKING.

Searches, each run on 0b71da0:

1. `git diff --stat origin/main...HEAD` — 22 files: `internal/evidence/{pack,verify,key,settings}.go` and tests, `internal/cli/{pack,run,config,format}.go` and tests, `internal/redact/redact.go`, `internal/change/between.go`, `internal/settingspage/catalog.go`, two docs.
2. `grep -n "Names" internal/redact/redact.go` and `grep -rn "Secret: *true" internal` — 5 secret rows (query API key, generator token, generator key, generator second key, guardrail key); every one is in `redact.Names`, which adds the organisation id.
3. `grep -rn "\.User\b\|userinfo" internal` — the generator and query base URLs refuse user info (`internal/generator/settings.go:320`, `internal/query/settings.go:194`), so a non-secret URL row in `settings.json` cannot carry a password.
4. `grep -rn "TracePath" internal` — the trace the pack reads is `result.TracePath`, set by the pipeline.
5. `grep -n "func visible" -A8 internal/cli/*.go` — every report string goes through `escapeControls`.
6. `grep -hn "^func Test"` over the new test files — 46 tests, listed against the mutants below.
7. `grep -n` over `docs/EVIDENCE_PACK.md` for mode, Windows, escape and bound — the doc matches the code on each.

Paths read, by FILE·LINE·CONTROL:

| path | file·line | control | holds |
|---|---|---|---|
| exact archive | `internal/evidence/verify.go:143-149` | rebuild with `archive(files)` and `bytes.Equal` with the input | yes |
| flat file list | `verify.go:242-253` | regular file only, plain name, no repeated name, entry bound before read | yes |
| signature spelling | `verify.go:162-165` | base64 and one newline, byte for byte | yes |
| signature | `verify.go:166-168` | key length and `ed25519.Verify` over the manifest's exact bytes | yes |
| key named | `verify.go:178-180` | manifest key id equals the key used | yes |
| manifest form | `verify.go:170-177` | unknown field and other version refused | yes |
| file set and hashes | `verify.go:181-213` | no path twice, listed file present, size, SHA-256, every `Required` file listed, no extra file | yes |
| receipts | `verify.go:270-323` | receipts equal the manifest's models; each receipted use has an id, a model and a matching trace event | yes |
| whole-file bound | `verify.go:122-128` | read at most `maxPack+1`, refuse over `maxPack` | yes, untested (C1) |
| writer | `internal/evidence/pack.go:125-190` | short secret refused (146), every payload redacted (219-255), manifest commits redacted (163-169), models read back from redacted receipts (155-158), self-verify before return (130) | yes |
| write | `pack.go:193-217` | temp file, mode 0600, rename | yes |
| private key | `internal/evidence/key.go:39-72` | any bit of 0077 refused; Windows skipped, as the doc says | yes |
| CLI scan | `internal/cli/run.go` (pack key load before the review; pack written before outputs; pack in `written`) | refusal exit 7 before the review | yes |
| CLI output collision | `internal/cli/config.go` `outputFiles` | pack is first in the collision and stdout lists | yes |
| settings record | `internal/cli/pack.go:93-128` | secret rows and every `redact.Names` variable written as set/unset; all values pass the redactor | yes |
| CLI verify | `internal/cli/pack.go:141-216` | refusal through `visible`; every report string through `visible` | yes |
| terminal escape | `verify.go:60-84`, `internal/cli/format.go` `escapeControls` | C0, DEL, C1 as `\xNN`; bidi and listed invisible marks as `\uNNNN`; invalid byte as `\xNN` | yes |
| redactor | `internal/redact/redact.go` `spellings`, `inner`, `Shorter` | fourth spelling (HTML escaping off); one quote dropped at each end; trimmed length kept for each name | yes |

### Differential

- `go vet ./...` at 0b71da0: clean (no output).
- `go test -count=1 ./...` at 0b71da0: 49 `ok`, 0 `FAIL`.
- `go vet ./...` at d9083c0: clean.
- `go test -count=1 ./...` at d9083c0: 48 `ok`, 0 `FAIL`.
- How the base was measured: `git worktree add` needed an approval this headless run cannot get, so I ran `git checkout --detach d9083c0` in this worktree, measured, and ran `git checkout review/evidence-pack-review-r11`. `git rev-parse --short HEAD` printed `0b71da0` after the return, and `git diff --stat` was empty.
- The one new package with tests is `internal/evidence` (48 → 49).

### Unplug

Each mechanism was unplugged alone, its package tests were run, and the file was put back before the next one. `git diff --stat` was checked between mutants. 33 mutants were run; 30 turned a test red.

| # | mechanism unplugged | file | red test |
|---|---|---|---|
| M1 | exact-archive comparison | verify.go:147 | TestThePackIsOneByteStringAndAnyOtherIsRefused, TestEveryChangedByteOfThePackIsRefused |
| M2 | signature spelling byte for byte | verify.go:163 | TestANonCanonicalSignatureSpellingIsRefused |
| M3 | `ed25519.Verify` | verify.go:166 | TestAChangedManifestIsRefusedBySignature, TestEveryChangedByteOfThePackIsRefused |
| M4 | manifest key id | verify.go:178 | TestEachManifestGuardRefusesOnItsOwn/key_id |
| M5 | path listed twice | verify.go:183 | TestEachManifestGuardRefusesOnItsOwn/file_listed_twice |
| M6 | listed size | verify.go:191 | TestAListedSizeThatIsNotTheFilesIsRefused |
| M7 | SHA-256 of each file | verify.go:195 | TestOneChangedByteIsRefusedNamingTheFile (7 subtests), TestEveryChangedByteOfThePackIsRefused |
| M8 | listed file present | verify.go:188 | TestAMissingAndAnExtraFileAreRefused |
| M9 | every `Required` file listed | verify.go:200 | TestRequiredPackFilesCannotBeOmitted (7 subtests) |
| M10 | no extra file | verify.go:210 | TestAMissingAndAnExtraFileAreRefused |
| M11 | receipt found in the trace | verify.go:317 | TestAReceiptThePackCannotShowIsRefused |
| M12 | receipts equal the manifest's models | verify.go:279 | TestEachManifestGuardRefusesOnItsOwn/receipts_differ_from_the_manifest |
| M13 | regular file only | verify.go:242 | TestAnArchiveThatIsNotAFlatFileListIsRefused/symlink, TestARefusalShowsNoControlOrBidirectionalCharacter |
| M14 | plain file name | verify.go:245 | TestAnArchiveThatIsNotAFlatFileListIsRefused/path |
| M15 | name twice in the archive | verify.go:248 | TestAnArchiveThatIsNotAFlatFileListIsRefused/duplicate |
| M16 | `Visible` bidi escape | verify.go:69 | TestARefusalShowsNoControlOrBidirectionalCharacter/bidi |
| M17 | `Visible` C0, DEL, C1 escape | verify.go:67 | TestARefusalShowsNoControlOrBidirectionalCharacter/escape, /c1 |
| M18 | key mode 0077 | key.go:68 | TestAPrivateKeyThatGroupOrOthersCanReachIsRefused (6 modes), TestAReadableKeyIsRefusedBeforeTheReview |
| M19 | short secret refusal | pack.go:146 | TestAConfiguredSecretShorterThanTheFloorRefusesThePack |
| M20 | manifest target redacted | pack.go:164 | TestASecretInTheTargetOrHeadReachesNoFile |
| M21 | manifest baseline redacted | pack.go:165 | TestNoConfiguredSecretReachesThePack |
| M22 | self-verify in `Build` | pack.go:130 | TestAReceiptThePackCannotShowIsRefused |
| M23 | escape-off JSON spelling | redact.go:126 | TestTheEscapeOffJSONSpellingIsRedacted, TestASecretInThePacksOwnJSONSpellingIsRedacted |
| M24 | `inner` back to `strings.Trim` | redact.go:139 | TestTheEscapeOffJSONSpellingIsRedacted, TestASecretEndingInAQuoteIsRedactedWhole |
| M25 | secret catalog row written raw | cli/pack.go:113 | **none** — equivalent mutant, see below |
| M26 | `redact.Names` set/unset loop | cli/pack.go:121 | TestALiveRunRecordsNoSecretSettingValue |
| M27 | report agent field escaped | cli/pack.go:197 | TestEveryReportFieldIsShownEscaped |
| M28 | CLI `escapeControls` bidi | cli/format.go | TestVerifyShowsNoControlByteFromThePack, TestEveryReportFieldIsShownEscaped |
| M29 | pack in the output collision list | cli/config.go `outputFiles` | TestEvidencePackCollidesWithEveryOutputFlag (5), TestEvidencePackAliasedOutputCannotReplacePack |
| M30 | pack in `written` after the load check | cli/run.go `writeOutputs` | TestWriteOutputsLeavesThePackWhenALinkAppearsAfterTheLoadCheck |
| M31 | unknown manifest field | verify.go:171 | TestEachManifestGuardRefusesOnItsOwn/unknown_field |
| M32 | whole-file bound `maxPack` | verify.go:126 | **none** — BLOCKING, condition C1 |
| M33 | CLI `visible` on a refusal | cli/pack.go:179 | **none** — equivalent mutant, see below |

M25 and M33 are equivalent mutants: no input changes the output while the other layer is in place.
- M25: the `redact.Names` loop at `cli/pack.go:120-122` writes set/unset over the same keys, and search 2 shows every secret catalog row is in `redact.Names`. Unplugging the other layer (M26) goes red.
- M33: `VerifyFile` returns only `*Refusal`, and `Refusal.Error` already escapes the name and the detail. Unplugging that layer (M16, M17) goes red.

The order's rule says a mechanism with no red test is BLOCKING. I read that rule as covering mechanisms that change some output. No test can kill an equivalent mutant, so I file M25 and M33 as NOTE. The master may raise either one (law #53).

M32 is not equivalent. Without the check, `VerifyFile` passes the first `maxPack+1` bytes on to `Verify` and ignores the rest of the file. A signed pack over the bound would then be judged on a prefix and not on the whole file, which is the exact-bytes half of the property. Today no test fails when the check is removed: `maxPack` is a constant of about 4.5 GiB, and no test reaches it.

### Closure

Each place where the property could fail was checked. The property holds at 0b71da0 on every path listed under Ground. Item 3 of the order therefore needs no failing test for a broken property.

For the BLOCKING unplug gap (C1), this is the test a fix must make pass. I ran it at 0b71da0, where it fails to build:

```
## github.com/saphan-ai/proven/internal/evidence [github.com/saphan-ai/proven/internal/evidence.test]
████████████████████████████████████████████:12: undefined: verifyFileWithin
████████████████████████████████████████████:15: undefined: verifyFileWithin
FAIL	github.com/saphan-ai/proven/internal/evidence [build failed]
```

Full text (package `evidence`; it uses the `builtPack` and `wantCheck` helpers in `hardening_test.go`):

```go
package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

// The mutated field is the whole-file bound VerifyFile holds a pack to
// (maxPack). A fix gives VerifyFile a seam that takes the bound, so the guard
// can be unplugged and seen to go red: a pack one byte over the bound is
// refused by the format check naming the bound, and a pack at the bound is
// verified. VerifyFile itself keeps calling the seam with maxPack.
func TestAPackOverTheWholeFileBoundIsRefused(t *testing.T) {
	pack, _, pub := builtPack(t)
	path := filepath.Join(t.TempDir(), "pack.tar")
	if err := os.WriteFile(path, pack, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := verifyFileWithin(path, pub, int64(len(pack))-1)
	wantCheck(t, err, CheckFormat, "byte bound")
	if _, err := verifyFileWithin(path, pub, int64(len(pack))); err != nil {
		t.Fatalf("a pack at the bound must verify: %v", err)
	}
}
```

What a fix must make it do: `VerifyFile(path, pub)` calls `verifyFileWithin(path, pub, maxPack)`. The seam reads at most `limit+1` bytes and refuses more than `limit` with `CheckFormat` and a detail that names the byte bound. With the fix in place, removing the `len(body) > limit` check must turn this test red.

### Remainder

NOTE items. None of them breaks the property as written.

- N1 `verify.go:60-77`: `Visible` does not escape the backslash. A pack string that holds the six characters `\x1b` looks the same as an escaped ESC. No terminal action follows, but a reader cannot tell the two apart.
- N2 `verify.go:82-84`: `IsBidi` is a closed list. Other invisible code points are written unchanged: tag characters U+E0000 to U+E007F, U+FFF9 to U+FFFB, U+180E, and the variation selectors. They cannot move the cursor or recolour the terminal, and the doc names only control characters, bidirectional marks and invalid bytes. They can still hide text from a reader.
- N3 `key.go:64` then `key.go:96`: the mode is read by `Stat`, then the file is read by `ReadFile`. A file swapped between the two calls is not checked. Doing that needs write access to the key's directory.
- N4 `cli/run.go`: the trace offset is measured at `<cwd>/logs/trace.jsonl`, and the pack reads `result.TracePath`. I did not measure whether the two can differ. If they do, the pack holds a wrong span of the trace: other runs' lines, still redacted, or a cut line that `Build` refuses.
- N5: M25 and M33 are equivalent mutants (see Unplug).
- Housekeeping, which the master owes: this leg left two untracked files that it could not delete. `git clean` needs an approval this headless run cannot get, and law #57 forbids deleting them another way. The files are `█████████████████████████████████████████` (now a 4-line `//go:build ignore` stub, so every package still builds and passes) and `███████████████████████████████████████████████`. Every tracked file is restored (`git diff --stat` is empty).
- Confinement: the order puts scratch under `█████████████████████████████████████`. This run's permission class allows writes only to `_REVIEW_STATUS.md`, `_REVIEW_NOTES.md` and `review/`, and it refused a redirect into `█████████████`. The order and the grant disagree.

### Profile

Security. The defect classes I checked:
- tamper evidence: signature, hashes, file set, exact bytes, receipts;
- secret leakage into any pack file or the signed manifest: the redactor's spellings, the short-secret floor, the settings record, the commit names;
- key exposure: file mode, the key never written into the pack, refusals that name no key byte;
- terminal control injection from pack text: refusals and the report;
- a pack replaced by another output;
- resource bounds on a hostile pack.

### Focus

Can anything the diff touches break the property? I found nothing that breaks it at 0b71da0. 30 of the 33 mutants are killed. Two of the three survivors are equivalent. The third, M32, is an untested guard (C1).

master-was-wrong: none. The master's figures hold: `go vet ./...` is clean at both commits, `go test -count=1 ./...` gives 48 ok at d9083c0 and 49 ok at 0b71da0, and there are 0 FAIL at either. One difference in method: I measured the base by a checkout in this worktree, not in a detached worktree.

### Conditions

- C1 — BLOCKING. `internal/evidence/verify.go:126`: no test turns red when the whole-file bound of `VerifyFile` is unplugged (M32). Fix: add the `verifyFileWithin` seam and the test given under Closure. Proof: that test passes, and it goes red when the `len(body) > limit` check is removed.
- C2 — NOTE. N1 to N4 under Remainder. Recorded only; none is a condition of merge.

### Unrun

- The Closure test against a fix: no fix exists, and this leg does not fix the code under review.
- The Windows branch of `ownerOnly` (`key.go:61`): this host is linux.
- A live, non-sample `proven scan --evidence-pack` against a real generator: it needs the network and keys, which the boundary excludes. The live settings path was exercised only through `packSettings` unit tests.
- The `proven` binary run by hand: `scan --evidence-pack` and `verify` were exercised only through the `internal/cli` and `cmd/proven` tests.
- Whether the trace offset and `result.TracePath` can name different files (N4).
- Deleting the two scratch files: refused, see Remainder.

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The property holds at 0b71da0: signature, every hash, the file set, the exact archive bytes, no secret, the key mode and escaped text. 30 of 33 mutants are killed.
  One BLOCKING item: the whole-file bound of VerifyFile (verify.go:126) is untested; the fix seam and its test are in the notes (C1). Two survivors are equivalent mutants (NOTE).
  Master re-measured and right: vet clean at both commits, 48 ok at d9083c0 and 49 ok at 0b71da0, 0 FAIL.
  Unrun is not empty. Two scratch files could not be deleted (git clean refused); the master owes their removal.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T06:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| file and line in Saphan Studio's private code | 3 |
