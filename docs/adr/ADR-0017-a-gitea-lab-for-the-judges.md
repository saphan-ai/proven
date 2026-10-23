# ADR-0017: proven ships a Docker Compose lab with Gitea and Gitea Actions, already configured, so a judge sees it review a pull request with nothing to set up

**Status:** accepted · **Date:** 2026-10-20
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-20
**Amended by:** [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md), point 7 (2026-10-21): the comments of point 4 are posted by reviewdog from Proven's SARIF, not by an adapter of Proven's own · [ADR-0022](ADR-0022-gitea-on-premises-is-an-equal-scenario.md) (2026-10-21): Gitea on premises is a scenario equal to GitHub; this lab is that scenario, packaged · [ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md) (2026-10-23): Proven's own image is tagged `latest` (point 6); the runner uses Docker in Docker (Consequences); the lab is started by `make up`
**Follows:** [ADR-0001](ADR-0001-go-behind-adapters.md) (adapters and sample mode), [ADR-0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md) (the pipeline shape)
**Trigger:** the owner, on what the judges receive ([`planning/CONVERSATION_2026-10-20.md`](../../hackathon/planning/CONVERSATION_2026-10-20.md), hop 7)

## Context

The owner:

> *"I want to ship a docker-compose with a local Gitea install with proven configured in it, so the judges will get a full working lab with the Gitea UI and can see it actually working. Nothing needed. Gitea Actions configured. If we have time we can also make a Jenkins example. But Gitea is a MUST."*

Gitea is a self-hosted Git service with pull requests and a web UI. Its CI, Gitea Actions, runs workflows written in GitHub Actions' syntax on a runner of its own, so the workflow of [ADR-0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md) can run there unchanged. Gitea has no code-scanning page for SARIF.

## Decision

1. **The repository ships a lab**: one Compose file that starts Gitea, a Gitea Actions runner registered to it, and a demo repository with an open pull request and proven's workflow. One command starts it; the judge opens Gitea in a browser and sees proven's review on the pull request.
2. **Nothing is needed beyond Docker.** No account, no key, no manual step. The lab runs proven in sample mode by default ([ADR-0001](ADR-0001-go-behind-adapters.md)). An optional environment file with real keys turns on the live services.
3. **The workflow in the lab is the same file a GitHub user would use** ([ADR-0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md)). The lab is how we show that one workflow runs on both.
4. **proven posts its review on the pull request through Gitea's API**, behind its own adapter. The SARIF and JSON reports are kept as the run's artifacts, because Gitea has nowhere to show SARIF.
5. **The lab binds to the judge's own machine only** (localhost). Its admin login is generated at first start and printed once; no password is written into the repository.
6. **Every image in the lab is pinned to a version**, never `latest`, so the lab a judge starts is the one we rehearsed. The versions are settings ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). **The one exception is Proven's own image, tagged `latest`**, with a dated tag pushed beside it ([ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md), the owner, 2026-10-23).
7. **The lab is rehearsed the way a judge runs it** before every daily push: from a clean checkout, one command, the review appears.
8. **Gitea is required; Jenkins is optional.** A Jenkins example running the same `proven ci` command is built only if time is left after the Gitea lab passes its rehearsal.

## Consequences

- A judge sees proven work in a pull request, not just in a terminal.
- The runner needs to start containers. Giving it the host's Docker socket is the simplest way and gives the lab's jobs control of the judge's Docker; running Docker inside the lab is safer and heavier. **Chosen: Docker in Docker** ([ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md), the owner, 2026-10-23); the lab's jobs get no control of the judge's Docker.
- Sample mode shows the whole review without Zetaris or NVIDIA. The live path needs keys, and the lab's README says which keys and what each one turns on.
- The Zetaris platform is not part of the lab by default: it is several GB and needs the sponsor's registry sign-in. Whether the lab gets an optional Zetaris profile is not part of this decision.
- The lab is one more thing to keep working through the build week. Its rehearsal is the daily judge rehearsal of the infrastructure order ([`proven-cursor-infra`](../../hackathon/fleet/orders/TRAIN_ORDER_proven-cursor-infra.md), W9).
