#!/bin/sh
# One demo pull request's report page: the review Proven wrote with `proven scan --md`, set in the way GitHub shows a
# Markdown comment (light and dark), written into the lab's page volume where the page's nginx serves it.
#
#   report.sh comment NUM              the review Proven posted on pull request NUM, as Markdown on standard output (exit 3
#                                      while it is not posted): the page is that comment, never a review made apart from it
#   report.sh render NUM < review.md   the page for pull request NUM, from the Markdown on standard input
#   report.sh stub NUM                 the page that says the report is not ready yet (the seed writes one per pull request,
#                                      so a Report link is never a dead link; `make reports` or make up replaces it)
#
# It runs in the seed image, as its unprivileged user, and reads no secret: the page carries no password, no token and no
# address but the lab's own. The settings it reads (all set by lab/compose.yaml for the seed): STATE_DIR (prs.tsv, written by
# the seed: number, branch, fixture, verdict), PRS_FILE (the list, for the title), REPORT_TEMPLATE, PAGE_DIR, LAB_URL,
# LAB_ADMIN_USER, LAB_DEMO_REPO, LAB_PAGE_URL and LAB_MODE.
set -eu
# The page directory is read by the page's nginx, a different user from this one. Whatever umask the seed runs under, the
# directory and its pages are world-readable: a 700 directory made every Report link a 403.
umask 022

here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
. "$here/render.sh"

STATE_DIR=${STATE_DIR:-/state}
PAGE_DIR=${PAGE_DIR:-/page}
PRS_FILE=${PRS_FILE:-/seed/prs.tsv}
REPORT_TEMPLATE=${REPORT_TEMPLATE:-/seed/report.html.tmpl}
TAB=$(printf '\t')
# The most Markdown one report takes in: Proven bounds its comment far below this.
REPORT_MAX_BYTES=1000000

fail() {
  echo "report: $1" >&2
  exit 1
}

mode=${1:-}
num=${2:-}
case $mode in
  comment | render | stub) ;;
  *) fail "usage: report.sh comment NUM, report.sh render NUM < review.md, or report.sh stub NUM" ;;
esac
case $num in
  '' | *[!0-9]*) fail "the pull request number is not a number" ;;
esac

# The comment: the first one on the pull request that the lab's bot account wrote and whose body starts with Proven's marker
# (lab/workflow/proven.yml posts it and edits it in place), with the marker line removed. The demo repository is public, so
# the read carries no credential. jq decodes the JSON string, escapes included.
if [ "$mode" = comment ]; then
  url="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/issues/$num/comments?limit=50"
  json=$(curl -fsS --max-time 20 "$url") || fail "the comments of pull request $num could not be read"
  review=$(printf '%s' "$json" | jq -r --arg bot "$LAB_BOT_USER" \
    'first(.[] | select(.user.login == $bot and (.body | startswith("<!-- proven-review"))) | .body) // empty') ||
    fail "the comments of pull request $num are not the list Gitea returns"
  [ -n "$review" ] || exit 3
  printf '%s\n' "$review" | sed '1d'
  exit 0
fi

# The branch and the verdict of pull request NUM are what the seed stored; the title is the list's.
branch=
while IFS="$TAB" read -r n b f v; do
  if [ "$n" = "$num" ]; then
    branch=$b
    break
  fi
done <"$STATE_DIR/prs.tsv"
[ -n "$branch" ] || fail "pull request $num is not in $STATE_DIR/prs.tsv"
title=
while IFS="$TAB" read -r b f c v t s; do
  case $b in '' | \#*) continue ;; esac
  if [ "$b" = "$branch" ]; then
    title=$t
    break
  fi
done <"$PRS_FILE"
[ -n "$title" ] || fail "branch $branch is not in $PRS_FILE"

stage_dir="$PAGE_DIR/reports"
mkdir -p "$stage_dir"
chmod 755 "$stage_dir"
stage="$stage_dir/.$num.html.new"
body=$(mktemp "$stage_dir/.body.XXXXXX")
trap 'rm -f "$stage" "$body" "$body.md"' EXIT INT TERM

badge_class=neutral
badge_text="Report"
how=
note=
footer='The review as Proven wrote it with <code>proven scan --md</code>. Open a block to read its evidence.'
if [ "$mode" = stub ]; then
  badge_text="Report not built"
  note='  <div class="note">This report has not been built. Run <code>make reports</code>, then reload. Until then the review is the comment on the pull request in Gitea.</div>'
  footer='This page is a placeholder: no review is rendered here yet.'
  printf '<p>No report has been built for this pull request yet. The review is the comment on the pull request in Gitea.</p>\n' >"$body"
else
  head -c "$REPORT_MAX_BYTES" | tr -d '\r' >"$body.md"
  [ -s "$body.md" ] || fail "no Markdown on standard input"
  first=$(grep -m 1 -v '^[[:space:]]*$' "$body.md" || true)
  case $first in
    *"Proven: blocked"*)
      badge_class=bad
      badge_text="Blocked"
      ;;
    *"Proven: clean"*)
      badge_class=ok
      badge_text="Clean"
      ;;
  esac
  awk -f "$here/md2html.awk" "$body.md" >"$body"
  # The folded "How this was answered" block: what the review's own text says about how it was made, read from that text and
  # nothing typed. The agents are Proven's fixed pipeline (AGENTS.md); the queries are the Query blocks the review shows.
  queries=$(grep -c '^\*\*Query\*\*' "$body.md" || true)
  snapshots=$(grep '^\*\*Query\*\*' "$body.md" | sed -n 's/.*snapshot \([A-Za-z0-9._-]*\).*/\1/p' | sort -u | paste -s -d, - | sed 's/,/, /g')
  models=$(sed -n 's/.*reviewed by `\([^`]*\)`.*/\1/p' "$body.md" | head -n 1)
  how="  <details class=\"how\"><summary>How this was answered</summary><ul>"
  how="$how<li><b>Agents, in order:</b> Planner, Investigator, Prover, Skeptic, Judge, Writer.</li>"
  how="$how<li><b>Queries run:</b> $queries"
  [ -z "$snapshots" ] || how="$how, on snapshot $(printf '%s' "$snapshots" | html_escape)"
  how="$how.</li>"
  [ -z "$models" ] || how="$how<li><b>Reviewer model:</b> $(printf '%s' "$models" | html_escape).</li>"
  how="$how<li><b>Receipts:</b> a live review gets a signed Nightshift receipt for each model call; they are listed on the Nightshift dashboard. A sample review calls no model, so it has none.</li>"
  how="$how</ul></details>"
  rm -f "$body.md"
  # A page with no block to open does not tell the reader to open one.
  if ! grep -q '<details' "$body"; then
    footer='The review as Proven wrote it with <code>proven scan --md</code>.'
  fi
  if [ "${LAB_MODE:-sample}" = sample ]; then
    note='  <div class="note">Sample mode: recorded answers, no model was called. This is the review Proven posted on this pull request.</div>'
  else
    note='  <div class="note">This lab reviews live. This is the review Proven posted on this pull request.</div>'
  fi
fi

[ -f "$REPORT_TEMPLATE" ] || fail "$REPORT_TEMPLATE is not mounted"

# The values go in through the environment and awk's index and substr, so no character in a title is special.
R_TITLE=$(printf '%s' "$title" | html_escape) \
  R_NUM=$num \
  R_BADGE_CLASS=$badge_class \
  R_BADGE_TEXT=$badge_text \
  R_PR_URL="${LAB_URL%/}/${LAB_ADMIN_USER:-proven}/${LAB_DEMO_REPO:-discount}/pulls/$num" \
  R_LAB_PAGE_URL=${LAB_PAGE_URL:-/lab} \
  R_NOTE=$note \
  R_HOW=$how \
  R_FOOTER=$footer \
  awk -v bodyfile="$body" '
    function fill(s, key, val,    out, i, tag) {
      tag = "@@" key "@@"
      out = ""
      while ((i = index(s, tag)) > 0) {
        out = out substr(s, 1, i - 1) val
        s = substr(s, i + length(tag))
      }
      return out s
    }
    /^@@BODY@@$/ { while ((getline l < bodyfile) > 0) print l; next }
    {
      s = $0
      s = fill(s, "TITLE", ENVIRON["R_TITLE"])
      s = fill(s, "NUM", ENVIRON["R_NUM"])
      s = fill(s, "BADGE_CLASS", ENVIRON["R_BADGE_CLASS"])
      s = fill(s, "BADGE_TEXT", ENVIRON["R_BADGE_TEXT"])
      s = fill(s, "PR_URL", ENVIRON["R_PR_URL"])
      s = fill(s, "LAB_PAGE_URL", ENVIRON["R_LAB_PAGE_URL"])
      s = fill(s, "NOTE", ENVIRON["R_NOTE"])
      s = fill(s, "HOW", ENVIRON["R_HOW"])
      s = fill(s, "FOOTER", ENVIRON["R_FOOTER"])
      print s
    }
  ' "$REPORT_TEMPLATE" >"$stage"

chmod 644 "$stage"
mv "$stage" "$stage_dir/$num.html"
trap - EXIT INT TERM
rm -f "$body"
echo "report: pull request $num: $mode written to reports/$num.html ($badge_text)"
