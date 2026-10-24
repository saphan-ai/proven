package main

import (
	"fmt"
	"os"
	"strconv"
)

// Usage:
//
//	guard check <pinned> <note>...   prints PASS, or FAIL and one stale note per line
//	guard describe                   prints the page's description of the guard
func main() {
	if len(os.Args) >= 2 && os.Args[1] == "describe" {
		fmt.Println(Describe())
		return
	}
	if len(os.Args) >= 3 && os.Args[1] == "check" {
		pinned, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bad pinned count:", err)
			os.Exit(2)
		}
		ok, stale := Check(os.Args[3:], pinned)
		if ok {
			fmt.Println("PASS")
			return
		}
		fmt.Println("FAIL")
		for _, s := range stale {
			fmt.Println("stale:", s)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "usage: guard check <pinned> <note>... | guard describe")
	os.Exit(2)
}
