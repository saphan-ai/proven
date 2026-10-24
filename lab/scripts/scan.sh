#!/bin/sh
# Scan the images we build with Docker Scout (scout-images.sh), then record the
# third-party images in docs/LAB.md. Fails closed: Docker Scout missing, a scan
# that cannot run, or a critical or high finding with a fix refuses the images
# we build, and this script exits non-zero naming the image. make push runs
# scout-images.sh before it publishes.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker

# With Scout missing there is nothing to record for the third-party images.
sh "$ROOT/lab/scripts/scout-images.sh" || exit 1

doc=$ROOT/docs/LAB.md
raw=$(mktemp)
trap 'rm -f "$raw"' EXIT
{
  echo "Scanned $(date -u +%Y-%m-%dT%H:%M:%SZ). Severities: $SCOUT_REFUSE_SEVERITIES, fix available only."
  echo
  for image in \
    "gitea/gitea:$GITEA_VERSION" \
    "gitea/act_runner:$ACT_RUNNER_VERSION" \
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
