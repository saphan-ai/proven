# Going live: the keys of lab/.env as Actions secrets of the demo repository. Sourced by seed.sh, and by
# lab/lab_test.go, which runs it against a stub of Gitea's API.
#
# It reads: LAB_LIVE_KEYS, the settings it names, API_TOKEN, GITEA_INTERNAL_URL, LAB_ADMIN_USER, LAB_DEMO_REPO.

# Going live (ADR-0052). Each setting named in LAB_LIVE_KEYS that is set and not empty becomes an Actions secret of
# the demo repository; one that is empty is removed from it, so taking a key out of lab/.env returns the lab to
# sample mode. The value reaches this container by name from the environment of `make up`. It goes to Gitea over a
# pipe (printf is a shell builtin, so no process lists it), is never written to a file, and is never printed:
# the log carries NAMES and counts only. Sets LIVE_STORED to the names stored, one line.
ensure_live_secrets() {
  LIVE_STORED=""
  for name in ${LAB_LIVE_KEYS:-}; do
    case $name in
      *[!A-Z0-9_]* | '')
        echo "seed: LAB_LIVE_KEYS holds a name that is not an Actions secret name" >&2
        exit 1
        ;;
    esac
    eval "value=\${$name:-}"
    url="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/actions/secrets/$name"
    if [ -z "$value" ]; then
      # 404: there was none. 204: it was removed.
      code=$(curl -sS --max-time 30 -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: token $API_TOKEN" "$url")
      if [ "$code" != "204" ] && [ "$code" != "404" ]; then
        echo "seed: could not remove the empty secret $name (HTTP $code)" >&2
        exit 1
      fi
      continue
    fi
    esc=$(printf '%s' "$value" | sed 's/\\/\\\\/g; s/"/\\"/g')
    code=$(printf '{"data":"%s"}' "$esc" | curl -sS --max-time 30 -o /dev/null -w '%{http_code}' -X PUT \
      -H "Authorization: token $API_TOKEN" -H 'Content-Type: application/json' --data-binary @- "$url")
    if [ "$code" != "201" ] && [ "$code" != "204" ] && [ "$code" != "200" ]; then
      echo "seed: could not store the secret $name (HTTP $code)" >&2
      exit 1
    fi
    LIVE_STORED="${LIVE_STORED:+$LIVE_STORED, }$name"
  done
  if [ -n "$LIVE_STORED" ]; then
    echo "seed: stored as Actions secrets: $LIVE_STORED"
    LAB_LIVE_STATE="keys stored: $LIVE_STORED"
  else
    echo "seed: no key is set; the lab reviews in sample mode"
    LAB_LIVE_STATE="no key is stored; reviews are sample reviews"
  fi
  export LAB_LIVE_STATE
}
