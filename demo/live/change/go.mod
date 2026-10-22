module example.com/share

go 1.26

// v0.16.0 is the release before the HTTP/2 rapid-reset fix in v0.17.0.
// CVE-2023-44487. The package is not imported, so the build stays offline.
require golang.org/x/net v0.16.0
