// Command osv-to-csv reads one OSV JSON advisory per file and writes the flat range CSV.
//
// It builds the OSV package table that the review's dependency check
// queries. OSV by package cannot be a Zetaris REST table (ADR-0023), so an
// administrator runs this on a directory of OSV advisories and puts the
// CSV in Zetaris's storage. The columns are described in docs/OSV_TABLE.md.
//
// Usage:
//
//	osv-to-csv <advisory-dir> <out.csv>
//
// <out.csv> may be - to write to stdout. A usage error or a file that
// cannot be read or written exits with status 2.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/saphan-ai/proven/internal/osv"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main without the process: it returns the exit status, so a test
// can drive it.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] == "" || args[1] == "" || stringsHasFlag(args) {
		fmt.Fprintln(stderr, "usage: osv-to-csv <advisory-dir> <out.csv>")
		return 2
	}
	rows, err := osv.ReadDir(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	// The whole CSV is built before anything is written, so a failure leaves
	// no half-written table.
	var buf bytes.Buffer
	if err := osv.WriteCSV(&buf, rows); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	if args[1] == "-" {
		if _, err := stdout.Write(buf.Bytes()); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 2
		}
		return 0
	}
	if err := os.WriteFile(args[1], buf.Bytes(), 0o644); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	return 0
}

// stringsHasFlag reports whether an argument looks like a flag. The command
// takes none, and "-" alone means stdout, not a flag.
func stringsHasFlag(args []string) bool {
	for _, a := range args {
		if len(a) > 0 && a[0] == '-' && a != "-" {
			return true
		}
	}
	return false
}
