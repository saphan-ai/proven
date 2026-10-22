// Command osv-to-csv reads one OSV JSON advisory per file and writes the flat range CSV.
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

func stringsHasFlag(args []string) bool {
	for _, a := range args {
		if len(a) > 0 && a[0] == '-' && a != "-" {
			return true
		}
	}
	return false
}
