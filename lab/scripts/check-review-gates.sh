#!/bin/sh
# Offline check of five properties of the lab, from a reviewer's shell check. It
# starts no container: docker and git are stubs on the PATH. Run from the
# repository root: make check-review-gates. It prints one PASS or FAIL line per
# property and `failed=N`, and exits non-zero when N is not 0.
#
# Each assertion reads the property, not its spelling, and has a negative
# control: the same assertion run on a copy of the tree with exactly the guarded
# line changed, which must read FAIL.
#   1. an environment LAB_BIND that is not loopback does not reach compose.
#      The guard refuses; it does not reset the value. The assertion reads the
#      refusal.
#   2. the admin password is not printed when stdout is not a terminal, and is
#      kept in a file at mode 0600.
#   3. with Docker Scout missing, make scan fails and builds nothing.
#   4. with Docker Scout missing or reporting a finding, make push fails and
#      never reaches `--push`.
#   5. the build context does not carry the .saphan work area.
set -u
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT=${ROOT:-$(CDPATH= cd -- "$here/../.." && pwd)}
export ROOT

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
cat >"$work/bin/git" <<'STUB'
#!/bin/sh
echo stub
STUB
chmod +x "$work/bin/docker" "$work/bin/git"

# tree DIR  — a copy of the files the scripts read. The scripts run against the
# copy, so a mutated line never touches the checkout.
tree() {
  mkdir -p "$1/lab" "$1/docs"
  cp -R "$ROOT/lab/scripts" "$1/lab/scripts"
  cp "$ROOT/lab/versions.env" "$1/lab/versions.env"
  cp "$ROOT/docs/LAB.md" "$1/docs/LAB.md"
  cp "$ROOT/.dockerignore" "$1/.dockerignore"
}

# run_script TREE SCOUT SCRIPT  — runs a lab script with the stubs; sets rc.
run_script() {
  : >"$LOG"
  rc=0
  sh -c 'ROOT=$1; STUB_SCOUT=$2; PATH=$3/bin:$PATH; export ROOT STUB_SCOUT PATH LOG; sh "$ROOT/lab/scripts/$4"' \
    sh "$1" "$2" "$work" "$3" >"$work/out" 2>"$work/err" </dev/null || rc=$?
}

# --- 1. LAB_BIND ----------------------------------------------------------
a1() {
  sh -c '
    ROOT=$1; LAB_BIND=0.0.0.0; export ROOT LAB_BIND
    . "$ROOT/lab/scripts/lib.sh"
    [ "$LAB_BIND" = 127.0.0.1 ] && exit 0
    ( require_loopback_bind ) >/dev/null 2>&1 && exit 1
    exit 0
  ' sh "$1" || return 1
  # The default stays loopback and is accepted.
  sh -c '
    ROOT=$1; export ROOT; unset LAB_BIND
    . "$ROOT/lab/scripts/lib.sh"
    [ "$LAB_BIND" = 127.0.0.1 ] && ( require_loopback_bind ) >/dev/null 2>&1
  ' sh "$1"
}

# --- 2. the password ------------------------------------------------------
secret=Zq7-reviewer-check-password
a2() {
  state=$work/a2state
  rm -rf "$state"
  sh -c '
    ROOT=$1; LAB_STATE_DIR=$2; export ROOT LAB_STATE_DIR
    . "$ROOT/lab/scripts/lib.sh"
    deliver_admin_password "$3"
  ' sh "$1" "$state" "$secret" >"$work/a2out" 2>"$work/a2err" </dev/null || return 1
  ! grep -qF -- "$secret" "$work/a2out" "$work/a2err" || return 1
  [ -s "$state/admin-password" ] && grep -qF -- "$secret" "$state/admin-password" || return 1
  [ "$(ls -ld "$state/admin-password" | cut -c1-10)" = "-rw-------" ] || return 1
  # up.sh prints no password itself: the only line that holds it besides the
  # pipe into the state volume is the call to deliver_admin_password.
  ! grep -E '(echo|printf)[^|]*\$\{?pass\}?' "$1/lab/scripts/up.sh" | grep -vq '| compose'
}

# --- 3. Scout missing is not success ---------------------------------------
a3() {
  run_script "$1" absent scan.sh
  [ "$rc" -ne 0 ] || return 1
  # Nothing but the Scout probe reached docker.
  ! grep -vE '^docker (scout version|version)' "$LOG" | grep -q .
}

# --- 4. push requires a passing scan --------------------------------------
a4() {
  for s in absent fail; do
    run_script "$1" "$s" push.sh
    [ "$rc" -ne 0 ] || return 1
    ! grep -q -e ' --push' "$LOG" || return 1
    ! grep -q 'buildx create' "$LOG" || return 1
  done
}

# --- 5. the build context ignores the work area ---------------------------
# a5 FILE  — .dockerignore rules, last match wins; ignored when the last rule
# that matches .saphan is not a re-include. A re-include of a path under .saphan
# opens the area.
a5() {
  awk '
    { sub(/\r$/, "") }
    /^[ \t]*(#|$)/ { next }
    {
      neg = 0; pat = $0
      if (pat ~ /^!/) { neg = 1; pat = substr(pat, 2) }
      sub(/^\//, "", pat); sub(/\/$/, "", pat)
      if (pat == "*" || pat == "**" || pat == ".*" || pat == ".saphan" || pat == "**/.saphan" || pat == ".saphan/**" || pat == "**/.saphan/**") {
        state = neg ? "in" : "out"
      } else if (neg && (pat ~ /^(\*\*\/)?[.]saphan\//)) {
        state = "in"
      }
    }
    END { exit (state == "out" ? 0 : 1) }
  ' "$1"
}

fail=0
name=
check() {
  if "$@"; then
    printf 'PASS %s\n' "$name"
  else
    printf 'FAIL %s\n' "$name"
    fail=$((fail + 1))
  fi
}

# control NAME CMD...  — the command must fail on the mutated input.
cfail=0
control() {
  cname=$1
  shift
  if "$@"; then
    printf 'CONTROL-FAIL %s (the assertion accepted the mutated input)\n' "$cname"
    cfail=$((cfail + 1))
  else
    printf 'CONTROL ok %s\n' "$cname"
  fi
}

tree "$work/real"

name=localhost_only_when_environment_overrides_default
check a1 "$work/real"
name=password_not_printed
check a2 "$work/real"
name=scout_unavailable_is_not_success
check a3 "$work/real"
name=push_requires_scan
check a4 "$work/real"
name=build_context_ignores_work_area
check a5 "$ROOT/.dockerignore"

# Negative controls. Each copy changes exactly the line the assertion guards.
tree "$work/n1"
sed 's/"\${LAB_BIND-}" | awk/127.0.0.1 | awk/' "$work/n1/lab/scripts/lib.sh" >"$work/n1/lib.new"
mv "$work/n1/lib.new" "$work/n1/lab/scripts/lib.sh"
control a1_guard_always_sees_loopback a1 "$work/n1"

tree "$work/n2"
sed 's/^    store_admin_password_file "\$1"$/    echo "$1"/' "$work/n2/lab/scripts/lib.sh" >"$work/n2/lib.new"
mv "$work/n2/lib.new" "$work/n2/lab/scripts/lib.sh"
control a2_function_prints_password a2 "$work/n2"
tree "$work/n2b"
echo 'echo "lab admin password: $pass"' >>"$work/n2b/lab/scripts/up.sh"
control a2_up_prints_password a2 "$work/n2b"

tree "$work/n3"
sed '/if ! docker scout version/,/^fi/s/exit 1/exit 0/' "$work/n3/lab/scripts/scout-images.sh" >"$work/n3/new"
mv "$work/n3/new" "$work/n3/lab/scripts/scout-images.sh"
control a3_scout_missing_exits_zero a3 "$work/n3"

tree "$work/n4"
sed '/scout-images.sh/d' "$work/n4/lab/scripts/push.sh" >"$work/n4/new"
mv "$work/n4/new" "$work/n4/lab/scripts/push.sh"
control a4_push_without_scan a4 "$work/n4"

printf '*\n!go.mod\n!.saphan\n' >"$work/di1"
control a5_work_area_re_included a5 "$work/di1"
printf '!go.mod\n.env\n' >"$work/di2"
control a5_no_rule_for_work_area a5 "$work/di2"

printf 'failed=%s\n' "$fail"
[ "$fail" -eq 0 ] && [ "$cfail" -eq 0 ]
