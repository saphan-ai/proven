# Reference set: Dependabot's alerts on this repository at fbdf7d3

GitHub Dependabot raised 51 alerts on this repository: 45 npm advisories in `docs-site/package-lock.json` and 6 Go
advisories for `golang.org/x/net` in `demo/live/change/go.mod`. We found them on our own repository and turned them into
a baseline for Proven: Proven must find the same vulnerable dependencies, with its own evidence, and the comparison is made
alert by alert.

`alerts.json` is frozen before any run. It names the commit the alerts belong to
(`fbdf7d32f57347fd89ae747ba0fe9faa28647414`, `main` when the alerts were read) and keeps each alert's number, severity,
GHSA and CVE ids, package, manifest, vulnerable range and first patched version, as the API returned them on 2026-10-26.
The npm alerts were fixed after that commit, so they are `fixed` here; the reference is the pinned commit, which git keeps.

The 6 Go alerts are on purpose: `demo/live/change` pins `golang.org/x/net v0.16.0`, the release before the HTTP/2
rapid-reset fix (CVE-2023-44487), so the live demo can show a known-exploited dependency being caught
(`demo/live/README.md`).

`proven-findings.json` is what Proven's dependency check kept when it ran over the manifests of the pinned commit, and
`REPORT.md` is the alert-by-alert comparison built from the two JSON files. Both are written by `make baseline-dependabot`
(or `make baseline-dependabot-manifests` for a run with no evidence store) and are never edited by hand. How to run it, what
evidence a run rests on and what the committed report shows are in [section 9 of the test README](../../../README.md).

## What the live run shows (2026-10-26, against our Zetaris)

Proven found all 6 Go alerts with its own evidence, and 15 more advisories for `golang.org/x/net@v0.16.0` that Dependabot
did not raise. The 45 npm alerts are `not checked`, not `missed`: Proven read every alerted npm package and version from
the lock file, but the live OSV table held no npm advisories at all when the run was made (read-only count on 2026-10-26
~08:10Z: Go 96 rows, Maven 20, SwiftURL 1, npm 0). A verdict of `missed` is given only where the store answered for the
alert's ecosystem and still named no row for its advisory. A KEV mark belongs to the advisory whose own CVE the KEV view
holds (here CVE-2023-44487), never to every advisory of the same version.
