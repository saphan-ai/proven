// Command proven reviews a directory or a pull request.
package main

import (
	"os"

	"github.com/saphan-ai/proven/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr, nil))
}
