//go:build windows

package cli

// linkCount refuses an existing regular trace file.
// The count is NumberOfLinks from GetFileInformationByHandle. That call is
// in package syscall, and a production file under internal/ must not import
// syscall, so the count is not read. Two is a second name: the load refuses
// the file with the same exit and the same sentence as a real hard link.
// A trace that already exists is refused on Windows, including a file that
// has only one name.
func linkCount(string) (uint64, error) {
	return 2, nil
}
