#!/bin/bash
# nightshift_vm.sh — run Meterless Nightshift inside a macOS virtual machine on an Apple-silicon Mac, apart from the host.
#
# Why a VM: the Nightshift macOS installers are not signed by a developer identity and not notarized by Apple, so macOS
# refuses them ("Apple could not verify ..."). A matching SHA-256 proves the file is the one Meterless published; it does
# not prove what the file does. The VM keeps whatever it does away from the host: no shared folder, no shared clipboard,
# no host directory mounted. Its network is Tart's default NAT (see README.md for what that still reaches).
#
# Environment, not product code (fleet tooling of proven, 2026-10-21). Decision record: hackathon/fleet/orders/
# TRAIN_ORDER_proven-nightshift-vm.md. Runbook: hackathon/fleet/infra/nightshift-vm/README.md.
#
# Settings (environment variables, or the file $HOME/.config/proven/nightshift-vm.env):
#   VM         the VM's name (default nightshift)
#   CPU MEM_MB DISK_GB   its size (default 4, 8192, 80)
#   CLIPBOARD  1 shares the clipboard with the guest (default 0). While it is shared, the app in the guest can read
#              everything copied on the host and replace it.
#
# Usage: nightshift_vm.sh check | install-tools | create | start | stop | status | copy-installer <dmg>
#                        | isolate | unisolate | reach [host ...]   (pf: keep the VM off the private networks)
# Bash 3.2 (macOS /bin/bash). Every step prints one line: ✔ or ✖.
set -u
CONF="${NIGHTSHIFT_VM_CONF:-$HOME/.config/proven/nightshift-vm.env}"
[ -f "$CONF" ] && . "$CONF"
CLIPBOARD="${CLIPBOARD:-0}"; VM="${VM:-nightshift}"; CPU="${CPU:-4}"; MEM_MB="${MEM_MB:-8192}"; DISK_GB="${DISK_GB:-80}";
IMG="${IMG:-ghcr.io/cirruslabs/macos-tahoe-base:latest}"
SUMS_URL="${SUMS_URL:-https://meterless.ai/Nightshift/M5+/SHA256SUMS}"
BIN=/usr/local/bin
export PATH="$BIN:$PATH"
LOG="${TMPDIR:-/tmp}/nightshift-vm-$VM.log"
KH="$HOME/.config/proven/nightshift-vm.known_hosts"
KEY="${KEY:-$HOME/.config/proven/nightshift-vm_ed25519}"   # a key for this VM only; without it ssh asks for the password
SSHO="-o UserKnownHostsFile=$KH"; [ -f "$KEY" ] && SSHO="$SSHO -o IdentitiesOnly=yes -i $KEY"

# A pinned release, older than two weeks when pinned (2026-10-21). The Homebrew formula in cirruslabs/cli fails on
# Homebrew 7 ("Calling `depends_on :macos` with `depends_on macos:` is disabled!"), so the release archive is used.
# Tart is signed with Developer ID (Team 9M2P8L4D89) and notarized. Tart's network helper Softnet is NOT used: it has no
# code signature and runs setuid root, which would put an unsigned binary on the host to contain another one.
TART_VER=2.40.1; TART_SHA=363e2701154a8155cbc1bb6d845430c9b42697d2a186bc49574471ca2877db46; TART_TEAM=9M2P8L4D89

# ok and no print the one line every step owes; no also stops the script, so a failed step never lets the next one run.
ok() { echo "✔ $*"; }
no() { echo "✖ $*"; exit 1; }
# vm_exists asks Tart's own list; an exact-name match, so a VM called nightshift2 is not taken for nightshift.
vm_exists() { command -v tart >/dev/null && tart list -q 2>/dev/null | grep -qx "$VM"; }
# vm_args returns the running `tart run` command line: its flags are the measurement of what the guest shares.
vm_args() { ps -axww -o args= | grep "tart run $VM" | grep -v grep | head -1; }

# check measures the host before anything is installed: macOS version (the Tahoe guest needs 26), free memory and disk
# against the VM's size plus headroom for the host, and whether the image registry answers.
check() {
  v=$(sw_vers -productVersion)
  case "$v" in 2[6-9].*) ok "$(sysctl -n machdep.cpu.brand_string), macOS $v";; *) no "macOS $v — the Tahoe guest needs macOS 26";; esac
  fr=$(vm_stat | awk '/page size of/{ps=$8} /Pages free/{f=$3} /Pages inactive/{i=$3} /Pages speculative/{s=$3} END{print int((f+i+s)*ps/1073741824)}')
  dk=$(df -g /System/Volumes/Data | awk 'NR==2{print $4}')
  # Headroom: 8 GiB of memory stays with the host; the disk needs the VM, the 25 GB image and 95 GB for the host to keep.
  need_m=$((MEM_MB/1024+8)); need_d=$((DISK_GB+25+95))
  echo "  memory: ${fr} GiB free+inactive - $((MEM_MB/1024)) for the VM = $((fr-MEM_MB/1024)) GiB left · disk: ${dk} GiB free - $((DISK_GB+25)) ($DISK_GB VM + 25 image) = $((dk-DISK_GB-25)) GiB left"
  [ "$fr" -ge "$need_m" ] && [ "$dk" -ge "$need_d" ] && ok "fits" || no "does not fit (needs $need_m GiB memory and $need_d GB disk free)"
  if vm_exists; then echo "  a VM named $VM exists"; else ok "no VM named $VM yet"; fi
  c=$(curl -s -o /dev/null -w '%{http_code}' -m 20 https://ghcr.io/v2/); case "$c" in 200|401) ok "ghcr.io answers ($c)";; *) no "ghcr.io answered $c";; esac
}

fetch() { # fetch <url> <sha256> <dir> — download and verify, or stop
  curl -fsSLo "$3/a.tar.gz" "$1" || no "download failed: $1"
  got=$(shasum -a 256 "$3/a.tar.gz" | awk '{print $1}')
  [ "$got" = "$2" ] && ok "sha256 matches the pin ($(basename "$1") ${2:0:12}…)" || no "sha256 $got is not the pinned $2"
  tar xzf "$3/a.tar.gz" -C "$3" || no "extract failed"
}

# install_tools puts the pinned Tart in /Applications after two proofs: the archive's SHA-256 matches the pin, and the
# app is signed by Cirrus Labs' team and notarized. A one-line wrapper in /usr/local/bin is the only file sudo writes.
install_tools() {
  d=$(mktemp -d)
  if [ -x /Applications/tart.app/Contents/MacOS/tart ] && /Applications/tart.app/Contents/MacOS/tart --version 2>/dev/null | grep -qx "$TART_VER"; then
    ok "Tart $TART_VER already in /Applications"
  else
    fetch "https://github.com/cirruslabs/tart/releases/download/$TART_VER/tart.tar.gz" "$TART_SHA" "$d"
    team=$(codesign -dv "$d/tart.app" 2>&1 | sed -n 's/^TeamIdentifier=//p')
    [ "$team" = "$TART_TEAM" ] && spctl --assess "$d/tart.app" 2>/dev/null && ok "Tart is signed by Team $team and notarized" || no "Tart signature check failed (team '$team')"
    rm -rf /Applications/tart.app && cp -R "$d/tart.app" /Applications/ || no "copy to /Applications failed"
    ok "Tart $TART_VER in /Applications/tart.app"
  fi
  printf '#!/bin/sh\nexec /Applications/tart.app/Contents/MacOS/tart "$@"\n' > "$d/tart"
  sudo mkdir -p "$BIN" && sudo install -o root -g wheel -m 0755 "$d/tart" "$BIN/tart" || no "could not write $BIN/tart (sudo)"
  ok "$BIN/tart → $(tart --version)"
  rm -rf "$d"
}

# create clones the base image and sizes the VM from the settings; it refuses rather than overwrite an existing VM.
create() {
  command -v tart >/dev/null || no "tart is not installed — run: $0 install-tools"
  vm_exists && no "a VM named $VM already exists — see: tart list"
  echo "  cloning $IMG (about 25 GB) …"
  tart clone "$IMG" "$VM" || no "tart clone failed"
  ok "cloned $VM"
  tart set "$VM" --cpu "$CPU" --memory "$MEM_MB" --disk-size "$DISK_GB" || no "tart set failed"
  ok "sized: $CPU vCPU, $MEM_MB MB, $DISK_GB GB"
}

start() {
  vm_exists || no "no VM named $VM — run: $0 create"
  [ -n "$(vm_args)" ] && { ok "already running: $(vm_args)"; return; }
  # No --dir: nothing of the host is shared. --no-clipboard unless CLIPBOARD=1: Tart shares the clipboard by default.
  cf=--no-clipboard; [ "$CLIPBOARD" = 1 ] && cf=
  nohup tart run "$VM" $cf >> "$LOG" 2>&1 &
  # The guest takes a while to boot: wait up to 5 minutes (60 × 5 s) for Tart to report its address.
  ip=""; for i in $(seq 1 60); do ip=$(tart ip "$VM" 2>/dev/null) && [ -n "$ip" ] && break; sleep 5; done
  [ -n "$ip" ] && ok "$VM running at $ip; its window is open on this Mac" || no "no address after 5 min — see $LOG"
  status
}

# status reads the running command line, not a remembered setting: a shared folder is always a failure, a shared
# clipboard is a failure unless CLIPBOARD=1 asked for it.
status() {
  vm_exists || { echo "no VM named $VM"; return 1; }
  tart get "$VM"
  a=$(vm_args)
  if [ -z "$a" ]; then echo "  not running"; return 0; fi
  case "$a" in *--dir*) echo "✖ a host folder is shared: $a"; return 1;; esac
  case "$a" in
    *--no-clipboard*) ok "running without a shared clipboard or folder: $a";;
    *) if [ "$CLIPBOARD" = 1 ]; then ok "running with the clipboard SHARED (CLIPBOARD=1), no shared folder: $a"; else echo "✖ the clipboard is shared: $a"; return 1; fi;;
  esac
}

stop() { tart stop "$VM" && ok "$VM stopped" || no "tart stop failed"; }

# copy_installer checks the installer twice against the SHA-256 Meterless publishes: on the host before it is sent, and
# inside the VM after the copy, so neither the download nor the transfer can change it unnoticed.
copy_installer() {
  f="${1:-}"; [ -f "$f" ] || no "usage: $0 copy-installer <path of the .dmg>"
  n=$(basename "$f")
  want=$(curl -fsS -m 20 "$SUMS_URL" | awk -v n="$n" '$2==n{print $1}')
  [ -n "$want" ] || no "$n is not in $SUMS_URL"
  got=$(shasum -a 256 "$f" | awk '{print $1}')
  [ "$got" = "$want" ] && ok "on the host: $n matches the published sha256 ${want:0:12}…" || no "on the host: $got is not the published $want"
  ip=$(tart ip "$VM" 2>/dev/null) || no "$VM is not running — run: $0 start"
  mkdir -p "$(dirname "$KH")"
  echo "  copying into the VM over ssh (the base image's user is admin; it asks for the password) …"
  scp -q $SSHO -o StrictHostKeyChecking=accept-new "$f" "admin@$ip:Downloads/$n" || no "scp failed"
  in=$(ssh $SSHO "admin@$ip" "shasum -a 256 'Downloads/$n'" | awk '{print $1}')
  [ "$in" = "$want" ] && ok "inside the VM: Downloads/$n matches ${want:0:12}…" || no "inside the VM: $in is not $want"
  echo "  Next, in the VM's window: open Downloads/$n, drag Nightshift to Applications, start it; when macOS refuses it,"
  echo "  System Settings → Privacy & Security → Open Anyway (the refusal and this step are in README.md)."
}

# --- isolate: keep the VM off the host's private networks with macOS's own pf firewall -------------------------------
# Measured 2026-10-21 before it was written: with Tart's NAT the guest reached TCP 22 on machines of the host's private
# network, on the host's VM gateway, and on the tailnet through the host's Tailscale. pf is part of macOS (signed by
# Apple); nothing new runs as root but pfctl. The rules live in an anchor under com.apple/, which /etc/pf.conf already
# evaluates and macOS loads at boot (com.apple.pfctl); a LaunchDaemon loads the anchor at boot and every 10 minutes.
VM_NET="${VM_NET:-192.168.64.0/24}"; VM_GW="${VM_GW:-192.168.64.1}"
ANCHOR=com.apple/proven-nightshift
PF_RULES=/etc/pf.anchors/proven-nightshift
PF_LOADER=/usr/local/libexec/proven-nightshift-pf
PF_PLIST=/Library/LaunchDaemons/com.proven.nightshift-pf.plist
PF_LABEL=com.proven.nightshift-pf

pf_rules() { cat <<EOF
# proven-nightshift: the Nightshift VM ($VM_NET) reaches the internet and the host's DNS, nothing private.
table <proven_private> const { 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 100.64.0.0/10, 169.254.0.0/16 }
pass in quick on ! lo0 proto { udp tcp } from $VM_NET to $VM_GW port 53 keep state
pass in quick on ! lo0 proto udp from any port 68 to any port 67 keep state
pass out quick on ! lo0 from $VM_GW to $VM_NET keep state
block drop in quick on ! lo0 from $VM_NET to <proven_private>
block drop in quick on bridge100 inet6 from any to fc00::/7
EOF
}
pf_loader() { cat <<EOF
#!/bin/sh
# proven-nightshift: load the anchor, enable pf if it is off. Idempotent.
/sbin/pfctl -a $ANCHOR -f $PF_RULES 2>/dev/null || exit 1
/sbin/pfctl -s info 2>/dev/null | grep -q 'Status: Enabled' || /sbin/pfctl -E >/dev/null 2>&1
exit 0
EOF
}
pf_plist() { cat <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>$PF_LABEL</string>
  <key>ProgramArguments</key><array><string>$PF_LOADER</string></array>
  <key>RunAtLoad</key><true/>
  <key>StartInterval</key><integer>600</integer>
</dict></plist>
EOF
}

# isolate writes the rules, the loader and the LaunchDaemon to a temp dir and checks all three (pf parses the rules,
# plutil the plist, /etc/pf.conf evaluates com.apple/*) before sudo installs anything. DRY=1 stops after the checks.
isolate() {
  d=$(mktemp -d); pf_rules > "$d/rules"; pf_loader > "$d/loader"; pf_plist > "$d/plist"
  /sbin/pfctl -n -f "$d/rules" >/dev/null 2>&1 && ok "pf rules parse" || no "pf rules do not parse: /sbin/pfctl -n -f $d/rules"
  plutil -lint -s "$d/plist" && ok "LaunchDaemon plist is valid" || no "plist invalid: $d/plist"
  grep -q 'anchor "com.apple/\*"' /etc/pf.conf && ok "/etc/pf.conf evaluates com.apple/*" || no "/etc/pf.conf has no com.apple/* anchor"
  if [ "${DRY:-0}" = 1 ]; then echo "DRY=1 — would install $PF_RULES, $PF_LOADER, $PF_PLIST and load $ANCHOR:"; cat "$d/rules"; rm -rf "$d"; return 0; fi
  sudo install -d -o root -g wheel -m 0755 /usr/local/libexec \
    && sudo install -o root -g wheel -m 0644 "$d/rules" "$PF_RULES" \
    && sudo install -o root -g wheel -m 0755 "$d/loader" "$PF_LOADER" \
    && sudo install -o root -g wheel -m 0644 "$d/plist" "$PF_PLIST" || no "install with sudo failed"
  ok "installed $PF_RULES, $PF_LOADER, $PF_PLIST"
  sudo launchctl bootout system/$PF_LABEL 2>/dev/null
  sudo launchctl bootstrap system "$PF_PLIST" || no "launchctl bootstrap failed"
  # launchctl returns before the loader has run; give it a moment, then read the anchor from pf itself.
  sleep 2
  n=$(sudo /sbin/pfctl -a $ANCHOR -s rules 2>/dev/null | grep -c .)
  st=$(sudo /sbin/pfctl -s info 2>/dev/null | awk '/^Status:/{print $2}')
  # The anchor holds five rules (pf_rules); fewer means it did not load as written.
  [ "$n" -ge 5 ] && [ "$st" = Enabled ] && ok "anchor $ANCHOR holds $n rules; pf is $st" || no "anchor holds $n rules, pf is '$st'"
  rm -rf "$d"
  echo "  Now measure from the guest: $0 reach"
}

# unisolate removes only what isolate installed; pf itself is left as macOS had it.
unisolate() {
  sudo launchctl bootout system/$PF_LABEL 2>/dev/null
  sudo /sbin/pfctl -a $ANCHOR -F all 2>/dev/null
  sudo rm -f "$PF_PLIST" "$PF_LOADER" "$PF_RULES" && ok "pf anchor flushed, LaunchDaemon and files removed (pf itself stays as macOS left it)"
}

reach() { # from inside the guest: what it can open. Usage: reach [host ...] (default: the host's gateway, the internet)
  ip=$(tart ip "$VM" 2>/dev/null) || no "$VM is not running"
  targets="$VM_GW ${*:-}"
  ssh $SSHO "admin@$ip" "for h in $targets; do nc -z -G 3 \$h 22 >/dev/null 2>&1 && echo \"✖ guest reaches \$h:22\" || echo \"✔ guest does not reach \$h:22\"; done; curl -s -o /dev/null -m 10 -w '  internet: HTTP %{http_code}\n' https://www.apple.com/"
}

case "${1:-}" in
  check) check ;;
  isolate) isolate ;;
  unisolate) unisolate ;;
  reach) shift; reach "$@" ;;
  install-tools) install_tools ;;
  create) create ;;
  start) start ;;
  stop) stop ;;
  status) status ;;
  copy-installer) copy_installer "${2:-}" ;;
  *) sed -n '2,21p' "$0" | sed 's/^# \{0,1\}//'; exit 2 ;;
esac
