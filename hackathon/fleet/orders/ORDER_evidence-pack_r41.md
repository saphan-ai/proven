# ORDER — evidence-pack · round 41

- change: evidence-pack
- round: 41
- base: main @ c1e4ba8
- executor: claude-code
- written: 2026-10-24T18:14:41Z
- task: round 4 — a verified pack holds every required file and no configured secret in any spelling, and every verify check is pinned by a test

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — evidence-pack
track: proven-enterprise
- train: proven-enterprise
- wagon: W2
- review: named
- repo: proven
- base: main @ c1e4ba8
- branch: fix/evidence-pack
- lands-on: main
- backend: claude-code
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: round 4 — a verified pack holds every required file and no configured secret in any spelling, and every verify check is pinned by a test

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[evidence-pack] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23: ADR-0035 — a signed evidence pack per review and `proven verify`; *"the settings used (never a secret)"*.

### THE REVIEWS OF ROUND 3 (tip 3b6bac5) — PASTED (law #44)
codex: **reject**, two BLOCKING. claude-code: **accept-with-conditions**, with conditions it marks BLOCKING.
- **codex B1 · required payload files can disappear from a verified pack.** `internal/evidence/verify.go:160-182`
  treats the signed manifest's list as the whole contract and rejects only archive entries absent from it;
  `settings.json` and `commits.json` are required by `docs/EVIDENCE_PACK.md:10-22`, but a manifest without them
  verifies. **Fix:** Verify requires every file the pack format declares, by name.
- **codex B2 · a short configured secret can enter the signed pack.** `pack.go:204-239` relies on the redactor,
  which ignores values shorter than six bytes (`redact.go:54-67`), while the settings accept any non-empty value.
  **Fix:** writing a pack refuses, by name, when any configured secret is shorter than the redactor's floor (the
  general refusal at load is proven-reviewer-core's pending fix; do not depend on it).
- **claude F1 · a secret in the pack's own JSON spelling is not redacted.** `pack.go:244-253` with `redact.go:83-99`:
  redact the spelling the pack's encoder writes (HTML escaping off), or redact values before encoding.
- **claude U2, U14, U18, U22, U23 (missing controls):** pin `verify.go:151` (signature spelling byte for byte),
  `verify.go:255` (model count vs receipts, no panic), `pack.go:149,152` (target/head redaction), a non-sample
  `packSettings` test for `internal/cli/pack.go:109-123`, and every rendered report field
  (`internal/cli/pack.go:191,193,206,209,223,229`).
- **claude owed (not blocking):** assert the detail in the `duplicate` subtest; verify `FileRef.Size` or drop it; bound
  `VerifyFile`'s read; test `ResolveCommit` including `--end-of-options`.

The reviewers' tests, verbatim (claude-code, then codex):
```go
// A configured secret must reach no file of the pack whatever characters it
// holds. The pack writes JSON with HTML escaping off; the redactor's needles are
// the raw value, %q, %+q and json.Marshal (HTML escaping on). A value holding a
// quote, an HTML character and a character %q escapes but JSON does not
// (U+00AD here) is written by the pack in a spelling that is none of them.
// A fix must make every file of the pack free of the value in every spelling
// the pack's own encoder gives it.
func TestReviewProbeASecretInThePacksOwnJSONSpellingIsRedacted(t *testing.T) {
	secret := "tok\"en<12\u00ad34"
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

### WHAT TO BUILD
1. **Rebase first:** `git rebase --onto main c1e4ba8` (main moved: the evidence-store interface, the redactor). Do
   not merge. Resolve against main's `redact.go`, `format.go`, `catalog.go`.
2. codex B1, B2 and claude F1; the five missing controls; the owed items where cheap.

### ACCEPTANCE
1. `go test -count=1 ./...` (whole module) green; `go vet ./...` clean.
2. Every reviewer test above passes (adapted to the package's style if needed) and was red on 3b6bac5's behaviour where
   it tests a defect.
3. Negative controls, named: remove the required-files check (B1), the short-secret refusal (B2), the JSON-spelling
   redaction (F1) — each turns its test red by name. Restore all.
4. Subjects match `^\[evidence-pack\] [a-z0-9/._-]+: .+$`, at most 72 characters; no `Co-authored-by` line.

### WHAT IS NOT IN SCOPE
- W1, W3. JSON or SARIF writers. The general short-secret refusal at load (proven-reviewer-core). Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, each finding with
its test, the negative controls, the whole-module counts. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private credential or variable name | 2 |
| private workspace path | 4 |
