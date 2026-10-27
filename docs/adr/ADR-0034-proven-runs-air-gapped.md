# ADR-0034: Proven runs air-gapped: no code, prompt or evidence leaves the organisation's network

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) (Nightshift), [ADR-0022](ADR-0022-gitea-on-premises-is-an-equal-scenario.md) (Gitea on premises), [ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) (the Prover's container)
**Trigger:** the owner, 23 October: features closer to enterprise, built on what Proven already has

## Context

Every service Proven calls can already run on the organisation's own machines: the models through Meterless
Nightshift, the safety model on a local model server, Zetaris locally, the forge as Gitea. What is missing is one mode
that keeps every call inside, and a proof that it does.

## Decision

1. **One setting, `PROVEN_NETWORK=local`**, turns on the air-gapped mode. In it Proven refuses at load any configured
   address that is not on the organisation's own network (loopback, private ranges, or hosts listed in a setting), and
   names each one.
2. **Every outbound call in that mode goes only to the configured local services**; the hosted NVIDIA API and any other
   public address are refused, not tried.
3. **The proof is a test**: a review in air-gapped mode with every public address blocked completes, and a probe shows
   no connection left the allowed hosts.
4. **The trust boundary (the owner, 2026-10-26).** The programs the operator installs and points Proven to (the
   container runtime, git, the Go toolchain) are the operator's trusted base. "Nothing leaves the network" covers
   everything Proven itself runs and reads, DNS and subprocesses included; the claim is not narrowed. For the one
   program started at a path the operator chooses, the container runtime, Proven checks two things in `local` mode: at
   start it prints and traces the absolute path and the version of the runtime, so the operator sees what runs; and it
   refuses a relative `PROVEN_TEST_RUNTIME` and a runtime that any user could replace (a world-writable file, or one
   under a world-writable directory without the sticky bit, at any level or along any link), by name. A refusal says what it found and what to do. The
   reviewer's alternative, refusing every runtime Proven cannot verify, is not adopted: a name is not an identity,
   and the operator's own `docker` or `podman` is the setup the docs name.
   The git policy belongs to the review, not to the process: a review fixes it, with its read limits, when it is
   opened, and every git command it starts carries it, whatever another review in the same process does. The forge
   address the action posts its summary comment to (`GITHUB_API_URL`) is judged in `local` mode by the same rule as
   every other address (`proven check-address`, before the scan), and is refused by name when it is outside; the
   comment is then posted with `curl` only, with no configuration file and no proxy, and nothing is installed.
5. **`docs/AIR_GAPPED.md`** tells an operator what to run locally, the settings, and how to check the mode holds.

## Consequences

- An enterprise can say: our code never leaves our network, and show the test.
- The lab of ADR-0017 is the reference set-up for it.
