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

service_running() {
  compose ps --status running --services 2>/dev/null | grep -x "$1" >/dev/null 2>&1
}

load_versions
