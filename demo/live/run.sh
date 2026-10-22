#!/bin/bash
# Live review of this demo. Settings come from the environment.
# The script does not print them.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
root=$(cd "$here/../.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

(cd "$root" && go build -o "$work/proven" ./cmd/proven)
(cd "$root" && go build -o "$work/livecheck" ./cmd/livecheck)
(cd "$root" && go build -o "$work/reviewlive" ./demo/live/review)

# proven is built above. On this branch, proven scan accepts only --sample,
# so a live review goes through reviewlive (review.Open with sample mode off).
"$work/livecheck"
"$work/reviewlive" "$here"
