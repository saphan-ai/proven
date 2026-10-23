# One-shot seed. Versions come from the build-arg. The process drops to a
# user that is not root after it prepares the state directories.

ARG RUNTIME_IMAGE
FROM ${RUNTIME_IMAGE}
RUN apk add --no-cache ca-certificates curl su-exec \
 && adduser -D -u 65532 -h /work proven
COPY seed.sh /usr/local/bin/seed.sh
RUN chmod 755 /usr/local/bin/seed.sh
USER proven
ENTRYPOINT ["/usr/local/bin/seed.sh"]
