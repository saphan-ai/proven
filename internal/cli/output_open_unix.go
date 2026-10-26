//go:build unix

package cli

import (
	"errors"
	"os"
	"syscall"
)

// outputOpenRefusesFinalLink is true where openOutputFile passes O_NOFOLLOW,
// so a symbolic link as the final component fails the open and is not followed.
const outputOpenRefusesFinalLink = true

// openOutputFile opens path for writing and truncates it, creating it when
// it is absent.
//
// WHY O_NOFOLLOW: the load check and the Lstat in writeOutputFile both run
// before this open. A final symbolic link swapped in after the Lstat would
// be followed by a plain create. O_NOFOLLOW makes this open fail with ELOOP
// when the final component is a symbolic link, and the file the link names
// is not created or replaced. The flag applies to the final component only;
// earlier components are still followed. Go 1.26 exposes it as
// syscall.O_NOFOLLOW on a unix build. Measured on linux/amd64 with
// go1.26.10: OpenFile with that flag leaves the target unchanged, and
// without it the target is replaced.
func openOutputFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscall.O_NOFOLLOW, 0o644)
}

// openFailedOnFinalLink reports the error O_NOFOLLOW returns for a final
// symbolic link. The same errno is also a link loop in an earlier component;
// the caller distinguishes them.
func openFailedOnFinalLink(err error) bool {
	return errors.Is(err, syscall.ELOOP)
}
