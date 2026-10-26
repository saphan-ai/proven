# REVIEW — json-sarif-outputs · round 6

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T04:29:17Z

## Summary

f4819cf: the JSON and SARIF writers and the load-time alias refusal meet items 1-6 for -o and every --*-output pair; 10 unplugs, all killed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

Reviewer: claude-code (Anthropic family), leg json-sarif-outputs-review-r6, 2026-10-24.
Verdict: **accept-with-conditions**. One condition (C1) is a defect of the same class as the round-2 BLOCKING; I file it
as a condition because the fix is small and local. The master may raise its class (law #53); I do not lower it.

### Ground

- `git rev-parse --short=7 HEAD` printed `f4819cf`. Scope: `git diff main...f4819cf`, 20 files, +5660/−19 (stat measured).
- `git diff main...f4819cf -- AGENTS.md internal/review | wc -l` → `0`. Boundary held.
- I read every changed Go file in full: `internal/cli/config.go` (refuseCollidingOutputs, outputIdentity, splitRoot,
  splitComponents), `run.go` (writeOutputs plus the write-time guard), `format.go`, `format_json.go`, `format_sarif.go`,
  all five test files, `sarifschema/{doc.go,schema_test.go}`, `docs/OUTPUT_FORMATS.md`, `testdata/sarif/README.md`. I also
  read the callers: `Run` maps a `writeOutputs` error to `exitFatal` (2), `run.go:100-103`; `refuse` builds a `cliError`
  with `exitConfig` (7), `config.go:812-815`; `resolve` calls `refuseCollidingOutputs` after `cfg.output` and the
  `--*-output` paths are set, `config.go:391`.
- Measured on this host (darwin, sandboxed, no network, no listen):
  - `go test -count=1 ./internal/cli/` → `ok github.com/saphan-ai/proven/internal/cli 19.210s`.
  - `go test ./...` → 30 `ok`, 7 `FAIL`. Six of the seven (`cmd/livecheck`, `internal/{generator,guardrail,pipeline,query,review}`)
    panic with `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. That is the
    sandbox, in packages this diff does not touch. The seventh, `internal/cli/sarifschema [setup failed]`, cannot fetch
    `github.com/santhosh-tekuri/jsonschema/█████████` (`tls: failed to verify certificate`). The module is not in the
    cache: `.../cache/download/github.com/santhosh-tekuri/jsonschema/v5/@v/` holds only `v5.3.1.lock`.
  - `go vet ./...` stops at the same missing module. `go vet ./internal/cli/` is clean.
  - Vendored schema: sha256 `c3b4bb2d6093897483348925aaa73af03b3e3f4bd4ca38cef26dcb4212a2682e`. Its prefix matches the
    `c3b4bb2d60938974` that `testdata/sarif/README.md` claims. Its `id` is the OASIS errata01 URL. `region.startLine` has
    `"minimum": 1` (line 1716).
  - Dependency delta (R4): one module, `santhosh-tekuri/jsonschema/v5 v5.3.1`, added in `go.mod` and `go.sum`. The only
    importer is `internal/cli/sarifschema/schema_test.go` (`grep -rln santhosh-tekuri internal cmd`), so it is test-only
    and does not reach the shipped binary. Go modules have no install-time scripts.
  - Commit messages `main..f4819cf`: no AI-attribution phrases (grep for co-authored / generated with / vendor names: no output).
  - Non-printable characters in the added code and docs: `grep -c -E '[^[:print:][:space:]]'` → `0`.
    remark: this instrument depends on the locale's `[:print:]`. Whether it classes Unicode format characters (bidi
    controls) as printable was not established, so a `0` here is weak evidence about bidi characters specifically.
- The author's and the master's self-assessment (the order's "master's verification") was set aside as evidence (R1).
  Where I could, I re-measured it, below.

### Differential

Item by item, against WHAT TO JUDGE:

1. **SARIF valid 2.1.0 for every result shape.** The structure half is MET. `TestSARIFShapes` (`format_sarif_test.go:225`)
   covers no evidence, unresolved, a failed service, an empty path, line 0, an end line before the start line, an unknown
   severity and the empty result. The writer's guards match the schema: `startLine >= 1` (`format_sarif.go:162`), endLine
   only when it is at or after startLine, `message.text` never empty (`sarifMessageText`), `ruleId` omitted when empty,
   `level` always one of note/warning/error. The schema half (`sarifschema/schema_test.go:126`, `:27`) is **UNRUN here**
   (see Unrun). By reading it: it validates each shape against the vendored schema, and its negative control removes
   `version` from the sample log and requires a rejection that names `version`. → conditional on C0.
2. **Evidence and the `unresolved` mark reach JSON and SARIF.** MET. `evidenceFrom` is the one function behind
   `findings[].evidence` and SARIF `results[].properties.evidence` (`format_json.go:172`, `format_sarif.go:150`), and
   `TestSARIFPropertiesAreTheContract` asserts the two arrays are equal (`output_contract_test.go:228-232`). Text lines
   versus the formats: class, rule, settings statement, author-model statement, reviewer statements, failed services,
   partial, family sentence and trace path are in the JSON keys and in the SARIF `runs[0].properties`
   (`TestSARIFCarriesEveryTextOnlyLine`, `TestJSONCarriesEveryTextOnlyLine`). The text's per-finding EPSS lines are
   `epss.Phrases(columns, rows)` (`internal/epss/epss.go:226`), a pure rendering of query rows that are already in
   `evidence`, so no data is lost. The SARIF location drops `end_line` when it is before `start_line`, or when
   `start_line < 1`, and no SARIF property carries it. `docs/OUTPUT_FORMATS.md` documents this; see Remainder.
3. **One run writes everything from one result; a write failure is exit 2; aliases are exit 7 for every pair.** MET for
   `-o` and every `--*-output` pair. The load check covers every pair of the four flags
   (`TestCollidingOutputPathsAreRefused`), `..`, a symlinked directory, link-then-`..`, a chain of links, a dangling link,
   a hard link, relative versus absolute and a trailing slash (`aliasTable`, `outputs_alias_test.go:36`). A write failure
   is exit 2 for each flag (`TestOutputWriteFailureExits2`). **NOT MET for standard output:** see C1, measured.
4. **JSON field names match `docs/OUTPUT_FORMATS.md`.** MET. `TestJSONDocumentIsTheContract` decodes the whole document
   and compares it with a literal, exactly. `assertNamesDocumented` checks that every emitted key appears in backticks
   in the doc. Unplugged both ways (see Unplug).
5. **Each claim has a test that fails when reverted.** MET for everything I could run (12 mutants, all killed; see
   Unplug). The schema-test unplug is unrun.
6. **ADR-0027.** MET. Exported names (`Formatter`, `FormatterFor`) have doc comments. Package `sarifschema` has a package
   comment that says why it exists. Each choice carries a WHY comment that names its decision: ADR-0020 §1 at the
   writers and `FormatterFor`, ADR-0013 at `refuseCollidingOutputs`, the RFC 3986 §4.2 reason at `artifactURI`, the
   write-time guard's reason at `run.go:239-243`. Each test comment names the property it proves and the field it mutates.

### Unplug

Each mutant was applied to the source, the named tests were run, and the source was restored with `git checkout --`.
`git diff --stat` is empty afterwards.

| # | mutation | killed by |
|---|---|---|
| U1 | `refuseCollidingOutputs` result ignored (`config.go:391`) | 9 tests red: AliasTable…AtLoad, CollidingOutputPaths…, SymlinkedDirectory…, SymlinkedFileAndHardLink…, RelativeAndAbsolute…, CollisionReproduction, DanglingAlias…, DanglingAliasChains…, LinkLoopAndLongChain… |
| U2 | symlinks treated as plain components in `outputIdentity` | AliasTable (6 link rows) ×2 tests, SymlinkedDirectory, DanglingAlias (12 ordered pairs), DanglingAliasChains, LinkLoop (2) |
| U3 | both `os.SameFile` arms disabled (load and write) | hard-link rows ×3 tests, `TestAliasCreatedAfterLoadIsRefusedAtWrite` |
| U4 | WARNING↔ERROR swapped in `sarifLevel` | `TestSARIFLevelFollowsSeverity` |
| U5 | SARIF `properties.evidence` dropped | `TestSARIFShapes/no_evidence`, `TestSARIFPropertiesAreTheContract` |
| U6 | SARIF `properties.unresolved` dropped | `TestSARIFShapes/unresolved`, `TestSARIFPropertiesAreTheContract` |
| U7 | JSON tag `from_receipt` → `fromReceipt` | `TestJSONDocumentIsTheContract` |
| U8 | doc row `receipt_id` → `receipt` in OUTPUT_FORMATS.md | `TestJSONDocumentIsTheContract`: "does not name: receipt_id" |
| U9 | `stringSlice` returns nil | JSONDocumentIsTheContract, JSONEmptyListsAreArraysNotNull, SARIFPropertiesAreTheContract |
| U10 | region guard `StartLine >= 1` → `>= 0` | `TestSARIFShapes/line_0` |

Not unplugged here: the schema test's own controls (the `sarifschema` package cannot build offline).

### Closure

- Round-2 codex BLOCKING (`--json -o X --sarif-output X` exit 0): **ADDRESSED.** `TestCollisionReproduction` returns
  exit 7 and X is not written. U1 shows it red without the check.
- Round-2 claude C1 (text-only lines in SARIF properties and JSON): **ADDRESSED** (`TestSARIFCarriesEveryTextOnlyLine`,
  `TestJSONCarriesEveryTextOnlyLine`, `family_sentence` added to both).
- C2 (table-driven schema test over edge shapes): **ADDRESSED in code.** The structure table ran green. The schema table
  is unrun here (C0).
- C3 (`level` per result id): **ADDRESSED** (U4 killed).
- C4 (decode-and-compare over every documented field, `[]` not null): **ADDRESSED** (U7, U9, U5 killed).
- C5 (every top-level JSON key named in the doc): **ADDRESSED** (U8 killed).
- Round-4 BLOCKING (two outputs through a symlinked directory exited 0): **ADDRESSED.**
  `TestOutputsThroughASymlinkedDirectoryAreRefused` and the `aliasTable` link rows pass, and U2 turns them red. The
  link-then-`..` row checks the kernel's order (the link is resolved before `..`), which the old Abs-first path got wrong.
- The schema tests moved to `internal/cli/sarifschema` so `internal/cli` builds without the module. **Confirmed by
  measurement:** on this host `internal/cli` is `ok` while `sarifschema` alone fails setup.

### Remainder

Measured with a temporary in-package probe test that calls `cli.Run` with `captureReviewer`, from a fresh `t.TempDir()`:

- **P1, stdout onto an output file (→ C1):** stdout is an `*os.File` opened on `r.json`, with args
  `scan --sarif --json-output r.json` → `exit 0`, stderr empty. The final `r.json` is the **JSON** body: the SARIF
  written to standard output was silently replaced. In a shell this is `proven scan --sarif --json-output r.json > r.json`.
- **P2, letter case on case-insensitive APFS:** `-o Out/r.json --sarif-output out/r.json` (only `out/` exists) →
  `exit 2`, with "-o and --sarif-output are the same file (out/r.json); -o is left as it was written". The doc names
  this gap. The write guard catches it and `-o` stays intact.
- **P3, Unicode normalization:** `-o é.json` (NFC) with `--sarif-output é.json` (NFD) → `exit 2`, same message, `-o`
  intact. **The doc does not name this case** (→ C2).
- **P4:** `-o r.json --sarif-output missing/../r.json` → exit 7, "same file as -o". The kernel would refuse that write
  with ENOENT, so this refusal is harmless but gives the wrong reason. Not a condition.
- **P5:** `-o d --sarif-output d/`, where `d` is a directory → exit 7, "same file". Harmless: the write would fail
  anyway. Not a condition.
- `assertNamesDocumented` (`output_contract_test.go:240`) treats a key as documented if it appears anywhere in
  backticks. A key renamed to another documented leaf name (`model`, `path`) would pass that half. The exact literal
  compare in the same test still catches the rename, so this is a limit of that instrument, not a hole in the contract.
- When `start_line < 1` or `end_line < start_line`, SARIF does not carry `end_line`. The doc states it.
- Declined to judge (R11), for the master to rule on: no `tool.driver.rules` / `version` / `informationUri` in the SARIF
  driver (valid 2.1.0; GitHub code scanning renders rule help from it); absolute `trace_path` in a SARIF uploaded to a
  public code-scanning surface (it is documented as possibly absolute); `familySentence` reads the trace file once per
  format (two or three reads per run).

### Profile

Default (correctness of a writer and a CLI seam), plus the destructive-path part of security S4. Files are written to
user-named paths, and one output overwriting another was the round-2 and round-4 defect. Writes go only to paths the
user names. Nothing is deleted. The load check only stats. The dependency delta is test-only. The non-printable scan
was run (see Ground).

### Focus

The property that must hold: two outputs of one run never end up as one file with exit 0. I spent most of the review on
the identity function: kernel-order resolution of links before `..`, relative link targets taken against the link's
directory, the hop bound, dangling final links, and the second-line write guard. Then on whether each claim's test
actually fails when its code is removed.

### Conditions

- **C0 (gate precondition, from Unrun).** Before the gate, the master runs and pastes, on a host with the module cached:
  `go test -count=1 -v ./internal/cli/sarifschema/` (both `TestSampleSARIFMatchesTheSchema` and
  `TestSARIFShapesMatchTheSchema` PASS) and `go vet ./...` clean. Add one schema unplug: set `startLine` to 0 in one shape
  (or weaken the guard at `format_sarif.go:162` to `>= 0`) and show `TestSARIFShapesMatchTheSchema/line_0` red.
  Remedy owner: master.
- **C1 (blocking-class, item 3; `run.go:198-266` and `config.go:391`).** Standard output, when it is a regular file, is
  not compared with the `--*-output` files. P1 reproduces it: exit 0, and one body silently replaces the other. This is
  the same defect class as the round-2 BLOCKING. Remedy: in `Run`, before the review runs, if `stdout` is an `*os.File`
  whose `Stat` is a regular file, `os.Stat` each `--*-output` path and compare it with `os.SameFile`. The shell has
  already created stdout's file, so an output that is the same file exists under some spelling, and `os.Stat` follows
  links. Also add stdout's `FileInfo` to the `written` list of the write-time guard. Refuse with exit 7, naming "standard output" and the flag. Test:
  `TestStdoutRedirectedOntoAnOutputIsRefused`. Mutated field: stdout is an `*os.File` on the `--json-output` path.
  Unplug: drop the stdout entry, and the test sees exit 0. If the owner rules this out of scope instead, the doc must
  say in one sentence that standard output redirected onto an output file is not detected.
- **C2 (owed, docs; `docs/OUTPUT_FORMATS.md:5`).** Next to the letter-case sentence, name Unicode normalization (NFC/NFD
  on APFS) as the same gap (P3). For both, say what happens: the write-time check stops the run with exit 2 and the
  earlier output stays as written. Not blocking.
- **C3 (owed, tests).** Add schema shapes for a finding whose message, title and id are all blank (`"finding"`, no
  `ruleId`), for an absolute path and a path with `:` in its first segment (`artifactURI`), and for a non-sample finding
  carrying both test and query evidence. The external `shapes` table and the internal `sarifShapes` table are kept in
  sync by hand. A test that compares the two tables' names would stop them drifting. Not blocking.

### Unrun

- `go test ./internal/cli/sarifschema/` (`TestSampleSARIFMatchesTheSchema`, `TestSARIFShapesMatchTheSchema`): refused
  at setup. `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/...": tls: failed to verify
  certificate: x509: OSStatus -26276`. The module is not in the local cache. → C0.
- The schema test's own unplugs: same cause. → C0.
- `go vet ./...` across `sarifschema`: same cause. `go vet ./internal/cli/` is clean.
- The master's "38 packages ok, 0 FAIL": not reproduced. Here it is 30 ok and 7 FAIL: six are sandbox listen denials
  outside the diff, and one is the setup failure above.
- The built-binary check "refuses `-o` and `--sarif-output` on one file with exit 7": the binary built
  (`go build ./cmd/proven`), but the role's command class does not admit running it. The same path was exercised
  in-process through `cli.Run` (`TestCollisionReproduction` and the alias table, exit 7).
- Residue: `git clean`, `rm` and `find -delete` are refused for this role. Two untracked files of mine remain, and the
  master should remove them: `████████████████████████████████████` (cut down to the single line `package cli`, so it
  changes no behaviour) and `██████████████████████████████████████████████` (the built binary and a sha256 helper,
  inside the order's confinement directory). No tracked file is modified (`git diff --stat` is empty).

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: f4819cf: the JSON and SARIF writers and the load-time alias refusal meet items 1-6 for -o and every --*-output pair; 10 unplugs, all killed.
  C1 (blocking-class): stdout redirected onto a --*-output file is not refused; exit 0 and one body is silently replaced (measured).
  C0: the sarifschema tests are UNRUN here (jsonschema module not cached, no network); the master runs them and pastes PASS before the gate.
  C2/C3 owed: the doc should name the NFC/NFD gap (it ends at write-time exit 2); add the missing schema shapes.
  Residue the master must remove: ████████████████████████████████████ and █████████████ (this role is refused rm and git clean).
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T11:40:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| e-mail or user name | 2 |
| file and line in Saphan Studio's private code | 2 |
