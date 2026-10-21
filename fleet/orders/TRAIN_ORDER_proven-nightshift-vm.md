# TRAIN ORDER — proven-nightshift-vm

- master: MASTER, the master of `proven-cursor-infra`, appointed to this train by the owner on 2026-10-21
- status: ACTIVE — written by the advisor on 2026-10-21, amended by the master the same day (§1a, W1, N4); the owner signs it
- track: proven-nightshift-vm
- original: the fleet's record (this copy is the same text, with the blackouts listed at the end)
- repo: proven (public record and the runbook), ████████████ (the VM's brief and report)
- base: main at the cut
- author: ADVISOR of the hackathon track "The Agent That Can Explain Why"

## §0 THE OWNER, VERBATIM

2026-10-21. The organisers' setup guide makes Meterless Nightshift the primary AI provider of every entry
([ADR-0019](../decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md)). Both macOS installers were refused by
macOS: *"Apple could not verify 'Nightshift Setup' is free of malware that may harm your Mac or compromise your privacy."*
The advisor measured them: unsigned, not notarized; the M5+ download matches Meterless's published SHA-256. The advisor
proposed three options: A, a macOS virtual machine on macm5; B, a separate macOS user; C, install on macm5 directly.

> *"A"*

## §1 WHAT THIS TRAIN IS

Nightshift runs inside a macOS virtual machine on macm5, apart from everything else on that machine: no shared folder, no
key of the fleet's inside it, no reach into the private networks (pf). The clipboard IS shared, by the owner's decision N6. NVIDIA is configured in Nightshift's Providers tab inside the VM. Its
runtime endpoint is reachable from the machines that run proven over the private network, with its own token, and from
nowhere else. One request through a pinned Nemotron lane answers, and a wrong token is refused.

**Why a VM.** macm5 holds the fleet's bridge certificate, the Cursor seats and the Zetaris registry login. The Nightshift
installer is not signed by a developer identity and not notarized by Apple, and it downloads further software. The SHA-256
match proves the file is the one Meterless published; it does not prove what the file does. A VM keeps whatever it does
away from the fleet's secrets.

**What the advisor read on 2026-10-21. CLAIMS FOR YOU TO TEST, not ground (law #44). Re-measure each one before you build
on it. Where you find the advisor wrong, write it in the wagon's `master-was-wrong:` field.**

| # | claim (advisor, 2026-10-21) | source | instrument |
|---|---|---|---|
| NS-A1 | Both macOS installers are unsigned and not notarized; the app inside the M1 image has an ad-hoc signature with no Team ID. | MEASURED by the advisor: `codesign -dv`, `spctl --assess`, `stapler validate` | repeat on the M5+ image |
| NS-A2 | `Nightshift-Setup-mac-m5.dmg` sha256 `823e0ee09e6e36a55bce84a314e75106382fcba019e67e3092e13f63edf2e2a4` matches `https://meterless.ai/Nightshift/M5+/SHA256SUMS`; the same list names `Nightshift-1.2.1-macos-arm64.dmg` sha256 `9fb3023672f379b38da8aa1405b916bfb9c9359fa5a73f6fe0ec5e7cf63ba087`. | MEASURED by the advisor | `shasum -a 256` inside the VM, after the copy |
| NS-A3 | The M5+ edition is for macOS Tahoe on M5 to M6; the M1+ edition for older macOS on M1 to M4. A VM's macOS version decides which edition runs in it. | the organisers' setup guide, p. 17 | the edition that installs and starts in the VM |
| NS-A4 | Nightshift's runtime is OpenAI-compatible, by default at `http://127.0.0.1:8088/v1`; it takes `Authorization: Bearer <token>`, `meterless-app`, `meterless-scope: user`; `GET /v1/models`, `POST /v1/chat/completions`. | the setup guide, pp. 19–20 | W5 |
| NS-A5 | If the runtime listens on 127.0.0.1 only, it is not reachable from outside the VM without a proxy inside it. | inference, not read | `lsof -nP -iTCP -sTCP:LISTEN` inside the VM |
| NS-A6 | Apple Silicon runs macOS guests with Apple's Virtualization framework; Tart and UTM both use it. | general knowledge, not measured here | the VM boots |

## §1a WHAT THE MASTER RE-MEASURED (2026-10-21, the master's readings, dated)

- **NS-A2 confirmed on macm5:** `shasum -a 256 ~/Downloads/Nightshift-Setup-mac-m5.dmg` = `823e0ee0…2e2a4`, the advisor's
  value. remark: it proves the file is the published one; it says nothing about what the file does (that is why the VM).
- **macm5:** Apple M5 Max, 18 cores, 128 GiB; macOS 26.5.2 (Tahoe). Free + inactive memory ~65.5 GiB (`vm_stat`); the
  data volume has 1.3 TiB free (`df -h`). remark: memory is a snapshot taken while other legs ran; a Zetaris VM is not on
  macm5, so nothing else claims these 8 GiB.
- **The arithmetic for W1 (law #46):** 4 of 18 cores; 8 of ~65.5 GiB free (≈57 GiB left); disk ≤ 25 GB image download +
  an 80 GB sparse VM disk ≈ 105 GB of 1.3 TiB. FITS.
- **NS-A3:** the host is an M5 on Tahoe; the guest image is Tahoe (`ghcr.io/cirruslabs/macos-tahoe-base`), so the M5+
  edition is the one to try first. Whether it runs on the VIRTUAL hardware is not measured until W2.
- **Tart on macm5:** not installed (`which tart` prints nothing). Tart's source licence is FSL-1.1-ALv2 (a host tool, not
  code in proven). `tart run` **shares the clipboard by default**: W1 passes `--no-clipboard`; no `--dir` is passed, so no
  folder is shared. The base image's user is `admin` / password `admin`, with SSH on.
- **Network — what the advisor did not cover:** Tart's default network is NAT, and from it the guest reaches macm5's
  private network and the host's own listening ports. **Softnet** (`--net-softnet`) lets the guest reach only globally
  routable addresses and the host's VM gateway — the private network and the fleet's machines are blocked. Softnet needs a
  root (setuid) helper on macm5. Decision **N4** below.

## §2 WAGONS

### W1 · `nightshift-vm-exists` — a macOS VM on macm5, isolated
- **what:** a macOS VM on macm5, with Tart (N1), cloned from `ghcr.io/cirruslabs/macos-tahoe-base:latest`, sized 4 vCPU, 8 GiB, 80 GB
  (arithmetic in §1a). Started with `--no-clipboard`, no `--dir`, and the network of N4. **No shared folder, no host
  directory mounted.** The clipboard: shared since N6.
- **owner act:** the creation script, rehearsed by the master with `DRY=1` and the output pasted (law #33).
- **dod:** the VM boots; its configuration shows no shared directory; its size matches the arithmetic.
- **result (measured by the master 2026-10-21 12:36–12:50Z):** the owner's script finished exit 0. `tart get`: darwin,
  4 CPU, 8192 MB, 80 GB; started as `tart run nightshift --no-clipboard` (no `--dir`). Inside the guest: macOS 26.6.2,
  4 CPU, 8 GiB, `/` 74 GiB, zero virtiofs mounts. **Network, NAT:** from the guest, TCP 22 answers on runner-cursor, on
  the Zetaris host and on macm5's VM gateway; the internet answers. remark: an open port proves reach, not access — each
  of those still asks for its own key.
- **master-was-wrong:** the master first chose Homebrew; the tap's formula is broken on Homebrew 7 (§3 Tools).

### W2 · `nightshift-installed-in-the-vm` — the M5+ installer, checked and installed in the VM
- **what:** the M5+ installer copied into the VM, its sha256 checked there against NS-A2, then installed. The email access
  gate and the "Open Anyway" step are the owner's, inside the VM's window.
- **dod:** the sha256 inside the VM pasted; Nightshift starts in the VM; NS-A1 repeated for the installed app and pasted.
- **measured by the master inside the VM (2026-10-21):** host and guest sha256 of `Nightshift-Setup-mac-m5.dmg` both
  `823e0ee0…2e2a4`, equal to Meterless's published list. NS-A1 holds for the M5+ image: the dmg is not signed and has no
  stapled ticket; `Nightshift Setup.app` (`ai.meterless.installer.mac-m5`) is ad-hoc signed, no Team ID. **New:**
  `spctl` in the guest answers `override=security disabled` — the Tahoe base image ships with Gatekeeper off, so the
  "Open Anyway" step may not appear in the VM.
- **master-was-wrong:** —

### W3 · `nvidia-in-nightshift` — NVIDIA in Providers, two Nemotron lanes pinned
- **owner acts, none by an agent:** the NVIDIA API key into Nightshift's Providers tab, inside the VM. It is never typed
  on macm5 outside the VM, never in a file in any repository, a prompt, an order or a report.
- **what:** two lanes, each pinned to a Nemotron model by name (the setup guide suggests a fast and a deep one), so that
  `meterless/auto` cannot route to another family ([ADR-0019](../decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md), point 4).
- **dod:** the lanes and their model names, pasted from Nightshift; the key's value nowhere.
- **progress 2026-10-21:** the owner put the NVIDIA key into Providers inside the VM, through a short clipboard window
  (VM restarted without `--no-clipboard`, the key pasted, the VM restarted with it; measured afterwards: running with
  `--no-clipboard`). The lanes' model names are owed by the owner. Runtime measured listening on `127.0.0.1:8088`
  (NS-A4, NS-A5 hold); a `meterless` process also listens on `*:8089`, which W4 must account for.
- **NVIDIA provider, as Nightshift shows it (pasted by the owner 2026-10-21):** "verified on the next routed request ·
  12 hosted models registered · trial credits + a per-minute cap on the account · not unlimited · unknown spend once
  credits are gone". remark: so every NVIDIA-routed request may cost once the trial credits end — W5 stays one request,
  run with the owner's consent at the call (law #49).
- **Measured inside the VM (master, 2026-10-21):** `GET 127.0.0.1:8088/v1/models` with no token → 401 `unauthorized`;
  with a wrong token → 401. `127.0.0.1:8089/` → 200 with no token, and that port listens on all of the guest's
  interfaces; what it serves is not yet read (W4 reads it before anything is exposed). Read since: its page is titled "Nightshift · Phone".
- **The two lanes (master, from the owner's `GET /v1/models` paste, 2026-10-21):** the runtime lists NVIDIA's hosted
  models by full name, so no alias is configured — a request names the model: **fast** `nvidia/nemotron-3.5-lightning-30b-a3b`,
  **deep** `nvidia/nemotron-3.5-super-vl-120b-a12b`. Eight `nvidia/` models are listed, beside 18 `meterless/` routing
  modes and the `local-runtime/` catalogue. remark: listing a name proves it is addressable, not that a request to it is
  not re-routed — W5 reads the model the response names. **master-was-wrong (the advisor's word "lanes"):** Nightshift's
  `lanes.toml` holds its LOCAL llama.cpp lanes only (7, declaring ~8 GB active in an 8 GiB VM); NVIDIA is not a lane there.
- **REOPENED by W5, then READ in the code (master, 2026-10-21):** Nightshift's external proxy does not obey a model id
  for identified traffic. `engine/proxy.js` `bindAppMode` (comment NS-APP-STREAM AS-31/34/35): *"the app_policy `mode`
  — decides the routing profile. The client's model id / meterless-mode header / sort knob is RECORDED (receipt, turn
  event) but not obeyed. First contact with no saved bind inherits the cockpit's current mode and persists it"*; the
  request handler passes `forceModel: bind.bound ? null : …`. An app's `lanes` allowlist names LOCAL lanes only
  (`select.js`, `resolveAppPolicy`). ⇒ **A cloud model cannot be pinned from proven.** Under `auto` the race may also
  pick a local lane in the VM (LFM2.5-8B, Llama-3.2-3B). What remains is to READ the answering model from the signed
  receipt (ADR-0019 point 4) and decide what proven does with an answer from the wrong model — decision **N7**.
  **master-was-wrong:** the order (and the advisor) assumed lanes could be pinned by name; they cannot.
- **The runtime token** is the `proxy-token` file in Nightshift's state directory inside the VM (the Nightshift CLI
  reads it from there, `transport.mjs` resolveToken); it is never printed.
- **master-was-wrong:** —

### W4 · `nightshift-reachable-on-the-private-network` — the runtime on the tailnet only
- **what:** the runtime reachable from runner-cursor and macm5 over the private network, and from nothing else. If it
  listens on 127.0.0.1 only (NS-A5), a proxy inside the VM forwards one port to it. The simplest shape to measure first:
  the VM joins the tailnet as its own node and serves that one port to the tailnet only.
- **The runtime token** comes from Nightshift's Console. The owner puts it in the environment of the machines that run
  proven; it is never written into a repository.
- **dod:** from runner-cursor, `GET /v1/models` answers with the token. **Negative controls (law #27):** the same call
  with one character of the token changed is refused; from a machine outside the tailnet, the port does not answer.
- **result (2026-10-21):** the owner saved the fenced tailnet ACL (its `tests` block passed, so the console accepted it);
  fleet reach re-measured from macm5 afterwards: TCP 22 answers on every online node and on the ███████████████
  subnet route, runner-cursor → macm5 answers. The owner's script joined the VM as node `nightshift` (█████████████,
  tag `tag:nightshift`, Tailscale 1.104.1 from Homebrew in the VM) and serves its tcp:8088 to 127.0.0.1:8088.
  **Measured by the master:** runner-cursor → nightshift:8088 open, no token → 401; → 8089 closed; → 22 closed. macm5 →
  8088 open, 8089 closed. █████ (outside the grant) → 8088 closed. From the VM: TCP 22 on runner-cursor, macm5, █████,
  macmini and the Zetaris host over the tailnet: closed; the internet: 200. **Token (the owner's script, 2026-10-21):**
  read from the VM's `proxy-token` (32 characters, never printed), written to `~/.config/proven/nightshift.env` (600)
  on runner-cursor and macm5 with `NIGHTSHIFT_URL`. From runner-cursor: with the token HTTP 200; with one character
  changed HTTP 401. The script's log holds no token (grep: 0). **W4 DONE.**
- **master-was-wrong:** —

### W5 · `a-nemotron-request-answers` — one Nemotron request answers
- **what:** from runner-cursor, one `POST /v1/chat/completions` to a pinned Nemotron lane, with the headers of NS-A4.
- **dod:** the answer and the model Nightshift reports for it, pasted; it is a Nemotron model.
- **Unplug test (law #41):** stop the VM; the same request fails with a named connection error, not an empty answer.
- **result (2026-10-21, evidence in `EVIDENCE_W5_2026-10-21.md`):** three refusals first (`nvidia_401`: Providers held the
  key's UUID, not its `nvapi-` value); then **answered** at 13:35:42Z, Nightshift's signed receipt `rcpt_req_xaswpcmb`:
  provider `nvidia`, cloud, **routed to `nvidia/nemotron-3-ultra-550b-a55b` although `nemotron-3.5-lightning-30b-a3b` was
  named** (app policy `auto`), 39 tokens, 0 USD. Unplug: VM stopped → `curl: (28) Connection timed out after 60002 ms`;
  restarted, the runtime answers again. **W5 DONE** (it is a Nemotron model). **Owed:** the pin — W3 is reopened for it.
- **master-was-wrong:** the master expected naming the model in the request to pin it; Nightshift's `auto` policy for the
  app overrode it (measured).

### W6 · `nightshift-runbook-and-record` — the runbook and the public record
- **what:** the human ceremony of W1–W4 as a numbered runbook in proven, `fleet/infra/nightshift-vm.md`, including what
  to do when macOS refuses the installer (law #39). The public copy of this order and its status stay current, with
  nothing internal in them.
- **dod:** the runbook exists; the deny-list grep over every public file prints nothing.
- **master-was-wrong:** —

## §3 DECISIONS OWED BY THE OWNER

- **N1 — the tool. DECIDED by the owner 2026-10-21: Tart.**
- **N2 — the VM's lifetime. DECIDED by the owner 2026-10-21:** *"Start today, we need to test it. Stopped after 6 Nov."*
  W1–W5 run today, before the build window opens on 2026-10-22 00:00 UTC; the VM is stopped after the results on
  6 November.
- **N4 — the VM's network. DECIDED by the owner 2026-10-21: NAT.** The master had recommended Softnet; the owner
  refused it: Softnet has no code signature and runs setuid root on the host, so it would contain one unsigned program
  by installing another, with more privilege and outside the VM. With NAT the guest can reach machines on macm5's
  network; anything there keeps its own authentication. Hardening, measured first: block that range with macOS's own
  `pf` firewall. **APPLIED 2026-10-21 (the owner: "pf i możesz zrobić skrypt launchpad"):** `nightshift_vm.sh isolate`
  installed anchor `com.apple/proven-nightshift` (6 rules) and LaunchDaemon `com.proven.nightshift-pf` (boot + every
  10 min); pf Enabled. **Measured from the guest, before → after:** TCP 22 on runner-cursor (LAN), the Zetaris host,
  macm5's VM gateway and runner-cursor over the tailnet: OPEN → closed. HTTPS to apple.com, NVIDIA's API and
  meterless.ai: 200 → 200; DNS answers; macm5 → VM ssh still opens. remark: survival of a macm5 reboot is NOT yet
  measured — the LaunchDaemon is the mechanism, the first reboot is its test.
- **Tools (measured 2026-10-21):** the `cirruslabs/cli` Homebrew formula fails on Homebrew 7.0.10 (*"Calling
  `depends_on :macos` with `depends_on macos:` is disabled!"*). Tart comes from its GitHub release 2.40.1 instead,
  sha256-pinned, signed by Cirrus Labs (Team 9M2P8L4D89) and notarized, through proven's
  `fleet/infra/nightshift-vm/nightshift_vm.sh`.
- **N5 — how the runtime is reached. DECIDED by the owner 2026-10-21: the VM joins the tailnet as its own node**
  (the master had recommended an ssh tunnel from macm5). **Condition, stated to the owner:** the node's traffic is
  WireGuard to public addresses, so the host's pf does not filter it — what the VM reaches on the tailnet, and who
  reaches it (including `*:8089`, which answers without a token), is decided by the tailnet ACL alone. ⇒ W4 tags the
  node `tag:nightshift`; the ACL lets nothing out of `tag:nightshift` and lets only runner-cursor and macm5 into
  `tag:nightshift:8088`. The ACL edit is the owner's (admin console); the master drafts it from the current policy.
  **DoD addition (negative controls):** from the VM, TCP 22 on runner-cursor's tailnet address does not answer; from
  runner-cursor, `nightshift:8089` does not answer.
- **N6 — the clipboard. DECIDED by the owner 2026-10-21: shared** (*"zostaw już ten schowek, może się przydać"*), after
  the NVIDIA key would not arrive without it. The trade, stated to the owner when he decided: while it is shared, the
  unsigned app in the guest can read everything copied on macm5 and replace it. `nightshift_vm.sh` keeps the default
  `--no-clipboard`; `CLIPBOARD=1` in `~/.config/proven/nightshift-vm.env` on macm5 records this choice, and `status`
  reds a shared clipboard without it (measured both ways).
- **N7 — what pins proven's reviews to Nemotron, since Nightshift will not. OPEN, the master recommends (a)+(b).**
  (a) proven refuses a review whose receipt's `routed_model` is not `nvidia/*` (product code, the build window; it is
  ADR-0019 point 4 made binding). (b) the Nightshift VM keeps no local chat lanes for `proven`'s modes, so the race has
  only NVIDIA rows to choose from (an owner act in Nightshift's cockpit; measured afterwards by a receipt). (c) set the
  app's mode to `deep` in the Apps panel — narrows the race to stronger rows but still not to one model.
  **(c) APPLIED by the owner 2026-10-21 ("zmieniłem na deep")**; measured in the VM: `state/app-policy.json` →
  `"proven": {"mode": "text.deep"}` (was `auto`). Nightshift counts 3 requests from app `proven`, all naming
  `nvidia/nemotron-3.5-lightning-30b-a3b`. Setup → Providers' *Default model* (the owner's screenshot) offers modes only:
  auto, fast, deep, cheap, local, fusion. **(a) and (b) stay open**; whether `deep` races any local lane is unmeasured
  until the next request's receipt.
  **Control request 1 under deep (13:50:56Z, the owner's run):** connected, then `curl: (28) Operation timed out after
  60002 milliseconds with 0 bytes received` — the master's 60 s limit, sized for one fast call, was too short for deep;
  the limit is now 300 s. **Read in the VM's logs the same minute:** the four LFM2.5 lanes are `gated off:
  LIQUID_FEATURES — left out of every race`; the two Llama-3.2-3B lanes are `parked for memory: need 4029 MB,
  MemAvailable ~3570 MB`; `provider keys: nvidia`. ⇒ today no local chat lane can win a race — N7 (b) holds BY
  ACCIDENT (memory), not by setting. **And Nightshift's receipts do not survive a restart:** `meterless.json`
  `receipts` was 0 after the VM restarted at 13:37Z (it held `rcpt_req_xaswpcmb` before), so evidence must be copied at
  the call — the scripts do.
  **Control request 2 under deep (13:55:43Z):** Nightshift answered `provider_error · ttfb_timeout:local-runtime/
  llama-3.2-3b-instruct` — **deep routed proven's request to the LOCAL 3B lane**, which had loaded (`worker`, 1 slot,
  1–2 tok/s, task 576 from request 1 still decoding past 600 tokens). **master-was-wrong:** "N7 (b) holds by accident" —
  it did not; the parked lane loaded. The owner stopped the runtime afterwards. ⇒ (a) is needed whatever (b) does.
- **N3 — the question to Meterless.** Recommended: ask in the Help Meterless channel for a notarized build, and whether
  a VM on our own machine counts for the eligibility rule *"Zetaris and Meterless must be running on your machine"*.

## §4 DEFINITION OF DONE

- Nightshift runs in a VM on macm5 with no shared folder, off the private networks, with the clipboard shared by N6; its
  installer's sha256 was checked inside the VM.
- NVIDIA is configured inside the VM only; two Nemotron lanes are pinned by name.
- The runtime answers from runner-cursor with the token, refuses a wrong token, and does not answer from outside the
  tailnet.
- One Nemotron request answers and names its model; stopping the VM turns it into a named error.
- The runbook is in proven, and the public record is current.

## §5 COST

- The VM and Nightshift: local energy, a named zero.
- Each Nemotron request through NVIDIA's hosted API is priced by NVIDIA's terms for the key, and runs only with the
  owner's consent at the call (law #49). W5 is one request.

## REDACTED

Parts of this file are blacked out (█). Nothing else was changed, except the `master` line (the seat is left out).

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17.

| what was blacked out | times | what it was |
|---|---|---|
| machine names | 2 | the machine that runs our local models |
| private repository names | 1 | our infrastructure repository |
| network addresses | 2 | tailnet addresses and the private subnet |
