# Local and offline operation

Set `PROVEN_NETWORK=local` to enforce the local network policy. Remote hosted modes are refused. Configure an allowed local model API, or disable models, and use a local SQLite snapshot. PostgreSQL and CouchDB must also be inside the configured network.

`PROVEN_NETWORK_ALLOW` lists additional allowed host names. Their resolved addresses are checked by the same policy. HTTP clients do not follow a redirect or use an external proxy to bypass that boundary. Git reads cannot fetch missing objects.

The external CLI model transport is refused in this mode: the HTTP policy cannot constrain a CLI's own network access. Use an allowed local API instead.

Prepare runner images, toolchains, dependencies and vulnerability data before going offline. Generated tests still run with no network or inherited secrets. A missing dependency or source is incomplete analysis, never a clean measurement.

The recorded sample works without a model server or container runtime:

```sh
make sample
```

For real changes, import your data and select a model or deterministic analysis through [configuration](CONFIGURATION.md). See [runner controls](RUNNER.md) and [database snapshots](DATABASES.md).
