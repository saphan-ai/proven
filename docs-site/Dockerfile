# Proven's documentation site: one image, built from the repository root.
#
#   docker build -f docs-site/Dockerfile -t proven-docs .
#   docker run -p 3000:3000 proven-docs
#
# The build context is the repository root: the site needs README.md, docs/ and
# the Go code its generated pages are written from. The build FAILS when a page
# is broken, names something internal, or is out of date with the code (the
# guards in docs-site/scripts/guards.mjs).

# The default matches GOLANG_IMAGE in lab/versions.env: the module's Go version.
ARG GOLANG_IMAGE=golang:1.26.10
ARG NODE_IMAGE=node:22-alpine

# generate: the pages that are written from the code, never by hand.
FROM ${GOLANG_IMAGE} AS generate
WORKDIR /repo
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY docs/CONFIGURATION.md ./docs/CONFIGURATION.md
COPY docs-site/scripts/generate.sh ./docs-site/scripts/generate.sh
RUN sh docs-site/scripts/generate.sh /gen

# deps: npm ci refuses a package.json and a lockfile that disagree, on purpose.
# The scripts are skipped here; next build writes .source itself.
FROM ${NODE_IMAGE} AS deps
WORKDIR /repo/docs-site
COPY docs-site/package.json docs-site/package-lock.json ./
RUN npm ci --ignore-scripts

# build: the guards run first under set -e, so a failing guard stops the image.
FROM ${NODE_IMAGE} AS build
ENV NEXT_TELEMETRY_DISABLED=1
WORKDIR /repo
# The whole repository, not only docs/: a page links to files across it (ADRs to
# hackathon/, docs to lab/ and test/), and the link guard can only tell a real
# file from a broken link when the file is here. docs-site/Dockerfile.dockerignore
# replaces the root .dockerignore for this build, so those directories are in the
# context; build output and secrets stay out.
RUN --mount=type=bind,target=/ctx \
    cd /ctx && tar --exclude=./docs-site -cf - . | tar -xf - -C /repo
COPY --from=deps /repo/docs-site/node_modules ./docs-site/node_modules
COPY docs-site ./docs-site
COPY --from=generate /gen /gen
WORKDIR /repo/docs-site
ENV GENERATED_FRESH_DIR=/gen
RUN set -e; npm test; npm run build

# run: only the standalone server and its static files.
FROM ${NODE_IMAGE} AS run
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
WORKDIR /app
RUN addgroup -g 10001 -S docs && adduser -S -u 10001 -G docs docs
COPY --from=build --chown=10001:10001 /repo/docs-site/.next/standalone ./
COPY --from=build --chown=10001:10001 /repo/docs-site/.next/static ./.next/static
COPY --from=build --chown=10001:10001 /repo/docs-site/public ./public
USER 10001:10001
EXPOSE 3000
# a real page: / redirects, so it would not say the site is serving pages
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:3000/docs/reference/cli || exit 1
CMD ["node", "server.js"]
