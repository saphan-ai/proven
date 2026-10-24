#!/bin/sh
# Build the images we publish for this machine and scan each with Docker Scout.
# Fails closed: it exits non-zero, naming the image, when Scout is missing, when
# a scan cannot run, or when a finding of SCOUT_REFUSE_SEVERITIES has a fix.
# make scan and make push both run it; make push does so before it publishes.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

proven_image="$PROVEN_IMAGE:latest"
seed_image="$SEED_IMAGE:$SEED_IMAGE_TAG"

if ! docker scout version >/dev/null 2>&1; then
  echo "scout: docker scout is not installed, so $proven_image and $seed_image are not scanned. Install the Docker Scout plugin; nothing is published without a scan." >&2
  exit 1
fi

sh "$ROOT/lab/scripts/image.sh"
sh "$ROOT/lab/scripts/seed-image.sh"

refuse=0
for image in "$proven_image" "$seed_image"; do
  echo "scout: $image"
  if ! docker scout cves "$image" \
    --only-severity "$SCOUT_REFUSE_SEVERITIES" \
    --only-fixed \
    --exit-code; then
    echo "scout: refusing $image: the scan failed or a $SCOUT_REFUSE_SEVERITIES finding has a fix" >&2
    refuse=1
  fi
done

if [ "$refuse" -ne 0 ]; then
  exit 1
fi
