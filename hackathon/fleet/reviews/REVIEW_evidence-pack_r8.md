# REVIEW — evidence-pack · round 8

- verdict: not recorded
- reviewer: not recorded, not recorded
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T10:27:28Z

## Summary

See the notes as filed below.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-pack

### Ground
- Reviewed: `main...3b6bac5` (merge-base `c1e4ba8`), branch `fix/evidence-pack`, HEAD of this worktree `3b6bac5` (measured `git rev-parse --short=7 HEAD`). Profile: security. 16 files, +2117/−4: `internal/evidence/{pack,verify,key,settings}.go` + tests, `internal/cli/{pack,run,config,format}.go` + tests, `internal/change/between.go` (`ResolveCommit`), `internal/settingspage/catalog.go`, `docs/EVIDENCE_PACK.md`, `docs/CONFIGURATION.md`.
- Read in full: `internal/evidence/pack.go`, `verify.go`, `key.go`, `pack_test.go`, `hardening_test.go`; `internal/cli/pack.go`, `pack_test.go`, `pack_hardening_test.go`; `internal/redact/redact.go` (the redactor the pack depends on, unchanged by this diff); `docs/EVIDENCE_PACK.md`; the diff hunks of `run.go`, `config.go`, `format.go`, `between.go`, `catalog.go`, `settings.go`, `CONFIGURATION.md`.
- Searches made to find every path the diff touches (each run with `grep -rn` in the worktree):
  - `TracePath` → the trace the pack reads is `review.Result.TracePath`, set from `LogDir = <wd>/logs` (`internal/cli/run.go:212`), the same file `traceStart` measures (`run.go` diff hunk) — consistent.
  - `LogDir` → one setter, `internal/cli/run.go:212`.
  - `Secret: *true` → five secret rows (`generator/second.go:31`, `generator/settings.go:74,80`, `guardrail/settings.go:70`, `query/settings.go:48`); all five are in `redact.Names` (`redact.go:28`), which also holds `ZETARIS_ORG_ID`.
  - `ResolveCommit|resolveCommits` → one production caller (`internal/cli/pack.go:82,84`), zero test callers.
  - `exitFatal|exitConfig` → 2 and 7 (`internal/cli/config.go:26-27`), matching the docs.
  - `MinLength` in `internal/evidence`, `internal/cli` → no use by the diff (main has since removed `redact.MinLength`; see Focus).
- Tests run at `3b6bac5` on this seat (go1.26.10 darwin/arm64):
  - `go test ./internal/evidence/ ./internal/cli/ ./internal/redact/ ./internal/change/ ./internal/settingspage/` → all `ok`.
  - `go test -count=1 -race ./internal/evidence/` → `ok`.
  - `go vet ./internal/evidence/ ./internal/cli/` → clean.
  - `go test ./...` → red packages: `cmd/livecheck`, `internal/cli/sarifschema [setup failed]`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`; six red tests (`TestAllServicesReady`, `TestLiveReadsTheProvisionedStream`, `TestLiveReviewCarriesTheRoutedModel`, `TestLiveSelectReadsRowsFromTheSQLEditor`, `TestServerPostsTheMeasuredRequest`, `TestUnsafePromptStopsTheStepAndNamesCategories`). Every cause read is the seat: `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`, and for sarifschema a module download refused (`tls: failed to verify certificate`).
  - Same seven packages at `main` (`5faabc9`): the same seven packages and the same six tests red, same causes. None of the red packages is touched by the diff.
- S4: a Go scan of all 16 changed files for C0/C1 controls, DEL, U+00AD, bidi/invisible marks (`evidence.IsBidi` set) and invalid UTF-8 found none. Non-ASCII present: `§`, `…`, `—`, and in `hardening_test.go:166` a literal U+FFFD used deliberately as the assertion's needle. No instruction-like comments seen in the diff.
- S4 dependency delta: `go.mod`/`go.sum` not in the diff; new imports are standard library only (`archive/tar`, `crypto/ed25519`, `crypto/x509`, `encoding/pem`, …). No destructive operation on derived paths: the only file write is `WriteFile` (temp file in the target's directory, chmod 0600, rename).
- R1: I read no executor status or self-assessment; the order's brief history was set aside as instructed.

### Differential
The property, clause by clause, at `3b6bac5`:

| clause | holds? | where it is enforced | evidence |
|---|---|---|---|
| Ed25519 signature over the manifest's exact bytes | yes | `verify.go:150-156`, key id `verify.go:166` | `TestAChangedManifestIsRefusedBySignature`, `TestAWrongKeyIsRefusedBySignature`, C5 `key id` |
| every file hash | yes | `verify.go:179-182` | `TestOneChangedByteIsRefusedNamingTheFile` (7 files) |
| the file set (none missing, none extra, none listed twice) | yes | `verify.go:170-193` | `TestAMissingAndAnExtraFileAreRefused`, C5 `file listed twice` |
| the exact archive bytes | yes | canonical rebuild `verify.go:131-137`, plus the signature file byte-for-byte `verify.go:151` | `TestThePackIsOneByteString…`, `TestEveryChangedByteOfThePackIsRefused`; the `:151` clause has no test (Unplug U2) |
| any change refused by name | yes | `Refusal{Check, File}` | `wantRefusal` asserts check and file |
| no configured secret in the pack | **no, for a narrow class of values** | `collect` redacts the JSON text (`pack.go:213,238`); the manifest's commits field by field (`pack.go:149-152`); models read back from redacted `receipts.json` (`pack.go:141`) | failing test F1 below |
| a readable private key refused | yes | `ownerOnly` `key.go:60-72`, before the review `run.go` hunk | `TestAPrivateKeyThatGroupOrOthersCanReachIsRefused`, `TestAReadableKeyIsRefusedBeforeTheReview` |
| pack text reaches the terminal escaped | yes | `Refusal.Error` → `Visible` (`verify.go:44-73`); `renderReport` → `visible` on every pack field (`cli/pack.go:189-230`); `escapeControls` gains the bidi case (`format.go`) | `TestARefusalShowsNoControlOrBidirectionalCharacter`, `TestVerifyShowsNoControlByteFromThePack`; six of the render sites have no test (Unplug U6) |

**F1 — a configured secret can reach `findings.json` (and any JSON file of the pack) unredacted.** `pack.go:244-253` writes JSON with `SetEscapeHTML(false)`. The redactor's needles (`redact.go:83-99`) are the raw value, `strconv.Quote`, `strconv.QuoteToASCII` and `json.Marshal` (HTML escaping ON). A value that holds (a) a `"` or `\` (so the raw needle misses), (b) one of `< > &` (so the `json.Marshal` needle, which writes `<`, misses) and (c) a character `%q` escapes and JSON does not — DEL, or a non-printable rune such as U+00AD — (so the `%q` needles miss) is written by the pack in a spelling that is none of the four. Trigger: `███████████████████████='tok"en<12­34'` and a finding message echoing it. Measured: the test below fails at `3b6bac5` with `findings.json holds the configured secret as "tok\\\"en<12­34"`. The pack verifies, so nothing downstream notices. Class: real but narrow — a service-issued token with this character mix is unlikely; the property as stated ("no configured secret") does not hold for it. The callers were read: every pack file passes through `put` → `marshal` → `r.Text` (`pack.go:208-214`); the trace is the run's JSONL, re-redacted with the same needles (`pack.go:238`).

Failing test (full text; place in `internal/evidence`, it uses `sampleInput`, `readArchive`, `marshal` from the package's tests and code; write the secret's U+00AD as the escape `­`):

```go
// A configured secret must reach no file of the pack whatever characters it
// holds. The pack writes JSON with HTML escaping off; the redactor's needles are
// the raw value, %q, %+q and json.Marshal (HTML escaping on). A value holding a
// quote, an HTML character and a character %q escapes but JSON does not
// (U+00AD here) is written by the pack in a spelling that is none of them.
// A fix must make every file of the pack free of the value in every spelling
// the pack's own encoder gives it.
func TestReviewProbeASecretInThePacksOwnJSONSpellingIsRedacted(t *testing.T) {
	secret := "tok\"en<12­34"
	in, pub := sampleInput(t)
	in.Redactor = redact.New(map[string]string{"███████████████████████": secret})
	in.Result.Findings[0].Message = "leaked " + secret
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := marshal(secret)
	if err != nil {
		t.Fatal(err)
	}
	spelled := strings.Trim(strings.TrimSpace(string(enc)), `"`)
	for name, body := range files {
		if bytes.Contains(body, []byte(spelled)) || bytes.Contains(body, []byte(secret)) {
			t.Errorf("%s holds the configured secret as %q", name, spelled)
		}
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatal(err)
	}
}
```

What a fix must make it do: every file of the pack is free of the value in the spelling the pack's own encoder (`SetEscapeHTML(false)`) gives it — e.g. add that spelling to `redact.spellings`, or redact string values before they are encoded. Run the test against the REBASED base: `internal/redact/redact.go` changed on main after `c1e4ba8` (see Focus).

### Unplug
Each mechanism the diff adds, unplugged once at `3b6bac5` (source edited, the named package's tests run with `-count=1`, the edit restored; `git diff --stat` empty after each):

| # | mechanism unplugged | site | test that turns red |
|---|---|---|---|
| U1 | canonical-archive byte equality | `verify.go:135` | `TestThePackIsOneByteStringAndAnyOtherIsRefused`, `TestEveryChangedByteOfThePackIsRefused` |
| **U2** | signature file byte-for-byte (`string(sigBytes) != Encode(sig)+"\n"`) | `verify.go:151` | **none** (run `-count=3`, all green). Go's base64 decoder skips `\r`/`\n` and accepts non-zero padding bits, so a pack whose `MANIFEST.sig` is spelled `…==\r\n`, holds an inner newline, or changes the padding bits of the last data character verifies without this clause; the canonical-archive check cannot see it because the archive is rebuilt from the changed file. Every single-bit flip of the last data character lands on a value that changes real bits, which is why `TestEveryChangedByte…` cannot kill it. **BLOCKING (order item 2).** |
| U3 | `ed25519.Verify` | `verify.go:154` | `TestAChangedManifestIsRefusedBySignature`, `TestEveryChangedByte…` |
| U4 | manifest version | `verify.go:163` | C5 `version` |
| U5 | manifest key id | `verify.go:166` | C5 `key id` |
| U5b | listed twice | `verify.go:171` | C5 `file listed twice` |
| U5c | `DisallowUnknownFields` | `verify.go:159` | C5 `unknown field` |
| U7 | missing file | `verify.go:176` | `TestAMissingAndAnExtraFileAreRefused` (refused as `hash` instead of `missing`) |
| U8 | file hash | `verify.go:180` | `TestOneChangedByteIsRefusedNamingTheFile` (7 subtests), `TestEveryChangedByte…` |
| U9 | extra file | `verify.go:190` | `TestAMissingAndAnExtraFileAreRefused` |
| U10 | non-regular entry | `verify.go:222` | `TestARefusalShowsNoControlOrBidirectionalCharacter` (4), `TestVerifyShowsNoControlByteFromThePack` |
| U11 | plain file name | `verify.go:225` | `TestAnArchiveThatIsNotAFlatFileListIsRefused/path` |
| U12 | duplicate archive name | `verify.go:228` | none — **property still holds**: the canonical rebuild keeps one entry, so the bytes differ and U1 refuses it as `format`. The `duplicate` subtest asserts only `CheckFormat`, so it cannot tell the two guards apart. Owed, not blocking. |
| U13 | entry size bound | `verify.go:231` | `TestAnEntryOverTheBoundIsRefusedBeforeItIsRead` |
| **U14** | receipt count equals model count | `verify.go:255` | **none.** Unplugged, one model use more in the (re-signed) manifest verifies, and one fewer **panics** (`index out of range [0] with length 0`). Reachable only by a holder of the signing key, but it is a mechanism the diff adds with no test. **BLOCKING (order item 2).** |
| U15 | receipt equals manifest model | `verify.go:259` | C5 `receipts differ from the manifest` |
| U16 | receipt found in the trace | `verify.go:297` | `TestAReceiptThePackCannotShowIsRefused` |
| U17 | trace redaction | `pack.go:238` | `TestNoConfiguredSecretReachesThePack` |
| **U18** | manifest `Target` and `Head` redaction | `pack.go:149,152` | **none.** Only `Baseline` is pinned (`pack_test.go:226`). Unplugged, a secret in the target path or the head name is written into `MANIFEST.json` in clear. **BLOCKING (order item 2).** |
| U19 | manifest models read back from redacted `receipts.json` | `pack.go:141,155` | `TestASecretInAModelUseReachesNoFileAndThePackVerifies` |
| U20 | private key mode `0077` | `key.go:68` | `TestAPrivateKeyThatGroupOrOthersCanReachIsRefused` (6 modes), `TestAReadableKeyIsRefusedBeforeTheReview` |
| U21 | key loaded before the review | `run.go:73-77` | `TestAReadableKeyIsRefusedBeforeTheReview`, `TestAMissingKeyIsNamedAndNotPrinted` (2) |
| **U22** | secret settings written as `set`/`unset` | `cli/pack.go:113-114` and `:120-122` | **none.** Unplugged together, the whole `internal/cli` suite stays green: the only CLI pack tests run `--sample`, and `packSettings` skips every environment setting in sample mode (`:109`). With both unplugged, a secret shorter than the redactor's minimum (6 on this base) is written into `settings.json` in clear. Each half alone is masked by the other (the `redact.Names` loop overwrites the catalog's five secret rows). **BLOCKING (order item 2).** |
| **U23** | `visible` on six report fields: path `:191`, created `:193`, proof level `:206`, failed services `:209`, head `:223`, target `:229` (`internal/cli/pack.go`) | — | **none.** Unplugged together, the whole `internal/cli` suite stays green; only model, route and baseline are pinned (`pack_hardening_test.go:43-46`). **BLOCKING (order item 2).** |
| U24 | `escapeControls` bidi case | `format.go` (diff hunk) | `TestVerifyShowsNoControlByteFromThePack` (raw U+202E) |
| — | `ResolveCommit` and its `--end-of-options` | `change/between.go:256-265` | not unplugged; measured instead: zero test callers (grep in Ground), so no test can turn red. Owed. |

For each BLOCKING survivor a test that kills it, run green at `3b6bac5` and red with the mechanism unplugged (U2 all 3 subtests red; U14 both subtests red, one by panic; U18 `MANIFEST.json holds the configured secret`; U22 all six secret names red; U23 `the output holds the raw character U+001B`):

```go
// internal/evidence — U2
// The signature file is the base64 and one newline, byte for byte. Each
// spelling below decodes to the same 64 signature bytes, so only the
// byte-for-byte clause at verify.go:151 can refuse it; the archive is rebuilt
// from the changed file, so the canonical-archive check passes.
func TestANonCanonicalSignatureSpellingIsRefused(t *testing.T) {
	pack, _, pub := builtPack(t)
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	body := strings.TrimSuffix(string(files[SignatureName]), "\n")
	last := body[len(body)-3] // the char before "==": its 4 low bits are padding
	alt := strings.IndexByte(alphabet, last) + 1
	for name, s := range map[string]string{
		"carriage return": body + "\r\n",
		"inner newline":   body[:10] + "\n" + body[10:] + "\n",
		"padding bits":    body[:len(body)-3] + string(alphabet[alt]) + "==\n",
	} {
		t.Run(name, func(t *testing.T) {
			m := rewrite(t, pack, func(n string, b []byte) []byte {
				if n == SignatureName {
					return []byte(s)
				}
				return b
			}, nil)
			_, err := Verify(m, pub)
			wantCheck(t, err, CheckSignature, "Ed25519")
		})
	}
}

// internal/evidence — U14
// The mutated field is the manifest's model count, signed again: one model use
// more than receipts.json holds must be refused by the count check at
// verify.go:255, and one fewer must be refused, not panic.
func TestAModelCountThatDiffersFromTheReceiptsIsRefused(t *testing.T) {
	pack, in, pub := builtPack(t)
	for name, edit := range map[string]func(m *Manifest){
		"one more":  func(m *Manifest) { m.Models = append(m.Models, m.Models[0]) },
		"one fewer": func(m *Manifest) { m.Models = m.Models[:0] },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := resign(t, pack, in.Key, editManifest(t, edit))
			_, err := Verify(mutated, pub)
			wantCheck(t, err, CheckReceipt, "the manifest names")
		})
	}
}

// internal/evidence — U18
// The manifest's commits are redacted field by field (pack.go:149-152), apart
// from commits.json. Only the baseline is pinned by a test today; the target
// and the head are the fields this test mutates.
func TestASecretInTheTargetOrHeadReachesNoFile(t *testing.T) {
	in, pub := sampleInput(t)
	in.Commits.Target = "/srv/" + secretMarker
	in.Commits.Head = "refs/heads/" + secretMarker
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if bytes.Contains(body, []byte(secretMarker)) {
			t.Errorf("%s holds the configured secret", name)
		}
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatal(err)
	}
}

// internal/cli — U22 (imports: strings, testing, internal/redact, internal/settingspage)
// A live (non-sample) run records every documented setting. Each secret
// setting must be recorded as "set" or "unset", and no value of one may appear,
// even a value too short for the redactor. The mutated fields are the values
// of the secret settings.
func TestALiveRunRecordsNoSecretSettingValue(t *testing.T) {
	env := map[string]string{}
	for _, sec := range settingspage.Catalog() {
		for _, row := range sec.Rows {
			if row.Secret {
				env[row.Name] = ""
			}
		}
	}
	for _, name := range redact.Names {
		env[name] = ""
	}
	i := 0
	for name := range env {
		env[name] = "k3y" + string(rune('A'+i)) // short: under the redactor's minimum
		i++
	}
	getenv := func(k string) string { return env[k] }
	got := packSettings(config{}, getenv, redact.FromEnv(getenv))
	for name, v := range env {
		if got[name] != "set" {
			t.Errorf("%s is recorded as %q, want set", name, got[name])
		}
		for k, rec := range got {
			if strings.Contains(rec, v) {
				t.Errorf("setting %s holds the value of %s", k, name)
			}
		}
	}
}

// internal/cli — U23 (imports: internal/evidence; uses assertNoControl)
// Every string of a verified report comes from the signer's pack, or from the
// pack's path, and reaches the terminal through visible. The mutated fields are
// each string field of the report and the path, one hostile value in all.
func TestEveryReportFieldIsShownEscaped(t *testing.T) {
	h := "x\x1b[2J" + string(rune(0x202e)) + "\x9by"
	rep := &evidence.Report{
		Created: h, KeyID: h,
		Commits:     evidence.Commits{Target: h, Baseline: h, BaselineResolved: h, Head: h, HeadResolved: h},
		Models:      []evidence.Model{{Agent: h, Route: h, Model: h, ReceiptID: h}},
		ProofCounts: map[string]int{h: 1},
		FailedSvcs:  []string{h},
	}
	assertNoControl(t, renderReport(h, rep))
}
```

Note on U22: once main's `redact.go` is in (it drops the length floor from `New`), the "short value" in that test must be one the rebased redactor still leaves alone, or the assertion must target the `set` value alone; re-check after the rebase.

### Closure
- S1 trust-boundary map:
  - **Pack file → `proven verify`** (the attacker writes every byte). Tampering: closed by the canonical archive + signature + hashes + file set (U1, U3, U7-U9 killed; bit-flip sweep green). Spoofing: a different key is refused by signature and by key id. Disclosure: none (verify reads, prints escaped). Denial: `VerifyFile` reads the whole file with `os.ReadFile` and no bound before the per-entry bound (`verify.go:113`) — a very large file costs memory; owed, not part of the property. Escalation: no extraction to disk; a non-regular entry or a non-plain name is refused; `path.Base` keeps names flat.
  - **Pack text → terminal.** Closed at `3b6bac5`: every pack-derived field passes `visible`, refusals pass `Visible` and then `visible` again. Gap is in tests only (U23).
  - **Environment / review result → pack writer** (service echoes, model routes, test output, query rows can hold a secret). Closed except F1 (spelling gap) and the untested halves U18, U22.
  - **Private key file → signer.** `ownerOnly` follows symlinks via `os.Stat`, refuses any `0077` bit, and the refusal names setting, path (`%q`) and mode, never key bytes. A stat failure falls through to `readPEM`, which names the path.
  - **argv `--baseline-commit`/head → `git rev-parse`.** `--end-of-options` stops a name read as an option; untested.
- S2: the closure member (the pack) was treated as attacker-written; every blocking finding was reproduced by a test that fails (F1 at `3b6bac5`; U2/U14/U18/U22/U23 with the mechanism unplugged).
- S3: authorization checks of the verifier (signature, key id) removed once each: red tests named (U3, U5). The private-key mode check removed: red (U20).

### Remainder
- `FileRef.Size` is signed into the manifest but never compared with the entry's length (`verify.go:170-183`); a signer's wrong size verifies. The report only prints the count, so nothing reads it today. Owed.
- `IsBidi` (`verify.go:78-80`) covers the Trojan-Source set (U+061C, U+200B-200F, U+2028-202E, U+2060-2069, U+FEFF); other invisible runes (U+00AD, U+180E, U+3164, tag characters U+E0000-E007F) pass unescaped. They do not move the cursor or reorder text, so they are outside the stated property; recorded so the owner can rule.
- The `duplicate` subtest (`pack_test.go:262-267`) asserts only `CheckFormat`; U12 survives because U1 also refuses. Owed: assert the detail.
- `ResolveCommit` (`change/between.go:256`) has no test; neither its happy path nor `--end-of-options`. Owed.
- Declined to judge (R11): pack retention and access (docs say the organisation decides); Windows mode semantics (`key.go:61` skips the check, documented); provider receipt signatures (documented as not checked); licences (out of scope by order).

### Profile
security (order `## Profile: security`). S1-S4 performed above. Check ranking (testing-bar §10): every check behind this verdict is a **suite** check (the project's own tests plus my mutants and probe tests). No external or project-config check bears on the property. Mutation on the changed files: 24 mutants (counting each named site once), 18 killed, 6 survived (U2, U12, U14, U18, U22, U23); U12 survives only because a second mechanism still refuses. No bar-weakening move found in the diff: no skip, no suppression, no tolerance edit, no assertion loosened.

### Focus
Can anything the diff touches break the property? Yes, in one place at `3b6bac5`: F1, a configured secret with a specific character mix reaches the pack's JSON files unredacted. Every other clause holds at the tip; five mechanisms that make it hold have no test that would notice their loss (U2, U14, U18, U22, U23).

- master-was-wrong: **"the wagon rebased onto main"** (order § WHAT IS UNDER REVIEW) does not hold. Measured: `git merge-base main 3b6bac5` = `c1e4ba8`; `git rev-parse main` = `5faabc9`; main is ahead by merges (`the-evidence-store-interface`, `security-piece-8-fixes`, `judges-lab-makefile`, `govulncheck-says-what-is-reached`, …) that touch files this diff touches or depends on: `internal/redact/redact.go` (drops `redact.MinLength` and the length floor in `New`), `internal/cli/format.go`, `internal/settingspage/catalog.go`. This verdict judges `3b6bac5` against `c1e4ba8`; the rebase can change the redactor that F1 and U22 depend on. A trial merge could not be run on this seat (Unrun).
- The master's test reading holds: at `3b6bac5` the whole module is green except seven packages / six tests whose failures are identical at `main` (`5faabc9`) and caused by the seat (loopback bind and module download refused).

### Conditions
1. **F1 (BLOCKING — the property does not hold).** `internal/evidence/pack.go:244-253` with `internal/redact/redact.go:83-99`: redact the spelling the pack's own encoder writes (HTML escaping off), or redact values before encoding. Done when `TestReviewProbeASecretInThePacksOwnJSONSpellingIsRedacted` (full text in Differential) passes on the rebased branch.
2. **U2 (BLOCKING — order item 2).** Add the signature-spelling test (Unplug) so `verify.go:151`'s byte-for-byte clause is pinned.
3. **U14 (BLOCKING — order item 2).** Add the model-count test so `verify.go:255` is pinned (and its removal cannot turn into a panic).
4. **U18 (BLOCKING — order item 2).** Add the target/head redaction test for `pack.go:149,152`.
5. **U22 (BLOCKING — order item 2).** Add a non-sample `packSettings` test for `internal/cli/pack.go:109-123`; re-check its short-value assumption against main's redactor after the rebase.
6. **U23 (BLOCKING — order item 2).** Add the every-field render test for `internal/cli/pack.go:191,193,206,209,223,229`.
7. **Rebase (BLOCKING for the gate, not for the code).** Rebase `fix/evidence-pack` onto current main (`5faabc9` or later), resolve against main's `redact.go` / `format.go` / `catalog.go`, and re-run `go test ./internal/evidence/ ./internal/cli/ ./internal/redact/` plus conditions 1-6 at the rebased tip.
8. Owed, not blocking (R6): assert the detail in the `duplicate` subtest (U12); verify `FileRef.Size` or drop it; bound `VerifyFile`'s read; test `ResolveCommit` including `--end-of-options`.

### Unrun
- Trial merge of `3b6bac5` into current main: `git merge-tree --write-tree --name-only main 3b6bac5` was refused by the seat's permission layer ("This command requires approval"); so were `git -C … diff`, a second worktree (`git worktree add` → `Operation not permitted` on the shared `.git`), and shell loops (`for` is not in the review class). Whether the branch merges cleanly into `5faabc9` and stays green there is **UNMEASURED**.
- Network-bound tests (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`) and `internal/cli/sarifschema`: red on this seat at both tip and main for seat reasons (loopback bind / module download refused). Their behaviour with the diff is not measured here; none of them is in the diff.
- Worktree cleanup not completed: `git clean -f -- █████████████████████████████████████████ ████████████████████████████████████` was refused ("This command requires approval"), and `rm` is not in the review class. Both untracked probe files are still in the worktree. Each is emptied to a bare package clause (vet clean; no test in it), so nothing they hold affects a test run. They must be deleted by a role that may do so. Every tracked file is restored: `git diff --stat` is empty.
- `TestAPrivateKeyThatGroupOrOthersCanReachIsRefused` was run on darwin only; the Windows branch of `ownerOnly` (`key.go:61`) was not run.

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 2 |
| private credential or variable name | 2 |
