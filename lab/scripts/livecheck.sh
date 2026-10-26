#!/bin/sh
# Name which live services are ready. Reads lab/.env when that file exists.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
sh "$ROOT/lab/scripts/build.sh"

load_lab_env || exit 1

# With ZETARIS=local and no address of your own, the check runs from this host, so it asks the loopback port the
# platform is published on (lab/zetaris/override.yml), not the name the lab's reviews use on the lab network.
if [ "${ZETARIS:-}" = "local" ] && [ -z "${ZETARIS_REST_URL:-}" ]; then
  ZETARIS_REST_URL="http://127.0.0.1:${ZETARIS_API_HOST_PORT}${ZETARIS_REST_PATH}"
  export ZETARIS_REST_URL
fi

exec "$ROOT/bin/livecheck"
