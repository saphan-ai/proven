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

# require_loopback_bind  — the lab publishes its port on LAB_BIND only when that
# is a loopback address: 127.0.0.0/8 or [::1]. An empty value binds every
# interface, so it is refused too. LAB_BIND may come from the environment.
require_loopback_bind() {
  if ! printf '%s\n' "${LAB_BIND-}" | awk '
    $0 == "[::1]" { exit 0 }
    {
      n = split($0, o, ".")
      if (n != 4 || o[1] != "127") exit 1
      for (i = 2; i <= 4; i++) {
        if (o[i] !~ /^[0-9]+$/ || length(o[i]) > 3 || o[i] + 0 > 255) exit 1
      }
      exit 0
    }'; then
    echo "LAB_BIND=${LAB_BIND-} is not a loopback address. The lab binds to 127.0.0.0/8 or [::1] only. Unset LAB_BIND or set it to 127.0.0.1." >&2
    exit 1
  fi
}

compose() {
  case ${1-} in
    up | run | create | start) require_loopback_bind ;;
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
