#!/bin/sh
# Data check. Reads the data files listed in manifest.txt and prints "ok",
# or an error and a non-zero exit.
set -eu

here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)

# every file in the manifest must exist in the data directory
count=0
while read -r name; do
  if [ ! -f "$here/records/$name" ]; then
    echo "missing: $name"
    exit 1
  fi
  count=$((count + 1))
done < "$here/manifest.txt"

# the index must list as many files as the manifest does
listed=$(wc -l < "$here/data/index.txt")
if [ "$listed" -ne "$count" ]; then
  echo "index lists $listed files, manifest lists $count"
  exit 1
fi

echo ok
