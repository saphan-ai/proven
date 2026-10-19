# CONVERSATION — a local guardrail model, 18 October 2026

This page continues [`CONVERSATION_2026-10-17.md`](CONVERSATION_2026-10-17.md): the owner and the advisor, preparing the infrastructure. Like the earlier pages, it says for every hop **who** spoke, **to whom**, what the advisor **reasoned** and **did**, and **which document** came out of it.

The owner wrote in English, so his words are quoted verbatim, typos included, with no translation. Internal names are blacked out (█, one block per character). The list is at the end.

## The people and agents in this conversation

- **The owner**: the one person on the team. He was flying home from Los Angeles, with a stop in Hong Kong.
- **The advisor**: the agent that works with the owner on plans and documents. It measures, recommends and writes things down. It does not dispatch work, gate or merge.
- **The infra master** (not yet appointed): the agent that will run the infrastructure trains.

## Hop by hop

### 1 · owner → advisor: a guardrail model for the fleet

The owner pasted a request and added a line of his own:

> *"We need to prepare our local LLM infrastructure for the NVIDIA model: Nemotron 3.5 Content Safety (4B), our local guardrail model.*
>
> *Write this as a parallel train order for the infra master.*
>
> *Target setup:*
> *1. The NVIDIA model is installed on █████ as a warm model. If it does not fit on that machine, take out the least used warm model and replace it with the NVIDIA one.*
> *2. The model is registered on the LiteLLM proxy.*
> *3. The agent-loop backend is tested with one test run on this model.*
>
> *If the master needs anything from me, the master stops and asks.*
>
> *The master runs this at the same time as the other work. Make it a separate train, or attach it as wagons that run in parallel.*
>
> *All clear?"*

> *"Hong Kong - they have banned traffic to Anthropic servers :)*
> *BKK now. let's keep planning and setting up the infra"*

A guardrail model judges whether a prompt, or an agent's answer, is safe. Our fleet already runs local guardrail models on one of its machines. The owner wants NVIDIA's model there too.

### 2 · advisor → the record, the machine and the model: measure first

Before writing anything, the advisor read three things.

- **Our own record.** On 5 October another master of our fleet installed three safety judges on that machine and measured them. Its analysis already listed Nemotron 3.5 Content Safety as an alternative kept for later. One question was left open: whether llama.cpp, the server that runs our local models, can render this model's prompt template.
- **The machine, read-only.** Which models are loaded, how much memory they use, how much disk is free. The models in memory take 32 of 112 GiB of GPU memory. NVIDIA's 4B model needs about 6 to 7 GiB, so it most likely fits without taking another model out. But the machine's swap was full, so the advisor wrote that the worst case has to be measured with the largest model actually loaded, not estimated. It could not read the server's configuration file, because its user has no access. It did not try another way in.
- **The model.** NVIDIA publishes it in a format our server does not load directly. A community conversion in the right format exists. The model card describes the answer format: `User Safety: safe|unsafe`, an optional `Response Safety:` line, and the categories.

**The advisor's reasoning.** Three points of the owner's request needed care:

- *"The least used warm model"* needs a reading of how often each model is used. Without one, any choice is a guess. And the least used model could be the one our coding agents depend on, which by our own contract is never taken out. Then two of the owner's rules collide, and only he can choose.
- *"Registered on the LiteLLM proxy"* and *"tested with one agent-loop run"* are not the same test. The agent loop reaches the machine directly, not through the proxy. So the run proves the proxy only if it goes through a route that uses the proxy.
- The model is a classifier, not a coding agent. One agent-loop run on it may end in a failure. That is a valid result, and the order says to report it as it is, not to change the task until it "passes".

### 3 · advisor → the infra master: the train order

The advisor wrote the train order `█████-nemotron-content-safety` into our private record: a separate train, run in parallel with `proven-cursor-infra`, by the same master. It has six wagons:

1. Re-measure the advisor's readings, read the live configuration, and decide whether the model fits by printed arithmetic.
2. Download the model and check its checksum against the one Hugging Face publishes.
3. Keep it loaded ("warm") next to the other judges, and check that it answers in the format of its card, with its latency.
4. Register it on the proxy, leaving the four existing models unchanged.
5. One agent-loop run on it, with an honest outcome.
6. Update the infrastructure records.

The advisor's readings go in as claims the master must test, not as facts. The order lists every place where the master stops and asks the owner, as the owner asked.

**Why there is no public copy of this order.** Our rule is that every hackathon document exists in both places. The advisor judged that this one is not a hackathon document. It is the fleet's own guardrail, and proven cannot use it: a judge's machine does not have our machine, so proven reaches NVIDIA through NVIDIA's hosted API. The advisor asked the owner whether he sees it differently. Until he answers, the order stays only in the private record. This page records the conversation either way.

### 4 · owner → advisor: English in this repository

While the advisor was working, the owner sent:

> *"Remember thay you write in proven in English"*

Everything the advisor writes here and in our record is in English. The advisor talks with the owner in Polish, because that is the language he set for our conversations.

### 5 · the owner corrected the advisor's redaction

On the plane the owner reviewed yesterday's changes. He took the advisor's commit back to edit it, and blacked out one word by hand in the public train order: the user name used to log in to a cluster node. The advisor had left it visible, reading it as a generic word that every such machine has. The owner's reading is the rule: user names are blacked out, generic ones included. The advisor did not undo his edit. It wrote the rule into its memory, and told the owner that the same word still stands in hop 8 of the previous page, already committed. Whether to black it out there is his decision.

### 6 · owner → advisor: keep the conversation pages up to date

> *"Remember to update conversations files"*

This page is the answer.

### 7 · owner → advisor: English, a black-out, and the guardrail goes public

> *"remember that I can spek Polish and you always in English.*
> *1 ████ - blackout*
> *2 make the order for █████ public - for the judges we will make option to run without guardrails. But we give guardrails as default as in saphan"*

The owner answered both of the advisor's questions, and set the language: he may write in Polish, and the advisor always answers in English.

**What the advisor did.**

- **The black-out.** In hop 8 of [`CONVERSATION_2026-10-17.md`](CONVERSATION_2026-10-17.md), the login user name is now blacked out in both places it stood, and listed in that page's `REDACTED` table.
- **The order is public now:** [`../fleet/orders/TRAIN_ORDER_nemotron-content-safety.md`](../fleet/orders/TRAIN_ORDER_nemotron-content-safety.md). The file name leaves out the machine's name. The original in our private record carries the owner's words and the name of its public copy. The reason given in hop 3 for keeping it private no longer holds: the guardrail is part of proven now.
- **The guardrail is part of proven's design.** By default proven checks prompts and answers with the guardrail model, as Saphan Studio does. A setting turns it off, so a judge can run proven without our machine. The top of [`../README.md`](../README.md) says so. Building the switch is proven's work from 22 October, not the infrastructure train's. One question is still open: whether NVIDIA's hosted API serves the same model. If it does, a judge with an NVIDIA key could keep the guardrail on.

This page covers hops 1 to 7 of 18 October.

## Not run

- proven's guardrail switch is not built. Whether NVIDIA's hosted API serves Nemotron 3.5 Content Safety was not checked.
- No model was downloaded, no configuration changed and no agent run started. Those are the master's acts, after the owner signs the order.
- The advisor's figures for memory and for the model's size are readings and estimates. The master re-measures them before acting.

---

## REDACTED

Parts of this file are blacked out (█). Nothing else in the quoted text was changed.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 3 | the machine that runs our local models (twice in the owner's words, once inside the train order's name) |
| account names | 1 | the account the owner asked to black out (hop 7) |
