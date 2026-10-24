//go:build windows

package cli

import (
	"os"
	"syscall"
)

// linkCount is how many directory names a file has.
// The file data Lstat returns on Windows has no count. Opening the file
// and reading ByHandleFileInformation.NumberOfLinks does. The open is
// not a follow of a symlink: the caller has already seen a regular file.
func linkCount(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var info syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(syscall.Handle(f.Fd()), &info); err != nil {
		return 0, err
	}
	return uint64(info.NumberOfLinks), nil
}
