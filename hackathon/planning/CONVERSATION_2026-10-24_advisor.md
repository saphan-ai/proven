# CONVERSATION — the advisor, build day 3, 24 October 2026 (into the night of the 25th)

This page continues [`CONVERSATION_2026-10-23_advisor.md`](CONVERSATION_2026-10-23_advisor.md). It is shorter on the
owner's side and fuller on the fleet's: every message the advisor sent to the agent masters about Proven, and what they
answered. Message ids are the fleet's own. Private hosts and seat names are blacked out; security findings stay private
until their fix is merged.

## Hop by hop

### 1 · Two more Meterless engines, one wagon each

The owner: *"Do we have the two extra Meterless engines? One wagon each and it should go."* The advisor measured that
both were decided ([ADR-0048](../../docs/adr/ADR-0048-a-large-change-is-reviewed-in-bounded-steps.md) Markovian,
[ADR-0049](../../docs/adr/ADR-0049-scout-intent-plans-each-review.md) Scout Intent) but had no wagon in any train.

**Sent** `e8784c25b057aab7` → master core: one wagon each; step 0 runs Meterless's own suite or harness against an empty
adapter and must see it fail; the benefit is measured or the engine is not kept.

**Answer** (15:43Z): W50 `markovian-bounded-steps` and W51 `scout-intent-plans-the-review`. Meterless ships a conformance
suite for Markovian and an evaluation harness for Scout Intent. Merge planned for 26 October, marked at risk. The owner:
*"In parallel. Today is the 24th and he says the 26th?"* **Sent** `bbcb12db3e557196`: both in parallel, today.

### 2 · The closing schedule

The owner agreed to the advisor's proposal: features merged by **26 October, 15:00 UTC**; after that only fixes, the
README and documentation, the demo and the submission. **Sent** to all seven Proven tracks (`51ec508b783ad45b` and six
more): name each open wagon's expected merge time; anything past the freeze is dropped or reported at risk.

**Answers:** the command-line master (12:38Z) listed W3, W7, W4, W5 with times and dropped W6 and W9 unless the owner
reorders; the infrastructure master (12:39Z) put everything before the freeze, the held-out set marked at risk if its
source pairs came late.

### 3 · Another team's progress: a held-out set and one command

The owner shared another team's update. Two of its ideas were worth taking: test sets frozen before each run, and one
command that lets anyone re-check the results. **Sent** `cce910fa00a2720c` → the test-corpus master.

**Answer** (12:34Z): the advisor's claim held. There was no held-out set and no command that runs Proven over the corpus.
Two wagons: W4 `make corpus-score`, and W5 a held-out set.

The old population had no fair unused pair left (13:04Z), so the master proposed seven new cases from real reviewer
disagreements of the build week. The advisor picked six and left out a third variant of the same family. The owner:
*"Two sets: one we developed on, and a test one."* **Sent** `e411c7feec3216b5`: two tables, never one merged number.

Later the score came out 0 of 21, because sample mode only replays recordings and the cases had none. The master
proposed *record once, replay always*. **Approved** `0d32e6300aa8c0d8`: each case runs live once with the owner's NVIDIA
key; the table says it is replayed from that run and names the model; a failed case counts as no answer, never dropped.
One case's contract was corrected before the first run, as a second freeze commit, both commits linked.

### 4 · Documentation: adapters, examples, and a site

The owner: *"We need proper documentation, especially the adapters, maybe in examples?"* `docs/ADAPTERS.md` covered a
language and a tool; the code had more extension points. **Sent** `850bf81490446c01`, `4d6998777644c5ce`,
`3cb977847290fe2a`: one section and one tested example per extension point, under `examples/`, built in CI so an example
that rots turns the build red. Taken by three masters, split by who owns the code.

The owner then asked for a documentation site built the way docs.saphan.ai is, in its own container, hosted on Saphan's
Cloudflare. **Sent** `58769394aff60381` and three follow-ups → the command-line master. **Taken** as W12 `proven-docs-site`.
The owner checked proven.ai (taken) and chose **proven.saphan.ai**.

### 5 · Questions from comparing with others

- **Does each agent see only the data its role allows?** **Sent** `228a225eda634a2c` → master core. It became W55
  `evidence-scope-per-role`.
- **What do the guardrail's checks show?** **Sent** `0b2436147f5fb83d`. **Answer:** every check, pass or block, is in the
  trace with its categories and latency, but the trace does not say whether the local or the hosted model checked. W54
  adds it.

### 6 · What is published about cost

The owner first wanted the daily spend in the README, then decided against it: the figures stay private, and public
documents carry the model that ran, not its cost. Proven's `AGENTS.md`, the decision template and the ledger were changed
the same evening. **Sent** to the five masters (`3d01d2b6586aa968` and four more): public copies carry the model and no
cost from 25 October.

### 7 · Master core splits

Master core had too much to finish before the freeze. The owner: *"Take half of it to a new master."* The advisor wrote
the kickoff of a new track, `proven-meterless-engines` (W50, W51, W54, OpenTelemetry, and core's share of the adapter
documentation). **Sent** `d0842a7e15678f81` → master core: what leaves, what stays.

**Answer** (16:15Z): none of the five had started, so all five moved whole. Core kept the evidence store work (W47, W49,
W52), more Zetaris sources (W43) and W55.

The new master's session was lost once and restarted. Its first W50 leg ran for one minute at 00:21Z (see hop 10).

### 8 · NVIDIA's model as a reviewer in the fleet

The owner: the fleet's NVIDIA reviews call NVIDIA's API directly, not through Nightshift; Nightshift stays where Proven's
own code uses it. **Sent** `8093857cd99ae27d`. The first live review by Nemotron finished at 15:49Z.

### 9 · Three model families, one commit

For a public side by side, the advisor asked two masters (`db67fb06c1713057`, `2a064f7e0c93a57e`) to add a Nemotron review
beside the OpenAI and Anthropic reviews on their next change that is not security work. The test-corpus master put it on
the held-out set. Just after midnight UTC:

| reviewer | verdict | what it found |
|---|---|---|
| OpenAI (Codex) | reject | one case's frozen answer key contradicts its own contract |
| Anthropic (Claude) | accept with BLOCKING conditions | the same key problem, and two labels that contradict the key |
| NVIDIA (Nemotron) | accept with conditions | one of those two labels, judged non-blocking; a test re-run live, failing exactly as the case says |

The advisor's first summary said Codex had found something on its own. It had not: Claude had raised the same key
problem. The README was corrected before it was pushed.

### 10 · Zero trust, caught in the act

At 00:21Z the first W50 leg tried to download Meterless's repository from GitHub. The fleet refused the connection
(`CONNECT tunnel failed, response 403`) and the run stopped. The owner: *"That refusal is a BEAUTIFUL show of zero
trust."* **Sent** `4830b8dbd2700224` → the new master: hand the engine's files to the run as one archive with its hash,
pinned to a commit; do not open the network.

### 11 · The review comment must read for a person

The owner ran `make up` on a fresh clone. The lab worked, but the review comment on the pull request was the raw text
format, one field per line. *"It MUST be formatted; it looks like crap now."* **Sent** `d47fa7116b53adf0` and
`c79d3a628fcc106d` → the command-line master: a short plain-language summary first; one finding per problem, so the same
dependency is not reported three times; the findings ordered by proof; plain names instead of internal ones; queries and
rows as real tables inside collapsed sections; and a "what to do" for each finding.

### 12 · Small things, same night

- The documentation index on `main` carried merge-conflict markers, and they had gone public in the morning push. The
  advisor fixed it (`177bba6`). The owner's pre-push check now refuses conflict markers anywhere in the tree.
- `docs/LAB.md` is linked right under `make up`.
- The Cursor coding sessions, promised in [ADR-0024](../../docs/adr/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md),
  were not published yet. **Sent** `2076a0ea0fa2ea19` → the infrastructure master. **Taken** as W15: excerpts beside
  each change, each tied to its original by a hash, with a blackout check that refuses a recording with anything left
  to black out.
- A read-only viewer for the CouchDB evidence store, with a link from each finding in the comment to its record.
  **Sent** `380c816d5dec5bfa` → master core.
- Two rulings for the security track went out (`1a146e162e5b21ec`); their subject stays private until the fixes are
  merged.
