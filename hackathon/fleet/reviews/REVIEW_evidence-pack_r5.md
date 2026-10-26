# REVIEW — evidence-pack · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T02:30:50Z

## Summary

Linux pack review found no reproduced defect; base and tip tests/vet pass. Windows ACLs were not run. Scratch cleanup was refused by the command guard; details in review notes.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground

- Reviewed `d9083c0...0b71da0` (the local `main` ref is absent; the recorded base resolves to `d9083c0`). `git rev-parse --short=7 HEAD` returned `0b71da0`. I did not read other wagon reviews.
- Discovery searches: `git diff --name-only d9083c0...HEAD`; `rg --files -g '*evidence*' -g '*pack*' -g '*verify*'`; `rg -n 'func |secret|token|Escape|verify|manifest|archive|signature' internal/evidence/{pack.go,verify.go} internal/cli/pack.go`; `rg -n 'TracePath|trace.jsonl|LogDir' internal/review internal/cli/run.go internal/trace`; `rg -n 'Secret:' internal/{generator,guardrail,query,settingspage}/*go`; `rg -n 'func Test' internal/{evidence,cli,redact}/*test.go`. Read writer, verifier, CLI call sites, redactor, trace writer, key loader, output collision handling and corresponding tests.
- `go test -count=1 ./...`: exit 0 on archived base and tip, respectively 58 and 59 packages enumerated by `go list ./...`. `go vet ./...`: exit 0 on both. `go test -count=1 ./internal/evidence ./internal/cli ./internal/redact`: exit 0 on tip. The master's clean-vet and clean-suite claim was reproduced; its exact 48/49 `ok` counts were not independently captured as counts.
- Dependency manifests and lockfiles have no changes (`git diff --name-only d9083c0...HEAD -- go.mod go.sum package.json package-lock.json` returned empty). `git diff --check d9083c0...HEAD` passed. No deleted test, weakened assertion, new suppression or bar change found. Test discipline ranking: Go vet is a project/tool check; Go tests and the mutation controls below are suite checks, the latter more discriminating; no external audit was claimed.

### Differential

- `internal/evidence/pack.go:125-189` refuses missing signing material and short configured secrets, redacts payload and manifest values, hashes every payload file, signs the manifest, then self-verifies; `internal/cli/pack.go:57-72` writes that result. `internal/evidence/pack.go:270-299` writes deterministic sorted tar entries.
- `internal/evidence/verify.go:135-224` compares the whole archive to its canonical reconstruction, validates the signature before reading manifest claims, checks listed sizes and SHA-256 digests, required and extra files, then receipt consistency. `internal/evidence/verify.go:231-263` rejects non-flat or repeated entries. `internal/cli/pack.go:141-224` passes pack-derived text through a visible-control encoder. `internal/evidence/key.go:55-73` checks Unix private-key mode before loading; `internal/cli/config.go:443-509` includes the pack in output-identity checks.
- No behavior outside the review's pack/verify integration was judged as a pass. No blocking or owed defect was reproduced within the ordered property.

### Unplug

- Independently copied the tip into `███████████████████████████████████████████`, inverted one guard at a time, ran the stated targeted tests and restored the scratch source before the next inversion. None of these mutants touched the reviewed checkout.
- Whole-archive byte comparison off (`internal/evidence/verify.go:147`): `TestThePackIsOneByteStringAndAnyOtherIsRefused` and `TestEveryChangedByteOfThePackIsRefused` fail; 2,809 of 12,800 single-bit changes then verify.
- Ed25519 verification off (`internal/evidence/verify.go:166`): `TestAChangedManifestIsRefusedBySignature` fails, accepting a changed manifest. SHA-256 comparison off (`internal/evidence/verify.go:195`): `TestOneChangedByteIsRefusedNamingTheFile` fails on all seven payload entries. Extra-file rejection off (`internal/evidence/verify.go:210`): `TestAMissingAndAnExtraFileAreRefused` fails. Required-file rejection off (`internal/evidence/verify.go:200`): `TestRequiredPackFilesCannotBeOmitted` fails, including missing commits/settings accepted.
- Payload redaction off (`internal/evidence/pack.go:228`): `TestNoConfiguredSecretReachesThePack` and `TestASecretInThePacksOwnJSONSpellingIsRedacted` fail. Private-key mode check off (`internal/evidence/key.go:68`): `TestAPrivateKeyThatGroupOrOthersCanReachIsRefused` and `TestAReadableKeyIsRefusedBeforeTheReview` fail. Bidirectional escape off (`internal/evidence/verify.go:69`): `TestARefusalShowsNoControlOrBidirectionalCharacter` fails. Receipt trace-match rejection off (`internal/evidence/verify.go:317`): `TestAReceiptThePackCannotShowIsRefused` fails. Pack output identity removed (`internal/cli/config.go:446`): `TestEvidencePackCollidesWithEveryOutputFlag` fails for five output spellings.
- False-world answer: a missing mechanism produced a failing targeted test, not the same answer as the intact implementation. The restored tip passes the full suite; no mutant was treated as a product defect.

### Closure

- FILE·LINE·CONTROL: `internal/evidence/verify.go:143-212` / canonical archive, signature, hash and exact membership: tests and unplug checks above turn red. `internal/evidence/pack.go:146-189` / redaction and signing: existing payload and spelling tests and redaction mutant turn red. `internal/evidence/key.go:55-73` / readable-key refusal: mode and CLI tests turn red. `internal/evidence/verify.go:60-86` with `internal/cli/pack.go:171-224` / terminal escaping: refusal and report tests cover pack-derived text and the bidi mutant turns red. `internal/cli/config.go:443-509` / pack/report aliasing: collision test turns red.
- No blocking finding requires a new failing Go test. A separate scratch test, `TestVerifyEscapesUnreadablePublicKeyPath`, passed: a missing public-key path containing an escape byte prints it quoted, not as a terminal control. This test was not added to the repository.

### Remainder

- Boundaries: pack input and filenames can be tampered with before verification; the verify CLI accepts only canonical archives signed by the supplied public key with matching payload digests and membership. Signer-controlled pack content is untrusted at the terminal and rendered visibly. Signing-key paths and permissions are local trust inputs. Pack output paths can alias other report outputs and are checked both before writing and against files already written.
- The pack cannot establish that a finding is true or authenticate the provider's receipt signature; `internal/evidence/verify.go:107-113` explicitly says so. Storage/retention access is the operator's decision. These limits do not negate the ordered integrity property. Declined to judge the correctness of underlying reviews, provider receipt provenance, Windows ACLs and filesystem races across actors; no broader guarantee is implied.
- No instruction-like comment or bidirectional/invisible character was found in the changed executable source beyond deliberate test fixtures for escaped text; no new dependency or destructive derived path was introduced. No quality-only restructuring requested. Ruling: preserve the documented Windows exception rather than silently treating a Unix mode check as cross-platform; the cost of being wrong is a false cross-platform security claim.

### Profile

- Security S1: spoofed archive structure, tampered payload/manifest and signer-supplied visible text have the corresponding checks above; disclosure is controlled by redaction of configured values but a pack still intentionally contains review evidence; privilege boundary is local key-file permissions. S2 closure uses attacker-written pack entries in existing tests, not trust in the author's pack. S3 mode and collision checks were removed once in scratch and their tests failed. S4 dependency delta empty, named secrets traced from configuration catalog through redactor, no changed recursive deletion or path-derived destructive operation.

### Focus

- The ordered integrity, configured-secret, Unix key-mode and pack-derived terminal-text properties hold in the measured Linux checkout. `master-was-wrong: none` for the base/tip vet and test pass claims; the claimed `ok` counts are not claimed as independently matched.

### Conditions

- Acceptance does not extend the private-key permission guarantee to Windows: `internal/evidence/key.go:55-73` explicitly excludes Windows, as the pack documentation says. A cross-platform claim requires a Windows-specific permission instrument and negative control. No other condition.

### Unrun

- Windows private-key ACL behavior: this run is on Linux and there is no Windows runner in this checkout. The documented Windows exemption is not treated as a successful key-permission check.
- Scratch cleanup: the scratch archive and mutants remain under `█████████████████████████████████████`. The attempted `rm -rf` was refused by the command guard as an `rm -f`-style operation. I did not retry the denied act in another spelling; `git status --porcelain` still reports `████████` as untracked. Remove it only through an authorized cleanup procedure.

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Linux pack review found no reproduced defect; base and tip tests/vet pass. Windows ACLs were not run. Scratch cleanup was refused by the command guard; details in review notes.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T02:26:05Z

<!-- ENGINE-EMITTED REVIEW SKELETON, rendered from ██████████████████████████████.
     Everything below this line contains WORKED FORMS to copy as applicable.
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy the applicable form OUT.

     Write-legal `- state:` tokens: spawned · recommendation-ready.
     Last act before done: set `- state:` above to recommendation-ready and add every
     return-time key shown below. Their names, value classes, and vocabularies
     are projected from the review family's rows, never copied here.

     `verdict`: exactly one write-legal verdict token, alone on the line — mandatory for `recommendation-ready`; no parenthetical and no conditions after it; the conditions belong in the review body the `- artifacts:` manifest names
     - verdict: <accept | accept-with-conditions | reject>

     `ref`: the review's returned artifact — mandatory for `recommendation-ready`; one of this document's own `- artifacts:` entries, VERBATIM (a basename that is not a declared entry is refused at delivery); never a git ref
     - ref: _REVIEW_NOTES.md

     `artifacts`: NESTED DASH LIST of paths the review returns; mandatory for `recommendation-ready`, including an explicit empty list when nothing is returned
     - artifacts:
       - _REVIEW_NOTES.md
       - _REVIEW_STATUS.md

     An entry names a FILE, or a DIRECTORY: a directory is delivered as ONE
     deterministic `.tar` archive (sorted entries, fixed mtimes — the same
     bytes, and so the same digest, every time) recorded under the entry's
     own name plus that suffix, which is the name the projected manifest and
     the teardown gate then read. A path that EXISTS and is neither is
     refused by what it IS, never as `missing`.

     Explicit empty manifest form when the review returns no files:
     - artifacts:

     🔴 THE ANALYSIS DOES NOT GO IN THIS FILE. It goes in _REVIEW_NOTES.md, whose
     mandatory members are H2 HEADINGS — never top-level dash bullets in
     _REVIEW_STATUS.md. A section heading written as `- ground:` here is an unknown
     key: it is now reported as a WARNING rather than refused, so your
     verdict still lands — but the analysis lands nowhere at all.

     The complete _REVIEW_NOTES.md form, projected from the same one section source
     this engine renders it from (██████████████████████████████):

     # REVIEW NOTES — evidence-pack

     ## Ground
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Differential
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unplug
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Closure
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Remainder
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Profile
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Focus
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Conditions
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unrun
     <replace with measured review content; write `none` only when none is the measured answer>
-->

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| file and line in Saphan Studio's private code | 2 |
