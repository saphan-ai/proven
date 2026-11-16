# Restricted runner

The default runtime is `docker`, using image `golang:1.26`, with `1` CPU and `512m` memory. Each side has a 2 minutes time limit. Removal has a separate limit of 2 seconds, within an overall margin of 3 seconds. Failed cleanup names the container that may remain.

Each checkout is mounted read-only. No network or inherited secrets are available. Toolchains and dependencies must already be present in the prepared image. The configured runtime and images are operator-trusted dependencies.
