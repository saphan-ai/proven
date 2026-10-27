#!/bin/sh
# Offline check of how the lab's admin password reaches the judge, with a stub
# docker on PATH. No Docker, no network. Run: make check-lab-password
#   1. stdout not a terminal: no password text on stdout or stderr; the password
#      is in a file at mode 0600 in the state directory, and only the path is
#      printed. An older, wider file at that path is replaced by a 0600 one.
#   2. stdout a terminal (a pty, when python3 is on the PATH): the password is
#      printed once and no file is written.
#   3. a password the judge supplies in LAB_ADMIN_PASSWORD is used, is not
#      printed and is not written to a file; one that is too short is refused
#      before any container starts.
#   4. a later make up prints no password and writes no file.
#   5. make clean removes the file.
# Each check has a negative control that mutates the one field it guards.
set -eu
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT=${ROOT:-$(CDPATH= cd -- "$here/../.." && pwd)}
export ROOT

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir "$work/bin"
LOG=$work/docker.log
LAB_STATE_DIR=$work/state
export LOG LAB_STATE_DIR

# STUB_PASSWORD: what `cat /state/password` returns, as a lab that is installed.
cat >"$work/bin/docker" <<'STUB'
#!/bin/sh
echo "docker $*" >>"$LOG"
case $* in
  *' ps -q'*) echo stubcontainer ;;
  *'--entrypoint cat seed /state/password'*) [ -n "${STUB_PASSWORD:-}" ] && printf '%s' "$STUB_PASSWORD" ;;
  'version'*) echo linux/amd64 ;;
esac
exit 0
STUB
chmod +x "$work/bin/docker"
# The real docker, when there is one, for the live Compose render in section 6.
real_docker=$(command -v docker 2>/dev/null || true)
PATH=$work/bin:$PATH
export PATH
# A terminal must never be asked a question by this check: the mode is set, and the ports are ones nothing holds.
PROVEN_LAB_MODE=sample
GITEA_HTTP_PORT=53300
LGTM_GRAFANA_PORT=53301
LAB_PAGE_PORT=53302
LAB_DOCS_PORT=53303
LGTM_OTLP_PORT=54318
export PROVEN_LAB_MODE GITEA_HTTP_PORT LGTM_GRAFANA_PORT LAB_PAGE_PORT LAB_DOCS_PORT LGTM_OTLP_PORT

fail=0
bad() {
  echo "check-lab-password: FAIL: $1" >&2
  fail=1
}
expect() {
  [ "$2" = "$3" ] || bad "$1: got '$2', want '$3'"
}

# contains FILE TEXT  — yes or no.
contains() {
  if grep -qF -- "$2" "$1"; then echo yes; else echo no; fi
}

# mode_of FILE  — the file's permission string, the same on Linux and macOS.
mode_of() {
  ls -ld "$1" | cut -c1-10
}

secret=Zq7-not-a-real-password-91
supplied=Judge-Own-Passw0rd

# --- 1. stdout redirected -------------------------------------------------
rm -rf "$LAB_STATE_DIR"
sh -c '. "$ROOT/lab/scripts/lib.sh"; deliver_admin_password "$1"' sh "$secret" >"$work/out" 2>"$work/err" || bad "redirected: deliver_admin_password failed: $(cat "$work/err")"
file=$LAB_STATE_DIR/admin-password
expect "redirected: stdout holds the password" "$(contains "$work/out" "$secret")" no
expect "redirected: stderr holds the password" "$(contains "$work/err" "$secret")" no
expect "redirected: stdout names the file" "$(contains "$work/out" "$file")" yes
expect "redirected: the file holds the password" "$(contains "$file" "$secret")" yes
expect "redirected: the file mode" "$(mode_of "$file")" "-rw-------"
expect "redirected: the directory mode" "$(mode_of "$LAB_STATE_DIR")" "drwx------"

# Negative controls: the same checks must read red on the mutated field.
printf '%s\n' "lab admin password: $secret" >"$work/leak"
expect "negative control: a leaked password is seen" "$(contains "$work/leak" "$secret")" yes
chmod 644 "$file"
expect "negative control: a 0644 file is seen" "$(mode_of "$file")" "-rw-r--r--"

# An older file at the path, wider and with another value, is replaced.
printf 'old\n' >"$file"
chmod 666 "$file"
sh -c '. "$ROOT/lab/scripts/lib.sh"; deliver_admin_password "$1"' sh "$secret" >"$work/out" 2>"$work/err" || bad "redirected: deliver_admin_password failed: $(cat "$work/err")"
expect "replaced: the file mode" "$(mode_of "$file")" "-rw-------"
expect "replaced: the old value is gone" "$(contains "$file" old)" no
expect "replaced: the new value is there" "$(contains "$file" "$secret")" yes

# --- 2. stdout a terminal -------------------------------------------------
if command -v python3 >/dev/null 2>&1; then
  rm -rf "$LAB_STATE_DIR"
  PTY_SECRET=$secret
  export PTY_SECRET
  rc=0
  python3 -I -c '
import os, pty, sys
secret = os.environ["PTY_SECRET"]
cmd = ["sh", "-c", ". \"$ROOT/lab/scripts/lib.sh\"; deliver_admin_password \"$1\"", "sh", secret]
try:
    pid, fd = pty.fork()
except OSError:
    sys.exit(77)
if pid == 0:
    os.execvp(cmd[0], cmd)
out = b""
while True:
    try:
        chunk = os.read(fd, 4096)
    except OSError:
        break
    if not chunk:
        break
    out += chunk
os.waitpid(pid, 0)
sys.stdout.write(out.decode())
' >"$work/pty" 2>&1 || rc=$?
  case $rc in
    0)
      expect "terminal: the password is printed" "$(contains "$work/pty" "$secret")" yes
      expect "terminal: it is printed once" "$(grep -cF -- "$secret" "$work/pty")" 1
      expect "terminal: no file is written" "$([ -e "$file" ] && echo yes || echo no)" no
      ;;
    77) echo "check-lab-password: UNRUN: the terminal case: this host opens no pty" >&2 ;;
    *) bad "terminal: the pty run failed with exit $rc" ;;
  esac
else
  echo "check-lab-password: UNRUN: the terminal case needs python3 for a pty" >&2
fi

# --- 3 and 4. the whole of up.sh, against the stub ------------------------
run_up() {
  : >"$LOG"
  rc=0
  sh "$ROOT/lab/scripts/up.sh" >"$work/out" 2>"$work/err" </dev/null || rc=$?
}

# The stub prints nothing for `cat /state/password`, so this is a first start.
unset LAB_ADMIN_PASSWORD STUB_PASSWORD LAB_BIND
rm -rf "$LAB_STATE_DIR"
run_up
expect "generated, redirected: exit" "$rc" 0
expect "generated, redirected: the flow starts a sample lab" "$(grep -c "^up: sample mode" "$work/out")" 1
expect "generated, redirected: no question is printed" "$(grep -c "\[sample/live\]" "$work/out")" 0
expect "generated, redirected: the ready banner ends the output" "$(tail -n 1 "$work/out" | cut -c1-17)" "Grafana dashboard"
expect "generated, redirected: the file exists" "$([ -s "$file" ] && echo yes || echo no)" yes
expect "generated, redirected: the file mode" "$(mode_of "$file")" "-rw-------"
gen=$(cat "$file")
expect "generated: the password is not empty" "$([ -n "$gen" ] && echo yes || echo no)" yes
expect "generated, redirected: stdout holds it" "$(contains "$work/out" "$gen")" no
expect "generated, redirected: stderr holds it" "$(contains "$work/err" "$gen")" no
expect "generated: it is the one the admin was created with" "$(contains "$LOG" "--password $gen")" yes

# A lab that is installed: no password printed, no file written.
rm -rf "$LAB_STATE_DIR"
STUB_PASSWORD=$gen
export STUB_PASSWORD
run_up
expect "later run: exit" "$rc" 0
expect "later run: stdout holds the password" "$(contains "$work/out" "$gen")" no
expect "later run: stderr holds the password" "$(contains "$work/err" "$gen")" no
expect "later run: a file is written" "$([ -e "$file" ] && echo yes || echo no)" no
unset STUB_PASSWORD

# A supplied password: used, not printed, not written.
LAB_ADMIN_PASSWORD=$supplied
export LAB_ADMIN_PASSWORD
run_up
expect "supplied: exit" "$rc" 0
expect "supplied: stdout holds it" "$(contains "$work/out" "$supplied")" no
expect "supplied: stderr holds it" "$(contains "$work/err" "$supplied")" no
expect "supplied: it is the one the admin was created with" "$(contains "$LOG" "--password $supplied")" yes
expect "supplied: a file is written" "$([ -e "$file" ] && echo yes || echo no)" no
# Negative control: the stub log does hold it, so the "not printed" checks above
# read a place where it would show if the script printed it.
expect "negative control: the log is a place the password shows" "$(contains "$LOG" "$supplied")" yes

# A supplied password that is too short is refused before any container.
LAB_ADMIN_PASSWORD=short
export LAB_ADMIN_PASSWORD
run_up
expect "short supplied: refused" "$([ "$rc" -ne 0 ] && echo refused || echo ran)" refused
expect "short supplied: touched docker" "$([ -s "$LOG" ] && echo yes || echo no)" no
unset LAB_ADMIN_PASSWORD

# --- 5. make clean --------------------------------------------------------
# clean.sh also removes ./bin and ./dist of the checkout, so this check runs the
# function it calls, and reads that clean.sh does call it.
rm -rf "$LAB_STATE_DIR"
run_up
expect "clean: the file exists before" "$([ -e "$file" ] && echo yes || echo no)" yes
sh -c '. "$ROOT/lab/scripts/lib.sh"; remove_admin_password_file'
expect "clean: the file is gone" "$([ -e "$file" ] && echo yes || echo no)" no
expect "clean: the directory is gone" "$([ -e "$LAB_STATE_DIR" ] && echo yes || echo no)" no
expect "clean: clean.sh removes the file" "$(contains "$ROOT/lab/scripts/clean.sh" remove_admin_password_file)" yes
# Negative control: a file that is not removed is seen.
rm -rf "$LAB_STATE_DIR"
run_up
expect "negative control: the file is seen when it is not removed" "$([ -e "$file" ] && echo yes || echo no)" yes

# --- 6. the password never enters the Compose environment ------------------
# Compose renders every value it interpolates, so a supplied password must not be
# named in lab/compose.yaml at all: the seed reads it from /state/password.
# names_password FILE  — yes when FILE names the variable.
names_password() {
  if grep -q 'LAB_ADMIN_PASSWORD' "$1"; then echo yes; else echo no; fi
}
expect "compose: the file names the password variable" "$(names_password "$ROOT/lab/compose.yaml")" no
expect "compose: up.sh passes the password to compose run" "$(grep -c 'LAB_ADMIN_PASSWORD=.*compose' "$ROOT/lab/scripts/up.sh")" 0
expect "compose: the seed reads /state/password" "$(contains "$ROOT/lab/seed/seed.sh" 'cat /state/password')" yes
# Negative control: the same file with the interpolation put back is seen.
awk '{ print } /^    environment:$/ && !done { print "      LAB_ADMIN_PASSWORD: ${LAB_ADMIN_PASSWORD:-}"; done = 1 }' "$ROOT/lab/compose.yaml" >"$work/compose.mutant"
expect "negative control: the interpolation is seen" "$(names_password "$work/compose.mutant")" yes

# Live: render the real config with a supplied password and read it for the value.
if [ -n "$real_docker" ] && "$real_docker" compose version >/dev/null 2>&1; then
  rc=0
  LAB_ADMIN_PASSWORD=$supplied "$real_docker" compose --env-file "$ROOT/lab/versions.env" \
    -f "$ROOT/lab/compose.yaml" --project-directory "$ROOT/lab" config >"$work/rendered" 2>&1 || rc=$?
  expect "live config: renders" "$rc" 0
  expect "live config: the rendered file holds the password" "$(contains "$work/rendered" "$supplied")" no
  LAB_ADMIN_PASSWORD=$supplied "$real_docker" compose --env-file "$ROOT/lab/versions.env" \
    -f "$work/compose.mutant" --project-directory "$ROOT/lab" config >"$work/rendered.mutant" 2>&1 || true
  expect "negative control: live config of the mutant holds the password" "$(contains "$work/rendered.mutant" "$supplied")" yes
else
  echo "check-lab-password: UNRUN: the live Compose render: docker compose is not on this host" >&2
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "check-lab-password: ok"
