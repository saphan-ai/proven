package cli

import (
	"fmt"
	"io"

	"github.com/saphan-ai/proven/internal/network"
)

// runCheckAddress is `proven check-address SETTING ADDRESS`: the action asks it
// whether the forge address reviewdog would post to may be reached under the
// network mode (ADR-0034). In public mode every address is allowed. In local
// mode the address is judged like every other: loopback, private, or on the
// allow list, and an IP literal, because reviewdog resolves a name with its
// own resolver. Exit 0 allowed, and the line ends with the mode; 7 refused,
// naming the setting; 2 for a wrong command line.
func runCheckAddress(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "check-address takes a setting name and an address: proven check-address SETTING ADDRESS")
		return exitFatal
	}
	pol, err := network.Load(getenv)
	if err != nil {
		ce, _ := network.As(err)
		fmt.Fprintln(stderr, visible(ce.Error()))
		return exitConfig
	}
	if err := pol.CheckProgramAddress(args[0], args[1]); err != nil {
		ce, _ := network.As(err)
		fmt.Fprintln(stderr, visible(ce.Error()))
		return exitConfig
	}
	// The last word is the mode Proven loaded, so the action decides its own
	// local-mode steps by Proven's reading of PROVEN_NETWORK, never its own.
	mode := network.ModePublic
	if pol.Local() {
		mode = network.ModeLocal
	}
	fmt.Fprintf(stdout, "%s is allowed under the network mode: %s\n", args[0], mode)
	return exitOK
}
