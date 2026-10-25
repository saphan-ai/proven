#!/bin/sh
# Setup script: creates a user, then adds the user to a team.
# Arguments: USER EMAIL. The API client is the program named by API_CMD.
# Prints "ok" and exits 0, or prints the failing step and exits 1.
# The steps reached are traced on stderr.
set -eu

user=${1:?usage: setup.sh USER EMAIL}
email=${2:?usage: setup.sh USER EMAIL}
api=${API_CMD:?API_CMD is not set}

echo "step create" >&2
status=$(sh "$api" create "$user" "$email" | cut -d ' ' -f 1)
# 422: the user already exists, so go on
if [ "$status" != 201 ] && [ "$status" != 422 ]; then
  echo create
  exit 1
fi

echo "step team" >&2
status=$(sh "$api" team "$user" | cut -d ' ' -f 1)
if [ "$status" != 204 ]; then
  echo team
  exit 1
fi

echo ok
