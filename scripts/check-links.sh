#!/bin/sh
# check-links.sh — every relative link in every Markdown file, and every repository path named in metadata.json,
# points at a file or directory that exists. Run from the repository root. Exit 0 when all resolve, 1 otherwise,
# with one line per broken link. ADR-0030: the product and the hackathon live apart, so a moved file must not leave
# a dangling link behind.
#
# Links checked: Markdown inline links `[text](target)` whose target is not a URL (scheme:), not an e-mail address and
# not an in-page anchor (#...). The target's `#anchor` and `?query` are ignored; the path is resolved against the
# directory of the file that contains it.
set -u
broken=0

# git ls-files: only files the repository tracks, so build output and local clutter are never checked.
for md in $(git ls-files '*.md'); do
  dir=$(dirname "$md")
  # A generated docs-site page is written as docs/ writes it (its links resolve from docs/, where the site's build resolves them).
  case "$md" in docs-site/generated/*) dir=docs ;; esac
  # One target per line; the sed keeps only what sits between "](" and the next ")".
  grep -o '\]([^)]*)' "$md" 2>/dev/null | sed 's/^](//; s/)$//; s/[[:space:]].*$//' | while IFS= read -r target; do
    case "$target" in
      ''|'#'*|*:*|mailto:*) continue ;;
    esac
    path=${target%%#*}
    path=${path%%\?*}
    [ -z "$path" ] && continue
    case "$path" in
      /*) resolved=".$path" ;;
      *) resolved="$dir/$path" ;;
    esac
    if [ ! -e "$resolved" ]; then
      echo "broken: $md -> $target"
    fi
  done
done > "${TMPDIR:-/tmp}/check-links.$$"

# metadata.json: every string value that names a path inside the repository (it starts with a top-level entry of the
# repository) must exist.
if [ -f metadata.json ]; then
  tops=$(git ls-tree --name-only HEAD | tr '\n' '|' | sed 's/|$//')
  grep -o '"[^"]*"' metadata.json | tr -d '"' | grep -E "^($tops)(/|$)" | sort -u | while IFS= read -r p; do
    p=${p%%#*}
    [ -e "$p" ] || echo "broken: metadata.json -> $p"
  done >> "${TMPDIR:-/tmp}/check-links.$$"
fi

if [ -s "${TMPDIR:-/tmp}/check-links.$$" ]; then
  cat "${TMPDIR:-/tmp}/check-links.$$"
  broken=$(wc -l < "${TMPDIR:-/tmp}/check-links.$$" | tr -d ' ')
fi
rm -f "${TMPDIR:-/tmp}/check-links.$$"
if [ "$broken" -gt 0 ]; then
  echo "check-links: $broken broken link(s)"
  exit 1
fi
echo "check-links: every relative link and metadata.json path resolves"
