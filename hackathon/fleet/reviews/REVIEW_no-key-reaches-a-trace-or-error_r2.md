# REVIEW — no-key-reaches-a-trace-or-error · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T14:37:17Z

## Summary

Re-review of 131862e..3cfc7f8 on claude-opus-5-5: C1, C2, C3, C4 all ADDRESSED; each fix unplugged turns a named test red.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — no-key-reaches-a-trace-or-error

Re-review, round 2 (R14). Scope: `131862e..3cfc7f8`, one commit, `[no-key-reaches-a-trace-or-error] pipeline: redact
findings`. Reviewer model: claude-opus-5-5. Profile: security (R9 — the diff is on the secret boundary).
All measurements below were taken by this leg on 2026-10-23 at `3cfc7f8`.

### Ground

- `git rev-parse --short=7 HEAD` → `3cfc7f8` (ROLE-LOCK holds).
- Diff read in full: `internal/pipeline/pipeline.go` (+redaction at the end of `Pipeline.Review`, `redactFindings`,
  `redactRows`, a `Build` refusal), `internal/agent/planner/planner.go` (two error strings), comment-only changes in
  `internal/redact/redact.go`, `internal/review/open.go`, `internal/trace/trace.go`, one `Redactor:` line in
  `internal/pipeline/pipeline_test.go`, and a new `internal/pipeline/result_redaction_test.go` (4 tests).
- Callers read: `Pipeline.Review` (`pipeline.go:241-264`) and every return of `p.review`. Every error return builds
  `candidate.Draft{}` (grep of `return candidate.Draft{}` in `pipeline.go`). The partial drafts (`pipeline.go:363`,
  `:1032`, deadline path) return a nil error, so they pass through the redaction at `pipeline.go:262`. No draft with
  findings leaves `Review` without being redacted.
- Every field `redactFindings` touches checked against `candidate.go:22-104` (`Risk`, `TestEvidence`, `QueryEvidence`,
  `Evidence`, `Refutation`, `Candidate`): every string and `[]string`/`[][]string` field is covered. My reflection
  test below confirms this mechanically.
- Live `Build` caller: `internal/review/open.go:57-68`. Its live branch always sets `popts.Redactor = redact.FromEnv(env)`
  before `pipeline.Build` (`open.go:150`). The other `Build` caller, `internal/review/doubled_test.go:151`, is
  `Sample: true`.
- Planner error sites (`grep Errorf planner.go`): three. None formats model text after the change.
- **Round-1 notes NOT read.** The order points at `██████████████████████████████████████████████` and
  `REVIEW_NOTES_p2-codex.md`. Neither was provisioned (`ls ███████████████████████` → only `REVIEW_NOTES_p2-claude.md`,
  which is the piece-2 review of `90d6951` with different C1–C4). I judged C1–C4 as the order pastes them.
- Round-1 tests: the named functions `TestW30TokenInAFindingDoesNotReachTheDraft` /
  `TestW30TokenInAFatalAgentErrorDoesNotReachTheError` were not provisioned either. The leg landed tests of the same
  description without the `W30` prefix (`result_redaction_test.go:57`, `:80`). I wrote my own independent versions
  under the round-1 names, plus one completeness test, and injected them with `go test -overlay`, so the worktree was
  never written. Files: `artifacts/no-key-reaches-a-trace-or-error-review-r2/run-20261023-141913-e314feea/scratch/`
  (`w30_review_r2_test.go`, `overlay.json`), outside the worktree.
  - `TestW30TokenInAFindingDoesNotReachTheDraft`: a sample review whose planner answer carries the marker in the title,
    message and rule_id; it marshals the WHOLE draft to JSON, asserts the marker is absent and a `[redacted: ` marker
    is present (positive control), and asserts at least one finding.
  - `TestW30TokenInAFatalAgentErrorDoesNotReachTheError`: the planner returns two risks whose id is the marker; it
    asserts an error, no marker in it, the refusal named, and no findings.
  - `TestW30RedactFindingsCoversEveryStringField`: reflection sets every exported string reachable from a `Candidate`
    to the marker, runs `redactFindings`, asserts none survives and the input is unchanged.
- At tip, all three are green, and so are the leg's four (`go test -run 'TestW30|TestTokenIn|TestRedactFindings|TestLivePipelineWithout'`).

### Differential

- Master's claim (2026-10-23): "34 packages ok at 3cfc7f8". **Does not hold in this sandbox; it is UNMEASURED, not
  refuted.** `go build ./...` and `go vet ./...` pass. `go test -count=1 ./...` gives 34 test packages: **28 ok,
  6 FAIL**. Every FAIL is the same panic, `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation
  not permitted`, raised by the sandbox before any assertion: `cmd/livecheck`, `internal/generator`,
  `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`. The panic kills the package binary, so
  every test after it in those packages also went unrun.
- Narrowing for the package the diff touches: `go test -count=1 -skip '<the 8 tests of guardrail_test.go>'
  ./internal/pipeline/` → `ok` (32 PASS lines). The diff touches no test in `guardrail_test.go`.
- remark: a green `internal/pipeline` here says nothing about the 8 skipped guardrail tests or the five other panicking packages; it does not stand in for the master's 34.

### Unplug

Each mutation was made in place with one edit, the tests were run, the edit was reverted, and then
`git diff --stat HEAD` printed nothing.

| # | Mutation | Red | Survived |
|---|---|---|---|
| U1 (C1) | `pipeline.go:262` `d.Findings = p.redactFindings(d.Findings)` commented out | `TestTokenInAFindingDoesNotReachTheDraft`, `TestW30TokenInAFindingDoesNotReachTheDraft` | — |
| U2 (C1, one field) | `pipeline.go:278` `c.Message` left unredacted | `TestW30TokenInAFindingDoesNotReachTheDraft`, `TestW30RedactFindingsCoversEveryStringField` | **`TestTokenInAFindingDoesNotReachTheDraft`, `TestRedactFindingsReachesEvidence`** (O1) |
| U3 (C2) | `planner.go:90` error re-formats `r.ID` | `TestTokenInAFatalAgentErrorDoesNotReachTheError`, `TestW30TokenInAFatalAgentErrorDoesNotReachTheError` | every test in `internal/agent/planner` (guarded only from the pipeline) |
| U4 (C3) | `pipeline.go:142` guard made `false && …` | `TestLivePipelineWithoutARedactorIsRefused` (run alone: `a live pipeline with no redactor was accepted`) | — |

In the package-wide U4 run the `httptest` panic stopped `internal/pipeline` before the C3 test, so I ran it alone.

### Closure

- **C1 — ADDRESSED.** `pipeline.go:256-263`: after every agent has run, every finding is passed through
  `redactFindings` (`:270-308`). It returns a deep copy (new evidence slices, copied `Test`/`Query` structs, copied rows)
  with the one redactor applied to every string field. Proof: U1 red on both end-to-end tests; the whole-draft JSON scan
  is green at tip.
- **C2 — ADDRESSED.** `planner.go:86-91`: both id errors name a 1-based position (`risk %d has no id`,
  `risk %d repeats an earlier id`) and format no model text. Proof: U3 red on both pipeline tests.
- **C3 — ADDRESSED.** `pipeline.go:139-144`: `!opts.Sample && opts.Redactor == nil` is refused before `trace.Open`.
  The test flips exactly `Options.Redactor` (negative control, law #27). The live caller always supplies one
  (`open.go:68`). Proof: U4.
- **C4 — ADDRESSED.** The secret-redaction comments that cited ADR-0025 (the container decision) now cite ADR-0005 and
  ADR-0013 point 4 (`pipeline.go:15`, `:141`, `:149`, `:260`, `redact.go:4`, `open.go:66`, `trace.go:170`). The ADR-0025
  citations that remain (`grep -rn ADR-0025 --include='*.go'`) are all about the container, where they belong.

### Remainder

Owed, R14 (outside the round-1 conditions; none opens another round):
- **O1 (owed, test strength)** — `result_redaction_test.go:57-74` puts the token only in the title, and
  `:98-117` checks five chosen fields. U2 shows that a Message leak survives both. Remedy: land a completeness test such
  as `TestW30RedactFindingsCoversEveryStringField` (reflection over every string field), so a field added to `Candidate`
  later without redaction turns red.
- **O2 (owed, test locality)** — U3 survives the planner package's own tests. Remedy: one test in
  `internal/agent/planner` that a repeated or empty id yields an error with no model text.
- **N1 (note)** — ADR-0005 holds no sentence about keys in output (grep of `fleet/decisions` for redaction language hits
  only ADR-0023). ADR-0013 point 4, "secrets masked", is the nearest written decision. The property itself lives in
  `SECURITY_REVIEW.md` piece 2, not in an ADR. The owner may want an ADR point that states it.

Declined to judge (R11): the trace and `Result.FailedServices` paths (round 1 of piece 2, commit `131862e`, outside
this diff), and whether the review layer adds unredacted text to `review.Result` after `Pipeline.Review`.

### Profile

security. S1: the boundary in scope is service/model text → `candidate.Draft` → `review.Result`, plus model text →
error. Disclosure across it is closed for findings (C1) and the planner's id errors (C2). S2: both blocking findings
are reproduced by a failing test under unplug (U1, U3). S3: each new guard was removed once and a test fails (U1–U4).
S4: no dependency delta (`go.mod`/`go.sum` untouched in the diff); no destructive path. A scan for invisible or
bidirectional characters was not run (Unrun).

### Focus

THE PROPERTY: no key reaches the trace, the logs, the reports, a PR comment, an error message or a container.
Within this diff it holds for the draft's findings and the planner's errors. On every return path of `Pipeline.Review`,
a finding either passes the redactor or the draft is empty.

- master-was-wrong: **unmeasured** — "34 packages ok at 3cfc7f8": 34 test packages exist, but 6 cannot run in this
  sandbox (loopback bind refused); 28 ok here.
- master-was-wrong: **partly** — the order names `REVIEW_NOTES_round1.md` and `REVIEW_NOTES_p2-codex.md` as
  provisioned; neither is in `████████████████████████`.

### Conditions

- **K1 (verification, master)** — run `go test -count=1 ./...` at `3cfc7f8` on a host that allows a loopback listener
  and paste the result. The six packages above must be `ok`; this leg could not measure them (see Unrun). If any is red
  for a reason other than the sandbox, this verdict does not stand.
- O1 and O2 are owed, not blocking (R6).
- The NVIDIA half of the owner's two-review rule (2026-10-22) is still owed; this review ran alone.

### Unrun

- `go test ./...` for `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/query`, `internal/review`
  in full, and the 8 `guardrail_test.go` tests of `internal/pipeline`: the sandbox refuses the bind (`listen tcp6
  [::1]:0: bind: operation not permitted`). This includes the piece-2 echo tests (`TestRuntimeTokenEchoedInTheReceiptIsNotWritten`,
  `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten`).
- The scratch copy the order asks for: `git worktree add --detach ██████████████████████████████████████████████████████████ 3cfc7f8`
  → `fatal: could not create directory of '…/sources/proven/.git/█████████████': Operation not permitted`. Test files
  were injected with `go test -overlay` from outside the worktree instead, and the mutations were made in place and
  reverted (`git diff --stat HEAD` empty).
- The round-1 review notes and the round-1 test bodies were not provisioned (Ground).
- No scan for invisible or bidirectional characters; no race detector; no live service.
- The scratch files in the run's artifacts directory (`scratch/`) cannot be deleted by this role (`rm` is not in the
  review class); they sit outside the worktree and outside git.

## The review status as filed

## REVIEW — no-key-reaches-a-trace-or-error
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Re-review of 131862e..3cfc7f8 on claude-opus-5-5: C1, C2, C3, C4 all ADDRESSED; each fix unplugged turns a named test red.
  Build and vet green; go test: 28 of 34 packages ok, and 6 not measurable here (sandbox refuses loopback bind).
  K1: the master runs the full suite on a host that allows a listener. Owed: O1 a field-completeness test, O2 a planner-local test.
  The round-1 notes were not provisioned; C1–C4 were judged from the order's paste. The NVIDIA review half is still owed.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T15:10:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| private workspace path | 3 |
