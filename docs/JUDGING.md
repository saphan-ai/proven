# The images judges run

During the judging window (27 October to 6 November 2026) every image a judge's run pulls from Docker Hub is referenced by
the tag **`:judging`**. A later push to `latest` or to `main` does not change it, so what a judge runs on day ten is what
was published on day one, and anyone can check that.

## Which images

| Image | Tag | Who pulls it |
|---|---|---|
| `saphan/proven` | `judging` | the GitHub Action (`action/Dockerfile` builds on it) and the workflow a GitHub user commits (`lab/workflow/proven.yml`) |
| `saphan/proven-lab-seed` | `judging` | published beside it, from the same scanned build |

The judges' lab builds its own images from the checkout and does not pull either one.

## How the tag is made

`make push-judging` runs `make push` (one build per platform, linux/amd64 and linux/arm64, written to a local layout and
pushed nowhere), scans every platform of both images with Docker Scout, and refuses, publishing nothing, if Scout reports
any critical or high finding, with a fix or without one. Only then does it copy the scanned layouts, unchanged, to `latest`,
the version tag and `:judging`, and read each tag back to check that it carries the digest that was scanned.
`make check-lab-gates` checks all of this offline, and fails when the Action or the committed workflow names any tag but
`:judging`.

The image is built on Alpine 3.24, whose git (2.54.0) is past the version CVE-2025-66413 affects. Docker Scout reported no
vulnerable package in the image built this way (27 October 2026).

## The published digests

| Image | Platforms | Digest of `:judging` | Built from commit | Docker Scout at push |
|---|---|---|---|---|
| `saphan/proven` | linux/amd64, linux/arm64 | added after the push | added after the push | added after the push |
| `saphan/proven-lab-seed` | linux/amd64, linux/arm64 | added after the push | added after the push | added after the push |

## Check it yourself

```bash
docker buildx imagetools inspect saphan/proven:judging --format '{{.Manifest.Digest}}'
docker pull saphan/proven:judging
docker inspect saphan/proven:judging --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
docker scout cves saphan/proven:judging --only-severity critical,high
```

The first line prints the digest in the table above. The third prints the commit the image was built from. The last prints
the scan a judge can repeat.
