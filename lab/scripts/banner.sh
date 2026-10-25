# The lab's closing banner. Sourced after lib.sh by up.sh, password.sh, status.sh and banner-check.sh.
# Every URL is built from the settings in lab/versions.env: LAB_BIND, the ports, the dashboard uid.

# The page the lab hands out. Its path is a setting, and so is its port.
lab_page_url() {
  printf 'http://%s:%s%s' "$LAB_BIND" "$LAB_PAGE_PORT" "$LAB_PAGE_PATH"
}

# What the Password line shows when stdout is not a terminal: where the password is, never the password.
# The file is in the lab's state volume; the page and `make password` on a terminal are the other ways to read it.
lab_password_location() {
  printf 'file /state/password in the volume %s_seed-state' "$COMPOSE_PROJECT_NAME"
}

# lab_banner PASSWORD
# Prints the banner and nothing after it. On a terminal the Password line carries PASSWORD. Anywhere else (a
# pipe, a file, a log) it carries the file's location and PASSWORD is not written at all.
lab_banner() {
  if [ -t 1 ]; then
    shown=$1
  else
    shown=$(lab_password_location)
  fi
  printf 'Start here: %s\n' "$(lab_page_url)"
  printf '%s\n' '########################################'
  printf '%s\n' '####    Your proven lab is ready    ####'
  printf '%s\n' '########################################'
  printf '\n'
  printf '%s\n' 'To see demo pull requests scanned by proven please open:'
  printf '\n'
  printf 'Gitea: http://%s:%s/login\n' "$LAB_BIND" "$GITEA_HTTP_PORT"
  printf 'Username: %s\n' "$LAB_ADMIN_USER"
  printf 'Password: %s\n' "$shown"
  printf '\n'
  printf 'Grafana dashboard: http://%s:%s/d/%s\n' "$LAB_BIND" "$LGTM_GRAFANA_PORT" "$LGTM_DASHBOARD_UID"
  shown=
}

# lab_read_password: the password from the lab's 0600 state file, or nothing. It is read in the seed image, which
# mounts the state volume, and goes to the caller's variable, never to a log.
lab_read_password() {
  compose run --rm --no-deps -T --user 0:0 --entrypoint cat seed /state/password 2>/dev/null || true
}
