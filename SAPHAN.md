# Saphan Studio, and how our fleet works

**Saphan Studio** (https://saphan.ai) is our control plane for fleets of AI coding agents. It decides which agent may do what, runs each agent inside our own boundary, and keeps a record of every order, report, review and decision. Saphan Studio is a pre-existing component of this entry. It runs the fleet, it is not judged, and its code is not in this repository.

This page explains the platform and the vocabulary you will meet in this repository: in [`planning/answers/`](planning/answers/), in [`fleet/`](fleet/) and in the commit history. It says what the platform does. It does not say how it works inside.

## Who does what

| role | who | what it does |
|---|---|---|
| **owner** | the one person on the team | sets the goals, decides, merges. Writes no code. |
| **master** | an orchestrating agent | turns a goal into a train of small changes, writes the orders, checks each round itself, and hands the owner a decision |
| **executor** | an agent that builds | builds one change in its own copy of the repository and reports back. In this project every executor is a Cursor agent. |
| **test author** | an agent | writes tests from the test plan, apart from the code, so the code cannot shape its own tests |
| **reviewer** | an agent that did not write the change | checks the change from a fresh context, breaks it on purpose, and returns a verdict with conditions |
| **tester** | an agent | uses the whole result the way a user would, against what each change promised |
| **scribe** | an agent | keeps the record: the bug backlog, the counts, the history |
| **advisor** | an agent | works with the owner on plans and documents |

## What the platform guarantees today

- **Every agent has its own identity**, separate from any person's account.
- **Agents never inherit the access of the person who started them.** Each task gets only the permissions it needs: a task that reads starts as a reader, a task that edits gets edit, and nobody gets delete just in case.
- **Network access is closed by default.** A task that needs a package registry gets it for that task only. The next task starts with nothing.
- **Our sandbox wraps the vendor's sandbox.** If the vendor's own sandbox is switched off, the agent is still inside ours.
- **An agent commits only in its own copy of the repository.** Merge and push are forbidden in its instructions and also blocked by the tool.
- **Every order is signed by the person who gave it**, so every task is linked to the person who asked for it.
- **Agents can say no. Only a person can say yes.** A reviewer's `accept` is advice. The decision is the owner's, from a short list: accept, accept with conditions, reject, close without shipping.
- **Every change keeps its record:** the order, who signed it, the agent's report, the review, the decision with its reason, the commits, and the name of the person who said yes.
- **Every rejection is written down with its reason**, and often it becomes the next rule the agents work by.

## Security, and where it runs

Saphan Studio is built for teams that cannot let agents loose. This is what the fleet behind this repository runs on every day.

| | what it means | docs |
|---|---|---|
| **Many vendors, one set of rules** | Claude Code, Codex, Qwen Code and Cursor agents, and local models, run under the same identities, policies and record. A vendor is a backend, not a different set of rules. | [the agents it drives](https://docs.saphan.ai/docs/getting-started/the-agents-it-drives) |
| **Closed by default** | A run starts with no network and nothing it was not given. Every permission is opened for one task and closes with it. | [egress](https://docs.saphan.ai/docs/security/egress) |
| **Our own sandbox** | Every run is confined by our sandbox, which wraps the vendor's own. Switch the vendor's sandbox off and the agent is still inside ours. | [sandbox matrix](https://docs.saphan.ai/docs/security/sandbox-matrix) |
| **macOS, Linux, containers** | Agents run on macOS and Linux machines, locally or on other machines of the fleet, and in Docker or Podman containers. | [runners](https://docs.saphan.ai/docs/security/8-runners-what-happens-during-a-run) |
| **The policy travels with the task** | Each task carries its own policy: which tools, which files it may read and write, where it may connect. Without its policy, a run does not start. | [agent security](https://docs.saphan.ai/docs/security/agent-security) |
| **Guardrails** | A checkpoint on the conversation between an agent and its model, with a plugin for NVIDIA NeMo Guardrails. | [guardrails](https://docs.saphan.ai/docs/security/guardrails) |
| **Every headless session recorded** | Agents run without a person watching, so every session is recorded: what the agent was told, what it did, what it returned. | [session recording](https://docs.saphan.ai/docs/security/session-recording) |
| **Secrets referenced, never stored** | An agent gets a reference to a credential at the moment of use. The value is never written into an order, a log or the record. | [secrets](https://docs.saphan.ai/docs/security/9-secrets) |
| **Evidence, not claims** | Every order, report, review and decision is kept with who signed it and when, so any change can be traced back to the person who asked for it and the person who said yes. | [audit and compliance](https://docs.saphan.ai/docs/audit-and-compliance) |
| **CLI and web console** | Everything the fleet does can be driven and read from the command line or from the web console. The console is the owner's view of the record. | [the console](https://docs.saphan.ai/docs/console) |

The documentation also says, chapter by chapter, what is built, on which platform it has been exercised, and what this release does not cover: [documented limits](https://docs.saphan.ai/docs/security/11-documented-limits).

**Everything about the platform is at [docs.saphan.ai](https://docs.saphan.ai/docs).**

## The words you will meet

| word | meaning |
|---|---|
| **train** | one goal, delivered as a series of small changes |
| **wagon** | one change in a train, numbered `W1`, `W2`, `W2b` … |
| **track** | the line of work a train belongs to, named by a slug |
| **stream** | the work on one wagon, named by a slug such as `the-cursor-egress-baseline` |
| **leg** | one agent run on a stream: an executor leg, a review leg, a tester leg |
| **round** | one pass of build, check and decide. A change sent back gets a new round, and the old rounds stay in the record. |
| **seat** | a place where an agent runs: one vendor's agent on one machine |
| **local / far** | on the owner's machine / on another machine of the fleet |
| **master verify** | the master's own check of a round, before the review |
| **named review** | the independent review of the last round, with its verdict |
| **conditions** | what a review asks for, each one `BLOCKING` (cannot land), `OWED` (lands, and the fix is owed) or `NOTE` |
| **Unrun** | the section of a review that lists everything the reviewer could not run, and why |
| **unplug** | removing a check on purpose to show that a test notices. If nothing turns red, the check is not protected. |
| **base / tip** | the commit a change started from / the newest commit of the change |
| **trunk** | the main branch (`develop` in our own repositories) |
| **census test** | a test that reads the whole code base and counts it against one rule, so a change anywhere that breaks the rule turns it red |
| **land** | merge into the trunk. Only the owner does it. |
| **the record** | every document of every round: orders, reports, reviews, decisions, measurements, bugs |

## The laws

The fleet works under a set of numbered laws, more than fifty of them. Every agent starts every run with the laws in its instructions. A law usually exists because something went wrong once, and it says what must happen so it does not happen again. The ones quoted in this repository:

| law | in plain words |
|---|---|
| **#41 unplug test** | A mechanism is done only when you can switch it off and show the alarm that fires. If nothing turns red when it is unplugged, only its happy path exists. |
| **#44 an order must be able to refute its author** | What a master writes in an order is a claim to be tested, not settled ground. Every order includes a step that, done honestly, proves the master wrong if the master is wrong: reproduce the problem before you change anything, and if you cannot, say so. |
| **#53 a review with Unrun items is not an accept** | If a review lists anything it could not run, its verdict cannot be a plain `accept`. A condition's weight can be raised by the master, never lowered. |
| **#55 a refusal describes itself** | When the system refuses an agent or a person, its first sentence says what was measured, what was refused, and what to do next. |

Some rules are the owner's own rulings rather than numbered laws. One you will meet in the answers: **"a reject from either is a reject."** When two reviewers disagree and no evidence settles it, the reject stands.

## Scale

The fleet has written millions of lines of production code across our projects. Saphan Studio alone, one of those projects, had about 1.1 million lines of code, 1.7 million lines of tests and 29,570 commits between 6 July and 7 October 2026. The owner has not written a line of its code since 6 July.

## Why this repository is different

For Saphan Studio the record is private. For this project it is public. The same fleet, under the same laws, writes its orders, reports, reviews and decisions for `proven`, and a copy of each one is published in [`fleet/`](fleet/), next to the code it explains.
