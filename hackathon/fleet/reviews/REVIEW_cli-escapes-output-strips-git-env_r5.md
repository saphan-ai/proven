# REVIEW — cli-escapes-output-strips-git-env · round 5

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T06:03:42Z

## Summary

claude-code review of a78d7d8: P4-F5 (text+stderr escape C0/DEL/C1/malformed UTF-8, marked continuation lines) and P2-F3 (merge-base git child on childenv.Minimal) hold; all 7 overlay unplugs turn red; TestCIMergeBaseIsTheForkPoint green.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-escapes-output-strips-git-env

Reviewer: claude-code (Anthropic family), run cli-escapes-output-strips-git-env-review-r6, 2026-10-24.
`git rev-parse --short=7 HEAD` printed `a78d7d8`. Verdict: **accept-with-conditions** (one documentation condition, C1).

### Ground

- HEAD `a78d7d8`; `git merge-base 0b954a9 a78d7d8` = `0b954a9`.
- The order says the scope is "2 commits". Measured: `git log 0b954a9..a78d7d8 -- internal/cli` shows **5** stream commits
  (025fb8c, 1e77598, 40a9570, 3f96bf0, a78d7d8) and nothing else. The three-dot diff also carries main's merges (91 files),
  and those are out of scope. Every stream commit touches only `internal/cli/` (`git show --stat` on each one).
  `AGENTS.md` and `internal/review` are untouched by the stream. The changes to `internal/review` in the three-dot stat come from main.
- No `Co-Authored-By` / "Generated with" in any message of the range (grep count 0). All subjects match `[slug] area: what`.
- `go vet ./...` is clean. `go test -count=1 ./internal/cli/ ./internal/childenv/` is ok.
  `TestCIMergeBaseIsTheForkPoint` PASS (11.5s) with the git child on `childenv.Minimal()`, so git still works under the minimal env.
- Full suite in this sandbox: 34 ok, 3 with no test files, 6 FAIL. All 6 failures are the same
  `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`, so this sandbox
  refuses a loopback listen. They are cmd/livecheck, internal/generator, guardrail, pipeline, query and review, and none is in the stream's write-set.
  The master's "40 packages ok" is **not reproduced here**; see Unrun.

### Differential

Property → site → control (FILE·LINE·CONTROL):

1. **Every change-borne text field is escaped.** I counted every print site in `internal/cli` (`grep -n Fprint|.Write(`):
   - `format.go:56,59,63,68,73,88,99,104` cover class, rule, settings, author-model, reviewer statements, failed service and reason, family sentence and trace path.
   - `format.go:149,157,167,170,176` cover name/ID, severity, title, EPSS phrases, location path, message and rule id.
   - `format.go:198,201,204` cover test path, name, command and output. Output goes through `visibleLines`.
   - `format.go:212,215,218,222` cover engine, snapshot, SQL, columns and rows.
   - The remaining sites print only constant text: `format.go:83,94,163` (fixed literals) and `run.go:42,46` (help/version constants).
   - Every stderr error goes through `printErr` (`run.go:116-118`). The verbose line (`run.go:99`) is escaped too.
   - `run.go:237` writes `primary`, which is the already-escaped text render.
   - **No field from the order's list is printed raw.** Control: `TestTextOutputShowsEveryPrintedFieldWithoutControlBytes`
     plants raw bytes under 23 field names plus FamilySentence and TracePath.
2. **The escaping covers the whole claimed range.** In `escapeControls` (`format.go:260-277`):
   - every malformed byte is checked first and shown as `\xNN` of the original byte;
   - `isControl` covers `<0x20`, `0x7f` and `0x80–0x9f` (the C1 code points, i.e. the 2-byte `\xc2\x80–\xc2\x9f` UTF-8 forms);
   - a raw single C1 byte (e.g. `\x9b`) is not valid UTF-8, so it takes the malformed-byte arm and is still escaped.

   Newline is kept only in `visibleLines`, and every continuation line there starts with `"    | "`. I checked Proven's own text lines:
   all of them start at column 0 or with exactly two spaces, so a forged "failed services: none" cannot appear at column 0.
   Controls: `TestControlContractBytes` and `TestMultiLineFieldCannotForgeProvenLine`.
3. **The git child does not inherit Proven's environment.** `run.go:339` sets `cmd.Env = childenv.Minimal()`, which passes only PATH and HOME (`childenv.go`).
   `gitMergeBase` is the only `exec.Command` in `internal/cli`. Control: `TestGitMergeBaseChildCarriesNoKeyFromProven`, which runs a fake `git` on PATH
   that dumps its env. It asserts that both the marker and the key name are absent and that PATH is present.
4. **JSON and SARIF are byte-identical.** True, but trivially so. `FormatterFor` refuses both `json` and `sarif` ("has no writer in this build")
   at base `0b954a9` (format.go:27,29) and at the tip alike. Neither format has a writer to compare.
5. **Tests**: see Unplug. All seven unplugs turned red.
6. **ADR-0027 WHY comments**:
   - The escaping has its byte contract written once (`format.go:236-247`). The continuation prefix's reason is at `format.go:230-234`.
   - The env line carries a `Property:` WHY comment (`run.go:336-338`).
   - Every test's doc comment names its property.
   - One sentence in the escaping comment is false; see Conditions.

### Unplug

Method: `go test -overlay review/overlay.json` swaps in copies of `format.go`/`run.go` (`review/*.gotxt`) whose switch is read from `review/unplug.txt`.
The checked-out source was never edited. Control run with `none`: ok. Each unplug, with the tests that turned red:

| unplug | what is removed | red |
|---|---|---|
| `all` | `escapeControls` returns its input | Stderr (both subtests), TextOutputShowsEveryPrintedField, MultiLineField, ControlContractBytes |
| `path` | `visible` dropped at the one call site `format.go:167` | TextOutputShowsEveryPrintedField, naming `field LocationPath` + `raw byte 0x1b` |
| `utf8` | malformed byte written raw | TextOutputShowsEveryPrintedField, ControlContractBytes |
| `prefix` | kept newline without `continuationPrefix` | MultiLineFieldCannotForgeProvenLine, ControlContractBytes |
| `c1` | U+0080–U+009F written raw | TextOutputShowsEveryPrintedField, ControlContractBytes |
| `stderr` | `printErr` writes `err.Error()` raw | StderrCarriesNoRawControlByte (both subtests) |
| `env` | `cmd.Env = nil` (inherit) | GitMergeBaseChildCarriesNoKeyFromProven ("the git child inherited Proven's key") |

remark: the overlay unplugs measure these seven removals only. A raw print site added later is caught only if a test fills it.

### Closure

- P4-F5 is closed for the text writer and for stderr. A finding whose fields carry ESC reaches the terminal as `\x1b`, and the tests prove it with a red unplug.
- P2-F3 is closed for the one subprocess `internal/cli` starts. The marker test is red with the fix removed, and git still resolves the fork point under the minimal env.
- The round-2 answers hold under measurement: the continuation prefix, tab as `\x09`, and a malformed byte shown as its original value.
- The round-3/4 answers hold under measurement: stderr goes through the same encoder, and the stderr test exits on its error with code 2 in both subtests.

### Remainder

None of these blocks the verdict. They are recorded for the train.
- **R1, ambiguity rather than control.** A literal backslash is not escaped, so change text containing the four characters `\x1b`
  prints exactly like an escaped ESC. No byte reaches the terminal raw, but a reader cannot tell the two apart.
- **R2, outside the claimed range.** Unicode format and bidi characters pass unchanged: U+200B–U+200F, U+202A–U+202E, U+2066–U+2069,
  and U+2028/U+2029. These cannot drive a terminal, but they can reorder how a path or title displays. The contract does not claim them.
- **R3, test gap.** The verbose line at `run.go:99` (target, baseline, head) is escaped but has no test, so removing that `visible` turns nothing red.
  Its values come from flags or CI variables, not from the change.
- **R4, the scope of property 3.** It is verified for `internal/cli`. Subprocesses started deeper down (the container runtime in `internal/testrun`)
  were already on `childenv` and were not part of this diff.

### Profile

Security class, terminal-output encoding and subprocess environment. Small surface: 2 production files and about 60 production lines.
The tests are proportionate: one byte-contract table, one test that fills every field, one forging test, one stderr integration test and one fake-git env test.
The work matches the fleet's testing bar: every property has a test whose false-world answer differs, as shown in Unplug.

### Focus

1. Field completeness: walked every print site.
2. Totality of the byte range: measured via the unplugs, including that a lone raw `\x9b` takes the malformed arm.
3. Line forging through the kept newline: checked that no Proven line starts with `"    | "`.
4. Git environment: unplugged, plus a regression run of `TestCIMergeBaseIsTheForkPoint`.
5. The JSON/SARIF claim: measured `encoding/json` directly.

### Conditions

- **C1 (documentation, ADR-0027; class: must fix before or with merge, no code behaviour change).** `internal/cli/format.go:247` says
  "JSON and SARIF need no such step: their encoders already escape." Measured on this toolchain:
  `json.Marshal("a\x1bb\u009bc\x7fd")` → `"a\u001bb` + raw U+009B + `c` + raw DEL + `d"`. So `encoding/json` escapes C0 but leaves
  **C1 and DEL raw** (probe: `go run review/jsonc1/main.go`). Neither writer exists in this build either (`format.go:27-30`).
  The sentence states a property that is false for half the contract's range, and it will mislead whoever builds the JSON/SARIF writer.
  Fix: correct the sentence. For example, say that no JSON/SARIF writer exists yet, and that `encoding/json` leaves C1 and DEL raw, so a future writer
  that is meant for a terminal must decide that step itself. The stream's own claim, "JSON and SARIF unchanged", is true, and only vacuously.

### Unrun

- **Six packages did not run in this sandbox**: cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query and internal/review.
  Each panics at `httptest` with `bind: operation not permitted`, so the rest of each package was never exercised here.
  None is in the stream's write-set (every stream commit touches only `internal/cli/`). The master's "40 packages ok" is unverified by this leg.
- I did not do a live terminal render, i.e. no `proven scan` on a real terminal, which the order forbids inside the checkout. The properties are proven on the writer's bytes.

## The review status as filed

## REVIEW — cli-escapes-output-strips-git-env
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: claude-code review of a78d7d8: P4-F5 (text+stderr escape C0/DEL/C1/malformed UTF-8, marked continuation lines) and P2-F3 (merge-base git child on childenv.Minimal) hold; all 7 overlay unplugs turn red; TestCIMergeBaseIsTheForkPoint green.
  One condition C1: format.go:247 claims JSON/SARIF encoders already escape — measured false for C1 and DEL (encoding/json leaves them raw), and neither writer exists; correct the sentence.
  Unrun: 6 non-stream packages panic on loopback bind in this sandbox; the master's 40-ok is unverified here.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T06:30:00Z

## REDACTED

none
