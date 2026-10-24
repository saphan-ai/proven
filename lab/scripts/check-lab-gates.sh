#!/bin/sh
# Offline check of the lab's two gates, with a stub docker on PATH. No Docker,
# no network, nothing is published. Run: make check-lab-gates
#   1. make up refuses a LAB_BIND that is not a loopback address, before any
#      container starts.
#   2. make scan and make push refuse when Docker Scout is missing, fails or
#      reports a finding, and make push never reaches `buildx build --push`.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
# make exports ROOT. Set ROOT to another checkout to run the same check on it.
ROOT=${ROOT:-$(CDPATH= cd -- "$here/../.." && pwd)}
export ROOT
. "$ROOT/lab/scripts/lib.sh"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir "$work/bin"
LOG=$work/docker.log
export LOG

# STUB_SCOUT: absent (no plugin), ok (scan passes), fail (scan reports a finding).
cat >"$work/bin/docker" <<'STUB'
#!/bin/sh
echo "docker $*" >>"$LOG"
case $1 in
  scout)
    [ "${STUB_SCOUT:-absent}" = absent ] && exit 1
    case $2 in
      version) exit 0 ;;
      cves) [ "$STUB_SCOUT" = fail ] && exit 2; exit 0 ;;
    esac
    ;;
  version) echo linux/amd64 ;;
esac
exit 0
STUB
chmod +x "$work/bin/docker"
PATH=$work/bin:$PATH
export PATH

fail=0
expect() {
  if [ "$2" != "$3" ]; then
    echo "check-lab-gates: FAIL: $1: got '$2', want '$3'" >&2
    fail=1
  fi
}

# run NAME CMD...  — sets rc and err (stderr) for the command, with a fresh log.
run() {
  : >"$LOG"
  rc=0
  "$@" >"$work/out" 2>"$work/err" </dev/null || rc=$?
  err=$(cat "$work/err")
}

has_err() {
  case $err in
    *"$2"*) ;;
    *)
      echo "check-lab-gates: FAIL: $1: stderr does not name '$2': $err" >&2
      fail=1
      ;;
  esac
}

logged() {
  grep -q -- "$1" "$LOG"
}

# 1. LAB_BIND. A refused value names LAB_BIND and the value, and no docker call
# is made. An accepted value goes on to `compose up`.
for bad in 0.0.0.0 '' 192.168.1.5 10.0.0.1 localhost 127.1 127.0.0.256 128.0.0.1 '::1' '[::]' '127.0.0.1:80' '127.0.0.1 '; do
  LAB_BIND=$bad
  export LAB_BIND
  run sh "$ROOT/lab/scripts/up.sh"
  expect "LAB_BIND='$bad' exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
  has_err "LAB_BIND='$bad'" "LAB_BIND=$bad "
  expect "LAB_BIND='$bad' touched docker" "$([ -s "$LOG" ] && echo yes || echo no)" no
done
for good in 127.0.0.1 127.0.0.2 127.255.255.254 '[::1]'; do
  LAB_BIND=$good
  export LAB_BIND
  run sh "$ROOT/lab/scripts/up.sh"
  expect "LAB_BIND='$good' reached compose up" "$(logged ' up -d gitea dind' && echo yes || echo no)" yes
done
unset LAB_BIND
run sh "$ROOT/lab/scripts/up.sh"
expect "LAB_BIND unset reached compose up" "$(logged ' up -d gitea dind' && echo yes || echo no)" yes

# 2. Scan before publish.
STUB_SCOUT=absent
export STUB_SCOUT
run sh "$ROOT/lab/scripts/scan.sh"
expect "scan without scout exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
has_err "scan without scout" "docker scout is not installed"
has_err "scan without scout" "$PROVEN_IMAGE"

run sh "$ROOT/lab/scripts/push.sh"
expect "push without scout exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
has_err "push without scout" "docker scout is not installed"
has_err "push without scout" "$SEED_IMAGE"
expect "push without scout reached buildx --push" "$(logged ' --push' && echo yes || echo no)" no
expect "push without scout made a builder" "$(logged 'buildx create' && echo yes || echo no)" no

STUB_SCOUT=fail
run sh "$ROOT/lab/scripts/push.sh"
expect "push with a finding exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
has_err "push with a finding" "refusing $PROVEN_IMAGE:latest"
has_err "push with a finding" "refusing $SEED_IMAGE:$SEED_IMAGE_TAG"
expect "push with a finding reached buildx --push" "$(logged ' --push' && echo yes || echo no)" no

run sh "$ROOT/lab/scripts/scan.sh"
expect "scan with a finding exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused

STUB_SCOUT=ok
run sh "$ROOT/lab/scripts/scout-images.sh"
expect "scout-images with a clean scan exit" "$rc" 0
run sh "$ROOT/lab/scripts/push.sh"
expect "push with a clean scan exit" "$rc" 0
expect "push with a clean scan reached buildx --push" "$(logged ' --push' && echo yes || echo no)" yes
first_scan=$(grep -n 'scout cves' "$LOG" | head -n 1 | cut -d: -f1)
first_push=$(grep -n -e '--push' "$LOG" | head -n 1 | cut -d: -f1)
expect "scan came before the push" "$([ "$first_scan" -lt "$first_push" ] && echo yes || echo no)" yes

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "check-lab-gates: ok"
