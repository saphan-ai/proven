# REVIEW — evidence-pack · round 4

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T00:41:11Z

## Summary

BLOCKING: a colliding output option overwrites the signed pack after it is verified; the scan returns success. A local sample-mode test fails at 5856357. Refuse pack/output collisions before review and preserve the pack at write time.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground
RED HEADER — BLOCKING: A successful scan can overwrite its signed evidence pack with a report when `--evidence-pack` and `-o` name the same file; the expected-behavior test fails at the reviewed SHA. See Differential.

At reviewed SHA 5856357 (2026-10-25), `git rev-parse --short=7 HEAD` returned `5856357`; `git merge-base d9083c0 HEAD` returned d9083c034cb0d297933c4acedbc99eac9b8c1d22. This checkout has no local `main` ref; `d9083c0...HEAD` is the equivalent reviewed diff. `git diff --name-only d9083c0...HEAD` listed 21 paths: docs/CONFIGURATION.md, docs/EVIDENCE_PACK.md, internal/change/between.go, internal/change/resolve_test.go, internal/cli/{config.go,format.go,pack.go,pack_hardening_test.go,pack_round4_test.go,pack_test.go,run.go}, internal/evidence/{hardening_test.go,key.go,pack.go,pack_test.go,round4_test.go,settings.go,verify.go}, internal/redact/{redact.go,shorter_test.go}, and internal/settingspage/catalog.go.

Searches used to follow the changed paths: `rg --files`, `rg -n 'func Test|^// The mutated' internal/evidence/*test.go internal/cli/pack*test.go`, `rg -n 'func writable|evidencePack|func refuseStdoutOnAnOutput|output.*same|trace' internal/cli/config.go internal/cli/run.go internal/cli/pack*.go`, `rg -n 'outputIdentity|evidencePack|output paths|duplicate|same file' internal/cli/config.go`, `rg -n 'func visible|Visible\\(|fmt\\.Fprint|pack' internal/cli/format.go internal/cli/run.go internal/cli/pack.go`, and `rg -n 'func .*Visible|IsBidi|Visible\\(' internal/evidence internal/cli/*test.go`. Read the pack writer, verifier, key loader, CLI config and writer, redactor and terminal renderer. `go.mod` pins Go 1.26; local `go version` is go1.26.5. The Makefile has build/lab commands, no generic unit-test target.

Tip measurement (before the temporary probe): `go vet ./... && go test -count=1 ./...` passed; 49 package rows reported `ok`, zero `FAIL`. Base measurement: `git archive d9083c0 | tar -xf - -C █████████████████████████████████████████` followed by `(cd █████████████████████████████████████████ && GOTOOLCHAIN=local go vet ./... && GOTOOLCHAIN=local go test -count=1 ./...)` exited 0, with 48 package rows `ok`, 10 `no test files`, zero `FAIL`; capture: `███████████████████████████████████████████████████`. The difference is one added passing package row, not a count of individual tests.

### Differential
RED HEADER — BLOCKING: `--evidence-pack` is excluded from output-file collision checks. FILE·LINE·CONTROL: internal/cli/config.go:453 lists output paths but omits the pack; internal/cli/config.go:394 invokes collision checking before assigning `cfg.evidencePack` at line 433; internal/cli/run.go:128 writes the signed pack before `writeOutputs` at line 134, which writes `-o` with `os.WriteFile` (internal/cli/run.go:299). The same path for `--evidence-pack` and `-o` therefore overwrites the signed pack with ordinary text and returns exit 0. An auditor cannot verify the promised pack. There is no guard on aliases or already-existing same files between the pack and ordinary outputs either. Fix must refuse pack/output collisions before the review starts (including aliased paths) and keep the pack intact; check again at write time if a symlink can change during the review.

At the tip, a test of expected behavior (full text below) printed `colliding output was accepted: ""` and failed. The test wrote no attack artifact and used only a temporary local sample-mode review.

```go
package cli

import (
 "bytes"
 "os"
 "path/filepath"
 "testing"

 "github.com/saphan-ai/proven/internal/evidence"
)

func TestEvidencePackOutputCannotReplacePack(t *testing.T) {
 target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
 t.Chdir(t.TempDir())
 privateKey, publicKey := packKeys(t)
 packPath := filepath.Join(t.TempDir(), "review.pack")
 var stdout, stderr bytes.Buffer
 exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, "-o", packPath, target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
 if exit == exitOK {
  t.Fatalf("colliding output was accepted: %q", stderr.String())
 }
 if body, err := os.ReadFile(packPath); err == nil {
  public, err := evidence.LoadPublicKey("test", publicKey)
  if err != nil { t.Fatal(err) }
  if _, err := evidence.Verify(body, public); err != nil { t.Fatalf("a written pack must remain verifiable: %v", err) }
 }
}
```

The expected RED is the named `TestEvidencePackOutputCannotReplacePack` assertion, not a compilation error. It is a verification test only; it does not implement a fix. The temporary test file is removed before return. Its captured failing output is `███████████████████████████████████████████████████████`.

### Unplug
Measured by Go `-overlay` source substitutions kept only under `█████████████████████████████████████████████` (not source edits), each with `go test -overlay <overlay.json> -count=1 -run <named test> -v ./internal/evidence`:
- Signature verification removed: `TestAChangedManifestIsRefusedBySignature` turns red (wanted signature refusal, got nil).
- Canonical archive equality removed: `TestThePackIsOneByteStringAndAnyOtherIsRefused` turns red on appended bytes, changed metadata and padding.
- Per-file hash comparison disabled: `TestOneChangedByteIsRefusedNamingTheFile` turns red for all seven payload files.
- Required-file guard disabled: `TestRequiredPackFilesCannotBeOmitted` turns red for required files (the result can be a different downstream format/receipt refusal; the test demands the missing-file refusal).
- Private-key mode guard disabled: `TestAPrivateKeyThatGroupOrOthersCanReachIsRefused` turns red for group/other-readable modes.
- Secret replacement disabled: `TestNoConfiguredSecretReachesThePack` turns red (pack holds the configured value).
- Invisible-mark escaping disabled: `TestARefusalShowsNoControlOrBidirectionalCharacter` turns red.
- Pack/output path collision: no existing test turns red because there is no guard to unplug; the new expected-behavior test turns red on the unmodified tip. This is the blocking unplug gap.

The initial canonical mutation yielded only a compile error (unused variable) and does not count; the corrected mutation above compiled and turned its named test red. The first required-file and key-mode substitutions made no effective change and do not count; corrected substitutions above did.

### Closure
Reject at 5856357: the declared per-review pack can be erased by a legitimate pair of output options while the command reports success. Signature, hash, file set and archive-byte guards have meaningful negative controls for packs that reach the verifier. The defect is in CLI output lifetime and is not closed by those verifier tests.

### Remainder
Offline receipt authenticity is explicitly not checked by the verifier; docs/EVIDENCE_PACK.md states the limitation. Sample-mode review and pack-output collision are in scope; hosted generator, an external auditor's file handling, and Windows permission behavior were not exercised. No product code was changed or committed in this review. The review order prohibits committing the probe or publishing a finding outside this review document.

### Profile
Security: the signing key mode check, hash, signature, canonical archive, required files, redaction and terminal-visible controls were inspected and unplugged as above. No credential was printed in the test captures. The pack/output collision invalidates availability of signed evidence rather than forging a signature. Scans for bidi/invisible instructions in the diff were limited to changed Go paths; no other independent finding is asserted.

### Focus
master-was-wrong: none on the clean tip suite (the master's quoted counts agree only insofar as package-row counts, not individual test counts, were measured). The baseline reading is separately recorded under Ground. The affirmative claim "a pack can be left for each review" is contradicted by the differential test at the reviewed SHA.

### Conditions
Blocking: refuse `--evidence-pack` collisions with `-o`, `--text-output`, `--json-output`, `--sarif-output` and their aliases before starting a review; preserve the signed pack if the filesystem changes after preflight. A test must assert a refusal and that no unverified overwritten pack is left. No condition is downgraded by this reviewer.

### Unrun
External hosted review and real remote receipts: no network outside local Go tests permitted. Windows key permissions: this seat is Linux. Full suite with the probe present: the named probe was run separately and removed. No test asserts the external origin of a signed receipt; that is out of offline verification scope, not a pass.

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: a colliding output option overwrites the signed pack after it is verified; the scan returns success. A local sample-mode test fails at 5856357. Refuse pack/output collisions before review and preserve the pack at write time.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T00:39:03Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 5 |
