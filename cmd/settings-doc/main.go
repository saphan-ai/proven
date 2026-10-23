// Command settings-doc writes the environment tables of docs/CONFIGURATION.md
// from the settings the packages declare.
//
// Run it, or go generate in internal/settingspage, after a setting changes.
// ADR-0013: every setting and its default are listed in one place.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/settingspage"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run rewrites the page and prints its path. It takes no arguments, so it
// always writes the same file.
func run(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "settings-doc writes docs/CONFIGURATION.md and takes no arguments")
		return 2
	}
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	path := filepath.Join(root, "docs", "CONFIGURATION.md")
	body, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	next, err := envdoc.Apply(string(body), envdoc.Render(settingspage.Catalog()))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// An unchanged page is not written again, so its modification time stays.
	if next == string(body) {
		fmt.Println(path)
		return 0
	}
	// The page keeps its own permissions.
	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(path); statErr == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(path, []byte(next), mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(path)
	return 0
}

// moduleRoot walks up from the working directory to go.mod, so the command
// works from any directory of the module.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	start := dir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		dir = parent
	}
}
