// Notice configuration for a chat service, with field validation.
//
// token_env is the name field: it holds the name of the environment variable
// that carries the API token, never the token itself. The token belongs in
// the credentials file (home/credentials). show prints the fields only.
//
// A bad value is refused when the field is set. The refusal message quotes
// the rejected value in full. The control writes a synthetic token into the
// credentials file, then checks set messages and show output.
package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const syntheticToken = "tok_synthetic_0000_not_a_secret"

var envName = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
var channelName = regexp.MustCompile(`^#[a-z0-9][a-z0-9_-]*$`)

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
	case "control":
		if len(args) != 1 {
			usage()
			os.Exit(2)
		}
		os.Exit(runControl(home))
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
	fmt.Fprintln(os.Stderr, "       noticecfg --home DIR control")
}

func configPath(home string) string {
	return filepath.Join(home, "config")
}

func credentialsPath(home string) string {
	return filepath.Join(home, "credentials")
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

func judge(field, value string) (string, bool) {
	switch field {
	case "token_env":
		if !envName.MatchString(value) {
			return fmt.Sprintf("refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: %s", value), false
		}
	case "channel":
		if !channelName.MatchString(value) {
			return fmt.Sprintf("refused: channel must be a channel name matching #[a-z0-9][a-z0-9_-]*; rejected value: %s", value), false
		}
	case "endpoint":
		if !validEndpoint(value) {
			return fmt.Sprintf("refused: endpoint must be an http or https URL with a host; rejected value: %s", value), false
		}
	}
	return "", true
}

func validEndpoint(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func setField(home, field, value string) (string, int) {
	if msg, ok := judge(field, value); !ok {
		return msg, 0
	}
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

// runControl writes the synthetic token into the credentials file, sets
// legal field values, and reports whether that token appears in the set
// messages or in show output.
func runControl(home string) int {
	if err := os.WriteFile(credentialsPath(home), []byte(syntheticToken+"\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	pairs := [][2]string{
		{"token_env", "CHAT_API_TOKEN"},
		{"channel", "#ops"},
		{"endpoint", "https://chat.example/hook"},
	}
	messages := make([]string, 0, len(pairs))
	for _, p := range pairs {
		msg, code := setField(home, p[0], p[1])
		if code != 0 {
			fmt.Fprintln(os.Stderr, msg)
			return code
		}
		messages = append(messages, msg)
	}
	blob := strings.Join(messages, "\n") + "\n" + showText(home)
	if strings.Contains(blob, syntheticToken) {
		fmt.Println("fail")
		return 1
	}
	fmt.Println("pass")
	return 0
}
