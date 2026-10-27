# REVIEW — evidence-pack · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T08:19:46Z

## Summary

24cfc28: signature, per-file hash and file-set checks hold and are pinned by mutants; key comes from the operator. Probes found four gaps against the order:

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground
- Reviewer: claude-code (opus), role REVIEW, run evidence-pack-review-r2, 2026-10-24.
- Role-lock: `git rev-parse --short=7 HEAD` printed `24cfc28`. Scope: `git diff fcbcb79...24cfc28`, 3 commits, 13 files, +1556/−4. The author's self-assessment (the master's "vet clean, 41/41") was set aside per R1 and re-measured below.
- Toolchain: `go version go1.26.10 darwin/arm64`.
- `go vet ./...`: clean (no output).
- `go test -count=1 ./...` at 24cfc28: every package touched by the diff passes: `internal/evidence`, `internal/cli`, `internal/change`, `internal/settingspage`, plus `cmd/proven`. Six packages fail: `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`. Each has exactly one failing test, and all six failures are `httptest: failed to listen on a port ... bind: operation not permitted`. Counted with `grep -c` over those packages: 6 of 6. This is the leg's sandbox refusing a listener, not the diff. None of the six packages is in the write-set.
- Probes: `██████████████████████████████████████████████████████`, run as `go test -count=1 -v ██████████████████████████████████████████████████████`. They use only the exported API and write nothing outside `t.TempDir()`.
- Order divergence, recorded under law #44: the order says the H1 of `_REVIEW_STATUS.md` is `# REVIEW — evidence-pack-review-claude`. The engine's channel grammar says `# REVIEW — evidence-pack` and explains that the verdict row binds to the reviewed slug. I followed the engine.

### Differential
Each property from the order, with its location, the control that holds it, and the measured result.

1. **P1: a pack verifies only if the signature, every hash and the file set all hold.** HOLDS for content, FAILS as worded.
   - Content integrity holds, and the controls are pinned. The signature check is `internal/evidence/verify.go:109`; its mutant fails `TestAChangedManifestIsRefusedBySignature`. The per-file hash is `verify.go:135`; its mutant fails `TestOneChangedByteIsRefusedNamingTheFile`. The missing check is `verify.go:131`. The extra-file check is `verify.go:145`; its mutant fails `TestAMissingAndAnExtraFileAreRefused`. The exit code is 2 (`internal/cli/pack.go:181`).
   - "Any changed byte is refused" is false. The probe `TestEveryByteFlip` flipped one bit at each of the 11776 positions of a pack, and **3377 single-bit flips (28.7%) still verify**. Those bytes are tar header metadata (mode, mtime, uname and similar) and block padding. **Bytes appended after the end-of-archive verify** (`err=<nil>`). **Whitespace added around `MANIFEST.sig`, together with a rewritten mode/uname/mtime on every header, verifies** (`TestOutsideSignedBytes`, `err=<nil>`). Causes: `readArchive` (`verify.go:166-199`) never compares the container to the canonical form, and `verify.go:105` applies `TrimSpace` to the signature file, which no hash covers.
   - None of these bytes changes what `verify` accepts or reports. The defect is that the pack file is not tamper-evident as a byte string. Two consequences follow. A stored digest of the pack file can differ from the original while `verify` still says "verified". The test comment at `internal/cli/pack_test.go:86`, "One byte changed anywhere in the pack is refused", overclaims: it passes only because `len/2` lands in content. I ran it 20 times (`-count=20`) and it was stable, so it is not flaky.
2. **P2: the signature covers what the manifest claims, and the key comes from the operator.** HOLDS.
   - The signature is over the exact manifest bytes (`pack.go:150`, `verify.go:109`). The manifest is decoded only after the signature passes. Every payload file is bound by its hash.
   - The verifying key comes from `--public-key`, else `██████████████████████` (`internal/cli/pack.go:167-172`). Nothing in the pack selects the key: `m.KeyID` is only compared against the operator's key (`verify.go:121`), never used to look one up.
   - Fields that are not signed (tar metadata, signature whitespace) change no verdict, as measured in P1.
3. **P3: no configured secret in any file of the pack, and the private key never enters the pack.** FAILS on one path.
   - `Build` writes `Models: models(in.Result)` into MANIFEST.json **without the redactor** (`internal/evidence/pack.go:134`). The same list goes into `receipts.json` through the redactor (`pack.go:204`).
   - Probe `TestSecretInModelUse` set the route to `https://x/<secret>`. Result: "pack contains secret: true", "secret in MANIFEST.json".
   - The same pack then **fails its own verification**: `receipt check failed on receipts.json: receipt 1 differs from the manifest's model use`. `Build` does not verify what it wrote, so the scan exits 0 with a pack no auditor can verify.
   - Every other field is redacted, as far as I could find: the payload files (`pack.go:192`, `pack.go:217`), the manifest commit names (`pack.go:128-131`), and the settings (`internal/cli/pack.go:124-126`, with secrets as `set`/`unset`).
   - The private key: `TestThePrivateKeyNeverEntersThePack` checks the seed and the full key bytes. `KeyError` (`key.go:21-32`) carries the setting, the path and a reason, never content. `readPEM` drops the OS error text (`key.go:78`). No log line in the diff prints the key.
4. **P4: the key file's owner-only permissions are checked or refused, and a missing or malformed key is refused at load with exit 7, before anything is written.** FAILS on permissions, HOLDS on the rest.
   - `readPEM` (`internal/evidence/key.go:71-85`) calls `os.ReadFile` with no `Stat` and no mode check. Probe `TestWorldReadableKey`: a `0666` private key loads with `err=<nil>`.
   - The docs (`docs/EVIDENCE_PACK.md`, "Keep ████████ where only ...") leave this to the operator. The diff contains no check and no refusal.
   - Missing, unreadable or malformed keys are refused through `loadPackKey` → `cliError{exitConfig}` (`internal/cli/pack.go:26-31`). This runs at `run.go:64-71`, before `openCall` and the review, and nothing has been written at that point (`writable` at `run.go:158` is a pure check). This is pinned by `TestAMissingKeyIsNamedAndNotPrinted`, which asserts exit 7, no pack and empty stdout. The CLI test covers unset and unreadable keys; the malformed case is covered only at unit level (`TestKeyRefusalsNameTheSettingAndNeverTheKey`).
5. **P5: shell rules, and pack bytes shown to a terminal pass through the existing `visible` encoder.** FAILS, and the order's premise is wrong.
   - **There is no `visible` encoder in this repository at 24cfc28.** `grep -rn -i -w visible --include='*.go' .` finds only five comments, and no `IsControl`/`IsPrint`/escape helper exists under `internal/` or `cmd/`. The order's premise is refuted. Measured 2026-10-24, on the tip only, because the leg's permission class denied `git grep` on other refs.
   - Bytes from the pack reach the terminal raw. The more serious path comes **before the signature check**. `readArchive` refusals put the tar entry name `hdr.Name` into `Refusal.File` (`verify.go:175-187`). `Refusal.Error` concatenates it (`verify.go:41-47`), and `runVerify` prints it to stderr (`internal/cli/pack.go:179`).
   - Probe `TestControlBytesInRefusal`: an entry named `a\x1b]0;owned\x07\x1b[2J` produced a refusal containing the raw ESC/BEL bytes. Anyone can craft that input; it needs no key.
   - After the signature check, `renderReport` (`internal/cli/pack.go:186-214`) prints the signer-controlled strings `Created`, the commit names, the agent/model/route/receipt fields, the proof-level keys and the failed-service names, with no encoding.
   - The other shell rules hold. A flag beats the variable (`pack.go:167-171`). Refusals name what is allowed: `known flags: --public-key` (`pack.go:160`), and "verify takes exactly one pack file". `--evidence-pack` is a value flag (`config.go` isValueFlag).
6. **P6: every claim has a test whose false-world answer differs.** PARTIAL. See Unplug.

### Unplug
Each guard was mutated in the working tree and restored with `git checkout -- internal/evidence/verify.go`. `git diff --stat` was empty after each restore.
- Pinned guards (the mutant makes a named test fail):
  - The signature (`verify.go:109`) is pinned by `TestAChangedManifestIsRefusedBySignature`. With the signature mutant, `TestAWrongKeyIsRefusedBySignature` still passed, because the KeyID guard at `:121` refuses under the same check name.
  - The hash (`verify.go:135`) is pinned by `TestOneChangedByteIsRefusedNamingTheFile`.
  - The extra-file check (`verify.go:145`) is pinned by `TestAMissingAndAnExtraFileAreRefused`.
- **Six surviving mutants**, applied together. `go test -count=1 ./internal/evidence ./internal/cli` was green with all six in place, so each one alone also survives. These are guards with no test:
  - `verify.go:118` (version check);
  - `verify.go:121` (manifest key id equals the operator's key);
  - `verify.go:126` (a file listed twice);
  - `verify.go:114` (`DisallowUnknownFields`);
  - `verify.go:186` (the `maxEntry` bound);
  - `verify.go:214` (receipts.json equals the manifest's models).
- Known by reading, not run: swapping the `--public-key` and `██████████████████████` precedence (`internal/cli/pack.go:167-171`) survives. No test sets both, so the flag-wins rule has no pin.
- What should fire if each failed property were unplugged, and does not today:
  - P1 container malleability: nothing fires.
  - P3 manifest-model leak: nothing fires at write time. The verify failure appears only later, at the auditor.
  - P4 permissions: nothing fires.
  - P5 encoding: nothing fires.

### Closure
- Callers read: `Run` → `loadPackKey` → `writePack` (`run.go:63-126`). `runVerify` is dispatched before `resolve` (`run.go:35-37`). `change.ResolveCommit` passes `--end-of-options` before the revision, so an option-shaped name is not read as a git flag. The trace path is derived in two places: `run.go:88` (cwd/logs/trace.jsonl) and `run.go:199` (`LogDir`). `trace.Open` is called from `pipeline.go:152`. Today the two derivations agree.
- Paired paths (R13): `Build` and `Verify` share no test-time round trip of the manifest models through redaction, which is why the P3 leak and its self-verify failure go unseen. The CLI scan→verify round trip (`TestSampleReviewWritesAPackThatVerifies`) runs in sample mode only, where no secret can reach a model use.
- Trust boundary (S1). The pack file is untrusted until its signature passes. Before that point the code reads: tar parsing, with each entry bounded at 512 MiB but **no bound on entry count or total size**, so a pack of many small entries is held entirely in memory; and refusal text (control bytes, P5). After the signature passes, the content is the signer's, and only terminal rendering remains a concern.

### Remainder
Owed findings. They do not block (R6).
- R8: `Build` does not re-verify its own output, so a pack that cannot verify is written with exit 0 (`internal/cli/pack.go:68-72`). Remedy: call `Verify(pack, key.Public())` before `WriteFile` and refuse on failure.
- R8: the trace offset is derived from `os.Getwd()` (`run.go:86-90`) and the body is read from `result.TracePath` (`internal/cli/pack.go:61`). These are two derivations of one path, and concurrent runs in the same directory interleave their lines into one pack. Remedy: take both from one source. The interleaving belongs in the docs.
- The pack destination is not checked as writable before the review (`run.go:59` checks only formats). A missing directory costs a whole review and then exits 2. Remedy: check the destination's directory in `writable`.
- `verify --help` is refused as an unknown flag (exit 2) instead of printing usage.
- `FileRef.Size` is signed and never compared to the data length. It is harmless, because the hash binds the data, but it is a field a reader may wrongly trust.
- No bound on entry count or total size (see Closure).

### Profile
security, chosen because the diff touches signatures, key handling and secret redaction (R9): S1 trust-boundary map (Closure); S2 every blocking finding reproduced by a probe (Differential 1, 3, 4, 5); S3 guard removal (Unplug); S4 secrets searched (P3), no dependency delta (go.mod/go.sum untouched by the diff), and no invisible or bidirectional characters introduced by the diff (pack contents are judged in P5). testing-bar: the project declares no metric, command or baseline, so I set no number. Bar ranking: every check here is a SUITE check, which is a NOTE, and the mutation results in Unplug are the strongest evidence I have.

### Focus
Properties 1–6 of the order only, over `fcbcb79...24cfc28`. Licences are out of scope.

Declined to judge (R11): live-mode receipt content, which needs network; ADR-0035 text not shipped in the diff; and the six non-diff packages that are red because of the sandbox.

### Conditions
Each condition is to be addressed before merge. The class is owed-blocking against the order's stated properties.

1. **C1 (P3)**: redact the manifest's model list at `internal/evidence/pack.go:134`, for example by building `Models` from the same redacted bytes as `receipts.json`. Add a test that puts the secret in `Route` or `Model`, then asserts the secret is absent from every file and that the pack verifies.
2. **C2 (P4)**: in `LoadPrivateKey`, refuse a private key file with any group or other permission bit, by name and with exit 7 (`internal/evidence/key.go:71-85`). Add a test with a `0644` key that fails before the fix. Docs (law #39): the refusal and its remedy, `chmod 600`, go into `docs/EVIDENCE_PACK.md`.
3. **C3 (P5)**: escape control and bidirectional characters in every pack-derived string that reaches stdout or stderr: `Refusal.Error` / `verify.go:41-47`, and `renderReport` at `internal/cli/pack.go:186-214`. No `visible` encoder exists, so one must be added or the order's premise corrected by the master. Add a test with an ESC in a tar entry name that fails before the fix.
4. **C4 (P1)**: make the pack byte-canonical or narrow the claim. Preferred: `Verify` rebuilds `archive(files)` and requires it to equal the input bytes exactly. The writer is deterministic, so this refuses metadata, padding, trailing-byte and signature-whitespace changes. Pin it with a test that flips each byte. Otherwise, change the docs and the comment at `internal/cli/pack_test.go:86` to say the content, not the file, is tamper-evident.
5. **C5 (P6)**: add tests that kill the six surviving mutants listed in Unplug, plus one that sets both `--public-key` and `██████████████████████` and asserts the flag wins.

### Unrun
- `go test` at the base `fcbcb79`: not run. A second checkout would place an artifact outside the worktree, which the order forbids, so pre-existing reds at the base are not separated by measurement. They are attributed by cause instead: six bind refusals in packages outside the write-set.
- The six sandbox-red packages (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`): not run green, because the seat refuses `bind` for `httptest`.
- The race detector (`-race`) was not run. The diff adds no goroutines.
- A live-mode (non-sample) pack with real provider receipts was not run, because there is no network.
- Checking `main` or any other ref for a `visible` encoder: `git grep` across refs was denied by permission class `██████████████████████████`, so only the tip's tree was searched.

## The review status as filed

## REVIEW — evidence-pack
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: 24cfc28: signature, per-file hash and file-set checks hold and are pinned by mutants; key comes from the operator. Probes found four gaps against the order:
  model uses enter MANIFEST.json unredacted (secret leaks, pack then fails its own verify); a 0666 private key loads; tar entry names reach stderr raw before the signature check (no `visible` encoder exists at 24cfc28);
  28.7% of single-bit flips verify (tar metadata, padding, trailing bytes, sig whitespace). Six verifier guards have no test. Conditions C1–C5 in _REVIEW_NOTES.md.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
    - probe_test.go
- updated: 2026-10-24T09:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| policy binding name | 1 |
| private credential or variable name | 3 |
| signing key id | 1 |
