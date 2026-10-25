#!/bin/sh
# A stand-in for `docker` for the offline lab checks. It is copied to a directory
# on PATH by check-lab-gates.sh and check-review-gates.sh; it starts nothing and
# publishes nothing. Every call is appended to $LOG. State lives in $LOG.state.
#
#   STUB_SCOUT          absent (no plugin) | ok (scans pass) | fail (every scan fails)
#   STUB_FAIL_PLATFORM  with ok: the scan of that platform's image fails, e.g. linux/arm64
#   STUB_DRIFT          1: `imagetools create` publishes a digest that is not its source
#
# A staging reference has an index digest and one digest per platform; each is a
# function of the reference, so the same reference always answers the same digest.
# `imagetools inspect` records what it answered in $LOG.state/served, and
# `imagetools create` records what it published in $LOG.state/published, so a check
# can compare what was scanned with what was published.
state=$LOG.state
mkdir -p "$state"
echo "docker $*" >>"$LOG"

# digest SEED  — a sha256-shaped value derived from SEED.
digest() {
  n=$(printf '%s' "$1" | cksum | cut -d' ' -f1)
  printf 'sha256:%064d\n' "$n"
}

published_digest() {
  [ -f "$state/published" ] || return 1
  awk -v r="$1" '$1 == r { d = $2 } END { if (d == "") exit 1; print d }' "$state/published"
}

case $1 in
  scout)
    [ "${STUB_SCOUT:-absent}" = absent ] && exit 1
    case $2 in
      version) exit 0 ;;
      cves)
        [ "$STUB_SCOUT" = fail ] && exit 2
        if [ -n "${STUB_FAIL_PLATFORM:-}" ] && [ -f "$state/served" ]; then
          ref=$3
          want=${ref#registry://}
          want=${want#*@}
          if awk -v p="$STUB_FAIL_PLATFORM" -v d="$want" '$1 == "platform" && $2 == p && $4 == d { f = 1 } END { exit !f }' "$state/served"; then
            exit 2
          fi
        fi
        exit 0
        ;;
    esac
    ;;
  version) echo linux/amd64 ;;
  buildx)
    case $2 in
      imagetools)
        case $3 in
          inspect)
            ref=$4
            fmt=
            [ "${5:-}" = --format ] && fmt=$6
            idx=$(published_digest "$ref" || digest "index $ref")
            case $fmt in
              *Manifest.Digest*)
                echo "index $ref $idx" >>"$state/served"
                echo "$idx"
                ;;
              *Manifests*)
                for p in linux/amd64 linux/arm64; do
                  d=$(digest "$p $ref")
                  echo "platform $p $ref $d" >>"$state/served"
                  echo "$p $d"
                done
                echo "unknown/unknown $(digest "attestation $ref")"
                ;;
            esac
            ;;
          create)
            shift 3
            tags=
            src=
            while [ $# -gt 0 ]; do
              case $1 in
                --tag) tags="$tags $2"; shift 2 ;;
                --builder) shift 2 ;;
                *) src=$1; shift ;;
              esac
            done
            d=${src#*@}
            [ "${STUB_DRIFT:-}" = 1 ] && d=$(digest "drift $src")
            for t in $tags; do
              echo "$t $d" >>"$state/published"
            done
            ;;
        esac
        ;;
    esac
    ;;
esac
exit 0
