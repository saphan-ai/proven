# What make up tells a person while it works: the plan before the first Docker call, and one status line after each
# phase. Sourced by up.sh after lib.sh.
#
# Colour only on a terminal: the stream a line goes to is a terminal, NO_COLOR is not set to anything (no-color.org),
# and TERM is not dumb. Piped, in CI, or with NO_COLOR, the same words print without a control character.
#
# A phase is begun with ui_begin and ends with ui_done, which prints
#
#   (empty line)
#   ==> [2/5] Gitea, Grafana and the lab's Docker are up (32 s)
#   (empty line)
#
# in green. Docker prints a lot between the phases; the empty lines and the colour keep these lines from drowning in it.
# When the script stops while a phase is open, ui_exit (the EXIT trap of up.sh) prints the same line in red, naming the
# phase that failed and the log to read.

UI_TOTAL=5
UI_N=
UI_TITLE=
UI_LOG=
UI_T0=
UI_ADJOIN=

# ui_color FD: exit 0 when a line written to file descriptor FD may carry colour.
ui_color() {
  [ -t "$1" ] || return 1
  [ -z "${NO_COLOR:-}" ] || return 1
  [ "${TERM:-}" != dumb ] || return 1
  return 0
}

# ui_paint FD CODE TEXT: TEXT, wrapped in the colour CODE when FD allows colour.
ui_paint() {
  if ui_color "$1"; then
    printf '\033[%sm%s\033[0m' "$2" "$3"
  else
    printf '%s' "$3"
  fi
}

# ui_pr_count: how many demo pull requests the lab seeds (the rows of lab/seed/prs.tsv).
ui_pr_count() {
  grep -v '^#' "$ROOT/lab/seed/prs.tsv" | grep -c -v '^$' || true
}

# ui_duration SECONDS: "32 s", or "1 min 5 s" from a minute on.
ui_duration() {
  if [ "$1" -ge 60 ]; then
    printf '%s min %s s' $(($1 / 60)) $(($1 % 60))
  else
    printf '%s s' "$1"
  fi
}

# ui_plan: said once, after the questions and before the first Docker call.
ui_plan() {
  ui_prs=$(ui_pr_count)
  echo
  echo "What will happen next:"
  echo "  1. Build Proven's image (about 1 min; the first run downloads more)"
  echo "  2. Start Gitea, Grafana and the lab's Docker for the CI runner"
  echo "  3. Create the demo repository and $ui_prs pull requests, and start the CI runner"
  echo "  4. Proven reviews each pull request (a few minutes)"
  echo "  5. Show you where to start"
  echo
}

# ui_begin N TITLE LOG: phase N has started. TITLE names it when it fails; LOG is what to read then.
ui_begin() {
  UI_N=$1
  UI_TITLE=$2
  UI_LOG=$3
  UI_T0=$(date +%s)
}

# ui_done MESSAGE: the open phase finished. One green line, an empty line on each side.
ui_done() {
  ud_secs=$(($(date +%s) - UI_T0))
  # A line that follows a line with nothing printed between them shares its empty line: UI_ADJOIN is set for it.
  [ -n "${UI_ADJOIN:-}" ] || echo
  UI_ADJOIN=
  ui_paint 1 32 "==> [$UI_N/$UI_TOTAL] $1 ($(ui_duration "$ud_secs"))"
  echo
  echo
  UI_N=
}

# ui_fail [DETAIL]: the open phase failed. The same line as ui_done, in red, on the error stream: which phase, what, how
# long it ran, and the log to read.
ui_fail() {
  uf_secs=$(($(date +%s) - UI_T0))
  echo >&2
  ui_paint 2 31 "==> [$UI_N/$UI_TOTAL] FAILED: $UI_TITLE${1:+: $1} (after $(ui_duration "$uf_secs")). Read: $UI_LOG" >&2
  echo >&2
  echo >&2
  UI_N=
}

# ui_exit STATUS: the EXIT trap. A phase still open and a status that is not zero is a failure.
ui_exit() {
  [ "$1" -ne 0 ] && [ -n "$UI_N" ] || return 0
  ui_fail
}
