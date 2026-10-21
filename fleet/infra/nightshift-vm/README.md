# Nightshift in a macOS VM

Meterless Nightshift is the primary AI provider of every hackathon entry (ADR-0019). Its macOS installers are not signed
by a developer identity and not notarized, so macOS refuses them. proven runs Nightshift inside a macOS virtual machine
on macm5, with nothing of the host shared: no folder, no clipboard. Its network is Tart's default NAT. Decision record: [`fleet/orders/TRAIN_ORDER_proven-nightshift-vm.md`](../../orders/TRAIN_ORDER_proven-nightshift-vm.md).

The tool is [`nightshift_vm.sh`](nightshift_vm.sh). Environment, not product code.

## Runbook

1. **Check the host.** `./nightshift_vm.sh check` prints the memory and disk arithmetic and refuses when
   the VM does not fit.
2. **Install Tart.** `./nightshift_vm.sh install-tools` downloads Tart 2.40.1 from its GitHub release, checks the
   archive against the SHA-256 pinned in the script, checks that Tart is signed by Cirrus Labs (Team 9M2P8L4D89) and
   notarized, and asks for your password once to put a `tart` launcher in `/usr/local/bin`.
   - Why not Homebrew: the `cirruslabs/cli` formula fails on Homebrew 7 (*"Calling `depends_on :macos` with
     `depends_on macos:` is disabled!"*).
   - Why not Tart's Softnet network helper, which would keep the VM off the host's private network: it has **no code
     signature** and runs **setuid root** on the host. Containing one unsigned program by installing another, with more
     privilege and outside the VM, is a worse trade. The VM uses NAT, and **with NAT it can reach machines on the host's
     network**; anything there must keep its own authentication. Blocking that range with macOS's own `pf` firewall is
     the hardening step, measured before it is applied.
3. **Create the VM.** `./nightshift_vm.sh create` clones `ghcr.io/cirruslabs/macos-tahoe-base:latest`
   (about 25 GB) and sizes it 4 vCPU, 8 GiB, 80 GB. The image's user is `admin`, password `admin`.
4. **Start it.** `./nightshift_vm.sh start` runs it with `--no-clipboard` (Tart shares the clipboard by
   default) and no `--dir`, waits for its address, and prints the configuration. `status` shows the same; `stop` stops it.
5. **Bring the installer in.** `./nightshift_vm.sh copy-installer ~/Downloads/Nightshift-Setup-mac-m5.dmg` checks the
   file against Meterless's published `SHA256SUMS`, copies it into the VM over ssh (it asks for the `admin` password),
   and checks the SHA-256 again inside the VM. Nothing is shared to do it.
6. **Install, inside the VM's window.** Open the image, drag Nightshift to Applications, start it.
   - **When macOS refuses it** (*"Apple could not verify 'Nightshift Setup' is free of malware…"*): click **Done**, open
     **System Settings → Privacy & Security**, scroll to the line about Nightshift, click **Open Anyway**, confirm with
     the VM's password, and start it again. Do this only inside the VM, never on the host. The Tahoe base image ships with Gatekeeper off (`spctl` answers `override=security disabled`), so the refusal may not appear in the VM at all.
   - Pass Nightshift's email access gate in the VM.
7. **Keep the VM off the private networks.** Measured on 2026-10-21 with NAT: the guest reached TCP 22 on machines of
   the host's private network, on the host itself, and on the tailnet through the host's Tailscale.
   `./nightshift_vm.sh isolate` (password once) installs a `pf` anchor `com.apple/proven-nightshift` that drops
   everything from the VM's subnet to 10/8, 172.16/12, 192.168/16, 100.64/10, 169.254/16 and IPv6 ULA, except DNS to the
   host's VM gateway, and lets connections the host opens to the VM come back. A LaunchDaemon
   (`/Library/LaunchDaemons/com.proven.nightshift-pf.plist`) loads it at boot and every 10 minutes. `pf` is part of
   macOS; nothing new runs as root. `DRY=1 ./nightshift_vm.sh isolate` shows the rules and installs nothing.
   `./nightshift_vm.sh reach <host> …` measures from inside the guest; `unisolate` removes it all.
8. **NVIDIA, inside the VM only.** Put the NVIDIA API key into Nightshift's Providers tab. Never type it on the host,
   never into a file, a repository, a prompt or a report. Pin two lanes to Nemotron models by name.

The network steps (the runtime reachable from proven's machines only) follow in this file when they are measured.

## Settings

Environment variables, or `~/.config/proven/nightshift-vm.env`: `VM` (default
`nightshift`), `CPU`, `MEM_MB`, `DISK_GB` (default 4, 8192, 80).
