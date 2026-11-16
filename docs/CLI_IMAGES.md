# Build your own image with licensed CLIs

**You build the vendor-CLI image in your own environment. Proven distributes the CLI-free product image and build recipes, not images containing Claude Code, Copilot or your subscription state.** This is the project's distribution boundary. Install only the tools you are entitled to use and keep the resulting image under your own control.

The pattern is a common base, a pinned customer-added vendor layer, a build-time version check and a final non-root user. Authentication happens afterwards in a private runtime profile. A credential is never a Docker build argument or copied layer.

## 1. Inspect the customer build configuration

The [Dockerfile](../examples/cli/Dockerfile) builds Proven from this checkout, prepares a Node runtime and offers three customer targets. The [Bake configuration](../examples/cli/docker-bake.hcl) names versions and output tags:

| Target | Local image | Customer-installed tools |
| --- | --- | --- |
| `claude` | `proven-claude:local` | Claude Code. |
| `copilot` | `proven-copilot:local` | GitHub Copilot CLI. |
| `dual` | `proven-dual:local` | Both. |

Both the published product and customer recipe use Alpine with explicitly pinned patched git/musl packages. The customer build installs vendor packages in a separate Node stage, then copies their package tree and Node binary into a fresh runtime. npm and its dependency tree are not part of that runtime. Both vendor packages provide musl variants. The context allowlist excludes credentials, local state and unrelated workspace material.

From the Proven repository root, with Docker Buildx:

```sh
docker buildx bake -f examples/cli/docker-bake.hcl --print
docker buildx bake -f examples/cli/docker-bake.hcl --load dual
docker run --rm proven-dual:local --version
docker run --rm --entrypoint claude proven-dual:local --version
docker run --rm --entrypoint copilot proven-dual:local --version
```

Choose `claude` or `copilot` instead of `dual` if you only need one. The build invokes vendor version checks and returns to UID/GID `10001:10001`. No login occurs while building. Defaults are exact CLI versions, not `latest`; the Node/Go base tags should be pinned to verified digests for your retained deployment.

Override your own build configuration explicitly, for example:

```sh
CLAUDE_VERSION=2.1.325 COPILOT_VERSION=1.0.112 \
  docker buildx bake -f examples/cli/docker-bake.hcl --load dual
```

Record the versions and resulting image identity. Changing a CLI version requires checking its flags and revalidating the wrapper. No release workflow in this repository builds or pushes these vendor targets.

## 2. Scan the image you built

Use your organization's scanner and block HIGH/CRITICAL before using or distributing your own image. A clean scan of the official Proven product does not cover the Node or vendor packages you added. npm is used only in the discarded installation stage.

If you choose Docker Scout, its scan may send image metadata to Docker's service. For an explicitly selected local platform, following this project's archive approach:

```sh
PROVEN_CLI_PLATFORM=$(docker image inspect proven-dual:local --format '{{.Os}}/{{.Architecture}}')
docker image save --platform "$PROVEN_CLI_PLATFORM" -o proven-dual.tar proven-dual:local
docker scout cves archive://proven-dual.tar \
  --only-severity critical,high --ignore-suppressed=false --exit-code
```

Keep the scan result, time, image ID and platform. If it fails, update affected base/packages or versions and rebuild; do not interpret an unscanned customer image as inheriting the product's clean result. [Containers](CONTAINERS.md) explains the stronger multi-platform product publication gate. These instructions do not publish your image.

## 3. Create separate private login homes

The recipe's final Linux arm64 build was checked on 16 November 2026: both CLI version commands, Git 2.56.1, non-root execution, writable private home, `timeout` and absence of npm passed. Docker Scout indexed 95 packages and reported 0 HIGH / 0 CRITICAL in that archive. The earlier Debian candidate was rejected for 15 HIGH results; the final recipe uses the fresh patched Alpine runtime above. This does not establish another architecture, future rebuild or authenticated subscription behavior.

Provision one volume per account/CLI, so the Claude container does not receive the Copilot profile and vice versa:

```sh
docker volume create proven-claude-home
docker volume create proven-copilot-home
```

Authenticate Claude Code using its own official flow:

```sh
docker run --rm -it --mount type=volume,src=proven-claude-home,dst=/opt/proven \
  --entrypoint sh proven-dual:local -c \
  'export CLAUDE_CONFIG_DIR="$HOME/.config/proven/claude"; claude auth login'
```

Authenticate Copilot and disable hooks in that dedicated profile:

```sh
docker run --rm -it --mount type=volume,src=proven-copilot-home,dst=/opt/proven \
  --entrypoint sh proven-dual:local -c \
  'export COPILOT_HOME="$HOME/.config/proven/copilot"; copilot login --device-code && copilot config disableAllHooks true'
```

Complete those vendor browser/device steps yourself. The profile stays in its volume when the disposable container exits. On headless Linux, a CLI can store credentials in profile files when a keychain is unavailable; protect the Docker host, volume access and backups accordingly. Do not assume a macOS keychain login can be moved by copying a directory. See [CLI authentication](CLI_MODELS.md).

## 4. Use the image as the model transport

For full test proving, run the **Proven binary on the host**, with its prepared [runner](RUNNER.md). The customer image runs only the model CLI. It receives the prompt on stdin and its own login volume; no target checkout or Docker socket is mounted into that model container.

```sh
export PROVEN_GENERATOR_MODE=command
export PROVEN_GENERATOR_MODEL=sonnet
export PROVEN_GENERATOR_COMMAND='["docker","run","--rm","-i","--cap-drop","ALL","--security-opt","no-new-privileges","--mount","type=volume,src=proven-claude-home,dst=/opt/proven","--entrypoint","timeout","proven-dual:local","--kill-after=2s","290s","proven-claude","{model}"]'

export PROVEN_SECOND_REVIEWER_MODE=command
export PROVEN_SECOND_REVIEWER_MODEL=auto
export PROVEN_SECOND_REVIEWER_COMMAND='["docker","run","--rm","-i","--cap-drop","ALL","--security-opt","no-new-privileges","--mount","type=volume,src=proven-copilot-home,dst=/opt/proven","--entrypoint","timeout","proven-dual:local","--kill-after=2s","290s","proven-copilot","{model}"]'
export PROVEN_GENERATOR_TIMEOUT=300s
export PROVEN_SECOND_REVIEWER_TIMEOUT=300s

PROVEN_REQUIRE_COMPLETE=true proven scan --baseline-commit <base-sha> \
  --timeout 1800 --json-output review.json /path/to/repository
```

Select actual model IDs available to each account. Replace the local image tag with the identity you recorded for repeatable operation. A non-default Docker context can be selected explicitly in the argv rather than relying on arbitrary environment inheritance.

The model containers need their provider connection. They are different from proof containers, which always have no network or inherited secrets. The host arrangement also avoids the daemon bind-path problem caused by running the entire reviewer inside a container. Proven does not enforce its HTTP network policy inside an arbitrary CLI; command mode remains refused under `PROVEN_NETWORK=local`.

## 5. Refresh or remove credentials

The example gives the CLI inside each model container its own 290-second limit, below Proven's 300-second process limit, with a 2-second kill margin. Killing a Docker client does not guarantee that its daemon-managed container stops immediately. An early review cancellation can therefore leave that container running until its internal bound; inspect active containers if cleanup is interrupted. For immediate process-group cancellation without this daemon boundary, use the host CLI recipes. Do not remove the internal bound when adapting the container command.

Re-run the appropriate login command to refresh that profile without rebuilding the image. Follow the vendor's logout/revocation flow to remove access. Removing a container does not remove its named volume, and deleting a local volume does not necessarily revoke a server-side token. Back up only if your policy permits storing those credentials.

Keep images and credentials separate throughout upgrades. Share a recipe and pinned versions with colleagues; each person provisions their own licensed CLI/account. The wrapper contract checks use fake executables and do not certify your account, quota, image scan or a successful live dual review. Validate a short prompt and known defect/control cases before enforcement.
