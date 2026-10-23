# STATUS — judges-lab-makefile · round 5

- state: ready-for-review
- executor: Claude Code executor
- model: Sonnet 5.5 (claude-sonnet-5-5)
- pool: not shown to this leg
- cost: about USD 0.87 of the USD 22 cap, as the session budget line showed at the last commit
- last commit: 24f1ef5234068dc5b6144ca1a38fe7295f4a630c "[judges-lab-makefile] lab: start the PR workflow, pull the test image"

## Summary

Three changes. The demo repository's Actions unit is turned on by name, and an existing pull request is closed and reopened when the unit had to be turned on. The seeded workflow posts to Gitea's lab address. `make up` pulls the Prover's test image into the inner Docker daemon. `make check-links` is a target. This seat could not run `make` or Docker, so the lab was not started here.

## What changed

1. **Why no task was created (a claim, not measured here).** Gitea starts a pull-request workflow only when the repository's Actions unit is on. The seed created the repository through the API and never set the unit. I could not read Gitea's default units on this seat. The fix does not depend on the default: `lab/seed/seed.sh` reads `has_actions` and, when it is not true, sets it and checks the answer. It sets `RETRIGGER`, and `ensure_pr` then closes and reopens the open pull request, because a pull request opened while the unit was off has no task. `lab/compose.yaml` also adds `repo.actions` to `DEFAULT_REPO_UNITS`.
2. **The workflow's post address.** The job runs on the inner daemon, where `127.0.0.1` is not Gitea, and `github.api_url` is built from the public root URL. The seeded copy now replaces it with `http://gitea:3000/api/v1`. The sed was run on `lab/workflow/proven.yml`; `API_URL` came out as that address. This is a second, separate reason `make e2e` could still fail after the first fix; it was found by reading, not by a run.
3. **The test image.** `make up` pulls `PROVEN_TEST_IMAGE` into the inner daemon. `TEST_IMAGE_DIGEST` in `lab/versions.env` pins the pull. **It is empty.** The seat had no network and WebFetch was refused, so no digest could be read, and none was invented. With it empty, `make up` pulls by tag and prints the digest it got. The master must write that digest into `lab/versions.env`. Both settings are in `docs/CONFIGURATION.md`, and `docs/LAB.md` says so.
4. **`make check-links`** runs `scripts/check-links.sh` from the repository root and is listed in the help.

## Acceptance

- `sh scripts/check-links.sh` exit 0: "every relative link and metadata.json path resolves". Run directly; `make check-links` was not run because `make` is refused on this seat.
- `git rev-list --merges 9303b2d..HEAD` printed no commit. No `Co-authored-by` line in `9303b2d..HEAD`. The author is the fleet identity.
- `go vet ./...` exit 0. `go test -count=1 ./...` exit 0, every package `ok`.
- `sh -n` on `seed.sh`, `up.sh`, `e2e.sh`: exit 0. Syntax only.

## Not run

- `make`, `make build`, `make up`, `make e2e`, `make check-links`, `make scan`, `make down`: `make` and `docker` are refused on this seat. The master runs them.
- The negative control (runner stopped, `make e2e` exits non-zero and names the runner): not run. It passed in round 4 and this round did not change that path, except one added hint line after the timeout message.
- The Gitea API answers (`has_actions`, PATCH of the pull request state, reopening) are written from Gitea's documented API and were not tried against Gitea 1.22.6.

## Rulings

- Reopen with `PATCH {"state":"open"}`. Cost if wrong: the reopen fails and `make up` exits with `seed: could not reopen the pull request`.
- `TEST_IMAGE_DIGEST` left empty rather than guessed. Cost: the test image is not yet pinned.
- `${PROVEN_TEST_IMAGE%%:*}` takes the repository name, so an image name with a registry port is not supported. Said in `docs/CONFIGURATION.md`.

## master-was-wrong

— Nothing in this round's order was contradicted. The cause named for the missing task (the Actions unit) is my reading of Gitea's behaviour, not a measurement; if the lab still shows no task after this, read the Gitea log for the pull request event.
