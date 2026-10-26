# REVIEW — evidence-pack · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T13:30:03Z

## Summary

Two blocking defects: a verified pack may omit required payload files, and a short configured secret may enter the signed pack. Focused tests reproduce both. The whole-module run has one unrelated red; base comparison was unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground

Reviewed commit `3b6bac51beb71f69bcd7a7ab6c685ce374eac6ad` (measured with `git rev-parse --short=7 HEAD`; short output `3b6bac5`). The feature delta was read with `git diff c1e4ba8..HEAD --name-only` and `git diff c1e4ba8..HEAD -- internal/evidence/pack.go internal/evidence/verify.go internal/evidence/key.go internal/cli/pack.go`. The requested `main...HEAD` name was unavailable locally; `origin/main...HEAD` includes other wagons, so the feature's parent was used to isolate this review. Search paths: `rg -n 'evidence|pack|verify|manifest|archive' internal cmd --glob '*.go'`; `rg -n 'evidencePack|writePack|visible\(|escapeControls|Secret|Redactor|TracePath' internal/cli internal/review internal/pipeline internal/redact --glob '*.go'`; `rg -n 'EnvToken|EnvKey|MinLength|len\(.*[Tt]oken|len\(.*[Kk]ey' internal/generator internal/query internal/guardrail internal/review`; and `rg -n '^func Test' internal/evidence/pack_test.go`.

The module pins Go 1.26; the seat ran Go 1.26.5. No Makefile or `.github` CI directory was present. `go test -count=1 ./internal/evidence ./internal/cli` passed both packages. `go vet ./internal/evidence ./internal/cli` passed. `go test -count=1 ./...` failed only in `internal/pipeline/TestMaterializedTreesAreReadableByOthers`: two `go.mod` files had mode 0664. That test and the tree materializer are outside this feature delta. A base suite was not run, so whether this red is also present on main is unverified. The feature's two new regression tests below were run through `go test -overlay █████████████████████████████████████████████████ -count=1 -run '^TestReview(RequiredPackFilesCannotBeOmitted|ConfiguredShortSecretNeverEntersPack)$' -v ./internal/evidence`; both failed for the expected reasons.

No coverage, mutation-score, lint-warning or type-warning baseline with a command is declared. No numeric bar was judged. Ranking of available checks: Go vet is a project/tool check; the Go suite and this review's overlay tests are suite checks; the isolated source mutants below are stronger suite checks. No test skip, assertion weakening, suppression or threshold change was found in the feature delta. No dependency manifest or lockfile changed in the feature delta.

### Differential

**BLOCKING 1 — required payload files can disappear from a verified pack.** `internal/evidence/verify.go:160-182` treats the signed manifest's list as the complete file contract and only rejects archive entries absent from that list. `settings.json` and `commits.json` are required by `docs/EVIDENCE_PACK.md:10-22`, but `checkReceipts` and `describe` never read either. A pack with either file omitted from both the archive and a freshly signed manifest returns success from `Verify`. The overlay test below measured both cases at this commit: `a pack without required file settings.json verified` and the same for `commits.json`. Control: enforce the format's seven required payload names before returning success, then keep the list/archive equality and hash checks. The signature and every listed hash still hold in the failing test; it isolates the required-file invariant.

**BLOCKING 2 — an accepted configured secret can enter the signed pack.** `internal/evidence/pack.go:204-239` depends on `Redactor.Text` to remove secret values. `internal/redact/redact.go:54-67` ignores values shorter than six bytes, while `internal/generator/settings.go:223` and the other secret settings accept nonempty values of that length. A result containing a configured five-byte value is signed and returned by `Build`; the overlay test below measured `a configured secret is present in the signed pack`. Control: refuse such a secret before the review or redact it in every pack field, with a regression test that checks pack bytes and verifies the resulting pack.

Full text of the two failing Go tests (package `internal/evidence`, using existing same-package fixtures):

```go
package evidence

import (
    "bytes"
    "crypto/ed25519"
    "encoding/base64"
    "encoding/json"
    "testing"

    "github.com/saphan-ai/proven/internal/redact"
)

func TestReviewConfiguredShortSecretNeverEntersPack(t *testing.T) {
    in, pub := sampleInput(t)
    const secret = "abcde"
    in.Redactor = redact.New(map[string]string{"███████████": secret})
    in.Result.Findings[0].Message = "service returned " + secret
    pack, err := Build(in)
    if err != nil { t.Fatal(err) }
    if bytes.Contains(pack, []byte(secret)) {
        t.Fatal("a configured secret is present in the signed pack")
    }
    if _, err := Verify(pack, pub); err != nil {
        t.Fatalf("the secret-free pack must verify: %v", err)
    }
}

func TestReviewRequiredPackFilesCannotBeOmitted(t *testing.T) {
    pack, in, pub := builtPack(t)
    for _, omitted := range []string{SettingsName, CommitsName} {
        t.Run(omitted, func(t *testing.T) {
            files, err := readArchive(pack)
            if err != nil { t.Fatal(err) }
            var m Manifest
            if err := json.Unmarshal(files[ManifestName], &m); err != nil { t.Fatal(err) }
            refs := m.Files[:0]
            for _, ref := range m.Files {
                if ref.Path != omitted { refs = append(refs, ref) }
            }
            m.Files = refs
            delete(files, omitted)
            manifest, err := marshal(m)
            if err != nil { t.Fatal(err) }
            files[ManifestName] = manifest
            files[SignatureName] = []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(in.Key, manifest)) + "\n")
            changed, err := archive(files)
            if err != nil { t.Fatal(err) }
            if _, err := Verify(changed, pub); err == nil {
                t.Fatalf("a pack without required file %s verified", omitted)
            }
        })
    }
}
```

### Unplug

The source guards were each replaced in a Go overlay under `█████████████████████████████████████`, then the named test was run with `-count=1`. Each returned exit 1 with a named assertion failure, not a compile error: canonical archive equality → `TestThePackIsOneByteStringAndAnyOtherIsRefused`; Ed25519 verify → `TestAChangedManifestIsRefusedBySignature`; file-hash comparison → `TestOneChangedByteIsRefusedNamingTheFile`; extra-file check → `TestAMissingAndAnExtraFileAreRefused`; receipt/trace match → `TestAReceiptThePackCannotShowIsRefused`; private-key mode check → `TestAPrivateKeyThatGroupOrOthersCanReachIsRefused`; terminal bidi escape → `TestVerifyShowsNoControlByteFromThePack`; payload redaction → `TestNoConfiguredSecretReachesThePack`. The hash mutant's `commits.json` and `settings.json` subtests were accepted, while other entries returned format or receipt refusals; this confirms the hash check is needed for those files. The unplug script and overlays remain inside the confined work directory; cleanup was refused by the command guard. No production source file was edited.

The new short-secret and required-file tests are red on the unmodified tip, so no unplug is needed to establish those defects. The required-file test mutates only the signed file list and matching archive membership. The short-secret test mutates only one configured value and one result message.

### Closure

Verdict: **reject**. Both findings can yield a successful signed pack while a stated evidence-pack guarantee is false. Repair both and rerun the two tests, the existing pack tests, and the full suite. The reviewer made no product-code change and no commit.

### Remainder

`internal/evidence/verify.go:142-162` does not compare a signed `FileRef.Size` to the actual bytes. The SHA-256 still binds the bytes, so this is a metadata accuracy question rather than a demonstrated bypass of the requested property. The master should decide whether the size field is contractual. `readFrom` in `internal/cli/pack.go:45-53` silently returns an empty trace on read failure; the review did not prove a normal `scan` can reach this state, so this is an unverified risk, not a finding. Declined to judge: OpenSSL key lifecycle, provider receipt signatures, storage retention, and unrelated wagons in `origin/main...HEAD`.

### Profile

Trust boundaries read: unsigned tar headers and entries into `readArchive`; signed manifest and public key into `Verify`; review result and environment-backed redactor into `Build`; private-key file into `LoadPrivateKey`; pack strings into CLI terminal output. Tampering with archive bytes, manifest bytes, listed payload hashes and an extra file is refused by the named tests and unplug controls above. Private-key group/other access is refused by the mode test. A five-byte configured secret crosses into the pack, and required file omission crosses the format boundary. No dependency change was introduced by this feature. No invisible or bidirectional character was observed in the feature source; the terminal escape mutant was killed by its test. No instruction-like comment was found in the feature source.

### Focus

The diff touches two ways to break the requested property: the verifier can accept a missing required evidence file, and the pack can carry a short configured secret. `master-was-wrong:` not established for the master's whole-module claim because the base suite was not run. The tip suite's single failure is named in Ground and was not counted as a feature finding.

### Conditions

No conditions substitute for the two blocking repairs. The master must keep their class at BLOCKING unless it records a justified upward reclassification at the gate.

### Unrun

Base-commit whole-module tests were not run: this review worktree is confined to the reviewed tip and the order forbids an external checkout. Race detector, coverage and a separate linter were not run: the feature does not add concurrent state, no bar is declared, and no linter command is configured. A live endpoint review and provider receipt-signature check were not run: the boundary permits local tests only, and the pack does not carry signed receipt bodies.

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Two blocking defects: a verified pack may omit required payload files, and a short configured secret may enter the signed pack. Focused tests reproduce both. The whole-module run has one unrelated red; base comparison was unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T13:22:28Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| private credential or variable name | 1 |
