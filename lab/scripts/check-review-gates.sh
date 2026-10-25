#!/bin/sh
# Offline check of five properties of the lab, from a reviewer's shell check. It
# starts no container: docker and git are stubs on the PATH. Run from the
# repository root: make check-review-gates. It prints one PASS or FAIL line per
# property and `failed=N`, and exits non-zero when N is not 0.
#
# Each assertion reads the property, not its spelling, and has a negative
# control: the same assertion run on a copy of the tree with exactly the guarded
# line changed, which must read FAIL.
#   1. no setting publishes the lab beyond loopback: the compose file names a
#      literal loopback address in each port line (read from the file alone), and
#      the scripts refuse the retired LAB_BIND.
#   2. the admin password is not printed when stdout is not a terminal, and is
#      kept in a file at mode 0600.
#   3. with Docker Scout missing, make scan fails and builds nothing.
#   4. with Docker Scout missing or reporting a finding, make push fails, makes
#      no final tag and builds only staging tags.
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

cp "$ROOT/lab/scripts/stub-docker.sh" "$work/bin/docker"
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
  cp "$ROOT/lab/compose.yaml" "$1/lab/compose.yaml"
  cp "$ROOT/docs/LAB.md" "$1/docs/LAB.md"
  cp "$ROOT/.dockerignore" "$1/.dockerignore"
}

# run_script TREE SCOUT SCRIPT  — runs a lab script with the stubs; sets rc.
run_script() {
  : >"$LOG"
  rm -rf "$LOG.state"
  rc=0
  sh -c 'ROOT=$1; STUB_SCOUT=$2; PATH=$3/bin:$PATH; export ROOT STUB_SCOUT PATH LOG; sh "$ROOT/lab/scripts/$4"' \
    sh "$1" "$2" "$work" "$3" >"$work/out" 2>"$work/err" </dev/null || rc=$?
}

# --- 1. loopback ----------------------------------------------------------
# a1 TREE  — the compose file is loopback by itself, because
# `docker compose -f lab/compose.yaml up` runs no lab script; and the scripts
# refuse the retired LAB_BIND and take their address from a constant.
# a1_compose FILE reads only the port lines of the file: each must name a literal
# 127.x.x.x or [::1] host address, with no variable in it. A port block written
# inline, a service on the host network, or a file with no port line is FAIL.
a1_compose() {
  awk '
    { sub(/\r$/, "") }
    /^[ \t]*network_mode:[ \t]*"?host"?[ \t]*$/ { bad = 1 }
    inports && /^[ \t]*(#|$)/ { next }
    inports && /^[ \t]*-/ {
      line = $0; sub(/^[ \t]*-[ \t]*/, "", line)
      if (line ~ /^"?(127(\.[0-9]+)(\.[0-9]+)(\.[0-9]+)|\[::1\]):[0-9$]/) found++
      else bad = 1
      next
    }
    { inports = 0 }
    /^[ \t]*ports:/ {
      rest = $0; sub(/^[ \t]*ports:[ \t]*/, "", rest)
      if (rest != "" && rest !~ /^#/) bad = 1
      inports = 1
    }
    END { exit (bad || found < 3) ? 1 : 0 }
  ' "$1"
}

a1() {
  a1_compose "$1/lab/compose.yaml" || return 1
  sh -c '
    ROOT=$1; LAB_BIND=0.0.0.0; LAB_HOST=0.0.0.0; export ROOT LAB_BIND LAB_HOST
    . "$ROOT/lab/scripts/lib.sh"
    [ "$LAB_HOST" = 127.0.0.1 ] || exit 1
    ( refuse_retired_bind ) >/dev/null 2>&1 && exit 1
    exit 0
  ' sh "$1" || return 1
  # With nothing set the guard lets the lab start.
  sh -c '
    ROOT=$1; export ROOT; unset LAB_BIND
    . "$ROOT/lab/scripts/lib.sh"
    ( refuse_retired_bind ) >/dev/null 2>&1
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
  run_script "$1" absent push.sh
  [ "$rc" -ne 0 ] || return 1
  ! grep -q -e 'buildx build' -e 'buildx create' -e 'imagetools create' "$LOG" || return 1
  run_script "$1" fail push.sh
  [ "$rc" -ne 0 ] || return 1
  ! grep -q 'imagetools create' "$LOG" || return 1
  # Every build names a staging tag, never a final one.
  ! grep 'buildx build' "$LOG" | grep -vq -e '--tag [^ ]*:staging-' || return 1
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

name=loopback_in_the_compose_file_and_bind_refused
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
sed 's/"\${LAB_BIND+set}" = set/"${LAB_BIND+set}" = never/' "$work/n1/lab/scripts/lib.sh" >"$work/n1/lib.new"
mv "$work/n1/lib.new" "$work/n1/lab/scripts/lib.sh"
control a1_guard_never_refuses a1 "$work/n1"

# The compose file is the entry point: each copy puts one port line back.
tree "$work/n1b"
sed '0,/"127.0.0.1:/s//"0.0.0.0:/' "$work/n1b/lab/compose.yaml" >"$work/n1b/new"
mv "$work/n1b/new" "$work/n1b/lab/compose.yaml"
control a1_compose_port_on_all_interfaces a1 "$work/n1b"
tree "$work/n1c"
sed '0,/"127.0.0.1:/s//"${LAB_BIND}:/' "$work/n1c/lab/compose.yaml" >"$work/n1c/new"
mv "$work/n1c/new" "$work/n1c/lab/compose.yaml"
control a1_compose_port_on_a_variable a1 "$work/n1c"
tree "$work/n1d"
sed '0,/"127.0.0.1:/s//"/' "$work/n1d/lab/compose.yaml" >"$work/n1d/new"
mv "$work/n1d/new" "$work/n1d/lab/compose.yaml"
control a1_compose_port_without_address a1 "$work/n1d"

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
sed '/^\[ "\$refused" -eq 0 \]/d' "$work/n4/lab/scripts/push.sh" >"$work/n4/new"
mv "$work/n4/new" "$work/n4/lab/scripts/push.sh"
control a4_push_without_scan a4 "$work/n4"

printf '*\n!go.mod\n!.saphan\n' >"$work/di1"
control a5_work_area_re_included a5 "$work/di1"
printf '!go.mod\n.env\n' >"$work/di2"
control a5_no_rule_for_work_area a5 "$work/di2"

printf 'failed=%s\n' "$fail"
[ "$fail" -eq 0 ] && [ "$cfail" -eq 0 ]
