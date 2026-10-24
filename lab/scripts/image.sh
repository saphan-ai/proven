#!/bin/sh
# Build Proven's image for the local platform and load it into the local engine.
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
echo "image: one platform ($platform). A local engine cannot load a manifest list."

docker buildx build \
  --file "$ROOT/Dockerfile" \
  --load \
  --platform "$platform" \
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
  --tag "$PROVEN_IMAGE:latest" \
  --tag "$PROVEN_IMAGE:$version" \
  "$ROOT"
