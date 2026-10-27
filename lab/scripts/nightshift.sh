#!/bin/sh
# make nightshift: start Meterless Nightshift in a Tart VM on this Mac (path C), wait for it, print where it is.
# macOS on Apple silicon only: Tart does not run anywhere else. On any other host this prints that and exits 0,
# because it is a fact about the host and not a failure of the lab. The VM itself is made and run by
# hackathon/fleet/infra/nightshift-vm/nightshift_vm.sh (measured steps); this script only orchestrates it.
# Every name, port and path is a setting in lab/versions.env (NIGHTSHIFT_*). The runtime token is never read or
# printed here: Nightshift shows it to you in the VM's own window.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

os=$(uname -s)
arch=$(uname -m)
if [ "$os" != "Darwin" ] || { [ "$arch" != "arm64" ] && [ "$arch" != "aarch64" ]; }; then
  echo "nightshift: Tart runs Nightshift's macOS VM and exists for macOS on Apple silicon only; this host is $os/$arch."
  echo "nightshift: NEXT — go live another way, both work here: path A (one free NVIDIA key) or path B (our services through the judges' gateway). docs/LAB.md, \"Go live\"."
  exit 0
fi

vmscript=$ROOT/$NIGHTSHIFT_VM_SCRIPT
if [ ! -f "$vmscript" ]; then
  die "nightshift: $vmscript is not a file (NIGHTSHIFT_VM_SCRIPT in lab/versions.env)"
fi
vm() {
  VM=$NIGHTSHIFT_VM_NAME /bin/bash "$vmscript" "$@"
}

if ! command -v tart >/dev/null 2>&1 && [ ! -x /usr/local/bin/tart ]; then
  die "nightshift: Tart is not installed. NEXT — run: /bin/bash $vmscript install-tools (it downloads a pinned release, checks its checksum and signature, and asks for your password once). Then make nightshift again."
fi
PATH=/usr/local/bin:$PATH
export PATH

vm check

if ! tart list -q 2>/dev/null | grep -qx "$NIGHTSHIFT_VM_NAME"; then
  echo "nightshift: no VM named $NIGHTSHIFT_VM_NAME yet; cloning the base image (about 25 GB, minutes) and sizing it"
  vm create
fi
vm start

ip=$(tart ip "$NIGHTSHIFT_VM_NAME" 2>/dev/null || true)
if [ -z "$ip" ]; then
  die "nightshift: the VM has no address yet. NEXT — wait a minute and run make nightshift again."
fi
base="http://$ip:$NIGHTSHIFT_PORT"

# Any HTTP status from the health path proves the app listens; a key is not needed for that.
waited=0
healthy=0
while [ "$waited" -lt "$NIGHTSHIFT_WAIT_SECONDS" ]; do
  code=$(curl -sS -o /dev/null --max-time 3 -w '%{http_code}' "$base$NIGHTSHIFT_HEALTH_PATH" 2>/dev/null || true)
  case $code in
    '' | 000) ;;
    *)
      healthy=1
      break
      ;;
  esac
  waited=$((waited + 5))
  sleep 5
done

echo "nightshift: the VM $NIGHTSHIFT_VM_NAME is running at $ip"
if [ "$healthy" -ne 1 ]; then
  echo "nightshift: Nightshift is not answering at $base within ${NIGHTSHIFT_WAIT_SECONDS}s."
  echo "nightshift: NEXT — the app is installed and started inside the VM by hand, once: docs/LAB.md, path C, steps 11 to 14. Then run make nightshift again."
  exit 0
fi
echo "nightshift: Nightshift answers at $base"
echo "nightshift: put this in lab/.env:  PROVEN_GENERATOR_BASE_URL=$base"
echo "nightshift: and its runtime token as METERLESS_RUNTIME_TOKEN. Nightshift shows the token in its own window inside the VM; this script never reads or prints it (docs/LAB.md, path C, step 15)."
echo "nightshift: then make up again, and open a new pull request."
