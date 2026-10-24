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

A name is looked up only when it is on the list. `localhost` is answered with the loopback addresses and an IP
literal answers itself; neither sends a query. Any other name is refused before a DNS query is sent.

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
3. **In the container.** The Prover's image is run with `--pull never` and `--network none`. If the image is not on
   the machine, the Prover is down, and the review says which image to load. The address of the container daemon is
   checked at load (see [The paths beyond the HTTP clients](#the-paths-beyond-the-http-clients)).
4. **In git.** Git is run so that it cannot fetch. A review of a partial clone that needs an object the clone does not
   hold fails by name, and git is not asked to fetch it.
5. **In name resolution.** A name is looked up only when it is on `PROVEN_NETWORK_ALLOW`. Any other name is refused
   before a DNS query is sent.

## The paths beyond the HTTP clients

The dial check sees the connections Proven's own HTTP clients make. Three other paths reach the network without them.
Each is named here with what Proven enforces on it and what you do.

| Path | What Proven enforces in `local` mode | Your act |
|---|---|---|
| **The container daemon.** The runtime (`docker`, `podman`) talks to a daemon, and where that daemon is comes from `DOCKER_HOST`, `CONTAINER_HOST`, `DOCKER_CONTEXT`, or the context named in `~/.docker/config.json`. | The address is checked when the settings load. A local socket (`unix://`, `npipe://`, `fd://`) passes. A `tcp://` or `ssh://` address passes only when its host is on your network, by the rule above. Any other address is refused by name, and no container starts. A docker context that cannot be read is refused by name as well. | Use a local daemon socket. A remote daemon on your own network is allowed when its host is a private address or is on `PROVEN_NETWORK_ALLOW`. The daemon's own registry access is yours to close: Proven runs the image with `--pull never`, and the image must already be on the daemon's host. |
| **Git.** In a partial clone, git fetches a missing object from the remote on demand when a command needs it. | Every git command Proven runs has `GIT_NO_LAZY_FETCH=1` and `-c protocol.allow=never`, and no terminal prompt. Both switches stop the fetch (measured on git 2.43). The review fails with a message naming `PROVEN_NETWORK=local` and the partial clone. Proven never runs `git fetch` or `git clone`. | Review a full clone, or fetch the objects yourself on a connected host before you go off the network. |
| **DNS.** A name on `PROVEN_NETWORK_ALLOW` is resolved by your host's resolver when Proven dials it. That query is itself a connection, to whatever resolver the host is set to use. | Only a name on the list, or an IP literal, reaches a resolver at all. `localhost` is never looked up. | Point the host's resolver at one on your network, or put the name in `/etc/hosts` or use an address. Proven cannot see where the host's resolver sends the query. |
| **A local resolver daemon.** `nscd` (glibc's name-service cache) and `systemd-resolved` answer on a unix socket (`/var/run/nscd/socket`, `/run/systemd/resolve/io.systemd.Resolve`), and may resolve a host name for a process by asking DNS on its behalf. A process that only looks up a user or a group (git does) also connects to the `nscd` socket. That connect is local IPC, but the daemon behind it is a DNS path. | Proven starts no child that asks a host name off the list. Git runs with `-c protocol.allow=never`, so it refuses a remote URL before it resolves the name. The container client dials only the checked daemon address. Measured on the test seat, git asked about `git://` and `https://` names off the list made no connect at all with the switch on, and made a connect to a DNS server on port 53 with it off. | Run a local resolver that knows only the names on `PROVEN_NETWORK_ALLOW`, or run neither daemon, or stop them on a host that must not resolve. Proven cannot see what a daemon on the host does with a name it was asked. |

## How to check the mode holds

1. **Run `livecheck` in the mode.** Set the settings above, then:

   ```bash
   PROVEN_NETWORK=local go run ./cmd/livecheck
   ```

   A setting that names an address off your network is refused, and the refusal names it. A service that is on your
   network and answering is `ready`. `livecheck` never prints a secret.
2. **Run the tests that prove it.**

   ```bash
   go test -count=1 -run 'TestAirGapped|TestLocal|TestDial|TestOffline' ./internal/review ./internal/network ./internal/testrun ./internal/change
   ```

   - `TestAirGappedReviewCompletesAndNoConnectionLeavesTheAllowedHosts` reviews a change with every address answered
     from memory by a dialer that records each connection. It passes when the review completes, the generator, the
     guardrail and Zetaris were each dialled, every dial went to an allowed host, and a public address asked for by
     hand never reached the dialer.
   - `TestAirGappedWitnessSeesEveryConnection` is the process-wide witness. It runs a review against real loopback
     services in a process whose `connect` system calls are all seen and answered by a seccomp filter, and the filter
     is inherited by git and by the container runtime. It passes when every connect went to an allowed address, git and
     a runtime stand-in were seen connecting (so the filter reaches them), a public address and a public name asked for
     by hand made no connect and no DNS query, a partial clone whose origin is public failed by name, and git asked
     about hosts off the list looked none up. A unix socket is listed as local IPC, never as a network address; the
     socket of a name-service daemon is listed apart and refused, so a lookup that would have gone through it shows as
     a DNS connect and fails the test. It runs on linux/amd64 and skips elsewhere, saying why.
   - `TestConnectKindTellsIPCFromTheNetwork` passes when IPC, a resolver socket and a network address are told apart.
   - `TestLocalNeverQueriesDNSForAPublicName` passes when no DNS query is attempted for a name off the list.
   - `TestLocalRefusesAPublicDaemonAddressAtLoad` and `TestLocalRefusesADockerContextWhoseDaemonIsPublic` pass when a
     public container daemon is refused at load.
   - `TestLocalRefusesAPublicAddressInEachAddressSetting` passes when each of the three address settings refuses a
     public address, naming the setting and the host.
3. **Watch the wire, if you want a second witness.** Run a review with a firewall rule that drops everything but your
   own ranges. A review that completes, with no service named failed for a refused dial, is the same proof from outside.

## What this does not do

- It does not make a service you run local: if `PROVEN_NETWORK_ALLOW` names a host that forwards to the internet,
  Proven allows the name, and the forward is yours to close.
- A proxy set in the environment is not used by Proven's live clients in the air-gapped mode.
- It does not see where your host's resolver, or a container daemon on your network, sends a query or a pull of its
  own. Proven checks the address it is given, not what the service at that address does next.
- The Prover's container runs with no network at all (ADR-0025), in either mode.
