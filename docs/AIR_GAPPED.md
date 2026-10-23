# Running Proven air-gapped

With `PROVEN_NETWORK=local`, no code, prompt or evidence leaves your network: Proven refuses, when it loads its
settings, any address that is not on your own network, and refuses again, when it dials, any connection to one
(ADR-0034). This page says what to run on your side, which settings to set, and how to check that the mode holds.

## What to run locally

Proven calls four services. Each has a local form; the mode needs all four to be local.

| Service | What runs on your network | Setting that points at it |
|---|---|---|
| The models | Meterless Nightshift, on a machine of yours | `PROVEN_GENERATOR_BASE_URL`, `METERLESS_RUNTIME_TOKEN` |
| The safety model | Your own model server serving the content-safety model (guardrail mode `server`) | `PROVEN_GUARDRAIL_BASE_URL`, `PROVEN_GUARDRAIL_KEY` |
| The evidence | Zetaris, on a machine of yours | `ZETARIS_REST_URL`, `ZETARIS_API_KEY`, `ZETARIS_ORG_ID` |
| The Prover | A container runtime on the machine that runs Proven, with the test image already on it | `PROVEN_TEST_RUNTIME`, `PROVEN_TEST_IMAGE` |

The forge is yours too: Gitea on premises is an equal scenario (ADR-0022), and the lab of ADR-0017 is the reference
set-up.

The hosted NVIDIA API is not one of the four. The guardrail's `nvidia` mode is refused in `local` mode, by name.

## The settings

| Setting | Value in the air-gapped mode |
|---|---|
| `PROVEN_NETWORK` | `local`. The default is `public`, which changes nothing. Any other value is refused. |
| `PROVEN_NETWORK_ALLOW` | Optional. Host names that count as inside your network, comma-separated, for example `nightshift.corp.example,zetaris.corp.example`. |
| `PROVEN_GUARDRAIL_MODE` | `server` (the default) or `off`. `nvidia` is refused. |

An address is on your network when its host is one of these:

- a loopback address (`127.0.0.0/8`, `::1`) or the name `localhost`;
- a private address: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` (RFC 1918), `fc00::/7` (RFC 4193), or a
  link-local address (`169.254.0.0/16`, `fe80::/10`);
- a host name listed in `PROVEN_NETWORK_ALLOW`.

A host name that is not on the list is refused even if it would resolve to a private address. Proven does not look a
name up to judge it: a lookup is itself a connection, and an answer can change. Put the name on the list, or use the
address.

Every other setting is on [CONFIGURATION.md](CONFIGURATION.md).

### What is refused, and when

1. **At load.** Each address setting is checked when it loads, and a refusal names the setting and the host:

   ```text
   configuration refused: PROVEN_GENERATOR_BASE_URL="https://nightshift.example.com/v1": host nightshift.example.com is outside the organisation's network (PROVEN_NETWORK=local); allowed: a loopback or private address, localhost, or a host listed in PROVEN_NETWORK_ALLOW
   ```

   The settings checked are `PROVEN_GENERATOR_BASE_URL`, `PROVEN_GUARDRAIL_BASE_URL` and `ZETARIS_REST_URL`. A
   guardrail address is checked even when the guardrail is `off`, so a hosted address cannot wait there to be switched on.
2. **At dial.** Each live client opens its connections through a dialer that refuses an address off your network.
   It holds when a setting was built some other way, and a redirect is never followed. A refused dial is a failed
   service in the review, named like any other.
3. **In the container.** The Prover's image is run with `--pull never`. If the image is not on the machine, the Prover
   is down, and the review says which image to load.

## How to check the mode holds

1. **Run `livecheck` in the mode.** Set the settings above, then:

   ```bash
   PROVEN_NETWORK=local go run ./cmd/livecheck
   ```

   A setting that names an address off your network is refused, and the refusal names it. A service that is on your
   network and answering is `ready`. `livecheck` never prints a secret.
2. **Run the test that proves it.** It reviews a change in the air-gapped mode with every address answered from
   memory by a dialer that records each connection and refuses any host that is not allowed:

   ```bash
   go test -count=1 -run 'TestAirGapped|TestLocal|TestDial' ./internal/review ./internal/network ./internal/testrun
   ```

   `TestAirGappedReviewCompletesAndNoConnectionLeavesTheAllowedHosts` passes when the review completes, every
   recorded dial went to an allowed host, and a public address asked for by hand never reached the dialer.
   `TestLocalRefusesAPublicAddressInEachAddressSetting` passes when each of the three address settings refuses a public
   address, naming the setting and the host.
3. **Watch the wire, if you want a second witness.** Run a review with a firewall rule that drops everything but your
   own ranges. A review that completes, with no service named failed for a refused dial, is the same proof from outside.

## What this does not do

- It does not make a service you run local: if `PROVEN_NETWORK_ALLOW` names a host that forwards to the internet,
  Proven allows the name, and the forward is yours to close.
- A proxy set in the environment is not used by Proven's live clients in the air-gapped mode.
- The Prover's container runs with no network at all (ADR-0025), in either mode.
