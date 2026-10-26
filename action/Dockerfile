# The Action's image: Proven's published image plus reviewdog and the entry point script.
# GitHub runs a container action as root so it can write the workspace and
# $GITHUB_OUTPUT, which the runner owns; the published image's user 65532 cannot.
# Pin PROVEN_IMAGE to a digest (saphan/proven@sha256:...) once make push has
# published the image: docs/ACTION.md, runbook step 1.
ARG PROVEN_IMAGE=saphan/proven:latest
FROM ${PROVEN_IMAGE}
USER root

# reviewdog (MIT) posts Proven's SARIF on the pull request when `comment` is true.
# An exact release, checked against its published checksum; the build refuses an
# empty or wrong checksum. The two values are the sha256 lines for the Linux
# x86_64 and arm64 archives in the release's checksums.txt, and equal the sha256
# of the archives as downloaded (both read 2026-10-25). docs/ACTION.md, runbook
# step 1, says how to move to a newer release.
ARG TARGETARCH
ARG REVIEWDOG_VERSION=0.21.2
ARG REVIEWDOG_SHA256_AMD64=30413aa3c7443e9c3c157fe5766cad40e3bb39a32e210ee69b710a8d5c4b8e51
ARG REVIEWDOG_SHA256_ARM64=56854c078853ea53c73e1e4d627e6a90d6b68185c46edd46aad8beb94564b948
RUN set -eu; \
    case "${TARGETARCH:-amd64}" in \
      amd64) arch=x86_64; sum="${REVIEWDOG_SHA256_AMD64}" ;; \
      arm64) arch=arm64; sum="${REVIEWDOG_SHA256_ARM64}" ;; \
      *) echo "reviewdog: no build for ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    [ -n "$sum" ] || { echo "reviewdog: set REVIEWDOG_SHA256_AMD64 and REVIEWDOG_SHA256_ARM64 in action/Dockerfile to the release checksums (docs/ACTION.md, runbook step 1)" >&2; exit 1; }; \
    file="reviewdog_${REVIEWDOG_VERSION}_Linux_${arch}.tar.gz"; \
    wget -q -O "/tmp/$file" "https://github.com/reviewdog/reviewdog/releases/download/v${REVIEWDOG_VERSION}/$file"; \
    echo "$sum  /tmp/$file" | sha256sum -c -; \
    tar -xzf "/tmp/$file" -C /usr/local/bin reviewdog; \
    rm -f "/tmp/$file"; \
    reviewdog -version

# curl edits the summary comment: BusyBox wget cannot send PATCH (post-comment.sh).
RUN apk add --no-cache curl

COPY entrypoint.sh post-comment.sh /action/
RUN chmod 755 /action/entrypoint.sh && chmod 644 /action/post-comment.sh
ENTRYPOINT ["/action/entrypoint.sh"]
