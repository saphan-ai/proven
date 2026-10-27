# Calls to the lab Gitea's API from the host. Sourced after lib.sh.
# The calls run in the seed image, on the lab network, as the lab admin.
# The admin's password stays in the lab's state volume: each call reads it there.

gitea_login() {
  if ! compose run --rm --no-deps -T --entrypoint test seed -s /state/password </dev/null 2>/dev/null; then
    die "the lab admin password is not in the lab's state volume; run make up"
  fi
}

# gitea_curl ARGS... — curl as the admin. Fails on an HTTP error.
gitea_curl() {
  # The password never reaches a command line (ps shows container processes too): the shell in the seed reads it
  # from the state volume and hands curl a config on fd 3, so stdin stays free for --data-binary @-.
  compose run --rm --no-deps -T --entrypoint sh seed -c '
    u=$1; shift
    pw=$(sed -e "s/[\\\\\"]/\\\\&/g" /state/password)
    curl -fsS --max-time 30 -K /dev/fd/3 "$@" 3<<EOF
user = "$u:$pw"
EOF' sh "$LAB_ADMIN_USER" "$@"
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
