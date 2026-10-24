#!/usr/bin/env bash
# PV-016 check. Run from the repository root: bash test/corpus/cases/PV-016/check.sh
# For base and for change it builds the tool, then prints one line per code-facing claim in claims.tsv:
#   claim, what the pages say (the quoted sentence, when the version has that page), what the code did.
# Exit non-zero on a build or run failure, or when a change-side quote is missing.
set -u
case_dir=test/corpus/cases/PV-016
work_root="$PWD/.saphan/work/corpus-pv-016"
mkdir -p "$work_root" || exit 1
work=$(mktemp -d "$work_root/check.XXXXXX") || exit 1
trap 'rm -rf "$work"; rmdir "$work_root" "$PWD/.saphan/work" "$PWD/.saphan" 2>/dev/null || :' EXIT
trap 'exit 1' HUP INT TERM

# Base and change must hold the same code.
for f in go.mod alerts.go main.go; do
  if ! diff -q "$case_dir/base/$f" "$case_dir/change/$f" >/dev/null; then
    echo "check: base/$f and change/$f differ" >&2
    exit 1
  fi
done

status=0
for version in base change; do
  bin="$work/tool-$version"
  if ! go -C "$case_dir/$version" build -o "$bin" . ; then
    echo "check: build of $version failed" >&2
    exit 1
  fi
  while IFS=$'\t' read -r id kind arg claimed page quote; do
    page_file="$case_dir/$version/docs/alerts/$page"
    if [ -f "$page_file" ] && grep -qF -- "$quote" "$page_file"; then
      says="\"$quote\" ($page)"
      has_page=1
    else
      says="(no such page)"
      has_page=0
      if [ "$version" = change ]; then
        echo "check: change $id: quote not found in $page" >&2
        status=1
      fi
    fi
    case "$kind" in
      dest)
        got=$("$bin" dest "$arg") || { echo "check: run failed: $version $id" >&2; status=1; continue; }
        did="setting $arg: $got"
        ;;
      deliver)
        got=$("$bin" emit "$arg") || { echo "check: run failed: $version $id" >&2; status=1; continue; }
        did="emit $arg: $got"
        ;;
      fields)
        got=$("$bin" fields "$arg") || { echo "check: run failed: $version $id" >&2; status=1; continue; }
        absent=""
        if [ "$has_page" = 1 ]; then
          for f in ${claimed//,/ }; do
            case ",$got," in
              *",$f,"*) ;;
              *) absent="$absent${absent:+,}$f" ;;
            esac
          done
          did="alert $arg: present=$got; claimed-but-absent=${absent:-none}"
        else
          did="alert $arg: present=$got"
        fi
        ;;
      *)
        echo "check: unknown kind $kind" >&2
        status=1
        continue
        ;;
    esac
    printf '%s | %s | pages say: %s | code did: %s\n' "$version" "$id" "$says" "$did"
  done < "$case_dir/claims.tsv"
done
exit "$status"
