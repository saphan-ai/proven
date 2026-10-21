#!/bin/bash
# zetaris_console.sh — open the console of a self-hosted Zetaris platform on this machine as http://localhost:3000.
#
# Why a tunnel: the Zetaris web console calls crypto.randomUUID(), which a browser offers only in a "secure context"
# (https, or http://localhost). Opened as http://<address>:3000 it fails at sign-in. An ssh tunnel to localhost works.
#
# Environment, not product code (fleet tooling of proven, 2026-10-21).
#
# Settings (environment variables, or the file $HOME/.config/proven/zetaris-console.env):
#   ZETARIS_SSH      the ssh target of the machine that runs the platform, e.g. user@host
#   ZETARIS_UI       the console's address as that machine sees it, e.g. 127.0.0.1:3000 or <private ip>:3000
#   ZETARIS_LOCAL    the local port to open (default 3000)
#
# Usage: zetaris_console.sh up | down | status
set -u
CONF="${ZETARIS_CONSOLE_CONF:-$HOME/.config/proven/zetaris-console.env}"
[ -f "$CONF" ] && . "$CONF"
LOCAL="${ZETARIS_LOCAL:-3000}"
SOCK="${TMPDIR:-/tmp}/zetaris-console-$LOCAL.sock"

need() { [ -n "${ZETARIS_SSH:-}" ] && [ -n "${ZETARIS_UI:-}" ] || { echo "set ZETARIS_SSH and ZETARIS_UI (or write them to $CONF)"; exit 2; }; }
up_now() { ssh -S "$SOCK" -O check "${ZETARIS_SSH:-x}" >/dev/null 2>&1; }
answers() { curl -s -o /dev/null -m 5 -w '%{http_code}' "http://localhost:$LOCAL/api/config"; }

case "${1:-}" in
  up)
    need
    if up_now; then echo "already up: http://localhost:$LOCAL"; exit 0; fi
    if lsof -nP -iTCP:"$LOCAL" -sTCP:LISTEN >/dev/null 2>&1; then echo "port $LOCAL is taken on this machine:"; lsof -nP -iTCP:"$LOCAL" -sTCP:LISTEN; exit 2; fi
    # nohup: the tunnel survives closing the terminal; -f puts ssh in the background after it connects.
    nohup ssh -M -S "$SOCK" -f -N -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 \
      -L "127.0.0.1:$LOCAL:$ZETARIS_UI" "$ZETARIS_SSH" >/dev/null 2>&1 </dev/null || { echo "the tunnel did not open"; exit 1; }
    sleep 1
    code=$(answers); echo "up: http://localhost:$LOCAL  (console answers $code)"; [ "$code" = 200 ]
    ;;
  down)
    need
    if up_now; then ssh -S "$SOCK" -O exit "$ZETARIS_SSH" 2>/dev/null; echo "down"; else echo "not running"; fi
    ;;
  status)
    need
    if up_now; then echo "up: http://localhost:$LOCAL  (console answers $(answers))"; else echo "down"; exit 3; fi
    ;;
  *)
    echo "usage: $0 up | down | status"; exit 2
    ;;
esac
