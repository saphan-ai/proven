# ADR-0022: Gitea on the team's own premises is a scenario equal to GitHub, not only the judges' lab

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Amends:** [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) (the Gitea lab for the judges) and [ADR-0021](ADR-0021-a-self-hosted-runner-demo-set-up-by-one-script.md), point 9
**Follows:** [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) (Nightshift), [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) (self-hosted runners first)
**Trigger:** the owner, after the GitHub runner demo ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../hackathon/planning/CONVERSATION_2026-10-21_advisor.md), hop 15)

## Context

[ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) brought Gitea in as a lab the judges can start with nothing set up, and [ADR-0021](ADR-0021-a-self-hosted-runner-demo-set-up-by-one-script.md) described it as the way for a judge without a GitHub account. The owner (translated from Polish):

> *"Exactly so. But Gitea as an equal on-premises scenario. That is important. Not everyone has their source code on the internet."*

Many teams keep their code on their own servers: regulated industries, defence, telecoms, anyone whose code may not leave the building. For them a reviewer that needs GitHub, or sends code to a model in someone else's cloud, is not an option at all.

## Decision

1. **Proven has two equal scenarios**, documented, built and rehearsed with the same care:

   | | GitHub | Gitea, on premises |
   |---|---|---|
   | where the code lives | GitHub | the team's own Gitea server |
   | the runner | GitHub-hosted, or the team's own (primary for live mode, [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md)) | the team's own Gitea Actions runner, always on their premises |
   | the workflow | Proven's workflow file | **the same file**: Gitea Actions reads GitHub Actions' syntax |
   | comments on the pull request | reviewdog, `github-pr-review` | reviewdog, `gitea-pr-review` |
   | SARIF | GitHub code scanning | kept as the run's artifact; Gitea has no page for it |
   | models | Nightshift on the team's network, or keys as secrets | Nightshift and local models on the team's network |
   | evidence | Zetaris on the team's machine | Zetaris on the team's machine |

2. **In the Gitea scenario, nothing needs to leave the team's network** except what the team chooses: Gitea, the runner, Proven, Nightshift with local models, and Zetaris all run on its own machines. Where Nemotron is reached through NVIDIA's hosted API, that is the one call that leaves, and the documentation names it.
3. **The evidence sources are public data that can be fetched once and kept.** A team that allows no outside call loads OSV, CISA KEV and FIRST EPSS from copies it brings in, and the review says how old its copy is. How the sources are loaded is the next decision.
4. **The judges' lab is the Gitea scenario, packaged.** The same Compose file a team uses to try the on-premises setup is the one the judges start ([ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md)); it is not a separate demo.
5. **Both scenarios are rehearsed before every daily push**: the Gitea lab from a clean checkout, and the GitHub runner demo ([ADR-0021](ADR-0021-a-self-hosted-runner-demo-set-up-by-one-script.md)). A change that breaks either is not merged.
6. **The README and the documentation present both side by side**, with the same depth, from installation to the first review.

## Consequences

- The front page says Proven works where the code is: on GitHub, or entirely on a team's own servers.
- Every feature is checked on both. Anything that works only on GitHub, such as code scanning's page for SARIF, says so and gives the Gitea equivalent.
- Fully offline use depends on local models being good enough for the review; whether they are is not measured yet.
- More to keep working through the build week: two rehearsals a day instead of one.
