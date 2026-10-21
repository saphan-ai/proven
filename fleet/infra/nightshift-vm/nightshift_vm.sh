#!/bin/bash
# nightshift_vm.sh — run Meterless Nightshift inside a macOS virtual machine on an Apple-silicon Mac, apart from the host.
#
# Why a VM: the Nightshift macOS installers are not signed by a developer identity and not notarized by Apple, so macOS
# refuses them ("Apple could not verify ..."). A matching SHA-256 proves the file is the one Meterless published; it does
# not prove what the file does. The VM keeps whatever it does away from the host: no shared folder, no shared clipboard,
# no host directory mounted. Its network is Tart's default NAT (see README.md for what that still reaches).
#
# Environment, not product code (fleet tooling of proven, 2026-10-21). Decision record: fleet/orders/
# TRAIN_ORDER_proven-nightshift-vm.md. Runbook: fleet/infra/nightshift-vm/README.md.
#
# Settings (environment variables, or the file $HOME/.config/proven/nightshift-vm.env):
#   VM         the VM's name (default nightshift)
#   CPU MEM_MB DISK_GB   its size (default 4, 8192, 80)
#
# Usage: nightshift_vm.sh check | install-tools | create | start | stop | status | copy-installer <dmg>
# Bash 3.2 (macOS /bin/bash). Every step prints one line: ✔ or ✖.
set -u
CONF="${NIGHTSHIFT_VM_CONF:-$HOME/.config/proven/nightshift-vm.env}"
[ -f "$CONF" ] && . "$CONF"
VM="${VM:-nightshift}"; CPU="${CPU:-4}"; MEM_MB="${MEM_MB:-8192}"; DISK_GB="${DISK_GB:-80}";
IMG="${IMG:-ghcr.io/cirruslabs/macos-tahoe-base:latest}"
SUMS_URL="${SUMS_URL:-https://meterless.ai/Nightshift/M5+/SHA256SUMS}"
BIN=/usr/local/bin
export PATH="$BIN:$PATH"
LOG="${TMPDIR:-/tmp}/nightshift-vm-$VM.log"
KH="$HOME/.config/proven/nightshift-vm.known_hosts"

# A pinned release, older than two weeks when pinned (2026-10-21). The Homebrew formula in cirruslabs/cli fails on
# Homebrew 7 ("Calling `depends_on :macos` with `depends_on macos:` is disabled!"), so the release archive is used.
# Tart is signed with Developer ID (Team 9M2P8L4D89) and notarized. Tart's network helper Softnet is NOT used: it has no
# code signature and runs setuid root, which would put an unsigned binary on the host to contain another one.
TART_VER=2.40.1; TART_SHA=363e2701154a8155cbc1bb6d845430c9b42697d2a186bc49574471ca2877db46; TART_TEAM=9M2P8L4D89

ok() { echo "✔ $*"; }
no() { echo "✖ $*"; exit 1; }
vm_exists() { command -v tart >/dev/null && tart list -q 2>/dev/null | grep -qx "$VM"; }
vm_args() { ps -axww -o args= | grep "tart run $VM" | grep -v grep | head -1; }

check() {
  v=$(sw_vers -productVersion)
  case "$v" in 2[6-9].*) ok "$(sysctl -n machdep.cpu.brand_string), macOS $v";; *) no "macOS $v — the Tahoe guest needs macOS 26";; esac
  fr=$(vm_stat | awk '/page size of/{ps=$8} /Pages free/{f=$3} /Pages inactive/{i=$3} /Pages speculative/{s=$3} END{print int((f+i+s)*ps/1073741824)}')
  dk=$(df -g /System/Volumes/Data | awk 'NR==2{print $4}')
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
  # No --dir: nothing of the host is shared. --no-clipboard: Tart shares the clipboard by default.
  nohup tart run "$VM" --no-clipboard >> "$LOG" 2>&1 &
  ip=""; for i in $(seq 1 60); do ip=$(tart ip "$VM" 2>/dev/null) && [ -n "$ip" ] && break; sleep 5; done
  [ -n "$ip" ] && ok "$VM running at $ip; its window is open on this Mac" || no "no address after 5 min — see $LOG"
  status
}

status() {
  vm_exists || { echo "no VM named $VM"; return 1; }
  tart get "$VM"
  a=$(vm_args)
  if [ -z "$a" ]; then echo "  not running"; return 0; fi
  case "$a" in *--dir*) echo "✖ a host folder is shared: $a"; return 1;; esac
  case "$a" in *--no-clipboard*) ok "running without a shared clipboard or folder: $a";; *) echo "✖ the clipboard is shared: $a"; return 1;; esac
}

stop() { tart stop "$VM" && ok "$VM stopped" || no "tart stop failed"; }

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
  scp -q -o UserKnownHostsFile="$KH" -o StrictHostKeyChecking=accept-new "$f" "admin@$ip:Downloads/$n" || no "scp failed"
  in=$(ssh -o UserKnownHostsFile="$KH" "admin@$ip" "shasum -a 256 'Downloads/$n'" | awk '{print $1}')
  [ "$in" = "$want" ] && ok "inside the VM: Downloads/$n matches ${want:0:12}…" || no "inside the VM: $in is not $want"
  echo "  Next, in the VM's window: open Downloads/$n, drag Nightshift to Applications, start it; when macOS refuses it,"
  echo "  System Settings → Privacy & Security → Open Anyway (the refusal and this step are in README.md)."
}

case "${1:-}" in
  check) check ;;
  install-tools) install_tools ;;
  create) create ;;
  start) start ;;
  stop) stop ;;
  status) status ;;
  copy-installer) copy_installer "${2:-}" ;;
  *) sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'; exit 2 ;;
esac
