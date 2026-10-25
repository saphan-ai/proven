# The lab's start page: the template with the settings, the login and the real pull requests filled in.
# Sourced by seed.sh, and by lab/lab_test.go, which renders a page offline and reads it.
#
# render_page reads these variables: PAGE_TEMPLATE, PAGE_DIR (where lab.html goes), the file /tmp/pr-rows.html
# (the table rows), CLEAN_COUNT, LAB_URL, LAB_ADMIN_USER, LAB_DEMO_REPO, LAB_ADMIN_PASSWORD, LAB_GRAFANA_URL,
# LAB_DOCS_URL and LAB_PAGE_URL. PR_ROWS_FILE names the rows file when it is not the default.

html_escape() {
  sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g'
}

# The copy goes to PAGE_DIR and nowhere else. The password is never printed and never put on a command line that
# is logged; the page file is the only place it is written.
render_page() {
  if [ ! -f "$PAGE_TEMPLATE" ]; then
    echo "seed: $PAGE_TEMPLATE is not mounted" >&2
    exit 1
  fi
  rows=${PR_ROWS_FILE:-/tmp/pr-rows.html}
  work=${PAGE_WORK:-/tmp/page.html}
  case $CLEAN_COUNT in
    1) clean_word=One ;;
    2) clean_word=Two ;;
    3) clean_word=Three ;;
    *) clean_word=$CLEAN_COUNT ;;
  esac
  # The value goes through the HTML escape, then the sed escape for the replacement text.
  pw=$(printf '%s' "$LAB_ADMIN_PASSWORD" | html_escape | sed 's/[\\&|]/\\&/g')
  sed \
    -e "s|@@GITEA_URL@@|${LAB_URL%/}|g" \
    -e "s|@@GRAFANA_URL@@|${LAB_GRAFANA_URL:-}|g" \
    -e "s|@@DOCS_URL@@|${LAB_DOCS_URL:-}|g" \
    -e "s|@@PAGE_URL@@|${LAB_PAGE_URL:-}|g" \
    -e "s|@@LAB_ADMIN_USER@@|$LAB_ADMIN_USER|g" \
    -e "s|@@LAB_DEMO_REPO@@|$LAB_DEMO_REPO|g" \
    -e "s|@@CLEAN_COUNT@@|$clean_word|g" \
    -e "s|@@LAB_PASSWORD@@|$pw|g" \
    "$PAGE_TEMPLATE" >"$work"
  awk -v rows="$rows" '
    /^@@PR_ROWS@@$/ { while ((getline line < rows) > 0) print line; next }
    { print }
  ' "$work" >"$PAGE_DIR/lab.html.new"
  rm -f "$work"
  chmod 644 "$PAGE_DIR/lab.html.new"
  mv "$PAGE_DIR/lab.html.new" "$PAGE_DIR/lab.html"
  if grep -q '@@[A-Z_]*@@' "$PAGE_DIR/lab.html"; then
    echo "seed: the start page still holds a placeholder" >&2
    exit 1
  fi
}
