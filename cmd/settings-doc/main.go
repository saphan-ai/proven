// Command settings-doc writes the environment tables of docs/CONFIGURATION.md
// from the settings the packages declare. A table envdoc.Check rejects is
// refused, and the page is not written.
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
	next, err := applyCatalog(settingspage.Catalog(), string(body))
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

// applyCatalog renders sections into body. Check runs first: a pipe or a
// newline in a cell would otherwise be written, and only a later test would
// notice. A refused catalog does not change the page.
func applyCatalog(sections []envdoc.Section, body string) (string, error) {
	if err := envdoc.Check(sections); err != nil {
		return "", err
	}
	return envdoc.Apply(body, envdoc.Render(sections))
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
