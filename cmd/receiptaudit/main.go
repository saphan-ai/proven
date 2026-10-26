// Command receiptaudit counts the model answers of Proven's live reviews that
// are in custody: the answer's Nightshift receipt verifies under a pinned
// Ed25519 key, the receipt's request id is the answer's response id, and the
// model that answered is the model requested.
//
//	receiptaudit -receipts receipts.jsonl -pubkey <base64 key> trace.jsonl [trace.jsonl ...]
//
// It prints the numbers per UTC day and every answer that fails, with its
// reason. It exits 1 when any answer fails and 2 when it could not run. It
// opens only the files it is given and dials nothing.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/saphan-ai/proven/internal/custody"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main with its streams as arguments, so a test drives it in-process.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("receiptaudit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	receipts := fs.String("receipts", "", "the receipts.jsonl file Nightshift wrote")
	pubkey := fs.String("pubkey", "", "the pinned Ed25519 public key, base64")
	pubkeyFile := fs.String("pubkey-file", "", "a file holding the pinned public key, base64")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *receipts == "" || (*pubkey == "") == (*pubkeyFile == "") || fs.NArg() == 0 {
		fmt.Fprintln(stderr, "usage: receiptaudit -receipts FILE (-pubkey BASE64 | -pubkey-file FILE) TRACE [TRACE ...]")
		return 2
	}
	key := *pubkey
	if *pubkeyFile != "" {
		b, err := os.ReadFile(*pubkeyFile)
		if err != nil {
			fmt.Fprintf(stderr, "receiptaudit: %v\n", err)
			return 2
		}
		key = string(b)
	}
	rep, err := custody.AuditFiles(*receipts, key, fs.Args())
	if err != nil {
		fmt.Fprintf(stderr, "receiptaudit: %v\n", err)
		return 2
	}
	custody.Write(stdout, rep)
	if rep.Failed() > 0 {
		return 1
	}
	return 0
}
