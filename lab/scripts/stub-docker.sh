#!/bin/sh
# A stand-in for `docker` and `skopeo` for the offline lab checks. It is copied to a
# directory on PATH, under both names, by check-lab-gates.sh and
# check-review-gates.sh; it starts nothing and publishes nothing. Every call is
# appended to $LOG. State lives in $LOG.state.
#
#   STUB_SCOUT          absent (no plugin) | ok (scans pass) | fail (every scan fails)
#   STUB_FAIL_PLATFORM  with ok: the scan of that platform fails, e.g. linux/arm64
#   STUB_FAIL_IMAGE     with ok: the scan of the layout directory of that name fails, e.g. seed
#   STUB_DRIFT          1:`skopeo copy` publishes a digest that is not its source's
#
# `buildx build --output type=oci,dest=F` writes a tar of a layout whose index
# digest is a function of the build's Dockerfile, so the same build always has the
# same digest. Every call that writes to a registry (`--push`, `docker push`,
# `buildx imagetools create`, `skopeo copy`) also appends a line to
# $LOG.state/writes, so a check can show that a refused run wrote nothing.
# `skopeo copy` records "tag digest" in $LOG.state/published, and
# `imagetools inspect` answers from it.
state=$LOG.state
mkdir -p "$state"
tool=${0##*/}
echo "$tool $*" >>"$LOG"

# digest SEED  — a sha256-shaped value derived from SEED.
digest() {
  n=$(printf '%s' "$1" | cksum | cut -d' ' -f1)
  printf 'sha256:%064d\n' "$n"
}

published_digest() {
  [ -f "$state/published" ] || return 1
  awk -v r="$1" '$1 == r { d = $2 } END { if (d == "") exit 1; print d }' "$state/published"
}

# registry_write WHAT  — a call that writes to a registry.
registry_write() {
  echo "$1" >>"$state/writes"
}

# layout_digest DIR  — the index digest in an unpacked layout.
layout_digest() {
  tr -d ' \n\t' <"$1/index.json" | grep -o '"digest":"sha256:[0-9a-f]*"' | head -n 1 | cut -d '"' -f 4
}

if [ "$tool" = skopeo ]; then
  case $1 in
    copy)
      shift
      src=
      dst=
      for a in "$@"; do
        case $a in
          oci:*) src=${a#oci:} ;;
          docker://*) dst=${a#docker://} ;;
        esac
      done
      registry_write "skopeo copy $dst"
      d=$(layout_digest "$src")
      [ "${STUB_DRIFT:-}" = 1 ] && d=$(digest "drift $src")
      echo "$dst $d" >>"$state/published"
      ;;
  esac
  exit 0
fi

case $1 in
  scout)
    [ "${STUB_SCOUT:-absent}" = absent ] && exit 1
    case $2 in
      version) exit 0 ;;
      cves)
        [ "$STUB_SCOUT" = fail ] && exit 2
        case $3 in
          oci-dir://*/"${STUB_FAIL_IMAGE:-@none}") exit 2 ;;
        esac
        platform=
        prev=
        for a in "$@"; do
          [ "$prev" = --platform ] && platform=$a
          prev=$a
        done
        [ -n "${STUB_FAIL_PLATFORM:-}" ] && [ "$platform" = "$STUB_FAIL_PLATFORM" ] && exit 2
        exit 0
        ;;
    esac
    ;;
  version) echo linux/amd64 ;;
  push) registry_write "docker push $2" ;;
  buildx)
    case $2 in
      build)
        file=
        dest=
        prev=
        for a in "$@"; do
          case $a in
            --push) registry_write "buildx build --push" ;;
          esac
          [ "$prev" = --file ] && file=$a
          case $a in
            type=oci,dest=*) dest=${a#type=oci,dest=}; dest=${dest%%,*} ;;
            type=registry*) registry_write "buildx build --output registry" ;;
          esac
          prev=$a
        done
        if [ -n "$dest" ]; then
          tmp=$state/layout.$$
          rm -rf "$tmp"
          mkdir -p "$tmp"
          printf '{"schemaVersion":2,"manifests":[{"mediaType":"application/vnd.oci.image.index.v1+json","digest":"%s","size":1}]}\n' "$(digest "layout $file")" >"$tmp/index.json"
          echo "layout $(digest "layout $file")" >>"$state/layouts"
          tar -cf "$dest" -C "$tmp" index.json
          rm -rf "$tmp"
        fi
        ;;
      imagetools)
        case $3 in
          inspect)
            ref=$4
            idx=$(published_digest "$ref" || digest "index $ref")
            echo "$idx"
            ;;
          create)
            registry_write "buildx imagetools create"
            ;;
        esac
        ;;
    esac
    ;;
esac
exit 0
