#!/bin/sh
# Print the version, the tools, the lab address, and every target's comment.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

printf '%s\n' 'proven'
printf 'version: %s\n' "$(git -C "$ROOT" describe --tags --always --dirty)"
if command -v go >/dev/null 2>&1; then
  printf 'go: %s\n' "$(go version)"
else
  printf '%s\n' 'go: not found'
fi
if command -v docker >/dev/null 2>&1; then
  ver=$(docker version --format '{{.Client.Version}}' 2>/dev/null || true)
  if [ -n "$ver" ]; then
    printf 'docker: %s\n' "$ver"
  else
    printf '%s\n' 'docker: found'
  fi
else
  printf '%s\n' 'docker: not found'
fi
if service_running gitea; then
  printf 'lab: http://%s:%s/\n' "$LAB_BIND" "$GITEA_HTTP_PORT"
  if service_running lgtm; then
    printf 'dashboard: http://%s:%s/d/%s\n' "$LAB_BIND" "$LGTM_GRAFANA_PORT" "$LGTM_DASHBOARD_UID"
  fi
else
  printf '%s\n' 'lab: not running'
fi
printf '\n'
awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z0-9_-]+:.*## / { printf "  %-14s %s\n", $1, $2 }' "$ROOT/Makefile"
