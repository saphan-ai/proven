#!/bin/sh
# Print the lab's login: the same lines make up ends with. On a terminal the password is shown; when stdout is
# redirected the line names the file that holds it and the password is not written.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
. "$ROOT/lab/scripts/banner.sh"
need_docker

pass=$(lab_read_password)
if [ -z "$pass" ]; then
  die "password: the lab is not installed (no password file in the lab's volume); run make up"
fi
lab_banner "$pass"
