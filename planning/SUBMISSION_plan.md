# Submission plan — Open Agent Hackathon 2026

Status: draft by advisor #7, 2026-10-16 ("why here" paragraph added on the owner's request, same day). The topic is not chosen yet (owner, by 2026-10-20). The description below is written for the leading candidate and changes if the topic changes. Every claim in it must be true on the day it is submitted.

## The form

| Field | Required | Limit | What we send |
|---|---|---|---|
| Description | yes | 150–2000 characters | text below |
| Video | yes | max 4 min | link (YouTube unlisted or Vimeo) |
| GitHub solution | yes | repo link | `saphan-ai/proven` (https://github.com/saphan-ai/proven): README, licence and planning now; code from 22.10; one push a day at 22:00 Bangkok (15:00 UTC) |
| Pitch deck | no | link or file | PDF, 10 slides |

Up to three submissions, one counts. Plan: a complete safety submission on 26.10 evening, the final one on 27.10 before 20:00 UTC (the form closes 23:45 UTC).

## Description (form field, 150–2000 characters; 1,864 — version of 2026-10-16)

proven is a pull-request reviewer that has to prove what it says.

The problem. AI code reviewers read a diff and write comments. A finding is still an opinion: nothing is run to prove it, reviewers disagree, and a developer must redo the work to learn which comment is real. No report says what was not checked, so "no findings" can mean "nobody looked". When two model families reviewed the same change in our own agent fleet, they disagreed in 31 of 41 cases.

How it works. Two reviewers from different model families, one of them NVIDIA Nemotron, propose findings. A finding is published only with evidence: a test the agent writes that fails on the change and passes on the base, run in a sandbox on both commits, repeated and in isolation; the changed lines broken on purpose to see if any test notices; citations from the code history, the issue and the docs that say why the code is the way it is. Each finding is PROVEN, QUESTION or DROPPED. What could not be run is listed with the reason. When reviewers disagree, the evidence decides, not a vote.

What is different. Other reviewers give opinions; proven gives evidence you can run. It says what it did not check, and it tells "your change broke this" from "already broken" and "flaky".

How it is built. Four Cursor agents write all the code, 22–27 October. One person decides and merges, and writes no code. The fleet runs on Saphan Studio, our control plane for agent fleets, a pre-existing component. The repository already holds the fleet's eleven answers on what to build, verbatim. From 22 October every change arrives with its order, report, review and decision. The architecture will change as we learn, and every change of design is a documented decision. Apache-2.0.

The last test. On the final day, proven reviews the last change of its own build. Its report is part of the submission.

## Project description, long version (2,941 characters, 2026-10-16; not for the form field)

proven is a pull-request reviewer that has to prove what it says.

The problem. AI code reviewers read a diff and write comments. The best of them add linters and static analysis. But a finding is still an opinion: nothing is run to prove it, two reviewers often disagree, and a developer has to redo the work to learn which comment is real. A false alarm costs as much time as a real bug. And no report says what was not checked, so "no findings" can mean "nobody looked". Our own fleet of coding agents has lived this for months: tests green, review blocking, and the reviewer right. When two model families reviewed the same change, they disagreed in 31 of 41 cases.

How it works. Two reviewers from different model families, one of them NVIDIA Nemotron, read the change and propose findings. A finding is published only with evidence. The agent writes a test that fails on the change and passes on the base, and runs it in a sandbox on both commits, repeated and in isolation, so a flaky test or a broken machine is not blamed on the author. It breaks the changed lines on purpose to show whether any test notices. It cites the code history, the linked issue and the docs to say why the code is the way it is. Every finding gets a label: PROVEN, with its evidence; QUESTION, when the evidence is not enough; DROPPED, when the evidence says the code is right. What could not be run is listed with the reason, and a clean verdict is impossible while that list is not empty. When the reviewers disagree, the evidence decides, not a vote.

How it differs from what is on the market. Other reviewers give opinions; proven gives evidence you can run yourself. It tells you what it did not check. It separates "your change broke this" from "this was already broken" and "this test is flaky". It explains why the code exists, not only what is wrong with it.

How it is built. Four Cursor agents write all the code between 22 and 27 October. One person, the owner, decides and merges, and writes no code. The fleet runs on Saphan Studio, our control plane for agent fleets, a pre-existing component. Review with evidence was already on its backlog; we held it back to build it here, from zero and in public. Saphan Studio stays closed; this part goes to the community under Apache-2.0.

Built in the open. Before any code, the repository already holds the question the owner put to the fleet and eleven answers from the fleet's agents, verbatim. They show green tests with a blocking review, agents refuting each other and themselves, and disagreements settled by a run, a rule or the owner. From 22 October every change arrives with its order, status report, review and decision. The architecture will change as we learn, and every change of design will be a documented decision with its reason.

The last test. On the final day, proven runs on its own repository and reviews the last change of the build. Its report is part of the submission.

## Pitch deck (PDF, 10 slides)

1. Title, one line, team.
2. The problem: reviewers that cannot show their evidence.
3. Why this topic: the fleet proposed it. How many proposals came in, how the owner chose, one or two quotes from the answers in `answers/`.
4. What it does: one finding with its evidence.
5. How it works: two reviewers, evidence runners, the PROVEN / QUESTION / DROPPED rule.
6. Results on the demo set: planted bugs found and proven, bait dropped, false alarms.
7. How it was built: numbers from the repo (pull requests, orders, reviews, changes sent back, owner decisions).
8. Sponsor technology: Nemotron as a reviewer; four Cursor agents as the only executors, and what they built.
9. What it does not do yet.
10. Links: repo, video.

Built 26–27.10 from the repo history. A skeleton with slides 1–3 can be ready by 21.10.

## Video (max 4 min, target 3:30)

| Time | Content |
|---|---|
| 0:00–0:20 | The problem in one sentence and one bad AI review comment. |
| 0:20–2:20 | Live run on a pull request with planted bugs and one bait: findings appear, tests go red on head and green on base, the bait is dropped. |
| 2:20–3:10 | How it was built: the console with this project's orders, reviews and one change sent back; a short clip of a coding agent at work. |
| 3:10–3:30 | Repo, licence, what is next. |

Footage from before 22.10 shows only pre-existing tools. Use it only as background for the "how it was built" part.

## Publishing rule for fleet documents

The owner's rule from 12.10 applies to every document that enters the repo:

1. Fleet documents for this project are written for the public repo from the start. They are not copies of internal documents.
2. Before each merge: a secret scan (gitleaks) and a deny-list check (internal host names, machine names, internal paths, private repository names) run on the pull request. The owner writes the check, not the fleet.
3. No document describes how the engine binds orders to policies, pins versions or works inside. A public order holds the goal, the properties to hold, the acceptance test and the evidence. Nothing more.
4. Workflows in `.github/` are written or approved by the owner only. Agent pull requests must not change them.
5. Video and screenshots come from a clean demo instance. If the real console is used, blur host names and paths before export.
