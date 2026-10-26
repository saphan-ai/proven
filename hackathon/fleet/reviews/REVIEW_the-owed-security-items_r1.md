# REVIEW — the-owed-security-items · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T20:33:03Z

## Summary

claude-opus-5-5 (claude-code), single review; the NVIDIA half is owed. Of the 17 items, all but two are fixed or fixed at base, with the original reviewers' tests green. The gaps: (C1) existing state dirs and audit files keep 0755/0644 (probes red); (C2) a linked-parent export is skipped, not refused (ruling owed); (C3) a trailing-slash checkout passes the link guard (no caller makes one). Build and vet ok; 33 packages ok, 6 listener-blocked by the sandbox; real container run unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-owed-security-items

Reviewer: claude-code, model claude-opus-5-5. Single review: the NVIDIA (Nightshift) half of the owner's two-review rule is OWED (route down, per the order).

### Ground

- `git rev-parse --short=7 HEAD` → `cc95001`, equal to the order's `- reviewed-sha:`. `go version` → `go1.26.8 darwin/arm64`.
- Scope read: `git log 16cf7b7..cc95001` = six leg commits (`b81467a` modes, `b6d7450` git stderr, `9fd1970` linked proven.yaml + sample limits, `b5a4c41` linked trees/mounts + `--pull never`, `7e13e0d` redactor spellings, `ec76eb7` query gate) and the absorb `cc95001`. The absorb also brings main's `the-test-container-is-locked-down` merge (`07b901d`, `ce40863`, `65dddaf`, `f7e9eef`) and two doc commits. I read those as context, not as the leg's work.
- Production diff read in full (`git diff 16cf7b7 cc95001` over the 14 non-test files). Callers read: `pipeline.go:161,448,458` (redactor wired into the trace and into `recording`), `pipeline.go:385-419` (skipped paths reach the trace as `Kind: skipped`), `pipeline.go:557-558` and `prover.go:141` (checkout paths are built by `filepath.Join`), `investigator.go:51-108` (model SQL refused; every query comes from `querytpl.Build`).
- The original reviewers' notes for P0, P1, P2 and P4 (`████████████████████████`) were read. Their tests were re-run against `cc95001` with `go test -overlay`. The test sources and overlay files are kept in the engine's run directory (`artifacts/the-owed-security-items-review/run-20261023-200740-91ce9152/overlay/`, `…/unplug/`), outside the worktree; see Remainder.
- Item ids: `SECURITY_REVIEW.md` and the leg's order were not provisioned, so I cannot read the P0-F2 … W32 O2 id table. I mapped each item **by content** to the original findings. The mapping is mine and unverified against the id table.
- Full suite, item 5 of the order: `go build ./...` exit 0. `go vet ./...` exit 0, no output. `go test -count=1 ./...`: 33 packages ok, 6 FAIL. All six fail only on `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` (sandbox): `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`. I re-ran `pipeline`, `query` and `review` with every test of the listener files skipped (37 names, the `-skip` list): all three `ok`. `generator`, `guardrail` and `cmd/livecheck` are listener tests throughout: UNMEASURED here.
remark: "33 ok + 6 sandbox FAIL" cannot tell a listener test that would fail on its merits from one that would pass; the listener tests of six packages are UNMEASURED on this seat, not green.

### Differential

- Master's claim (2026-10-24) "39 packages ok": **partly re-measured.** 39 packages have tests (33 + 6). 33 are ok here, and the non-listener parts of 3 more are ok. The listener tests of 6 packages could not run on this seat. Not contradicted, not confirmed.
- Master's claim "the real container run passes at cc95001 on █████████████ with `--pull never`": **UNRUN.** No container runtime is admitted to this seat (`docker` is not in the review class's allow list). The argv carries `--pull never` (`container.go:345`, `TestRunLineNeverPullsTheImage` ok), and an absent image becomes the Prover down with the pull named (`container.go:472-484`, `TestAbsentImageIsTheProverDownAndNamesThePull` ok). The real run is not measured.
- Leg's claim "negative controls for 1, 3, 8 and 17 red": I did not re-run the leg's controls. I ran one of my own (Unplug).
- Order header conflict, named: the order's H1 for `_REVIEW_STATUS.md` reads `# REVIEW — the-owed-security-items-review`. The engine-injected channel grammar requires `# REVIEW — the-owed-security-items`, the reviewed stream, so the verdict binds where `saphan merge` looks. I followed the engine. The order's line is a defect in the order template.

### Unplug

- Redactor spellings (`7e13e0d`): I overlaid `internal/redact/redact.go` with a mutant whose `spellings` returns the raw value only. `go test -overlay …/unplug/overlay.json ./internal/redact/ ./demo/live/review/` → both FAIL: `TestAnEscapedFormOfAValueIsRedactedToo` (`%q`, `%+q` and JSON forms not redacted) and `TestRedactTextHidesAConfiguredSecret` (`%q` form). The control reddens.
- P1 D4, P4 F1/F2 and P2 O1: the original reviewers' tests were red at their SHAs (their notes). They are green here (Closure), so the fixes are what turned them. The leg's own new tests all pass (22 named, `-run` list in Closure).
- Not unplugged by me: `behindALink`, the `proven.yaml` Lstat refusal, the `oneSelect` quote refusal, the mode changes. Each would need a full-file mutant copy. My existing-directory probes below serve as the mode control in the other direction: they are red today.

### Closure

FILE·LINE·CONTROL per item. Mapping by content, as stated in Ground.

| # | Original finding | Fix site | Verdict | Control (run at cc95001) |
|---|---|---|---|---|
| 1 | P0 (codex) piece-8 modes: world, hmem, trace create 0755/0644 | `world.go:142,548,568` · `hmem.go:153,167,536,555` · `trace.go:124,128` | **fixed for a NEW path; NOT fixed for an EXISTING one** | leg `TestStateIsReadableOnlyByItsUser` ×2 + `TestTraceIsReadableOnlyByItsUser` ok. My probes `TestProbeTraceOnExistingDirectory` and `TestProbeHmemOnExistingDirectory` **FAIL**: `logs: mode 755, want 700`, `trace.jsonl: mode 644, want 600`, `state: mode 755`, `hmem.jsonl: mode 644`. → C1 |
| 2 | P0 (codex) meterless-conform session dirs 0755 | `cmd/meterless-conform/main.go:219,222,225` | fixed (new session dirs; a session id is fresh, so an existing dir is not the normal case) | package ok in the full suite |
| 3 | P0 (codex) piece-3 linked target `proven.yaml` followed | `review/config.go:193-208` | fixed: Lstat, refused by name, target never quoted | `TestLinkedConfigIsRefusedByName`, `TestRegularConfigInTheTargetIsStillRead` ok |
| 4 | P1 D3/C3 export reads through a linked parent | `change/archive.go:68-77,211-231` | **fixed in substance, not in the words of C3**: the path is left out and traced as `skipped`; it is not refused | original `TestReviewExportWorktreeStaysInsideTheTarget`: the leak assertion passes, the refusal assertion **FAILS** (`ExportWorktree must refuse…`). Leg `TestExportWorktreeDoesNotCopyThroughALinkedParent` ok. My probe with a two-level link: no byte copied, `skipped=["pkg/sub/a.txt"]`. → C2 |
| 5 | P1 D4/C4 linked prepared trees / linked checkout mounted | `pipeline/trees.go:80-84` · `testrun/container.go:417-427` | fixed | original `TestReviewLinkedTreesAreNotPreparedTrees` ok. Original `TestReviewCheckoutThatIsALinkIsNotMounted` ok (`found=false`, no `-v`). Probe with a trailing-slash spelling **FAILS** (link mounted), but no caller can produce it → C3 |
| 6 | P1 remainder: image pulled at test time | `container.go:345,472-484` | fixed in argv; real run UNRUN | `TestRunLineNeverPullsTheImage`, `TestAbsentImageIsTheProverDownAndNamesThePull` ok |
| 7 | P1 codex F1/F2, claude C1/C2 (hardening, free command) | main's merge (`07b901d`…), at base | fixed at base | original codex hardening test (with `--pull` added) ok. Codex command test adapted to `Name: "printf ready"` → refused, `Found=false` |
| 8 | P2 N3 / codex: runtime and git inherit the environment | `childenv.go`, `between.go:16-23`, `container.go:588` (main) | fixed at base | original `TestReviewSecretNeverReachesRuntimeProcess` ok |
| 9 | git failure text quoted stdout (repository content) | `change/between.go:49-108` | fixed: stderr only, clipped at 512 bytes, without blocking the child | `TestFailureMessageNeverQuotesStandardOutput`, `TestFailureMessageIsClipped`, `TestMessagesRefuseALogOverTheFileBound` ok |
| 10 | P2 F1/F2 service echo → trace/result | `pipeline.go:161,448,458`, `trace/redact.go` (at base) | fixed at base; end-to-end tests UNRUN (listener) | original codex unit test **FAILS**, but it builds `recording` and `trace.Open` with no redactor, which production never does (`pipeline.go:161,448`). It cannot pass by construction. Claude's end-to-end tests are in the tree as `TestRuntimeTokenEchoedInTheReceiptIsNotWritten` / `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten`, and are listener-blocked here |
| 11 | P2 O1/C3 refusal tests that cannot fail | `review/secret_echo_test.go` | fixed | `TestARefusedSecretIsNamedNotPrinted` (3 subtests), `TestARefusedOrgIDIsNamedNotPrinted` ok |
| 12 | P2 O2/C4 duplicated, untested demo redactor | `demo/live/review/main.go:148` → `redact.FromEnv` | fixed | `demo/live/review` test ok; it reddens under the redact unplug |
| 13 | P2 N1 redaction misses escaped spellings | `redact/redact.go:65-97` | fixed for `%q`, `%+q`, JSON (see Focus 4) | `TestAnEscapedFormOfAValueIsRedactedToo` ok, red when unplugged |
| 14 | P2 codex: org id printed / unknown argument echoed | `redact.Names` includes `ZETARIS_ORG_ID` (at base) | fixed | original `TestReviewSecretNeverPrintedOnUnknownArgument`, `TestReviewOrgIDNeverPrintedOnConfigurationError` ok. The service-echo variant is listener-blocked |
| 15 | P4 F1 `quoteLiteral` backslash / codex non-UTF-8 + ESC | `querytpl/querytpl.go:238-248,345-362` (ADR-0032, at base) | fixed at base | original test text 1 aimed at `literal` with a match-anything format (byte backstop alone) ok. Aimed at `OSVPackageSQL` for Go/npm/PyPI ok. Codex OSV controls ok. Leg `TestLiteralByteBackstop…`, `TestEveryIdParameter…`, `TestPackageNamesRefuseEveryControlByte` ok |
| 16 | P4 F2 gate fooled by comment/escape | `query/live.go:201-227` | fixed (at base) and widened by `ec76eb7` (`"` and backtick) | original test text 2, both tests ok. Leg `TestSelectGateSeesAStatementHiddenBehindAnotherQuoteCharacter` ok |
| 17 | P4 codex: model value changes the predicate | `investigator.go:51-55,64,94` (ADR-0032, at base) | fixed at base **by design**: no model SQL reaches `Query` | original `TestModelValueCannotChangeQueryPredicate` still **FAILS** against `oneSelect`. It is one balanced SELECT, and the gate is the second line by its own comment (`live.go:185-190`). The first line, the template, is where the property holds now |
| — | sample review ran under an earlier review's size bounds | `review/open.go:42-52` | fixed | `TestOpenAppliesTheSizeSettings` ok |

**Focus 2 — modes, existing directory: NO.** `os.MkdirAll` and `os.OpenFile(O_CREATE)` set a mode only when they create. A logs or state directory left at 0755 by an earlier Proven, and `trace.jsonl`, `hmem.jsonl` and `world.jsonl` at 0644, keep those modes for ever. `state.json` does become 0600, because it is rewritten through a fresh `.tmp` and a rename, unless a stale `state.json.tmp` survives with the old mode. Measured by the two probes above.

**Focus 3 — links:** `proven.yaml`: refused (Lstat). Export: a tracked path under a linked parent is skipped, and a tracked link was skipped before. Prepared trees: Lstat, so a link is not a prepared tree. Mount source: `Container.Run` refuses a checkout that Lstat reports as a link. The one spelling that passes is a trailing `/` (`lstat(2)` resolves it, and `filepath.Abs` then strips it for `-v`). Every caller joins its paths with `filepath.Join`, so it cannot reach that spelling today.
remark: these are link checks at one instant. A swap between the Lstat and the read or mount (TOCTOU) and a hard link are outside what any of the tests can see.

**Focus 4 — redactor spellings** (probe `TestProbeRedactSpellings`, six values × 19 forms):
- covered: raw; `%q`; `%+q`; `%#q`; `%v`/`%#v`; `json.Marshal`; JSON with `SetEscapeHTML(false)` (the trace's encoder); a value refused for a trailing line break and printed `%q` (the trim makes the core match); inside a `Bearer` header.
- not covered: double escaping (`%q` of `%q`, JSON of `%q`) when the value holds `"`, `\`, a control byte or invalid UTF-8; URL query/path escaping and HTML entity escaping when the value holds a reserved or `"<&>` byte; base64; hex; case change; any prefix (a 128-byte clip of a longer value, `secret[:8]`).
remark: for a value made only of letters, digits, `-` and `_` (the usual key shape), the only misses are base64, hex, case and prefix. The double-escape and URL/HTML misses need an escapable byte in the key.
remark: the trace redacts event text before JSON encoding (`trace/redact.go`), so a single escape level is what the trace path needs. I did not trace every other writer to that depth.

### Remainder

- Artifact place, named: the order puts artifacts under `████████████████████████████████████████████`. That directory does not exist. Creating it would leave `?? ████████` in the worktree: `████████` is not in `.gitignore`, and `mkdir`/`rm` are not in my grant, so I could not remove it after. I put the overlay test sources in the engine's run directory instead (`██████████████████████████████████████████████████████████████████████████████████████████████████████████████████` and `…/unplug/`), an admitted directory outside the worktree. Ruling: "leave the worktree as you found it" outranks the artifact path. Cost if wrong: the master copies 11 small files. `git status --porcelain` after the work equals the start, plus this file and `_REVIEW_STATUS.md`.
- Declined to judge (R11), returned to the master or owner: (a) `MakeReadable` (main's `65dddaf`, absorbed, not the leg's) makes the exported trees under `p.workDir` 0755/0644 so the container user can read them. While a review runs, the reviewed repository's content is readable by other local users. That is an ADR-0025 trade against "readable only by its user", and the owner rules on it. (b) Whether `oneSelect` should also refuse `$` (dollar quoting: `SELECT $$'$$ FROM t; DELETE FROM t; SELECT '` → `oneSelect=true`). Spark SQL does not read dollar quotes, and the gate is the second line behind ADR-0032, so I class it as a NOTE.
- Skipped-path trace text: a path skipped for a linked parent is traced as `Text: "symlink"` (`pipeline.go:407`), but the path itself is not a link. This is a wording NOTE.
- Item-id mapping unverified (Ground). The W31 C3–C6 and W32 O1–O2 notes were not provisioned. Rows 11–14 and the sample-limits row are my best match to those ids.

### Profile

security (R9: confinement, secrets, query construction, file modes).
- S1 boundaries: change working tree → export/mount (links: Focus 3); change config → reader (`proven.yaml` link refused); Proven → children (git/runtime env minimal, measured); service text → trace/result (redactor, Focus 4); model → SQL (templates only; the gate is the second line); Proven → disk (modes: Focus 2, the existing-path gap).
- S2: each condition below is reproduced by a test I ran (C1 probes, C2 original test, C3 probe). No payload is used: each test reads a mode, a path or a boolean.
- S3: one guard removed (redact spellings), and its tests turn red. No authorization object is in this diff.
- S4: no dependency delta (`go.mod` untouched in `16cf7b7..cc95001`). No destructive operation is added. `os.RemoveAll` exists only on Proven-made temp copies (`container.go:435,440`). Invisible/bidi scan: not run.

### Focus

1. Items: rows 2, 3, 5–16 and the sample-limits row are fixed, with the notes given (6: the real run is UNRUN; 10: end-to-end UNRUN). Row 1 is fixed for new paths only (C1). Row 4 is fixed in substance, by skip rather than refusal (C2). Row 17 is fixed at base by design, and the original gate test stays red by definition.
2. Modes: new paths 0700/0600 yes. Existing directories and append-only files: **no** (C1).
3. Links: none followed by any caller's spelling. The `Container.Run` guard misses a trailing-slash spelling that no caller produces (C3).
4. Redactor: Focus 4 table.
5. Build/vet/test: build ok, vet ok. Tests: 33 ok, 6 sandbox-blocked listener packages, and 3 of those ok without their listener tests.
master-was-wrong: not shown. Both dated claims are UNMEASURED on this seat (Differential), not contradicted.

### Conditions

- **C1 — OWED, to land before merge (property "Proven writes its own state readable only by its user").** `trace.go:124-128`, `hmem.go:153,167,536`, `world.go:142,568`: bring an existing directory to 0700 and an existing state or audit file to 0600 when opening it (`os.Chmod` after `MkdirAll`, `f.Chmod` after `OpenFile`), or refuse a wider mode by name. Proof: `TestProbeTraceOnExistingDirectory` and `TestProbeHmemOnExistingDirectory` (texts in the run directory) turn green, plus a world twin. Negative control: drop the chmod and they redden.
- **C2 — OWED, a ruling for the master.** Original P1 C3 said "refuses"; the leg skips and traces. I recommend recording skip-and-name as meeting C3: it matches how a tracked link is treated, and no byte from outside is copied. Then reword the trace text for a linked parent (`pipeline.go:407`). The alternative is to make `ExportWorktree` return an error. Either way, the original test's refusal assertion is the open half.
- **C3 — OWED, defence in depth, not blocking.** `container.go:419-426`: Lstat the `filepath.Clean`ed (or `filepath.Abs`) path, the same string that becomes the `-v` source. Proof: `TestProbeCheckoutLinkWithTrailingSlash` turns green.
- **C4 — OWED by the owner's rule.** The second, NVIDIA review through Nightshift. Its route is down today (the order).

### Unrun

- Real container run (`--pull never`, hardened argv): `docker` is not in this role's allow list. The master's █████████████ measurement stands unmeasured by me.
- The httptest listener tests of `cmd/livecheck`, `internal/generator`, `internal/guardrail`, and the listener files of `internal/pipeline`, `internal/query`, `internal/review`: `listen tcp6 [::1]:0: bind: operation not permitted`. This includes P2 claude's F1/F2 end-to-end tests and P2 codex's service-echo livecheck test. I did not use the sandbox override (law #57).
- The leg's tests run at base `16cf7b7` (red before the fix): this needs a second checkout, which this role cannot make (`git worktree` not attempted; `mkdir` denied).
- Unplug of `behindALink`, the `proven.yaml` refusal, the `oneSelect` quote refusal and the mode changes (each needs a full-file mutant).
- `git check-ignore`: answered "This command requires approval", so it was not used. Ignore status was read from `.gitignore` and `git status --ignored` instead.
- Invisible/bidirectional character scan of the diff.
- The id table (`SECURITY_REVIEW.md`, W31/W32 notes): not provisioned.

## The review status as filed

## REVIEW — the-owed-security-items
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: claude-opus-5-5 (claude-code), single review; the NVIDIA half is owed. Of the 17 items, all but two are fixed or fixed at base, with the original reviewers' tests green. The gaps: (C1) existing state dirs and audit files keep 0755/0644 (probes red); (C2) a linked-parent export is skipped, not refused (ruling owed); (C3) a trailing-slash checkout passes the link guard (no caller makes one). Build and vet ok; 33 packages ok, 6 listener-blocked by the sandbox; real container run unrun.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T00:00:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 4 |
| machine name | 2 |
| private workspace path | 1 |
