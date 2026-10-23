// Command proven reviews a directory or a pull request.
//
// main is deliberately thin: everything it does is in internal/cli, so the
// tests can run the whole command line in-process with their own
// environment, output streams and reviewer, and the binary adds nothing a
// test cannot see.
package main

import (
	"os"

	"github.com/saphan-ai/proven/internal/cli"
)

// main passes the real arguments, environment and streams to cli.Run and
// exits with the code it returns (ADR-0020 §2). The nil opener means the
// reviewer comes from review.Open, as in every production run.
func main() {
	os.Exit(cli.Run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr, nil))
}
