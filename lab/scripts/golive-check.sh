#!/bin/sh
# Check the go-live wiring offline (no Docker, no network, no key): the keys of lab/.env become Actions secrets and
# an empty one does not (against a stub of Gitea's API), the workflow's live/sample choice and the first lines of
# its comment, the loopback bind of the start page, and `make nightshift` on a host that is not a Mac.
# The tests are in lab/lab_test.go; this runs the ones named GoLive.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
if ! command -v go >/dev/null 2>&1; then
  die "check-golive: go is not installed (the checks are Go tests in lab/lab_test.go)"
fi
cd "$ROOT"
exec go test -count=1 -run 'GoLive' ./lab
