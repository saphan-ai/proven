# Saphan Studio, and how this fleet uses it

## Saphan Studio and Proven

**[Saphan Studio](https://saphan.ai/)** is a product that runs fleets of AI coding agents under human authority. It decides what each agent run may do, runs it within limits, and keeps the record of every request, report, review and decision. Its creator is [Marcin "Wato" Marzec](https://wato.org/).

**Proven** is a separate, new, open-source project: the code reviewer this fleet builds at the Open Agent Hackathon 2026. Saphan Studio is a pre-existing component of that entry. It runs the fleet that builds Proven; it is not part of the submission, it is not judged, and its code is not in this repository.

This page has three parts: what Saphan Studio does today, how this particular fleet uses it, and the words you will meet in this repository. The technical reference for what the product does, and what it does not do yet, is **[docs.saphan.ai](https://docs.saphan.ai/docs)**. Where this page and the documentation differ, the documentation is the reference.

## What Saphan Studio does today

These are product capabilities as the documentation describes them on 19 October 2026, each with its page and its limits.

| | what it does | docs |
|---|---|---|
| **Many vendors, one record** | Runs Claude Code, Codex, Cursor, Qwen Code and Droid agents, its own agent loop, and local models through an OpenAI-compatible connection, under the same identities, policies and record. Each backend's page says on which systems it has been exercised. **Cursor**, which Proven's executors use, has been added. | [the agents it drives](https://docs.saphan.ai/docs/getting-started/the-agents-it-drives) |
| **Limited network** | A run reaches its own model and the package registries for its kind of work; other destinations are refused unless a policy opens them, and opening the network is recorded with a reason. A machine that cannot enforce this refuses the run instead of running it open. Some kinds of run cannot carry this yet and are refused. | [egress](https://docs.saphan.ai/docs/security/egress) |
| **Confinement** | Where the platform confines a run, it limits what the run may write on the machine. This covers some agent and operating-system combinations, not all; where it does not, the record says that no confinement is claimed. Confinement limits writes, not reads. | [sandbox matrix](https://docs.saphan.ai/docs/security/sandbox-matrix) |
| **Human approval** | Approval is a person's act by default. Automation is possible only through an explicit, time-limited delegation that a person signs. The platform never merges code itself; merging stays a human act in your own Git tooling. | [agent security](https://docs.saphan.ai/docs/security/agent-security) |
| **A record of everything** | Every request, report, review and decision is kept with who made it and when, so a change can be traced to the person who asked for it and the person who approved it. | [audit and compliance](https://docs.saphan.ai/docs/audit-and-compliance) |
| **Session recording** | Agents run without a person watching, so their sessions are recorded by default. A run that is not recorded says so, with the reason. The exchange with the model is recorded only for agents the guardrail covers; for the others there is a run record and output, but no model transcript. | [session recording](https://docs.saphan.ai/docs/security/session-recording) |
| **Secrets by reference** | An agent gets a reference to a credential when it needs it; the value is not written into a request, a log or the record. | [secrets](https://docs.saphan.ai/docs/security/9-secrets) |
| **Guardrails** | A checkpoint on the conversation between an agent and its model, with a plugin for NVIDIA NeMo Guardrails. | [guardrails](https://docs.saphan.ai/docs/security/guardrails) |
| **Command line and web console** | Everything can be driven and read from the command line or the web console. | [the console](https://docs.saphan.ai/docs/console) |
| **Forges without webhooks** | Pull requests on a Gitea forge are read by polling, so the forge needs no public endpoint pointing back at the platform. | [pull requests and forges](https://docs.saphan.ai/docs/working-with-git/pull-requests-and-forges) |
| **Read-only MCP server** | An assistant or an agent can read the record over MCP, so nobody has to relay fleet state by hand. It answers questions and takes no action. | [MCP server](https://docs.saphan.ai/docs/mcp-server) |

What the product does not cover yet is listed chapter by chapter in the **[documented limits](https://docs.saphan.ai/docs/security/11-documented-limits)**.

## How this fleet uses it

Saphan Studio is the product. What follows is how *this* fleet is set up to build Proven: our own operating policy, stricter in places than the product's defaults.

- **The owner approves every merge himself.** No delegation is used for merging.
- **Agents never push.** They commit in their own copy of the repository; the owner pushes once a day.
- **The author of a change never reviews it.** Every change is checked by an agent that did not write it, starting from a fresh context.
- **Every document is published.** For Saphan Studio the record is private. For Proven, a copy of every request, report, review and decision is published in [`fleet/`](fleet/), next to the code it explains.

### Who does what

| role | who | what it does |
|---|---|---|
| **owner** | Marcin, the one person on the team | sets the goals, rules on scope, settles what the evidence cannot settle, approves every merge |
| **master** | an agent that coordinates | turns a goal into small changes, writes each request, checks each round, hands the owner a decision |
| **executor** | an agent that builds | builds one change in its own copy of the repository and reports back. For Proven, every executor is a Cursor agent. |
| **test author** | an agent | writes tests from the test plan, separately from the code |
| **reviewer** | an agent that did not write the change | checks the change and returns a verdict with conditions |
| **tester** | an agent | uses the result the way a user would |
| **scribe** | an agent | keeps the record: the bug backlog, the counts, the history |
| **advisor** | an agent | works with the owner on plans and documents |

### One change, end to end

1. The owner asks for something. A **master** writes the request (an *order*): what to build, how a reviewer will check it, and what is out of scope.
2. An **executor** builds it in its own copy of the repository and writes a *status report*: what it built, what it ran, what failed.
3. A **reviewer** that did not write the change checks it and writes a *review*. It can say no; it cannot say yes.
4. The **owner** reads the review and decides: accept, accept with conditions, reject, or close without shipping. His reason is written down.
5. If the change goes back, it gets a new round. The documents of every round stay.

### The hardware this fleet runs on

Everything runs on premises, in one person's home office. No cloud: only the model vendors' APIs are reached over the
internet.

| machine | specification | what runs on it |
|---|---|---|
| **Apple MacBook Pro** (macm5) | Apple M5 Max, 18 cores, 128 GB unified memory, 4 TB SSD | Saphan Studio itself: the engine, the record and the console (one app served from one binary); the masters' sessions; Meterless Nightshift in its own macOS virtual machine |
| **Apple Mac mini** (macmini) | Apple M4, 10 cores, 16 GB, 1 TB external SSD | agent runs and reviews that need Apple's platform |
| **AMD Ryzen AI Max+ 395 mini PC** | 32 threads, 128 GB memory shared with the integrated Radeon GPU, 2 TB NVMe | local inference on the AMD GPU: NVIDIA's Nemotron Content Safety (Proven's guardrail), three more guard models, and local coding and reasoning models |
| **AMD Ryzen 9 9950X workstation** | 16 cores / 32 threads, 128 GB, 2 × 2 TB NVMe on ZFS, Proxmox | every Linux virtual machine of the fleet: the agent runners (Cursor, Claude Code, Codex, NVIDIA Nemotron, Qwen), the builders, Zetaris in its own virtual machine, and the judges' lab Gitea |
| **three Intel NUCs** | Proxmox cluster | observability (Prometheus, Grafana), the model router (LiteLLM), network routing |

In all: about 436 GB of memory and 11 TB of SSD, most of it in three machines.

## Skills, rules and orchestration

**Skills.** Every agent run gets the skills its task needs, from a library of 32, in the open Agent Skills format. A skill brings the right instrument for a job: which tool, which check, which command. Each project switches on its own set; that is the owner's decision.

| area | skills |
|---|---|
| languages | Go, TypeScript and Node, Python, Rust, .NET, JVM and Spring, Kotlin for Android, Swift for iOS |
| front end | front-end engineering, React, accessibility (WCAG), web vitals, browser testing, terminal UIs, native mobile UI |
| back end and data | HTTP API conventions, contract-first boundaries, Postgres, schema migrations, idempotent writes, MCP servers |
| security | web security hardening, privacy and personal data, supply chain for Go, npm and Python |
| delivery | the testing bar, feature flags and rollout, semantic versioning, deprecation, observability, decision records |

Several skills build on open-source work and credit it in their own files: Addy Osmani's agent-skills (MIT) and Anthropic's skills (Apache-2.0).

**Rules.** The fleet works under written rules that the owner ratifies. Most exist because something went wrong once. In broad groups:

- **Human authority and ownership.** A person sets the goals and approves what ships. An agent can recommend and refuse; it cannot approve.
- **Evidence and verification.** A claim counts when a run shows it. What could not be checked is reported as not checked.
- **Security and isolation.** Each run gets only what its task needs, and secrets are never written down.
- **Accountability and traceability.** Every request, result and decision is recorded with who made it.

**Orchestration.** The work moves through four functions: **planning and coordination** (masters), **implementation** (executors and test authors), **independent verification** (reviewers and testers), and **record keeping and reporting** (the scribe, and every document in [`fleet/`](fleet/)).

## The words you will meet

| word | meaning |
|---|---|
| **order** | the request a master writes for one change |
| **status report** | the executor's account of what it built and ran |
| **review** | an independent check of the change, with a verdict |
| **decision** | the owner's ruling on a change, with his reason |
| **round** | one pass of build, check and decide; a change sent back gets a new round |
| **conditions** | what a review asks for, each one **BLOCKING** (cannot be merged), **OWED** (can be merged, the fix is owed) or **NOTE** |
| **not run** | the part of a report or review that lists what could not be checked, and why |
| **train / wagon** | one goal delivered as a series of small changes / one of those changes |
| **seat** | a place where an agent runs: one vendor's agent on one machine |
| **land** | merge into the main branch. Only the owner does it. |

## Scale

Saphan Studio, one of the projects this fleet works on, had about 1.1 million lines of code and 1.7 million lines of tests, and 29,570 commits were made to it between 6 July and 7 October 2026. The owner has not written a line of its code since 6 July. These figures were first published on this page on 16 October 2026; the counting method is not published here, and line and commit counts measure size, not quality.
