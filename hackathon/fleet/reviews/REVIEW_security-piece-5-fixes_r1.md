# REVIEW — security-piece-5-fixes · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T06:33:32Z

## Summary

W45 at a37939e (diff main...a37939e; the order's base 00d4ab9 is a tree, not a commit): properties 1, 3, 4, 5 hold, each unplug reddens a test (14 of 15 mutants).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — security-piece-5-fixes

Reviewer: named REVIEW leg `security-piece-5-fixes-review`, claude-code, model `claude-opus-5-5`, 2026-10-24.
Profile: security (S1–S4). Verdict: **accept-with-conditions** (two conditions, neither blocking a gate on its own
in my reading; the master classifies). PRIVATE until merged.
The NVIDIA half of the owner's two-review rule (2026-10-22) is **owed**: the route was down, and this review ran alone.

### Ground

- `git rev-parse --short=7 HEAD` → `a37939e` (role-lock satisfied).
- **master-was-wrong (law #44):** the order's base `00d4ab9` is not a commit. `git merge-base 00d4ab9 a37939e` →
  `error: object 00d4ab93571e81fcb4bed7ba88f69b706235eb36 is a tree, not a commit`; `git cat-file -t 00d4ab9` → `tree`.
  The diff I reviewed is `main...a37939e`. `git merge-base refs/heads/main a37939e` → `d510ac5`, which is `a37939e^2`
  (the main absorbed by the master's merge). So the reviewed diff is exactly the five leg commits:
  `8bdf2c5` result, `f45c212` pipeline, `e006156` memory, `a211fe9` agent, `929bffd` judge. 37 files, +744/−118.
  remark: `main` has moved since the absorb (`refs/heads/main` → `6547849`); this verdict is on `a37939e`, not on a landing against `6547849`.
- Toolchain: `go build ./... && go vet ./...` clean. `go test -count=1 ./...`: every package `ok` except six that
  panic at their first `httptest` listener (`listen tcp6 [::1]:0: bind: operation not permitted`):
  `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`.
  Re-run of those six with the listener-file tests skipped (`-skip` of every `Test*` in the nine `httptest` test files):
  `pipeline ok 24.3s`, `review ok 61.0s`, `query ok`, `generator ok`, `guardrail ok`; `cmd/livecheck` still FAILs on
  `TestOrgIDEchoedByAServiceIsNotPrinted` (same listener panic). See Unrun.
- Instruments that write nothing to the worktree: every probe and mutant ran via `go test -overlay <json>`; sources,
  mutants and overlays live in this run's artifact directory (`tools/mut`, `tools/runmut`, `mutants/`, `probes/`).
  `git status --porcelain` at the end shows only the four untracked entries present at spawn.
- Prompt builders census: `grep -rn -E 'Instruct\(|Repair\(|Prompt\(|\.Generate\(|Complete\('` over `internal cmd`
  (non-test). Builders: Planner (`planner.go:29`), Investigator (`investigator.go:37`), Prover (`prover.go:39`),
  Skeptic (`skeptic.go:35`), `answer.Repair` (`answer.go:76`). The only callers of `Complete` are `answer.go:97,105`,
  the guard `guardrail/wrap.go:37`, the recorder `pipeline.go:1222`, and `cmd/livecheck/main.go:130` (fixed `"hello"`).
  There is **no Fixer agent in this tree** (`internal/agent/writer` is deterministic: no generator call). The generator's
  own framing is `live.go:240`: one `user` message whose content is the prompt, JSON-marshalled into the HTTP body.

### Differential

Property by property, FILE·LINE·CONTROL:

1. **Query rows count only when bound to the change — HOLDS on the support side.**
   `judge.go:73` counts a query only if `e.Query.Bound && len(Rows) > 0`. `Bound` is set by Proven only:
   `osv/check.go:375` (the deterministic check's own rows) and `pipeline.go:809` via `osv.Facts.Binds`
   (`check.go:115-127`: package of an in-range advisory, or an advisory/alias id of such a row). The Investigator never
   sets it (`investigator.go:131-134`). `Request` is redacted with the rest of the evidence (`pipeline.go:322-329`).
   The dependency check now runs before the Investigator (`pipeline.go:561-565`) and its result is reused by `addOSV`.
   Control: `TestQueryRowsUnrelatedToTheChangeAreNotEvidence` + negative control `TestQueryRowsBoundToTheChangesPackageAreEvidence`.
   remark: "bound to the change" is not "bound to the claim" — measured under Remainder R1.
2. **Refutation stored only with evidence — HOLDS for every refutation written by this binary; FAILS for a store written
   by the pre-fix binary** (Condition C1). Write side: `pipeline.go:1072` sets `Evidenced: judge.Refutes(c)`;
   `world.go:384` remembers only `Refuted && HasEvidence && evidenced`; `trace.go:121-124` carries it (`omitempty`, so an
   old trace replays as unevidenced — correct). Read side: `pipeline.go:688` → `hmem.MatchRefutation`
   (`hmem.go:385-402`) suppresses on any record with matching keys; the record carries no evidenced mark and
   `schemaVersion` (`hmem.go:33`) is still `1`, so a pre-fix `state.json` is read as is.
3. **Change bytes and model titles enter prompts only as encoded data — HOLDS at every builder that exists.**
   Planner: `planner.go:57-67` one JSON array, invalid UTF-8 parts as `*_base64` (`:46-52`). Investigator `:39-43`,
   Prover `:41-44`, Skeptic `:37-40`: one JSON object each via `answer.Encode` (`answer.go:44-54`, `SetEscapeHTML(false)`).
   Repair: previous answer as one JSON string (`answer.go:85-87`); `original` is the already-encoded prompt.
   Generator framing: `live.go:240` marshals the prompt as a JSON string — no byte can add a message.
   No builder still splices raw text. Fixer: does not exist (Ground).
   remark: encoding makes the frame injective; it does not stop a model from obeying instructions written inside the data — Remainder R2.
4. **Disabled guardrail named in draft and result — HOLDS.** `pipeline.go:273-275` (after the review, also on a
   deadline-cut draft, which returns through the same path), `open.go:279-282` into `Result`, `format.go:66-71` in the
   text format. `json`/`sarif` have no writer in this build (`format.go:26-27` refuses them), so text is every format.
   One sentence, `candidate.GuardrailOffStatement` (`candidate.go:176`).
5. **Stores receive the guarded generator — HOLDS.** `pipeline.go:470-473`, after `gen` is wrapped; no store calls it
   today (`hmem.go:254`, `world.go:233` say so). Control: `TestStoresHoldTheGuardedGenerator`.

### Unplug

Each mutant replaced exactly one span (asserted match-once), applied by `-overlay`, run on the package(s) named:

| mutant | what was unplugged | red test(s) |
|---|---|---|
| M1 `judge.go:73` | `Bound \|\| queryRelates(...)` | `judge: TestDropsQueryRowsThatOnlyRepeatTheFindingsWords`; `pipeline: TestQueryRowsUnrelatedToTheChangeAreNotEvidence`, `TestNegativeControl_F3RecordedQueryRow`, `TestNegativeControl_PackageTheAdvisoryDoesNotName` |
| M2 `pipeline.go:808` | `Binds` always true | the same three pipeline tests |
| M3 `world.go:384` | drop `evidenced` | `TestSkepticWordAloneDoesNotSuppressTheNextReview` |
| M4 `pipeline.go:1072` | `Evidenced` always false | `TestRefutedFindingIsAbsentOnTheNextReview`, `TestNegativeControl_ClearHMEMStore` |
| M5 `pipeline.go:274` | draft statement dropped | `TestGuardOffIsNamedInTheDraft` |
| M6 `open.go:281` | result statement dropped | `review: TestResultCarriesTheGuardrailOffStatement` (the master's own unplug, reproduced) |
| M7 `format.go:67` | text line dropped | `cli: TestTextSaysWhenTheGuardrailWasOff` |
| M8 `pipeline.go:472` | stores get raw `p.gen` | `TestStoresHoldTheGuardedGenerator` |
| M9/M10/M11 | Investigator/Prover/Skeptic back to `"id: …\ntitle: …"` | each package's `TestTitleWithLineBreakAddsNoLineToThePrompt` (+ every sample-keyed test) |
| M13 `planner.go:47` | invalid UTF-8 folded | `TestChangesDifferingInInvalidBytesNeverShareOnePrompt` |
| M14 `check.go:107` | aliases not recorded as facts | `TestNegativeControl_CVEAbsentFromKEVIsNotCalledNotExploited`, `TestKEVFindingGetsItsEPSSRow`, `TestCVEWithNoEPSSRowAttachesNothing` |
| M15 `answer.go:69` | data header dropped | `TestChangeContentCannotCreateASecondFileFrame` |
| **M12 `answer.go:87`** | **Repair writes the previous answer raw** | **none — `./internal/agent/...` all `ok`** (Condition C2) |

M12 against my probe `TestRV_RepairKeepsPreviousAsData` (overlay): `FAIL … previous answer "plain" not one JSON string`.
remark: the M-table measures that a test reddens when a fix is removed; it cannot show the fix covers paths no test drives (Remainder).

### Closure

Security reviewers' probes, re-run from `████████████████████████` (identifiers renamed, bodies unchanged) on `a37939e`:

- claude P1 (`QueryRowsUnrelatedToTheChangeAreNotEvidence`) **PASS**; its control (no rows → dropped) PASS.
- claude P2 (`SkepticWordAloneDoesNotSuppressTheNextReview`) **PASS**.
- claude P3 (`TwoDifferentChangesNeverShareOnePrompt`) **PASS** (committed verbatim as `planner/prompt_data_test.go`).
- claude P4 (`GuardOffIsNamedInTheDraft`) **PASS**; codex `GuardrailOffIsVisibleInTheReviewResult` **PASS**.
- codex `ChangeContentCannotCreateASecondPlannerFileFrame` — committed in stricter form (0 marker lines, 1 JSON record) **PASS**.

My own probes (overlay only):
- `TestRV_TitleBuildersRoundTrip`, `TestRV_PlannerRoundTrip`: eight hostile strings (forged JSON closers, `\nDATA:\n`,
  `Task:` lines, U+2028/2029, NUL, chat-template tokens, backslash runs) → exactly one data header, one line, one JSON
  value, fields round-trip. **PASS**.
- `TestRV_RepairKeepsPreviousAsData` **PASS** on `a37939e`, **FAIL** on M12.
- `TestRVG2Write` (under an overlay restoring the pre-fix `world.go` rule) then `TestRVG2Read` (fixed tree, Skeptic
  refutes nothing): **FAIL** — `F1 dropped on the upgraded binary by a record the pre-fix binary stored from a bare flag
  (H-MEM "mem_ce2874b0744e71f8")`. Control: P2 above, both phases on the fixed binary → F1 kept.
- `TestRV_G3_UnrelatedClaimOnBoundRows` **FAIL** (kept, reason `"kept"`); control with an unaffected package → dropped (PASS).

C1–C5 of the d82cac2 security review: C1 closed (support side), C2 closed for new writes / open for old stores (C1 below),
C3 closed for framing, role separation not done (R2), C4 closed, C5 closed.

### Remainder

Owed, not conditions of this verdict; each measured or located:

- **R1 · rows bound to the change, not to the claim.** Measured: Planner title "SQL injection in the discount helper"
  on `billing/discount.go`, Investigator `osv-by-package` for `evil.example/leftpad` → finding **kept** as evidenced
  (`TestRV_G3_UnrelatedClaimOnBoundRows`). The order's property 1 is met as written; the claim's words remain the
  model's. Remedy for the master: require the bound subject to appear in the finding's rule/path, or show the bound
  subject beside the title. Kin of the d82cac2 review's O4 (Prover evidence bound to the change, not the claim).
- **R2 · no role separation.** `live.go:240` still sends task and data in one `user` message; both prior reviews asked
  for the task as a system message. Encoding closes framing; it does not separate authority.
- **R3 · the refutation side still word-matches.** `judge.go:114, 122-165`: `QueryContradicts` uses `queryRelates`
  (a 4+-letter word shared with the model's text) plus a denial cell; and any Skeptic-named existing test that passes
  on the change refutes (`judge.go:103`). Such a refutation now counts as `Evidenced` and is stored in H-MEM. The sample
  tables carry no denial cell, so not reproduced end to end — read only. Remedy: bind refuting rows the way property 1
  binds supporting rows.
- **R4 · behaviour change on a down dependency check.** `pipeline.go:561-565`: if `osv.Check` fails, `Facts` is empty
  and every Investigator row is unbound, so model findings fall to `dropped: no evidence`. Safe direction; the
  `QueryNote` says rows were unbound, not that the check was down. Read only.
- **R5 · `Facts` matching is case- and normalisation-exact** (`check.go:103,119-125`; PyPI names fold elsewhere,
  `manifest/gomod.go:182`). Only false negatives (rows left unbound). Read only.
- Declined (R11): live-model behaviour with the new data header; TLS of live clients; licences (out of scope).

### Profile

Security.
- S1 trust boundary: change → prompts (encoded, P3) → model answers → instruments chosen by models (bound by Proven
  on the support side; unbound on the refutation side, R3) → Judge → H-MEM (evidenced writes only; legacy reads, C1)
  → result (guardrail state named, P4).
- S2: two gaps reproduced by failing tests (G2, G3); M12 gap shown by mutant + probe.
- S3: fifteen mutants, fourteen red, one green (C2).
- S4: no `go.mod`/`go.sum` change in the diff (`git diff --stat main...a37939e` lists neither); no secret, bidi or
  network addition seen in the 37-file diff (read).

### Focus

Does each of the five properties hold on every path, and which test turns red when it is removed?
Properties 1, 3, 4, 5 hold on every path present in the tree, each guarded by a test that reddens when unplugged
(Unplug table), except the Repair path of property 3, which holds but has no guarding test (C2). Property 2 holds for
every refutation this binary writes and fails for refutations stored by the pre-fix binary (C1). Every prompt builder
encodes; none splices raw text; there is no Fixer.

### Conditions

- **C1 · H-MEM written by the pre-fix binary still suppresses evidenced findings.**
  Where: `internal/hmem/hmem.go:33` (`schemaVersion = 1` unchanged), `:385-402` (`MatchRefutation` reads every record),
  `internal/pipeline/pipeline.go:688`. Proof: `TestRVG2Write` → `TestRVG2Read` FAIL (Closure). Remedy, one of: bump
  `schemaVersion` so `Open` refuses an old `state.json` by name (with the clear command in the refusal), or mark records
  written under the evidenced rule and have `MatchRefutation` ignore unmarked ones; add the two-phase test as a
  regression. Class: the master decides; I hold it non-blocking only if no persisted pre-fix store exists outside tests.
- **C2 · the Repair encoding has no guarding test.** Where: `internal/agent/answer/answer.go:85-87`. Proof: mutant M12
  leaves `./internal/agent/...` green. Remedy: commit `TestRV_RepairKeepsPreviousAsData` (text below), which fails on
  M12 and passes on `a37939e`.

```go
// in internal/agent/answer (package answer_test), with imports bytes, encoding/json, strings, testing,
// answer, skeptic. hostile is any list of strings with quotes, line breaks and "\nDATA:\n".
func TestRV_RepairKeepsPreviousAsData(t *testing.T) {
	orig := skeptic.Prompt("F1", "t")
	for _, s := range hostile {
		r := answer.Repair(answer.Spec{Task: "T", Shape: "S", Example: "E"}, s, orig)
		_, after, ok := strings.Cut(r, "Previous answer, written as one JSON string:\n")
		if !ok {
			t.Fatal("no previous-answer section")
		}
		line, _, _ := strings.Cut(after, "\n")
		var got string
		if err := json.Unmarshal([]byte(line), &got); err != nil || got != s {
			t.Fatalf("previous answer %q not one JSON string: %q (%v)", s, line, err)
		}
		if bytes.Count([]byte(r), []byte("\nDATA:\n")) != 1 {
			t.Fatalf("repair prompt grew a data header from the previous answer %q", s)
		}
	}
}
```

- **C3 · the Unrun list is non-empty (law #53), so this is not an `accept`.** The 77 listener-file tests (Unrun),
  including `TestGuardrailOffCompletesAndSaysNoGuardrail`, must be green on a seat that allows loopback, on the
  landed tree, before merge. The master's 2026-10-24 macm5 reading covers `a37939e`; a re-absorb of `main`
  (now `6547849`) needs it again.

Full probe texts (pipeline G2/G3/P1/P2/P4, answer round-trips) are in the run's artifact directory under `probes/`.

### Unrun

- Listener-bound tests (sandbox refuses loopback: `listen tcp6 [::1]:0: bind: operation not permitted`, measured
  2026-10-24): every `Test*` in `cmd/livecheck/main_test.go`, `internal/generator/live_test.go`,
  `internal/guardrail/live_test.go`, `internal/pipeline/guardrail_test.go`, `internal/query/live_test.go`,
  `internal/review/{live_generator,open,secret_echo,zetaris}_test.go` — 77 functions, including
  `TestGuardrailOffCompletesAndSaysNoGuardrail`, `TestUnsafePromptStopsTheStepAndNamesCategories`,
  `TestUnsafeAnswerIsNotUsed`, `TestClosedGuardrailRefusesTheStepByName`, `TestOpenSampleReviewKeepsF1AndF2`. Some of
  these open no listener but were skipped as collateral of the file-level skip. The master's outside-sandbox green
  (2026-10-24, macm5) is the only reading of them; I did not reproduce it.
- `-race` and `-count>1`: not run.
- Live mode (Nightshift, NVIDIA, guardrail, Zetaris): not run (order: no network).
- The NVIDIA independent review (owner's rule 2026-10-22): owed, route down.
- Mutants were run on their target packages only, not the whole module.

## The review status as filed

## REVIEW — security-piece-5-fixes
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: W45 at a37939e (diff main...a37939e; the order's base 00d4ab9 is a tree, not a commit): properties 1, 3, 4, 5 hold, each unplug reddens a test (14 of 15 mutants).
  Conditions: C1 an H-MEM store written by the pre-fix binary still suppresses evidenced findings (two-phase probe red); C2 Repair encoding has no guarding test (mutant stays green);
  C3 77 listener tests unrun in this sandbox. Remainder: rows bound to the change, not to the claim (measured); refutation side still word-matches; no role separation.
  Reviewed on claude-code, model claude-opus-5-5; the NVIDIA second review is owed.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T07:20:00Z

## REDACTED

| what | times |
|---|---|
| private workspace path | 1 |
