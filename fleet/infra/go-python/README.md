# fleet/infra/go-python

The build image for proven's executors: Go and Python in one container, with Node and a C compiler for cgo.

**Environment, not product code.** The owner decided on 2026-10-20 (decision D2 of the train `proven-cursor-infra`) that
this file lands before the build window opens on 22 October, dated, and marked as environment.

| tool | version | where it comes from |
|---|---|---|
| Go | 1.26.8 | the base image |
| Node | 22 | the base image |
| Python | 3.13 | Debian 13, this file's one layer |
| gcc, libc headers | Debian 13 | this file's one layer; `CGO_ENABLED=1` |

The native lane on runner-cursor carries the same Go, Python and Node versions, so a reviewer can tell a code fault from
a lane fault.

**The base image is not published.** `BASE` is our Go image, pinned by digest. A stranger cannot build this file as it
stands: replace `BASE` with any Debian 13 image that carries Go 1.26.8 and Node 22.

Build (on runner-cursor):

```
docker build --build-arg BASE=<our Go image>@sha256:<digest> -t proven-go-python fleet/infra/go-python
docker run --rm proven-go-python sh -c 'go version; python3 --version; node --version; gcc --version | head -1; go env CGO_ENABLED'
```

The digest of the built image is recorded in `fleet/status/STATUS_proven-cursor-infra.md`.
