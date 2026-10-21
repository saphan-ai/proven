# ADR-0021: The demo includes a self-hosted GitHub runner in Docker, which one setup script connects to a judge's repository from only its name and a token

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Follows:** [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) (the self-hosted runner as the primary live path), [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) (the Gitea lab)
**Amended by:** [ADR-0022](ADR-0022-gitea-on-premises-is-an-equal-scenario.md), point 9: Gitea is not only for judges without a GitHub account; it is an equal, on-premises scenario
**Trigger:** the owner, on showing the self-hosted runner to the judges ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../planning/CONVERSATION_2026-10-21_advisor.md), hop 13)

## Context

[ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) made a team's own runner Proven's main way to run live, so that model keys stay on the team's network. The owner (translated from Polish):

> *"We can show the self-hosted runner in the demo too. It will be in Docker, and a setup script will configure the GitHub repository to send jobs to it. What the judge will need to configure the demo is the repository's name on GitHub and a key for it, so the script configures them automatically if the judge does not want to read the documentation."*

What GitHub requires, from its [documentation on adding self-hosted runners](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/add-runners): the repository's owner adds a runner; the runner's `config` script needs the repository's URL and a registration token, and *"the token expires after one hour"*; runners can also be added through GitHub's REST API. GitHub recommends self-hosted runners only for private repositories, because forks of a public one can run code on them.

## Decision

1. **The judge gives two things**: the repository's name (`owner/name`) and a GitHub token for it. Nothing else is asked for; every other value has a default ([ADR-0013](ADR-0013-no-hardcoded-settings.md)).
2. **One script does the rest**, and prints each step as it goes:
   1. checks the token's access before changing anything, and stops with the missing permission named if it falls short;
   2. asks GitHub's REST API for a runner registration token;
   3. starts the runner in a Docker container on the judge's machine, registered to that repository with a label of its own;
   4. commits Proven's workflow to the repository, with `runs-on:` set to that label ([ADR-0020](ADR-0020-proven-command-line-and-action-interface.md));
   5. pushes a demo branch and opens a pull request, which sends the first job to the runner;
   6. prints where to watch: the pull request, the run, the review's comments and its SARIF.
3. **The token is used and forgotten.** The script reads it from the environment or a hidden prompt, never from a command-line argument; it is not written to disk, not logged, and not passed into the runner container. The runner keeps only its own registration. The script says at the end that the judge may revoke the token.
4. **The token asks for the least it needs.** The documentation recommends a fine-grained token limited to that one repository, with: Administration (read and write, to register the runner), Contents and Workflows (read and write, to commit the workflow and the demo branch), Pull requests (read and write, to open the pull request and post the review), and Metadata (read). The script lists any permission that is missing.
5. **A private repository is recommended**, as GitHub recommends for self-hosted runners. On a public repository the script still works, after it has turned on GitHub's setting that requires approval before workflows from outside contributors run, and it says so.
6. **Proven runs inside the runner's container, not in a container of its own.** The runner's image carries Proven's binary, and the workflow calls `proven ci` in a `run:` step, so the runner needs no access to the judge's Docker.
7. **Model keys stay on the judge's machine.** Live mode reads the Nightshift token and the Zetaris key from the runner container's environment, set by the judge locally; GitHub never holds them. Without them, the demo runs in sample mode.
8. **One teardown command** removes the runner from the repository and stops its container. It needs the token again, because GitHub only removes a runner with a fresh removal token; the script asks for it the same way.
9. **The Gitea lab stays** ([ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md)): it needs no GitHub account at all. This demo is for a judge who wants to see Proven on GitHub itself.

## Consequences

- A judge who does not read the documentation runs one command with two values and watches a review appear on a pull request in their own repository.
- The token in point 4 is powerful while it lives. The script's checks, its silence about the token, and the advice to revoke it are what keep that short.
- The runner in the demo stays registered until teardown, rather than being ephemeral for each job ([ADR-0020](ADR-0020-proven-command-line-and-action-interface.md), point 6): an ephemeral runner must register again after every job, which needs the token again. The documentation says so, and recommends ephemeral runners for anything beyond a demo.
- The script is product work, written from 22 October, and rehearsed from a clean machine the way a judge will run it.
