#!/bin/sh
# Build ./bin/proven and ./bin/livecheck with Go, or with Docker, or say how to install Go.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

need=$(awk '$1 == "go" { print $2; exit }' "$ROOT/go.mod")
mkdir -p "$ROOT/bin"
cd "$ROOT"

use_go=0
if command -v go >/dev/null 2>&1; then
  got=$(go env GOVERSION)
  got=${got#go}
  if go_at_least "$need" "$got"; then
    use_go=1
  fi
fi

if [ "$use_go" -eq 1 ]; then
  CGO_ENABLED=0 go build -o "$ROOT/bin/proven" ./cmd/proven
  CGO_ENABLED=0 go build -o "$ROOT/bin/livecheck" ./cmd/livecheck
  exit 0
fi

if command -v docker >/dev/null 2>&1; then
  docker run --rm \
    --user "$(id -u):$(id -g)" \
    -e CGO_ENABLED=0 \
    -e GOOS="$(host_goos)" \
    -e GOARCH="$(host_goarch)" \
    -e GOCACHE=/tmp/gocache \
    -e GOPATH=/tmp/gopath \
    -e HOME=/tmp \
    -v "$ROOT:/src:ro" \
    -v "$ROOT/bin:/out" \
    -w /src \
    "$GOLANG_IMAGE" \
    sh -c 'go build -o /out/proven ./cmd/proven && go build -o /out/livecheck ./cmd/livecheck'
  exit 0
fi

install_go_help >&2
exit 1
