# REVIEW — evidence-pack · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T07:56:11Z

## Summary

Reject: four reproduced blocking gaps in secret handling, key permissions, terminal encoding and tar completeness; see review notes.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground
RED HEADER — The signed pack can disclose a configured secret through the unredacted manifest model list. Reject until the signer uses one consistently redacted model list and the full-archive negative control passes.

Review of `fcbcb79...24cfc28`, 13 changed files, against ADR-0035 and the review order. `git rev-parse --short=7 HEAD` printed `24cfc28`. Go version: `go1.26.5 linux/amd64`; module declares Go 1.26. Reviewed both the pack writer and verifier, CLI caller and rendering, key loader, and their tests. No module manifest or lockfile changed. No project coverage or mutation baseline was supplied. The author's 41/41 statement was treated as a claim, not evidence.

### Differential
- Base `fcbcb79`: `go test -count=1 ./...` on a local `git archive` checkout, with umask 022, exited 0; 40 package lines reported `ok`. `go vet ./...` exited 0.
- Tip `24cfc28`: `go test -count=1 ./internal/evidence ./internal/cli` exited 0. `go test -count=1 ./...` with inherited group-writable umask exited 1 at pre-existing `TestMaterializedTreesAreReadableByOthers` (`go.mod` mode 664); with umask 022 the full suite exited 0, with 41 package lines reporting `ok`. `go vet ./...` and `git diff --check fcbcb79...24cfc28` exited 0. A package count is not a test count; tests were not counted individually.
- Test-bar weakening: no removed tests, weakened assertions, skips, suppressions, thresholds, or changed dependency manifests found in the three-dot diff. Check ranking: Go compiler/vet are toolchain checks; project Go configuration and diff check are project checks; `go test` is an author-controlled suite check. No recorded coverage or mutation threshold to compare. No exceptions seen.

### Unplug
- Signed manifest, changed listed file, missing/extra named file: existing `TestAChangedManifestIsRefusedBySignature`, `TestOneChangedByteIsRefusedNamingTheFile`, `TestAMissingAndAnExtraFileAreRefused` mutate their named fields; those tests passed. `TestAWrongKeyIsRefusedBySignature` demonstrates externally supplied public-key refusal. The writer's `TestNoConfiguredSecretReachesThePack` mutates finding, evidence, trace, settings and baseline but not `Result.Models`.
- Locally run isolated reproduction (`████████████████████████████████████████████████`) used generated Ed25519 keys, no network, and only locally created data. With a secret in `Result.Models[0].Model`, the signed tar contained the configured secret. With that secret removed from the model, the baseline verified; appending one byte to the otherwise valid tar still verified; a mode-0644 private key loaded without error; an ESC sequence in a model came out of `proven verify` raw with exit 0. These are false-world answers different from the intended refusal or encoding. The initial secret-bearing pack itself failed receipt comparison because `receipts.json` was redacted while manifest models were not.
- Missing negative controls: a key-mode test expecting exit 7, a model-field secret test checking every archive member and verification, a terminal-control-byte test, and a trailing-data rejection test. Merely unplugging the current tests does not expose these failures.

### Closure
**REJECT.** Four blocking findings within this diff:
1. **Secret disclosure and inconsistent signed model data** — `internal/evidence/pack.go:134` writes `models(in.Result)` into the signed manifest without the redactor; `internal/evidence/pack.go:192` redacts the payload copy, so `internal/evidence/verify.go:214` rejects a secret-bearing model on receipt mismatch only *after* the private value has already entered the pack. Trigger: a configured token echoed in a generator model field. Local reproduction: `configured secret in signed pack=true`; the initial verify failed on `receipts.json`. Remedy: redact the model list before signing and use the same sanitized list in the receipts payload; test a secret in every manifest model field, including escaped forms, and assert the entire archive contains no secret while verification succeeds.
2. **Private key mode is not checked** — `internal/evidence/key.go:71-80` reads the configured PEM file directly; `internal/cli/run.go:63-71` calls this loader before review and maps only returned errors to exit 7. Trigger: a valid mode-0644 private key. Local reproduction: `group-readable private key refused=false error=<nil>`. Remedy: stat the opened key (and verify owner-only permissions on the file actually read) before parsing; refuse insecure mode at load with exit 7; test the mode-only mutation and no output written.
3. **Untrusted pack text reaches the terminal without encoding** — `internal/cli/pack.go:187-210` prints model, route, receipt, commit fields and pack path using `%s`, and `internal/cli/pack.go:177-182` sends the rendered report directly to stdout. Trigger: a signed pack whose model contains an ESC control sequence. Local reproduction: `terminal raw control emitted=true exit=0`. Remedy: pass each pack-derived terminal field, including refusal file names and the supplied pack path, through the shell's visible encoder at the output boundary; test control and bidirectional bytes with an otherwise valid signed pack.
4. **Additional archive bytes are accepted** — `internal/evidence/verify.go:164-192` ends archive parsing on the first tar EOF, without ensuring the input byte stream ends there. Trigger: append one byte to a fully valid signed tar. Local reproduction: baseline verifies and `appended byte refused=false error=<nil>`. Remedy: define and enforce a canonical complete tar envelope or explicitly reject trailing bytes; test appended bytes as well as a changed listed-file byte. The existing midpoint-byte test does not distinguish this case.

### Remainder
The public key is loaded from `--public-key` or the configured environment, not from the tar (`internal/cli/pack.go:139-168`); signature and listed payload hashes are checked. Missing/extra named regular entries and duplicate tar entry names are refused. No claim of provider receipt signature verification: `notChecked` explicitly says it is unavailable offline. No product files changed. The pre-existing umask-sensitive test is not a finding against this diff; the first inherited-umask suite failure is disclosed above. Public copy and bridge registration are outside this confined review leg; the master handles those records. Ruling: interpret “one byte changed” to include a byte appended to the archive; if policy intentionally permits trailing bytes, the owner must resolve the ADR-0035 promise explicitly.

### Profile
Security: signer reads configured local file; pack carries signed manifest and untrusted review/model text; auditor supplies verifying public key; terminal reads rendered pack fields. Confidentiality fails at the manifest, integrity fails for trailing data, signer confidentiality boundary is too permissive, terminal trust boundary fails. Backend: readArchive holds the archive bytes and individual entries; large-file and resource-exhaustion analysis beyond these acceptance properties was declined. UI: tested the terminal escape condition. No external dependency delta, network activity, or installation.

### Focus
File, line, trigger, control and reproduction are in Closure. Scope is only the reviewed three-dot diff. Fixes belong to the executor; this reviewer did not change product code. The master must re-run the named negative controls on a fix and obtain another independent review as required by the track.

### Conditions
Reject; there are no conditions that make this version acceptable without fixes. Each of the four blocking findings needs a failing-before/passing-after test using the mutated field named in Closure. No class is downgraded by the passing suite.

### Unrun
No full guard-by-guard mutation campaign, no live model/provider run, no receipt signature verification (pack lacks the signed receipt body), no coverage or mutation metric (none recorded), and no public bridge document registration from this confined review leg. These do not soften the reject verdict.

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject: four reproduced blocking gaps in secret handling, key permissions, terminal encoding and tar completeness; see review notes.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T07:52:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
