// Notice configuration for a chat service, before field validation.
//
// token_env is the name field: it holds the name of the environment variable
// that carries the API token, never the token itself. The token belongs in
// the credentials file (home/credentials). show prints the fields only.
//
// This version stores any value in any field.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	home, args := parseArgs(os.Args[1:])
	if home == "" || len(args) == 0 {
		usage()
		os.Exit(2)
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	switch args[0] {
	case "set":
		if len(args) != 3 {
			usage()
			os.Exit(2)
		}
		msg, code := setField(home, args[1], args[2])
		if code != 0 {
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(code)
		}
		fmt.Println(msg)
	case "show":
		if len(args) != 1 {
			usage()
			os.Exit(2)
		}
		fmt.Println(showText(home))
	default:
		usage()
		os.Exit(2)
	}
}

func parseArgs(args []string) (string, []string) {
	if len(args) >= 2 && args[0] == "--home" {
		return args[1], args[2:]
	}
	return "", args
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: noticecfg --home DIR set FIELD VALUE")
	fmt.Fprintln(os.Stderr, "       noticecfg --home DIR show")
}

func configPath(home string) string {
	return filepath.Join(home, "config")
}

func load(home string) map[string]string {
	cfg := map[string]string{}
	b, err := os.ReadFile(configPath(home))
	if err != nil {
		return cfg
	}
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || k == "" {
			continue
		}
		cfg[k] = v
	}
	return cfg
}

func store(home string, cfg map[string]string) error {
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, cfg[k])
	}
	return os.WriteFile(configPath(home), []byte(b.String()), 0o600)
}

func setField(home, field, value string) (string, int) {
	cfg := load(home)
	cfg[field] = value
	if err := store(home, cfg); err != nil {
		return err.Error(), 2
	}
	return "accepted", 0
}

func showText(home string) string {
	cfg := load(home)
	if len(cfg) == 0 {
		return "(empty)"
	}
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%s=%s", k, cfg[k])
	}
	return b.String()
}
