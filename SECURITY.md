# Security

Proven reviews code changes. To do that it reads a repository it did not write, runs tests from a change nobody has
reviewed yet, asks language models about that change, and queries an evidence platform. Each of these is a place where
the change under review could try to reach further than a review should let it. This page says what Proven protects,
how, and what is still being checked.

What is described here was read from `main` on 23 October 2026 (`2d3126b`). How the code as a whole is checked, piece
by piece, is in [SECURITY_REVIEW.md](docs/security/SECURITY_REVIEW.md).

## What Proven treats as untrusted

| Input | Why it is untrusted |
|---|---|
| The change: its code, its tests, its commit messages | it is what is being reviewed; its author may be careless or hostile |
| Files of the reviewed repository: manifests, lock files, `proven.yaml` | they come with the change |
| A model's answer | a model can be steered by the text of the change it reads |
| An evidence row from Zetaris or a sample table | Proven reports it as evidence, so it must be shown as it was returned, never acted on |

## What Proven protects

- **Your secrets.** The keys Proven uses (the model provider's runtime token, the guardrail key, the Zetaris
  credentials) are settings, read from the environment and never written to the trace, the reports or the terminal.
  `livecheck` names what is missing and never prints a value. See [docs/CONFIGURATION.md](docs/CONFIGURATION.md).
- **The machine Proven runs on.** Code from the change runs only inside a container, and only when a container runtime
  is present ([ADR-0025](docs/adr/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md)).
- **The honesty of the review.** A finding reaches the review only with evidence: a test that fails on the change and
  passes on the base, or a query with its rows. A model's opinion alone is dropped, and the trace says so.

## How

### The change's code runs in a container with no network and no secret

The Prover runs each test twice, on the base and on the change, each time in a fresh container started with:

- `--network none`: the container has no network at all;
- no `-e` or `--env` flag: no environment variable, and so no secret, is passed in;
- the checkout mounted **read-only** as its only volume;
- CPU and memory limits from the settings, and a time limit; a timed-out container is removed;
- `--rm`: nothing is kept after the run.

With no container runtime, the Prover is off and runs no code: the review names it as a failed service instead of
running the test on the host.

### A model is checked before and after every call

The guardrail (NVIDIA Nemotron content safety,
[ADR-0005](docs/adr/ADR-0005-guardrail-on-by-default.md)) is on by default and checks every prompt and every
answer. Turning it off is a setting, and a run without it says so in its trace. Two different model families draft and
challenge each finding ([ADR-0006](docs/adr/ADR-0006-a-skeptic-from-another-model-family.md)), so one steered
answer is not enough to make a finding.

### Proven never writes exploit code

A finding is proved by a test of the expected behaviour, never by an exploit
([ADR-0007](docs/adr/ADR-0007-no-exploit-code.md)).

### Evidence queries carry the change's values as literals

Proven builds every evidence query itself, from a fixed list of query templates
([QUERIES.md](docs/QUERIES.md), [ADR-0032](docs/adr/ADR-0032-proven-builds-every-query-from-a-template.md)). A
model picks a template and gives values; it never writes SQL. Each value, whether it comes from a model or from the
change, is checked against a strict format and goes into the query only as a quoted string literal that holds exactly
that value. A value outside its format is refused, never escaped. Proven does not claim a read-only Zetaris key.

### Proven keeps its memory outside the repository it reviews

Its state and its trace are written in the directory it starts from, and it refuses to start when that directory is
inside the reviewed repository, so a review never writes into the code it reviews.

### Images and the lab

Proven's images are minimal, run as a user that is not root, and are scanned by Docker Scout before every push; the
judges' lab binds to localhost only, generates its admin password at first start, and runs the Prover's containers in
Docker in Docker, never through the host's Docker
([ADR-0026](docs/adr/ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md)). The lab and the image are
being built in the build week; this paragraph describes the decision until they land.

## Running Proven in CI

- Give the workflow's token only what it needs: read the code, write pull-request comments.
- Run Proven on `pull_request`, never on `pull_request_target`, so a change from a fork never runs with your
  repository's secrets.
- Keep its keys in the CI system's secret store; Proven reads them from the environment and never prints them.
- On a self-hosted runner, give the runner a container runtime so the Prover can run tests in isolation; without one
  the Prover stays off rather than running the change's code on the runner
  ([ADR-0021](docs/adr/ADR-0021-a-self-hosted-runner-demo-set-up-by-one-script.md)).

## Weaknesses found and fixed

Each line names the piece of the security review that found it, what did not hold, the change that fixed it, and the
tests that prove the fix.

| Piece | What did not hold | Fixed by | Proved by |
|---|---|---|---|
| 1, the Prover's container | The command run inside the Prover's container could come from a model's answer, so text in the change, passed through a model, could choose what ran in the sandbox. Both reviewers of the piece named it. | `1d23854` (23 October): Proven builds the command itself from the test's file and name; a model's answer that carries a command is refused, and a name that is not a valid test name never reaches the container | `internal/agent/prover/command_test.go`, `internal/agent/skeptic/command_test.go`: `TestAnswerThatCarriesACommandIsRefused`, `TestValidNameReachesTheRunnerAsFileAndNameOnly`, `TestExecuteRefusesAnInvalidNameWithoutCallingTheRunner` and three more |
| 2, secrets | A secret that a service echoed back (the model provider's token in a receipt, the guardrail key in its safety categories, the Zetaris organisation id) could reach the trace, the result, a refusal reason or an error message; a refused organisation id was printed; `livecheck` printed an unknown argument before redacting it; and the refusal tests used empty secrets, so they could not fail. Both reviewers of the piece named parts of it. | `37eabfb` (23 October): one redactor, built from the values actually configured, stands at every writer of the trace, the result and every error; a live review without it is refused | `internal/review/secret_echo_test.go`, `internal/trace/redact_test.go`, `internal/pipeline/secret_test.go`, `cmd/livecheck/secret_test.go`: `TestRuntimeTokenEchoedInTheReceiptIsNotWritten`, `TestGuardrailKeyEchoedInTheCategoriesIsNotWritten`, `TestARefusedSecretIsNamedNotPrinted`, `TestRefusedOrgIDIsNotPrinted`, `TestLivePipelineWithoutARedactorIsRefused` and more |
| 1, the Prover's container | The test ran as the image's default user with no limit on processes, no capability drop, no block on privilege gain and a writable root filesystem; and the container runtime process inherited Proven's whole environment, keys included. Both reviewers of the piece named the first; one named the second. | `175fc39` (23 October): the container runs as an unprivileged numeric user (root is refused), with a process limit, every capability dropped, `no-new-privileges` and a read-only root with a bounded tmpfs; the runtime process and git get a minimal environment | `internal/testrun` hardening and environment tests |
| 3, the reviewed repository as input | With a baseline commit, the reviewed repository's own `base/` and `change/` directories could stand in for the git range, so the agents could read trees that the range did not select; and no read of the repository's content had a size limit. Both reviewers of the piece named the second. | `b3e2648` (23 October): the git range alone decides the trees when a baseline commit is given, and every read of repository content (file reads, the archive, `proven.yaml`, the author record) is size-bounded | `TestReviewP3BaselineReviewIgnoresBaseAndChangeDirectoriesInTheRepository` and the size-bound tests of `internal/change` and `internal/review` |
| 4, evidence queries | A value from the change entered a query as a quoted literal that doubled single quotes but let a backslash, a control byte or invalid UTF-8 through; SQL written by a model reached the evidence platform as text; and the read-only gate counted statements without reading comments or escapes. Both reviewers of the piece named parts of it. | `460da35` (23 October, ADR-0032): every query is built by Proven from a fixed template; each value is checked against a strict format and may hold no quote, backslash, control byte or non-ASCII byte; a model only picks a template and its values; the live gate reads statements strictly | `internal/querytpl` tests, `TestOneSelectRefusesASecondStatementBehindACommentOrEscape`, `TestLocalEngineRefusesTheSameTexts` |
| 0, 1 and 2, smaller items | Proven's state and trace were created readable by other users; a `proven.yaml` that is a link was read through it, and its refusal quoted the linked file; a linked directory could be mounted into the Prover's container, and its image could be pulled at test time; a secret written with escapes was not hidden. | `f9628fd` (23 October): state directories 0700 and files 0600; a linked `proven.yaml` and a linked mount are refused; the image is never pulled at test time; escaped spellings of a secret are hidden | the tests of `internal/review`, `internal/testrun`, `internal/redact` and the state stores in that change |
| 2 and 4, the command line | Text written to the terminal and to standard error could carry a control byte taken from the change; and the one git call of the command line ran with Proven's whole environment, keys included. | `310bce5` (24 October): every control byte from the change is shown escaped on the terminal and on standard error, and that git call gets a minimal environment | `TestTextOutputCarriesNoControlByteFromTheChange`, `TestStderrCarriesNoRawControlByteFromTheChange`, `TestGitMergeBaseChildCarriesNoKeyFromProven` |

## What is still being checked

Proven was built in one week. Its code is being reviewed as a whole, piece by piece, against the properties above; the
plan, the pieces and their status are in [SECURITY_REVIEW.md](docs/security/SECURITY_REVIEW.md). A weakness found there is written
down here once it is fixed, with the change that fixed it.

## Reporting a vulnerability

Please report it privately through GitHub's private vulnerability reporting, which is on for this repository:
[Report a vulnerability](https://github.com/saphan-ai/proven/security/advisories/new), or the same button on the
Security tab. Not in a public issue. Say what you found, how to reproduce it, and which commit you used. Only the
maintainers see the report until a fix is published.
