//go:build windows

package cli

import (
	"os"
	"syscall"
)

// linkCount is how many directory names a file has: NumberOfLinks from
// GetFileInformationByHandle. Windows has no other way to tell a second hard
// link from one name, and a log file with a second name must be refused
// (security review piece 8). The handle is opened for
// its attributes only; nothing is read, written, started or dialled, which is
// why this file is a syscall exemption in internal/pipeline/imports_test.go.
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
