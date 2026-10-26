# Going live: the keys of lab/.env as Actions secrets of the demo repository. Sourced by seed.sh, and by
# lab/lab_test.go, which runs it against a stub of Gitea's API.
#
# It reads: LAB_LIVE_KEYS, the settings it names, API_TOKEN, GITEA_INTERNAL_URL, LAB_ADMIN_USER, LAB_DEMO_REPO.

# Going live (ADR-0052). Each setting named in LAB_LIVE_KEYS that is set and not empty becomes an Actions secret of
# the demo repository; one that is empty is removed from it, so taking a key out of lab/.env returns the lab to
# sample mode. The value reaches this container by name from the environment of `make up`. It goes to Gitea over a
# pipe (printf is a shell builtin, so no process lists it), is never written to a file, and is never printed:
# the log carries NAMES and counts only. Sets LIVE_STORED to the names stored, one line.
# feature_word VALUE: how the start page names a feature's stored setting. Anything but on or off is not stored by the
# workflow's guard either, so it reads as Proven's own default.
feature_word() {
  case $1 in
    on | off) printf '%s' "$1" ;;
    *) printf "Proven's default" ;;
  esac
}

ensure_live_secrets() {
  LIVE_STORED=""
  FEATURES_STORED=""
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
    case " ${LAB_FEATURE_KEYS:-} " in
      *" $name "*) FEATURES_STORED="${FEATURES_STORED:+$FEATURES_STORED, }$name" ;;
      *) LIVE_STORED="${LIVE_STORED:+$LIVE_STORED, }$name" ;;
    esac
  done
  if [ -n "$FEATURES_STORED" ]; then
    echo "seed: stored the Meterless feature settings: $FEATURES_STORED"
  fi
  # The Meterless features are settings of a review, not keys: they are stored like keys but never make a lab live. The
  # line the start page shows is built from the values stored just now, which is what the review will be given.
  # H-MEM and the World Model have no switch in Proven, so they are always on and say so.
  scout=$(feature_word "${PROVEN_LAB_SCOUT_INTENT:-}")
  markov=$(feature_word "${PROVEN_MARKOVIAN:-}")
  LAB_FEATURES_STATE="H-MEM on (always), World Model on (always), Scout Intent $scout, Markovian $markov"
  export LAB_FEATURES_STATE
  if [ -n "$LIVE_STORED" ]; then
    echo "seed: stored as Actions secrets: $LIVE_STORED"
    LAB_LIVE_STATE="keys stored: $LIVE_STORED"
  else
    echo "seed: no key is set; the lab reviews in sample mode"
    LAB_LIVE_STATE="no key is stored; reviews are sample reviews"
  fi
  export LAB_LIVE_STATE
}
