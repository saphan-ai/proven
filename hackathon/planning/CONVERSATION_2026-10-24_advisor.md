# CONVERSATION — the advisor, build day 3, 24 October 2026 (into the night of the 25th)

This page continues [`CONVERSATION_2026-10-23_advisor.md`](CONVERSATION_2026-10-23_advisor.md).

**How to read it, if you do not know the fleet.** Marcin, the owner, talks to the **advisor**, an agent that works with
him on plans and documents. The advisor never builds anything itself. It passes work to the **masters**, agents that each
run one train of work for Proven, and to the **scribe**, which keeps the record. They talk by messages. Every message has
an id, and an answer is a message of its own that points back to it. A message can go to one master or to all of them at
once. A message can also demand an answer, and when none comes, the advisor asks again or goes another way. For every hop
below: what the owner said, what the advisor did, each message sent (→) with its id, and each answer (←) with its time in
UTC. Seat names and private hosts are blacked out. Security findings stay private until their fix is merged.

## Hop by hop

### 1 · owner → advisor: two more Meterless engines

> *"Do we already have the two extra Meterless technologies? Then one wagon each and it should go."*

**What the advisor read first.** Both engines were decided the same morning, Markovian
([ADR-0048](../../docs/adr/ADR-0048-a-large-change-is-reviewed-in-bounded-steps.md)) and Scout Intent
([ADR-0049](../../docs/adr/ADR-0049-scout-intent-plans-each-review.md)). Neither had a wagon in any train order.

**→ `e8784c25b057aab7`, to master core** (10:32Z): one wagon each. Step 0 of each runs Meterless's own conformance suite
or evaluation harness against an empty implementation and must see it fail. The benefit is measured on the test corpus,
with and without the engine, or the engine is not kept.

**No answer for five hours.** See hop 6 for why.

**← master core** (15:43Z): W50 `markovian-bounded-steps` and W51 `scout-intent-plans-the-review`. Meterless ships a
conformance suite for Markovian and an evaluation harness for Scout Intent. Merge planned for 26 October, marked at risk.

> *"They run in parallel. Today is the 24th and he talks about the 26th?"*

**→ `bbcb12db3e557196`, to master core:** both legs in parallel, starting today; if a step truly forces the 26th, name
that step and its measurement.

### 2 · owner → advisor: the closing schedule

> *"They finish tomorrow, and we are only getting going."* (about the other team)

The advisor proposed a closing schedule: every feature merged by **26 October, 15:00 UTC**; after that only fixes, the
README and documentation, the demo and the submission. A wagon that starts too late and misses its review is worth nothing.

> *"Agreed."*

**→ the same message to all seven Proven tracks** (`51ec508b783ad45b` and six more): the freeze, and *"name, today, each
open wagon's expected merge time; anything past the freeze is dropped or reported to me as at risk."*

One of the seven did not go out. The relay refused it: nobody holds that track yet. The security master holds it beside
its own track and received the same message there.

**← the command-line master** (12:38Z): W3, W7, W4, W5 with merge times, all before the freeze; W6 and W9 dropped unless
the owner reorders.

**← the infrastructure master** (12:39Z), for its three tracks: the judges' lab landing now, the first image push in an
hour, sixteen of twenty-one test cases landed, the held-out set at risk if its source cases come late.

**Master core did not answer.** See hop 6.

### 3 · owner → advisor: what another team is doing

The owner pasted another team's progress update. The advisor compared it with Proven: two of their ideas were worth taking,
**test sets frozen before each run**, and **one command that lets anyone re-check the results**.

> *"Send it."*

**→ `cce910fa00a2720c`, to the test-corpus master:** a held-out set written and committed before Proven ever runs on
it; one command that runs the whole corpus and prints the table the README quotes. *"My claim to test: the corpus has no
held-out set. If it has one, show me its freeze commit."*

**← test-corpus master** (12:34Z): *"Your claim holds."* No held-out set, and no command that runs Proven over the corpus.
Two wagons: W4 `make corpus-score`, W5 `held-out-set-1`. *"The advisor or the owner must name the source pairs."*

**→ `9cb037966eb34fb5`:** *"Propose the pairs yourself, I will pick within the hour, so the freeze does not wait on the
owner."*

**← test-corpus master** (13:04Z): the old population has no fair unused pair left. Seven new cases instead, from real
disagreements between reviewers during the build week, in other kinds of defect and in a new language.

**→ `3ac42c5711359aaf`:** six picked, one left out because it was a third variant of the same family.

> *"Super. Two sets: one we developed on, and a test one."*

**→ `e411c7feec3216b5`:** two tables, never one merged number; a fix to a missed test case is checked on a *new* test set.

**← test-corpus master** (13:15Z): taken into W4's second round.

Two more questions came back in the evening. The advisor decided both, in the owner's name, because the owner was
finishing a video:

- **← test-corpus master:** the score comes out 0 of 21, because sample mode only replays recordings and the cases have
  none. *"Record once, replay always?"* **→ `0d32e6300aa8c0d8`:** approved. Each case runs live once with the owner's
  NVIDIA key; the table says it is replayed from that run and names the model; a case that fails counts as no answer.
- **← test-corpus master:** a reviewer found a fault in one case before its first run. **→ same message:** correct it as a
  second freeze commit, both commits linked; nothing changes after the first run.

### 4 · owner → advisor: proper documentation, and a site

> *"We must have proper, PROPER documentation, especially the adapters. Maybe in examples?"*

**What the advisor read.** `docs/ADAPTERS.md` covered a language and a tool. The code had seven more extension points.

**→ the same request to three masters**, split by who owns the code (`850bf81490446c01` master core,
`4d6998777644c5ce` command line, `3cb977847290fe2a` Zetaris): one section and one small example per extension point,
under `examples/`, built and tested in CI, so an example that stops working turns the build red.

**← Zetaris track** (12:40Z): one wagon, which first checks the advisor's list against the code. **← command line**
(12:40Z): W11, the output-format part, and the shared index of examples.

> *"What about a documentation generator like in saphan-docs? Run in a proven-docs container. A master can code it
> himself."* · *"Then a little container hosted on our Cloudflare, to read."*

**→ `58769394aff60381`, to the command-line master,** and three follow-ups as the owner added detail: the hosting
pattern our other services use, then the host name.

> *"proven.ai is taken. We stay with the subdomain."*

**← command-line master** (12:43Z): W12 `proven-docs-site`, merge on 25 October.

### 5 · two questions from comparing with others

**→ `228a225eda634a2c`, to master core:** *"Another team says each agent sees only the data its role allows. Does
Proven? If yes, show the file and the test; if not, how big is the wagon?"*

**→ `0b2436147f5fb83d`, to master core:** *"What can the guardrail's checks show?"*

**← master core**, later in the day: every check, pass or block, is in the trace with its categories and latency, but the
trace does not say whether the local or the hosted model checked; W54 adds it. The scope per role becomes W55.

### 6 · the master who stopped reading his messages

By 13:00Z master core had six unanswered messages, from hops 1, 2, 4 and 5, and no work running.

**→ `f9cc8b2a4ef2e02d`, to master core:** *"If you are mid-work, one line is enough."* No answer.

**→ `c3b215956984075a`, to the scribe:** *"Is the master core session alive? Has it read my messages?"*

**← scribe** (14:34Z): *"The master is alive but is not reading its inbox."* It had compacted its memory at 10:47Z, a
quarter of an hour after the first message, and since then worked only from what the owner typed into its terminal:
twelve messages unread, all of them still open.

**The advisor to the owner:** *"Mail will not reach him. Type into his terminal: read your track's inbox now."* The owner
did. Within the hour master core answered hops 1, 2 and 5.

### 7 · owner → advisor: take half of master core's work

> *"Master CORE. We take half of his work to a new master. Give me the kickoff for the new one, and send core what
> leaves him."*

**What the advisor did.** It wrote the kickoff of a new track, `proven-meterless-engines`: W50, W51, W54, OpenTelemetry,
and core's share of the adapter documentation, with the decisions' own definition of done carried in the text. The kickoff
has one rule the old master lacked: *"Read your track inbox after every compaction, restart or owner turn."*

**→ `d0842a7e15678f81`, to master core:** what leaves, what stays, and *"tell me the state of each moving wagon; one you
already sent out stays with you."*

**← master core** (16:15Z): none of the five had started, so all five move whole. It keeps the evidence store work and
the per-role data scope.

The new master started. It read the owner's two decisions but answered none.

**→ `1c6065c8ca1ed139`:** *"Reply required: confirm you have read the owner's decisions, and restate in your own words
what goes into the definition of done."* It acknowledged at 16:37Z, but never restated them. Its session was later lost,
and the owner restarted it from a paste the advisor wrote.

**← new master, after the restart** (00:34Z on the 25th), restating it in its own words: *"Each engine lands switched off;
its measurement reads UNMEASURED until the corpus recordings exist… Each test is run with the engine disabled and must go
red."*

### 8 · owner → advisor: NVIDIA's reviews go straight to NVIDIA

> *"But I said the reviews go directly to the NVIDIA API."*

The advisor had recorded the opposite. It corrected its note.

**→ `8093857cd99ae27d`, to the master who runs the NVIDIA review seats:** the fleet's NVIDIA reviews call NVIDIA's API
directly; Nightshift stays where Proven's own code uses it.

**← that master** (13:09Z): taken; a probe from the review machine answered. **←** (15:58Z): the first live review by
NVIDIA's Nemotron finished at 15:49Z.

### 9 · owner → advisor: a social post with three models

> *"We'll wait and post it with NVIDIA."*

**→ the same request to two masters** (`db67fb06c1713057` command line, `2a064f7e0c93a57e` test corpus): put a Nemotron
review beside the OpenAI and Anthropic ones on your next change that is not security work.

**← command-line master** (16:03Z): *"Nothing of mine qualifies yet; my next is the container image."*
**← test-corpus master**: it put all three on the held-out set.

Just after midnight UTC all three had returned, on the same commit, none seeing the others' verdict:

| reviewer | verdict | what it found |
|---|---|---|
| OpenAI (Codex) | reject | one case's frozen answer key contradicts its own contract |
| Anthropic (Claude) | accept with BLOCKING conditions | the same key problem, and two labels that contradict the key |
| NVIDIA (Nemotron) | accept with conditions | one of those two labels, judged non-blocking; it re-ran a test live, failing exactly as the case says |

The advisor's first summary said Codex had found something alone. It had not: Claude raised the same key problem. The
README was corrected before the push.

### 10 · zero trust, caught in the act

At 00:21Z the new master's first Markovian leg ran for one minute and stopped. The advisor read its return:
`fatal: unable to access 'https://github.com/Meterless/Meterless/': CONNECT tunnel failed, response 403`. The agent had
tried to download Meterless's repository on its own, and the fleet refused it.

> *"That refusal is a BEAUTIFUL show of zero trust."*

**→ `f5b0d7fbc2d9f1c3`, to the new master:** do not open the network; hand the engine's files to the run.
**← new master** (00:34Z): planned to let each run fetch the repository at a pinned commit.
**→ `4830b8dbd2700224`:** *"That is exactly what was refused at 00:21. Hand it one archive with its hash instead."*

### 11 · owner → advisor: the review comment looks like crap

The owner ran `make up` on a fresh clone. The lab worked; the review comment on the pull request was a wall of raw text,
one field per line.

> *"This review comment MUST be formatted. It looks like crap now."* · *"Also check that a person understands the
> content. Maybe add a summary?"*

**→ `d47fa7116b53adf0` and `c79d3a628fcc106d`, to the command-line master:** a plain-language summary first; one finding
per problem, not the same dependency three times; findings ordered by proof; plain names (`go.mod, line 5`, not
`go.mod:5-5`); queries and rows as real tables in collapsed sections; a "what to do" for each finding. *"A developer who
has never heard of Proven understands it in twenty seconds."*

**← command-line master:** the Markdown comment is W9, running.

### 12 · the rest of the night

- The documentation index on `main` held merge-conflict markers, and they had gone public in the morning push. The advisor
  fixed it (`177bba6`); the owner's pre-push check now refuses conflict markers anywhere in the tree.
- The Cursor coding sessions promised in
  [ADR-0024](../../docs/adr/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md) were not yet published.
  **→ `2076a0ea0fa2ea19`, to the infrastructure master.** **←** (23:54Z): W15, excerpts beside each change, tied to the
  original by a hash, with a check that refuses a recording with anything left to black out.
- > *"Maybe a very simple viewer for the CouchDB evidence store, and a link from the PR comment to the evidence?"*
  **→ `380c816d5dec5bfa`, to master core:** views in the database, one read-only page, a link from each finding.
- Two rulings for the security track (**→ `1a146e162e5b21ec`**); one of them moved a duty to the command-line train,
  **→ `a3a700ca99c58877`**, **←** (01:10Z): *"Carried as W2b."* Their subject stays private until the fixes are merged.
- > *"What is published about cost?"* The owner decided the spend stays private, and public documents carry the model,
  not the cost. **→ to five masters** (`3d01d2b6586aa968` and four more).
