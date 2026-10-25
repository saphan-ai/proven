#!/bin/sh
# Stop the lab and remove its containers. Volumes stay until make clean.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
need_docker
compose down
