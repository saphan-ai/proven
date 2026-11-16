# CUSTOMER BUILD RECIPE ONLY. Proven does not publish these vendor-CLI images.
# No login, token, subscription state or credential file belongs in this build.
ARG GO_BASE=golang:1.26.10-alpine3.24
ARG NODE_BASE=node:24-alpine3.24
ARG RUNTIME_BASE=alpine:3.24
FROM ${GO_BASE} AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/proven ./cmd/proven

# npm is a build tool only. Copy the customer-installed package tree, never
# npm's own dependency tree, into the fresh Alpine runtime below.
FROM ${NODE_BASE} AS install-base
FROM install-base AS claude-install
ARG CLAUDE_VERSION=2.1.325
RUN npm install --prefix /opt/cli --no-audit --no-fund @anthropic-ai/claude-code@${CLAUDE_VERSION} \
 && /opt/cli/node_modules/.bin/claude --version
FROM install-base AS copilot-install
ARG COPILOT_VERSION=1.0.112
RUN npm install --prefix /opt/cli --no-audit --no-fund @github/copilot@${COPILOT_VERSION} \
 && /opt/cli/node_modules/.bin/copilot --version
FROM install-base AS codex-install
ARG CODEX_VERSION=0.169.0
RUN npm install --prefix /opt/cli --no-audit --no-fund @openai/codex@${CODEX_VERSION} \
 && /opt/cli/node_modules/.bin/codex --version
FROM install-base AS qwen-install
ARG QWEN_VERSION=0.27.4
ARG SHARP_VERSION=0.35.5
# Qwen 0.27.4 pins an affected optional sharp. Resolve its patched version
# explicitly instead of suppressing the image scanner's HIGH result.
RUN mkdir -p /opt/cli \
 && printf '{"private":true,"overrides":{"sharp":"%s"}}\n' "$SHARP_VERSION" > /opt/cli/package.json \
 && npm install --prefix /opt/cli --no-audit --no-fund @qwen-code/qwen-code@${QWEN_VERSION} \
 && /opt/cli/node_modules/.bin/qwen --version
FROM claude-install AS dual-install
ARG COPILOT_VERSION=1.0.112
RUN npm install --prefix /opt/cli --no-audit --no-fund @github/copilot@${COPILOT_VERSION} \
 && /opt/cli/node_modules/.bin/copilot --version

FROM ${RUNTIME_BASE} AS base
ARG SECURITY_REPOSITORY=https://dl-cdn.alpinelinux.org/alpine/edge/main
ARG GIT_PACKAGE_VERSION=2.56.1-r0
ARG MUSL_PACKAGE_VERSION=1.2.6-r6
RUN apk upgrade --no-cache \
 && printf '%s\n' "@security ${SECURITY_REPOSITORY}" >> /etc/apk/repositories \
 && apk add --no-cache ca-certificates coreutils libstdc++ \
    "git@security=${GIT_PACKAGE_VERSION}" "musl@security=${MUSL_PACKAGE_VERSION}" "musl-utils@security=${MUSL_PACKAGE_VERSION}" \
 && test "$(apk info -v | sed -n 's/^git-\([0-9].*\)$/\1/p')" = "${GIT_PACKAGE_VERSION}" \
 && test "$(apk info -v | sed -n 's/^musl-\([0-9].*\)$/\1/p')" = "${MUSL_PACKAGE_VERSION}" \
 && test "$(apk info -v | sed -n 's/^musl-utils-\([0-9].*\)$/\1/p')" = "${MUSL_PACKAGE_VERSION}" \
 && addgroup -g 10001 proven \
 && adduser -D -u 10001 -G proven -h /opt/proven proven \
 && mkdir -p /work && chown 10001:10001 /work
COPY --from=install-base /usr/local/bin/node /usr/local/bin/node
COPY --from=build /out/proven /usr/local/bin/proven
COPY --chmod=755 examples/cli/claude-text.sh /usr/local/bin/proven-claude
COPY --chmod=755 examples/cli/copilot-text.sh /usr/local/bin/proven-copilot
COPY --chmod=755 examples/cli/codex-text.sh /usr/local/bin/proven-codex
COPY --chmod=755 examples/cli/qwen-text.sh /usr/local/bin/proven-qwen
ENV HOME=/opt/proven PATH=/opt/cli/node_modules/.bin:$PATH
WORKDIR /work
ENTRYPOINT ["proven"]
USER 10001:10001

FROM base AS claude
COPY --from=claude-install /opt/cli /opt/cli
RUN claude --version && ! command -v npm

FROM base AS copilot
COPY --from=copilot-install /opt/cli /opt/cli
RUN copilot --version && ! command -v npm

FROM base AS dual
COPY --from=dual-install /opt/cli /opt/cli
RUN claude --version && copilot --version && ! command -v npm

FROM base AS codex
COPY --from=codex-install /opt/cli /opt/cli
RUN codex --version && ! command -v npm

FROM base AS qwen
COPY --from=qwen-install /opt/cli /opt/cli
RUN qwen --version && ! command -v npm
