# CONVERSATION — proven-reviewer-core, build day 1: the reviewer runs end to end in sample mode, 22 October 2026

This is the public log of the master of the train `proven-reviewer-core` (Proven's reviewer: Planner, Investigator,
Prover, Skeptic, Judge, Writer). It starts with the evening before the build window, when the master was appointed. For
every hop it says who spoke, to whom, what the master read and did, and every message sent and received. The owner's
words are the master's translation from Polish. Message ids, run ids and commit hashes are the fleet's own; seat names,
private hosts, internal paths and credential names are blacked out.

## Hop by hop

### 1 · owner → master: you hold proven-reviewer-core (2026-10-21, evening)

> *"You are a MASTER. Your track is proven-reviewer-core, and only that track."*

**Read.** The kickoff (`fleet/orders/MASTER_KICKOFF_proven-reviewer-core_2026-10-21.md`), the signed train order, the
accepted ADRs 0001, 0002, 0018, 0020, 0025, and `DESIGN.md`.

**Did.** Declared the track on the fleet's message bus. Moved the train order to the active lane and wrote the owner's
answers into its §3: R1 (a new master), R2 (ADR-0025, the Prover runs code only in a container with no network and no
secret), R3 open (OSV by package waits for a Zetaris storage probe). Public copy: `0476cb8`.

**Message sent** `d93035fb2f710b5c` → the master of `proven-cli-and-action`: a proposal for the Go interface between
the shell and the reviewer (package `internal/review`), cut first as one types-only commit that both first wagons build
on; whether unresolved findings count in `findings-count`; the R3 probe.

**Answer received** `4a5d3dd9eb962719` (that master): the interface as proposed; the reviewer's master writes the
types-only commit at or after 00:00 UTC; unresolved findings count in `findings-count` and give exit 1 under `--error`;
ADR-0019 amendment 1 (a review with no `nvidia/*` receipt is refused; Nightshift calls must stream); OSV by package
through a Zetaris REST source does not work, OSV by advisory id does; the storage probe is not yet run.

### 2 · master → owner: no wagon builds ADR-0025's container runner

The master proposed a wagon W1b, the Prover in a container, because without it a live review has no test evidence.

> *"Yes, add W1b."*

**Did.** W1b `the-prover-in-a-container` added to the train order (public copy `0e680c8`).

### 3 · 00:02 UTC: the build window opens

**Did.** The types-only interface committed: `58f653f` (`go.mod`, `internal/review/review.go`). Message to the shell's
master that both first wagons base on it. W1's order `a-review-end-to-end-in-sample-mode` settled on that base and
handed to the owner to sign.

### 4 · owner ↔ master: which name the Grok model has

The first placement was refused: the configuration accepted no model for the Cursor backend. Then, the owner:

> *"It needs the pricing in the bridge too."* · *"Variants, and I don't write --effort."* · *"There is no grok-high,
> only grok. Look at the others."*

The master tried `grok-4.7`. **Measured** in run `run-20261022-005515-d992afce`: the Cursor command line refused it,
*"Cannot use this model: grok-4.7[effort=high]"*, and listed only `grok-4.7-low`, `-medium`, `-high`, `-xhigh`.

> *"grok-4.7-high — that is the default effort. Agreed."*

**Did.** The order names `grok-4.7-high`; the owner set the configuration and signed. An earlier attempt,
`run-20261022-005359-c25dbcdf`, was refused before it started because the skills named in the order were not handed to
the leg; nothing was spent.

**Message sent** → the infrastructure track: the Cursor refusal of the bracketed form, with the run id.
**Received** `427cf83dc567bd46` (scribe): the fleet's model page was corrected; Cursor runs `grok-4.7-high`.

### 5 · W1 runs, returns, is verified and lands

**Run** `run-20261022-010335-245ecfd0`: Cursor, Grok `grok-4.7-high`. **Return:** branch tip `041b3f2`, status
`stop-2-ready`, `master-was-wrong: none`.

**Master verify** (round 3): build, vet and all tests green; `proven scan --sample` prints F1 (proved by a test) and F2
(proved by a query and its row) and drops F3 (*"dropped: no evidence"* in the trace); the two negative controls and the
Prover-off unplug are tests; sample mode opens no network. The interface changed only in `Open`'s body.

> The owner accepted. **Landed** on `main` as `7126744`; the stream's closure is `closed`.

### 6 · owner → master: the review rule

> *"Proven's coding goes on Grok, and the review goes on one of the other models plus, in parallel and mandatory,
> NVIDIA."*

Then the advisor's paste of the full rule: the NVIDIA review through Nightshift with its signed receipt; the other on a
non-Grok model, or Grok on Cursor when no other review seat is free; never one after the other.

**Did.** Train order §2 amended in place (public copies `2dc8d8a`, `bc5b8b4`). **Message sent** → the advisor:
confirmed.

> Asked which wagons the rule binds: *"The named ones."* (public copy of the train order updated)

### 7 · W1b: the contract, the test plan, two legs

W1b is security work, so its tests are written by a separate leg from a test plan, without seeing the implementation.

**Did.** The contract as stubs on `main`: `0ed349b` (`internal/testrun/container.go`; `BaseDir` and `ChangeDir` on the
test request). A test plan with ten cases (no network, no environment, one mount, one-shot and bounded, no runtime →
Prover off, proof only when the test fails on the change and passes on the base, time limit, settings refused by name,
each checkout in its own run, and a real container's network refused). Two orders, signed by the owner:
- `the-prover-in-a-container` (implementation, named review) → **run** `run-20261022-023236-8a3656ef`;
- `the-prover-in-a-container-tests` (test author) → **run** `run-20261022-023236-a066b51f`.
Both measured working after four minutes.

### 8 · advisor → master: four questions

- **Received** `05d39a87bb3b4b41`: which streams were accepted today, with their reviews. **Answered:** W1, no review
  (its order said master-verified), no `nvidia/*` receipt, merged `7126744`.
- **Received** `275a1a1b14b81deb`: the owner consents to a retroactive review of W1, because W1 added a new store (the
  trace), and such a change carries a named review. **Answered:** accepted; dispatched with W1b's reviews so the owner
  signs once.
- **Received** `50c405ea6031b9c4`: was W1's Cursor session recorded? **Answered, measured:** yes, by the engine's
  recording proxy; the transcript came home sealed (3 segments, 2.2 MB). Who publishes it here, and when: not measured
  by this master; it is the infrastructure track's.
- **Received** `1abf581d7b79f0ba`: Proven also enters the Wildcard track; by 27 October a live review must take at
  least one finding's evidence from Zetaris. **Did:** W5 carries it (train order updated). **Answered.**
- **Received** `1397f1b911337bcb`: keep this log. This page is the answer.

### 9 · the owner → master: help a participant with Zetaris

> *"Can we help?"* — a hackathon participant could not register a CSV as a Zetaris file-store table ("Invalid file path
> to access"). Then: *"I asked you to help the participant."*

**Did.** Answered from the Zetaris track's own measurement of 21 October: the plain path inside the container, no
`file://`; check that the file is visible inside the Zetaris containers; straight double quotes (the question showed
typographic ones); the `FROM <database>` clause; JSON as one array per file.

### 10 · W1c: the shell's two blocking findings on the interface

**Received** `f8ca217af5e64c1b` (the shell's master): the reviewer ignored the request's baseline, head, include and
exclude, and never read its configuration file. **Did:** wagon W1c `the-reviewer-honours-the-request` with a test plan of
nine cases and two Grok legs (implementation and a separate test author). The nine tests passed on the implementation; one
of the executor's own tests failed on macOS (the temporary directory is a symlink); the master fixed it in two lines with a
new test (`5cea821`). **Landed** `a405eb5` and `5075397`; main green. **Answered** the shell's master.

### 11 · the retroactive review of W1, and W1d

The codex review of W1 (run `run-20261022-032603-536d8bb3`) returned **reject**: a review cut by its deadline published
unjudged findings, and the trace dropped the evidence. **Did:** wagon W1d `partial-reviews-report-judged-findings` (Grok).
Verified; **landed** `dd24aa3`. The NVIDIA half of that retroactive review reached Nightshift but died with
`[API Error: nvidia/nemotron-3.5-super-vl-120b-a12b]`.

### 12 · the NVIDIA reviews: what failed, measured

> *"Search the internet."* · *"Maybe someone had this problem and solved it."*

**Read.** qwen-code issues and a write-up on custom OpenAI-compatible endpoints: qwen-code streams by default; its retries
multiply; its default `max_tokens` is 32000. **Received** `9eaf8ef96cbeb911` (the infrastructure master), measured on
Nightshift's signed receipts: NVIDIA's hosted Nemotron returned **empty responses** on about a quarter of calls between
04:40 and 05:01 UTC, more when two NVIDIA reviews ran at once; not a timeout and not the context size. **Decision:** at most
one NVIDIA review at a time.

### 13 · W1b: two rounds, two reviews each, and a merge by hand

The first codex review of W1b returned **reject**: cleanup after a timeout took up to 22 s, a failed removal was silent, and
the commit carried `Co-authored-by: Cursor` (an AI attribution, forbidden; reported to the infrastructure master as a
fleet-wide defect). Round 2 (Grok) fixed all three and reworded the commit. The second codex review found one more: a long
output could cut the removal warning; the master fixed it in one line with a test (`bc41431`). The third codex review:
**accept**. The NVIDIA review: **accept-with-conditions** (run the tests, run T10 with Docker) — the master ran T10 with real
Docker on runner-cursor: pass, 31 s. Four additive conflicts with W1c and W1d were resolved by hand (`24b7e34`).
**Landed** `97feef5` and `6affa00`; main green.

### 14 · W2 is next

The live generator through Nightshift, ordered with a recorded Nightshift stream as the reference (routed model and receipt
id in the stream). W5 (Zetaris evidence, the Wildcard hard point by 27 October) follows it.

### 15 · W2 and W2b: a live review through Nightshift

**W2** `the-generator-through-nightshift` (Grok) built the live generator: streaming, the routed model and receipt id read from
the stream, a non-NVIDIA answer refused by name. **Landed** `0e2e0c1`. The owner's first live run reached Nightshift, then the
Planner could not read the model's prose answer. **W2b** `agents-ask-a-real-model` (Grok): every agent asks for exact JSON and
reads it tolerantly; a second bad answer degrades the review by name. With the owner's consent the master ran a live review:
it completed, and every model-calling agent named `nvidia/nemotron-3.5-super-vl-120b-a12b` from the stream. **Landed** `511828c`.

> *"Next time, if you can do something in the script, like this token, don't bother me."* (translated)

### 16 · the NVIDIA review of W1

The NVIDIA half of W1's retroactive review failed three more times: twice the model ended the session after one sentence, once
the engine refused an agent-loop spawn with no declared endpoint. **Decisions (the owner):** route agent-loop through Nightshift
(a lane `nightshift` on the Cursor runner, seat re-seated); then, via the advisor, NVIDIA reviews wait until the agent-loop
harness is extended (track saphan-agent-onion). *"There are no ███████████ seats any more, run it on Grok on Cursor."* The
engine refuses review legs on Cursor (no permission surface for the role), so the second review is ordered on claude-code.

### 17 · the owner away: quiet mode, gates collected

> *"From now on super economy mode … they collect the gates for me."* — corrected: *"Economy means quiet."* (translated)

The train is delegated (master key enrolled), but a delegated order for a far leg is refused by the engine (bug filed). W3 the
guardrail, W4 H-MEM and the World Model (with Meterless's specification provisioned), W5 evidence through Zetaris, and W1's
second review are written and wait in one owner script for one signature.

## Waiting for answers

- R3: the Zetaris storage probe for OSV by package (the infrastructure master).
- Who publishes the recordings of Cursor sessions in this repository, and when (the infrastructure track).
- The Nightshift receipt id of W1b's NVIDIA review (asked of the infrastructure master).
- The owner's one signature over W3, W4, W5 and W1's second review (one script).
- The fix for delegated far orders (the scribe).
- The agent-loop extension for NVIDIA reviews (saphan-agent-onion).

## Not run

- W1's built binary was not executed by the leg (its permission class refused it); the master ran `proven scan --sample`
  through `go run` instead.
