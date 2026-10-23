#!/bin/sh
# Scout first. Refuse on a critical or high finding with a fix. Then push.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

sh "$ROOT/lab/scripts/scan.sh"

if ! docker buildx inspect proven-builder >/dev/null 2>&1; then
  docker buildx create --name proven-builder --driver docker-container
fi

created=$(oci_created)
revision=$(oci_revision)
version=$(oci_version)
name="$PROVEN_REGISTRY/$PROVEN_IMAGE"

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
