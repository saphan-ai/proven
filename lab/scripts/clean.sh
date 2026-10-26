#!/bin/sh
# Remove the lab's containers, volumes and local images, and ./bin and ./dist.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

if command -v docker >/dev/null 2>&1; then
  compose down -v --rmi local || true
  version=$(oci_version)
  docker rmi "$PROVEN_IMAGE:latest" "$PROVEN_IMAGE:$version" "$SEED_IMAGE:$SEED_IMAGE_TAG" >/dev/null 2>&1 || true
else
  echo "docker: not found, no containers removed"
fi
remove_admin_password_file
rm -rf "$ROOT/bin" "$ROOT/dist"
