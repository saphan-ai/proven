# Shared settings loader. Sourced by the lab scripts. ROOT is the repository root.
# The Makefile exports ROOT. A direct run of a lab/scripts file derives it.
if [ -z "${ROOT:-}" ]; then
  here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
  case $here in
    */lab/scripts)
      ROOT=$(CDPATH= cd -- "$here/../.." && pwd)
      export ROOT
      ;;
    *)
      echo "ROOT is not set" >&2
      exit 1
      ;;
  esac
fi

load_versions() {
  file=$ROOT/lab/versions.env
  if [ ! -f "$file" ]; then
    echo "missing $file" >&2
    exit 1
  fi
  while IFS= read -r line || [ -n "$line" ]; do
    case $line in
      '' | \#*) continue ;;
    esac
    key=${line%%=*}
    val=${line#*=}
    case $key in
      *[!A-Za-z0-9_]*)
        echo "lab/versions.env: refused key" >&2
        exit 1
        ;;
    esac
    eval "isset=\${$key+x}"
    if [ -z "$isset" ]; then
      export "$key=$val"
    fi
  done <"$file"
}

# The lab's published ports name 127.0.0.1 in lab/compose.yaml itself, so no
# setting and no way of starting that file moves them. LAB_HOST is that address,
# for the URLs the scripts print; it is a constant, never read from the
# environment. LAB_BIND, which once chose the address, is retired.
LAB_HOST=127.0.0.1

# refuse_retired_bind  — a LAB_BIND in the environment, empty or not, is refused
# by name and no container starts: whoever set it expected another address.
refuse_retired_bind() {
  if [ "${LAB_BIND+set}" = set ]; then
    echo "LAB_BIND=${LAB_BIND} is retired. The lab publishes on 127.0.0.1 only and no setting changes that. Unset LAB_BIND." >&2
    exit 1
  fi
}

# The admin password reaches the judge in one of two ways, and never through a
# log. On a terminal it is printed once, by deliver_admin_password. Anywhere
# else (CI, a redirect, a pipe) it is written to a file only the user can read,
# in the lab's state directory, and only the path is printed. make clean removes
# the file. LAB_STATE_DIR moves the directory.
stdout_is_terminal() {
  [ -t 1 ]
}

admin_password_file() {
  printf '%s\n' "${LAB_STATE_DIR:-$ROOT/lab/.state}/admin-password"
}

# store_admin_password_file PASSWORD  — the file is created at mode 0600 by the
# umask, never created wider and narrowed after. An older file or link at the
# path is removed first, so the mode is the new file's.
store_admin_password_file() {
  f=$(admin_password_file)
  d=$(dirname "$f")
  (
    umask 077
    mkdir -p "$d"
    rm -f "$f"
    printf '%s\n' "$1" >"$f"
  )
}

# deliver_admin_password PASSWORD  — called on the first make up only.
deliver_admin_password() {
  if stdout_is_terminal; then
    echo "lab admin password: $1"
    echo "This password is printed once. make clean removes it with the lab volume."
  else
    store_admin_password_file "$1"
    echo "lab admin password: not printed, because stdout is not a terminal"
    echo "lab admin password file: $(admin_password_file)"
    echo "The file is readable only by you. make clean removes it."
  fi
}

# remove_admin_password_file  — make clean. Removes the file, and the directory
# when nothing else is in it.
remove_admin_password_file() {
  f=$(admin_password_file)
  rm -f "$f"
  rmdir "$(dirname "$f")" 2>/dev/null || true
}

# lab_require_loopback NAME VALUE: the page carries the lab's password, so an address set for it must be loopback.
# Anything else is refused by name, before a container starts. Exit 0 for 127.0.0.1, ::1, [::1] and localhost.
# The page itself is published on 127.0.0.1 in lab/compose.yaml, like every port of the lab, whatever LAB_PAGE_BIND
# says: a LAB_PAGE_BIND set to another address is refused by name, so nobody believes the page moved.
lab_require_loopback() {
  case $2 in
    127.0.0.1 | ::1 | '[::1]' | localhost) return 0 ;;
  esac
  echo "refused: $1=$2 is not a loopback address. The start page carries the lab's password and is published on 127.0.0.1 only. NEXT — unset $1, or set it to 127.0.0.1." >&2
  return 1
}

# load_lab_env: read lab/.env (when it exists) into the environment. A line is KEY=value with KEY made of letters,
# digits and underscores; anything else is refused by line number, never by content, and an EMPTY value is left
# unset. The file is read, never executed. A NON-EMPTY value already in the environment wins (a setting that
# lab/versions.env defaults to empty, such as ZETARIS_COMPOSE_FILE, is filled from here).
load_lab_env() {
  read_lab_env_file || return 1
  map_guide_names
}

# map_guide_names: the organisers' Participant Setup Guide names two settings that Proven spells another way. A judge's
# .env from the guide holds METERLESS_BASE_URL and METERLESS_MODEL; the lab reads them as PROVEN_GENERATOR_BASE_URL and
# PROVEN_GENERATOR_MODEL, unless the Proven name is set too (it wins). METERLESS_RUNTIME_TOKEN and the ZETARIS_* names
# are the same in both and need no mapping.
map_guide_names() {
  if [ -z "${PROVEN_GENERATOR_BASE_URL:-}" ] && [ -n "${METERLESS_BASE_URL:-}" ]; then
    PROVEN_GENERATOR_BASE_URL=$METERLESS_BASE_URL
    export PROVEN_GENERATOR_BASE_URL
  fi
  if [ -z "${PROVEN_GENERATOR_MODEL:-}" ] && [ -n "${METERLESS_MODEL:-}" ]; then
    PROVEN_GENERATOR_MODEL=$METERLESS_MODEL
    export PROVEN_GENERATOR_MODEL
  fi
}

read_lab_env_file() {
  envfile=$ROOT/lab/.env
  [ -f "$envfile" ] || return 0
  n=0
  while IFS= read -r line || [ -n "$line" ]; do
    n=$((n + 1))
    case $line in
      '' | \#*) continue ;;
    esac
    key=${line%%=*}
    val=${line#*=}
    case $key in
      '' | *[!A-Za-z0-9_]*)
        echo "lab/.env: refused line $n (a setting is NAME=value)" >&2
        return 1
        ;;
    esac
    # A setting of the lab, not a way to change what runs: these names would change the shell, the loader or docker.
    case $key in
      PATH | IFS | HOME | SHELL | ENV | BASH_ENV | CDPATH | ROOT | LD_* | DYLD_* | DOCKER_* | COMPOSE_*)
        echo "lab/.env: refused line $n (that name is not a lab setting)" >&2
        return 1
        ;;
    esac
    [ -n "$val" ] || continue
    eval "isset=\${$key:+x}"
    if [ -z "$isset" ]; then
      export "$key=$val"
    fi
  done <"$envfile"
}

compose() {
  case ${1-} in
    up | run | create | start) refuse_retired_bind ;;
  esac
  docker compose \
    --env-file "$ROOT/lab/versions.env" \
    -f "$ROOT/lab/compose.yaml" \
    --project-directory "$ROOT/lab" \
    "$@"
}

die() {
  echo "$1" >&2
  exit 1
}

need_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    die "docker: not found"
  fi
}

# go_at_least NEED GOT  — GOT is 1.26.8, NEED is 1.26. Exit 0 when GOT >= NEED.
go_at_least() {
  awk -v need="$1" -v got="$2" 'BEGIN {
    split(need, a, ".")
    split(got, b, ".")
    for (i = 1; i <= 3; i++) {
      av = a[i] + 0
      bv = b[i] + 0
      if (bv > av) exit 0
      if (bv < av) exit 1
    }
    exit 0
  }'
}

host_goos() {
  os=$(uname -s)
  case $os in
    Linux) printf '%s\n' linux ;;
    Darwin) printf '%s\n' darwin ;;
    *) die "unsupported OS $os" ;;
  esac
}

host_goarch() {
  arch=$(uname -m)
  case $arch in
    x86_64 | amd64) printf '%s\n' amd64 ;;
    aarch64 | arm64) printf '%s\n' arm64 ;;
    *) die "unsupported architecture $arch" ;;
  esac
}

oci_created() {
  date -u +%Y-%m-%dT%H:%M:%SZ
}

oci_revision() {
  git -C "$ROOT" rev-parse HEAD
}

oci_version() {
  git -C "$ROOT" describe --tags --always --dirty | tr '/+' '--'
}

install_go_help() {
  need=$(awk '$1 == "go" { print $2; exit }' "$ROOT/go.mod")
  cat <<EOF
Go is not installed at version ${need} or later, and Docker is not available, so proven cannot be built.

macOS:
  brew install go
  The package must be Go ${need} or later. The upstream builds are at https://go.dev/dl/

Linux:
  Install Go ${need} or later from https://go.dev/dl/
  A distribution package is fine when its version is ${need} or later.

Then run: make build
EOF
}

# lab_write_secrets: the credentials named in LAB_SECRET_KEYS (lab/versions.env) reach the seed as ONE file of the
# lab's state volume, /state/secrets, written over standard input at mode 0600 (the way the admin password is). They are
# not an environment variable of any container, not in lab/compose.yaml, not an argument of any process: printf is a
# builtin, so no process lists a value. A name that is empty or unset is left out, so the seed removes that secret.
# The lines are NAME=value; a value holds no line break (flow_secret_ok, and the loader refuses one in lab/.env).
lab_write_secrets() {
  lws_dropped=
  for lws_name in $LAB_SECRET_KEYS; do
    eval "lws_val=\${$lws_name:-}"
    case $lws_val in
      *'
'*) lws_dropped="${lws_dropped:+$lws_dropped, }$lws_name" ;;
    esac
  done
  if [ -n "$lws_dropped" ]; then
    echo "up: refused: $lws_dropped holds a line break. NEXT — put the value on one line in lab/.env." >&2
    return 1
  fi
  {
    for lws_name in $LAB_SECRET_KEYS; do
      eval "lws_val=\${$lws_name:-}"
      if [ -n "$lws_val" ]; then
        printf '%s=%s\n' "$lws_name" "$lws_val"
      fi
    done
  } | compose run --rm --no-deps -T --entrypoint sh seed -c 'umask 077 && cat > /state/secrets'
}

# lab_hand_secrets: lab_write_secrets, and when it fails the transfer file is removed before the failure is returned:
# a writer that stopped after it opened /state/secrets would otherwise leave every credential on the volume.
lab_hand_secrets() {
  if ! lab_write_secrets; then
    lab_remove_secrets
    return 1
  fi
}

# lab_remove_secrets: the transfer file is removed after the seed, whether it finished or failed. The seed removes it
# too, on every exit path of its own; this is the second hand, for a seed that was killed.
lab_remove_secrets() {
  compose run --rm --no-deps -T --entrypoint rm seed -f /state/secrets </dev/null >/dev/null 2>&1 || true
}

service_running() {
  compose ps --status running --services 2>/dev/null | grep -x "$1" >/dev/null 2>&1
}

load_versions
