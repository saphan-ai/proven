# Go + Python build image for proven's executors.
# Environment, not product code: written 2026-10-20, before the build window, by the owner's decision D2.
#
# BASE is our Go image (Debian 13, Go 1.26.8, Node 22), pinned by digest. That image is not published, so a stranger
# cannot build this file as it stands: replace BASE with any Debian 13 image that carries Go 1.26.8 and Node 22.
ARG BASE
FROM ${BASE}

USER root
# One layer: Python 3.13 (Debian 13's own minor, the same as the native lane), venv and pip; a C compiler and libc
# headers for cgo (the base image has none).
RUN apt-get update \
 && apt-get install -y --no-install-recommends python3 python3-venv python3-pip gcc libc6-dev \
 && rm -rf /var/lib/apt/lists/*
ENV CGO_ENABLED=1
# Node 24 (the owner, 2026-10-21): the Cursor agent runs its legs with its own bundled Node 24.5.0, so the image carries the
# same major instead of the base image's Node 22. Copied from the official image, pinned by digest.
COPY --from=docker.io/library/node@sha256:1a9c44e9441b054d8360d14f12bd24e0da3112b5bcf2db4ecd16e1b008387e69 /usr/local/bin/node /usr/local/bin/node
COPY --from=docker.io/library/node@sha256:1a9c44e9441b054d8360d14f12bd24e0da3112b5bcf2db4ecd16e1b008387e69 /usr/local/lib/node_modules /usr/local/lib/node_modules
USER 65534
