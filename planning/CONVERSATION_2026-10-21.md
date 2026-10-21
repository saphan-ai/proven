# CONVERSATION — the day before the build, 21 October 2026

This page continues [`CONVERSATION_2026-10-20.md`](CONVERSATION_2026-10-20.md). For every hop it says who spoke, to whom, what the advisor reasoned and did, and which document came out of it. The owner's words are in English; where he wrote in Polish, they are given in the advisor's translation.

## Hop by hop

### 1 · owner → advisor: the test suite, on Cursor

> *"OK, now it is time to start coding the test suite. Maybe we will use Cursor executors for that? This will be a perfect test before real coding starts tomorrow."*

**What the advisor read.** The rule of this repository that project code starts on 22 October, and the owner's ruling of 19 October that the test corpus's case code is built before that and declared as prior art ([`../tests/README.md`](../tests/README.md)). The 21 case contracts are written; their code is not. The fleet's record showed that the two engine fixes Cursor needed had landed overnight but were not yet installed on the machines.

**The advisor's advice.** Yes, with an order: install the fixes, repeat the smoke turn, and only if Cursor runs with its shell working and without the owner opening its network, start with one case, then the rest. It asked one question: the rule says tests are written apart from the code, and tomorrow the same Cursor seats write proven.

### 2 · owner → advisor: warm them up, and find the bugs today

> *"[The train] is there already, waiting for its task. Exactly so: every run of ours has a fresh context. But since tomorrow they write the competition work, we have to warm them up a bit and find the bugs, because there certainly are some. It's a new adapter, so I'd rather catch them today than tomorrow and lose time."* (translated from Polish)

**What the advisor did.** It wrote [`../fleet/orders/ADVISOR_NOTE_cursor-shakeout_2026-10-21.md`](../fleet/orders/ADVISOR_NOTE_cursor-shakeout_2026-10-21.md) for the master of the test corpus: two checks before the first run, eleven things to exercise on purpose (Go and TypeScript, downloads, a host the policy does not allow, local against far, two runs at once, the return, the model pin, cost, a timeout, commit messages, and a test that must stay failing), and how to report what breaks. A failure today counts as a result.

## Not run

- Whether the fixes are installed, and whether the test corpus's master has been appointed, were not yet in the fleet's record when this page was written. The owner says the master is appointed.
