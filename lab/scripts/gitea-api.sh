# Calls to the lab Gitea's API from the host. Sourced after lib.sh.
# The calls run in the seed image, on the lab network, as the lab admin.
# The admin's password is read once from the lab's state volume.

gitea_login() {
  LAB_PASSWORD=$(compose run --rm --no-deps -T --entrypoint cat seed /state/password 2>/dev/null || true)
  if [ -z "$LAB_PASSWORD" ]; then
    die "the lab admin password is not in the lab's state volume; run make up"
  fi
}

# gitea_curl ARGS... — curl as the admin. Fails on an HTTP error.
gitea_curl() {
  compose run --rm --no-deps -T --entrypoint curl seed \
    -fsS --max-time 30 -u "$LAB_ADMIN_USER:$LAB_PASSWORD" "$@"
}

# json_first KEY — read JSON on stdin, print the first string value of KEY.
json_first() {
  awk -v key="$1" '
    { buf = buf $0 }
    END {
      if (match(buf, "\"" key "\":\"[^\"]*\"")) {
        print substr(buf, RSTART + length(key) + 4, RLENGTH - length(key) - 5)
      }
    }'
}

gitea_repo_api() {
  printf 'http://%s:%s/api/v1/repos/%s/%s' \
    "$GITEA_INTERNAL_HOST" "$GITEA_CONTAINER_HTTP_PORT" "$LAB_ADMIN_USER" "$LAB_DEMO_REPO"
}
