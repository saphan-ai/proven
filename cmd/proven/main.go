// Command proven runs a review. This wagon implements sample mode only.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getwd, os.Getenv))
}

func run(args []string, stdout, stderr io.Writer, getwd func() (string, error), env func(string) string) int {
	if len(args) == 0 || args[0] != "scan" {
		fmt.Fprintln(stderr, "usage: proven scan --sample <fixture>")
		return 2
	}
	sample := false
	target := ""
	var include, exclude []string
	baseline := ""
	configPath := ""
	for _, arg := range args[1:] {
		switch {
		case arg == "--sample":
			sample = true
		case strings.HasPrefix(arg, "--exclude="):
			exclude = append(exclude, strings.TrimPrefix(arg, "--exclude="))
		case strings.HasPrefix(arg, "--include="):
			include = append(include, strings.TrimPrefix(arg, "--include="))
		case strings.HasPrefix(arg, "--baseline-commit="):
			baseline = strings.TrimPrefix(arg, "--baseline-commit=")
		case strings.HasPrefix(arg, "--config="):
			configPath = strings.TrimPrefix(arg, "--config=")
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "unknown flag %s\nusage: proven scan --sample <fixture>\n", arg)
			return 2
		case target != "":
			fmt.Fprintln(stderr, "usage: proven scan --sample <fixture>")
			return 2
		default:
			target = arg
		}
	}
	if !sample || target == "" {
		fmt.Fprintln(stderr, "usage: proven scan --sample <fixture>")
		return 2
	}
	wd, err := getwd()
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	rev, err := review.Open(review.Options{
		ConfigPath: configPath,
		Sample:     true,
		LogDir:     filepath.Join(wd, "logs"),
		Env:        env,
	})
	if err != nil {
		var cfg *review.ConfigError
		if errors.As(err, &cfg) {
			fmt.Fprintln(stderr, cfg.Error())
			return 7
		}
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	res, err := rev.Review(context.Background(), review.Request{
		Target:         target,
		BaselineCommit: baseline,
		Include:        include,
		Exclude:        exclude,
	})
	if err != nil {
		var cfg *review.ConfigError
		if errors.As(err, &cfg) {
			fmt.Fprintln(stderr, cfg.Error())
			return 7
		}
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	printResult(stdout, res)
	return 0
}

func printResult(w io.Writer, res review.Result) {
	fmt.Fprintf(w, "review class: %s\n", res.Class)
	fmt.Fprintf(w, "review rule: %s\n", res.ClassRule)
	if res.AuthorModelStatement != "" {
		fmt.Fprintln(w, res.AuthorModelStatement)
	}
	for _, line := range res.ReviewerStatements {
		fmt.Fprintln(w, line)
	}
	for _, f := range res.Findings {
		fmt.Fprintf(w, "finding %s %s %s\n", f.ID, f.Severity, f.Title)
		fmt.Fprintf(w, "  %s:%d-%d\n", f.Location.Path, f.Location.StartLine, f.Location.EndLine)
		fmt.Fprintf(w, "  %s\n", f.Message)
		for _, e := range f.Evidence {
			switch e.Kind {
			case review.EvidenceTest:
				fmt.Fprintf(w, "  evidence test %s %s fails_on_change=%t passes_on_base=%t\n", e.Test.Path, e.Test.Name, e.Test.FailsOnChange, e.Test.PassesOnBase)
				fmt.Fprintf(w, "  command: %s\n", e.Test.Command)
				fmt.Fprintf(w, "  output: %s\n", strings.TrimSpace(e.Test.Output))
			case review.EvidenceQuery:
				fmt.Fprintf(w, "  evidence query %s snapshot=%s\n", e.Query.Engine, e.Query.Snapshot)
				fmt.Fprintf(w, "  SQL: %s\n", e.Query.SQL)
				fmt.Fprintf(w, "  columns: %s\n", strings.Join(e.Query.Columns, ", "))
				for _, row := range e.Query.Rows {
					fmt.Fprintf(w, "  row: %s\n", strings.Join(row, " | "))
				}
			}
		}
	}
	if len(res.FailedServices) == 0 {
		fmt.Fprintln(w, "failed services: none")
	} else {
		for _, s := range res.FailedServices {
			fmt.Fprintf(w, "failed service: %s: %s\n", s.Service, s.Reason)
		}
	}
	if res.TracePath != "" {
		fmt.Fprintf(w, "trace: %s\n", res.TracePath)
	}
}
