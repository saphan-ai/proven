# Proven

**A code reviewer that has to show its evidence, built in public by one person and a fleet of AI coding agents.**

## Quick start

```bash
make up
```

**Step by step, with what you need first and what to do if something fails: [docs/LAB.md](docs/LAB.md).**

`make up` builds Proven and starts the judges' lab on this machine. It prints the lab address and a generated admin password, once. Open that address, then the open pull request: the sample review is a comment on it. `make down` stops the lab.

Proven will review pull requests: proposed changes to a codebase. Every problem it reports must come with evidence you can check yourself, such as a test you can run or a data query you can repeat. If it cannot show the evidence, it does not report the problem.

[Marcin "Wato" Marzec](https://wato.org/) used fleets of AI coding agents on commercial telco software, and built [Saphan Studio](https://saphan.ai/) to direct and check their work. Now he is taking that fleet to the [Open Agent Hackathon 2026](https://hackathon.genai.works/event/open-agent-hackathon-2026), organized by GenAI Works, from **22 to 27 October 2026**, in the track *The Agent That Can Explain Why*, to build Proven. The agents write the code. Marcin sets the direction, makes the decisions, reviews the outcomes and approves every merge.

**Coding started on 22 October 2026, at 00:00 UTC.** Proven's code is written only inside the build window. Everything before 22 October was preparation, and all of it is public.

<a href="https://vimeo.com/1236933110"><img src="assets/proven-loop.gif" alt="Proven: a pull-request reviewer that has to prove what it says. The opening slides of the introduction video, looping; click to watch it." width="100%"></a>

<p align="center"><a href="https://vimeo.com/1236933110"><b>▶ Watch the introduction</b></a>: the idea, the evidence, the team (6 min). An introduction, not a product demo.</p>

## Why this exists

Marcin has spent almost three decades in internet infrastructure and security, including products that run at mobile operators. When he began using fleets of AI coding agents on commercial telco software, the hard part was not getting code written. It was knowing what to trust: who asked for a change, who checked it, and on what evidence.

That work led to patent applications, filed in the U.S. and Europe and still pending, and to **[Saphan Studio](https://saphan.ai/)**: a product that runs fleets of coding agents under human authority and keeps the record of their work. Saphan Studio is itself built by one person working with an agent fleet. The agents develop the platform that governs their own work; Marcin decides what ships.

Now that fleet takes on something new. At the hackathon it will build **Proven**, an open-source tool for checking code of the kind it writes itself.

## Impact: the problem at its real size

This repository is not here to show off a fleet. It shows, from the inside, how a fully agentic coding factory works,
because that is the only way to show the size of the problem Proven solves.

Most solo developers who code with AI agents feel the problem only lightly: an agent says "done", the tests are green,
and now and then something is quietly wrong. At one or two agents you can still read every change yourself. At a fleet
you cannot, and the question stops being "can agents write the code?" and becomes **"which of this can I trust, and on
what evidence?"** That is why pull-request review and security are not extras here. **They are the foundation.** Without
them, a fleet that writes code faster only produces unchecked code faster.

So we do not ask you to imagine the scale. We show it:

- **One person.** Marcin decides, rules on disputes and merges every change by hand. No agent ever merges or pushes.
- **A coding fleet** that plans, builds, tests and reviews, with every change reviewed by a model family other than
  the one that wrote it.
- **Six days**, from 22 to 27 October 2026, to build a working evidence-based reviewer with its own lab, a GitHub
  Action, four Meterless engines, live data through Zetaris, NVIDIA Nemotron as a reviewer, and a security review of
  its own code on top.
- **Every step in the open**: each order, each builder's report, each review and each decision is in this repository.
- **And Proven is not the only thing the fleet builds this week.** In the same six days, under the same one person,
  it keeps building three other projects: Saphan Studio itself (the platform that runs the fleet), the
  [Quake Watch server](https://github.com/saphan-ai/quake-watch-server), an earthquake alert service built in public
  as a step-by-step example, and its Android app. Four projects at once, one person deciding and merging all of them.

By the end of day four, 100 changes had been merged. Counted from the fleet's own record on 25 October, 240
independent reviews raised **201 blocking problems** in work that an agent had called finished, and **116 of them were
stopped before the change reached the main branch**. That is the problem Proven exists for, measured on the project
that builds it, not estimated.

**And do not take our word for the quality of the code: judge it yourself.** Every line is public, with the order
that asked for it, the review that checked it and the decision that merged it. The fleet is not ashamed of its work. It is proud of it.
Open any package, read any test, pick any merge, and tell us what you find.

## Honest limits

What Proven does not do yet, or does not do well, measured and dated. This list is updated with every push.

- **Scout Intent: how well it sorts changes is measured; whether it improves reviews is not.** On a blind set of 46
  changes frozen before the run (English, Polish and Thai, including disguised ones) it got about two thirds right;
  the exact count is being re-checked after a review found a flaw in how we counted, so we do not quote a number until
  it is confirmed. In some misses a security change was taken for a documentation change and got a lighter review than
  it should have (26 October). It also asks for clarification more often than its own ceiling allows. Whether its plans
  give the same findings for less work is not measured yet.
- **Markovian: its effect on reviews is not measured yet.** It passes its conformance checks (8 of 8), but whether
  reviewing a large change in bounded steps finds as much as one whole review is still to be measured.
- **Scout Intent and Markovian are off by default today.** Turning all four engines on by default, each switchable
  off by a setting or when `make up` asks, is being built.
- **H-MEM reads its whole state file** when it opens and rewrites it on every write. A size cap is being added; until
  then it is meant for one repository's memory, not a large organisation's.
- **The corpus score was recorded while NVIDIA's API was failing** (HTTP 500 and timeouts on 26 October), so the NVIDIA
  column shows mostly failed calls, every failure named. A second column on another model family (GLM-5.3) is recorded
  once and labelled "run 1 only".
- **The lab's demo runs on sample data.** Live mode needs Zetaris and Meterless Nightshift, your own or ours through an
  authenticated gateway.
- **The review spine is one long function** (`pipeline.review`, about 500 lines). Splitting it is accepted debt until
  after 6 November (see *Next*).

## Where things stand

Updated every evening. The newest day comes first.

### Day five, 26 October (midday update: NVIDIA issues, emergency workaround in progress)

**NVIDIA's API is failing in waves.** Since the afternoon of 25 October UTC, NVIDIA's hosted
`nemotron-3.5-super-vl-120b-a12b` answers many calls with HTTP 500 ("Failed to generate completions") or times out.
We measured it before blaming anything: the same prompt at every size from 8 KB to 256 KB went through 37 times out of
37 in a quiet window, and inside a failure window even a tiny "Reply with OK" prompt failed 3 times out of 5. So it is
not our prompts. A support ticket with the request ids is open with NVIDIA.

**The emergency workaround, being set up now:**
- **Meterless Nightshift stays Proven's primary AI provider.** If NVIDIA is down when judging starts, Proven's reviews
  fall back to **GLM-5.3 on QwenCloud**. Nightshift's current provider list does not include QwenCloud, so this
  fallback calls QwenCloud directly, outside Nightshift, and says so: every review names the model that answered and
  that it was the fallback. A QwenCloud provider in Nightshift would let the fallback run through Nightshift too.
- **NVIDIA stays in the loop as the guardrail**: a smaller NVIDIA safety model, running on our own machine, checks
  prompts and answers.
- The moment Nemotron is stable again, it goes back to being Proven's default reviewer.

**What already landed today:**
- **World Model passes 8 of 8** on Meterless's own conformance runner, on `main`. H-MEM's 13 of 13 is in its last
  review rounds.
- **The first real pull request** on this repository, [#1](https://github.com/saphan-ai/proven/pull/1): a small fix
  written by the fleet from a finding in our own code review. Proven's GitHub Action ran on it; its first run was hit
  by the NVIDIA outage and posted no comment, which is being fixed so a degraded review still says what failed.
- **The whole record is public**: every change merged since the build window opened now has its order, the
  builder's report, every review and the decision in [hackathon/fleet/](hackathon/fleet/) (127 orders, 178 reports,
  154 reviews, 87 decisions).
- **Dependency alerts from 51 to 6**: the documentation site's packages were upgraded, and the 6 left are the live
  demo's planted vulnerable dependency, on purpose
  ([security review](docs/security/SECURITY_REVIEW.md#dependency-alerts-on-this-repository)).
- **[Honest limits](#honest-limits)**: what Proven does not do yet, measured and dated, now at the top of this page.

**Still being built before the feature freeze:** an interactive `make up` (sample or live, judges can use our Zetaris
and Nightshift through an authenticated gateway), Proven's pull-request comment in Markdown, and the judges' gateway.

### Day four, 25 October

Day four of six. **Thirty-two more changes were merged**, 100 in four days. With Scout Intent and Markovian, all four
Meterless engines are now in Proven, the two new ones switched off until their benefit is measured.

**NVIDIA's model now writes Proven's code, not only reviews it.** NVIDIA made its hosted API free for us, and the fleet
added NVIDIA executor seats beside the Cursor, Claude and OpenAI ones. The first piece of Proven written by Nemotron, a
small, tested example of a model-provider adapter, took 16 minutes and is merged (`3beaaab`). Thank you, NVIDIA.

**A pull request that fixes itself** ([ADR-0050](docs/adr/ADR-0050-one-automated-fix-attempt-proved-by-the-same-evidence.md)).
After a review, Proven makes one fix attempt for every proved finding, judges it only by the same tests and queries that
proved the findings, and ends with "fixed automatically, waiting for a person" or "rejected". Proven never commits to
the author's branch. The review result now carries the patch and the verdict; the pull-request comment shows them next.

What else came in today:
- **Proven as a GitHub Action**, so it can run on every pull request of a repository.
- **One signed evidence pack per review**, and an **organisation policy** that sets the floor for every repository.
- **Each agent sees only the evidence its role allows**, set in configuration.
- **A custody check on model answers**: an answer counts only when its signed receipt verifies and names the model that
  was asked. Counted for 25 October: Nightshift signed 12 receipts that day and all 12 signatures verify, under the key
  whose fingerprint was recorded on 22 October. 11 of the 12 calls failed (8 cut at the provider's deadline, 2 aborted,
  1 empty answer), and the one that succeeded was a probe, not a review. So no live Proven answer of that day reached the
  check: 0 answers in custody, and custody on a real review stays **UNMEASURED** until a live run gets an answer.
- **Kotlin and Gradle**, the third language added as one small adapter, and **four small, tested adapter examples** in
  [`examples/`](examples/).
- **Evidence in CouchDB as well as Zetaris**, with a read-only viewer of a review's evidence records.
- **More evidence through Zetaris**: packages published to do harm, and project health.
- **A review you can watch**: OpenTelemetry traces and metrics, with a Grafana dashboard in the judges' lab.
- **The level of proof in JSON and SARIF**, so a pipeline can tell a proved finding from an opinion.
- **The documentation site** is built, served from its own container; its public address comes with the next update.
- **The held-out test set**: six cases frozen before Proven ever ran on them.
- **Three more security fixes**: what the command line outputs can no longer be steered by the reviewed change, it
  writes its state and trace outside the reviewed repository, and a vulnerable network library was raised to its fixed
  version. What was fixed: [SECURITY.md](SECURITY.md).
- **Each review can be planned before a model runs, and it is switched OFF.** Meterless's Scout Intent, built in Go
  (`internal/scout`), classifies a change as dependency, refactor, feature, security-sensitive, documentation only or
  configuration, and picks the plan inside the configured review classes. It never lowers a configured class, and a
  configured class keeps every agent it requires. The setting is `scout_intent` and it is **off by default**: a person
  turns it on in `proven.yaml`, and until then every review runs the full plan. Its **benefit is UNMEASURED**: the model
  calls and the time saved per review on the test corpus cannot be stated until the corpus recordings exist. The call
  counts in [the architecture page](docs/ARCHITECTURE.md#scout-intent-plans-each-review-adr-0049) are scripted, not live,
  and are labelled so ([ADR-0049](docs/adr/ADR-0049-scout-intent-plans-each-review.md)).
- **Its evaluation, as it is.** Meterless's eval harness, pinned at upstream commit `1742dec5`, ran twice.
  [On Meterless's own regression set](test/conformance/meterless/SCORECARD_scout-intent.txt) (32 examples) every metric is
  1.000 and `clarification_rate` is 0.188; it shows the engine reproduces the set's seeds, not that it generalises.
  [On Proven's change set](test/conformance/meterless/SCORECARD_scout-intent-changes.txt) (14 fixtures) the gate **fails**:
  `clarification_rate` is 0.429 against the 0.25 ceiling, by design, because 6 of the 14 fixtures ask for clarification;
  `clarification_agreement` is 1.000, so it asked at the right times on this small authored set. `injection_precision` and
  `override_frequency` have no positive examples there, so they measure nothing. Neither scorecard measures cost, speed or
  whether a plan keeps its findings on real changes.
- **Large pull requests in bounded steps, switched OFF by default.** Proven carries Meterless's Markovian engine. When it is switched on with `PROVEN_MARKOVIAN`, a change above a configured size is reviewed one file or one hunk at a time, carrying only what is established, what is suspected and what awaits evidence. A claim that awaits evidence is carried whole, or the review stops by name: it never loses one in silence. A bounded review that stops ends with a named error and does not yet fall back to a whole review. Saphan Studio has its own engine, patent pending; for the hackathon we use Meterless's ([ADR-0048](docs/adr/ADR-0048-a-large-change-is-reviewed-in-bounded-steps.md), [how it works](docs/ARCHITECTURE.md#bounded-steps-meterlesss-markovian-engine-adr-0048)).
  - **Conformance:** Meterless's own runner, pinned at upstream commit `1742dec5`, gives `markovian 8/8 checks passed` ([scorecard](test/conformance/meterless/SCORECARD_markovian.txt)).
  - **Benefit: UNMEASURED.** Model tokens per step on the test corpus cannot be read until the corpus recordings exist. The figures in the architecture page's measurement table come from a scripted generator and are labelled scripted. We keep the engine only if it finds the same problems for less; until that is measured it stays off.

**How the fleet got faster.** The median time from an order's signature to its agent starting work fell from 220
minutes on 22 October to 74, then 32, then 21 today. By midday the agents had exchanged 165 messages among themselves;
of the 40 that required an answer, 31 were answered, with a median of 9.3 minutes. The day's log shows how they talk:
[CONVERSATION_2026-10-25_advisor.md](hackathon/planning/CONVERSATION_2026-10-25_advisor.md).

**The hardware**, all of it in one home office, no cloud: [SAPHAN.md](hackathon/SAPHAN.md#the-hardware-this-fleet-runs-on).

### Day three, extra time: three model families, one commit

Just after midnight UTC, NVIDIA's Nemotron returned its first review in the fleet. The change was the held-out test set:
six new cases, frozen before Proven ever runs on them. Three reviewers from three model families read the same commit,
each from a fresh context, none seeing the others' verdict:

| reviewer | verdict | what it found |
|---|---|---|
| OpenAI (Codex) | reject | one case's frozen answer key contradicts its own frozen contract |
| Anthropic (Claude) | accept with BLOCKING conditions | the same answer-key problem, and in two more cases the check's label contradicts the answer key |
| NVIDIA (Nemotron) | accept with conditions | one of those two label mismatches, judged non-blocking; it also re-ran one case's Go test live and confirmed it fails exactly as the case says |

Three problems in all. Claude found all three; Codex found one and Nemotron another, each also found by Claude. Codex
alone or Nemotron alone would have let two of the three through; the three verdicts did not even agree on whether the
change could ship. The reviews are published with the change's decision.

**Zero trust, caught in the act.** At 00:21 UTC an agent began building the Markovian engine and, a minute in, tried to
download Meterless's repository from GitHub. The fleet refused the connection (`CONNECT tunnel failed, response 403`)
and the run stopped. Nobody had to notice: an agent reaches its own model and its package registries, nothing else,
and opening anything more is a recorded decision. The engine's files will reach the agent the allowed way, handed to
the run with a pinned commit.

### Day three, 24 October

Day three of six, published late. **Thirty-three more changes were merged since the last update**, 68 in three days. The fleet has grown to five agent
masters, and NVIDIA's Nemotron joined the reviewers this evening as a third model family, running straight on NVIDIA's
API. Its first verdicts, side by side with Claude's and OpenAI's on the same change, come with tomorrow's update.

**The day's theme: Proven became something you can extend.** A new language, a new tool or a new evidence source is
one small adapter, not a change to the reviewer:
- **Two new languages through one adapter each**: PHP (Composer) and Rust (Cargo), each with a sample scenario.
- **A tool's run is evidence**, like a test or a query: the Go race detector, and `govulncheck`, which says whether a
  vulnerable function is actually reached from your code, only imported, or not measured.
- **One interface for where the evidence lives**: Zetaris or a local store, with one conformance test that
  fails when two stores answer differently.

What else came in today:
- **Every finding gets a level of proof**: proved by a test, proved by a query, or only an opinion. The reviewer
  computes it; the outputs do not print it yet, and that fix is under way. When both reviewers name the same place
  without evidence, Proven shows it apart as an agreed opinion and never counts it.
- **A proposed fix is proved by the same test** that proved the finding: it must pass on the base, fail on the change,
  and pass on the change with the fix, without breaking any other test. It stays a suggestion, never a commit.
- **The second reviewer can be any model**, and the generator can call NVIDIA's hosted API directly as well as through
  Nightshift.
- **JSON and SARIF output**, so a review can feed GitHub's code scanning or any pipeline.
- **A lab for judges**: one Makefile builds Proven and runs it against a local Gitea, on your own machine.
- **Security review, round two**: the fixes for the model-facing piece and the data-on-disk piece are merged, each
  proved by the reviewers' failing tests. What was fixed: [SECURITY.md](SECURITY.md).
- **The test corpus is built**: the code for the cases Proven is measured on. Tomorrow it gets a score, and a second,
  held-out set of cases, frozen before Proven ever sees them, so the score cannot be tuned to the test.

**Privacy, in one sentence:** every prompt and every answer is checked by NVIDIA's safety model, which can run on your
own machine, so that check need not send your code anywhere.

**See it run:** [what shipped on day three, step by step in a terminal](hackathon/SHIPPED_DAY_3.md).

### Day two, 23 October

Day two of six is behind us. **Eighteen more changes were merged** (35 in two days), each written by an agent, checked again by its master on a clean checkout, and approved by Marcin; the larger ones were also reviewed by two models other than the one that wrote them.

**The day's biggest step: Proven's code was put through a security review of its own.** Not one diff at a time, but
the whole code, cut into nine pieces, each one place where something untrusted meets something Proven protects: the
sandbox that runs the change's code, the keys, the reviewed repository, the evidence queries, the models, the output.
Each piece was reviewed by **two model families on the same commit**, Anthropic's Claude and OpenAI's, neither seeing
the other's verdict. They did not always agree: on two of the first three pieces one reviewer accepted with conditions
and the other rejected the same code and proved it with failing tests. With one reviewer, those pieces would have
passed. **The first weakness was found, fixed, proved by six tests and merged the same day**: the command run inside
the sandbox could come from a model's answer, and now Proven builds it itself. Tomorrow NVIDIA's Nemotron reviews the
pieces again as a third family. The plan, the status of every piece and the verdicts side by side:
[docs/security/SECURITY_REVIEW.md](docs/security/SECURITY_REVIEW.md); what was fixed: [SECURITY.md](SECURITY.md).

What else came in today:
- **Proven's own command line**: `proven scan` for a developer and `proven ci` for a pipeline, with the flags and exit codes a CI pipeline already expects. It took 14 rounds: coded by Grok on Cursor, reviewed each round by Claude and OpenAI.
- **More dependency evidence**: `go.sum`, npm, `requirements.txt` and `poetry.lock` are read; each changed package is checked against OSV by version range, the known-exploited list, and its EPSS score.
- **One sample scenario per kind of evidence and refusal**, all runnable with no network and no keys.
- **A missing key no longer stops a review**: the service is named as down, and a guardrail that is down refuses every model call instead of letting it through unchecked.
- **Code written to be read**: the masters went over the code written so far and added the comments a judge needs.
- **We reshaped the repository, because Proven is no longer a toy**: the product at the top, laid out the way a Go project is, and the build's record in [`hackathon/`](hackathon/), so that after the hackathon what is left is a product. Eleven design decisions were written today ([ADR-0026 to ADR-0036](docs/adr/)).
- **Two more agent masters joined**, one for the command line and one for the security review, because one master was carrying four streams of work.

**See it run:** [what shipped on day two, step by step in a terminal](hackathon/SHIPPED_DAY_2.md).

### What works today with the sponsors' technology

| Sponsor | What works, measured | In a review |
|---|---|---|
| **Zetaris** | The platform runs on our own machine and its checks pass 9 of 9. Proven's own key reads over REST and is refused every write. OSV by package answers from a file in Zetaris storage, 0.10 to 0.15 s per query. | The Investigator's evidence (KEV, EPSS, OSV) is shaped as Zetaris queries; sample mode runs the same SQL over local tables. [setup](hackathon/fleet/infra/zetaris/README.md) |
| **Meterless** | Nightshift answers every model call with a signed receipt: 241 receipts by 25 October, all 241 signatures verify. H-MEM and the World Model run in Go and pass Meterless's own conformance suites 13 of 13 and 8 of 8. Markovian is in Proven too and passes Meterless's runner 8 of 8 (pinned at upstream commit `1742dec5`); it is OFF by default, and its benefit is UNMEASURED. **Scout Intent**, which plans each review, runs Meterless's own eval harness (pinned at upstream commit `1742dec5`): on Meterless's regression set every metric is 1.000 and `clarification_rate` is 0.188; on Proven's 14 change fixtures `clarification_rate` is 0.429 against a ceiling of 0.25, so that gate fails, because 6 of the 14 are ambiguous on purpose. | H-MEM and the World Model run in every review, sample mode included; the fleet's own reviews also go through Nightshift. Markovian reviews a large change in bounded steps only when `PROVEN_MARKOVIAN` is switched on; by default the review is unchanged. **Scout Intent is OFF by default** (`scout_intent`), so no review plans itself unless a person switches it on; its benefit, the model calls and time saved per review on the test corpus, is **UNMEASURED** until the corpus recordings exist. [NVIDIA and Nightshift](hackathon/fleet/infra/nvidia/README.md) · [Markovian](docs/ARCHITECTURE.md#bounded-steps-meterlesss-markovian-engine-adr-0048) · [how it plans](docs/ARCHITECTURE.md#scout-intent-plans-each-review-adr-0049) |
| **NVIDIA** | Nemotron 3.5 Super answers through Nightshift, named in each signed receipt. Nemotron 3.5 Content Safety runs on our own machine as the guardrail. | Every prompt and answer passes the guardrail; the Skeptic that tries to refute a finding defaults to Nemotron. In the fleet, Nemotron now reviews changes beside OpenAI and Anthropic, straight on NVIDIA's API. |
| **Cursor** | The command line was written by Cursor agents on Grok 4.7 and landed today; up to four agents code at once on one Linux machine, each inside its own jail, and every run's model traffic is recorded. | Cursor is how Proven is built. [how it works](hackathon/fleet/infra/cursor/README.md) |

### Next: everything still in the plan

Each day, what lands moves from this list to the day's summary.

**Already done from this plan** (checked on `main`, 26 October):
- **Run Proven live with a single free NVIDIA key** ([ADR-0031](docs/adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md)).
- **An evidence pack an auditor can check offline**: `proven verify` checks the signature and every hash, then works
  out each finding's level of proof again from the recorded evidence and refuses a pack whose verdict does not follow
  ([ADR-0035](docs/adr/ADR-0035-every-review-leaves-a-signed-evidence-pack.md); the forgery bench and its counts are in
  [EVIDENCE_PACK.md](docs/EVIDENCE_PACK.md#the-forgery-bench)).
- **Watch a review live**: a Grafana dashboard of the agents and the two reviewers, through OpenTelemetry
  ([ADR-0045](docs/adr/ADR-0045-a-review-is-observable-through-opentelemetry.md)).
- **An organisation policy with exceptions that expire**: an exception accepts one known finding until a date, with an
  owner and a reason ([ADR-0036](docs/adr/ADR-0036-an-organisation-policy-and-exceptions-that-expire.md)).
- **More sources, first part**: CVSS severity scores and licences ([ADR-0039](docs/adr/ADR-0039-more-evidence-sources.md)).
- **More languages, first part**: PHP, Rust, Kotlin, and Java projects built with Gradle.
- **Worked, tested adapter examples** for extending Proven, and the documentation site, built.

**Still in the plan:**

- **NVIDIA's models, already local.** Today NVIDIA's safety model runs on our own AMD Ryzen AI Max+ 395 machine with 128 GB of memory and checks every prompt and answer of Proven's reviews, while Nemotron reviews and writes code through NVIDIA's API. A DGX Spark is the same class of machine, built on NVIDIA's own silicon: on it the same fleet could run Nemotron for reviewing and coding locally beside the guardrail, tune NeMo Guardrails on the fleet's own record of what it refused and why, and keep every line of code in the room.
- **A score anyone can recompute, on two sets**: one command runs Proven over the cases it was built on and, separately, over a held-out set frozen before Proven ever saw it; the two scores are never merged into one number.
- **Scout Intent and Markovian measured**: on the corpus, with each engine on and off, the same findings for less, or the engine stays off ([ADR-0048](docs/adr/ADR-0048-a-large-change-is-reviewed-in-bounded-steps.md), [ADR-0049](docs/adr/ADR-0049-scout-intent-plans-each-review.md)).
- **The documentation site live at proven.saphan.ai**, and a worked example for the adapter kinds still without one.
- **The coding sessions, published**: excerpts of the Cursor agents' recorded sessions next to the change they built, each tied to its original by a hash ([ADR-0024](docs/adr/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md)).
- **Reports people can read**, in HTML and Markdown, next to JSON and SARIF for machines ([ADR-0033](docs/adr/ADR-0033-the-text-review-is-written-for-a-person.md)).
- **A readable comment on the pull request**, in GitHub and in Gitea, with a plain-language summary first and the automated fix as suggested changes.
- **Your own data**: plug in an internal list of banned packages, how critical each service is, or who owns the code, and Proven judges a change against what your organisation knows ([ADR-0046](docs/adr/ADR-0046-a-data-source-is-one-adapter.md)).
- **More languages**: C# and Ruby, and Java projects built with Maven, through the same small adapter ([ADR-0038](docs/adr/ADR-0038-a-language-is-one-small-adapter.md)).
- **An air-gapped installation**, where no code leaves your network ([ADR-0034](docs/adr/ADR-0034-proven-runs-air-gapped.md)).
- **Accepted risks shown in every output**, on top of the organisation policy.
- **NVIDIA's model in every review of the fleet's work**, beside OpenAI's and Anthropic's, now that its first review is in.
- **A measurement on 21 real review disputes**: how often Proven is right compared with a single model.

**Known debt, accepted on purpose.** The function that runs a review, `pipeline.review`, is one long function of about
500 lines. It should become one function per stage (Planner, Investigator, Prover, Skeptic, Judge, Writer), which is about
a day of work. We know it, and we accept the debt during the feature freeze: a refactor of the review's spine two days
before the deadline is a bigger risk than the long function itself. We split it after 6 November.

One day to go.

### Day one, 22 October

**Seventeen changes were merged into Proven on day one**, each one written by an agent, checked again by its master on a clean checkout on another machine, and approved by Marcin. Every merge commit carries the request it answers and the reason it was accepted.

**See it run:** [what shipped on day one, step by step in a terminal](hackathon/SHIPPED_DAY_1.md).

What Proven can do at the end of day one, in sample mode (recorded data, no network):
- **The six reviewer agents run end to end** behind the command-line shell's request, and honour every field of it: baseline, head, include and exclude paths, and a configuration file.
- **The Skeptic tries to refute every finding** with a model from another family, and **the Judge settles it by evidence**. A finding nobody could settle is reported as unresolved, never as true.
- **The Prover writes the test that proves a finding** and runs it only inside a container, once on the base and once on the change.
- **The Investigator's evidence comes through Zetaris**: KEV, EPSS and OSV queries over REST, and the same queries run locally over samples in sample mode.
- **Every prompt and every answer passes a guardrail**, on by default, with two backends: NVIDIA's safety model on our own machine, or NVIDIA's hosted API. An unsafe prompt or answer stops the step and is recorded with its categories.
- **A minimum of Meterless's H-MEM and World Model in Go**, so a finding refuted once does not come back in the next review. Meterless's own conformance suites run against it, and the scorecard is recorded as it is: H-MEM 9 of 13, World Model 8 of 8. The failures are the sections not built yet.
- **The live generator calls Meterless Nightshift**, reads which model answered, and refuses an answer that did not come from NVIDIA.
- **One command checks every live service a review needs** before a demo, and names each one that is not ready, without printing a secret.

| what | status | where |
|---|---|---|
| Choosing the topic, with eleven answers from the fleet | done, 16 October | [answers](hackathon/planning/answers/) |
| The design, as recorded decisions | 25 decisions accepted, 17–21 October | [decisions](docs/adr/) |
| Building Proven | **17 changes merged on day one**, all from the reviewer train; the shell's first change is still in progress | [git log](https://github.com/saphan-ai/proven/commits/main) · [reviewer order](hackathon/fleet/orders/TRAIN_ORDER_proven-reviewer-core.md) · [shell order](hackathon/fleet/orders/TRAIN_ORDER_proven-cli-and-action.md) |
| How the agents code with Cursor | four Cursor accounts on one Linux machine, up to four agents coding at once on Grok 4.7, each inside a jail, its own network and an egress door, with every model request recorded | [how it works](hackathon/fleet/infra/cursor/README.md) · [order](hackathon/fleet/orders/TRAIN_ORDER_proven-cursor-infra.md) |
| Recording every Cursor coding session | every session is recorded by the proxy; the first published recordings will show the coding itself, as excerpts tied to the sealed original | [ADR-0024](docs/adr/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md) |
| Zetaris, the sponsor's data platform | running on our own machine; the Investigator reads its evidence through Zetaris over REST; live Zetaris evidence in a review is next | [order](hackathon/fleet/orders/TRAIN_ORDER_proven-zetaris-local.md) · [ADR-0023](docs/adr/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) · [setup](hackathon/fleet/infra/zetaris/README.md) |
| Meterless Nightshift and NVIDIA Nemotron | running in its own VM; 241 signed receipts checked by 25 October, every signature verifies, and a receipt ties each answer to the model that gave it | [NVIDIA page](hackathon/fleet/infra/nvidia/README.md) · [runbook](hackathon/fleet/infra/nightshift-vm/README.md) · [ADR-0019](docs/adr/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) |
| A local NVIDIA safety model | done: running on our own machine, used by the guardrail | [order](hackathon/fleet/orders/TRAIN_ORDER_nemotron-content-safety.md) |
| A test corpus from 28 real review disagreements | the 21 case contracts are prior art, written before 22 October; their code is written inside the build window: the first five cases are coded on branches, not merged yet, and the other 16 are ordered, one agent per case | [tests](test/README.md) · [order](hackathon/fleet/orders/TRAIN_ORDER_proven-test-corpus.md) |
| A held-out set of six fresh cases | frozen in commit `0fabc62` and amended before any run in `b4bab25` (HO-003's reference modes made exact) and `31ab87c` (HO-002's answer key aligned with its contract), both after independent reviews: contracts and answer keys, written before Proven is run on any of them; one run follows, misses published | [held-out set 1](test/corpus/held-out/README.md) |
| Reviews of the fleet's own changes | every change is checked again by its master; independent reviews run in parallel by two model families. NVIDIA Nemotron reviews resume once they run on Saphan Studio's own agent loop (planned for 23 October); the trains keep a list of the changes waiting for one | [conversation logs](hackathon/planning/CONVERSATIONS.md) |
| Public updates | daily, 15:00 UTC | this repository |

**Saphan Studio's Cursor backend, built in parallel.** While the agents coded Proven, Saphan Studio's own fleet kept improving the Cursor backend that those agents run on, and each fix reached the live coding sessions as soon as it landed. Today it gained confinement for Cursor on macOS, Grok models named with their effort (`grok-4.7-high`), every Cursor spawn speaking HTTP/1.1 so the recording proxy can read it, placement of orders signed by a train's master, and a reader of a Cursor account's usage, so the scheduler can place Cursor runs by itself. Still to do: a Cursor seat in a container, and publishing the first session recordings.

Before Proven exists, we also wrote down what it will be tested on: [`tests/`](test/README.md) turns real review disagreements from our fleet into checkable cases, and is declared as prior art.

The preparation work sets up existing machines and tools for the agents. It is not product code, and it trains no models. Saphan Studio is a pre-existing component; the organizers confirmed that it may run the fleet.

## Why a tool like this is needed

One example from the fleet's earlier work, before Proven:

An agent wrote a protection against deleting the main branch of a repository, plus 16 tests to show it worked. All 16 passed. An independent reviewer then **removed the new protection** and ran the tests again. All 16 still passed: a different safeguard, in git itself, was stopping the deletions one step earlier. The tests proved git, not the fix. ([source](hackathon/planning/answers/master-the-bridge-commits-without-a-shared-lock.md))

Passing tests do not always prove what they claim to prove. A reviewer that asks "what does this evidence actually show?" catches what a green checkmark hides.

## What Proven is meant to do

*The intent, as planned before the build window. What already works is listed under [Where things stand](#where-things-stand).*

- **Every finding carries evidence.** A test that fails on the change and passes without it, or a recorded query and its results, with sources. This applies to every finding, including when two reviewers agree. Agreement is never a substitute for proof.
- **The reviewer should come from a different AI model family than the one that wrote the code.** A reviewer from the same family tends to share the author's assumptions; another family can challenge them. The change's own commit or pull request says which model wrote it, and by default an NVIDIA model reviews it. Every model is configurable, and no setup is blocked; each review says which model reviewed it and whether that is the author's.
- **For special kinds of review, such as security, two independent reviews.** Best with two models other than the author's; with fewer, the second review is a fresh agent that sees the code for the first time. In the fleet's own record, reviewers from two model families looked at the same change within 30 minutes and left a readable verdict 40 times, and reached different verdicts 28 times ([source](hackathon/planning/answers/README.md); first published as 31 of 41, corrected on 19 October). That is an observation from one fleet, not a benchmark, but it shows why a second opinion is worth having.
- **When reviews conflict, the evidence decides, not a vote.** Both sides' evidence is examined: what it actually establishes, what it assumes, and whether it reproduces. What stays unresolved is reported as unresolved, and a person decides.

Proven will not prove a whole program correct or find every vulnerability. It makes each claim it does make checkable.

How one review is meant to run, and where each sponsor's technology does the work:

```
  pull request ──▶ proven ci   (GitHub Action · Gitea Actions · command line)
                       │
  ┌────────────────────▼─────────── Proven's agents ─────────────────────┐
  │ Planner ─▶ Investigator ─▶ Prover ─▶ Skeptic ─▶ Judge ─▶ Writer      │
  │    ▲                                               │                 │
  │    └──────────── weak finding: back for evidence ──┘                 │
  └──────┬──────────────────────┬──────────────────────────┬─────────────┘
         │ SQL over REST        │ every model call         │ read · write
         ▼                      ▼                          ▼
  ┌──────────────┐  ┌──────────────────────────┐  ┌──────────────────────┐
  │ ZETARIS      │  │ METERLESS NIGHTSHIFT     │  │ METERLESS ENGINES    │
  │ (Data)       │  │ primary AI provider      │  │ H-MEM: refuted       │
  │  OSV         │  │ (Power)                  │  │  findings, rulings   │
  │  CISA KEV    │  │   └─▶ NVIDIA NEMOTRON    │  │ World Model: claims, │
  │  FIRST EPSS  │  │       reviewer, Skeptic  │  │  evidence, verdicts  │
  │  git history │  │                          │  │ Scout Intent: plans  │
  │              │  │                          │  │  each review, OFF    │
  │              │  │                          │  │ Markovian: OFF by    │
  │              │  │                          │  │  default, large      │
  │              │  │                          │  │  change in steps     │
  └──────┬───────┘  └──────────────────────────┘  └──────────┬───────────┘
         │ the query and its rows                            │ JSONL
         ▼                                                   ▼
  every finding with its evidence ─▶ SARIF · JSON · text · PR comment, logs/

  NVIDIA Nemotron 3.5 Content Safety checks every prompt and every answer.
```

The technical design, including the sponsors' technology (Zetaris, Meterless, NVIDIA, Cursor), is in **[DESIGN.md](hackathon/DESIGN.md)**.

## How Proven differs

Large vendors are moving the same way: a security finding should be proved, not only reported. Proven takes that idea
somewhere they do not:

- **Open source, and it runs where your code is**: on GitHub, on your own Gitea, on your own servers, or fully
  air-gapped. It reviews every pull request, not a nightly scan.
- **More than one model, in the open.** Proven names the models that reviewed each change. Connect a second reviewer
  and you see both verdicts, and exactly where they disagree.
- **An agent is not trusted, even when it is not malicious.** A model's answer is treated as untrusted input: the
  change's code runs only in a container with no network and no secret, and the command run there is built by Proven,
  never taken from a model.
- **Proof anyone can rerun, and the test fails first**: a test that passes on the old code and fails on the new, or a
  query and the rows it returned. A finding whose test does not fail on the change is not proved. No exploit code is
  written or kept.
- **A fix is proved too**: the same test that showed the problem shows the fix works.

## What this week gave us, and thank you

For us as developers, this week was a mega experience. Writing at this speed and at this scale, without dropping the
rule that every change is reviewed and every claim is proved, taught us more than months of normal work.

- **Scale shows what small projects hide.** A way of reading data that passes every test on a small project breaks
  when the data has grown. At one or two agents you never see it; at a fleet working on four projects at once you see
  it within hours. Each time, we turned the lesson into a written rule for every agent, so the next change is written
  right the first time.
- **Rigour is what makes speed safe.** Two independent reviews on every change, from different model families and run
  at the same time, caught problems that green tests did not, and they did it without slowing the work down.
- **Measure first, then say it.** More than once a first explanation turned out wrong when we measured it. What we
  publish here is what we measured.
- **We fixed a lot along the way**, in Proven and in how the fleet itself works, and the record of every fix is in this
  repository.

**Thank you to [GenAI Works](https://hackathon.genai.works/event/open-agent-hackathon-2026)** for organising the Open
Agent Hackathon 2026 and for the chance to take part, and to the sponsors whose technology Proven is built on:
**Zetaris**, **Meterless**, **NVIDIA** and **Cursor**. Building on your platforms under a deadline was the best possible
test of what an agent fleet can do.

## Who builds it

Two different groups of AI agents appear in this project. Do not mix them up:

| | who | what they do |
|---|---|---|
| **The fleet that builds Proven** | coding agents run by Saphan Studio, with Marcin | plan each change, write the code, test it independently, review it, and hand Marcin a decision |
| **The agents inside Proven** | Proven's own reviewers, once built | review someone else's pull request: plan, gather evidence, try to refute, judge, write the report |

In the fleet, four Cursor agents write the product code. Other agents plan the work, write tests separately from the code, and review each change without having written it. Marcin sets the goals, rules on scope, settles disputes the evidence cannot settle, and approves every merge. An agent can recommend; only he can approve. No agent ever pushes: each change is built in its own copy of the repository, and only Marcin merges and pushes. How the fleet is organized is explained in **[SAPHAN.md](hackathon/SAPHAN.md)**.

## Follow the work

- **[Documentation](docs/README.md)**: how to run Proven, starting with a review on samples (no network, no keys), its architecture and every setting.
- **[The diary](hackathon/planning/CONVERSATIONS.md)**: every conversation between Marcin and the fleet, day by day.
- **[Watch it being built, slice by slice](hackathon/fleet/)**: from 22 October, each change arrives with the request that asked for it, the builder's report, the independent review and Marcin's decision.
- **[Design decisions](docs/adr/)**: what was decided, and why.
- **[How each change went](hackathon/fleet/LEDGER.md)**: every change's rounds, the conditions its reviews raised, and its outcome.
- **[Security](SECURITY.md)**: what Proven treats as untrusted, what it protects and how, and **[how its code is reviewed for security as a whole](docs/security/SECURITY_REVIEW.md)**, piece by piece.
- **[How to code with Saphan Studio, step by step](https://github.com/saphan-ai/quake-watch-server)**: published alongside Proven, a sample app built as a series. First the server, an earthquake alert service on the USGS public feed; the mobile clients follow in their own repositories. Every step is recorded as it is built: the order, the agents' work, the review and the owner's decision.

**Updates:** once a day at **15:00 UTC** (22:00 in Bangkok; 17:00 in Warsaw until 25 October, 16:00 after). The last push comes before submissions close on 27 October, 23:45 UTC.

## Timeline: what each push brought

Every public push is tagged, so you can see exactly what existed before the build window and what each day added.

| push | tag | what came in |
|---|---|---|
| before 22 October, 00:00 UTC: **prior art** | [`prior-art`](https://github.com/saphan-ai/proven/tree/prior-art) (`bf05dd2`, 21 October, 19:53 UTC) | Preparation only, no product code: the topic and the fleet's eleven answers, 25 design decisions, the trains' orders, the machines and services for the agents (Cursor seats, Zetaris, Meterless Nightshift, NVIDIA's safety model), and the 21 test-corpus case contracts |
| 22 October, 00:03 UTC: **coding starts** | [`start`](https://github.com/saphan-ai/proven/compare/prior-art...start) (`3a1a9f0`) | The first line of product code: the agreed seam between the command-line shell and the reviewer, types only |
| 22 October, 15:17 UTC: **day 1** | [`day-1`](https://github.com/saphan-ai/proven/compare/start...day-1) | 17 changes merged: the six reviewer agents end to end in sample mode, the Skeptic and the Judge, the Prover writing and running its test in a container, evidence through Zetaris, the guardrail, the H-MEM and World Model minimum with Meterless's conformance scorecard, the live generator through Nightshift, the live-service check, and a small live demo a judge can rerun. The first documentation and the Cursor and NVIDIA setup pages. [What shipped, in a terminal](hackathon/SHIPPED_DAY_1.md). |
| 23 October, 15:00 UTC: **day 2** | [`day-2`](https://github.com/saphan-ai/proven/compare/day-1...day-2) | 18 changes merged: Proven's command line (`scan`, `ci`), dependency evidence for go.sum, npm, requirements.txt and poetry.lock with OSV ranges, KEV and EPSS, one sample scenario per kind of evidence, a missing key named as a down service. A security review of the whole code in nine pieces, with its first fix. The repository split into the product and `hackathon/`; ADR-0026 to ADR-0036. [What shipped, in a terminal](hackathon/SHIPPED_DAY_2.md). |
| 24 October, 23:00 UTC (late): **day 3** | [`day-3`](https://github.com/saphan-ai/proven/compare/day-2...day-3) | 33 changes merged: PHP and Rust through one language adapter each, tool runs as evidence (the race detector, `govulncheck` with reachability), one interface for the evidence store, every finding given a level of proof, fixes proved by the same test, any model as the second reviewer, NVIDIA's hosted API as a generator, JSON and SARIF output, a judges' lab with Gitea, the second round of security fixes, and the test corpus's code. NVIDIA Nemotron joins the fleet's reviewers. [What shipped, in a terminal](hackathon/SHIPPED_DAY_3.md). |
| 25 October, 17:00 UTC: **day 4** | [`day-4`](https://github.com/saphan-ai/proven/compare/day-3...day-4) | 32 changes merged: NVIDIA's Nemotron writes its first Proven code, Meterless's Scout Intent and Markovian (all four Meterless engines in, the two new ones off by default), one automated fix attempt per review (ADR-0050), Proven as a GitHub Action, a signed evidence pack, an organisation policy, evidence scope per agent role, a custody check on model answers, Kotlin and Gradle, adapter examples, CouchDB as an evidence store with a viewer, more Zetaris evidence, OpenTelemetry with a Grafana dashboard, the documentation site built, the held-out test set. |

**Contact:** Marcin Marzec, [marcin.marzec@wato.org](mailto:marcin.marzec@wato.org), [LinkedIn](https://www.linkedin.com/in/marcinmarzec).

**Found a vulnerability?** Please report it privately: [Report a vulnerability](https://github.com/saphan-ai/proven/security/advisories/new) on this repository's Security tab (GitHub's private vulnerability reporting is on). Not in a public issue. See [SECURITY.md](SECURITY.md).

## Repository layout

The repository holds two things: **the product**, and, in one folder, **the record of how it was built in the hackathon
week** ([ADR-0030](docs/adr/ADR-0030-the-product-and-the-hackathon-live-apart.md)).

| Path | What it is |
|---|---|
| `cmd/`, `internal/` | Proven's code: the command line and the reviewer |
| `docs/` | how to install, run and configure Proven; `docs/adr/` holds the design decisions, `docs/security/` the security review |
| `docs-site/` | the documentation website, built from this repository's Markdown and from the code |
| `SECURITY.md` | what Proven protects, and how to report a vulnerability |
| `test/` | the test corpus (`test/corpus/`) and the Meterless conformance runner (`test/conformance/`) |
| `testdata/`, `demo/`, `assets/` | sample scenarios, the live demo, images |
| `hackathon/` | the build week's record: `fleet/` (every change's order, report, review and decision), `planning/` (the topic, the daily conversations), `SAPHAN.md`, `DESIGN.md` |
| `metadata.json`, `metadata.schema.json` | the hackathon's submission metadata, at the root until the hackathon ends |

## Documentation site

The documentation website is built from this repository's own Markdown and from the code, and packaged as one container image. Run it locally:

```bash
docker build -f docs-site/Dockerfile -t proven-docs .
docker run -p 3000:3000 proven-docs
```

Then open `http://localhost:3000/docs`. The command-line, settings and adapter pages are written from the code during the build, and the build fails when a link is broken, a page matches a known shape of an internal path, a private host name or a credential (the list is in `docs-site/scripts/guards.mjs`; a value outside those shapes is not caught), or a generated page is out of date. After a flag, a setting or an adapter interface changes, regenerate the committed copies with `sh docs-site/scripts/generate.sh docs-site/generated`.

Deploy it to Cloudflare, from `docs-site/`, with one command:

```bash
CLOUDFLARE_ACCOUNT_ID=<your account id> CLOUDFLARE_API_TOKEN=<a token that may edit Workers, Containers and DNS> npx wrangler@4.45.0 deploy
```

It needs Docker running, and it serves `proven.saphan.ai`: the route in `docs-site/wrangler.jsonc` and its `DOCS_HOST` setting name that host, and a different host is changed in both.

## For technical readers

| | |
|---|---|
| [docs/SAMPLE_MODE.md](docs/SAMPLE_MODE.md) | try Proven with no network and no keys: the sample review, its findings and its trace |
| [DESIGN.md](hackathon/DESIGN.md) | Proven's design, the sponsors' technology, evidence sources, open questions |
| [SAPHAN.md](hackathon/SAPHAN.md) | the platform that runs the fleet, its roles and vocabulary |
| [AGENTS.md](AGENTS.md) | the rules every agent follows in this repository |
| [planning/](hackathon/planning/) | how the topic was chosen, and the fleet's answers |
| [fleet/](hackathon/fleet/) | orders, reports, reviews, decisions and the ledger of every change, with their templates |

**Pre-existing components:** [Saphan Studio](https://saphan.ai/) runs the fleet; it is not part of this submission and not in this repository. The [Meterless engine specifications](https://github.com/Meterless/Meterless) (Apache-2.0) are open source; we build to them and do not copy their reference code. Model providers, tools and data sources are listed as they are used.

Before publication Marcin reviews every document. Credentials, internal host names and private infrastructure details are blacked out, and each document says where.

**Licence:** the code is Apache-2.0 ([LICENSE](LICENSE)); the documentation and the build record (the Markdown pages,
`docs/` and `hackathon/`) are CC BY 4.0 ([LICENSE-docs](LICENSE-docs)).
