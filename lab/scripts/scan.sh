#!/bin/sh
# Scout the images we build when the plugin is installed, then record
# third-party images in docs/LAB.md. A critical or high finding with a fix
# available refuses the images we build. make push does not call this script.
# With no plugin, the scan is read on Docker Hub after make push.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

if ! docker scout version >/dev/null 2>&1; then
  echo "scan: docker scout is not installed. The scan is read on Docker Hub after make push."
  echo "scan: Proven is ${PROVEN_IMAGE}, tags latest and the git describe value."
  echo "scan: the seed is ${SEED_IMAGE}:${SEED_IMAGE_TAG}."
  exit 0
fi

sh "$ROOT/lab/scripts/image.sh"
sh "$ROOT/lab/scripts/seed-image.sh"

refuse=0
for image in "$PROVEN_IMAGE:latest" "$SEED_IMAGE:$SEED_IMAGE_TAG"; do
  echo "scan: $image"
  if ! docker scout cves "$image" \
    --only-severity "$SCOUT_REFUSE_SEVERITIES" \
    --only-fixed \
    --exit-code; then
    echo "scan: refusing $image: a $SCOUT_REFUSE_SEVERITIES finding has a fix" >&2
    refuse=1
  fi
done

doc=$ROOT/docs/LAB.md
raw=$(mktemp)
trap 'rm -f "$raw"' EXIT
{
  echo "Scanned $(date -u +%Y-%m-%dT%H:%M:%SZ). Severities: $SCOUT_REFUSE_SEVERITIES, fix available only."
  echo
  for image in \
    "gitea/gitea:$GITEA_VERSION" \
    "gitea/act_runner:$ACT_RUNNER_VERSION" \
    "$LGTM_IMAGE:$LGTM_VERSION" \
    "docker:$DOCKER_VERSION-dind" \
    "docker:$DOCKER_VERSION-cli"; do
    echo "### $image"
    echo
    echo '```text'
    if docker scout cves "$image" \
      --only-severity "$SCOUT_REFUSE_SEVERITIES" \
      --only-fixed >"$raw" 2>&1; then
      cat "$raw"
    else
      cat "$raw"
      echo "(scout exited non-zero; the lab still uses this pinned image)"
    fi
    echo '```'
    echo
  done
} >"$raw.section"

if ! grep -q '<!-- scout-third-party:start -->' "$doc"; then
  die "docs/LAB.md is missing the scout markers"
fi

awk -v section="$raw.section" '
  $0 == "<!-- scout-third-party:start -->" {
    print
    while ((getline line < section) > 0) print line
    skip = 1
    next
  }
  $0 == "<!-- scout-third-party:end -->" { skip = 0 }
  skip != 1 { print }
' "$doc" >"$doc.tmp"
mv "$doc.tmp" "$doc"
rm -f "$raw.section"

if [ "$refuse" -ne 0 ]; then
  exit 1
fi
