# planning/

How the topic of this project was chosen, before any code. Everything here was written before the build window opened on 22 October 2026. The story in order is below, under "How the topic was chosen"; every conversation since is in [`CONVERSATIONS.md`](CONVERSATIONS.md).

| file | what it is | written by |
|---|---|---|
| `CONVERSATIONS.md` | **start here:** the diary, one line per hop of every conversation page, each linking to its hop | the advisor |
| `CONVERSATION_2026-10-16.md` | choosing the topic: who said what to whom, hop by hop, with the owner's words in the original Polish | the advisor |
| `CONVERSATION_2026-10-17.md` | preparing the machines for the Cursor executors: what the advisor measured, reasoned and ordered, and the four decisions left to the owner; then the design reshaped for the sponsors' technology, and the infrastructure grown with it (hops 11 to 13) | the advisor |
| `CONVERSATION_2026-10-18.md` | a local NVIDIA guardrail model: what the advisor measured, the train order made public, the guardrail on by default in proven, the owner's correction of a redaction, and this diary | the advisor |
| `CONVERSATION_2026-10-19.md` | the front page rewritten for readers who have never heard of agent fleets | the advisor |
| `EVENT.md` | the hackathon: dates, tracks, judging, and the owner's rules for taking part | the advisor, from the owner's decisions |
| `QUESTION_TO_THE_FLEET.md` | the question the owner sent to the masters and the scribe: constraints, a request to critique the owner's candidate, and a template for proposals | the advisor |
| `PROPOSAL_advisor.md` | the advisor's own answer on the same template: *proof-carrying code review* | the advisor |
| `SUBMISSION_plan.md` | what we submit and how: description, pitch deck, video, publishing rules | the advisor |
| `answers/` | the fleet's eleven answers, verbatim, and a README on what they show | ten masters and the scribe |

The fleet's answers came back on 16 October, the same evening as the question, and are published verbatim. The same rule as for every other document applies: credentials, internal host names and details of our private infrastructure are removed, and a document says where something was removed.

`EVENT.md`, `QUESTION_TO_THE_FLEET.md` and `PROPOSAL_advisor.md` are published verbatim, as the fleet read them on 16 October. Internal details are blacked out (█), not deleted, and each file ends with a `REDACTED` section. `SUBMISSION_plan.md` is a living plan and changes as we go.

## How the topic was chosen (16 October 2026)

*This story stood in the main README until 19 October 2026, when the front page was rewritten for a wider audience. It is moved here unchanged, except for its links.*

The topic was chosen in one day: in one conversation between the owner and the advisor, and in a question sent to the fleet. The owner's words are translated from Polish. The whole exchange, hop by hop, with his words in the original Polish, is in [`CONVERSATION_2026-10-16.md`](CONVERSATION_2026-10-16.md).

**Who is who.** The **owner** is the one person on the team. The **masters** are the agents that orchestrate the fleet's work. The **scribe** is the agent that keeps the fleet's record. The **advisor** is an agent that works with the owner on plans and documents. Everyone except the owner is an AI agent.

1. **Where we started.** On 12 October the owner decided that the fleet would enter, and set the rules: the fleet builds, he writes no code, Cursor joins as an executor, all project code is new from 22 October, and everything is recorded and published after his review. The organizers confirmed that our own tools may run the fleet as a pre-existing component.
2. **The idea came from our own work.** The owner and the advisor were discussing how Saphan Studio could use its integration with GitHub and Gitea. The owner asked: *"What if we built a feature like the one Claude Code has, but plugged into our own infrastructure and process: pull-request review on GitHub?"*
3. **The owner set the frame.** The topic should come from the fleet, through a survey of the masters and the scribe. The project should be open source on GitHub, with one part taken into the Saphan backend. Everything is written in English.
4. **The owner's candidate.** While the question to the fleet was being written, he added: *"Write your own code review with different models?"*
5. **The question to the fleet.** The advisor wrote [`QUESTION_TO_THE_FLEET.md`](QUESTION_TO_THE_FLEET.md): six constraints, a request to critique the owner's candidate first, up to two proposals per author on one template, and one question for everyone: which question about your own work could you not answer this month? The owner pasted it to the ten masters and the scribe. The deadline was 19 October; eleven answers came back the same evening, between 20:03 and 20:19 UTC.
6. **The advisor's proposal.** The advisor answered on the same template in [`PROPOSAL_advisor.md`](PROPOSAL_advisor.md): *proof-carrying code review*. A finding counts only with evidence: a test that fails on the change and passes on the base, or a citation from the code history, an issue or the docs. Two reviewers from different model families. When they disagree, the evidence decides, not a vote.
7. **The submission form.** The owner read the form and decided what we send: a pitch deck with the fleet's reasoning, a video, and this repository, public from today, with the documents behind every change. The plan is in [`SUBMISSION_plan.md`](SUBMISSION_plan.md).
8. **Why it waited for the hackathon.** Review with evidence was already a feature we wanted in Saphan Studio. The owner: *"We can write in the submission that it is already in the backlog and waits for the hackathon to be developed. We won't open the whole platform, but we will give part of it to the community."* The same day the advisor filed the feature in our backlog, held until 22 October, so that nothing overlapping this project is built before the window opens.
9. **Publish the conversation now.** The owner: *"There is no code yet, but there is already the conversation about what we will write."* That is why `planning/` is here before 22 October.
10. **The name, and with it the topic.** The advisor proposed two names and advised against using the name of the Saphan feature, so that it stays clear what is open. The owner: *"proven is a good name. That's what we do."* The original plan was that the owner picks from the fleet's answers by 20 October. He named it at 20:12 UTC, while the last two answers were still being written. The answers are published verbatim in [`answers/`](answers/).
11. **How we publish.** The owner pushes once a day, at 22:00 Bangkok time.
12. **The sponsors reshape the design (17 October).** The owner read the hackathon's technical guide and found that our design did not meet it: *"it does not meet the requirements, because we have to use all the sponsors' technologies."* Every entry must integrate Zetaris and Meterless, and NVIDIA wherever it helps. Zetaris became the layer every piece of evidence is queried through, and Meterless the reviewer's memory and state. The owner also widened what the reviewer reads: *"Once we get the platform we can go wild. For example, pulling CVE databases as data, other data sources for the evaluating agent."* He chose Go, behind adapters: *"Later we will use it in Saphan. That is the point of this game."* The track stays the same: *"There will be a better explanation."* The hop-by-hop record is in [`CONVERSATION_2026-10-17.md`](CONVERSATION_2026-10-17.md), hops 11 and 12.
