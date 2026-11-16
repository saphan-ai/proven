# Proven's image: a static binary, git, and the container client, on Alpine.

ARG GOLANG_IMAGE=golang:1.26.10
ARG RUNTIME_IMAGE=alpine:3.24
FROM ${GOLANG_IMAGE} AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/proven ./cmd/proven

FROM ${RUNTIME_IMAGE}
ARG CREATED=unset
ARG REVISION=unset
ARG VERSION=unset
LABEL org.opencontainers.image.created="${CREATED}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.title="proven" \
      org.opencontainers.image.source="https://github.com/saphan-ai/proven"
# The Docker CLI is Docker's own static release, pinned by version and sha256 per platform. Alpine's docker-cli
# package is built with an older Go, whose standard library carries critical and high findings with fixes
# (Docker Scout, 2026-10-25); the static release is built with a current Go. Proven runs tests through this CLI.
ARG DOCKER_CLI_VERSION=29.9.2
ARG DOCKER_CLI_SHA256_AMD64=9ac225999fd5e863f7694c728ef23d480616b1e5639328fa309befd4243d69eb
ARG DOCKER_CLI_SHA256_ARM64=873b3e0207857c3b2f28392fb922f3ea6efe37620c16d4d4aaf2fef409c90b73
ARG TARGETARCH
# Selected security packages are pinned above the vanilla base. The tagged
# repository keeps unrelated dependencies on the stable branch where possible.
# Missing fixed versions fail the build instead of falling back to base packages.
ARG ALPINE_SECURITY_REPOSITORY=https://dl-cdn.alpinelinux.org/alpine/edge/main
ARG GIT_PACKAGE_VERSION=2.56.1-r0
ARG MUSL_PACKAGE_VERSION=1.2.6-r6
RUN set -eu; \
    case "${TARGETARCH:-amd64}" in \
      amd64) arch=x86_64; sum="${DOCKER_CLI_SHA256_AMD64}" ;; \
      arm64) arch=aarch64; sum="${DOCKER_CLI_SHA256_ARM64}" ;; \
      *) echo "docker cli: no static build for ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    apk upgrade --no-cache; \
    printf '%s\n' "@security ${ALPINE_SECURITY_REPOSITORY}" >> /etc/apk/repositories; \
    apk add --no-cache ca-certificates "git@security=${GIT_PACKAGE_VERSION}" "musl@security=${MUSL_PACKAGE_VERSION}" "musl-utils@security=${MUSL_PACKAGE_VERSION}"; \
    test "$(apk info -v | sed -n 's/^git-\([0-9].*\)$/\1/p')" = "${GIT_PACKAGE_VERSION}"; \
    test "$(apk info -v | sed -n 's/^musl-\([0-9].*\)$/\1/p')" = "${MUSL_PACKAGE_VERSION}"; \
    test "$(apk info -v | sed -n 's/^musl-utils-\([0-9].*\)$/\1/p')" = "${MUSL_PACKAGE_VERSION}"; \
    git --version; \
    wget -q -O /tmp/docker.tgz "https://download.docker.com/linux/static/stable/${arch}/docker-${DOCKER_CLI_VERSION}.tgz"; \
    echo "${sum}  /tmp/docker.tgz" | sha256sum -c -; \
    tar -xzf /tmp/docker.tgz -C /tmp docker/docker; \
    install -m 755 /tmp/docker/docker /usr/local/bin/docker; \
    rm -rf /tmp/docker.tgz /tmp/docker
RUN adduser -D -u 65532 -h /work proven \
 && mkdir -p /opt/proven/fixture \
 && chown -R proven:proven /work /opt/proven
COPY --from=build /out/proven /usr/local/bin/proven
COPY --chown=proven:proven testdata/fixture/discount /opt/proven/fixture/discount
USER proven
WORKDIR /work
ENTRYPOINT ["proven"]
