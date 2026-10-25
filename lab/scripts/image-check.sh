#!/bin/sh
# Check the image `make image` built against ./bin/proven (`make image-check`).
# The image tag is resolved the way image.sh resolves it: PROVEN_IMAGE from the
# environment, else from lab/versions.env.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker
sh "$ROOT/scripts/image-check.sh" "$PROVEN_IMAGE:latest" "$ROOT/bin/proven"
