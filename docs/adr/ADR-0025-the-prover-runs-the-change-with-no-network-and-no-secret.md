# ADR-0025: The Prover runs the change's code in a container with no network and no secret

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (a test is evidence), [ADR-0007](ADR-0007-no-exploit-code.md) (no exploit code), [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) §6 (the runner's protections), [ADR-0013](ADR-0013-no-hardcoded-settings.md)
**Trigger:** decision R2 of the train order [`proven-reviewer-core`](../../hackathon/fleet/orders/TRAIN_ORDER_proven-reviewer-core.md) ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../hackathon/planning/CONVERSATION_2026-10-21_advisor.md))

## Context

The Prover proves a finding with a test that fails on the change and passes on the base, and it runs both
([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)). So Proven runs code from the change under review:
code it has not reviewed yet, written by someone it does not know. No accepted decision said where that code runs.

[ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) §6 already protects the runner Proven runs on: one job
then gone, no secret beyond Proven's own, a narrow network. Inside one review, Proven itself holds its tokens for
Nightshift and Zetaris, and the change's code must not reach them. The advisor recommended a container with no network
and no secret. The owner (translated from Polish): *"R2 — OK."*

## Decision

1. **The change's code, and its tests, run only inside a container that Proven starts for that one run**, never in
   Proven's own process or on the host directly.
2. **The container has no network.** A test that needs one fails, and the review says the test could not run and why.
3. **The container holds no secret.** No environment variable, file or mount of Proven's own carries a token, a key or a
   credential into it. Amended 2026-10-24 (the master, with the container lock-down): the container receives exactly two
   environment values, both fixed constants in Proven's code and never read from Proven's environment: `HOME=/tmp` and
   `GOCACHE=/tmp/go-build`, so a Go test can build on the bounded scratch under a read-only root. Nothing else is passed
   with `-e`, and the container runtime and git run with a minimal environment of their own.
4. **The container sees only the checkout it needs**, and nothing it writes leaves it except the test results Proven
   reads back.
5. **Its image, time limit and resource limits are settings**, with documented defaults, validated on load
   ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). There is no setting that gives it a network or a secret.
6. **Where no container runtime is available, the Prover does not run code.** The finding then has a query as its
   evidence or is reported as unresolved, and the review says that the Prover was off and why. It never falls back to
   running the code outside a container.
7. **Proven builds the command the container runs; a model never does** (the master, 2026-10-23, merged with the wagon
   `the-prover-runs-only-its-test`). The Prover's model names a test by its file and its name and nothing that is
   executed. Proven builds the argument list from a fixed template per language (Go: `go test -run '^<Name>$'` in the
   test's package), with the name and the path checked against strict patterns that keep them inside the checkout, and
   runs it with no shell. An answer that carries a command, or a name or path that fails the checks, is refused with a
   named reason and no container starts. Proof: `TestArgvIsTheTemplateWithNoShell` and the Prover's refusal tests.

## Consequences

- A change cannot reach Proven's tokens, the team's network or the internet through its own tests.
- Tests that need the network cannot be evidence; the review says so instead of passing or failing them.
- A runner without Docker reviews with queries only. GitHub's Linux runners have Docker; the judges' lab has it.
- **Not known yet:** the start-up cost of a container per test run on a large change, and whether a rootless runtime is
  needed on runners where Docker is not allowed.
