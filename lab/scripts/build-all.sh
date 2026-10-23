#!/bin/sh
# Cross-compile proven and livecheck for darwin and linux, amd64 and arm64.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

need=$(awk '$1 == "go" { print $2; exit }' "$ROOT/go.mod")
mkdir -p "$ROOT/dist"
cd "$ROOT"

build_matrix() {
  for os in darwin linux; do
    for arch in amd64 arm64; do
      CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
        go build -trimpath -ldflags '-s -w' -o "$ROOT/dist/proven_${os}_${arch}" ./cmd/proven
      CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
        go build -trimpath -ldflags '-s -w' -o "$ROOT/dist/livecheck_${os}_${arch}" ./cmd/livecheck
    done
  done
}

use_go=0
if command -v go >/dev/null 2>&1; then
  got=$(go env GOVERSION)
  got=${got#go}
  if go_at_least "$need" "$got"; then
    use_go=1
  fi
fi

if [ "$use_go" -eq 1 ]; then
  build_matrix
elif command -v docker >/dev/null 2>&1; then
  docker run --rm \
    --user "$(id -u):$(id -g)" \
    -e CGO_ENABLED=0 \
    -e GOCACHE=/tmp/gocache \
    -e GOPATH=/tmp/gopath \
    -e HOME=/tmp \
    -v "$ROOT:/src:ro" \
    -v "$ROOT/dist:/out" \
    -w /src \
    "$GOLANG_IMAGE" \
    sh -c 'set -eu
      for os in darwin linux; do
        for arch in amd64 arm64; do
          GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w" -o /out/proven_${os}_${arch} ./cmd/proven
          GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-s -w" -o /out/livecheck_${os}_${arch} ./cmd/livecheck
        done
      done'
else
  install_go_help >&2
  exit 1
fi

cd "$ROOT/dist"
files=$(find . -maxdepth 1 -type f ! -name SHA256SUMS | sed 's|^\./||' | sort)
if command -v sha256sum >/dev/null 2>&1; then
  # shellcheck disable=SC2086
  sha256sum $files >SHA256SUMS
else
  # shellcheck disable=SC2086
  shasum -a 256 $files >SHA256SUMS
fi
