# Scanning the CLIs we add to Proven

**Measured on 16 November 2026, Linux arm64.** Scanning customer-built CLI images found vulnerable components inside Qwen Code's dependency tree and Cursor's bundled Node runtime. We updated those components in the customer recipes and repeated the scans. The three tested images then reported **0 HIGH / 0 CRITICAL**.

These were local development builds, without authenticated accounts. No vendor-CLI image was published by Proven. The official Proven product remains CLI-free; users build and control their own vendor layers.

## What the scans found

| Customer CLI | Pinned CLI version | Reported component | Before | Change | After |
| --- | --- | --- | --- | --- | --- |
| OpenAI Codex | `0.169.0` | No HIGH/CRITICAL component reported | 0 HIGH, 0 CRITICAL | No dependency exception needed | 0 HIGH, 0 CRITICAL |
| Qwen Code | `0.27.4` | Optional dependency `sharp 0.35.4` | 1 HIGH, 0 CRITICAL | Explicit npm override to `sharp 0.35.5` | 0 HIGH, 0 CRITICAL |
| Cursor | `2026.11.13-eacade1` | Node `24.22.1` inside the vendor archive | 7 HIGH, 0 CRITICAL | Replace the bundled runtime with official Node `24.22.3`, keeping Node major 24 | 0 HIGH, 0 CRITICAL |

The scope was the final image filesystem, including vendor-added packages. The filter included HIGH and CRITICAL findings, with suppressed findings still included. We did not hide advisories or reinterpret a failing scan as success.

Captured outputs: [Codex baseline](audits/2026-11-16/codex-baseline.txt), [Qwen before](audits/2026-11-16/qwen-before.txt), [Qwen after](audits/2026-11-16/qwen-after.txt), [Cursor before](audits/2026-11-16/cursor-before.txt), [Cursor after](audits/2026-11-16/cursor-after.txt). Only terminal formatting, trailing whitespace and the machine's local archive directory were normalized.

## Qwen Code: the CLI version did not pin a patched dependency

Qwen Code `0.27.4` declared `sharp 0.35.4` as an optional dependency. Docker Scout matched it to **GHSA-wq5f-xc86-pv6w / CVE-2026-96889**. The advisory identifies `sharp <0.35.5` and a vulnerability in its upstream librsvg dependency; `0.35.5` is the patched package. [GitHub's reviewed advisory](https://github.com/advisories/GHSA-wq5f-xc86-pv6w).

The [customer Dockerfile](../examples/cli/Dockerfile) now creates an explicit npm override before installation:

```json
{"private":true,"overrides":{"sharp":"0.35.5"}}
```

The CLI remains pinned to `0.27.4`. `SHARP_VERSION` is a separate build argument, also declared in the [build configuration](../examples/cli/docker-bake.hcl). The rebuilt image still passed `qwen --version`, and its repeated HIGH/CRITICAL scan returned zero findings.

The advisory's described RCE conditions concern glibc-based Linux. Our Qwen image uses Alpine/musl. The scanner result establishes package presence and an advisory match; we did **not** demonstrate an exploit or establish those runtime conditions in this image. We updated the dependency rather than rely on a platform-specific dismissal.

## Cursor: an archive brought its own older Node

The pinned Cursor archive contained its own Node `24.22.1`. Updating the operating-system packages did not replace that executable. Docker Scout reported seven HIGH advisories against it:

| Advisory | Reported area |
| --- | --- |
| `CVE-2026-123863` | HTTP/2 header lifetime. |
| `CVE-2026-123864` | Reentrant TLS/HTTP parser execution. |
| `CVE-2026-123865` | Abandoned HTTP/2 streams. |
| `CVE-2026-123866` | Buffer lifetime during socket writes. |
| `CVE-2026-123872` | TLS session/SNI policy. |
| `CVE-2026-123874` | HTTP/2 stream lifetime during writes. |
| `CVE-2026-126051` | TLS server crash. |

Node's **11 November 2026 security release** lists these fixes in `24.22.3`. Several affect server behavior; their presence in a CLI does not establish that its model-connection path exposes them. [Official Node release notes](https://nodejs.org/en/blog/release/v24.22.3).

The [Cursor recipe](../examples/cli/Dockerfile.cursor) now:

1. Downloads a version-pinned Cursor archive and checks its SHA-256 for the selected architecture.
2. Removes the bundled Node in the installation stage.
3. Copies the patched Node binary from `node:24.22.3-bookworm-slim` into the fresh runtime image.
4. Runs the vendor's version command as the non-root user.

The Cursor code stays at the same version, and Node stays on the same major version. This limits the compatibility change; it does not replace authenticated functional testing. The rebuilt image passed the version check and reported zero HIGH/CRITICAL findings.

Cursor also needed a different base from our other CLI recipes. Its bundled Linux executable failed on Alpine with `fcntl64: symbol not found`, including with `gcompat`. Its separate customer recipe therefore uses Ubuntu/glibc. The official Proven image and the Claude, Copilot, Codex and Qwen customer recipes retain Alpine, with the explicit Git/musl security packages described in [container builds](CONTAINERS.md).

## How the results were checked

Each image was exported for the **explicit local platform**, then scanned from that archive with Docker Scout `1.23.1`:

```sh
docker image save --platform linux/arm64 -o proven-qwen-arm64.tar proven-qwen:local
docker scout cves archive://proven-qwen-arm64.tar \
  --only-severity critical,high --ignore-suppressed=false --exit-code
sha256sum proven-qwen-arm64.tar
```

Repeat for the other images. On macOS, `shasum -a 256` supplies the archive checksum. The explicit archive avoids attributing a scan of another platform to the image being tested. Docker Scout remains a service-backed scanner; a local archive input is not a claim that image metadata stays offline.

### Retained archive identities

These SHA-256 values identify the exact local archives measured, not a promise that a later rebuild has identical bytes. The archives themselves are not committed or published.

| Measurement | Archive SHA-256 |
| --- | --- |
| Codex baseline | `b57bad36aa205a7033089e0b11e7815b06783e067f3dbe731b03a354a68ad6f3` |
| Qwen before | `51155a340ecbe66064fb5debb46e2abe3924eb96f7ffd0ca29517a3c0fd0ea6c` |
| Qwen after | `389f84164a2f8b04707b022a8883b0996c1906d703a3b6918a6025bc3a78a939` |
| Cursor before | `0aaca638f0a3dc45222783dc21b4927f8e71846ef44be784d078fde81f5a2860` |
| Cursor after | `a25c04a83460bb6d11ad9d300bcaed145fe8ba3521f06d60a29b1ce3671cd8d2` |

Docker reported these final local image identities:

```text
Codex  sha256:c8cf233ab7eeeeb385f192f869171ead5170f875cf74997036684820f1f76efd
Qwen   sha256:353c3a75af98da57ea782dc604d9f14d5877aad34ad2914c7fed59d0388ec0b5
Cursor sha256:4c80a74e20493187923b32a9d437f594a937667c922bbb9a1497e3c78bf363e6
```

## What this establishes

The selected archives had no HIGH/CRITICAL findings in the scanner's database at measurement time, after the explicit fixes. Build-time version checks and account-free adapter tests exercised installation and invocation. They did not establish authenticated model behavior, exploit reachability, absence of lower-severity issues or permanent safety.

These measurements cover **Linux arm64 only**. An amd64 image, another CLI release, updated base tag or changed dependency resolution needs its own checks. For retained builds, pin base images by reviewed digest as well as pinning the vendor package/version.

The useful operational rule is to scan the **finished customer image**. A vendor version pin, a valid download checksum and a clean base-image scan each answer a different question. None substitutes for inspecting the dependency tree and runtime that the final CLI layer actually added.

The project's [publication gate](../scripts/publish-image.sh) scans both product-image platforms and refuses vendor CLI executables in the official image. Customer recipes remain a separate build and verification responsibility. [Start with Proven](START.md) explains that ownership without requiring users to assemble the model transport by hand.
