# One-shot seed. Versions come from the build-arg. The image runs as a user
# that is not root, and that user owns the state directories, so a new named
# volume starts with that owner and nothing starts the seed as root.
# The default matches RUNTIME_IMAGE in lab/versions.env. make passes the
# setting, and that value wins. An empty default makes BuildKit warn that
# the base name is empty (measured 2026-10-23, line 5 of this file).

ARG RUNTIME_IMAGE=alpine:3.23
FROM ${RUNTIME_IMAGE}
RUN apk add --no-cache ca-certificates curl \
 && adduser -D -u 65532 -h /work proven
COPY seed.sh /usr/local/bin/seed.sh
RUN chmod 755 /usr/local/bin/seed.sh \
 && mkdir -p /state /reg \
 && chown proven:proven /state /reg \
 && chmod 700 /state /reg
USER proven
ENTRYPOINT ["/usr/local/bin/seed.sh"]
