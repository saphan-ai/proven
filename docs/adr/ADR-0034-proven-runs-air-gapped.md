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
4. **`docs/AIR_GAPPED.md`** tells an operator what to run locally, the settings, and how to check the mode holds.

## Consequences

- An enterprise can say: our code never leaves our network, and show the test.
- The lab of ADR-0017 is the reference set-up for it.
