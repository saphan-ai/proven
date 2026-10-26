#!/bin/sh
# Check the closing banner offline (no Docker, no network). Four things must hold:
#   1. with the settings as lab/versions.env has them, the banner is the owner's block, line for line;
#   2. when stdout is not a terminal, the Password line names the file and the password text is nowhere in the output;
#   3. on a terminal, the Password line carries the password;
#   4. every URL comes from the settings: change a setting and the line changes, the others do not.
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
Start here: http://127.0.0.1:3002/lab
########################################
####    Your proven lab is ready    ####
########################################

To see demo pull requests scanned by proven please open:

Gitea: http://127.0.0.1:3000/login
Username: proven
Password: PASSWORD

Grafana dashboard: http://127.0.0.1:3001/d/proven-review
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
last=$(printf '%s\n' "$out" | tail -n 1)
if [ "$last" != "Grafana dashboard: http://127.0.0.1:3001/d/proven-review" ]; then
  bad "something is printed after the Grafana line: $last"
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
  if [ "$tty_out" != "$want_tty" ]; then
    bad "the terminal banner is not the owner's block with the password on the Password line"
    printf '%s\n' "$tty_out" >&2
  fi
fi

# 4: the URLs come from the settings.
set_out=$(emit_piped env LAB_BIND=10.1.2.3 LAB_PAGE_BIND=10.1.2.4 GITEA_HTTP_PORT=4000 LGTM_GRAFANA_PORT=4001 LGTM_DASHBOARD_UID=other-uid \
  LAB_PAGE_PORT=4002 LAB_PAGE_PATH=/elsewhere LAB_ADMIN_USER=someone COMPOSE_PROJECT_NAME=mylab)
for line in \
  'Start here: http://10.1.2.4:4002/elsewhere' \
  'Gitea: http://10.1.2.3:4000/login' \
  'Username: someone' \
  'Password: file /state/password in the volume mylab_seed-state' \
  'Grafana dashboard: http://10.1.2.3:4001/d/other-uid'; do
  if ! printf '%s\n' "$set_out" | grep -qxF "$line"; then
    bad "with other settings the banner lacks the line: $line"
  fi
done
case $set_out in
  *127.0.0.1* | *3000* | *3001* | *3002*) bad "a URL in the banner is not built from the settings: a default survived" ;;
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
