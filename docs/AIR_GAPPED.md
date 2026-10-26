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
| The evidence | Zetaris, or CouchDB (`PROVEN_EVIDENCE_STORE=couchdb`), on a machine of yours | `ZETARIS_REST_URL`, `ZETARIS_API_KEY`, `ZETARIS_ORG_ID`; or `PROVEN_COUCHDB_URL`, `PROVEN_COUCHDB_USER`, `PROVEN_COUCHDB_PASSWORD` |
| The telemetry (optional) | An OTLP/HTTP collector on your network. Telemetry is off unless one of these is set | `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`, `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` |
| The Prover | A container runtime on the machine that runs Proven, with the test image already on it | `PROVEN_TEST_RUNTIME`, `PROVEN_TEST_IMAGE` |

The forge is yours too: Gitea on premises is an equal scenario (ADR-0022), and the lab of ADR-0017 is the reference
set-up.

The hosted NVIDIA API is not one of the four. The guardrail's `nvidia` mode is refused in `local` mode, by name.
The generator's `nvidia` mode is refused in `local` mode, by name.

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

A name on the list is not trusted for its answer. Every address the lookup returns is judged by the same rule as an
address you write down: a loopback, private or link-local address, or one you listed. If any answer is outside, the
whole lookup is refused, naming the host and the address, and no connection is made. The dial judges the address again
before it opens the socket.

Proven looks a name up with its own resolver, not the C library's, and asks for the absolute name, so the host's
search domains add no other name. The resolver reads its name servers from the host (`/etc/resolv.conf`; on a platform
with no such file, Go's default of the loopback addresses). Each connection to a name server goes through the same
address check: a server outside the allowed set is refused by name and address before one byte is sent. If every
server in `/etc/resolv.conf` is outside, the lookup fails and the review says which server was refused. Name servers
on loopback (including `systemd-resolved`'s `127.0.0.53`), a private range or your allow list are used.

Every other setting is on [CONFIGURATION.md](CONFIGURATION.md).

### What is refused, and when

1. **At load.** Each address setting is checked when it loads, and a refusal names the setting and the host:

   ```text
   configuration refused: PROVEN_GENERATOR_BASE_URL="https://nightshift.example.com/v1": host nightshift.example.com is outside the organisation's network (PROVEN_NETWORK=local); allowed: a loopback or private address, localhost, or a host listed in PROVEN_NETWORK_ALLOW
   ```

   The settings checked are `PROVEN_GENERATOR_BASE_URL`, `PROVEN_GUARDRAIL_BASE_URL`, `ZETARIS_REST_URL`,
   `PROVEN_COUCHDB_URL` (for the evidence store and, when `PROVEN_COUCHDB_RECORDS` is set, for the evidence-record
   writer), the three OTLP endpoint variables (`OTEL_EXPORTER_OTLP_ENDPOINT`,
   `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`, `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`) and, when a
   second reviewer is set (ADR-0037), `PROVEN_SECOND_REVIEWER_BASE_URL`, with or without its key. The OTLP endpoints
   are checked in sample mode too, because sample mode reads them. A
   guardrail address is checked even when the guardrail is `off`, so a hosted address cannot wait there to be switched on.
2. **At dial.** Each live client, the CouchDB store's, the evidence-record writer's and both OTLP exporters' included, opens its connections
   through a dialer that refuses an address off your network.
   It holds when a setting was built some other way, and a redirect is never followed. A refused dial is a failed
   service in the review, named like any other.
3. **In the container.** The Prover's image is run with `--pull never` and `--network none`. If the image is not on
   the machine, the Prover is down, and the review says which image to load. The address of the container daemon is
   checked at load (see [The paths beyond the HTTP clients](#the-paths-beyond-the-http-clients)).
4. **In git.** Git is run so that it cannot fetch, in a sample review as well: sample mode reads `PROVEN_NETWORK`, a value it cannot read is refused by name, and a sample review with a baseline commit runs git with the same switches. A review of a partial clone that needs an object the clone does not
   hold fails by name, and git is not asked to fetch it.
5. **In name resolution.** A name is looked up only when it is on `PROVEN_NETWORK_ALLOW`. Any other name is refused
   before a DNS query is sent.

## The paths beyond the HTTP clients

The dial check sees the connections Proven's own HTTP clients make. Three other paths reach the network without them.
Each is named here with what Proven enforces on it and what you do.

| Path | What Proven enforces in `local` mode | Your act |
|---|---|---|
| **The container daemon.** The runtime (`docker`, `podman`) talks to a daemon, and where that daemon is comes from `DOCKER_HOST`, `CONTAINER_HOST`, `DOCKER_CONTEXT`, the context named in `~/.docker/config.json`, or, for podman, the service destinations of its `containers.conf` files and its `podman-connections.json` files. | Only `docker` and `podman` are run in `local` mode: any other `PROVEN_TEST_RUNTIME` is refused by name, because Proven cannot tell where its daemon is. For podman every configured destination is checked, not only the active one. Proven reads every place podman takes a configuration from, in three directories: `/usr/share/containers`, `/etc/containers` and `$HOME/.config/containers`. In each it reads `containers.conf`, `containers.rootless.conf` and `containers.rootful.conf`; the `*.conf` files of `containers.conf.d`, `containers.rootless.conf.d` and `containers.rootful.conf.d`, and of each directory inside those (podman keeps one per user id, such as `containers.rootless.conf.d/1000`); and `podman-connections.json` (a rootful podman keeps its connections in `/etc/containers`). It judges them whatever user runs it, because it cannot tell which mode the runtime will use. The variables that move podman's configuration (`CONTAINERS_CONF`, `CONTAINERS_CONF_OVERRIDE`, `XDG_CONFIG_HOME`, `PODMAN_CONNECTIONS_CONF`) are never passed to the runtime, so these are the files it reads. **A directory Proven finds and cannot list is refused by name, not read as empty**, and so is a file it cannot read. With `HOME` unset, local mode refuses: the runtime would read the home of the passwd entry, which Proven does not judge. The names of the rootless and rootful variants are taken from the containers.conf(5) manual and were not checked against a podman on the test seat. Podman's `containers.conf` files and their drop-ins are read as TOML, with the parser podman uses, so a destination is found however the file spells it (a dotted key, a quoted key or table name, an inline table, a literal string, an escape, a key in other case). A file Proven cannot parse, or one whose destinations are not the types podman reads, is refused by name: a file Proven cannot read with certainty is not judged to be safe. A command named `docker` may be podman (the podman-docker package installs one), and Proven cannot tell which, so a `docker` runtime is judged by the podman sources as well as the docker ones: a podman destination outside your network refuses it even when the command is a real docker. The address is checked when the settings load. A local socket (`unix://`, `npipe://`, `fd://`) passes. A `tcp://` address passes only when its host is an IP address on your network (loopback, private or link-local). An **`ssh://`** address is refused whatever its host is, for Docker (`DOCKER_HOST`, `CONTAINER_HOST`, a docker context) and for every podman destination: ssh picks its route from the ssh configuration under `HOME` (`ProxyJump`, `ProxyCommand`, `Match`), so a private IP in the address does not say where it connects, and Proven cannot check every hop. A host **name** is refused, listed or not: the runtime, or ssh under it, would resolve the name with its own resolver, through the name servers of the host's configuration and to an answer Proven never judged, and Proven hands the runtime no name. The same holds for every podman destination. Any other address is refused by name, and no container starts. A docker context that cannot be read is refused by name as well, and so is a docker or podman configuration file larger than 1 MiB: the runtime reads all of it, so Proven never judges only a prefix. | Use a local daemon socket; an ssh tunnel you open yourself to a local socket also works. A remote daemon on your own network is allowed when it is given as its private IP address (`tcp://10.0.0.7:2376`); a name, even one on `PROVEN_NETWORK_ALLOW`, is refused. The daemon's own registry access is yours to close: Proven runs the image with `--pull never`, and the image must already be on the daemon's host. |
| **Git.** In a partial clone, git fetches a missing object from the remote on demand when a command needs it. | Every git command Proven runs has `GIT_NO_LAZY_FETCH=1`, `-c protocol.allow=never` and one `-c protocol.<name>.allow=never` for `file`, `git`, `http`, `https`, `ssh` and `ext`, so a per-protocol policy in `.git/config` or `~/.gitconfig` cannot reopen a transport, and no terminal prompt. Both switches stop the fetch (measured on git 2.43). A failed read of a partial clone (a repository with a promisor remote) is named, whatever words git uses: the review fails with a message naming `PROVEN_NETWORK=local` and the partial clone. Proven never runs `git fetch` or `git clone`. **No git command Proven starts runs a filter driver or a signature program.** A driver is a program that `.gitattributes` in the reviewed repository and the git configuration of the user or the system choose (`filter.<name>.clean`, `smudge`, `process`; git-lfs is one, and its smudge connects over HTTPS to its own endpoint, which the switches above do not govern). `git archive` runs the smudge driver, so in local mode the tree of a commit is read with `git ls-tree` and `git cat-file`, which convert nothing: a file is the blob git stores, and a repository that keeps files in git-lfs gives its pointer files. The diff of a working tree and the file lists of one run the clean driver and ask the file system monitor (`core.fsmonitor` names a program, and it ran in the test), so they are read with the attributes of an empty tree, `core.attributesFile=/dev/null`, `core.fsmonitor=false` and `GIT_ATTR_NOSYSTEM=1` (this needs git 2.40 or later; an older git refuses the switch by name). The reader of one blob (`git show`) and of the log convert nothing. `git log` would check the signature of a signed commit by starting `gpg.program`, `gpg.x509.program` or `gpg.ssh.program` when `log.showSignature` is set, and the commit's own header picks which one; every offline git command carries `-c log.showSignature=false`, so none starts (test `TestLocalMessagesStartNoSignatureProgram`; on a host where git cannot write a temporary file under Proven's minimal environment the behaviour test cannot go red, and the switch is then held by `TestOfflineGitCommandCarriesBothSwitches`). git reads `.git/info/attributes` of the reviewed clone whatever the switches say, and a line in it can name a filter driver, so in local mode a clone whose `.git/info/attributes` sets any attribute is refused by name before the working tree is read; a file that is absent or holds only comments is not refused (test `TestLocalInfoAttributesStartNoFilter`: the driver leaves no marker in local mode, and does online). `core.sshCommand` and the credential helpers belong to the transports, which are closed, and a hook runs only when git writes to the repository, which Proven's commands do not (not measured: no hook-writing command is run). | Review a full clone, or fetch the objects yourself on a connected host before you go off the network. If the review refuses the clone for `.git/info/attributes`, empty or remove that file. |
| **DNS.** A name on `PROVEN_NETWORK_ALLOW` is resolved when Proven dials it. The query is itself a connection, to a name server, and the answer is an address Proven will connect to. | Only a name on the list, or an IP literal, reaches a resolver at all. `localhost` is never looked up. Proven's own resolver connects only to a name server inside the allowed set, and refuses an outside server by name and address. An answer outside the allowed set is refused before any dial. | List at least one name server in `/etc/resolv.conf` that is on loopback, a private range or your allow list, and make it answer the listed names with addresses on your network. A name server that then forwards to the internet is yours to close, as with a forwarding service. |
| **A local resolver daemon.** `nscd` (glibc's name-service cache) and `systemd-resolved` answer on a unix socket (`/var/run/nscd/socket`, `/run/systemd/resolve/io.systemd.Resolve`), and may resolve a host name for a process by asking DNS on its behalf. A process that only looks up a user or a group (git does) also connects to the `nscd` socket. That connect is local IPC, but the daemon behind it is a DNS path. | Proven starts no child that asks a host name off the list. Git runs with `-c protocol.allow=never`, so it refuses a remote URL before it resolves the name. The container client dials only the checked daemon address. Measured on the test seat, git asked about `git://` and `https://` names off the list made no connect at all with the switch on, and made a connect to a DNS server on port 53 with it off. | Run a local resolver that knows only the names on `PROVEN_NETWORK_ALLOW`, or run neither daemon, or stop them on a host that must not resolve. Proven cannot see what a daemon on the host does with a name it was asked. |

## Trust boundary

The programs you install and point Proven to are your trusted base: the container runtime (`docker` or `podman`), `git`
and the Go toolchain. Proven does not judge what they do on their own. "Nothing leaves the network" covers everything
Proven itself runs and reads: its HTTP clients, its DNS lookups, the git commands it starts, and the container
commands it starts. For a program it starts, Proven checks the configuration that could send it off the network (see
the table above). It does not check the program's code.

Two checks follow from this, both in `local` mode and both on the container runtime, because it is the one program
Proven starts at a path you choose (`PROVEN_TEST_RUNTIME`):

- **It prints what runs.** At start Proven prints one line to standard error and writes the same line to the trace:
  `proven: container runtime /usr/bin/docker: Docker version 29.9.1, build abc`. It is the absolute path Proven found
  and the first line of `<runtime> --version`, which is given five seconds. A runtime that does not answer is shown as
  `version unknown`; that is not a refusal.
- **It refuses a runtime anyone could have replaced.** A runtime given as a relative path (`./docker`, `bin/docker`) is
  refused, because it names a different program in each directory. A bare name (`docker`, `podman`) is found through
  `PATH` once. The file found (after links are followed) is refused when it is writable by every user, and so is a
  file in a directory that is writable by every user and has no sticky bit. The refusal names
  `PROVEN_TEST_RUNTIME`, says what it found, and says what to do.

Your act when one of them refuses: set `PROVEN_TEST_RUNTIME` to an absolute path, or leave it unset for `docker`, and
remove the world-write permission (`chmod o-w <file>` or `chmod o-w <directory>`) or install the runtime in a directory
only its owner can change. Public mode checks none of this and prints nothing.

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
   - `TestLocalAllowedNameRejectsOutsideDNSAnswer` passes when a listed name whose DNS answer is a public address is
     refused by `Lookup`, with the same resolver shown to return that address. `TestLocalDialRefusesAnOutsideAnswerByNameAndAddress`
     passes when the dial refuses it, naming the host and the address, with no connection.
   - `TestLocalResolverDialsOnlyAnAllowedNameServer` passes when a public name server is refused by name and address
     and a loopback, private or listed one is dialled.
   - `TestLocalRefusesAPublicCouchDBAddress`, `TestLoadCouchLocalRefusesAnOutsideAddress`,
     `TestLocalRefusesAPublicTelemetryEndpoint` (each of the three variables), `TestLocalRefusesAPublicTelemetryEndpointInSampleMode`
     and `TestSetupLocalRefusesAnOutsideEndpoint` pass when an outside CouchDB or OTLP address is refused at load, by
     name. `TestLoadRecordsLocalRefusesAnOutsideAddress` passes when the evidence-record writer refuses the same address.
     `TestSetupLocalDialsBothExportersThroughThePolicy` and `TestRecorderLocalClientDialsThroughThePolicy` pass when
     both exporters and the record writer dial through the policy.
     The witness above also reviews through CouchDB and an OTLP collector on loopback, and a public address given to
     the store by hand reaches no connect.
   - `TestLocalDaemonNameIsNotResolvedByTheRuntime` passes when a daemon address given as a listed name is refused.
   - `TestAbsoluteNameEndsWithADot` and `TestLocalAsksAListedNameOnceAsItself` pass when a listed name is asked as the
     absolute name, once, with no search suffix.
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
- It does not see where a name server on your network, or a container daemon on your network, sends a query or a pull
  of its own. Proven checks the address it is given, not what the service at that address does next.
- The Prover's container runs with no network at all (ADR-0025), in either mode.
