#!/bin/sh
# Name which live services are ready. Reads lab/.env when that file exists.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
sh "$ROOT/lab/scripts/build.sh"

envfile=$ROOT/lab/.env
if [ -f "$envfile" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    case $line in
      '' | \#*) continue ;;
    esac
    key=${line%%=*}
    case $key in
      *[!A-Za-z0-9_]*)
        echo "lab/.env: refused line" >&2
        exit 1
        ;;
    esac
    val=${line#*=}
    if [ -n "$val" ]; then
      export "$key=$val"
    fi
  done <"$envfile"
fi

exec "$ROOT/bin/livecheck"
