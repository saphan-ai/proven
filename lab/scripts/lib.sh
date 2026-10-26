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

# lab_require_loopback NAME VALUE: the page carries the lab's password, so its address must be loopback. Anything
# else is refused by name, before a container starts. Exit 0 for 127.0.0.1, ::1, [::1] and localhost.
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

service_running() {
  compose ps --status running --services 2>/dev/null | grep -x "$1" >/dev/null 2>&1
}

load_versions
