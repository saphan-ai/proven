# md2html.awk: the Markdown that `proven scan --md` writes, as HTML for the lab's report page. POSIX awk, so it runs in the
# seed image (busybox) and on a developer's machine alike.
#
# It reads the subset Proven writes (docs/OUTPUT_FORMATS.md): paragraphs, **bold**, `code`, pipe tables, bullet lists,
# fenced code blocks, backslash escapes, HTML entities Proven already escaped, and the three raw-HTML shapes it emits
# (<details><summary>...</summary>, </details>, <sub>...</sub>). Nothing else is passed through: every other < or > in a
# line is written as an entity, so the page can carry no tag the review did not mean to carry. No script, no link, no
# image, no remote address is ever written by this file.
#
# Not rendered, because Proven does not write them: headings, emphasis with *, links, images, nested lists, indented code.

# ent(s): s as HTML text. An entity already in s (&lt; &#42;) is kept; a bare & becomes &amp;.
function ent(s,    out, c, rest) {
  out = ""
  while (length(s) > 0) {
    c = substr(s, 1, 1)
    if (c == "&") {
      rest = substr(s, 1, 12)
      if (match(rest, /^&[A-Za-z0-9#]+;/)) {
        out = out substr(s, 1, RLENGTH)
        s = substr(s, RLENGTH + 1)
        continue
      }
      out = out "&amp;"
    } else if (c == "<") {
      out = out "&lt;"
    } else if (c == ">") {
      out = out "&gt;"
    } else {
      out = out c
    }
    s = substr(s, 2)
  }
  return out
}

# lit(s): s as literal text, as inside a code span or a fence: every & is itself.
function lit(s) {
  gsub(/&/, "\\&amp;", s)
  gsub(/</, "\\&lt;", s)
  gsub(/>/, "\\&gt;", s)
  return s
}

# plain(s): text outside code spans. A backslash before punctuation is dropped, and an escaped * is kept from bold.
function plain(s,    out, c, n, bold, a, b) {
  out = ""
  while (length(s) > 0) {
    c = substr(s, 1, 1)
    n = substr(s, 2, 1)
    if (c == "\\" && n ~ /[][!"#$%&'()*+,.\/:;<=>?@\\^_`{|}~-]/) {
      if (n == "*") {
        out = out "&#42;"
      } else if (n == "&") {
        out = out "&amp;"
      } else if (n == "<") {
        out = out "&lt;"
      } else if (n == ">") {
        out = out "&gt;"
      } else {
        out = out n
      }
      s = substr(s, 3)
      continue
    }
    out = out c
    s = substr(s, 2)
  }
  # Entities and stray angle brackets, over the text that is left.
  out = ent_keep(out)
  # **bold**
  bold = ""
  while ((a = index(out, "**")) > 0) {
    b = index(substr(out, a + 2), "**")
    if (b == 0) {
      break
    }
    bold = bold substr(out, 1, a - 1) "<strong>" substr(out, a + 2, b - 1) "</strong>"
    out = substr(out, a + 2 + b + 1)
  }
  return bold out
}

# ent_keep(s): like ent, but the entities plain() itself wrote are already in s, so only < > and a bare & change.
function ent_keep(s) {
  return ent(s)
}

# inline(s): code spans first (their text is literal), then the rest.
function inline(s,    out, before, code) {
  out = ""
  while (match(s, /`[^`]+`/)) {
    before = substr(s, 1, RSTART - 1)
    code = substr(s, RSTART + 1, RLENGTH - 2)
    out = out plain(before) "<code>" lit(code) "</code>"
    s = substr(s, RSTART + RLENGTH)
  }
  return out plain(s)
}

function trim(s) {
  sub(/^[ \t]+/, "", s)
  sub(/[ \t]+$/, "", s)
  return s
}

# cells(line, parts): splits a pipe-table row; \| is a pipe inside a cell. Returns the cell count.
function cells(line, parts,    n, i, raw) {
  gsub(/\\\|/, "\001", line)
  n = split(line, raw, "|")
  # raw[1] is before the first pipe and raw[n] after the last: both empty in a well-formed row.
  for (i = 2; i < n; i++) {
    parts[i - 1] = trim(raw[i])
    gsub(/\001/, "|", parts[i - 1])
  }
  return n - 2
}

function close_block() {
  if (mode == "para") {
    print "</p>"
  } else if (mode == "list") {
    print "</ul>"
  } else if (mode == "table") {
    print "</tbody></table></div>"
  } else if (mode == "head") {
    print "</thead></table></div>"
  }
  mode = ""
}

function table_head(line,    n, i, c, cls) {
  n = cells(line, hdr)
  ncols = n
  print "<div class=\"table-wrap\"><table><thead><tr>"
  for (i = 1; i <= n; i++) {
    print "<th>" inline(hdr[i]) "</th>"
  }
  print "</tr></thead><tbody>"
  mode = "head"
}

function table_row(line,    n, i, cls, v) {
  n = cells(line, row)
  print "<tr>"
  for (i = 1; i <= n; i++) {
    cls = ""
    if (hdr[i] == "Severity") {
      v = tolower(row[i])
      if (v == "error" || v == "warning" || v == "info") {
        cls = " class=\"sev sev-" v "\""
      }
    }
    print "<td" cls ">" inline(row[i]) "</td>"
  }
  print "</tr>"
}

BEGIN { mode = ""; fence = 0 }

{
  line = $0
  sub(/\r$/, "", line)
}

fence {
  if (line ~ /^```/) {
    print "</code></pre>"
    fence = 0
  } else {
    print lit(line)
  }
  next
}

line ~ /^```/ {
  close_block()
  lang = substr(line, 4)
  gsub(/[^A-Za-z0-9_-]/, "", lang)
  print "<pre class=\"fence\"><code" (lang != "" ? " class=\"lang-" lang "\"" : "") ">"
  fence = 1
  next
}

line ~ /^[ \t]*$/ {
  close_block()
  next
}

line ~ /^<details><summary>.*<\/summary>$/ {
  close_block()
  inner = substr(line, 19, length(line) - 18 - 10)
  print "<details><summary>" ent(inner) "</summary>"
  next
}

line == "</details>" {
  close_block()
  print "</details>"
  next
}

line ~ /^<sub>.*<\/sub>$/ {
  close_block()
  print "<p class=\"foot\"><sub>" inline(substr(line, 6, length(line) - 5 - 6)) "</sub></p>"
  next
}

line ~ /^\|/ {
  if (mode == "table" || mode == "head") {
    if (mode == "head") {
      # The delimiter row (| --- | --- |) is the line after the header; it carries no cell.
      if (line ~ /^\|[ :|-]+\|[ \t]*$/) {
        mode = "table"
        next
      }
      mode = "table"
    }
    table_row(line)
    next
  }
  close_block()
  table_head(line)
  next
}

line ~ /^- / {
  if (mode != "list") {
    close_block()
    print "<ul>"
    mode = "list"
  }
  print "<li>" inline(substr(line, 3)) "</li>"
  next
}

{
  if (mode != "para") {
    close_block()
    print "<p>"
    mode = "para"
  } else {
    print "<br>"
  }
  print inline(line)
}

END {
  if (fence) {
    print "</code></pre>"
  }
  close_block()
}
