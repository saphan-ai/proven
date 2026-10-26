#!/bin/sh
# Check the closing banner offline (no Docker, no network). Four things must hold:
#   1. with the settings as lab/versions.env has them, the banner is the owner's block, line for line;
#   2. when stdout is not a terminal, the Password line names the file and the password text is nowhere in the output;
#   3. on a terminal, the Password line carries the password;
#   4. every URL comes from the settings: change a setting and the line changes, the others do not. The address is
#      127.0.0.1 and no setting moves it: a LAB_BIND or LAB_PAGE_BIND in the environment leaves it where it is.
# The terminal case needs script(1). Without it the check says so and exits 1, unless BANNER_CHECK_NO_TTY_OK=1.
# `lab_banner` reads the terminal test itself, so the same function is run both ways.
set -eu

here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT=$(CDPATH= cd -- "$here/../.." && pwd)
export ROOT

# --emit: print the banner with the test password. Used under script(1) for the terminal case.
if [ "${1:-}" = "--emit" ]; then
  . "$here/lib.sh"
  . "$here/banner.sh"
  lab_banner "$BANNER_CHECK_PASSWORD"
  exit 0
fi

# --no-tty-ok is BANNER_CHECK_NO_TTY_OK=1: the terminal case may be named as not run.
if [ "${1:-}" = "--no-tty-ok" ]; then
  BANNER_CHECK_NO_TTY_OK=1
fi

secret=banner-check-secret-4f9a1c
fail=0
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

bad() {
  echo "banner-check: FAIL: $1" >&2
  fail=1
}

# The settings the banner reads, cleared so lab/versions.env decides.
clean_env() {
  env -u LAB_BIND -u LAB_PAGE_BIND -u GITEA_HTTP_PORT -u LGTM_GRAFANA_PORT -u LGTM_DASHBOARD_UID -u LAB_PAGE_PORT -u LAB_PAGE_PATH \
    -u LAB_ADMIN_USER -u COMPOSE_PROJECT_NAME "$@"
}

emit_piped() {
  # command substitution: stdout is a pipe, never a terminal
  clean_env BANNER_CHECK_PASSWORD="$secret" "$@" sh "$here/banner-check.sh" --emit
}

expected_defaults() {
  cat <<'EOF'
########################################
####    Your proven lab is ready    ####
########################################

To see demo pull requests scanned by proven please open:

Gitea: http://127.0.0.1:3300/user/login
Username: proven
Password: PASSWORD

Grafana dashboard: http://127.0.0.1:3301/d/proven-overview
Wait for every review: make demo-check
Help: https://github.com/saphan-ai/proven/blob/main/docs/WALKTHROUGH.md

>>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<
EOF
}

location='file /state/password in the volume proven-lab_seed-state'

# 1 and 2: redirected, with the defaults.
out=$(emit_piped env)
want=$(expected_defaults | sed "s|^Password: PASSWORD\$|Password: $location|")
if [ "$out" != "$want" ]; then
  bad "the redirected banner is not the owner's block with the file location on the Password line"
  printf '%s\n' "$out" >&2
fi
case $out in
  *"$secret"*) bad "the password text is in the redirected output" ;;
esac
# The banner ends with the START HERE line and exactly ONE empty line after it. Command substitution strips trailing
# newlines, so this reads the bytes from a file.
emit_piped env >"$work/banner.out"
if [ "$(tail -n 2 "$work/banner.out" | head -n 1)" != ">>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<" ] || [ -n "$(tail -n 1 "$work/banner.out")" ] || [ "$(tail -c 2 "$work/banner.out" | od -An -tx1 | tr -d ' ')" != "0a0a" ]; then
  bad "the banner does not end with the START HERE line and exactly one empty line"
fi
if grep -q "$(printf '\033')" "$work/banner.out"; then
  bad "the redirected banner holds a control character"
fi

# 3: on a terminal.
tty_out=
if script --version 2>&1 | grep -q util-linux; then
  tty_out=$(clean_env BANNER_CHECK_PASSWORD="$secret" script -qec "sh $here/banner-check.sh --emit" /dev/null | tr -d '\r')
elif script -q /dev/null true >/dev/null 2>&1; then
  # BSD script(1) (macOS) reads the terminal settings from ITS OWN stdin and stops when stdin is not a terminal,
  # which is how `go test` and a pipe run this file. So the case runs there only from a terminal; otherwise it is
  # skipped by name below, never reported as passed. util-linux script(1) above makes its own pseudo-terminal.
  if [ -t 0 ]; then
    tty_out=$(clean_env BANNER_CHECK_PASSWORD="$secret" script -q /dev/null sh "$here/banner-check.sh" --emit | tr -d '\r')
  else
    bsd_no_stdin_tty=1
  fi
fi
if [ -z "$tty_out" ] || [ -n "${bsd_no_stdin_tty:-}" ]; then
  if [ -n "${bsd_no_stdin_tty:-}" ] && [ "${BANNER_CHECK_NO_TTY_OK:-}" = "1" ]; then
    echo "banner-check: terminal case NOT RUN: BSD script(1) needs a terminal on stdin and this run has none; run make check-banner from a terminal (BANNER_CHECK_NO_TTY_OK=1)"
  elif [ "${BANNER_CHECK_NO_TTY_OK:-}" = "1" ]; then
    echo "banner-check: terminal case NOT RUN: script(1) is not available (BANNER_CHECK_NO_TTY_OK=1)"
  else
    bad "terminal case not run: script(1) is not available; set BANNER_CHECK_NO_TTY_OK=1 to accept that"
  fi
else
  want_tty=$(expected_defaults | sed "s|^Password: PASSWORD\$|Password: $secret|")
  esc=$(printf '\033')
  plain_tty=$(printf '%s\n' "$tty_out" | sed "s/${esc}\[[0-9;]*m//g")
  if [ "$plain_tty" != "$want_tty" ]; then
    bad "the terminal banner is not the owner's block with the password on the Password line"
    printf '%s\n' "$tty_out" >&2
  fi
  # On a terminal the START HERE line is bold and green, and only that line carries colour.
  if ! printf '%s\n' "$tty_out" | grep -q "^${esc}\[1;32m>>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<${esc}\[0m\$"; then
    bad "on a terminal the START HERE line is not bold green"
  fi
  if [ "$(printf '%s\n' "$tty_out" | grep -c "$esc")" -ne 1 ]; then
    bad "on a terminal something besides the START HERE line carries colour"
  fi
  # NO_COLOR is respected on a terminal.
  if script --version 2>&1 | grep -q util-linux; then
    nc_out=$(clean_env NO_COLOR=1 BANNER_CHECK_PASSWORD="$secret" script -qec "sh $here/banner-check.sh --emit" /dev/null | tr -d '\r')
    if printf '%s\n' "$nc_out" | grep -q "$esc"; then
      bad "NO_COLOR is set and the terminal banner still holds a control character"
    fi
  fi
fi

# 4: the URLs come from the settings.
set_out=$(emit_piped env LAB_BIND=10.1.2.3 LAB_PAGE_BIND=10.1.2.4 LAB_HOST=10.1.2.5 GITEA_HTTP_PORT=4000 LGTM_GRAFANA_PORT=4001 LGTM_DASHBOARD_UID=other-uid \
  LAB_PAGE_PORT=4002 LAB_PAGE_PATH=/elsewhere LAB_ADMIN_USER=someone COMPOSE_PROJECT_NAME=mylab)
for line in \
  '>>>>>>>>  START HERE:  http://127.0.0.1:4002/elsewhere  <<<<<<<<' \
  'Gitea: http://127.0.0.1:4000/user/login' \
  'Username: someone' \
  'Password: file /state/password in the volume mylab_seed-state' \
  'Grafana dashboard: http://127.0.0.1:4001/d/other-uid'; do
  if ! printf '%s\n' "$set_out" | grep -qxF "$line"; then
    bad "with other settings the banner lacks the line: $line"
  fi
done
case $set_out in
  *3300* | *3301* | *3302*) bad "a URL in the banner is not built from the settings: a default survived" ;;
esac
case $set_out in
  *10.1.2.*) bad "an address setting moved a URL in the banner off 127.0.0.1" ;;
esac

# The page's path is written in two places: the setting and the nginx file. They must agree.
path=$(clean_env sh -c ". '$here/lib.sh'; printf '%s' \"\$LAB_PAGE_PATH\"")
if ! grep -q "location = $path {" "$ROOT/lab/page/nginx.conf"; then
  bad "lab/page/nginx.conf does not serve LAB_PAGE_PATH ($path)"
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "banner-check: ok"
