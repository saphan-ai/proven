#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
mkdir -p "$root/dist"
for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do
  target_os=${target%/*}
  target_arch=${target#*/}
  echo "Building $target"
  (cd "$root" && CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build -trimpath -o "dist/proven-$target_os-$target_arch" ./cmd/proven)
done
(cd "$root/dist" && if command -v sha256sum >/dev/null 2>&1; then sha256sum proven-* > checksums.txt; else shasum -a 256 proven-* > checksums.txt; fi)
