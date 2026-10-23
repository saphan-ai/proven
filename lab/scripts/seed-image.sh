#!/bin/sh
# Build the one-shot seed image. It is pinned by SEED_IMAGE_TAG, not latest.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

if ! docker buildx version >/dev/null 2>&1; then
  die "docker buildx is not available"
fi

created=$(oci_created)
revision=$(oci_revision)
version=$(oci_version)
platform=$(docker version --format '{{.Server.Os}}/{{.Server.Arch}}')
echo "seed-image: one platform ($platform). A local engine cannot load a manifest list."
echo "seed-image: make push publishes $PUSH_PLATFORMS as one tag."

docker buildx build \
  --file "$ROOT/lab/seed/Dockerfile" \
  --load \
  --platform "$platform" \
  --build-arg "RUNTIME_IMAGE=$RUNTIME_IMAGE" \
  --label "org.opencontainers.image.created=$created" \
  --label "org.opencontainers.image.revision=$revision" \
  --label "org.opencontainers.image.version=$version" \
  --label "org.opencontainers.image.title=proven-lab-seed" \
  --tag "$SEED_IMAGE:$SEED_IMAGE_TAG" \
  "$ROOT/lab/seed"
