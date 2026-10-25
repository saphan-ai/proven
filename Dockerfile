# Proven's image: a static binary, git, and the container client, on Alpine.
# The defaults match GOLANG_IMAGE and RUNTIME_IMAGE in lab/versions.env.
# make passes the settings, and those values win. An empty default makes
# BuildKit warn that the base name is empty.

ARG GOLANG_IMAGE=golang:1.26.10
ARG RUNTIME_IMAGE=alpine:3.22
FROM ${GOLANG_IMAGE} AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/proven ./cmd/proven \
 && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/livecheck ./cmd/livecheck

FROM ${RUNTIME_IMAGE}
ARG CREATED=unset
ARG REVISION=unset
ARG VERSION=unset
LABEL org.opencontainers.image.created="${CREATED}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.title="proven" \
      org.opencontainers.image.source="https://github.com/saphan-ai/proven"
RUN apk add --no-cache ca-certificates git docker-cli \
 && adduser -D -u 65532 -h /work proven \
 && mkdir -p /opt/proven/fixture \
 && chown -R proven:proven /work /opt/proven
COPY --from=build /out/proven /usr/local/bin/proven
COPY --from=build /out/livecheck /usr/local/bin/livecheck
COPY --chown=proven:proven testdata/fixture/discount /opt/proven/fixture/discount
USER proven
WORKDIR /work
ENTRYPOINT ["proven"]
