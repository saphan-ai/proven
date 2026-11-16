#!/bin/sh
# Build locally, check both architectures, scan, then publish the same image.
set -eu
image=${1:?usage: publish-image.sh REGISTRY/IMAGE:REVISION-TAG}
case "$image" in */*:?*) ;; *) echo 'Use an explicit registry/namespace image tag' >&2; exit 2 ;; esac
case "$image" in *:latest) echo 'Use a unique revision tag rather than latest' >&2; exit 2 ;; esac
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
revision=$(git -C "$root" rev-parse HEAD)
if [ -n "$(git -C "$root" status --porcelain --untracked-files=all)" ]; then
  echo 'Commit the reviewed source before publishing an image' >&2
  exit 2
fi
created=$(date -u +%Y-%m-%dT%H:%M:%SZ)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
set -- docker buildx build
if [ -n "${BUILDX_BUILDER:-}" ]; then set -- "$@" --builder "$BUILDX_BUILDER"; fi
"$@" --pull --platform linux/amd64,linux/arm64 --provenance=mode=max --sbom=true \
  --build-arg "CREATED=$created" --build-arg "REVISION=$revision" \
  --build-arg "VERSION=${image##*:}" --load -t "$image" "$root"

for platform in linux/amd64 linux/arm64; do
  docker run --rm --platform "$platform" --network none "$image" --version
  docker run --rm --platform "$platform" --network none --entrypoint /bin/sh "$image" -ec \
    'git --version; mkdir /tmp/git-smoke; git init -q /tmp/git-smoke; git -C /tmp/git-smoke status --porcelain'
  docker run --rm --platform "$platform" --network none "$image" scan --sample /opt/proven/fixture/discount
  # A non-zero scanner result stops this script before push. No pipeline, base
  # exclusion, fix-availability filter or suppression may turn failure into success.
  report="$work/scout-${platform#*/}.sarif"
  docker scout cves --exit-code --ignore-suppressed=false --only-severity critical,high \
    --format sarif --output "$report" --platform "$platform" "local://$image"
  # Count every returned result, including an accepted suppression. A scanner
  # policy must not make a nonempty HIGH/CRITICAL report publishable.
  python3 - "$report" <<'PY'
import json,sys
with open(sys.argv[1]) as f: report=json.load(f)
runs=report.get('runs')
if report.get('version')!='2.1.0' or not isinstance(runs,list) or not runs:
    raise SystemExit('Scanner did not produce a valid SARIF run')
if any(not isinstance(run,dict) or not isinstance(run.get('results',[]),list) for run in runs):
    raise SystemExit('Scanner results are malformed')
if any(run.get('results',[]) for run in runs):
    raise SystemExit('HIGH/CRITICAL findings block publication, including suppressed findings')
print('HIGH/CRITICAL release gate: no findings')
PY
done

docker push "$image"
docker buildx imagetools inspect "$image"
printf 'Published %s from %s after both platform scans passed\n' "$image" "$revision"
