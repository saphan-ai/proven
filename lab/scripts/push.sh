#!/bin/sh
# Publish the images we build: what is scanned is what is published.
#   1. Each image is built once for every platform in PUSH_PLATFORMS and pushed to
#      a staging tag (staging-<revision>). A final tag is never built.
#   2. Docker Scout scans, by digest, the image of every platform in that build.
#   3. Only when every scan of every image is green, the final tags are made from
#      the scanned index digests (`buildx imagetools create`), and each final tag
#      is read back and must carry the digest that was scanned.
# A missing Scout plugin, a scan that cannot run, a fixable SCOUT_REFUSE_SEVERITIES
# finding or a missing platform on any image publishes nothing: no final tag is made.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

if ! docker scout version >/dev/null 2>&1; then
  die "push: refusing to publish: the docker scout plugin is not installed, so nothing can be scanned. Install it; nothing is published without a scan."
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
proven_stage="$name:staging-$revision"
seed_stage="$seed_name:staging-$SEED_IMAGE_TAG-$revision"

echo "push: $name:latest and $name:$version, one tag each, platforms $PUSH_PLATFORMS."
echo "push: $seed_name:$SEED_IMAGE_TAG, one tag, platforms $PUSH_PLATFORMS."
echo "push: building once to $proven_stage and $seed_stage; the final tags are made only from scanned digests."

# The one build of each image. It pushes a staging tag, never a final tag.
docker buildx build \
  --builder proven-builder \
  --file "$ROOT/Dockerfile" \
  --platform "$PUSH_PLATFORMS" \
  --provenance=mode=max \
  --sbom=true \
  --push \
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
  --tag "$proven_stage" \
  "$ROOT"

docker buildx build \
  --builder proven-builder \
  --file "$ROOT/lab/seed/Dockerfile" \
  --platform "$PUSH_PLATFORMS" \
  --provenance=mode=max \
  --sbom=true \
  --push \
  --build-arg "RUNTIME_IMAGE=$RUNTIME_IMAGE" \
  --label "org.opencontainers.image.created=$created" \
  --label "org.opencontainers.image.revision=$revision" \
  --label "org.opencontainers.image.version=$version" \
  --label "org.opencontainers.image.title=proven-lab-seed" \
  --tag "$seed_stage" \
  "$ROOT/lab/seed"

# index_digest REF  — the digest of the manifest list REF names.
index_digest() {
  docker buildx imagetools inspect "$1" --format '{{.Manifest.Digest}}'
}

# platform_digests REF  — one line per image in the list: "os/arch digest".
platform_digests() {
  docker buildx imagetools inspect "$1" --format '{{range .Manifest.Manifests}}{{.Platform.OS}}/{{.Platform.Architecture}} {{.Digest}}{{"\n"}}{{end}}'
}

# scan_staged REF  — scans, by digest, the image of every platform in
# PUSH_PLATFORMS. Sets scanned_index to the index digest that was scanned. Returns
# non-zero when any scan did not pass or a platform is not in the list.
scan_staged() {
  stage=$1
  repo=${stage%:*}
  scanned_index=$(index_digest "$stage")
  if [ -z "$scanned_index" ]; then
    echo "scout: refusing $stage: its digest could not be read" >&2
    return 1
  fi
  list=$(platform_digests "$stage")
  ok=0
  for platform in $(printf '%s' "$PUSH_PLATFORMS" | tr ',' ' '); do
    d=$(printf '%s\n' "$list" | awk -v p="$platform" '$1 == p { print $2; exit }')
    if [ -z "$d" ]; then
      echo "scout: refusing $stage: it has no image for $platform" >&2
      ok=1
      continue
    fi
    echo "scout: $stage $platform $d"
    if ! docker scout cves "registry://$repo@$d" \
      --only-severity "$SCOUT_REFUSE_SEVERITIES" \
      --only-fixed \
      --exit-code; then
      echo "scout: refusing $stage $platform $d: the scan failed or a $SCOUT_REFUSE_SEVERITIES finding has a fix" >&2
      ok=1
    fi
  done
  return "$ok"
}

refused=0
scan_staged "$proven_stage" || refused=1
proven_index=$scanned_index
scan_staged "$seed_stage" || refused=1
seed_index=$scanned_index

[ "$refused" -eq 0 ] || die "push: refusing to publish: a Docker Scout scan did not pass (see the image and platform named above). No final tag was made. The staging tags $proven_stage and $seed_stage hold a build that was not accepted; they are not a release."

# publish REPO DIGEST TAG...  — the final tags are made from REPO@DIGEST, the
# index that was scanned, and each is read back to carry that digest.
publish() {
  repo=$1
  digest=$2
  shift 2
  tagargs=
  for t in "$@"; do
    tagargs="$tagargs --tag $t"
  done
  # shellcheck disable=SC2086
  docker buildx imagetools create --builder proven-builder $tagargs "$repo@$digest"
  for t in "$@"; do
    got=$(index_digest "$t")
    if [ "$got" != "$digest" ]; then
      die "push: $t carries $got, not the scanned $digest"
    fi
  done
}

publish "${proven_stage%:*}" "$proven_index" "$name:latest" "$name:$version"
publish "${seed_stage%:*}" "$seed_index" "$seed_name:$SEED_IMAGE_TAG"
echo "push: published the scanned digests $proven_index and $seed_index."
