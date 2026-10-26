#!/bin/sh
# Load the public Nightshift receipts into the lab's Loki, for the Nightshift dashboard (make nightshift-receipts; make up runs
# it too). The file is lab/nightshift/receipts.public.jsonl, or NIGHTSHIFT_RECEIPTS_FILE: the receipts of Proven's own calls
# only, in the lab's public shape, with no prompt text. The importer (lab/seed/nightshift-import.sh) leaves out any other app
# and drops prompt_preview even if the file still has it. The full receipts file stays with whoever runs Nightshift.
#
# No file is not a failure: the lab ships without one until the public receipts are put there, and the dashboard says so.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"

file=${NIGHTSHIFT_RECEIPTS_FILE:-$ROOT/lab/nightshift/receipts.public.jsonl}
if [ ! -f "$file" ]; then
  echo "nightshift-receipts: no public receipts file at $file; the Nightshift dashboard says so. To load them, put the file there and run make nightshift-receipts."
  exit 0
fi
need_docker
compose run --rm --no-deps -T -e FORCE="${FORCE:-}" --entrypoint sh seed -c 'exec sh /usr/local/bin/nightshift-import.sh' <"$file"
