#!/bin/sh
# Offline check of the lab's two gates, with a stub docker on PATH. No Docker,
# no network, nothing is published. Run: make check-lab-gates
#   1. make up refuses a LAB_BIND in the environment, before any container
#      starts. (The ports themselves are loopback in lab/compose.yaml; that is
#      checked by the Go test and by check-review-gates.sh.)
#   2. make scan refuses when Docker Scout is missing, fails or reports a finding.
#   3. make push publishes what Docker Scout scanned, and nothing else:
#        - Scout missing: nothing is built and nothing is published;
#        - a failed scan of any image or any platform: no final tag is made;
#        - on success the final tags are made from the digests that were scanned,
#          every platform of every image was scanned first, and no final tag is
#          ever built;
#        - a final tag that reads back another digest is an error.
#      Each property also has a negative control: the same checks run on a copy
#      of push.sh with exactly the guarding line changed, and must fail.
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

cp "$ROOT/lab/scripts/stub-docker.sh" "$work/bin/docker"
cat >"$work/bin/git" <<'STUB'
#!/bin/sh
echo stub
STUB
chmod +x "$work/bin/docker" "$work/bin/git"
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
  rm -rf "$LOG.state"
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

# 1. LAB_BIND is retired. Any value, an empty one included, is refused by name,
# and no docker call is made. Unset goes on to `compose up`.
for bad in 0.0.0.0 '' 192.168.1.5 10.0.0.1 localhost 127.0.0.1 127.0.0.2 '[::1]' '::1' '[::]' '127.0.0.1:80'; do
  LAB_BIND=$bad
  export LAB_BIND
  run sh "$ROOT/lab/scripts/up.sh"
  expect "LAB_BIND='$bad' exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
  has_err "LAB_BIND='$bad'" "LAB_BIND=$bad "
  expect "LAB_BIND='$bad' touched docker" "$([ -s "$LOG" ] && echo yes || echo no)" no
done
unset LAB_BIND
run sh "$ROOT/lab/scripts/up.sh"
expect "LAB_BIND unset reached compose up" "$(logged ' up -d gitea dind' && echo yes || echo no)" yes

# 2. make scan.
STUB_SCOUT=absent
export STUB_SCOUT
run sh "$ROOT/lab/scripts/scan.sh"
expect "scan without scout exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
has_err "scan without scout" "docker scout is not installed"
has_err "scan without scout" "$PROVEN_IMAGE"

STUB_SCOUT=fail
run sh "$ROOT/lab/scripts/scan.sh"
expect "scan with a finding exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused

STUB_SCOUT=ok
run sh "$ROOT/lab/scripts/scout-images.sh"
expect "scout-images with a clean scan exit" "$rc" 0

# 3. make push.
name=$PROVEN_REGISTRY/$PROVEN_IMAGE
seed_name=$PROVEN_REGISTRY/$SEED_IMAGE
STUB_SCOUT=absent
run sh "$ROOT/lab/scripts/push.sh"
expect "push without scout exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
has_err "push without scout" "docker scout plugin is not installed"
expect "push without scout built" "$(logged 'buildx build' && echo yes || echo no)" no

STUB_SCOUT=fail
run sh "$ROOT/lab/scripts/push.sh"
expect "push with a finding exit" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
has_err "push with a finding" "refusing $name:staging-"
has_err "push with a finding" "refusing $seed_name:staging-"
expect "push with a finding made a final tag" "$(logged 'imagetools create' && echo yes || echo no)" no

# push_ok TREE  — runs push.sh of TREE through five scenarios and reads the
# stub's log. Returns 0 when every property holds; otherwise sets why.
why=
bad() {
  why=$1
  return 1
}

run_push() {
  : >"$LOG"
  rm -rf "$LOG.state"
  rc=0
  ROOT=$1 STUB_SCOUT=$2 STUB_FAIL_PLATFORM=${3:-} STUB_DRIFT=${4:-} \
    sh "$1/lab/scripts/push.sh" >"$work/out" 2>"$work/err" </dev/null || rc=$?
}

# no build ever names a final tag, and none comes after the first scan.
builds_stage_only() {
  grep 'buildx build' "$LOG" | grep -e "--tag $name:latest" -e "--tag $name:stub" -e "--tag $seed_name:$SEED_IMAGE_TAG " >/dev/null && return 1
  first_scan=$(grep -n 'scout cves' "$LOG" | head -n 1 | cut -d: -f1)
  last_build=$(grep -n 'buildx build' "$LOG" | tail -n 1 | cut -d: -f1)
  [ -z "$first_scan" ] || [ "$last_build" -lt "$first_scan" ]
}

# Every create names a scanned index digest, every platform of that index was
# scanned before it, and every published digest is one of those index digests.
published_were_scanned() {
  creates=$(grep -n 'buildx imagetools create' "$LOG" | cut -d: -f1)
  [ -n "$creates" ] || return 1
  served=$LOG.state/served
  for ln in $creates; do
    line=$(sed -n "${ln}p" "$LOG")
    src=${line##* }
    repo=${src%@*}
    dig=${src#*@}
    ref=$(awk -v r="$repo:" -v d="$dig" '$1 == "index" && index($2, r) == 1 && $3 == d { print $2; exit }' "$served")
    [ -n "$ref" ] || return 1
    for p in $(printf '%s' "$PUSH_PLATFORMS" | tr ',' ' '); do
      pd=$(awk -v p="$p" -v r="$ref" '$1 == "platform" && $2 == p && $3 == r { print $4; exit }' "$served")
      [ -n "$pd" ] || return 1
      sl=$(grep -n "scout cves registry://$repo@$pd" "$LOG" | head -n 1 | cut -d: -f1)
      [ -n "$sl" ] && [ "$sl" -lt "$ln" ] || return 1
    done
  done
  [ -s "$LOG.state/published" ] || return 1
  while read -r tag pubd; do
    awk -v d="$pubd" '$1 == "index" && $3 == d { f = 1 } END { exit !f }' "$served" || return 1
  done <"$LOG.state/published"
}

push_ok() {
  t=$1
  run_push "$t" absent
  [ "$rc" -ne 0 ] || bad "scout missing: push exited 0" || return 1
  ! grep -q -e 'buildx build' -e 'buildx create' -e 'imagetools create' "$LOG" || bad "scout missing: it built or published" || return 1

  run_push "$t" fail
  [ "$rc" -ne 0 ] || bad "failing scans: push exited 0" || return 1
  ! grep -q 'imagetools create' "$LOG" || bad "failing scans: a final tag was made" || return 1
  builds_stage_only || bad "failing scans: a build named a final tag" || return 1

  # The second platform's image fails; the first is clean.
  run_push "$t" ok linux/arm64
  [ "$rc" -ne 0 ] || bad "arm64 scan failed: push exited 0" || return 1
  ! grep -q 'imagetools create' "$LOG" || bad "arm64 scan failed: a final tag was made" || return 1
  [ "$(grep -c 'scout cves' "$LOG")" -ge 4 ] || bad "arm64 scan failed: not every platform of every image was scanned" || return 1
  # The first platform's image fails; the second is clean.
  run_push "$t" ok linux/amd64
  [ "$rc" -ne 0 ] || bad "amd64 scan failed: push exited 0" || return 1
  ! grep -q 'imagetools create' "$LOG" || bad "amd64 scan failed: a final tag was made" || return 1

  run_push "$t" ok
  [ "$rc" -eq 0 ] || bad "clean scans: push exited $rc" || return 1
  builds_stage_only || bad "clean scans: a build named a final tag, or came after a scan" || return 1
  published_were_scanned || bad "clean scans: a published digest is not one that was scanned" || return 1
  grep -q "imagetools create.* --tag $name:latest" "$LOG" || bad "clean scans: $name:latest was not made" || return 1
  grep -q "imagetools create.* --tag $seed_name:$SEED_IMAGE_TAG" "$LOG" || bad "clean scans: the seed tag was not made" || return 1

  run_push "$t" ok '' 1
  [ "$rc" -ne 0 ] || bad "drift: a final tag with another digest exited 0" || return 1
}

# mutate NAME AWKPROGRAM  — a copy of the checkout's scripts under NAME with push.sh
# changed in exactly one place.
mutate() {
  m=$work/$1
  mkdir -p "$m/lab"
  cp -R "$ROOT/lab/scripts" "$m/lab/scripts"
  cp "$ROOT/lab/versions.env" "$m/lab/versions.env"
  awk "$2" "$ROOT/lab/scripts/push.sh" >"$m/lab/scripts/push.sh.new"
  mv "$m/lab/scripts/push.sh.new" "$m/lab/scripts/push.sh"
  if cmp -s "$ROOT/lab/scripts/push.sh" "$m/lab/scripts/push.sh"; then
    echo "check-lab-gates: FAIL: control $1 did not change push.sh" >&2
    fail=1
  fi
}

# control NAME  — push_ok must fail on the mutated copy.
control() {
  if push_ok "$work/$1"; then
    echo "check-lab-gates: FAIL: control $1: push_ok accepted the mutated push.sh" >&2
    fail=1
  else
    echo "check-lab-gates: control ok: $1 ($why)"
  fi
}

tree=$work/real
mkdir -p "$tree/lab"
cp -R "$ROOT/lab/scripts" "$tree/lab/scripts"
cp "$ROOT/lab/versions.env" "$tree/lab/versions.env"
if ! push_ok "$tree"; then
  echo "check-lab-gates: FAIL: push: $why" >&2
  fail=1
fi

# Negative controls. Each changes exactly one line of push.sh.
mutate c_gate '/^\[ "\$refused" -eq 0 \]/ { next } { print }'
control c_gate
mutate c_first_platform_only '{ sub(/\| tr .,. . .\)/, "| cut -d, -f1)"); print }'
control c_first_platform_only
mutate c_scan_the_tag '{ sub(/"registry:\/\/\$repo@\$d"/, "\"registry://$stage\""); print }'
control c_scan_the_tag
mutate c_rebuild_final_tag '{ print } /^\[ "\$refused" -eq 0 \]/ { print "docker buildx build --builder proven-builder --push --tag \"$name:latest\" \"$ROOT\"" }'
control c_rebuild_final_tag
mutate c_publish_other_digest '{ sub(/"\$\{proven_stage%:\*\}" "\$proven_index"/, "\"${proven_stage%:*}\" \"$seed_index\""); print }'
control c_publish_other_digest
mutate c_no_readback '/got=\$\(index_digest/ { print "    got=$digest"; next } { print }'
control c_no_readback

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "check-lab-gates: ok"
