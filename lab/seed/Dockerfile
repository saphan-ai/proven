# One-shot seed. Versions come from the build-arg. The process drops to a
# user that is not root after it prepares the state directories.
# The default matches RUNTIME_IMAGE in lab/versions.env. make passes the
# setting, and that value wins. An empty default makes BuildKit warn that
# the base name is empty (measured 2026-10-23, line 5 of this file).

ARG RUNTIME_IMAGE=alpine:3.22
FROM ${RUNTIME_IMAGE}
RUN apk add --no-cache ca-certificates curl su-exec \
 && adduser -D -u 65532 -h /work proven
COPY seed.sh /usr/local/bin/seed.sh
RUN chmod 755 /usr/local/bin/seed.sh
USER proven
ENTRYPOINT ["/usr/local/bin/seed.sh"]
