# REVIEW — the-reviewer-model-and-the-review-class · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T13:32:28Z

## Summary

REJECT on claude-opus-5-5 (Anthropic), a9aced6...2d373c2. B1 blocking: author_record follows symlinks out of the review target, and a file outside reaches the printed author statement (probe P4, settings supplied by the reviewed tree's own proven.yaml).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-reviewer-model-and-the-review-class

Reviewer model: claude-code `claude-opus-5-5` (Anthropic). Scope: `a9aced6...2d373c2`, HEAD measured `2d373c2`.
Verdict: **reject**. One blocking finding (B1). Everything else is owed work.

### Ground
- `git rev-parse --short=7 HEAD` → `2d373c2`. `git diff --stat a9aced6...2d373c2` → 13 files, +1480 −79; commits `2aaa777` and `2d373c2`.
- Specification read from the worktree: ADR-0008 to ADR-0014 in `fleet/decisions/`, plus ADR-0020 (where `proven.yaml` is loaded from) and `AGENTS.md` § Configuration.
- `go build ./...` and `go vet ./...` → clean (go1.26.8 darwin/arm64).
- `go test -count=1 ./...` → 5 packages FAIL (generator, guardrail, pipeline, query, review). Each red is `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`: 5 of 5 panics counted by grep. **Classification: environment** (this leg's sandbox refuses a localhost bind), not the diff. A panic aborts the rest of its package, so I re-ran with the listener tests skipped: `internal/review`, `internal/pipeline`, `internal/guardrail` → **ok**. `internal/generator` and `internal/query` still hit a listener test on every skip, and the diff touches neither. `cmd/proven`, `internal/change` → **ok** on the full run.
- `go run ./cmd/proven scan --sample testdata/fixture/discount` from the worktree root → exit 2: `hmem: state directory …/state/hmem is inside the repository`. **Classification: pre-existing, not the diff.** The guard comes from `0596982` (`git merge-base --is-ancestor 0596982 a9aced6` → true), and the state directory is derived from `LogDir = wd/logs` (`internal/pipeline/memory.go:26`). The same scan run from a temporary directory is `TestScanSamplePrintsF1AndF2NotF3` (`cmd/proven/main_test.go:11`), and it is green. It asserts the new lines `review class: security`, `review rule: go.mod`, `author model not stated; default reviewer NVIDIA` and `special-review model: not configured`.

### Differential
The master's claims (law #44), re-measured:
- "tests green at 2d373c2": **holds only where measurable.** Green except the listener tests, which this host cannot run (Ground).
- "unknown family `Google` and unknown class `frontend` each refused with exit 7": **holds.** `TestScanUnknownFamilyExits7` and `TestScanUnknownClassExits7` are green, and `TestConfigRefusals` (`internal/review/decide_test.go:176`) is green.
- "removing highest-risk-wins turns `TestHighestRiskWinsOverEarlierLowerRule` and `TestMobileOutranksUserInterface` red": **holds exactly.** Mutant M1 (Unplug) turns those two red and nothing else.
- Not claimed by the master and found here: B1 (a file outside the target reaches the review output) and O1 (the reviewed tree's own `proven.yaml` lowers its class).

### Unplug
Mutants were applied with `go test -overlay`. The mutated copies live only in the run artifact directory (`…/artifacts/the-reviewer-model-and-the-review-class-review/run-20261022-132107-d573211e/unplug/`), and no source file was edited.
- **M1 · ADR-0014 point 3:** `highest()` ignores the class order (`decide.go:99`, `if r < bestRank` → `if false && …`). RED: `TestHighestRiskWinsOverEarlierLowerRule` and `TestMobileOutranksUserInterface`. All other tests green. **Held, guarded.**
- **M2 · ADR-0014 point 4 + ADR-0011 point 2:** in `chooseClass` the fallback returns the literal `"backend"` (`decide.go:85`), and `readAuthors` ignores `cfg.ReadCommits` (`decide.go:180`). Result: **all green, so the mutant survives.** No test sets a non-default `fallback_class`, and no test sets `author_commit_trailers: off`. → O3.
- Probes were added by overlay as one test file (`…/probe/zz_probe_test.go`), never in the tree:
  - P1: the reviewed tree adds `proven.yaml` with `path_rules: **=docs`, `doubled_classes: off` and docs first in the order. Class `security` by `go.mod`, doubled, becomes **`docs` by `**`, not doubled**. → O1.
  - P2: `--config` sets `author_record: MODEL.txt`, and `MODEL.txt` in the target is a symlink to a file outside it. Result: `author Outside:outside-model … from record`, **read from outside the target**. → B1.
  - P4: P1 and P2 together, with no `--config`. The target's own `proven.yaml` sets `author_record: MODEL.txt` and `author_trailer_keys: https`, and the symlink points to an outside file holding `https://user:███████████████████████████`. The statement printed is `author //user:███████████████████████████ family family unknown from record`. → B1.
  - P3: malformed and unknown inputs. `Assisted-by: claude-opus-5-5` (no agent) → "author model not stated". `Gemini:gemini-3-pro` → "family unknown", and the review continues. `Co-authored-by: Claude <█████████████████████>` → "author model not stated". An 8 MiB description line followed by a trailer → read, no error. → O4.

### Closure
Each ADR point the order cut: FILE·LINE · the CONTROL that would turn red without it.
- ADR-0011 p1 (`Assisted-by:` by default, read from every commit) — **held.** `config.go` default `TrailerKeys: {"Assisted-by"}`, `decide.go:180` `readAuthors`, `between.go:49` `Messages`. Controls: `TestDescriptionNamesTheAuthorModel` (default key), `TestMessagesExcludesTheBaselineCommit`. Gap: no test reads the default key from a *commit* (the commit test uses a custom key).
- ADR-0011 p2 (sources configurable: commits on, description on, record off) — **held in code.** Commits-off is unguarded (M2). Description: `TestDescriptionOffIgnoresThePullRequest`. Record off by default: `TestRecordFileOffByDefault`. "Order of precedence" is implemented as a union of the sources: declined to judge (Remainder).
- ADR-0011 p3 (trailer keys configurable) — **held.** Control: `TestCommitTrailerUsesTheConfiguredKey`. P3 shows that the usual `Co-authored-by: Name <email>` form never parses as `agent:model` (O4).
- ADR-0011 p4 (family table configurable; unknown → "family unknown") — **held for patterns.** Controls: `TestFamilyTableDefaults`, `TestConfiguredFamilyTableReplacesTheDefault`. The family *names* are a closed list (`config.go` `familyNames`) → O5.
- ADR-0011 p5 (each review states which author models it is independent of) — **held.** `decide.go` `reviewerLine` / `familyRelation`. Control: `TestDescriptionNamesTheAuthorModel` (`… (Anthropic): no`, `… (NVIDIA): yes`).
- ADR-0011 p6 (nothing stated → "author model not stated", default reviewer) — **held.** Controls: `TestNoTrailerSaysAuthorModelNotStated`, `TestAuthorStatementWhenNothingIsNamed`.
- ADR-0011 p7 (never writes a trailer) — **held by reading.** `internal/change/between.go` runs only `git log`, `rev-parse`, `diff` and `ls-files`. The exemption in `internal/pipeline/imports_test.go` names them. No test would redden if a write verb were added (Remainder).
- ADR-0014 p1 (deterministic, by configuration) — **held.** No model is called in `applyDecision`. P1: the configuration may come from the reviewed tree itself (O1).
- ADR-0014 p2 (rules map patterns to classes) — **held.** Control: `TestClassRules`.
- ADR-0014 p3 (highest risk wins, configurable order) — **held, guarded.** M1 red. Control for the configured order: `TestConfiguredOrderCanPutDocsAboveSecurity`.
- ADR-0014 p4 (configurable fallback, backend by default) — **default held** (`TestClassRules` `billing/discount.go` → fallback). **The non-default value is unguarded** (M2).
- ADR-0014 p5 (ships a default rule set, *documented with the other defaults*) — **half held.** `config.go` `defaultPathRules` ships it. No document lists it: the new keys appear in no `.md` file, and `AGENTS.md:73-79` still describes a `timeout`-only file → O2.
- ADR-0014 p6 (the review states its class and the rule) — **held.** `cmd/proven/main.go:96-103`. Control: `TestScanSamplePrintsF1AndF2NotF3` (`review rule: go.mod`).
- ADR-0014 p7 (closed class list; an order that leaves a class out is refused, the refusal lists the allowed values) — **held.** `config.go` `parseClassOrder` / `parseClassTokens` / `parsePathRules`. Controls: `TestConfigRefusals` (asserts a non-empty `Allowed`), `TestScanUnknownClassExits7`.
- ADR-0014 p8 (no heuristic of its own) — **held by reading** `chooseClass`. `raise_class` is parsed and not applied. Control: `TestRaiseClassOnDoesNotChangeTheClass`.
- ADR-0013 p4a (every setting validated on load; a wrong value stops proven) — **held for every key in `configKeys`.** `config.go` `applyKey`, exit 7 via `main.go`. `author_record` refuses an absolute path or `..` on load, but **does not refuse a symlink** (B1).

Seam `internal/review/review.go`: **additive only.** It adds `Request.Description`, nine fields to `Result` and the new type `AuthorModel`. Nothing is renamed or removed (the diff hunk realigns the existing `Result` fields' whitespace; their names and types are unchanged). `open.go` removes `timeoutInTarget`, which is unexported and replaced by `loadSettingsInTarget` with the same behaviour for `timeout`.

Input handling (order item 3):
- No input reaches a shell. `between.go:59` passes argv directly, with `--end-of-options`, and both revisions are checked by `ensureCommit` first.
- No regular expression is compiled from user text (`grep regexp` → none in the changed code). Patterns go to `path.Match`, which is not a regex and is validated on load by `validPattern`.
- **A file path outside the repository: YES**, through a symlink (B1).
- A long input states and continues. An unknown model or family states "family unknown" and continues. A malformed trailer is dropped without a statement (O4).

### Remainder
**BLOCKING**
- **B1 · `internal/review/decide.go:286-297` (`recordPaths`, literal branch) and `:299-323` (glob branch), and `readRecord` `:263` `os.ReadFile`.**
  - **Defect:** `author_record` follows symlinks. `os.Stat` and `os.ReadFile` resolve the link, so a file that sits outside the review target is read.
  - **Guard defeated:** `parseRecord` (`config.go`) refuses an absolute path and `..` precisely to keep this read inside the target, and a symlink inside the target bypasses that.
  - **Asset:** files readable on the host that runs proven (a CI runner).
  - **Reach:** any line of the outside file shaped `Key: a:b`, where `Key` is a configured trailer key, is printed in the review output's author statement. P4 shows that the reviewed tree can supply both `author_record` and `author_trailer_keys` through its own `proven.yaml` (ADR-0020: `proven.yaml` in the target is the default), with no operator setting.
  - **Remedy:** resolve each candidate with `filepath.EvalSymlinks`, require the resolved path to stay under the resolved target, and skip a symlink in the glob walk (or refuse it by name).
  - **Proof expected:** a test that creates `MODEL.txt` as a symlink to a temp file outside the target and asserts a named refusal, or no author from `record`. It is red at `2d373c2` and green after the fix.

**OWED** (R6: not blocking, each with a remedy)
- **O1 · `internal/review/open.go` `Review` → `loadSettingsInTarget`.** The reviewed tree's own `proven.yaml` now chooses its own class and doubling (P1: security → docs, doubled → not). It was harmless while the file held only `timeout`; this diff extends it to the settings that decide how strict the review is.
  - This conflicts in spirit with ADR-0014 p9 ("never lower") and ADR-0012 p2 ("however small"). Loading from the target is ADR-0020's default, so this is **order-mandated in part (R12); the owner decides.**
  - Smallest remedy: a default rule `proven.yaml=security`, so a change to its own configuration is reviewed at least as security. Alternatively, read the configuration from the baseline commit, or state in the review that the settings came from the reviewed tree.
- **O2 · ADR-0014 p5 / ADR-0013 p2.** The new defaults (`path_rules`, `class_order`, `fallback_class`, `doubled_classes`, the family table, the trailer keys and sources) are documented nowhere, and `AGENTS.md:73-79` describes a `timeout`-only file. Remedy: one defaults table in `AGENTS.md` or the README.
- **O3 · ADR-0013 Consequences ("each setting tested at its default and at least one other value").** Surviving mutant M2. Untested non-default values: `fallback_class`, `author_commit_trailers: off`, `default_reviewer`, `doubled_classes` as a list (only `off` is tested, and only parsed), and `author_record` as a glob, absolute (refusal) or with `..` (refusal). Remedy: one table test per key.
- **O4 · `decide.go:211-236` `scanTrailers` / `splitAgentModel`.** A trailer whose value is not `agent:model` (`Assisted-by: claude-opus-5-5`, or the usual `Co-authored-by: Name <email>`) is dropped in silence, and the review then says "author model not stated" although the change stated one. Remedy: report a recognised key with an unreadable value as "author statement unreadable: <line>" rather than "not stated".
- **O5 · `config.go` `familyNames()`.** Family names are a closed list of four, so `gemini-*=Google` is refused on load. A team using Gemini can never get a "yes"/"no" independence label, only "family unknown". ADR-0014 p7's closed list is about classes, and ADR-0011 p4 / ADR-0013 p1 make the family table a setting. Possibly order-mandated (the master measured the refusal as intended): **owner decides.**
- **O6 · `decide_test.go:111-121` `TestSecurityIsDoubledAndBackendIsNot`.** It builds `lines` and then asserts its own literal (`lines[1]`), which is tautological. Only the `cfg.Doubled` asserts measure anything.

**DECLINED TO JUDGE** (R11; the master rules)
- ADR-0011 p2 "order of precedence" is implemented as a union of all enabled sources with de-duplication by model. That is consistent with p5, but it is not a precedence.
- `special_reviewer` defaults to empty, so a default security review is "doubled" with one reviewer plus the line "special-review model: not configured". ADR-0008 p4 says a second model "is configured", while ADR-0010 p2 says no setup is refused.
- `familyRelation` prints `yes`/`no` meaning *same family as the author*. ADR-0009 p5 asks whether the reviewer *is the author's model*, and an exact-model match is not stated separately.
- ADR-0011 p7 has no test that would redden if a write verb were added to `between.go`.
- `matchSegs` recursion with several `**` segments is polynomial in path depth. The patterns can come from the reviewed tree (O1). This was not measured.

**RESIDUE OF THIS LEG**
- **`logs/trace.jsonl` is untracked and was created by the ordered `go run ./cmd/proven scan …`** (`LogDir = wd/logs`). `rm` is outside my permission class (law #57), so the master removes it.
- The scratch files are outside the worktree, in the run artifact directory.

### Profile
Security (R9: the diff adds configuration and a file read driven by configuration), plus backend. S1 trust boundaries:
- the reviewed tree → settings (O1);
- the reviewed tree → file reads (B1);
- commit text and PR description → the review output (bounded; no shell, no regex).

S4: the dependency delta is none (no `go.mod`/`go.sum` change in the diff). I found no invisible or bidirectional characters in the changed files by reading them; no byte scan was run.

### Focus
The order's four items:
- (1) the ADR points: Closure;
- (2) the seam: additive (Closure);
- (3) the input boundaries: B1, O1, O4;
- (4) build, vet and test: Ground.

### Conditions
The verdict is `reject`, so no conditions of acceptance. What a re-review needs (R14: it judges only these and the fix diff):
- **B1 fixed, with its failing-before test.**
- The master rules on O1 and O5 (owner decision, order-mandated in part).
- O2, O3, O4 and O6 are owed and may land in the same round or later.

### Unrun
- Listener-binding tests in `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query` and `internal/review` (for example `TestLiveReviewCarriesTheRoutedModel`). Refusal: `listen tcp6 [::1]:0: bind: operation not permitted`. The sandbox refuses localhost binds. These must be run on a host that permits them.
- `go run ./cmd/proven scan --sample testdata/fixture/discount` **from the worktree root**. It was refused with exit 2 by the pre-existing hmem guard (Ground). Its equivalent run from a temporary directory, `TestScanSamplePrintsF1AndF2NotF3`, is green.
- The independent NVIDIA review through Nightshift, as ordered: that route is down today, so this review ran alone and **the NVIDIA half is owed.**

## The review status as filed

## REVIEW — the-reviewer-model-and-the-review-class
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: REJECT on claude-opus-5-5 (Anthropic), a9aced6...2d373c2. B1 blocking: author_record follows symlinks out of the review target, and a file outside reaches the printed author statement (probe P4, settings supplied by the reviewed tree's own proven.yaml).
  ADR-0011 p1-7, ADR-0014 p1-8 and ADR-0013 p4a are held; highest-risk-wins unplug reddens the two tests the master named. Owed: O1 the tree's proven.yaml lowers its own class (owner), O2 defaults undocumented, O3 fallback/commits-off untested (mutant survives), O4/O5/O6.
  Unrun: listener tests (sandbox bind denied) and the NVIDIA half (route down). Master removes logs/trace.jsonl left by the ordered scan.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T14:10:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 3 |
