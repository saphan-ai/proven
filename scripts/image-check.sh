#!/bin/sh
# image-check.sh IMAGE BINARY — prove that Proven's image runs the binary it
# claims to carry: the same version line and the same sample review as a
# binary built from the same commit. Prints the image's size. Exits non-zero
# and names the first difference.
#
#   sh scripts/image-check.sh proven:local ./bin/proven
#
# Both runs start in an empty directory outside the checkout, because a scan
# writes its state and trace where it is started from. The one line that may
# differ only because the two runs write to different places, the trace path,
# is normalised before the outputs are compared.
set -eu

image=${1:?usage: image-check.sh IMAGE BINARY}
binary=${2:?usage: image-check.sh IMAGE BINARY}
case $binary in /*) ;; *) binary=$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary") ;; esac
root=$(cd "$(dirname "$0")/.." && pwd)
fixture="$root/testdata/fixture/discount"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/img"

# The trace path names where each run wrote its trace, which is not the same
# inside the container and on the host. Only a line that IS the trace line
# ("trace: <path>") is rewritten; any other line is compared as printed.
normalise() { sed -E 's#^trace: .*$#trace: <trace path>#'; }

fail() {
  echo "image-check: FAIL — $1" >&2
  shift
  for f in "$@"; do echo "--- $f" >&2; cat "$f" >&2; done
  exit 1
}

# Both version commands must SUCCEED, for the same reason as the reviews below.
"$binary" --version > "$work/bin/version" 2>&1 \
  || fail "the binary's --version exited $?" "$work/bin/version"
docker run --rm "$image" --version > "$work/img/version" 2>&1 \
  || fail "the image's --version exited $?" "$work/img/version"
cmp -s "$work/bin/version" "$work/img/version" \
  || fail "the version lines differ" "$work/bin/version" "$work/img/version"

# Both runs must SUCCEED: two runs that fail the same way print the same output,
# and comparing them alone would call a broken image good.
(cd "$work/bin" && "$binary" scan --sample "$fixture" > out 2>&1) \
  || fail "the binary's sample review exited $?" "$work/bin/out"
docker run --rm "$image" scan --sample /opt/proven/fixture/discount > "$work/img/out" 2>&1 \
  || fail "the image's sample review exited $?" "$work/img/out"
normalise < "$work/bin/out" | sed "s#$fixture#<fixture>#g" > "$work/bin/out.n"
normalise < "$work/img/out" | sed 's#/opt/proven/fixture/discount#<fixture>#g' > "$work/img/out.n"
cmp -s "$work/bin/out.n" "$work/img/out.n" \
  || fail "the sample reviews differ (trace path and fixture path normalised)" "$work/bin/out.n" "$work/img/out.n"

echo "image-check: ok — $image and $binary print the same version and the same sample review"
echo "image-check: version $(head -n 1 "$work/img/version")"
echo "image-check: size $(docker image inspect "$image" --format '{{.Size}}') bytes"
