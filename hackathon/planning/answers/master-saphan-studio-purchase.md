# Answer — master of track `saphan-studio-purchase`

Author: the master of `saphan-studio-purchase` (Solo purchase portal, saphan-oauth sign-up, lease, production deploy).
Date: 2026-10-16. Evidence base: 21 named review runs on this track, 2026-10-11 → 15, about 71 conditions.
Paths are relative to the workspace. `TP` means `██████████████████████████saphan-studio-purchase/`, and `RA` means
`███████████████████████████████<slug>/…/_REVIEW_NOTES.md`.

## 1. Critique of "Proof-carrying code review"

**The core is right.** On this track the findings that ended arguments were the ones backed by a run. Weak findings were
not wrong; they were opinions with nothing to check them against. Six things in my record that the proposal gets wrong
or misses:

**1. "Fails on head, passes on base" is the wrong proof for half of the real findings.** The commonest strong finding on
my track was *"this guard is not tested"*, not *"this line is wrong"*. Examples:
- `solo-portal` C7: ███████████████████████ is untested.
- `solo-lease-contract-and-go-verifier` C1: █████████ ████████ ███████ ███ ███ ███████.

For these, no test fails on head, because the code is correct and simply unguarded. The proof that worked was a **mutant
kill**: `TP/MASTER_VERIFY_solo-lease-contract-and-go-verifier_round-2` turned `████████████` into `if false && …` and
named the tests that went red. The reviewer's own mutants also produced the useful negatives:
- `node-leg-keeps-its-node-grant`: the mutants are "EQUIVALENT … none is missing".
- `console-activates-solo-lease` M8: the mutant survived, but "the property holds" because the outer guard still refuses.

⇒ PROVEN needs two evidence kinds: *failing test* for defects and *surviving mutant* for missing coverage. A surviving
mutant that turns out to be equivalent must be able to DROP a finding, with the reason.

**2. A quarter of the disagreements were not settleable by evidence, and the tool must say so.** About 10 of the 71
conditions needed a person's decision about intent or scope, not a fact:
- `policy-projection-on-land` F1: the probe proved the door refuses. Whether it *should* was the owner's call: „Commit
  lądującego” ("the landing human's commit").
- `setup-passkey-signs-machine-admit` C1/C2: the owner ruled „Rozdziel i ratyfikuj” ("split and ratify") and
  „(b) Zawęzić” ("narrow it").
- `executor-node-is-a-technology-class-the-engine-picks` C1: the owner ruled „Zostawmy node — szeroko” ("keep Node wide").

In each of these the evidence was complete and the reviewers were both right about the facts. **Add a fourth label,
DECISION**: the evidence is assembled, the choice is named, and the tool refuses to pick. A tool that settles these
"by evidence" will settle them wrongly and confidently.

**3. On my track the false alarms came from the author, not the reviewer.** No reviewer finding was refuted in any
executor return. Every refutation on record went against the **order's premise**, i.e. the PR description:
- `setup-account-is-the-saphan-account` C1: my order claimed that activation without an email was refused. The leg showed
  it was not, by reading `███████████████████████████████████████` and adding a test (`TP/MASTER_VERIFY_setup-account…_round-2`).
- `executor-node-runs-shell`: "master reading 1 was wrong in part". Both grants already allowed sh/bash, confirmed with
  `git show <base>:<file>`.
- `node-leg-keeps-its-node-grant`: the wrong bytes were chosen at dispatch, not by the later writer.

⇒ **Check the PR description's claims as candidate findings too**, with the same three labels. "The author says X; the
code at base shows not-X" is the cheapest proven finding there is, and no reviewer I know produces it.

**4. The misses were environment facts, and a container on base/head would not have caught them either.** What reached
production unreviewed on `deploy-portal-on-cloudflare` (`TP/orders/ORDER_W9c_…`, `ORDER_W9c_r2_…`, purchase commit `9886a55`):
- The Worker forwarded `https://` to the container, and the platform refuses that, so every request answered 503.
- The container had no inactivity timeout. It was stopped and restored every minute, and the owner's password save
  "hung" with it.
- The installer was run under `dash`/`sh`, not `bash`. A one-space break in `supervisor.test.sh:77` was caught only by
  the master's run.
- The portal's install line pointed at an empty bucket (404, `TP/TRAIN_ORDER.md`, *"publish a preview release"*).

Review r5 had listed "Cloudflare platform behaviour" and "container idle behaviour" as **Unrun**, which was honest and
was then lost in the notes. Review r1 had seen the swallowed `catch (_)` that hid the 503 cause, but filed it inside
another item, and the remedy fixed only that item. ⇒ Three changes:
- (a) Make **UNRUN** a first-class output: "what I could not run, and the one command that would". In this record it was
  the most valuable line and the least visible.
- (b) One finding per claim: never a defect nested inside another one's remedy.
- (c) Let the evidence runner take an **environment matrix** (shell, OS, runtime, database). A pass on one
  environment is labelled as such.

**5. Evidence must name where it ran.** `solo-portal` C3: the leg's test passed on pg-mem, which "ignores its WHERE",
so the fix was "unverified on PostgreSQL". `local-machine-cert-from-own-ca`: the master measured "base 90 · tip 91,
flake", while the reviewer measured 188 FAIL on another seat, and the base stayed UNRUN. Two honest runs disagreed
because they ran in different places. ⇒ Every evidence item carries its environment, and a disagreement between two
*runs* is shown as such, not averaged.

**6. The biggest risk is the one my track lived: a reviewer that cannot run anything.** My review seats could run `go`,
`git`, `grep` and `diff`, but not `sh`, `npm` or `docker`. The W9b, W2 and deploy reviews were "judged by reading —
nothing executable". `TP/TRAIN_ORDER.md:212` (the W0f entry): *"W9b took ten rounds, the last six blind, because its legs could not run
their own shell harnesses."* Rounds 4–9 got no named review at all, because the review seat was refused placement.
⇒ The day-2 signal in the proposal is the right one. I would make it stricter: by day 2, run the planted bugs on two
languages and two shells, or cut.

**What would make a reviewer's output useful to me, instead of something I triage:** every condition sorted into
**fix** (the leg does it), **run** (someone must execute this — the reviewer could not) or **decide** (a person
chooses), with the evidence attached. On this track that split was roughly 55 / 6 / 10. I sorted all 71 by hand.

## Proposal 1: Why is production down? Deploy incident explainer

- **track:** The Agent That Can Explain Why.
- **the problem, in one sentence a judge understands:** after a deploy the service answers 503, and the cause is spread
  over the code diff, the platform's logs, the platform's rules and side effects nobody reads — so the person on call
  guesses.
- **who has this problem outside our fleet:** every small team on a container or serverless platform (Cloudflare,
  Fly, Render, Cloud Run) without an SRE; the dev who deployed on Friday.
- **the data sources the agent investigates:**
  1. the deploy diff and its commits;
  2. platform logs and the deploy record;
  3. live probes it runs itself (`curl` with timing, DNS from two resolvers);
  4. **side-channel witnesses**: backup timestamps, cron output, restart counters. In our case the R2 dump cadence
     showed the container restarting every minute while its own stdout was invisible;
  5. the platform's documentation pages.
- **what "explained with evidence" means here — one answer:**
  ```
  CAUSE 1 · PROVEN — the container is stopped and restored every ~60 s
  Evidence  1. backups land every 60 s (12:01, 12:02, 12:03 …); the schedule says every 5 min
            2. probe: 3 of 10 requests between :58 and :05 answer 503 in < 50 ms; the rest 200
            3. worker.js:41 creates the container with no inactivity timeout (added in deploy a1b2…)
            4. platform docs, "Container lifecycle": the default sleep is 60 s idle
  Fix       set the inactivity timeout on start and on keepalive; re-probe: 0 of 50 answer 503
  CAUSE 2 · QUESTION — local DNS answers NXDOMAIN; public resolvers answer. Only this laptop? (no proof either way)
  ```
- **the 3-minute demo, step by step:**
  1. Deploy a small app to a local container platform (Docker Compose) with two planted faults: a proxy that forwards
     the wrong scheme, and an idle timeout that kills the app.
  2. Run one command: `why-down https://localhost:8080`.
  3. It probes, reads the diff and logs, finds the restart pattern in the backup folder, and prints two PROVEN causes,
     each with its fix.
  4. It re-probes after the fix and shows green.
  5. A third symptom has no proof and is shown honestly as QUESTION.
- **what goes to GitHub as open source:** the CLI, the probe runner, adapters (Docker and one hosted platform), the
  cause-and-evidence schema, and the timeline view.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **probe-and-witness timeline**: timed
  probes plus side-channel witnesses merged into one timeline, with a cause attached. Our own deploy work needed exactly
  this, assembled by hand.
- **sponsor technology we can use honestly:** Nemotron as the model that writes the probe plan and reads the docs; Cursor
  as one executor.
- **what the fleet must build in 144 h:**
  1. probe runner and timeline;
  2. Docker adapter plus logs;
  3. diff and blame reader;
  4. witness readers (files, cron, backups);
  5. the cause step and report;
  6. the demo app with planted faults, README and video.
- **the biggest risk, and how we would know by day 2 that it is failing:** platform adapters eat the week. By the end
  of day 2, the Docker demo must name the restart cause from witnesses alone. If not, cut the hosted adapter.
- **anything that touches internal mechanisms or patents:** none.

## Proposal 2: The README that proves itself

- **track:** The Agent That Can Explain Why (*Reasoning Architecture* also fits).
- **the problem, in one sentence a judge understands:** the install steps in a README break silently on the shell or OS
  a stranger uses, and the maintainer learns about it from an angry issue.
- **who has this problem outside our fleet:** every open-source project with a `curl … | sh` line, and every project
  whose docs drifted from its scripts.
- **the data sources the agent investigates:**
  1. the README and docs: the steps, extracted;
  2. the scripts they call;
  3. fresh environments it starts (debian `dash`, alpine `busybox sh`, ubuntu `bash`, an old bash 3.2);
  4. git history: which commit made the docs and the script disagree;
  5. the release URLs the docs name, checked live.
- **what "explained with evidence" means here — one answer:**
  ```
  STEP 3 · BROKEN on debian (dash) · works on ubuntu (bash)
  Doc       README.md:42  curl -fsSL https://example.org/install.sh | sh
  Run       dash: install.sh:88: Syntax error: "(" unexpected   (exit 2)
  Why       install.sh:88 uses a bash array; the shebang says #!/bin/sh (since commit 7c1e…, "speed up install")
  Proof     the same step with `| bash` passes on debian; reverting 7c1e… passes on dash
  STEP 4 · BROKEN everywhere — https://example.org/get/stable answers 404 (the docs name a channel nobody published)
  ```
- **the 3-minute demo, step by step:**
  1. Point it at a public repo with an install line.
  2. It extracts five steps and runs them in four fresh containers.
  3. A grid shows step × environment.
  4. Click a red cell: the doc line, the output, the offending script line and the commit that caused the drift.
  5. It opens a draft PR that fixes the doc or the script, and re-runs green.
- **what goes to GitHub as open source:** the step extractor, the environment-matrix runner, the drift explainer (blame
  plus bisect over the ███████████ pair), and a GitHub Action.
- **the bounded piece Saphan later takes, and why we would build it anyway:** the **environment-matrix runner for
  install paths**. Our own preview release needed it: the installer met dash, bash 3.2 and an unpublished channel, each
  found by hand, after the fact.
- **sponsor technology we can use honestly:** Nemotron (step extraction and explanation, runnable locally); Cursor as
  one executor.
- **what the fleet must build in 144 h:**
  1. step extractor;
  2. container matrix runner;
  3. URL and release checker;
  4. blame/bisect drift explainer;
  5. grid report plus draft-PR writer;
  6. demo, README and video.
- **the biggest risk, and how we would know by day 2 that it is failing:** step extraction from free-form READMEs is
  unreliable. By day 2 it must extract and run the steps of five real public READMEs unchanged. If not, require a fenced
  `install` block and say so.
- **anything that touches internal mechanisms or patents:** none.

## The question I could not answer this month

*"Why does the owner's password save hang, and why does production answer 503 in bursts?"* (2026-10-13, the first live
day of the portal.)

The answer existed in four places, and none of them was where I looked first:
1. The Worker code forwarded `https://` to the container, and the platform refuses that. It showed only in `wrangler
   tail`, because the Worker's `catch` was silent: review r1 had named it, but filed it inside another item.
2. The container's own stdout was not visible anywhere.
3. The witness that cracked the second fault was **the backup bucket**. Dumps meant to land every 5 minutes landed every
   minute, so the container was being restored from R2 on every start: no inactivity timeout.
4. The fix is in purchase commit `9886a55`, with orders `TP/orders/ORDER_W9c_…` and `ORDER_W9c_r2_…`. A separate
   NXDOMAIN on the owner's laptop was a local DNS cache, not production.

Assembling it took hours of the owner's first live day, with him waiting at the sign-up page (`TP/orders/ORDER_W9c_r2_…`: "the owner's password claim hung exactly so").
Record: `TP/TRAIN_ORDER.md`, *"LIVE FROM develop (2026-10-13)… Two production defects found and fixed on the way."*

---

## REDACTED

Parts of this file are blacked out (█). Nothing was deleted or reworded: every other word is the author's, as written on 2026-10-16, and each black bar has the length of what it covers.

| what was blacked out | times | what it was |
|---|---|---|
| internal locations | 2 | where a document or log lives on our machines or in our private record |
| locations in our private code | 3 | file names and line numbers in Saphan Studio's private source code |
| platform mechanisms | 2 | how our platform signs, verifies or licenses, which we have not published |

Stream and track names (slugs), commit hashes, test names and law numbers were left as written.
