// Command proven reviews a directory or a pull request.
package main

import (
	"os"

	"github.com/saphan-ai/proven/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr, nil))
}

// familySentence is the review's same-family line, when the trace has one.
func familySentence(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Kind == "family" && e.Text != "" {
			return e.Text
		}
	}
	return ""
}
