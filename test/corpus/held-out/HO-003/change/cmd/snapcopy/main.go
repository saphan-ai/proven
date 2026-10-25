// Command snapcopy copies the tree at the first argument to the second, then
// prints one line per entry of the copy: its path and its mode.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"ho003change/snapshot"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: snapcopy SRC DST")
		os.Exit(2)
	}
	src, dst := os.Args[1], os.Args[2]
	if err := snapshot.Copy(src, dst); err != nil {
		fmt.Fprintln(os.Stderr, "copy:", err)
		os.Exit(1)
	}
	err := filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			return err
		}
		fmt.Printf("%s %04o\n", rel, info.Mode().Perm())
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "walk:", err)
		os.Exit(1)
	}
}
