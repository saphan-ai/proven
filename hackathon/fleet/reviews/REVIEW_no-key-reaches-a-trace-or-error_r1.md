# REVIEW — no-key-reaches-a-trace-or-error · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T12:42:02Z

## Summary

claude-code / claude-opus-5-5. Trace holds: one redactor at trace.Writer.Append, world and H-MEM fed from it. Fails the wagon's own claim on two paths, each with a failing test: B1 finding title, message and evidence reach review.Result unredacted; B2 the planner duplicate-id error returns model text verbatim.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — no-key-reaches-a-trace-or-error

**RED — the order's artifact confinement and the review role's write-set disagree, and the review role cannot clean up.**
The order puts every artifact under `████████████████████████████████████████████████████` and asks for a clean
`git status --porcelain`. The permission class `██████████████████████████` admits only
`[_REVIEW_STATUS.md _REVIEW_NOTES.md review/]` for writes and refuses `rm`. A Bash redirect into `█████████████` was
refused (`the path "██████████████████████████" is inside the workdir but OUTSIDE the write-set of role review`). The
Write tool had already put 8 scratch files there without a refusal. I stopped writing there once the hook named the
write-set (law #57). `████████` is untracked and stays untracked; the master owes its removal (list under Unrun). The
same class of refusal stopped the earlier piece-2 reviewer's cleanup, so this is not a one-off.

Reviewer: claude-code, model `claude-opus-5-5`. Role-lock: `git rev-parse --short=7 HEAD` printed `131862e`. Range
`21af174...131862e`, two commits. Profile: security.

Verdict: **reject**. The trace now holds: every string field of every event goes through one redactor at
`trace.Writer.Append`. The H-MEM and world stores are fed from that redacted file, so they hold too. The wagon's own
claim is "no configured secret reaches the trace, **the result** or **an error**, even when a service echoes it". That
claim fails on two paths, and each has a test that fails at the tip:
**(B1)** a finding's title, message and evidence reach `review.Result` unredacted;
**(B2)** a fatal agent error returns a model-echoed value verbatim.

### Ground

- Toolchain: `go build ./... && go vet ./...` both exit 0 at `131862e` on this host (darwin).
- `go test -count=1 ./...` at `131862e`: 28 packages `ok`, 5 `[no test files]`, **6 `FAIL`**: `cmd/livecheck`,
  `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`. Every one of
  the six fails the same way: `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not
  permitted`. This host's confinement refuses a loopback listener. A rerun with the harness sandbox disabled gave the
  same panic. These six are UNMEASURED here, not red: the bind refusal makes a green and a broken package look the same.
- Master's claim (2026-10-23, macm5): "34 packages ok". 28 ok plus 6 bind-blocked makes 34 packages with tests, so the
  count matches. Their GREEN state is **not re-measured** here.
- Writers enumerated with `grep -rnE "os\.WriteFile|os\.Create|OpenFile|json\.NewEncoder|\.Write\(|Fprint"` over
  `internal cmd demo`. Every file writer that is not a test, and its exposure:
  - `internal/trace/trace.go:128,189` writes the trace. **Redacted** (`trace.go:172` → `trace/redact.go`).
  - `internal/world/world.go:548,568` writes `state.json` and `world.jsonl`. Fed by `pipeline.fold`
    (`internal/pipeline/memory.go:54-61`), which reads **the trace file from disk**, so it is redacted transitively.
  - `internal/hmem/hmem.go:536,555` writes `hmem.jsonl` and state. It is reached only through
    `world.apply → RememberRefutation` (`world.go:383`), so it is redacted transitively.
  - `internal/testrun/written.go`, `internal/change/archive.go` write checkouts. These carry repository content, not
    service text.
  - `cmd/livecheck/main.go:80,93,100` print terminal lines. **Redacted** (`rd.Text`), including the unknown argument.
  - `demo/live/review/main.go:34,52,57` print terminal lines. **Redacted** (`redactText`).
  - `cmd/proven/main.go` runs sample mode only and loads no secret (as read by the earlier review; not changed here).
- Every use of the redactor outside its own package (`grep -rn "redactor\|Redactor\|redact\."`):
  `pipeline.go:147` (trace), `:941` (failure reason), `:1062` (models kept on the draft), `review/open.go:68` (built
  from env, live only), `cmd/livecheck/main.go:78`, `demo/live/review/main.go:148`. **Nothing touches
  `draft.Findings` or the error `Review` returns.**
- Production builders: `pipeline.Build` has one caller, `internal/review/open.go:150`. `trace.Open` has one caller,
  `pipeline.go:138`.

### Differential

The original reviewers' tests (from `████████████████████████`), run against `131862e`. The Claude reviewer's tests
are already in the tree; the Codex reviewer's tests were added by `go test -overlay` and change no tracked file.

| test (origin) | result at `131862e` |
|---|---|
| `TestRuntimeTokenEchoedInTheReceiptIsNotWritten` (p2-claude, now `internal/review/secret_echo_test.go`) | UNMEASURED — httptest bind refused |
| `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten` (p2-claude, review package) | UNMEASURED — bind refused |
| `TestARefusedSecretIsNamedNotPrinted` (p2-claude, review package) | **PASS** |
| `TestARefusedSecretIsNamedNotPrinted` (p2-claude, livecheck version) | not in the tree; its review-package twin passes. The livecheck refusal path is covered by `TestRefusedOrgIDIsNotPrinted` **PASS** |
| `TestReviewSecretNeverPrintedOnUnknownArgument` (p2-codex) | **PASS** — closed by `cmd/livecheck/main.go:78-80` |
| `TestReviewOrgIDNeverPrintedOnConfigurationError` (p2-codex) | **PASS** — closed by `internal/query/settings.go:128,362` |
| `TestReviewOrgIDNeverPrintedWhenServiceEchoesIt` (p2-codex) | UNMEASURED — bind refused |
| `TestReviewConfiguredSecretNeverReachesTraceFromReceipt` (p2-codex) | **FAIL** `the configured secret reached the trace`. This is expected from the design: the test builds `recording` and `trace.Open` by hand with no redactor, and the redactor is injected only by `review.Open`. It does not show a production leak (one builder, see Ground). It does show the silent nil default in O1. |
| `TestReviewSecretNeverReachesRuntimeProcess` (p2-codex) | **FAIL** `the configured secret reached the runtime process`. Not addressed by this wagon (Remainder R1). |

The leg's own tests that need no network: `TestConfiguredValueEchoedInAReceiptIsNotWritten` (4 subtests) **PASS**,
pipeline `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten` **PASS**, trace `TestAppendRedacts…` and
`TestAWriterWithoutARedactor…` **PASS**, `internal/redact` (7 tests) **PASS**, `TestLivePipelineWithoutAGuardrailIsRefused`
**PASS**.

New tests by this reviewer, network-free, at pipeline level (overlay, not committed). They run the leg's own
seam: `Build(Options{Sample: true, Redactor: redact.FromEnv(...)})` with a generator wrapper that echoes the token in
the model's answer text.

- `TestW30TokenInAFindingDoesNotReachTheDraft` → **FAIL**:
  `the draft's finding carries the token: "Discount is applied twice marker-w30-find-91ab"`. In the same run the trace
  was clean (`trace carries marker: false; trace carries [redacted: true`). So the trace holds and the result does not.
- `TestW30TokenInAFatalAgentErrorDoesNotReachTheError` → **FAIL**:
  `review error: planner: duplicate id marker-w30-dup-44cd`.

Master-was-wrong (law #44): **partly**. The 34-package count is consistent. The negative control "trace writer
bypassed → the receipt and categories tests red" could not be repeated here (Unplug). The leg's account says the
redactor "stands at the trace writer, the recorded models, a failure's reason". That is accurate, but it is narrower
than the wagon's task ("the result"), and the master's order carried the task.

### Unplug

- U1, trace writer bypassed (`trace/redact.go:10` forced to return `e`): **not run**. A mutant file needs a write
  location, and the only admitted one is the write-set above. Read from the code instead: a mutant there makes
  `TestAppendRedactsEveryTextFieldAndLeavesTheCallersEventAlone` fail, because it asserts on the written bytes. The
  leg's pipeline tests also read the trace file. This is UNMEASURED; the reading is not a measurement.
- U2, `review/open.go:68` removed (no redactor in live mode): **not run**, for the same reason. Read from the code:
  only the httptest-based review tests (`secret_echo_test.go`) build through `review.Open`. The pipeline tests pass
  `Redactor` explicitly and would stay green. So on a host that refuses a listener, nothing turns red when this guard
  is unplugged. See O1.
- The two W30 tests above are the unplug evidence in the other direction: they show the paths that have no guard.

### Closure

**B1 (BLOCKING) — a finding's text reaches `review.Result` unredacted.**
- Trigger: the generator echoes a configured value in a risk's title or message, or Zetaris rows or Prover output
  carry one.
- Path: planner output → `candidate` → `draft.Findings` → `internal/review/open.go:241-255` (`fill` copies `Title`,
  `Message`, and `toEvidence` copies `Test.Output`, `Query.Rows` and the rest) → `Result.Findings`.
- No redactor is applied on that path (Ground, the redactor-use list). The demo redacts its own stdout, but
  `Result` is the API that a report, SARIF or pull-request-comment writer will read. `SECURITY.md:23` promises
  "never written to … the reports".
- Proof: `TestW30TokenInAFindingDoesNotReachTheDraft` fails at `131862e`.
- Remedy: apply the one redactor to every string of each finding and its evidence where the draft is assembled, or
  in `fill`. Then assert on `Result.Findings` in the leg's review-level echo test and in a network-free
  pipeline-level twin.

**B2 (BLOCKING for the wagon's "or an error" claim; narrow trigger) — a fatal agent error carries model text.**
- Path: `internal/agent/planner/planner.go:88` `fmt.Errorf("planner: duplicate id %s", r.ID)` →
  `pipeline.soften` returns it as fatal (`pipeline.go:926`) → `review/open.go:229` returns it verbatim.
- This is the only agent error that formats model text (`grep -rnE "errors.New\(|Errorf\(" internal/agent`).
- Proof: `TestW30TokenInAFatalAgentErrorDoesNotReachTheError` fails.
- Remedy: redact the error at the review seam (`runner.Review`), or drop the id from the message.

Both tests, full text: `██████████████████████████████████████████████████████████████████████████████████`, until the
master removes it. Short form: a `generator.Generator` wrapper around `generator.NewSample(fx.Answers)`.
- It rewrites the planner's answer, either appending the marker to `"Discount is applied twice"` or replacing the
  risk list with two risks sharing `id=marker`.
- It strips the marker from incoming prompts so that the recorded answers still match.
- It builds with `Redactor: redact.FromEnv(token=marker)`.
- It asserts that no `d.Findings[i].Title+Message` and no `err.Error()` contains the marker.

### Remainder

- **O1 (OWED) — a live pipeline with a nil redactor redacts nothing, silently.**
  - Where: `pipeline.go:92-95,146-147`, "Nil redacts nothing". This is a silent fallback that hides the invariant
    (R8).
  - Paired path (R13): `Build` already refuses a live pipeline with no guardrail
    (`TestLivePipelineWithoutAGuardrailIsRefused`). The redactor has no matching refusal.
  - Remedy: refuse `Sample: false` with a nil `Redactor`, with a test that runs without a network.
- **O2 (OWED) — comments cite the wrong decision (ADR-0027 point 4).**
  - The comments cite ADR-0025 at `pipeline.go` (package doc and `:142`), `trace/trace.go:172`, `review/open.go:66`
    and the `internal/redact` package comment. ADR-0025 is "the Prover runs the change … in a container with no
    network and no secret"; it says nothing about the trace or the result.
  - The decision the code actually carries out is `SECURITY.md:23` together with ADR-0013 point 4 ("secrets
    masked").
  - Remedy: re-point the citations. In all other respects the comments meet ADR-0027: every new exported name has a
    doc comment, every new test states its property with a `Mutated field:` line, and there is no AI attribution.
- **N1 (NOTE) — the redaction rule, short and empty values.**
  - An empty value, or one under `MinLength = 6` bytes after trimming, is never redacted (`redact.go:402,430`).
    Documented in `demo/live/README.md`; tested in `TestShortAndEmptyValuesAreNotRedacted`.
  - What holds: the loaders' refusals never print such a value (`TestARefusedSecretIsNamedNotPrinted`,
    `TestARefusedOrgIDIsNamedNotPrinted` PASS).
  - What does not: a 1–5 digit `ZETARIS_ORG_ID` that a service echoes reaches every writer.
  - A false positive in the other direction: a 6–18 digit org id is replaced wherever those digits appear, inside
    query rows, line counts or test output in the trace. The trace's evidence then differs from the result's
    evidence, and the event id is computed over the replaced text (`trace.go:172`).
- **N2 (NOTE) — transformed echoes are not caught.** The match is the exact bytes (`strings.ReplaceAll`). Each of
  these passes through: a change of case; percent, base64 or `%q` encoding (the escaped line break the earlier N1
  named); a prefix or a truncation; a value split across two fields or two stream chunks; and an echo of
  a value with internal whitespace changed. What holds: a JSON `\u` escape in a service's wire answer, because it is
  decoded before the redactor runs, and the trace's own JSON encoding, which happens after it. All of this is read
  from the code; none of it was run.
- **R1 (carried, out of this wagon's stated surfaces) — the runtime subprocess inherits the full environment.**
  `internal/testrun/container.go:432`; `TestReviewSecretNeverReachesRuntimeProcess` still FAILS. The container itself
  gets no `-e` (ADR-0025 point 3 holds). The host `docker` and `git` processes hold every key. This is piece-2 scope;
  it is not named in W30's task.
- Closed by this wagon (from the earlier reviews): livecheck's unknown-argument echo, the refused org id print, the
  org id added to every redaction set, the untested duplicate redactor in the demo (now `internal/redact`, tested),
  the receipt and categories paths into the trace, `Models` and `FailedServices`.

### Profile

Security.
- **S1 boundary map:**
  - environment → loaders: refusals blank the key and the org id.
  - loaders → `redact.FromEnv` in `review.Open`.
  - service text → trace (redacted) → world and H-MEM (redacted transitively).
  - service text → `draft.Models` and `draft.Failed` (redacted).
  - service text → `draft.Findings` → `Result.Findings`: **disclosure, B1**.
  - model text → a fatal error: **disclosure, B2**.
  - process → `docker` and `git`: the full environment crosses (R1).
- **S2:** B1 and B2 are each reproduced by a failing test.
- **S3:** the guards could not be removed in place (Unplug). Their test coverage was read instead; U2 has no
  network-free test (O1).
- **S4:** no dependency delta (`go.mod` and `go.sum` untouched in the range). Secrets searched (Ground). No destructive
  path. Grepping the commits, comments and test names for finding text (`F[0-9]`, `N[0-9]`, `finding`, `reviewer`,
  `security-piece`) found none. The scan for invisible or bidirectional characters was not run (Unrun).
- Testing bar: no metric is declared, so none is claimed as met. Every check here is a suite check.

### Focus

Does any key reach the trace, the result or an error at `131862e`?
- **The trace: no**, for exact-byte echoes of 6 or more bytes. One writer, every field, tested.
- **The result: yes.** `Models` and `FailedServices` are redacted; `Findings` and their evidence are not (B1).
- **An error: yes**, on one narrow path (B2).
- Refused settings, including the org id: named, never printed.

### Conditions

- **C1 (BLOCKING, B1):** redact every finding and evidence string before it becomes `Result`. Proof: a network-free
  test asserting on `Result.Findings` (or `draft.Findings`) fails before the fix and passes after.
- **C2 (BLOCKING, B2):** no model-supplied text in the error `Review` returns. Proof:
  `TestW30TokenInAFatalAgentErrorDoesNotReachTheError` or an equivalent turns green.
- **C3 (OWED, O1):** `Build` refuses a live pipeline with no redactor, with a test that needs no network.
- **C4 (OWED, O2):** re-point the ADR-0025 citations to the decision the code actually carries out.
- **C5 (OWED, for the master):** re-run on a host that permits a loopback listener the six packages that are
  UNMEASURED here, and repeat the leg's negative control.
- N1, N2 and R1 are notes. R1 is returned to the piece-2 train.
- No class was changed by this reviewer.

### Unrun

- All httptest-based tests, including the leg's `internal/review/secret_echo_test.go` and livecheck's echo tests, are
  refused with `listen tcp6 [::1]:0: bind: operation not permitted`. A retry with the harness sandbox disabled gave
  the same refusal.
- The negative controls U1 and U2 need a mutant file, and the write-set admits none outside
  `_REVIEW_*`/`review/`:
  `the redirection … writes a file … OUTSIDE the write-set of role review under permission class
  ██████████████████████████`.
- The invisible and bidirectional character scan: `grep -P` is unsupported by this grep, and the `LC_ALL=C` prefix is
  refused (`"LC_ALL=C" is NOT listed in permission class ██████████████████████████`).
- The race detector was not run. No live service was called. No real container was started.
- `git check-ignore` asked for approval in a headless run, so it was not run.
- **Cleanup owed by the master** (`rm` is refused for this role). Untracked under
  `████████████████████████████████████████████████████`: `overlay.json`, `overlay2.json`,
  `livecheck/security_review_test.go`, `pipeline/security_review_test.go`, `pipeline/zz_w30_review_test.go`,
  `testrun/security_review_test.go`, `review/zz_w30_review_test.go`. No tracked file was changed (`git status
  --porcelain` shows `████████` as the only addition of mine).
- The second, independent NVIDIA review through Nightshift is owed (the route is down today, per the order).

## The review status as filed

## REVIEW — no-key-reaches-a-trace-or-error
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: claude-code / claude-opus-5-5. Trace holds: one redactor at trace.Writer.Append, world and H-MEM fed from it. Fails the wagon's own claim on two paths, each with a failing test: B1 finding title, message and evidence reach review.Result unredacted; B2 the planner duplicate-id error returns model text verbatim.
  Owed: a live Build with a nil redactor should be refused; comments cite ADR-0025 (the container) for trace redaction.
  build+vet ok; 6 httptest packages UNMEASURED (loopback bind refused on this host); untracked ████████ left for the master (rm refused).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T13:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 8 |
| policy binding name | 3 |
| private workspace path | 1 |
