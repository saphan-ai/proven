//go:build windows

package cli

// linkCount reports one name for every file.
// NumberOfLinks comes from GetFileInformationByHandle, in package syscall.
// In Go 1.26.10 that call is GetFileInformationByHandle(handle, *ByHandleFileInformation).
// A production file under internal/ must not import syscall, so the count
// is not read. One means an existing trace file is not refused. A second
// name on Windows is not distinguished from one name.
func linkCount(string) (uint64, error) {
	return 1, nil
}
