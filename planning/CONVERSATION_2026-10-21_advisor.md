# CONVERSATION — the advisor, the day before the build, 21 October 2026

This page continues [`CONVERSATION_2026-10-20.md`](CONVERSATION_2026-10-20.md). The infrastructure master's conversation for the same days is on [`CONVERSATION_2026-10-21.md`](CONVERSATION_2026-10-21.md). For every hop it says who spoke, to whom, what the advisor reasoned and did, and which document came out of it. The owner's words are in English; where he wrote in Polish, they are given in the advisor's translation.

## Hop by hop

### 1 · owner → advisor: the test suite, on Cursor

> *"OK, now it is time to start coding the test suite. Maybe we will use Cursor executors for that? This will be a perfect test before real coding starts tomorrow."*

**What the advisor read.** The rule of this repository that project code starts on 22 October, and the owner's ruling of 19 October that the test corpus's case code is built before that and declared as prior art ([`../tests/README.md`](../tests/README.md)). The 21 case contracts are written; their code is not. The fleet's record showed that the two engine fixes Cursor needed had landed overnight but were not yet installed on the machines.

**The advisor's advice.** Yes, with an order: install the fixes, repeat the smoke turn, and only if Cursor runs with its shell working and without the owner opening its network, start with one case, then the rest. It asked one question: the rule says tests are written apart from the code, and tomorrow the same Cursor seats write proven.

### 2 · owner → advisor: warm them up, and find the bugs today

> *"[The train] is there already, waiting for its task. Exactly so: every run of ours has a fresh context. But since tomorrow they write the competition work, we have to warm them up a bit and find the bugs, because there certainly are some. It's a new adapter, so I'd rather catch them today than tomorrow and lose time."* (translated from Polish)

**What the advisor did.** It wrote [`../fleet/orders/ADVISOR_NOTE_cursor-shakeout_2026-10-21.md`](../fleet/orders/ADVISOR_NOTE_cursor-shakeout_2026-10-21.md) for the master of the test corpus: two checks before the first run, eleven things to exercise on purpose (Go and TypeScript, downloads, a host the policy does not allow, local against far, two runs at once, the return, the model pin, cost, a timeout, commit messages, and a test that must stay failing), and how to report what breaks. A failure today counts as a result.

### 3 · an error of the advisor's, and its repair

Writing this page, the advisor first saved it under the name the infrastructure master had already used for his own conversation of 19–21 October, and so replaced his page. It found this in the commit's line count before anything was pushed, put the master's page back exactly as he had written it, and moved its own hops here.

### 4 · owner → advisor: tell the scribe to reinstall

> *"Write to the scribe that he has to do the reinstall."* (translated from Polish)

> *"You can save in proven what such a message and its answer look like, with black-outs where needed."* (translated from Polish)

**What the advisor did.** It measured macm5 first: both fixes were already in its build. It could not read runner-cursor from its seat, said so in the message, and sent it through the fleet's message system. While the answer was pending, the owner made runner-cursor trusted; the advisor measured it, found the fixes there too, and told the scribe nothing needed redoing. The scribe's answer showed he had installed a still newer build on every machine. The message, its delivery note and the answer are in [`../fleet/messages/MESSAGE_reinstall-cursor-fixes-runner-cursor_2026-10-21.md`](../fleet/messages/MESSAGE_reinstall-cursor-fixes-runner-cursor_2026-10-21.md), with what they show.

Sending the message, the advisor also named the test corpus's track in it, and the fleet's message system took that as the advisor declaring itself a head of that track. The binding cannot be released. The advisor will forward anything for the test corpus's master that reaches it by mistake.

The advisor also found a handover from the scribe, sent to it on 19 October and left unread for two days. Two of its points had been done by others since; three were questions for the owner.

### 5 · owner → advisor: three answers

> *"The five cases stay. Prior-art declaration: we say that the test suite, with all the code it needs, is prior art. The executors' tasks stay visible: this is part of the whole thing, to show how the orders are passed to them, how they work and how their work is reviewed. If all private data is blacked out, that is fine and even required. We will also prepare the orders for coding in advance, so they are written in Markdown, and run them when the coding window starts. Git will prove that the running happened after the start."* (translated from Polish)

**The advisor's note on the last point.** A commit's date is written by whoever makes the commit, so on its own it proves nothing about when the work was done. Saphan Studio records each run's start time when the agent starts; the status report carries that run id and time, beside the commits. Together they show the work started after 22 October 00:00 UTC.

**What the advisor did.** [`../fleet/README.md`](../fleet/README.md) now says that the orders are written before the build and run inside it, what proves the timing, and that orders, reports and reviews are public with private data blacked out. It has a new row for `messages/`. The two answers about the test corpus went to the scribe, who raised them.

## Not run

- Whether the fixes are installed, and whether the test corpus's master has been appointed, were not yet in the fleet's record when this page was written. The owner says the master is appointed.
