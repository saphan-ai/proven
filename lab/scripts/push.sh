#!/bin/sh
# Push every image we build. Each tag is one manifest list of PUSH_PLATFORMS.
# Nothing is published until Docker Scout has scanned every image: scout-images.sh
# runs first and a refusal stops this script before any builder or push.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

sh "$ROOT/lab/scripts/scout-images.sh" || die "push: refusing to publish: the Docker Scout scan did not pass (see the image named above)"

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
seed="$PROVEN_REGISTRY/$SEED_IMAGE:$SEED_IMAGE_TAG"

echo "push: $name:latest and $name:$version, one tag each, platforms $PUSH_PLATFORMS."
echo "push: $seed, one tag, platforms $PUSH_PLATFORMS."
echo "push: Docker Scout passed for the local build of each image."

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
  --tag "$name:latest" \
  --tag "$name:$version" \
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
  --tag "$seed" \
  "$ROOT/lab/seed"
