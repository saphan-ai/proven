# TRAIN ORDER — proven-reviewer-core

- master: MASTER (claude-code), appointed by the owner on 2026-10-21 (R1)
- status: SIGNED by the owner on 2026-10-21 (draft 5eb97ecb0c6125c955ad414f167f9d2b, valid until withdrawn); §3 then filled
  in by the master with the owner's answers, no other change. **No wagon is dispatched before 2026-10-22 00:00 UTC.**
- track: proven-reviewer-core
- repo: proven (`saphan-ai/proven`), the product code
- base: main at the cut, 2026-10-22 or later
- author: ADVISOR of the hackathon track "The Agent That Can Explain Why"

## §0 THE OWNER, VERBATIM

2026-10-21 (translated from Polish): *"Working first, pretty later."* · *"Explore today, build tomorrow."* ·
*"Act."* (said to the advisor, who then found that no order existed yet for the reviewer itself).

The design is the owner's, in his accepted decisions; this order only puts them in a building sequence.

## §1 WHAT THIS TRAIN IS

The reviewer itself: what reads a change, finds what is wrong with it, proves it or drops it, lets a model from another
family try to refute it, and writes the review. Its outer shell (command line, formats, Action, comments) is the train
[`proven-cli-and-action`](TRAIN_ORDER_proven-cli-and-action.md); **the seam between the two is the Go interface that train
defines in its W1.** Where the two disagree, that train's W1 is the contract and this train adapts.

**"Working first" is the shape of the train.** W1 is a thin slice: the whole review end to end, in sample mode, with every
agent as simple as it can be, so the CLI prints a real review on the first day. Every later wagon deepens one part of a
review that already runs. The sample mode stays working at every step: a judge with no keys sees the whole review
([ADR-0001](../decisions/ADR-0001-go-behind-adapters.md), [ADR-0019](../decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) point 7).

**The decisions this train builds, all accepted:**
[0001](../decisions/ADR-0001-go-behind-adapters.md) Go behind adapters ·
[0002](../decisions/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) evidence is a test or a Zetaris query ·
[0003](../decisions/ADR-0003-meterless-engines-in-go.md) and [0018](../decisions/ADR-0018-meterless-engines-minimum-and-their-contract.md) H-MEM and World Model, minimum, one generator, JSONL ·
[0005](../decisions/ADR-0005-guardrail-on-by-default.md) guardrail on by default ·
[0006](../decisions/ADR-0006-a-skeptic-from-another-model-family.md) the Skeptic ·
[0007](../decisions/ADR-0007-no-exploit-code.md) no exploit code ·
[0008](../decisions/ADR-0008-reviewer-model-chosen-against-the-authors-model.md)–[0014](../decisions/ADR-0014-review-class-chosen-by-configured-rules.md) the reviewer's model, its family, review classes, no hardcoded settings ·
[0015](../decisions/ADR-0015-zetaris-through-its-rest-api.md) and [0023](../decisions/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) Zetaris over REST ·
[0019](../decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) Nightshift the primary provider ·
[0024](../decisions/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md) every Cursor session recorded.
The agents are those of `DESIGN.md`: Planner, Investigator, Prover, Skeptic, Judge, Writer.

**What the advisor read on 2026-10-21. CLAIMS FOR YOU TO TEST, not ground (law #44).** Re-measure each before a wagon
builds on it; where the advisor is wrong, the wagon's `master-was-wrong:` field says so.

| # | claim (advisor) | source | instrument |
|---|---|---|---|
| RC-1 | Nightshift answers OpenAI-compatible `POST /v1/chat/completions` with the headers of ADR-0019 point 3, and its answer carries a signed receipt naming the routed model. | the setup guide; the `proven-nightshift-vm` train measured one request and its receipt | one request from runner-cursor, the receipt's field names pasted |
| RC-2 | The receipt fields are `meterless.routed_model` and `meterless.winning_route.provider`. | ADR-0019 point 4, from the guide | the same receipt |
| RC-3 | Our own guardrail model server answers safe/unsafe with categories. | the `nemotron-content-safety` train, done | one safe and one unsafe call |
| RC-4 | Zetaris answers KEV, EPSS (fixed CVE or page) and OSV by advisory id as REST tables, read live with Proven's own key; OSV by package does not. | ADR-0023, measured 2026-10-21 | one SELECT each with Proven's key |
| RC-5 | OSV by package is answerable from a copy in Zetaris's storage (the owner's option 1). | **NOT MEASURED**: the probe was asked of the master on 2026-10-21 and has no answer yet | the probe's result; until then W5 builds the three working sources only |
| RC-6 | Meterless publishes specifications and conformance suites for H-MEM and the World Model that run against an implementation through a TypeScript adapter. | ADR-0003, ADR-0018 | the suites' location pasted, and one suite run against an empty adapter (it must fail) |

## §2 WAGONS

Executors: Cursor seats on runner-cursor, `grok-4.7-high` by default; `--model` on the dispatch line. **The owner, 2026-10-22
(translated from Polish):** *"Proven's coding goes on Grok, and the review goes on one of the other models plus, in
parallel and mandatory, NVIDIA."* Every review of this train is therefore two independent reviews of the same diff, run in
PARALLEL, never one after the other, and neither sees the other's verdict before returning: **(a) mandatory**, an
`nvidia/*` model through Nightshift with its signed receipt (ADR-0019 amendment 1: no `nvidia/*` receipt = the review is
refused; the Nightshift call streams); **(b)** a non-Grok model when such a review seat is free; when none is free, Grok
on a Cursor seat is allowed (the owner's rule as relayed by the advisor, 2026-10-22). The rule binds wagons whose review is **named**;
master-verified wagons stay master-verified (the owner, 2026-10-22: *"nazwane"* — "the named ones"). Every session
recorded (ADR-0024). Reviews on `codex` with `--role review`, the focus phrased as the property that must hold (law #9,
law #56). Intermediate wagons are master-verified; the last carries the stream's named review (law #29).

### W1 `a-review-end-to-end-in-sample-mode`
- **what:** the six agents as the simplest code that can work, behind the seam of `proven-cli-and-action` W1: a change
  in → a review out, in sample mode, with no network and no key. The generator interface of ADR-0018 point 2 with a
  sample generator; the World Model as an append-only JSONL trace in the run's `logs/` (ADR-0018 point 3); a fixture
  change with one finding that is proved by a test, one proved by a sample query, and one with no evidence.
- **dod:** `proven scan --sample <fixture>` prints a review with the two proved findings and **without** the third
  (ADR-0002 point 2); the trace follows each finding from claim to verdict. **Negative control (law #27):** give the
  third finding evidence in the fixture and it appears; remove the test of the first and it disappears.
- **master-was-wrong:** —

### W1b `the-prover-in-a-container`
- **added:** by the owner 2026-10-21, on the master's proposal (no wagon built ADR-0025's runner): *"tak, dodać W1b."*
  (translated: "yes, add W1b").
- **what:** ADR-0025: the test-run adapter's live implementation. The change's code and tests run only in a container
  Proven starts for that one run: no network, no secret, only the checkout it needs; image, time and resource limits are
  settings validated on load (ADR-0013), with no setting that gives a network or a secret. No container runtime → the
  Prover runs no code, and the review says so (point 6). Never a fallback to the host.
- **dod:** on runner-cursor, the W1 fixture's F1 is proved by a test run in the container: fails on the change, passes on
  the base. **Negative controls:** a test that opens a network connection fails inside the container and the review says
  it could not run and why; a test that reads Proven's token finds none. **Unplug (law #41):** no container runtime on
  the path → the Prover is named off, F1 is not reported as proved, and no process of the change runs on the host
  (measured by a test).
- **master-was-wrong:** —

### W1c `the-reviewer-honours-the-request`
- **added:** by the master, 2026-10-22, from two BLOCKING findings of the review of `proven-cli-and-action` W1b (message
  `f8ca217af5e64c1b`): the reviewer ignored the request's baseline, head, include and exclude, and never read its
  configuration file.
- **what:** the request reaches the pipeline (changed files from the baseline, selection by include/exclude, a missing
  baseline is fatal); the configuration file is read with a closed key set, refused by name on load (exit 7 in the shell).
- **dod:** the test plan's nine cases, written by a separate test author, green on the implementation; `--exclude='*'`
  prints no finding; `--baseline-commit=does-not-exist` fails naming it.
- **master-was-wrong:** —

### W1d `a-partial-review-reports-only-judged-findings`
- **added:** by the master, 2026-10-22, from the retroactive codex review of W1 (verdict reject, run
  run-20261022-032603-536d8bb3): two BLOCKING findings on `main`.
- **what:** (1) a review cut by its deadline reports only findings the Judge kept; an unjudged candidate is never in the
  result (ADR-0002 point 2); (2) the trace carries the full evidence (test outcome and output; query columns and rows)
  and the full refutation (attempted, refuted, reason) (ADR-0018 point 3). Dispatched after W1c lands (both touch the
  pipeline).
- **dod:** a test with a deadline after planning: no unjudged finding in the result; the trace of F1, F2 and F3 carries
  the full evidence and refutation. **Negative control:** remove the partial-path filter and the deadline test turns red.
- **master-was-wrong:** —

### W2 `the-generator-through-nightshift`
- **what:** the live generator: Nightshift by default (ADR-0019), any OpenAI-compatible endpoint by setting; base URL,
  token and model or routing mode from the environment, validated on load (ADR-0013). Each review names the route and
  model read from the receipt, never assumed (ADR-0019 point 4). RC-1, RC-2 measured. **Also (claim of the
  infrastructure master, 2026-10-21, to test):** ADR-0019 amendment 1 (proven `383898c`): a review whose receipt names no
  `nvidia/*` model, or has no receipt, is refused; every Nightshift call streams, because Nightshift cuts a request that
  sends no headers within 45 s.
- **dod:** one review on runner-cursor through Nightshift names the model from the receipt. **Negative controls:** a
  wrong token → the adapter's named error and the review degrades (ADR-0001 point 3), never a hang; a response without a
  receipt → the review says the model is unknown, it does not guess. **Unplug (law #41):** stop the Nightshift VM; the
  same review runs degraded and says which service failed.
- **master-was-wrong:** —

### W3 `the-guardrail`
- **what:** ADR-0005: every prompt and every answer through the guardrail adapter, on by default; two backends (our
  model server, NVIDIA's hosted API); a setting turns it off and the trace says so; `unsafe` stops the step and is
  written with its categories. RC-3 measured.
- **dod:** a fixture prompt that the guardrail marks unsafe stops its step, the trace names the categories. **Negative
  control:** the same run with the guardrail off completes and the trace says *no guardrail*. **Unplug:** guardrail
  server down → the step is refused by name, not passed through.
- **master-was-wrong:** —

### W4 `world-model-and-h-mem-minimum`
- **what:** ADR-0018 points 1, 4, 5, 6: the World Model's store, ingest and context builder; H-MEM's store and
  retrieval; the World Model's memory backed by H-MEM. The feature a user sees: a finding refuted in one review does not
  come back in the next (ADR-0006 point 4). RC-6 measured; the conformance suites run where they run (ADR-0018 point 7).
- **dod:** two reviews of the same fixture: the finding refuted in the first is absent from the second, and the trace
  says why. A run killed halfway and started again gives the same state (ADR-0018 point 5, a test that does exactly
  that). The conformance results pasted, or the reproducible feature test where a suite does not run, with the reason.
  **Negative control:** clear H-MEM and the refuted finding returns.
- **master-was-wrong:** —

### W5 `evidence-through-zetaris`
- **what:** the Zetaris adapter over REST (ADR-0015, ADR-0023): the Investigator's queries over KEV, EPSS and OSV by
  advisory id, read with Proven's own key; the same queries on a local engine over small samples in sample mode
  (ADR-0002 point 4). Settings from the environment. OSV by package only once RC-5 is measured (an ADR written from the
  probe; ADR-0025 is R2).
- **dod:** a fixture change that adds a dependency version covered by an advisory on the KEV list gives a finding whose
  evidence is the query and its rows (ADR-0007 point 2), live and in sample mode. **Negative control:** a version the
  advisory does not cover gives no finding. A join drops a CVE absent from one side: the review never reads a missing
  row as "not exploited" (ADR-0023). **Unplug:** Zetaris down → degraded, named.
  **Hard point, Wildcard track (the owner via the advisor, 2026-10-22):** by **2026-10-27**, a LIVE Proven review draws
  at least one finding's evidence from Zetaris: the query's SQL and rows appear in the trace, recorded and demonstrable,
  not only in sample mode. W5 carries it; W5 is dispatched as soon as W2 (the live generator) lands.
- **master-was-wrong:** —

### W6 `the-skeptic-and-the-judge`
- **what:** ADR-0006: the Skeptic, from another model family than the agent that made the finding (Nemotron through
  Nightshift by default, a setting), tries to refute every finding; the Judge settles by the evidence, never by a vote;
  a refuted finding is dropped and written to the trace and to H-MEM. Findings the evidence does not settle are reported
  as unresolved (`unresolved-count`, ADR-0020).
- **dod:** a fixture where the Skeptic's refutation carries a test that passes on the change: the finding is dropped. A
  fixture where the refutation has no evidence: the finding stays. **Negative control:** set the Skeptic to the same
  family as the finder and the review says so, by name (ADR-0009, ADR-0010).
- **master-was-wrong:** —

### W7 `the-reviewer-model-and-the-review-class`
- **what:** ADR-0008 to ADR-0014: the author's model read from the `Assisted-by:` trailer and the other configured
  sources; the family table; the reviewer chosen against the author's family; review classes by configured path rules,
  highest-risk wins; security reviews doubled. Every list closed, validated on load, refused with the allowed values.
- **dod:** a table test per rule. **Negative controls:** an unknown family in the table → refused on load, exit 7; a
  path rule naming an unknown class → refused on load; a change with no trailer → the review says the author's model is
  unknown and which default it used.
- **review:** named — the stream's review (law #29).
- **master-was-wrong:** —

## §3 DECISIONS OWED BY THE OWNER

- **R1 — the master. DECIDED by the owner 2026-10-21:** *"We can have a new master."* (translated from Polish). A second
  master holds this train; the shell and the reviewer are built in parallel.
- **R2 — where the Prover runs the change's tests. DECIDED by the owner 2026-10-21:** *"OK."* The Prover runs the change's
  code only in a container with no network and no secret; with no container runtime it runs no code. Written as
  [ADR-0025](../decisions/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md).
- **R3 — OSV by package. OPEN.** Waits for the Zetaris storage probe (RC-5), asked of the infrastructure master; the owner
  chose option 1 on 2026-10-21. The infrastructure master, 2026-10-21: OSV by package through a Zetaris REST source is
  impossible (the nested POST body is sent as a string and OSV answers 400); OSV by advisory id works; the storage probe
  is not yet run, its result promised before 2026-10-22 00:00 UTC. Until the probe answers, W5 builds KEV, EPSS and OSV by advisory id only. The ADR for OSV
  by package is written from the probe's result (ADR-0025 is now taken by R2).

## §4 DEFINITION OF DONE

- `proven scan` reviews a change end to end, live through Nightshift and Zetaris, and in sample mode with no keys.
- Every finding in a review has a test or a query with its rows; a finding without either is not in the review.
- The Skeptic from another family tries every finding; the evidence decides; what it cannot decide is reported as
  unresolved.
- A finding refuted once does not come back; the World Model's trace in `logs/` follows each finding from claim to
  verdict.
- Every model a review used is named from the provider's receipt; every refusal names what was refused and what is
  allowed; every wagon's report has its run id, start time and session recording.

## §5 COST

- Cursor seats: subscription; each seat's allowance is the scarce quantity (law #51). Each wagon's quote is the master's,
  before dispatch; its cap is twice the quote (law #16).
- Model calls in tests: sample mode by default; a live call through Nightshift to NVIDIA's hosted API is priced by
  NVIDIA's terms and runs only with the owner's consent at the call (law #49).
- Reviews on `codex`: quoted at the review order.
