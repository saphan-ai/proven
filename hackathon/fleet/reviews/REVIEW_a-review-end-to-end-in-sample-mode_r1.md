# REVIEW — a-review-end-to-end-in-sample-mode · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T03:31:43Z

## Summary

Reject: partial results can include F3 without evidence, and the trace omits proof/refutation details. Codex reviewer; exact model and cost unexposed. Full Go package tests passed; direct source inversion unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — a-review-end-to-end-in-sample-mode

### Ground
Reviewed `3a1a9f0...041b3f2` at `041b3f2` on 2026-10-22. Scope: the two commits and 46 changed files. Ran `go test ./internal/... ./cmd/proven/...` with local caches: all packages passed. Built the CLI and ran `scan --sample` against the discount fixture from the review scratch directory: F1 and F2 appeared; F3 did not; the trace recorded F3's drop. `git diff --check 3a1a9f0...041b3f2` passed. No dependency manifest changed; `go.mod` pins Go 1.26 and no external module. Read the paired sample/live adapters and the public `internal/review` contract. The project has no declared numeric coverage, mutation or lint baseline in this diff; there is no invented metric or exception. Check ranking: repository tests are suite-owned; `git diff --check` is a project check; no external checker was run.
remark: The successful sample scan says nothing about deadlines or later live adapters; its false-world answer for those cases is identical.

### Differential
The master's F1/F2/F3, no-network-import, and ordinary append-path claims held in the measured commands. The master's stronger claim that the result contains only evidenced findings does not hold on the partial path. The trace exists but its evidence and refutation events omit essential content. The note that the network guard is armed holds, but no agent uses its guarded HTTP client (`internal/pipeline/pipeline.go:42-58`, `████████████████████████████████`). These distinctions are observations, not assumptions from the master's status.

### Unplug
The existing negative controls remove only `recorded_tests` for F1 (`internal/pipeline/pipeline_test.go:93-138`) or add a row to `tables.sql` for F3 (`internal/pipeline/pipeline_test.go:54-91`); both passed in the suite. `TestDropsWhenThereIsNoEvidence` (`internal/agent/judge/judge_test.go:10-15`) would reject changing `judge.Decide` to keep empty evidence. The source-level inversion itself was UNRUN: the available executor has no `apply_patch` command/tool, and review rules prohibit substituting a different write mechanism for source edits. The no-network test `TestSampleModeOpensNoNetwork` (`internal/pipeline/pipeline_test.go:15-26`) only observes an armed, unused client and zero attempted dials; removing the dial refusal while leaving `Arm` intact would not turn it red. `TestGuardDialIsVisible` (`█████████████████████████████████████`) directly exercises the transport and would turn red if its dial stopped refusing. The import census (`internal/pipeline/imports_test.go:12-70`) refuses `net`, `net/http`, and `os/exec` imports across the listed sample packages; it explicitly exempts `internal/netguard`. A sample test-run swap to the live adapter (`internal/pipeline/pipeline_test.go:28-52`) reports the Prover as off rather than executing change code.
remark: Zero observed dials does not prove network blocking: an unused guard reports zero whether its transport refuses or permits a dial.

### Closure
**BLOCKING — Partial reviews publish unevidenced findings.** `internal/pipeline/pipeline.go:79-88` builds candidates before judging them; deadline returns at `:91-93`, `:112-114`, `:135-137` and `:157-159` call `state.partial`, which calls `writer.Write` at `:245-249`. `internal/agent/writer/writer.go:18-24` accepts every candidate whose `Dropped` default remains false; only `internal/agent/judge/judge.go:8-17` enforces evidence, and it has not run. A scratch Go reproduction passing an unjudged F3 candidate into the actual writer printed `partial=true findings=1 first=F3 evidence=0`. A deadline after claims can therefore return F3 with an empty evidence list, violating `internal/review/review.go:76-94`. Remedy: apply the evidence predicate on every partial-return path (including mid-agent cancellation), trace a verdict for each returned candidate, and add a cancellation-after-planning test that asserts every emitted finding has qualifying evidence.

**BLOCKING — The append-only trace omits the proof and refutation it claims to preserve.** `internal/pipeline/pipeline.go:266-285` reduces F1 test evidence to path/name and F2 query evidence to SQL plus a row count, discarding the recorded test outcomes/output, query columns and rows. `:140-150` records only `Refutation.Reason`, discarding `Attempted` and `Refuted`. The sample scan's JSONL contains a test name and `rows=1`, not the actual result row or outcomes. `internal/trace/trace.go:27-56` does use O_APPEND and propagates Write/Sync errors; that does not recover missing event data. Remedy: persist structured evidence and refutation values with their finding id, and assert JSONL round-trip for the F1 outcome and F2 row. Never claim full replay from the current trace.

**OWED — The review request loses its selection fields.** `internal/review/open.go:35-36` forwards only `Target`, while `internal/review/review.go:60-72` also defines baseline/head, include/exclude and minimum severity. `internal/pipeline/pipeline.go:60-63` always loads a fixture directory. A caller supplying a narrower path set or commit range gets the same result as an unrestricted request, without a refusal. Remedy: implement the selection semantics at the seam or explicitly reject nondefault fields before reviewing, and test the refusal/selection. The CLI currently accepts only `--sample`, which limits present exposure but not the published `Reviewer` interface.

**OWED — Sample guard is not a process-wide egress barrier.** `internal/pipeline/pipeline.go:49-55` arms `netguard.Guard` but never passes its client into an agent or adapter; the static import test is the operative check in the current source. A later live adapter would require its own explicit no-egress boundary, not an unused client. Remedy: route any network-capable dependency through the guarded transport and add a sample-run negative control that makes an attempted dial fail visibly. The present sample generator/query/test-run implementations contain no network/exec calls in their normal paths.

### Remainder
`internal/change/read.go:40-64` reads leaf symlinks by following them, with no containment check; fixture content is local in sample mode, but a future live generator might receive out-of-root bytes. Remedy before enabling live generation: resolve links within an allowed target root or refuse them and test a symlink outside the fixture. The current CLI invokes sample mode only (`cmd/proven/main.go:41-64`); this is a forward-looking boundary finding, not evidence of present exfiltration. The sample SQL engine parses SELECT without a documented resource bound; no performance claim was made. No toolchain denial occurred. No bars or tests were weakened in the diff. Declined to judge: licenses (owner scope), production provider networking (not implemented), throughput/performance (no baseline).

### Profile
Backend plus security, because the change adds a JSONL persistence path, local change-file reads and future query/model adapters. The only present writer is `trace.Open` with `O_APPEND`, and it checks errors; it does not close its file descriptor or distinguish runs sharing `logs/trace.jsonl`. The profile's highest risk is an unauditable partial result, not a measured external network call. No credentials, policies or private implementation details are reproduced here.

### Focus
1. Judge: `internal/agent/judge/judge.go:8-17` drops no-evidence candidates on the completed path; `internal/agent/writer/writer.go:18-24` trusts default `Dropped=false` on partial paths.
2. Sample boundary: `internal/testrun/sample.go:31-54` returns recorded outcomes and never executes a command; `internal/query/sample.go:16-43` loads local SQL; the network guard is unused on this path.
3. Trace: `internal/trace/trace.go:27-56` appends and syncs, and `internal/pipeline/pipeline.go:266-285` strips evidence details before append; write errors propagate rather than silently producing success.
4. Seam: `internal/review/open.go:35-71` maps failures and partial status but never forwards selection fields. `internal/pipeline/pipeline.go:257-263` labels recorded model calls `unknown` and `FromReceipt=false`; this is explicit lack of receipt, not a guessed provider identity.
5. Later live boundary: fixture leaf symlinks in `internal/change/read.go:40-64` and a guard not wired into adapters require containment before live mode.

### Conditions
Verdict: reject. BLOCKING findings require corrected code and regression tests for deadline-after-planning and full evidence/refutation trace payload; the master should then request a fresh independent review. OWED findings require a stated seam decision and a wired egress control before exposing non-sample review. No review condition is downgraded here.

### Unrun
Direct source mutation of Judge and the no-network transport (missing `apply_patch` tool), a timed cancellation between planning and judging, full live mode (no provider), network observation beyond the package import census, and base-vs-tip numeric testing-bar metrics (none declared) were not run. The reviewer model's exact identifier, monthly-pool before/after and charged cost were not exposed to this leg; do not invent them. No bridge registration or public `fleet/reviews` copy was made: the mandated review return is confined to the two channel files and the owner must reconcile the separate publication rule. No claim in this review depends on those unrun checks passing.

## The review status as filed

## REVIEW — a-review-end-to-end-in-sample-mode
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject: partial results can include F3 without evidence, and the trace omits proof/refutation details. Codex reviewer; exact model and cost unexposed. Full Go package tests passed; direct source inversion unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T03:29:36Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 2 |
