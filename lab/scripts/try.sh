#!/bin/sh
# Try one thing Proven does, on this machine, and see what it prints (make try SCENARIO=forged | prover). The start page's
# "Try yourself" buttons for the two scenarios that are not a pull request name these commands.
#
#   forged   Proven reviews a change and signs an evidence pack; one byte of a copy is changed; proven verify accepts the
#            pack and refuses the copy. The key pair is made here, in a scratch folder that is removed when the script ends.
#   prover   a container started with the flags Proven's live Prover uses (internal/testrun, DefaultContainerConfig) tries to
#            reach the network, looks for a secret, and tries to write to the code: it can do none of the three.
#
# Both run in Proven's own image, with no network and no password, token or key of the lab: the script reads none. The only
# key is the throw-away one the forged scenario makes for itself.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
load_versions

scenario=${1:-}
case $scenario in
  forged | prover) ;;
  *) die "try: the scenario is '$scenario'; it must be forged or prover. NEXT — run make try SCENARIO=forged, or make try SCENARIO=prover." ;;
esac
need_docker

image="$PROVEN_IMAGE:latest"
if ! docker image inspect "$image" >/dev/null 2>&1; then
  die "try: the image $image is not built. NEXT — run make up (or make image), then try again."
fi

say() { printf '\n%s\n' "$1"; }

try_forged() {
  command -v openssl >/dev/null 2>&1 || die "try: forged makes a key pair with openssl, which is not on the path. NEXT — install openssl, then run make try SCENARIO=forged again."
  work=$(mktemp -d)
  trap 'rm -rf "$work"' EXIT INT TERM
  chmod 700 "$work"
  openssl genpkey -algorithm ed25519 -out "$work/pack.key" >/dev/null 2>&1
  chmod 600 "$work/pack.key"
  openssl pkey -in "$work/pack.key" -pubout -out "$work/pack.pub" >/dev/null 2>&1
  # The container runs as the caller, so it can read the 0600 key and write the pack; git is told the folder is its own.
  run() {
    docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges:true \
      --user "$(id -u):$(id -g)" -v "$work:/pack" -e HOME=/tmp -e PROVEN_LOG_DIR=/tmp/log \
      -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=safe.directory -e GIT_CONFIG_VALUE_0='*' \
      -e SAMPLE_MODE=true -e PROVEN_PACK_KEY=/pack/pack.key -e PROVEN_PACK_PUBLIC_KEY=/pack/pack.pub \
      --entrypoint sh "$image" -c "$1"
  }
  say "1. Proven reviews the discount change and signs an evidence pack."
  run 'proven scan --evidence-pack /pack/review.pack /opt/proven/fixture/discount >/dev/null; ls -l /pack/review.pack | cut -d" " -f5- | sed "s/^/   pack: /"'
  say "2. An auditor verifies the pack with the public key."
  run 'proven verify /pack/review.pack' | sed 's/^/   /'
  say "3. One byte of a copy is changed, and the auditor verifies the copy."
  run 'cp /pack/review.pack /pack/forged.pack && printf X | dd of=/pack/forged.pack bs=1 seek=700 conv=notrunc 2>/dev/null'
  if out=$(run 'proven verify /pack/forged.pack' 2>&1); then
    printf '%s\n' "$out" | sed 's/^/   /'
    die "try: the forged pack was accepted. This is a defect in Proven; please report it."
  fi
  printf '%s\n' "$out" | sed 's/^/   /'
  say "The forged copy is refused, and the real pack is accepted: a pack that is changed after it is signed does not verify."
}

try_prover() {
  say "A container started with the flags Proven's live Prover uses: no network, no secret passed in, the code read-only."
  docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges:true --read-only --cpus 1 --memory 512m \
    -v "$ROOT/testdata/fixture/discount:/code:ro" --workdir /code \
    --entrypoint sh "$image" -c '
      if wget -q -T 3 -O /dev/null http://example.com 2>/dev/null; then echo "   network: REACHED example.com (this must not happen)"; exit 1; else echo "   network: none, the request to example.com failed"; fi
      n=$(env | grep -c -i -E "token|secret|password|key" || true)
      echo "   secrets in the environment: $n"
      [ "$n" = 0 ] || exit 1
      if touch /code/changed 2>/dev/null; then echo "   code: WRITABLE (this must not happen)"; exit 1; else echo "   code: read-only, the write failed"; fi
    '
  say "The Prover can read the change and run its tests. It cannot reach the network, cannot see a secret, and cannot change the code."
}

"try_$scenario"
