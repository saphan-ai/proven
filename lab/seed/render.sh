# The lab's start page: the template with the settings, the login and the real pull requests filled in.
# Sourced by seed.sh, and by lab/lab_test.go, which renders a page offline and reads it.
#
# render_page reads these variables: PAGE_TEMPLATE, PAGE_DIR (where lab.html goes, and the only place it is staged),
# the file /tmp/pr-rows.html (the table rows), CLEAN_COUNT, LAB_URL, LAB_ADMIN_USER, LAB_DEMO_REPO, LAB_ADMIN_PASSWORD,
# LAB_GRAFANA_URL, LAB_DOCS_URL, LAB_PAGE_URL, LAB_LIVE_STATE
# (what the stored keys say, one line for the page). PR_ROWS_FILE names the rows file when it is not the default.

html_escape() {
  sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g'
}

# The page file is the only place the password is written, and it is written inside PAGE_DIR (the page volume)
# and nowhere else: the staging file is made there, with mode 0600, from one sed-and-awk pipe, so no copy of the
# page with the password in it ever exists outside the volume. A render that stops between the stages removes the
# staging file. The password is never printed and is never on a command line: awk reads it from its environment.
render_page() {
  if [ ! -f "$PAGE_TEMPLATE" ]; then
    echo "seed: $PAGE_TEMPLATE is not mounted" >&2
    exit 1
  fi
  rows=${PR_ROWS_FILE:-/tmp/pr-rows.html}
  stage="$PAGE_DIR/.lab.html.new"
  case $CLEAN_COUNT in
    1) clean_word=One ;;
    2) clean_word=Two ;;
    3) clean_word=Three ;;
    *) clean_word=$CLEAN_COUNT ;;
  esac
  # The value goes through the HTML escape; awk puts it in with index and substr, so no character in it is special.
  pw=$(printf '%s' "$LAB_ADMIN_PASSWORD" | html_escape)
  trap 'rm -f "$stage"' EXIT INT TERM
  (
    umask 077
    sed \
      -e "s|@@GITEA_URL@@|${LAB_URL%/}|g" \
      -e "s|@@GRAFANA_URL@@|${LAB_GRAFANA_URL:-}|g" \
      -e "s|@@DOCS_URL@@|${LAB_DOCS_URL:-}|g" \
      -e "s|@@PAGE_URL@@|${LAB_PAGE_URL:-}|g" \
      -e "s|@@LAB_ADMIN_USER@@|$LAB_ADMIN_USER|g" \
      -e "s|@@LAB_DEMO_REPO@@|$LAB_DEMO_REPO|g" \
      -e "s|@@CLEAN_COUNT@@|$clean_word|g" \
      -e "s|@@LIVE_STATE@@|${LAB_LIVE_STATE:-sample: no keys}|g" \
      "$PAGE_TEMPLATE" |
      LAB_PAGE_PW=$pw awk -v rows="$rows" '
        /^@@PR_ROWS@@$/ { while ((getline line < rows) > 0) print line; next }
        {
          i = index($0, "@@LAB_PASSWORD@@")
          if (i > 0) {
            $0 = substr($0, 1, i - 1) ENVIRON["LAB_PAGE_PW"] substr($0, i + 16)
          }
          print
        }
      ' >"$stage"
  )
  if grep -q '@@[A-Z_]*@@' "$stage"; then
    echo "seed: the start page still holds a placeholder" >&2
    exit 1
  fi
  chmod 644 "$stage"
  mv "$stage" "$PAGE_DIR/lab.html"
  trap - EXIT INT TERM
}
