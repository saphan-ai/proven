//go:build unix

package cli

import (
	"fmt"
	"os"
	"syscall"
)

// linkCount is how many directory names a file has.
// On Linux and Darwin that number is Stat_t.Nlink. The field is uint64 on
// Linux and uint16 on Darwin; both convert to the same count.
func linkCount(path string) (uint64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st == nil {
		return 0, fmt.Errorf("file status has no link count")
	}
	return uint64(st.Nlink), nil
}
