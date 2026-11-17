# Container builds and security updates

The product image uses Alpine as its base and installs its runtime packages during the build. A vanilla base image is not the final security state of the product. Operator model keys and evidence databases are not baked into the image. The runtime includes the CLI and a recorded demonstration fixture; the build context is allow-listed.

## Runtime package overlay

Vendor CLIs are absent from the published product. [Customer CLI images](CLI_IMAGES.md) are separate recipes users build themselves; account state arrives at runtime. Their added Node/vendor packages require their own security checks; npm is confined to the installation stage.

The product publication script additionally refuses an image exposing `claude`, `copilot`, `codex`, `agent`, `cursor-agent` or `qwen` on either architecture. Customer build targets are separate and are never selected by the product release script.

The Dockerfile first applies available stable Alpine updates, then installs selected signed packages from a tagged security repository. Build arguments explicitly select the repository and exact package versions:

| Argument | Current default |
| --- | --- |
| `RUNTIME_IMAGE` | `alpine:3.24` |
| `ALPINE_SECURITY_REPOSITORY` | `https://dl-cdn.alpinelinux.org/alpine/edge/main` |
| `GIT_PACKAGE_VERSION` | `2.56.1-r0` |
| `MUSL_PACKAGE_VERSION` | `1.2.6-r6` for both `musl` and `musl-utils` |
| `SQLITE_PACKAGE_VERSION` | `3.54.1-r0` for `sqlite-libs` in `Dockerfile.action` only; Python's runtime dependency. |

The tag limits selection of the newer branch to the requested packages and necessary dependencies. This is a curated overlay, not a wholesale distribution upgrade. Package signatures remain enforced. Missing pinned versions or version mismatches fail the build; there is no fallback to older base packages. The runtime dependency set is recorded in the SBOM and must be tested and scanned for each architecture.

The newer pins were selected after the previous image scan reported CVE-2026-105437 in git and CVE-2026-107891 in musl. Distribution backports and scanner version matching can differ; consult the [Git security announcement](https://github.blog/open-source/git/git-security-vulnerabilities-announced-7/) and [Alpine musl tracker](https://security.alpinelinux.org/vuln/CVE-2026-107891) when changing these pins. The release gate uses the findings of a fresh scan of the finished image, including base packages, without a suppression or ignore-base shortcut.

The Docker client is separately downloaded from Docker's static release and checked against a pinned architecture-specific SHA-256. Updating a package version also requires reviewing its checksum where applicable, rebuilding, exercising Git and the sample review, and rescanning both architectures.

The generic product image does not install Python or a native SQLite package: Proven's SQLite adapter is compiled Go code. The GitHub Action image adds Python for read-only comment collection and explicitly updates its SQLite library. It verifies the library works and keeps the `sqlite3` command-line shell absent. See the [17 November image check](audits/2026-11-17/IMAGE_CHECK.md) for measured scan results and the initial package match.

## Publishing

Use a unique source-revision tag. `make publish-image IMAGE=your-registry/proven:your-tag` builds and loads both supported Linux architectures, checks the executable, recorded example and actual offline Git/SQLite model-free flow, scans an explicit single-platform archive of each with Docker Scout, and pushes only after every check passes. The archives avoid local multi-platform selection differences between Scout versions. It requires a Buildx builder supporting both platforms, a containerd image store capable of loading a multi-platform image, and `docker image save --platform`. `BUILDX_BUILDER` can select a prepared builder.

Docker Scout can send image metadata to its service; invoking this publishing command opts into that scan. A scanner error or any reported HIGH/CRITICAL finding stops publication. The gate also reads the SARIF result list and rejects suppressed findings even if the scanner exits successfully. Fix packages and rebuild rather than skipping that gate. Ordinary local development can still use `make image`; it does not publish or imply a successful security scan. `make scan` is also available for a local Trivy check. The publishing script uses Python 3 to check SARIF.

Record the source revision, final image index digest, platform digests, package versions, scanner results and scan date. SBOM and provenance describe the build; they do not replace the scan. An image clean against today's scanner database can acquire findings later and needs a new review before the next release.

In self-review, update `PROVEN_REVIEW_IMAGE` only to the digest of the validated image. Keep the previous digest available for reproducibility. Runner images and host Git are separate components and need their own updates; updating the product image does not patch the GitHub runner or another operator's host.
