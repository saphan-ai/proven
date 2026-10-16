# planning/answers/

On 16 October 2026 the owner sent one question to the fleet: [`../QUESTION_TO_THE_FLEET.md`](../QUESTION_TO_THE_FLEET.md). It asked every master and the scribe to critique his candidate topic first, from their own reviews, and then to propose something better if they could.

**Eleven answers came back the same evening: ten masters and the scribe.** The eleven files were finished within sixteen minutes of each other. Each master answered from the record of its own track: the streams, the rounds, the reviews, the counts, the commits. Most of them say it in the first lines: *"Everything below is from the record; where I did not measure, I say so."*

The answers are published **verbatim**. Nothing was reworded and nothing was deleted. Where an answer named something internal, it is blacked out (█), and every file ends with a `REDACTED` section that says what was blacked out and how many times.

There is no code in this repository yet. This folder is the best picture we have of how the fleet works before the build starts: how the agents talk, how they disagree, how a disagreement ends, and why the tool we are about to build is needed.

## The answers

| file | author | the track it answered from | its own proposal |
|---|---|---|---|
| [`scribe.md`](scribe.md) | the scribe | the whole record: about 8,500 documents, 92,390 run rows | *Is this red mine?* A failure explainer |
| [`master-a-finished-leg-is-returned-by-the-engine-not-by-the-master.md`](master-a-finished-leg-is-returned-by-the-engine-not-by-the-master.md) | master | how a finished agent run is brought home | *Why did this test fail: regression, flake, or the box?* |
| [`master-saphan-agent-cursor.md`](master-saphan-agent-cursor.md) | master | Cursor as a backend of the fleet | *Claim checker for pull requests* |
| [`master-saphan-agent-onion.md`](master-saphan-agent-onion.md) | master | the sandbox and its proxy to model hosts | *Would your tests notice if this fix vanished?* |
| [`master-saphan-message-is-one-verb-for-every-role-and-rides-the-bridge.md`](master-saphan-message-is-one-verb-for-every-role-and-rides-the-bridge.md) | master | messages between roles; the secret scanner | *Why is this test red?* |
| [`master-saphan-studio-purchase.md`](master-saphan-studio-purchase.md) | master | the purchase portal and its production deploy | *Why is production down?* and *The README that proves itself* |
| [`master-the-bridge-commits-without-a-shared-lock.md`](master-the-bridge-commits-without-a-shared-lock.md) | master | git hooks and trunk protection | *Regression or flake? A failure that explains itself* |
| [`master-the-console-is-one-app-served-from-the-binary.md`](master-the-console-is-one-app-served-from-the-binary.md) | master | the web console | *Threat-to-test design review* and *Screen-truth checker* |
| [`master-the-fleet-commits-as-itself-and-writes-to-the-customers-templates.md`](master-the-fleet-commits-as-itself-and-writes-to-the-customers-templates.md) | master | the fleet's own git identity | *Why is this test red? Triaging CI failures with evidence* |
| [`master-the-gateway-is-the-only-door-to-the-store.md`](master-the-gateway-is-the-only-door-to-the-store.md) | master | the gateway in front of the store | *Why did the merge break?* and *Flake, pre-existing or mine?* |
| [`master-what-the-docs-refresh-found-in-the-engine.md`](master-what-the-docs-refresh-found-in-the-engine.md) | master | 97 places where the engine contradicts its own help | *Does the help tell the truth?* |

The words used below (master, stream, round, review, verify, unplug, land) are explained in [`../../SAPHAN.md`](../../SAPHAN.md).

## 1. The tests are green and the review says no

This is the case the answers come back to again and again. The author's tests pass. A reviewer, who did not write the code, rejects the change, and proves why.

- **Three times in two days.** On one track, the master's own verification was green three times, and three times the review said `reject`. Each time the reviewer was right, and the next round fixed what it found. The master wrote: *"In my record the common and expensive case is 'the test suite is green and the reviewer is right'."* ([finished-leg](master-a-finished-leg-is-returned-by-the-engine-not-by-the-master.md))
- **Sixteen green rows that tested the wrong thing.** A change to stop deletions of the trunk came with 16 tests, all green. The reviewer removed the new check, and all 16 stayed green: git itself refused the deletions one step earlier. *"The tests proved git, not the fix."* ([bridge lock](master-the-bridge-commits-without-a-shared-lock.md))
- **Thirteen rounds, seven reviews, seven rejections.** Every rejection was a run, not an opinion, and every one held up. Each fix created the next defect in its own blast radius. ([fleet identity](master-the-fleet-commits-as-itself-and-writes-to-the-customers-templates.md))
- **The author's proof was wrong, and the author was the master.** The master "proved" a fix by breaking it and watching a test go red. The reviewer showed that the test was guarding a copy of the code, not the code that runs, by changing exactly the production line and keeping the suite green. ([sandbox](master-saphan-agent-onion.md))
- **"All five acceptance items pass by named tests."** The reviewer ran the package the change touched and found two reds the change had introduced. ([gateway](master-the-gateway-is-the-only-door-to-the-store.md))
- **A secret the scanner let through.** The reviewer wrote one test with a credential shaped in a way the fix had not considered. It passed before the change and failed after it. It was the only finding on that stream that nobody could argue with. ([messages](master-saphan-message-is-one-verb-for-every-role-and-rides-the-bridge.md))

## 2. The agents disagree, and they say so

The scribe counted, across the whole record:

| what | count |
|---|---|
| review returns in the record | 1,087 |
| streams with two or more reviews | 207 |
| designed blind pairs (two model families, one order, the same commit) | 14 |
| … of which the two reviewers disagreed | **11 of 14** |
| reviews by two model families of the same target within 30 minutes | 41 |
| … with different verdicts | **31 of 41 (76 %)** |
| reject rate by model family, over 2,075 readable review runs | codex 80 % · claude 27 % · qwen 10 % |

The disagreements run in every direction, not only between two reviewers:

- **An executor refuted its master with a test.** The order stated how one part worked. The executor wrote a test that showed it did not work that way. ([gateway](master-the-gateway-is-the-only-door-to-the-store.md))
- **An executor refused to build on a wrong premise.** The order told it to reuse two things. It stopped as `blocked` and showed that neither existed. ([sandbox](master-saphan-agent-onion.md))
- **An executor refuted two masters.** Two masters had written down the cause of a failed landing. The executor read the code path and showed that the cause they named could not happen. ([bridge lock](master-the-bridge-commits-without-a-shared-lock.md))
- **A test author and an implementation disagreed**, and a recording of real traffic settled it. ([Cursor](master-saphan-agent-cursor.md))
- **The owner overruled a master's fix.** The master proposed a remedy for a security finding. The owner saw that it did not work. ([console](master-the-console-is-one-app-served-from-the-binary.md))
- **The advisor's own list was re-measured**, and it was wrong in both directions: some defects it listed did not exist, and some it dropped were real. ([docs refresh](master-what-the-docs-refresh-found-in-the-engine.md))

And the agents name their own mistakes, in their own answers:

- *"MASTER WAS WRONG."* (a ledger entry, quoted in [messages](master-saphan-message-is-one-verb-for-every-role-and-rides-the-bridge.md))
- *"W7, the worst one, and it is mine."* A merge the master made silently deleted 51–52 files of the trunk. ([Cursor](master-saphan-agent-cursor.md))
- *"It needed three review rounds, one of them because my own proof was wrong."* ([sandbox](master-saphan-agent-onion.md))
- *"My first explanation of that one … was wrong. My own next two runs refuted it, and I withdrew it."* ([finished-leg](master-a-finished-leg-is-returned-by-the-engine-not-by-the-master.md))
- *"So the cause we had both written down was wrong, and the true one is still unmeasured."* ([bridge lock](master-the-bridge-commits-without-a-shared-lock.md))

## 3. How a disagreement ends

Not by a vote, and not by the loudest agent.

1. **By a run.** A probe, a reproduction, a test, or breaking the code on purpose. Every answer says it in some form: *"every disagreement that was settled was settled by a run, never by a second opinion."* Before a fix, the executor must first reproduce each finding, and it may come back with "refuted". That is our law #44: an order must be able to refute its author.
2. **By a rule, when there is no evidence either way.** The owner's rule for two reviewers is *"a reject from either is a reject."* The scribe found that it decided every designed pair. Only 2 of the 11 disagreeing pairs could be settled by measurement, and both times the reviewer who said no was right.
3. **By the owner, when the question is what we want rather than what is true.** One stream went through 21 review rounds, and every finding was true and proven. It ended when the owner ruled on scope: what that layer promises and what it does not. On another track, about 10 of 71 review conditions needed a person's decision; the owner's rulings are quoted there in Polish. ([bridge lock](master-the-bridge-commits-without-a-shared-lock.md), [purchase](master-saphan-studio-purchase.md))
4. **Never by silence about what was not checked.** A review that lists anything under `Unrun` cannot say `accept` (law #53). On 16 October that rule made the owner choose one more round over a landing. It cost about three hours.

The scribe also reports the one time this failed: a real conflict that closed with no arbitration document.

## 4. Why the tool is needed

Every answer ends with *"the question I could not answer this month"*. At least five of them use the same words: the answer was spread over **four places**. Finding it took *"most of an hour"*, *"hours"*, *"most of an evening"*. Some of the questions:

- Why did 27 agent runs die on one day with one line of error, and is the cause still there?
- Why does every local agent on one machine answer 502 after an install that passed seven rounds and several reviews?
- Why did a merge delete half of a package that the change never touched?
- Why does production answer 503 in bursts, on the first live day, with the owner waiting at the sign-up page?
- Why has nothing moved for two hours?

The scribe measured the size of it. **Every one of 3,030 failure rows in our store has an empty cause.** For 52 runs the engine itself marked as unexplained, the cause was readable on disk in 36 cases. *"In each one, the explanation existed in a second data source (stderr, git, raw test output, host state), and the tool that reported the outcome never read it. That is exactly the shape of the Explain Why track."*

## 5. What the fleet asked us to change

Every answer agrees with the core of the candidate: a finding counts only when a run proves it. Every answer also says that *"fails on the change, passes on the base"* is not enough on its own, and says why from its own record.

| what the answers ask for | how many of the 11 answers |
|---|---|
| "could not run" as its own outcome, with the reason (`UNRUN`, `UNMEASURED`, `COULD NOT RUN`) | 11 |
| breaking the code on purpose (mutation, unplug) as evidence, not only a failing test | 9 |
| running the change merged with today's trunk, not only against its own base | 5 |
| a label for "the evidence is clear and the choice is a person's" (`DECISION`, `NEEDS-A-RULING`, `PROVEN-OUT-OF-SCOPE`) | 4 |

Almost every answer also asks that every failing test is run again **alone, on the same machine, on both sides**, before it is called a finding. Their false alarms came from the environment far more often than from a model.

On two model families, the fleet pushes back. Several masters write that the second opinion that decided things was never a second model, but *"a second instrument"*, *"a second role"* or *"a second environment"*. The scribe's numbers point the other way: two families disagree three times out of four. Both views are here, and both are measured.

The fleet sent fourteen proposals. Six of them are the same tool seen from different tracks: *why is this test red: my change, already broken, flaky, or the machine?* Most of the authors offer it as the evidence runner of the candidate rather than as a rival.

The owner reads these answers and decides what goes into the first order of the build. That order will be published in [`../../fleet/orders/`](../../fleet/orders/) on 22 October.
