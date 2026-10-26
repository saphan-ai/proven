# STATUS — lab-demo · round 5

- state: ready-for-review
- executor: a Claude Code agent (an executor seat; not a Cursor agent, as the order named the backend)
- model: Sonnet 5.5 (`claude-sonnet-5-5`)
- last commit: see the branch tip
- updated: 2026-10-26 UTC

## Summary
The named review rejected round 4: the Access token entered Compose. It was worse than that one token. Every credential of the seed (`PROVEN_GENERATOR_KEY`, `PROVEN_GUARDRAIL_KEY`, `METERLESS_RUNTIME_TOKEN`, `ZETARIS_API_KEY`, `JUDGES_GATEWAY_TOKEN`, `JUDGES_GATEWAY_ACCESS_TOKEN`) was passed by name from the `make up` shell into the seed container's environment, and the Zetaris administrator key did the same for the register container. Now `make up` writes the credentials into the seed's state volume as one 0600 file, `/state/secrets`, over standard input. The seed reads it, stores each non-empty one as an Actions secret, and removes the file on every way out. No container has a credential in its environment, and `lab/compose.yaml` names none. Nothing was run against Docker.

## Red items (read these first)
1. **No Docker on this seat, so the live controls are not run.** The sentinel check on `docker compose config`, `docker inspect`, process arguments and logs, "the Actions secret reaches a completed review", one sample run, one run with a dummy key, and the inherited host checks (password, listeners, one real review per seeded PR, dashboard ingestion, paths A/B/C) are all UNRUN. The offline controls below stand in for the transport only. They do not prove the live behaviour. `make check-lab-secrets` is the live control, written here, never executed.
2. **Path B is not shown to complete a review.** The review said an authentication probe is not a review. Nothing in this round changes that; it stays open until a Docker host runs it, or path B is left unavailable until W18.
3. **The three broken documentation links are still there** (`OSV_TABLE.md`, `ZETARIS_USL.md`, `POLICY.md` in `docs-site/generated/settings.md`). The order for this round did not name them; not touched.

## What was built
- `lab/versions.env`: `LAB_SECRET_KEYS`, the six credentials. Classed secret because each one opens a service. Everything else in `LAB_LIVE_KEYS` (modes, URLs, a model name, an organisation id, the two feature switches) opens nothing and stays in the Compose environment by name.
- `lab/compose.yaml`: the six names are gone from the seed's environment, and `ZETARIS_ADMIN_API_KEY` is gone from `zetaris-register`. The seed gets only the list of names (`LAB_SECRET_KEYS`).
- `lab/scripts/lib.sh`: `lab_write_secrets` writes `NAME=value` lines over stdin with `umask 077` (`printf` is a builtin; no value is an argument). A value with a line break is refused by name before anything is written. `lab_remove_secrets` removes the file.
- `lab/scripts/up.sh`: writes the file just before the seed, removes it after the seed whatever its status.
- `lab/seed/seed.sh`, `lab/seed/golive.sh`: the seed removes the file on EXIT, TERM and INT, and right after the secrets are stored. A name in `LAB_SECRET_KEYS` is read from the file only, never from the environment. A name absent from the file is treated as empty, so the secret is removed as before.
- `lab/zetaris/register.sh`, `lab/scripts/zetaris.sh`: the administrator key goes to the register container on standard input.
- `lab/scripts/check-lab-secrets.sh` and `make check-lab-secrets`: the live control for a Docker host.
- Docs: `docs/LAB.md` (a section "How a credential travels"), ADR-0052 amended in place, `docs/CONFIGURATION.md`.
- Livecheck reads the keys from `lab/.env` on the host as before.

## Controls
Each names the field it mutates.
- `TestSecretsNoCredentialNameIsInCompose`. Mutated field: one credential name put back into the seed's environment block (`JUDGES_GATEWAY_ACCESS_TOKEN`): red. Second mutation: `ZETARIS_ADMIN_API_KEY` put back into the register service: red.
- `TestSecretsTheListsAreHeldTogether`: the seed's environment names exactly `LAB_LIVE_KEYS` minus `LAB_SECRET_KEYS`; every credential is in `LAB_LIVE_KEYS`.
- `TestSecretsTravelOnStandardInputIntoAFileOnlyTheSeedCanRead`: with a sentinel in each name and a recording stand-in for `docker`, no sentinel is in any argument; each is on stdin and in the file; the file is mode 0600; an empty credential is not written. Mutated field: the `-c` script of the one call carries a value: the argument check goes red. Mutated field: `umask 077` to `022`: the mode check goes red.
- `TestSecretsAValueWithALineBreakIsRefusedBeforeAnythingIsWritten`: docker is never called.
- `TestSecretsTheFileIsGoneAfterASeedThatFailedMidway`: the real `seed.sh` stops at its first check and the file made for it is gone. Mutated field: the EXIT trap replaced by a no-op: the file stays, red. The failure is the earliest one (a required setting is missing), not a failure after the secrets were stored; a later failure uses the same trap and is not separately shown.
- `TestSecretsTheSeedStoresACredentialFromTheFileAndRemovesNothingItDidNotWrite`: with the Access token only in the file and nothing in the environment, the stub of Gitea receives a PUT of the sentinel; a plain setting still comes from the environment; a credential present only in the environment is not stored.
- Round-3 and round-4 controls: `go test -count=1 ./lab/...` passes in full (the register test now feeds the key on stdin; its missing-key control now expects "no administrator key arrived").

## Evidence
- `go vet ./...`: exit 0. `gofmt -l lab cmd internal`: nothing. `go test -count=1 ./lab/... ./cmd/livecheck/`: ok.
- `sh -n` on every `.sh` under `lab/scripts`, `lab/seed`, `lab/zetaris`: exit 0.
- `sh lab/scripts/golive-check.sh`, `check-lab-gates.sh`, `check-review-gates.sh`, `check-lab-password.sh`, `e2e-parse-check.sh`: exit 0.

## Not run
- Docker, Compose, Gitea, the runner: none. Everything under red item 1.
- `make check-lab-secrets` and the mutation "put one name back into Compose" against a real `docker compose config`: the offline version of that mutation is run (above); the live one is not.
- A pseudo-terminal: the terminal cases are not run on this seat.
- `node docs-site/scripts/guards.mjs`: the command needed approval and was not run.
- `make` itself is not on the seat's command list; the scripts and tests behind the targets were run directly.

## Judgement calls
1. One file, `NAME=value` lines, rather than one file per name. The order said one file. Cost if wrong: a value with a line break would split; `lab_write_secrets` refuses one by name, and the flow already limits a token's characters.
2. `JUDGES_GATEWAY_URL` is classed a setting, not a credential: it is an address the judge was sent, and the seed shows it on the start page as a link. If it should be secret, one name moves from the Compose list to `LAB_SECRET_KEYS`.
3. The Zetaris administrator key goes on standard input of the register container, not through the state volume: that container does not mount the volume and runs as another user, so a file the seed owns at 0600 would not be readable there. Cost if wrong: none for the property asked (no environment, no Compose).
4. The seed removes the file right after the secrets are stored, as well as on exit, to shorten the time it exists. `make up` removes it again after the seed for a seed that was killed.
5. Host environment: the `docker` client process on the host still has the credentials in its own environment (they are in the judge's shell). Compose does not forward them, since no service names them. This is not a container environment.

## master-was-wrong
(a) The master read the Access token as one instance of the transport; it was every name from `PROVEN_GENERATOR_MODE:` to `PROVEN_LAB_SCOUT_INTENT:`. That reading holds for the credentials, and not for the rest of the block: the modes, addresses, model name, organisation id and the two feature switches are not credentials. They were left where they are. (b) The order named the seed. One more place carried a credential: the register container's `ZETARIS_ADMIN_API_KEY`. It is moved too. (c) The review's remedy, "send the credential directly to the seed's secret-writing request", was not taken literally: the seed is a separate container started by Compose, so the value has to reach it somehow; the state volume is the same route the lab password already uses.

## Blocked on / next
Nothing blocks. Next: the master runs the live controls on a Docker host: `make up` with a sentinel in each credential's name, then `make check-lab-secrets`; one sample run; one run with a dummy key; then the inherited host checks.

## REDACTED
Nothing was blacked out: no credentials, host names, user names, absolute paths or private-record names appear above (0 items).
