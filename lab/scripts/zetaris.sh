#!/bin/sh
# Start Zetaris on this machine for the lab (path C): `make up ZETARIS=local`, or `make zetaris`.
# up.sh calls this after the lab's network exists. Sourced settings: lab/versions.env and lab/.env.
#
# What this does NOT do, and why: it does not carry Zetaris's images. They are the sponsor's, in the sponsor's
# registry, behind a sign-in; the terms under which a judge may pull and run them are not in this repository
# (searched 2026-10-26: hackathon/fleet/infra/zetaris holds a runbook and a console tool, no licence or terms file).
# So the target uses YOUR bundle (the zip from your own Zetaris account) and YOUR registry sign-in, and says so.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker
load_lab_env

if [ "${ZETARIS:-}" != "local" ]; then
  echo "zetaris: ZETARIS is '${ZETARIS:-}', not 'local'; nothing to start. Run: make up ZETARIS=local"
  exit 0
fi

echo "zetaris: Zetaris's images come from the sponsor's registry with YOUR sign-in and YOUR bundle; this lab carries neither."

if [ -z "${ZETARIS_COMPOSE_FILE:-}" ] || [ ! -f "${ZETARIS_COMPOSE_FILE:-}" ]; then
  die "zetaris: ZETARIS_COMPOSE_FILE is '${ZETARIS_COMPOSE_FILE:-}', which is not a file. NEXT — unpack your zetaris-platform.zip, sign in to the sponsor's registry with the command that came with it, then set ZETARIS_COMPOSE_FILE=<the unpacked docker-compose.yml> in lab/.env (docs/LAB.md, path C)."
fi

for p in "$ZETARIS_UI_HOST_PORT" "$ZETARIS_API_HOST_PORT"; do
  case $p in
    '' | *[!0-9]*) die "zetaris: a Zetaris host port is '$p', not a number (ZETARIS_UI_HOST_PORT, ZETARIS_API_HOST_PORT)" ;;
  esac
done

zcompose() {
  docker compose -p "$ZETARIS_PROJECT" \
    -f "$ZETARIS_COMPOSE_FILE" -f "$ROOT/lab/zetaris/override.yml" \
    --project-directory "$(dirname "$ZETARIS_COMPOSE_FILE")" "$@"
}

if ! zcompose pull; then
  die "zetaris: pulling the platform's images failed. If the message above says unauthorized, sign in to the sponsor's registry with the command from your own Zetaris account (a shell that is not logged), then run make up ZETARIS=local again."
fi
zcompose up -d

# The REST service answers a request without a key with an error status; any HTTP status proves it is listening.
waited=0
up=0
while [ "$waited" -lt "${ZETARIS_WAIT_SECONDS:-180}" ]; do
  code=$(curl -sS -o /dev/null --max-time 3 -w '%{http_code}' "http://127.0.0.1:${ZETARIS_API_HOST_PORT}${ZETARIS_REST_PATH}/" 2>/dev/null || true)
  case $code in
    '' | 000) ;;
    *)
      up=1
      break
      ;;
  esac
  waited=$((waited + 5))
  sleep 5
done
if [ "$up" -ne 1 ]; then
  die "zetaris: the REST service did not answer on 127.0.0.1:${ZETARIS_API_HOST_PORT} within ${ZETARIS_WAIT_SECONDS:-180}s. NEXT — run the bundle's own verify.sh (zetaris-local.md step 8)."
fi
echo "zetaris: the platform answers on http://127.0.0.1:${ZETARIS_API_HOST_PORT} (console: http://127.0.0.1:${ZETARIS_UI_HOST_PORT}, open it as localhost)"

# Registering the evidence is the administrator's act. It needs the administrator's key, the organisation id and the
# name of Proven's non-admin user, all from lab/.env. Without them the platform is up and the step says what is missing.
if [ -n "${ZETARIS_ADMIN_API_KEY:-}" ] && [ -n "${ZETARIS_ORG_ID:-}" ] && [ -n "${ZETARIS_PROVEN_USER:-}" ]; then
  printf '%s' "$ZETARIS_ADMIN_API_KEY" | compose --profile zetaris run --rm -T zetaris-register
else
  echo "zetaris: the evidence definitions are NOT registered yet: ZETARIS_ADMIN_API_KEY, ZETARIS_ORG_ID and ZETARIS_PROVEN_USER must all be set in lab/.env (docs/LAB.md, path C, steps 6 to 8). Then run make zetaris."
fi
