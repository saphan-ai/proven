#!/bin/sh
# Sample review from a directory outside the repository. No keys and no Docker.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
sh "$ROOT/lab/scripts/build.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
(
  cd "$tmp"
  env -u METERLESS_RUNTIME_TOKEN \
    -u PROVEN_GUARDRAIL_KEY \
    -u ZETARIS_API_KEY \
    -u ZETARIS_REST_URL \
    -u ZETARIS_ORG_ID \
    -u DOCKER_HOST \
    "$ROOT/bin/proven" scan --sample "$ROOT/testdata/fixture/discount"
)
