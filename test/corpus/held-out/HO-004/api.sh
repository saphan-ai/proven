#!/bin/sh
# Stand-in for the service's web API. No network: it keeps its users in a file.
#   api.sh create USER EMAIL   prints "<status> <reason>"
#   api.sh team USER           prints "<status> <reason>"
# Create answers 201 for a new user, 422 "user exists" for a user that exists,
# and 422 "email refused" for an email address it will not take (any address
# at the domain blocked.example). Team answers 204 when the user exists and 404
# when it does not. The state directory is named by API_STATE.
set -eu

state=${API_STATE:?API_STATE is not set}
users="$state/users.txt"
[ -f "$users" ] || : > "$users"

case "${1:-}" in
  create)
    user=${2:?usage: api.sh create USER EMAIL}
    email=${3:?usage: api.sh create USER EMAIL}
    if grep -qx "$user" "$users"; then
      echo "422 user exists"
    else
      case "$email" in
        *@blocked.example) echo "422 email refused" ;;
        *) echo "$user" >> "$users"; echo "201 created" ;;
      esac
    fi
    ;;
  team)
    user=${2:?usage: api.sh team USER}
    if grep -qx "$user" "$users"; then
      echo "204 added"
    else
      echo "404 no such user"
    fi
    ;;
  *)
    echo "usage: api.sh create|team ..." >&2
    exit 2
    ;;
esac
