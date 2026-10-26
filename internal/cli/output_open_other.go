//go:build !unix

package cli

import "os"

// outputOpenRefusesFinalLink is false where the open has no O_NOFOLLOW.
// writeOutputFile's Lstat is then the only guard, and a final symbolic link
// swapped in after that Lstat is followed.
const outputOpenRefusesFinalLink = false

// openOutputFile opens path for writing and truncates it. This system has
// no open flag that refuses a final symbolic link, so a link at path is
// followed.
func openOutputFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
}

// openFailedOnFinalLink is never true here: the open does not report a
// final symbolic link as its own error.
func openFailedOnFinalLink(error) bool { return false }
