#!/bin/sh
# Exercise publication ordering with fake clients; never builds or contacts a registry.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir "$work/bin"
cat > "$work/bin/git" <<'SH'
#!/bin/sh
case "$*" in *rev-parse*) echo 0123456789abcdef ;; esac
SH
cat > "$work/bin/docker" <<'SH'
#!/bin/sh
printf '%s\n' "$*" >> "$FAKE_DOCKER_LOG"
case "$*" in
  run*"for cli in claude copilot"*)
    if [ "${FAKE_VENDOR_CLI:-}" = true ]; then exit 1; fi ;;
  scout*cves*)
    args="$*"
    while [ "$#" -gt 0 ]; do
      if [ "$1" = --output ]; then
        if [ "${HIDDEN_CVE:-}" = true ]; then
          printf '%s' '{"version":"2.1.0","runs":[{"results":[{"ruleId":"CVE-TEST","suppressions":[{"status":"accepted"}]}]}]}' > "$2"
        else
          printf '%s' '{"version":"2.1.0","runs":[{"results":[]}]}' > "$2"
        fi
        break
      fi
      shift
    done
    case "$args" in *"${FAIL_PLATFORM:-not-a-platform}"*) exit 2 ;; esac ;;
esac
SH
chmod +x "$work/bin/git" "$work/bin/docker"
PATH="$work/bin:$PATH"
export PATH
FAKE_DOCKER_LOG="$work/commands"
export FAKE_DOCKER_LOG
for platform in linux/amd64 linux/arm64; do
  : > "$FAKE_DOCKER_LOG"
  FAIL_PLATFORM=$platform; export FAIL_PLATFORM
  if sh "$root/scripts/publish-image.sh" example/proven:test > "$work/out" 2>&1; then
    echo "FAIL: publication accepted a failed $platform scan" >&2; exit 1
  fi
  python3 - "$FAKE_DOCKER_LOG" <<'PY_CHECK'
import sys
assert not any(s.startswith('push ') for s in open(sys.argv[1])), 'pushed before successful scans'
PY_CHECK
done
unset FAIL_PLATFORM
: > "$FAKE_DOCKER_LOG"
HIDDEN_CVE=true; export HIDDEN_CVE
if sh "$root/scripts/publish-image.sh" example/proven:test > "$work/out" 2>&1; then
  echo 'FAIL: publication accepted a suppressed vulnerability' >&2; exit 1
fi
python3 - "$FAKE_DOCKER_LOG" <<'PY_CHECK'
import sys
assert not any(s.startswith('push ') for s in open(sys.argv[1])), 'pushed with a suppressed vulnerability'
PY_CHECK
unset HIDDEN_CVE
FAKE_VENDOR_CLI=true; export FAKE_VENDOR_CLI
: > "$FAKE_DOCKER_LOG"
if sh "$root/scripts/publish-image.sh" example/proven:test > "$work/out" 2>&1; then
  echo 'FAIL: publication accepted a vendor CLI in the product image' >&2; exit 1
fi
python3 - "$FAKE_DOCKER_LOG" <<'PY_CHECK'
import sys
assert not any(s.startswith('push ') for s in open(sys.argv[1])), 'pushed a vendor CLI image'
PY_CHECK
unset FAKE_VENDOR_CLI
: > "$FAKE_DOCKER_LOG"
sh "$root/scripts/publish-image.sh" example/proven:test > "$work/out" 2>&1
python3 - "$FAKE_DOCKER_LOG" <<'PY'
import sys
lines=open(sys.argv[1]).read().splitlines()
scans=[i for i,s in enumerate(lines) if s.startswith('scout cves ')]
pushes=[i for i,s in enumerate(lines) if s.startswith('push ')]
builds=[s for s in lines if s.startswith('buildx build ')]
assert len(builds)==1 and '--load' in builds[0] and '--push' not in builds[0], lines
assert len(scans)==2 and len(pushes)==1 and max(scans)<pushes[0], lines
assert all('--exit-code' in lines[i] and '--ignore-suppressed=false' in lines[i] for i in scans)
assert all('archive://' in lines[i] for i in scans)
exports=[i for i,s in enumerate(lines) if s.startswith('image save --platform ')]
assert len(exports)==2 and exports[0]<scans[0] and exports[1]<scans[1], lines
PY
echo 'publish-image: failures on either architecture block push; both successful scans precede push'
