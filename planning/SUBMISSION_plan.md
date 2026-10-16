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

## Description (draft, leading candidate)

proven: proof-carrying code review. AI code reviewers give opinions, and a developer has to check each one. A false alarm costs as much time as a real bug. We built a pull-request reviewer that has to prove what it says. Two reviewers from different model families read the change. Every finding becomes either a test run in a sandbox on the base and head commits, or a citation to the code history, an issue or the docs. Each finding is marked PROVEN, QUESTION or DROPPED, with the evidence attached. When the reviewers disagree, the evidence decides, not a vote. Each finding also says why the code is the way it is, so the reviewer explains why something is wrong, not only that it is.

How it was built is part of the entry. A fleet of AI agents, with Cursor as one of the executors, wrote all of the code in six days. One person set the goals, made the decisions and reviewed every change. He wrote no code. The repository holds, next to the code, the work orders, status reports and reviews behind each change, so anyone can follow why each change was made and who checked it. NVIDIA Nemotron is one of the reviewer models.

Why here: review with evidence was already on the backlog of Saphan Studio, our control plane for agent fleets. We held it back so it could be built during the hackathon, from zero and in public. Saphan Studio stays closed and is a pre-existing component that only runs the fleet. This part goes to the community: the reviewer is open source under Apache-2.0, and Saphan will use the same library for its own pull-request review.

## Pitch deck (PDF, 10 slides)

1. Title, one line, team.
2. The problem: reviewers that cannot show their evidence.
3. Why this topic: the fleet proposed it. How many proposals came in, how the owner chose, one or two quotes from the answers in `answers/`.
4. What it does: one finding with its evidence.
5. How it works: two reviewers, evidence runners, the PROVEN / QUESTION / DROPPED rule.
6. Results on the demo set: planted bugs found and proven, bait dropped, false alarms.
7. How it was built: numbers from the repo (pull requests, orders, reviews, changes sent back, owner decisions).
8. Sponsor technology: Nemotron as a reviewer, Cursor as an executor and what it built.
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
