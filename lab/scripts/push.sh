#!/bin/sh
# Publish the images we build: nothing reaches a registry before every platform of
# every image has been scanned, and what is scanned is what is published.
#   1. Each image is built once for every platform in PUSH_PLATFORMS into a local
#      OCI layout (`buildx build --output type=oci`). No build pushes anything.
#   2. Docker Scout scans, from that layout, the image of every platform.
#   3. Only when every scan of every image is green, skopeo copies each layout,
#      byte for byte and with its digests kept, to its final tags, and each tag
#      is read back and must carry the digest of the layout that was scanned.
# A missing Scout plugin, a missing skopeo, a scan that cannot run, a fixable
# SCOUT_REFUSE_SEVERITIES finding or a missing platform on any image publishes
# nothing at all: the first registry write comes after the last green scan.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

if ! docker scout version >/dev/null 2>&1; then
  die "push: refusing to publish: the docker scout plugin is not installed, so nothing can be scanned. Install it; nothing is published without a scan."
fi
if ! command -v skopeo >/dev/null 2>&1; then
  die "push: refusing to publish: skopeo is not installed, and it is what copies a scanned layout to the registry. Install it; nothing is published without a scan."
fi
if ! docker buildx version >/dev/null 2>&1; then
  die "docker buildx is not available"
fi

if ! docker buildx inspect proven-builder >/dev/null 2>&1; then
  docker buildx create --name proven-builder --driver docker-container
fi

created=$(oci_created)
revision=$(oci_revision)
version=$(oci_version)
name="$PROVEN_REGISTRY/$PROVEN_IMAGE"
seed_name="$PROVEN_REGISTRY/$SEED_IMAGE"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "push: $name:latest and $name:$version, one tag each, platforms $PUSH_PLATFORMS."
echo "push: $seed_name:$SEED_IMAGE_TAG, one tag, platforms $PUSH_PLATFORMS."
echo "push: building once to local layouts and scanning them; the registry is written only after every scan is green."

# unpack NAME  — the layout NAME.tar that buildx wrote, as a directory Scout reads.
unpack() {
  mkdir "$work/$1"
  tar -xf "$work/$1.tar" -C "$work/$1"
  rm -f "$work/$1.tar"
}

# The one build of each image. It writes a local layout and pushes nothing.
docker buildx build \
  --builder proven-builder \
  --file "$ROOT/Dockerfile" \
  --platform "$PUSH_PLATFORMS" \
  --provenance=mode=max \
  --sbom=true \
  --output "type=oci,dest=$work/proven.tar,name=$name:scan" \
  --build-arg "GOLANG_IMAGE=$GOLANG_IMAGE" \
  --build-arg "RUNTIME_IMAGE=$RUNTIME_IMAGE" \
  --build-arg "CREATED=$created" \
  --build-arg "REVISION=$revision" \
  --build-arg "VERSION=$version" \
  --label "org.opencontainers.image.created=$created" \
  --label "org.opencontainers.image.revision=$revision" \
  --label "org.opencontainers.image.version=$version" \
  --label "org.opencontainers.image.title=proven" \
  --label "org.opencontainers.image.source=https://github.com/saphan-ai/proven" \
  "$ROOT"
unpack proven

docker buildx build \
  --builder proven-builder \
  --file "$ROOT/lab/seed/Dockerfile" \
  --platform "$PUSH_PLATFORMS" \
  --provenance=mode=max \
  --sbom=true \
  --output "type=oci,dest=$work/seed.tar,name=$seed_name:scan" \
  --build-arg "RUNTIME_IMAGE=$RUNTIME_IMAGE" \
  --label "org.opencontainers.image.created=$created" \
  --label "org.opencontainers.image.revision=$revision" \
  --label "org.opencontainers.image.version=$version" \
  --label "org.opencontainers.image.title=proven-lab-seed" \
  "$ROOT/lab/seed"
unpack seed

# layout_digest NAME  — the digest of the index the layout holds: the digest every
# registry read of the published image must give back.
layout_digest() {
  tr -d ' \n\t' <"$work/$1/index.json" | grep -o '"digest":"sha256:[0-9a-f]*"' | head -n 1 | cut -d '"' -f 4
}

# scan_layout NAME  — scans, from the layout, the image of every platform in
# PUSH_PLATFORMS. Returns non-zero when any scan did not pass or a platform is
# not in the layout.
scan_layout() {
  ok=0
  if [ -z "$(layout_digest "$1")" ]; then
    echo "scout: refusing $1: its layout has no index digest" >&2
    return 1
  fi
  for platform in $(printf '%s' "$PUSH_PLATFORMS" | tr ',' ' '); do
    echo "scout: $1 $platform $(layout_digest "$1")"
    if ! docker scout cves "oci-dir://$work/$1" \
      --platform "$platform" \
      --only-severity "$SCOUT_REFUSE_SEVERITIES" \
      --exit-code; then
      echo "scout: refusing $1 $platform: the scan failed, the platform is missing, or a $SCOUT_REFUSE_SEVERITIES finding was reported" >&2
      ok=1
    fi
  done
  return "$ok"
}

refused=0
scan_layout proven || refused=1
scan_layout seed || refused=1

[ "$refused" -eq 0 ] || die "push: refusing to publish: a Docker Scout scan did not pass (see the image and platform named above). Nothing was pushed to any registry."

# publish NAME TAG...  — copies the scanned layout to each tag and reads each tag
# back: it must carry the digest of the layout.
publish() {
  layout=$1
  shift
  digest=$(layout_digest "$layout")
  for t in "$@"; do
    skopeo copy --all --preserve-digests "oci:$work/$layout" "docker://$t"
    got=$(docker buildx imagetools inspect "$t" --format '{{.Manifest.Digest}}')
    if [ "$got" != "$digest" ]; then
      die "push: $t carries $got, not the scanned $digest"
    fi
  done
}

# :judging is the tag every image a judge's run pulls is referenced by during the judging
# window (docs/JUDGING.md). It moves only when the owner asks for it (make push-judging), so
# an ordinary push to latest never changes what judges run.
judging_tag() {
  if [ "${PUSH_JUDGING:-}" = yes ]; then
    printf '%s:judging' "$1"
  fi
}

proven_digest=$(layout_digest proven)
seed_digest=$(layout_digest seed)
publish proven "$name:latest" "$name:$version" $(judging_tag "$name")
publish seed "$seed_name:$SEED_IMAGE_TAG" $(judging_tag "$seed_name")
echo "push: published the scanned digests $proven_digest and $seed_digest."
