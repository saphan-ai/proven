#!/bin/sh
# Follow the lab logs. SERVICE= names one service.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker
if [ -n "${SERVICE:-}" ]; then
  compose logs -f "$SERVICE"
else
  compose logs -f
fi
