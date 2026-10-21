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

### 6 · owner → advisor: the organisers' guides

The owner shared the event's Welcome Guide, the Meterless engine guide and index, and the organisers' 30-page Participant Setup Guide, and asked the advisor to read what it could reach itself.

**What the advisor found.**

- **Meterless engines** are reference architectures each team implements in its own language, with no service and no key: what [ADR-0003](../fleet/decisions/ADR-0003-meterless-engines-in-go.md) already decided.
- **A Power rule:** *"your app must support Meterless Nightshift as a primary AI provider."* Nightshift is Meterless's application that serves model requests on a machine; NVIDIA's models are reached through it.
- **Eligibility:** *"Zetaris and Meterless must be running on your machine to be eligible."* Our own Zetaris platform is required, not optional.
- **Zetaris cannot read a file sitting on a disk**: data comes from storage it reaches, or from a web API registered as a table. That shapes how the evidence sources are loaded; the choice is open.
- The setup guide's facts about the local Zetaris platform went to the infrastructure master for his order.

> *"ADRs: yes."* (translated from Polish)

**What the advisor did.** It wrote [ADR-0018](../fleet/decisions/ADR-0018-meterless-engines-minimum-and-their-contract.md) (two engines at their minimum, one injected generator, JSONL in `logs/`, Markovian only if time is left) and [ADR-0019](../fleet/decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) (Nightshift as the primary provider, NVIDIA through it by default, other routes kept as settings). `.env` files are now in `.gitignore`, as the event's security rule asks.

### 7 · owner → advisor: macOS refuses the installer

> *"Apple could not verify 'Nightshift Setup' is free of malware that may harm your Mac or compromise your privacy. Nightshift's website looks nice, but then it's weak."* (translated in part from Polish)

**What the advisor checked**, without running anything: the disk image is not signed; the app inside carries only an ad-hoc signature, with no developer identity, and is not notarized by Apple. A note inside the image from Meterless says it is *"not notarized by Apple yet"* and tells users how to open it anyway. The download was also the edition for M1 to M4 Macs, while the guide sends M5 Macs to another one.

### 8 · owner → advisor: the checksums, and a virtual machine

The M5 edition was refused the same way. Meterless publishes SHA-256 sums, and the owner found the list for the M5+ edition. The advisor compared them: the download matches. It said what that shows (the file is the one Meterless published) and what it does not (what the file does, or who built it), and proposed three places to run it: a macOS virtual machine on macm5, a separate macOS user, or macm5 itself.

> *"A"*

**What the advisor did.** It wrote the order [`../fleet/orders/TRAIN_ORDER_proven-nightshift-vm.md`](../fleet/orders/TRAIN_ORDER_proven-nightshift-vm.md): a VM with no shared folder or clipboard, the installer's checksum checked again inside it, NVIDIA configured only inside it, two lanes pinned to Nemotron models, the endpoint reachable over the private network with its own token, one request that must name a Nemotron model, and a runbook.

### 9 · owner → advisor: three answers before onboarding

The advisor read the event's help-desk channel and the sponsors' guides that differ from each other, and left three questions open for the organisers. The owner settled them (translated from Polish):

> *"Pre-existing components are fine: they have confirmed it."*

> *"If they don't give us their credits, we use our own. We'll manage."*

> *"There is no other way. Security is security and zero trust is zero trust. It's poor that they give an unsigned binary to the public to install."*

**What follows.** The test corpus and Saphan Studio stay declared as pre-existing, in the same track. Cursor and NVIDIA run on our own subscriptions and keys unless the sponsors' credits arrive, and the ledger records which. Nightshift stays inside its virtual machine, whatever the answer about a signed build.

### 10 · owner → advisor: a `metadata.json` for everyone

Participants report that the sponsor folders ask for each sponsor's use to be described "in your README and metadata.json", while no submission document defines that file. Nobody had answered them. The owner (translated from Polish):

> *"Let's draft the format and share it with the other participants. Placement in the root folder. Let's help ourselves."*

**What the advisor did.** It wrote a community draft, [`../metadata.schema.json`](../metadata.schema.json), a JSON Schema anyone may copy, and proven's own [`../metadata.json`](../metadata.json), which follows it. The schema asks what judges need: the track, the team, each sponsor's use with its status (planned, in progress, working), where it is in the code and the evidence that it works, the components that existed before the build, and how to run it without the team's keys. It says plainly that it is not official and that an official format, if one comes, wins. proven's file marks most of its sponsor use as planned, because the build has not started. Validated with `ajv` (JSON Schema 2020-12): proven's file passes; a copy with an unknown track and a copy with an unknown engine are refused.

## Not run

- Whether the fixes are installed, and whether the test corpus's master has been appointed, were not yet in the fleet's record when this page was written. The owner says the master is appointed.
