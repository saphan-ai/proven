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
# is normalised before the outputs are compared, and only on the last line.
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
# normalise FIXTURE: the run's own fixture path becomes <fixture> (a fixed
# string, never a pattern), and the LAST line, when it is the trace line, has
# its path replaced. No other line is changed, so no other difference can hide.
normalise() {
  awk -v fx="$1" '
    { line = $0
      out = ""
      while ((i = index(line, fx)) > 0) { out = out substr(line, 1, i - 1) "<fixture>"; line = substr(line, i + length(fx)) }
      held[NR] = out line }
    END { if (NR > 0 && held[NR] ~ /^trace: /) held[NR] = "trace: <trace path>"
          for (n = 1; n <= NR; n++) print held[n] }'
}

fail() {
  echo "image-check: FAIL — $1" >&2
  shift
  for f in "$@"; do echo "--- $f" >&2; cat "$f" >&2; done
  exit 1
}

# Both version commands must SUCCEED, for the same reason as the reviews below.
"$binary" --version > "$work/bin/version" 2>&1 \
  || fail "the binary's --version exited $?" "$work/bin/version"
docker run --rm --pull=never "$image" --version > "$work/img/version" 2>&1 \
  || fail "the image's --version exited $?" "$work/img/version"
cmp -s "$work/bin/version" "$work/img/version" \
  || fail "the version lines differ" "$work/bin/version" "$work/img/version"

# Both runs must SUCCEED: two runs that fail the same way print the same output,
# and comparing them alone would call a broken image good.
(cd "$work/bin" && "$binary" scan --sample "$fixture" > out 2>&1) \
  || fail "the binary's sample review exited $?" "$work/bin/out"
docker run --rm --pull=never "$image" scan --sample /opt/proven/fixture/discount > "$work/img/out" 2>&1 \
  || fail "the image's sample review exited $?" "$work/img/out"
normalise "$fixture" < "$work/bin/out" > "$work/bin/out.n"
normalise /opt/proven/fixture/discount < "$work/img/out" > "$work/img/out.n"
cmp -s "$work/bin/out.n" "$work/img/out.n" \
  || fail "the sample reviews differ (trace path and fixture path normalised)" "$work/bin/out.n" "$work/img/out.n"

# The version line does not name a build yet, so the commit is checked from the
# image's revision label (make image sets it from git) against this checkout.
label=$(docker image inspect "$image" --format '{{index .Config.Labels "org.opencontainers.image.revision"}}')
if head=$(git -C "$root" rev-parse HEAD 2>/dev/null); then
  [ "$label" = "$head" ] \
    || fail "the image was built from $label, this checkout is $head"
  echo "image-check: revision $head (the image's label and this checkout agree)"
else
  echo "remark: this directory is not a git checkout, so the image's revision ($label) is not compared"
fi

echo "image-check: ok — $image and $binary print the same version and the same sample review"
echo "image-check: version $(head -n 1 "$work/img/version")"
echo "image-check: size $(docker image inspect "$image" --format '{{.Size}}') bytes"
