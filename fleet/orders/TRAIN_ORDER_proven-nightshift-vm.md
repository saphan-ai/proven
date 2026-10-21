# TRAIN ORDER — proven-nightshift-vm

- master: MASTER, the master of `proven-cursor-infra`, appointed to this train by the owner on 2026-10-21
- status: ACTIVE — written by the advisor on 2026-10-21, amended by the master the same day (§1a, W1, N4); the owner signs it
- track: proven-nightshift-vm
- original: the fleet's record (this copy is the same text)
- repo: proven (public record and the runbook), saphan-infra (the VM's brief and report)
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
shared clipboard, no key of the fleet's inside it. NVIDIA is configured in Nightshift's Providers tab inside the VM. Its
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
  (arithmetic in §1a). Started with `--no-clipboard`, no `--dir`, and the network of N4. **No shared folder, no shared
  clipboard, no host directory mounted.**
- **owner act:** the creation script, rehearsed by the master with `DRY=1` and the output pasted (law #33).
- **dod:** the VM boots; its configuration shows no shared directory; its size matches the arithmetic.
- **master-was-wrong:** —

### W2 · `nightshift-installed-in-the-vm` — the M5+ installer, checked and installed in the VM
- **what:** the M5+ installer copied into the VM, its sha256 checked there against NS-A2, then installed. The email access
  gate and the "Open Anyway" step are the owner's, inside the VM's window.
- **dod:** the sha256 inside the VM pasted; Nightshift starts in the VM; NS-A1 repeated for the installed app and pasted.
- **master-was-wrong:** —

### W3 · `nvidia-in-nightshift` — NVIDIA in Providers, two Nemotron lanes pinned
- **owner acts, none by an agent:** the NVIDIA API key into Nightshift's Providers tab, inside the VM. It is never typed
  on macm5 outside the VM, never in a file in any repository, a prompt, an order or a report.
- **what:** two lanes, each pinned to a Nemotron model by name (the setup guide suggests a fast and a deep one), so that
  `meterless/auto` cannot route to another family ([ADR-0019](../decisions/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md), point 4).
- **dod:** the lanes and their model names, pasted from Nightshift; the key's value nowhere.
- **master-was-wrong:** —

### W4 · `nightshift-reachable-on-the-private-network` — the runtime on the tailnet only
- **what:** the runtime reachable from runner-cursor and macm5 over the private network, and from nothing else. If it
  listens on 127.0.0.1 only (NS-A5), a proxy inside the VM forwards one port to it. The simplest shape to measure first:
  the VM joins the tailnet as its own node and serves that one port to the tailnet only.
- **The runtime token** comes from Nightshift's Console. The owner puts it in the environment of the machines that run
  proven; it is never written into a repository.
- **dod:** from runner-cursor, `GET /v1/models` answers with the token. **Negative controls (law #27):** the same call
  with one character of the token changed is refused; from a machine outside the tailnet, the port does not answer.
- **master-was-wrong:** —

### W5 · `a-nemotron-request-answers` — one Nemotron request answers
- **what:** from runner-cursor, one `POST /v1/chat/completions` to a pinned Nemotron lane, with the headers of NS-A4.
- **dod:** the answer and the model Nightshift reports for it, pasted; it is a Nemotron model.
- **Unplug test (law #41):** stop the VM; the same request fails with a named connection error, not an empty answer.
- **master-was-wrong:** —

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
  `pf` firewall.
- **Tools (measured 2026-10-21):** the `cirruslabs/cli` Homebrew formula fails on Homebrew 7.0.10 (*"Calling
  `depends_on :macos` with `depends_on macos:` is disabled!"*). Tart comes from its GitHub release 2.40.1 instead,
  sha256-pinned, signed by Cirrus Labs (Team 9M2P8L4D89) and notarized, through proven's
  `fleet/infra/nightshift-vm/nightshift_vm.sh`.
- **N3 — the question to Meterless.** Recommended: ask in the Help Meterless channel for a notarized build, and whether
  a VM on our own machine counts for the eligibility rule *"Zetaris and Meterless must be running on your machine"*.

## §4 DEFINITION OF DONE

- Nightshift runs in a VM on macm5 with no shared folder or clipboard; its installer's sha256 was checked inside the VM.
- NVIDIA is configured inside the VM only; two Nemotron lanes are pinned by name.
- The runtime answers from runner-cursor with the token, refuses a wrong token, and does not answer from outside the
  tailnet.
- One Nemotron request answers and names its model; stopping the VM turns it into a named error.
- The runbook is in proven, and the public record is current.

## §5 COST

- The VM and Nightshift: local energy, a named zero.
- Each Nemotron request through NVIDIA's hosted API is priced by NVIDIA's terms for the key, and runs only with the
  owner's consent at the call (law #49). W5 is one request.
