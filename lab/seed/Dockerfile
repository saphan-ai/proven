# One-shot seed. Versions come from the build-arg. The image runs as a user
# that is not root, and that user owns the state directories, so a new named
# volume starts with that owner and nothing starts the seed as root.
# The default matches RUNTIME_IMAGE in lab/versions.env. make passes the
# setting, and that value wins. An empty default makes BuildKit warn that
# the base name is empty (measured 2026-10-23, line 5 of this file).

ARG RUNTIME_IMAGE=alpine:3.24
FROM ${RUNTIME_IMAGE}
RUN apk add --no-cache ca-certificates curl jq \
 && adduser -D -u 65532 -h /work proven
COPY seed.sh render.sh golive.sh report.sh md2html.awk nightshift-import.sh /usr/local/bin/
# /page is the start page's volume. The page service, another user, reads it,
# so it is 755 where /state and /reg are 700.
RUN chmod 755 /usr/local/bin/seed.sh /usr/local/bin/render.sh /usr/local/bin/golive.sh /usr/local/bin/report.sh /usr/local/bin/md2html.awk /usr/local/bin/nightshift-import.sh \
 && mkdir -p /state /reg /page \
 && chown proven:proven /state /reg /page \
 && chmod 700 /state /reg \
 && chmod 755 /page
USER proven
ENTRYPOINT ["/usr/local/bin/seed.sh"]
